package utils

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func setupMonitoringDB(t *testing.T) {
	t.Helper()
	tmp := t.TempDir() + "/"
	prev := CONFIGFOLDER
	CONFIGFOLDER = tmp
	if err := InitMetricsDatabase(); err != nil {
		t.Fatal("InitMetricsDatabase:", err)
	}
	t.Cleanup(func() {
		CloseMetricsDatabase()
		CONFIGFOLDER = prev
	})
}

func parseFilter(t *testing.T, raw string) map[string]interface{} {
	t.Helper()
	var f map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &f); err != nil {
		t.Fatal("filter fixture:", err)
	}
	return f
}

func TestEventFilterTranslator(t *testing.T) {
	cases := []struct {
		name    string
		dialect string
		filter  string
		wantSQL string
		wantArg []interface{}
	}{
		{
			name:    "implicit equality",
			dialect: DialectSQLite,
			filter:  `{"level": "error"}`,
			wantSQL: "(level = ?)",
			wantArg: []interface{}{"error"},
		},
		{
			name:    "comparison on date accepts RFC3339",
			dialect: DialectSQLite,
			filter:  `{"date": {"$gte": "2024-01-01T00:00:00Z"}}`,
			wantSQL: "((date >= ?))",
			wantArg: []interface{}{int64(1704067200000)},
		},
		{
			name:    "extended-json date wrapper still parses",
			dialect: DialectSQLite,
			filter:  `{"date": {"$lt": {"$date": "2024-01-01T00:00:00Z"}}}`,
			wantSQL: "((date < ?))",
			wantArg: []interface{}{int64(1704067200000)},
		},
		{
			name:    "in list",
			dialect: DialectSQLite,
			filter:  `{"level": {"$in": ["error", "warning"]}}`,
			wantSQL: "((level IN (?, ?)))",
			wantArg: []interface{}{"error", "warning"},
		},
		{
			name:    "or group",
			dialect: DialectSQLite,
			filter:  `{"$or": [{"level": "error"}, {"eventId": "cosmos.proxy"}]}`,
			wantSQL: "(((level = ?) OR (event_id = ?)))",
			wantArg: []interface{}{"error", "cosmos.proxy"},
		},
		{
			name:    "regex becomes a prefix LIKE",
			dialect: DialectSQLite,
			filter:  `{"eventId": {"$regex": "^cosmos"}}`,
			wantSQL: `((event_id LIKE ? ESCAPE '\'))`,
			wantArg: []interface{}{"cosmos%"},
		},
		{
			name:    "escaped dots in a regex are literal (the deployments tab filter)",
			dialect: DialectSQLite,
			filter:  `{"eventId": {"$regex": "^cosmos\\.scheduler\\."}}`,
			wantSQL: `((event_id LIKE ? ESCAPE '\'))`,
			wantArg: []interface{}{"cosmos.scheduler.%"},
		},
		{
			name:    "slashes are literal, escaped or not",
			dialect: DialectSQLite,
			filter:  `{"object": {"$regex": "^route@/api\\/v1"}}`,
			wantSQL: `((object LIKE ? ESCAPE '\'))`,
			wantArg: []interface{}{"route@/api/v1%"},
		},
		{
			name:    "wildcards and end anchor",
			dialect: DialectSQLite,
			filter:  `{"label": {"$regex": "deploy.*done$"}}`,
			wantSQL: `((label LIKE ? ESCAPE '\'))`,
			wantArg: []interface{}{"%deploy%done"},
		},
		{
			name:    "json path is bound, never interpolated (sqlite)",
			dialect: DialectSQLite,
			filter:  `{"data.route": "web"}`,
			wantSQL: "(CAST(json_extract(data, ?) AS TEXT) = ?)",
			wantArg: []interface{}{"$.route", "web"},
		},
		{
			name:    "json path is bound, never interpolated (postgres)",
			dialect: DialectPostgres,
			filter:  `{"data.a.b": "x"}`,
			wantSQL: "((data::jsonb #>> ?::text[]) = ?)",
			wantArg: []interface{}{"{a,b}", "x"},
		},
		{
			name:    "exists on a json path",
			dialect: DialectSQLite,
			filter:  `{"data.route": {"$exists": true}}`,
			wantSQL: "((CAST(json_extract(data, ?) AS TEXT) IS NOT NULL))",
			wantArg: []interface{}{"$.route"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotSQL, gotArgs, err := TranslateEventFilter(c.dialect, parseFilter(t, c.filter))
			if err != nil {
				t.Fatal("translate:", err)
			}
			if gotSQL != c.wantSQL {
				t.Fatalf("sql = %q, want %q", gotSQL, c.wantSQL)
			}
			if !reflect.DeepEqual(gotArgs, c.wantArg) {
				t.Fatalf("args = %#v, want %#v", gotArgs, c.wantArg)
			}
		})
	}
}

func TestEventFilterRejectsUnsupported(t *testing.T) {
	cases := []string{
		`{"$where": "this.level == 'error'"}`,
		`{"level": {"$expr": 1}}`,
		`{"password": "x"}`,
		`{"label": {"$regex": "a|b"}}`,
		`{"label": {"$regex": "\\d+"}}`,
		`{"label": {"$regex": "a\\"}}`,
		`{"data.route; DROP TABLE events": "x"}`,
		`{"$or": []}`,
		`{"level": {"$in": []}}`,
	}

	for _, raw := range cases {
		if _, _, err := TranslateEventFilter(DialectSQLite, parseFilter(t, raw)); err == nil {
			t.Fatalf("expected %s to be rejected", raw)
		}
	}
}

// A hostile value must travel as a parameter: it matches its own row and nothing else,
// and the table it names is still there afterwards.
func TestEventFilterValueIsParameterized(t *testing.T) {
	setupMonitoringDB(t)

	payload := `'; DROP TABLE events;--`
	if err := InsertEvent(Event{EventId: "cosmos.test", Label: payload, Level: "error", Date: time.Now()}); err != nil {
		t.Fatal("InsertEvent:", err)
	}
	if err := InsertEvent(Event{EventId: "cosmos.test", Label: "innocent", Level: "error", Date: time.Now()}); err != nil {
		t.Fatal("InsertEvent:", err)
	}

	events, total, err := QueryEvents(EventQuery{
		From:   time.Now().Add(-time.Hour),
		To:     time.Now().Add(time.Hour),
		Levels: []string{"error"},
		Filter: map[string]interface{}{"label": payload},
	})
	if err != nil {
		t.Fatal("QueryEvents:", err)
	}
	if total != 1 || len(events) != 1 || events[0].Label != payload {
		t.Fatalf("injection payload was not treated as a value: total %d, events %+v", total, events)
	}

	// the table survived, so nothing was executed
	if _, _, err := QueryEvents(EventQuery{From: time.Now().Add(-time.Hour), To: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal("events table is gone:", err)
	}
}

func TestEventQueryPagination(t *testing.T) {
	setupMonitoringDB(t)

	base := time.Now().Add(-time.Minute)
	for i := 0; i < 5; i++ {
		if err := InsertEvent(Event{
			EventId: "cosmos.test",
			Level:   "info",
			Date:    base.Add(time.Duration(i) * time.Second),
			Data:    map[string]interface{}{"i": i},
		}); err != nil {
			t.Fatal("InsertEvent:", err)
		}
	}

	q := EventQuery{
		From:   base.Add(-time.Hour),
		To:     base.Add(time.Hour),
		Levels: []string{"info", "warning", "error", "important", "success"},
		Limit:  2,
	}

	first, total, err := QueryEvents(q)
	if err != nil || total != 5 || len(first) != 2 {
		t.Fatalf("first page: %d events, total %d, %v", len(first), total, err)
	}

	q.Cursor = first[len(first)-1].Id
	second, _, err := QueryEvents(q)
	if err != nil || len(second) != 2 {
		t.Fatalf("second page: %d events, %v", len(second), err)
	}
	if second[0].Id >= first[len(first)-1].Id {
		t.Fatalf("cursor did not advance: %d then %d", first[len(first)-1].Id, second[0].Id)
	}
	if second[0].Data["i"] == nil {
		t.Fatalf("event data did not round-trip: %+v", second[0])
	}

	// the level ladder excludes what it does not list
	q.Cursor = 0
	q.Levels = []string{"error"}
	_, total, err = QueryEvents(q)
	if err != nil || total != 0 {
		t.Fatalf("level filter: total %d, %v", total, err)
	}
}

// Older clients send page=0 on the first load, which the ObjectID cursor silently
// ignored. An integer cursor must keep meaning "first page", not "id < 0".
func TestEventQueryZeroCursorIsFirstPage(t *testing.T) {
	setupMonitoringDB(t)

	base := time.Now().Add(-time.Minute)
	for i := 0; i < 3; i++ {
		if err := InsertEvent(Event{EventId: "cosmos.test", Level: "info", Date: base.Add(time.Duration(i) * time.Second)}); err != nil {
			t.Fatal("InsertEvent:", err)
		}
	}

	q := EventQuery{From: base.Add(-time.Hour), To: base.Add(time.Hour), Levels: []string{"info"}}

	unset, _, err := QueryEvents(q)
	if err != nil || len(unset) != 3 {
		t.Fatalf("unset cursor: %d events, %v", len(unset), err)
	}

	for _, cursor := range []int64{0, -1} {
		q.Cursor = cursor
		page, _, err := QueryEvents(q)
		if err != nil {
			t.Fatalf("cursor %d: %v", cursor, err)
		}
		if len(page) != len(unset) || page[0].Id != unset[0].Id {
			t.Fatalf("cursor %d returned %d events, want the same first page as an unset cursor (%d)",
				cursor, len(page), len(unset))
		}
	}
}

func TestNotificationsRoundTrip(t *testing.T) {
	setupMonitoringDB(t)

	for i := 0; i < 3; i++ {
		if err := InsertNotification(Notification{
			Recipient: "alice",
			Title:     "t",
			Message:   "m",
			Level:     "info",
			Date:      time.Now(),
			Actions:   []NotificationActions{{Text: "go", Link: "/x"}},
		}); err != nil {
			t.Fatal("InsertNotification:", err)
		}
	}
	if err := InsertNotification(Notification{Recipient: "bob", Date: time.Now()}); err != nil {
		t.Fatal("InsertNotification bob:", err)
	}

	notifications, err := ListNotifications("alice", 0, 20)
	if err != nil || len(notifications) != 3 {
		t.Fatalf("ListNotifications = %d, %v", len(notifications), err)
	}
	if len(notifications[0].Actions) != 1 || notifications[0].Actions[0].Link != "/x" {
		t.Fatalf("actions did not round-trip: %+v", notifications[0])
	}
	if notifications[0].Read {
		t.Fatal("fresh notification should be unread")
	}

	// another recipient's id must not be markable
	bobs, _ := ListNotifications("bob", 0, 20)
	matched, err := MarkNotificationsRead("alice", []int64{bobs[0].ID})
	if err != nil || matched != 0 {
		t.Fatalf("cross-recipient mark: matched %d, %v", matched, err)
	}

	matched, err = MarkNotificationsRead("alice", []int64{notifications[0].ID, notifications[1].ID})
	if err != nil || matched != 2 {
		t.Fatalf("MarkNotificationsRead: matched %d, %v", matched, err)
	}
	notifications, _ = ListNotifications("alice", 0, 20)
	if !notifications[0].Read || !notifications[1].Read || notifications[2].Read {
		t.Fatalf("read flags wrong: %+v", notifications)
	}

	// keyset pagination walks backwards
	page, err := ListNotifications("alice", notifications[0].ID, 20)
	if err != nil || len(page) != 2 {
		t.Fatalf("cursor page = %d, %v", len(page), err)
	}
}

func TestDatabaseRetention(t *testing.T) {
	setupMonitoringDB(t)

	old := time.Now().AddDate(0, -2, 0)
	fresh := time.Now()

	for _, d := range []time.Time{old, fresh} {
		if err := InsertEvent(Event{EventId: "cosmos.test", Level: "info", Date: d}); err != nil {
			t.Fatal("InsertEvent:", err)
		}
		if err := InsertNotification(Notification{Recipient: "alice", Date: d}); err != nil {
			t.Fatal("InsertNotification:", err)
		}
	}

	db, err := MonitoringWriteDB()
	if err != nil {
		t.Fatal(err)
	}
	insert := MonitoringRebind("INSERT INTO metric_values(node, key, granularity, date, value, avg_index, expire) VALUES(?, ?, ?, ?, ?, ?, ?)")
	// one already expired, one still live
	if _, err := db.Exec(insert, MonitoringNode(), "cosmos.a", "raw", 1, 1, 0, NowMillis()-1000); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(insert, MonitoringNode(), "cosmos.b", "raw", 2, 1, 0, NowMillis()+3600000); err != nil {
		t.Fatal(err)
	}

	RunDatabaseRetention()

	count := func(query string, args ...interface{}) int {
		var n int
		if err := db.QueryRow(MonitoringRebind(query), args...).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}

	cutoff := TimeToMillis(time.Now().AddDate(0, -1, 0))
	if n := count("SELECT COUNT(*) FROM events WHERE date < ?", cutoff); n != 0 {
		t.Fatalf("month-old events survived retention: %d", n)
	}
	if n := count("SELECT COUNT(*) FROM notifications WHERE date < ?", cutoff); n != 0 {
		t.Fatalf("month-old notifications survived retention: %d", n)
	}
	if n := count("SELECT COUNT(*) FROM events WHERE date >= ?", cutoff); n < 1 {
		t.Fatal("retention deleted fresh events")
	}
	if n := count("SELECT COUNT(*) FROM notifications WHERE date >= ?", cutoff); n != 1 {
		t.Fatalf("retention deleted fresh notifications: %d", n)
	}
	if n := count("SELECT COUNT(*) FROM metric_values"); n != 1 {
		t.Fatalf("expected only the live metric bucket to remain, got %d", n)
	}
}

func TestWildcardToLike(t *testing.T) {
	cases := map[string]string{
		"cosmos.system.cpu":     `cosmos.system.cpu%`,
		"cosmos.system.*":       `cosmos.system.%%`,
		"cosmos.disk.a_b":       `cosmos.disk.a\_b%`,
		"cosmos.*.netRx.*":      `cosmos.%.netRx.%%`,
		"cosmos.100%.something": `cosmos.100\%.something%`,
	}
	for in, want := range cases {
		if got := WildcardToLike(in); got != want {
			t.Fatalf("WildcardToLike(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMonitoringRebind(t *testing.T) {
	if got := rebindFor(DialectSQLite, "a = ? AND b = ?"); got != "a = ? AND b = ?" {
		t.Fatalf("sqlite rebind changed the query: %q", got)
	}
	if got := rebindFor(DialectPostgres, "a = ? AND b = ?"); got != "a = $1 AND b = $2" {
		t.Fatalf("postgres rebind = %q", got)
	}
}

func TestPostgresDSN(t *testing.T) {
	got := postgresDSN(DatabaseConfig{
		PostgresHost:     "db.internal:6543",
		PostgresDatabase: "cosmos",
		PostgresUsername: "user",
		PostgresPassword: "p@ss word",
	})
	want := "postgres://user:p%40ss%20word@db.internal:6543/cosmos"
	if got != want {
		t.Fatalf("postgresDSN = %q, want %q", got, want)
	}
	if got := postgresDSN(DatabaseConfig{PostgresHost: "db.internal"}); got != "postgres://db.internal:5432/cosmos" {
		t.Fatalf("default port/database = %q", got)
	}
}

// The node column must never be empty, or two servers sharing a Postgres collide.
func TestResolveNodeName(t *testing.T) {
	config := Config{}
	config.Database.NodeName = "explicit"
	if got := resolveNodeName(config); got != "explicit" {
		t.Fatalf("NodeName override ignored: %q", got)
	}

	config.Database.NodeName = ""
	config.ConstellationConfig.ThisDeviceName = "constellation-node"
	if got := resolveNodeName(config); got != "constellation-node" {
		t.Fatalf("constellation fallback ignored: %q", got)
	}

	config.ConstellationConfig.ThisDeviceName = ""
	if got := resolveNodeName(config); got == "" {
		t.Fatal("node name must never be empty")
	}
}
