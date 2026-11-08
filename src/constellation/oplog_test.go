package constellation

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/azukaar/cosmos-server/src/utils"
)

// makeEnvelope builds a db envelope the way publishOp does, so the tests
// exercise the real encode path rather than a hand-written payload.
func makeEnvelope(t *testing.T, m utils.Mutation, origin string) OpEnvelope {
	t.Helper()

	doc, err := utils.EncodeOpDoc(m)
	if err != nil {
		t.Fatal("EncodeOpDoc:", err)
	}
	filter, err := utils.NormalizeOpFilter(m.Table, m.Filter)
	if err != nil {
		t.Fatal("NormalizeOpFilter:", err)
	}
	return OpEnvelope{
		V: oplogEnvVersion, Epoch: utils.GetOplogEpoch(), Kind: "db",
		Table: m.Table, Op: m.Op, Filter: filter, Doc: doc,
		Origin: origin, ReqID: utils.GenerateRandomString(8),
	}
}

// roundTrip marshals and unmarshals an envelope, so a test never applies the
// in-memory value the encoder happened to leave behind.
func roundTrip(t *testing.T, env OpEnvelope) OpEnvelope {
	t.Helper()

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatal("marshal envelope:", err)
	}
	var out OpEnvelope
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal("unmarshal envelope:", err)
	}
	return out
}

func TestUnitOplogEnvelopeCodec(t *testing.T) {
	setupTestEnv(t, nil)

	// the struct json tags hide Password/APIKey; the wire form must not
	user := utils.User{Nickname: "alice", Password: "secret-hash", MFAKey: "mfa", Role: utils.ADMIN,
		CreatedAt: time.Now().UTC().Truncate(time.Second)}
	env := roundTrip(t, makeEnvelope(t, utils.Mutation{Table: "users", Op: "insert", Doc: user}, "self"))

	doc, err := utils.DecodeOpDoc(env.Table, env.Op, env.Doc)
	if err != nil {
		t.Fatal("DecodeOpDoc:", err)
	}
	got, ok := doc.(utils.User)
	if !ok {
		t.Fatalf("decoded doc is %T, want utils.User", doc)
	}
	if got.Password != user.Password || got.MFAKey != user.MFAKey {
		t.Fatalf("secrets dropped by the wire form: %+v", got)
	}
	if !got.CreatedAt.Equal(user.CreatedAt) || got.Role != user.Role {
		t.Fatalf("fields lost in round-trip: %+v", got)
	}

	device := utils.ConstellationDevice{DeviceName: "d1", IP: "192.168.201.5", APIKey: "device-key",
		PublicKey: "priv", Tags: []string{"a", "b"}}
	env = roundTrip(t, makeEnvelope(t, utils.Mutation{Table: "devices", Op: "insert", Doc: device}, "self"))
	doc, err = utils.DecodeOpDoc(env.Table, env.Op, env.Doc)
	if err != nil {
		t.Fatal("DecodeOpDoc devices:", err)
	}
	gotDev := doc.(utils.ConstellationDevice)
	if gotDev.APIKey != device.APIKey || gotDev.PublicKey != device.PublicKey {
		t.Fatalf("device secrets dropped: %+v", gotDev)
	}
	if strings.Join(gotDev.Tags, ",") != "a,b" {
		t.Fatalf("tags lost: %v", gotDev.Tags)
	}

	// update maps travel already converted to DB values, so replicas write
	// byte-identical rows no matter what Go type the caller passed
	env = roundTrip(t, makeEnvelope(t, utils.Mutation{
		Table:  "devices",
		Op:     "updateMany",
		Filter: map[string]interface{}{"DeviceName": "d1", "Blocked": false},
		Doc:    map[string]interface{}{"Blocked": true, "Tags": []string{"x"}},
	}, "self"))

	var fields map[string]interface{}
	if err := json.Unmarshal(env.Doc, &fields); err != nil {
		t.Fatal("unmarshal update doc:", err)
	}
	if fields["Blocked"] != float64(1) {
		t.Fatalf("bool not normalized to its DB form: %#v", fields["Blocked"])
	}
	if fields["Tags"] != `["x"]` {
		t.Fatalf("tags not normalized to their DB form: %#v", fields["Tags"])
	}
	if env.Filter["Blocked"] != float64(0) {
		t.Fatalf("filter not normalized: %#v", env.Filter["Blocked"])
	}

	// an unknown column is rejected before it can reach the log
	if _, err := utils.EncodeOpDoc(utils.Mutation{Table: "devices", Op: "update",
		Doc: map[string]interface{}{"NotAColumn": 1}}); err == nil {
		t.Fatal("unknown field accepted into an op")
	}
}

func TestUnitOplogSeqContiguity(t *testing.T) {
	cases := []struct {
		seq, last uint64
		want      seqAction
	}{
		{1, 0, seqApply},
		{8, 7, seqApply},
		{7, 7, seqSkip}, // redelivery
		{3, 7, seqSkip}, // stale replay
		{9, 7, seqGap},  // retention dropped 8
		{99, 0, seqGap}, // brand-new node behind a long log
	}
	for _, c := range cases {
		if got := oplogSeqAction(c.seq, c.last); got != c.want {
			t.Errorf("oplogSeqAction(%d, %d) = %v, want %v", c.seq, c.last, got, c.want)
		}
	}
}

func TestUnitOplogApplyIsIdempotent(t *testing.T) {
	setupTestEnv(t, nil)

	env := roundTrip(t, makeEnvelope(t, utils.Mutation{
		Table: "users", Op: "insert", Doc: utils.User{Nickname: "bob", Password: "h"},
	}, "someone-else"))

	if err := oplogApplyEnvelope(env, 1); err != nil {
		t.Fatal("first apply:", err)
	}
	if seq := utils.GetLastAppliedSeq(); seq != 1 {
		t.Fatalf("seq not committed with the data: %d", seq)
	}

	// replaying the same sequence is gated before apply is ever reached
	if got := oplogSeqAction(1, utils.GetLastAppliedSeq()); got != seqSkip {
		t.Fatalf("replay of seq 1 not skipped: %v", got)
	}

	count, err := utils.CountUsers()
	if err != nil || count != 1 {
		t.Fatalf("CountUsers = %d, %v (want 1)", count, err)
	}
}

func TestUnitOplogRejectionCommitsSeqAndErrorsWaiter(t *testing.T) {
	setupTestEnv(t, nil)

	// users are keyed by nickname, so the duplicate arrives as a PRIMARY KEY
	// violation — it must still be a rejection, not a loop-halting fault
	user := utils.User{Nickname: "dup", Password: "h"}
	if err := oplogApplyEnvelope(roundTrip(t, makeEnvelope(t,
		utils.Mutation{Table: "users", Op: "insert", Doc: user}, "peer")), 1); err != nil {
		t.Fatal("first insert:", err)
	}

	env := roundTrip(t, makeEnvelope(t, utils.Mutation{Table: "users", Op: "insert", Doc: user}, "self"))
	wait := registerOpWaiter(env.ReqID)
	defer releaseOpWaiter(env.ReqID)

	// a rejection is an outcome, not a fault: it must not halt the loop
	if err := oplogApplyEnvelope(env, 2); err != nil {
		t.Fatal("rejection escalated to a loop-halting error:", err)
	}

	// ...and it still consumes its sequence, so every replica stays aligned
	if seq := utils.GetLastAppliedSeq(); seq != 2 {
		t.Fatalf("rejected op did not consume its seq: %d", seq)
	}

	select {
	case err := <-wait:
		var ec *utils.ErrConstraint
		if !errors.As(err, &ec) {
			t.Fatalf("waiter got %v, want *utils.ErrConstraint", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("originator was never told its op was rejected")
	}

	count, _ := utils.CountUsers()
	if count != 1 {
		t.Fatalf("rejected row was written anyway: %d users", count)
	}
}

func TestUnitOplogWaiterRegistry(t *testing.T) {
	setupTestEnv(t, nil)

	wait := registerOpWaiter("req-1")

	// an op from another node must not release our waiter
	notifyOpWaiter("req-other", nil)
	select {
	case <-wait:
		t.Fatal("waiter released by an unrelated op")
	case <-time.After(50 * time.Millisecond):
	}

	notifyOpWaiter("req-1", nil)
	select {
	case err := <-wait:
		if err != nil {
			t.Fatalf("waiter got %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("waiter never released")
	}

	// after release the id is gone, and notifying it must not panic or block
	releaseOpWaiter("req-1")
	notifyOpWaiter("req-1", nil)
}

func TestUnitOplogWriteModeMatrix(t *testing.T) {
	// constellation disabled: single-node installs stay on the direct path
	setupTestEnv(t, func(c *utils.Config) { c.ConstellationConfig.Enabled = false })
	if got := oplogWriteMode(); got != oplogDirect {
		t.Fatalf("disabled constellation: mode %v, want direct", got)
	}

	setupTestEnv(t, nil)
	writeNebulaYML(t, map[string]interface{}{
		"cstln_device_name": "self",
		"cstln_ip":          "192.168.201.1/24",
		"cstln_api_key":     "k",
	})

	// enabled but no peers: still standalone, still direct
	if got := oplogWriteMode(); got != oplogDirect {
		t.Fatalf("no peers: mode %v, want direct", got)
	}

	// a peer manager exists, so this node is part of a constellation
	deviceCacheMux.Lock()
	CachedDevices = map[string]utils.ConstellationDevice{
		"peer": {DeviceName: "peer", IP: "192.168.201.2/24", CosmosNode: 2},
	}
	deviceCacheMux.Unlock()

	// pre-NATS boot (migrations, new install): direct, single writer by construction
	NebulaStarted.Store(false)
	if got := oplogWriteMode(); got != oplogDirect {
		t.Fatalf("pre-NATS boot: mode %v, want direct", got)
	}

	// up but not attached to the log: read-only, whoever you are
	NebulaStarted.Store(true)
	oplogAttached.Store(false)
	if got := oplogWriteMode(); got != oplogReadOnly {
		t.Fatalf("detached: mode %v, want read-only", got)
	}

	oplogAttached.Store(true)
	if got := oplogWriteMode(); got != oplogPublish {
		t.Fatalf("attached: mode %v, want publish", got)
	}
}

// Regression: a node mid-RestartNebula must be read-only, never direct. stop()
// clears NebulaStarted and the device cache together while the apply loop
// detaches, so a real cluster member briefly looks both stopped and standalone —
// and since the apply loop answers a device write with `go RestartNebula()`, that
// window opens after every device write. A direct write there is never published
// and is silently discarded by the next snapshot.
func TestUnitOplogRestartWindowIsReadOnly(t *testing.T) {
	setupTestEnv(t, nil)
	writeNebulaYML(t, map[string]interface{}{
		"cstln_device_name": "self", "cstln_ip": "192.168.201.1/24", "cstln_api_key": "k",
	})

	// this node has materialized from a log
	if err := utils.MarkOplogBootstrapped(utils.GetOplogEpoch()); err != nil {
		t.Fatal("MarkOplogBootstrapped:", err)
	}

	// exactly what stop() leaves behind: detached, not started, cache emptied
	oplogAttached.Store(false)
	NebulaStarted.Store(false)
	deviceCacheMux.Lock()
	CachedDevices = map[string]utils.ConstellationDevice{}
	CachedDeviceNames = map[string]string{}
	deviceCacheMux.Unlock()

	if got := oplogWriteMode(); got != oplogReadOnly {
		t.Fatalf("restart window: mode %v, want read-only — a direct write here forks this node from the log", got)
	}

	// and the write path must actually refuse rather than write locally
	err := publishOp(utils.Mutation{Table: "users", Op: "insert", Doc: utils.User{Nickname: "forked"}})
	if !errors.Is(err, utils.ErrReadOnly) {
		t.Fatalf("write during restart window returned %v, want ErrReadOnly", err)
	}
	if count, _ := utils.CountUsers(); count != 0 {
		t.Fatalf("write during restart window landed locally: %d users", count)
	}

	// ResetNebula disables constellation but leaves the sequence behind; such a
	// node must become writable again rather than be stranded read-only
	config := utils.ReadConfigFromFile()
	config.ConstellationConfig.Enabled = false
	utils.SetBaseMainConfig(config)

	if got := oplogWriteMode(); got != oplogDirect {
		t.Fatalf("after constellation disabled: mode %v, want direct (stale seq must not strand it)", got)
	}
}

func TestUnitOplogReadOnlyRejectsWritesButDropsBestEffort(t *testing.T) {
	setupTestEnv(t, nil)
	writeNebulaYML(t, map[string]interface{}{
		"cstln_device_name": "self", "cstln_ip": "192.168.201.1/24", "cstln_api_key": "k",
	})
	deviceCacheMux.Lock()
	CachedDevices = map[string]utils.ConstellationDevice{
		"peer": {DeviceName: "peer", IP: "192.168.201.2/24", CosmosNode: 2},
	}
	deviceCacheMux.Unlock()
	NebulaStarted.Store(true)
	oplogAttached.Store(false)

	err := publishOp(utils.Mutation{Table: "users", Op: "insert", Doc: utils.User{Nickname: "nope"}})
	if !errors.Is(err, utils.ErrReadOnly) {
		t.Fatalf("write on a read-only node returned %v, want ErrReadOnly", err)
	}
	if count, _ := utils.CountUsers(); count != 0 {
		t.Fatalf("read-only node wrote locally anyway: %d users", count)
	}

	// LastLogin is best-effort: dropped rather than written locally, because a
	// local-only write would fork this node's materialization forever
	err = publishOp(utils.Mutation{
		Table: "users", Op: "update", BestEffort: true,
		Filter: map[string]interface{}{"Nickname": "nope"},
		Doc:    map[string]interface{}{"LastLogin": time.Now()},
	})
	if err != nil {
		t.Fatalf("best-effort op failed instead of being dropped: %v", err)
	}
}

func TestUnitOplogDirectWritePath(t *testing.T) {
	setupTestEnv(t, func(c *utils.Config) { c.ConstellationConfig.Enabled = false })

	if err := publishOp(utils.Mutation{
		Table: "users", Op: "insert", Doc: utils.User{Nickname: "solo", Password: "h"},
	}); err != nil {
		t.Fatal("direct write:", err)
	}

	u, err := utils.GetUser("solo")
	if err != nil || u.Password != "h" {
		t.Fatalf("direct write did not land: %+v %v", u, err)
	}
	// a direct write is not a log entry and must not move the position
	if seq := utils.GetLastAppliedSeq(); seq != 0 {
		t.Fatalf("direct write moved the log position to %d", seq)
	}
}

func TestUnitOplogOIDCSigningKeyDomain(t *testing.T) {
	setupTestEnv(t, nil)

	apply := func(key string) {
		t.Helper()
		state, err := json.Marshal(OIDCSigningKeyPayload{Key: key})
		if err != nil {
			t.Fatal(err)
		}
		if err := oplogDomains[DomainOIDCSigningKey].Apply(state); err != nil {
			t.Fatal(err)
		}
	}

	apply("cluster-key")
	http := utils.GetMainConfig().HTTPConfig
	if http.OIDCSigningKey != "cluster-key" || !http.OIDCSigningKeyShared {
		t.Fatalf("key not installed: %q shared=%v", http.OIDCSigningKey, http.OIDCSigningKeyShared)
	}

	apply("")
	if got := utils.GetMainConfig().HTTPConfig.OIDCSigningKey; got != "cluster-key" {
		t.Fatalf("an empty state wiped the key: %q", got)
	}
}

func TestUnitOplogDomainRegistry(t *testing.T) {
	setupTestEnv(t, nil)

	for _, name := range []string{DomainAuthKeys, DomainDNS, DomainAPITokens, DomainRoles,
		DomainOpenIDClients, DomainFileCACrt, DomainFileCAKey, DomainFileRclone, DomainDatabase,
		DomainDNSZones, DomainZoneCerts, DomainOIDCSigningKey} {
		d, ok := oplogDomains[name]
		if !ok {
			t.Fatalf("domain %q is not registered", name)
		}
		if d.Apply == nil || d.Snapshot == nil {
			t.Fatalf("domain %q is missing Apply or Snapshot", name)
		}
	}

	// config domain: apply full state, read it back through Snapshot
	tokens := map[string]utils.APITokenConfig{"ci": {Description: "ci token"}}
	state, err := json.Marshal(tokens)
	if err != nil {
		t.Fatal(err)
	}
	if err := applyDomainLocal(DomainAPITokens, state); err != nil {
		t.Fatal("applyDomainLocal api_tokens:", err)
	}
	if got := utils.GetMainConfig().APITokens["ci"].Description; got != "ci token" {
		t.Fatalf("api_tokens not applied: %q", got)
	}

	snap, err := oplogDomains[DomainAPITokens].Snapshot()
	if err != nil {
		t.Fatal("snapshot api_tokens:", err)
	}
	var back map[string]utils.APITokenConfig
	if err := json.Unmarshal(snap, &back); err != nil || back["ci"].Description != "ci token" {
		t.Fatalf("api_tokens snapshot round-trip failed: %s (%v)", snap, err)
	}

	// file domain: applied via tmp+rename, and an empty payload never truncates
	fileState, _ := json.Marshal(FilePayload{Data: "aGVsbG8="}) // "hello"
	if err := applyDomainLocal(DomainFileCACrt, fileState); err != nil {
		t.Fatal("applyDomainLocal file:ca.crt:", err)
	}
	data, err := readConfigFile("ca.crt")
	if err != nil || string(data) != "hello" {
		t.Fatalf("ca.crt = %q, %v", data, err)
	}

	empty, _ := json.Marshal(FilePayload{})
	if err := applyDomainLocal(DomainFileCACrt, empty); err != nil {
		t.Fatal("empty file payload:", err)
	}
	if data, _ := readConfigFile("ca.crt"); string(data) != "hello" {
		t.Fatalf("an empty payload wiped the CA: %q", data)
	}

	// snapshotDomains covers every registered domain
	all := snapshotDomains()
	for name := range oplogDomains {
		if _, ok := all[name]; !ok {
			t.Errorf("snapshotDomains omitted %q", name)
		}
	}

	if err := applyDomainLocal("no-such-domain", state); err == nil {
		t.Fatal("unknown domain accepted")
	}
}

// The database domain replicates the Postgres connection only; NodeName stays
// node-local as this node's metrics identity.
// Apply directly in both: the zones reaction restarts the HTTP server, absent in a unit env
func applyDomainNoReact(t *testing.T, name string, state interface{}) {
	t.Helper()
	raw, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	utils.ConfigLock.Lock()
	err = oplogDomains[name].Apply(raw)
	utils.ConfigLock.Unlock()
	if err != nil {
		t.Fatal("apply "+name+":", err)
	}
}

func TestUnitOplogZonesDomains(t *testing.T) {
	setupTestEnv(t, nil)

	applyDomainNoReact(t, DomainDNSZones, []utils.DNSZoneConfig{
		{Zone: "domain.com", DNSChallengeProvider: "cloudflare"},
		{Zone: "other.org", DNSChallengeProvider: "cloudflare"},
	})
	applyDomainNoReact(t, DomainZoneCerts, map[string]utils.ZoneCert{
		"domain.com": {TLSCert: "cert-a", TLSKey: "key-a"},
		"other.org":  {TLSCert: "cert-b", TLSKey: "key-b"},
	})

	if got := utils.GetMainConfig().HTTPConfig; len(got.DNSZones) != 2 || got.ZoneCerts["other.org"].TLSKey != "key-b" {
		t.Fatalf("zones or zone certs not applied: %+v", got.DNSZones)
	}

	// a node that has no certificate yet can't wipe everyone's
	applyDomainNoReact(t, DomainZoneCerts, map[string]utils.ZoneCert{})
	if len(utils.GetMainConfig().HTTPConfig.ZoneCerts) != 2 {
		t.Fatal("an empty zone_certs state must not be applied")
	}

	// the certificate of a deleted zone goes with it
	applyDomainNoReact(t, DomainDNSZones, []utils.DNSZoneConfig{{Zone: "domain.com", DNSChallengeProvider: "cloudflare"}})
	certs := utils.GetMainConfig().HTTPConfig.ZoneCerts
	if _, ok := certs["other.org"]; ok || certs["domain.com"].TLSCert != "cert-a" {
		t.Fatalf("zone certs after zone deletion: %+v", certs)
	}
}

func TestUnitOplogDatabaseDomain(t *testing.T) {
	setupTestEnv(t, nil)

	// Apply reads the file, not the in-memory config
	local := utils.ReadConfigFromFile()
	local.Database.NodeName = "node-local"
	utils.SetBaseMainConfig(local)

	payload := DatabasePayload{
		PostgresHost: "db.example.com:5432", PostgresDatabase: "cosmos",
		PostgresUsername: "cosmos", PostgresPassword: "pg-secret",
	}
	state, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}

	// Apply directly: React would reopen the monitoring store, absent in a unit env
	d := oplogDomains[DomainDatabase]
	utils.ConfigLock.Lock()
	err = d.Apply(state)
	utils.ConfigLock.Unlock()
	if err != nil {
		t.Fatal("apply database domain:", err)
	}

	got := utils.GetMainConfig().Database
	if got.PostgresHost != payload.PostgresHost || got.PostgresDatabase != payload.PostgresDatabase ||
		got.PostgresUsername != payload.PostgresUsername || got.PostgresPassword != payload.PostgresPassword {
		t.Fatalf("connection not applied: %+v", got)
	}
	if got.NodeName != "node-local" {
		t.Fatalf("NodeName = %q, want \"node-local\" — a replicated identity puts every "+
			"node's metrics under the same name", got.NodeName)
	}

	snap, err := d.Snapshot()
	if err != nil {
		t.Fatal("snapshot database:", err)
	}
	var back DatabasePayload
	if err := json.Unmarshal(snap, &back); err != nil {
		t.Fatal("unmarshal snapshot:", err)
	}
	if back != payload {
		t.Fatalf("snapshot round-trip lost fields: %+v", back)
	}
	if strings.Contains(string(snap), "node-local") {
		t.Fatalf("snapshot carries this node's metrics identity: %s", snap)
	}
}

// A node that predates a domain must skip its ops, not halt — the apply loop is
// contiguous, and halting would strand it for the rest of a rolling upgrade.
func TestUnitOplogUnknownDomainIsSkipped(t *testing.T) {
	setupTestEnv(t, nil)

	env := OpEnvelope{
		V: oplogEnvVersion, Epoch: utils.GetOplogEpoch(), Kind: "domain",
		Domain: "domain-from-a-newer-node", Op: "set", Doc: json.RawMessage(`{"anything":1}`),
		Origin: "peer", ReqID: utils.GenerateRandomString(8),
	}
	if err := oplogApplyEnvelope(roundTrip(t, env), 1); err != nil {
		t.Fatalf("unknown domain halted the apply loop: %v", err)
	}
	if seq := utils.GetLastAppliedSeq(); seq != 1 {
		t.Fatalf("skipped entry did not consume its seq: %d — the loop would gap on the next op", seq)
	}

	// a known domain's genuine failure stays fatal
	bad := env
	bad.Domain = DomainDatabase
	bad.Doc = json.RawMessage(`"not-an-object"`)
	if err := oplogApplyEnvelope(roundTrip(t, bad), 2); err == nil {
		t.Fatal("a malformed payload for a known domain was swallowed")
	}
	if seq := utils.GetLastAppliedSeq(); seq != 1 {
		t.Fatalf("a failed apply committed its seq anyway: %d", seq)
	}
}

func TestUnitOplogEpochFencesSubjects(t *testing.T) {
	setupTestEnv(t, nil)

	if name := oplogStreamName(1); name != "cosmos-oplog-e1" {
		t.Fatalf("stream name = %q", name)
	}
	// a stale-epoch node publishes to a subject no current stream listens on,
	// which is the whole fencing mechanism
	if oplogSubjectPrefix(1) == oplogSubjectPrefix(2) {
		t.Fatal("epochs share a subject prefix, publishes would not be fenced")
	}
	if !strings.HasPrefix(oplogSubjectWildcard(7), "cosmos._global_.oplog.e7.") {
		t.Fatalf("wildcard = %q", oplogSubjectWildcard(7))
	}

	subject := oplogSubject(OpEnvelope{Epoch: 3, Kind: "domain", Domain: DomainFileCACrt})
	if strings.Count(subject, ".") != 5 || strings.Contains(subject, ":") {
		t.Fatalf("domain subject is not a safe token: %q", subject)
	}

	// an entry from another epoch is not applied as if it were ours
	env := roundTrip(t, makeEnvelope(t, utils.Mutation{
		Table: "users", Op: "insert", Doc: utils.User{Nickname: "ghost"},
	}, "peer"))
	env.Epoch = 99
	if env.Epoch == utils.GetOplogEpoch() {
		t.Fatal("test setup: epochs should differ")
	}
}

// Bug A: a founder seeds at seq 0 and serves snapshots stamped seq 0, so a joiner
// that installs one is STILL at seq 0. Keyed on the sequence, the always-snapshot
// rule re-fired forever and the joiner never attached — its log showed repeated
// "Fast-forwarded to seq 0" and zero "Attached" lines.
func TestUnitOplogBootstrapMarkerSurvivesSeqZero(t *testing.T) {
	setupTestEnv(t, nil)

	// founder: real rows, nothing applied from a log yet
	if err := utils.CreateDevice(utils.ConstellationDevice{DeviceName: "founder", IP: "192.168.201.1"}); err != nil {
		t.Fatal("CreateDevice:", err)
	}
	if utils.IsOplogBootstrapped() {
		t.Fatal("a node that has never materialized from the log must not read as bootstrapped")
	}

	// founder-seed path marks it without any sequence moving
	if err := utils.MarkOplogBootstrapped(utils.GetOplogEpoch()); err != nil {
		t.Fatal("MarkOplogBootstrapped:", err)
	}
	if !utils.IsOplogBootstrapped() {
		t.Fatal("founder-seed did not mark the node bootstrapped")
	}
	if seq := utils.GetLastAppliedSeq(); seq != 0 {
		t.Fatalf("founder should still be at seq 0, got %d", seq)
	}

	// a joiner installing that founder's (epoch 1, seq 0) snapshot must come out
	// bootstrapped, or it snapshots forever
	dump, epoch, seq, err := utils.BuildSnapshot()
	if err != nil {
		t.Fatal("BuildSnapshot:", err)
	}
	if seq != 0 {
		t.Fatalf("test premise wrong: founder snapshot should be stamped seq 0, got %d", seq)
	}

	setupTestEnv(t, nil) // fresh node standing in for the joiner
	if utils.IsOplogBootstrapped() {
		t.Fatal("fresh joiner must not start bootstrapped")
	}
	if err := utils.ApplyLogicalDump(dump, epoch, seq); err != nil {
		t.Fatal("ApplyLogicalDump:", err)
	}
	if !utils.IsOplogBootstrapped() {
		t.Fatal("joiner still un-bootstrapped after installing a seq-0 snapshot — it would re-snapshot forever")
	}
	if _, err := utils.GetDeviceByName("founder", false); err != nil {
		t.Fatal("joiner did not receive the founder's rows:", err)
	}

	// a reform bumping the epoch must invalidate the marker without a clear path
	if err := utils.SetOplogState(2, 0); err != nil {
		t.Fatal("SetOplogState:", err)
	}
	if utils.IsOplogBootstrapped() {
		t.Fatal("bootstrapped marker survived an epoch change")
	}
}

// Bug B: an enrolled node sitting at seq 0 (it joined via a founder's seq-0
// snapshot) fell past the read-only guard into a direct write during its own
// post-install mesh bounce. The write returned 200 having never been published.
func TestUnitOplogEnrolledAtSeqZeroIsReadOnly(t *testing.T) {
	setupTestEnv(t, nil)
	writeNebulaYML(t, map[string]interface{}{
		"cstln_device_name": "self", "cstln_ip": "192.168.201.1/24", "cstln_api_key": "k",
	})

	// exactly the joiner's state: bootstrapped from a seq-0 snapshot, mid-bounce
	if err := utils.MarkOplogBootstrapped(utils.GetOplogEpoch()); err != nil {
		t.Fatal("MarkOplogBootstrapped:", err)
	}
	if seq := utils.GetLastAppliedSeq(); seq != 0 {
		t.Fatalf("test premise wrong: want seq 0, got %d", seq)
	}
	oplogAttached.Store(false)
	NebulaStarted.Store(false)

	if got := oplogWriteMode(); got != oplogReadOnly {
		t.Fatalf("enrolled node at seq 0: mode %v, want read-only — a direct write here never publishes", got)
	}

	err := publishOp(utils.Mutation{
		Table: "devices", Op: "insert",
		Doc: utils.ConstellationDevice{DeviceName: "from-agent", IP: "192.168.201.91"},
	})
	if !errors.Is(err, utils.ErrReadOnly) {
		t.Fatalf("write returned %v, want ErrReadOnly", err)
	}
	if count, _ := utils.CountDevices(map[string]interface{}{}); count != 0 {
		t.Fatalf("write landed locally without publishing: %d devices", count)
	}
}

// The snapshot install is replace-all, so "is this snapshot empty" has to be a
// content question. An empty store marshals to 25 bytes, not zero, which made the
// original length check accept exactly the payload it existed to reject.
func TestUnitOplogEmptySnapshotRejected(t *testing.T) {
	setupTestEnv(t, nil)

	empty, err := utils.BuildLogicalDump()
	if err != nil {
		t.Fatal("BuildLogicalDump:", err)
	}
	if len(empty) == 0 {
		t.Fatal("test premise wrong: an empty dump should still be non-empty bytes")
	}
	if !dumpIsEmpty(empty) {
		t.Fatalf("empty store not detected as empty: %s (%d bytes)", empty, len(empty))
	}

	// a dump with rows must not be mistaken for empty
	if err := utils.CreateDevice(utils.ConstellationDevice{DeviceName: "d1", IP: "192.168.201.5"}); err != nil {
		t.Fatal("CreateDevice:", err)
	}
	full, err := utils.BuildLogicalDump()
	if err != nil {
		t.Fatal("BuildLogicalDump 2:", err)
	}
	if dumpIsEmpty(full) {
		t.Fatalf("populated store reported empty: %s", full)
	}

	// users alone are enough to make it non-empty
	if err := utils.DeleteDevicesLocal(map[string]interface{}{}); err != nil {
		t.Fatal("DeleteDevicesLocal:", err)
	}
	if err := utils.CreateUser(utils.User{Nickname: "solo"}); err != nil {
		t.Fatal("CreateUser:", err)
	}
	usersOnly, _ := utils.BuildLogicalDump()
	if dumpIsEmpty(usersOnly) {
		t.Fatalf("store with users reported empty: %s", usersOnly)
	}

	// garbage is treated as empty rather than installed
	if !dumpIsEmpty([]byte("not json")) {
		t.Fatal("undecodable dump not treated as empty")
	}
}

// ResetNebula wipes devices while the apply loop is still attached (it calls only
// stop(), which does not detach). Published, that delete-all carries an empty
// filter and becomes an unqualified DELETE on every node in the constellation.
func TestUnitOplogLocalWriteNeverPublishes(t *testing.T) {
	setupTestEnv(t, nil)
	writeNebulaYML(t, map[string]interface{}{
		"cstln_device_name": "self", "cstln_ip": "192.168.201.1/24", "cstln_api_key": "k",
	})

	if err := utils.CreateDevice(utils.ConstellationDevice{DeviceName: "d1", IP: "192.168.201.5"}); err != nil {
		t.Fatal("CreateDevice:", err)
	}

	// the state ResetNebula actually runs in: attached, so the ladder would publish
	oplogAttached.Store(true)
	defer oplogAttached.Store(false)

	if got := oplogWriteMode(); got != oplogPublish {
		t.Fatalf("test premise wrong: mode %v, want publish", got)
	}

	// the local path must ignore the ladder entirely — no publish, no waiter, no
	// error from having no JetStream context
	if err := utils.DeleteDevicesLocal(map[string]interface{}{}); err != nil {
		t.Fatal("DeleteDevicesLocal must work while attached:", err)
	}
	if count, _ := utils.CountDevices(map[string]interface{}{}); count != 0 {
		t.Fatalf("local delete did not apply: %d devices", count)
	}

	// and it must not have moved the log position
	if seq := utils.GetLastAppliedSeq(); seq != 0 {
		t.Fatalf("local write moved the log position to %d", seq)
	}
}

// Formation: an HA constellation has no JetStream until its second manager
// enrolls, so the creator has to write straight to SQLite in the meantime.
func TestUnitOplogFormationWriterWritesDirect(t *testing.T) {
	setupTestEnv(t, nil)
	writeNebulaYML(t, map[string]interface{}{
		"cstln_device_name": "self", "cstln_ip": "192.168.201.1/24", "cstln_api_key": "k",
	})
	deviceCacheMux.Lock()
	CachedDevices = map[string]utils.ConstellationDevice{
		"peer": {DeviceName: "peer", IP: "192.168.201.2/24", CosmosNode: 2},
	}
	deviceCacheMux.Unlock()
	NebulaStarted.Store(true)
	oplogAttached.Store(false)

	// the same node without the licence is read-only: this is what every peer is
	// during formation, and it is what makes the writer a single writer
	if got := oplogWriteMode(); got != oplogReadOnly {
		t.Fatalf("unlicensed node: mode %v, want read-only", got)
	}

	if err := utils.SetFormationWriter(utils.GetOplogEpoch()); err != nil {
		t.Fatal("SetFormationWriter:", err)
	}
	if got := oplogWriteMode(); got != oplogDirect {
		t.Fatalf("formation writer: mode %v, want direct", got)
	}

	if err := publishOp(utils.Mutation{
		Table: "users", Op: "insert", Doc: utils.User{Nickname: "founder", Password: "h"},
	}); err != nil {
		t.Fatal("formation write:", err)
	}
	if count, _ := utils.CountUsers(); count != 1 {
		t.Fatalf("formation write did not land locally: %d users", count)
	}

	// The licence ends when the log exists, NOT when this node attaches to it: a
	// direct write in that gap would never reach the log and would be erased by
	// the first snapshot anyone takes.
	oplogStreamSeen.Store(true)
	if got := oplogWriteMode(); got != oplogReadOnly {
		t.Fatalf("formation writer after stream creation: mode %v, want read-only", got)
	}
}

// The licence carries the epoch it was granted for, so a reform elsewhere
// invalidates it with no explicit clear path — the same reason the bootstrapped
// marker is epoch-tied.
func TestUnitOplogFormationLicenceIsEpochTied(t *testing.T) {
	setupTestEnv(t, nil)

	if utils.IsFormationWriter() {
		t.Fatal("a fresh node holds a formation licence")
	}
	if err := utils.SetFormationWriter(utils.GetOplogEpoch()); err != nil {
		t.Fatal("SetFormationWriter:", err)
	}
	if !utils.IsFormationWriter() {
		t.Fatal("licence not granted")
	}

	// what installing a post-reform snapshot does to this node
	if err := utils.SetOplogState(utils.GetOplogEpoch()+1, 0); err != nil {
		t.Fatal("SetOplogState:", err)
	}
	if utils.IsFormationWriter() {
		t.Fatal("a stale licence survived the epoch bump — this node could write outside the new log")
	}
}

func TestUnitOplogReformEntersFormationAtNewEpoch(t *testing.T) {
	setupTestEnv(t, nil)

	before := utils.GetOplogEpoch()
	if err := utils.MarkOplogBootstrapped(before); err != nil {
		t.Fatal("MarkOplogBootstrapped:", err)
	}
	if err := utils.CommitOplogSeq(42); err != nil {
		t.Fatal("CommitOplogSeq:", err)
	}

	epoch, err := utils.ReformOplogEpoch()
	if err != nil {
		t.Fatal("ReformOplogEpoch:", err)
	}

	if epoch != before+1 || utils.GetOplogEpoch() != before+1 {
		t.Fatalf("epoch = %d (kv %d), want %d", epoch, utils.GetOplogEpoch(), before+1)
	}
	// the old cluster's positions mean nothing in the new log
	if seq := utils.GetLastAppliedSeq(); seq != 0 {
		t.Fatalf("last_applied_seq = %d, want 0", seq)
	}
	if utils.IsOplogBootstrapped() {
		t.Fatal("reform left the node claiming materialized state for an epoch it has not seen")
	}
	if !utils.IsFormationWriter() {
		t.Fatal("survivor did not take the formation licence, so it stays read-only with no way back")
	}
}

func TestUnitOplogReplicaTargetTracksManagerCount(t *testing.T) {
	setupTestEnv(t, func(c *utils.Config) { c.ConstellationConfig.NATSReplicas = 3 })

	setCache := func(devices ...utils.ConstellationDevice) {
		cache := map[string]utils.ConstellationDevice{}
		for _, d := range devices {
			cache[d.DeviceName] = d
			if d.PublicHostname != "" {
				cache[d.PublicHostname] = d
			}
		}
		deviceCacheMux.Lock()
		CachedDevices = cache
		deviceCacheMux.Unlock()
	}

	// the cache indexes a device under its name AND its public hostnames, and an
	// agent is not a cluster member — neither may inflate the count
	setCache(
		utils.ConstellationDevice{DeviceName: "a", CosmosNode: 2, PublicHostname: "a.example.com"},
		utils.ConstellationDevice{DeviceName: "b", CosmosNode: 2},
		utils.ConstellationDevice{DeviceName: "ag", CosmosNode: 1},
	)
	if got := countKnownManagers(); got != 2 {
		t.Fatalf("countKnownManagers() = %d, want 2", got)
	}
	if got := oplogTargetReplicas(); got != 1 {
		t.Fatalf("two managers: target %d replicas, want 1 (a 3-way group cannot be placed)", got)
	}

	setCache(
		utils.ConstellationDevice{DeviceName: "a", CosmosNode: 2},
		utils.ConstellationDevice{DeviceName: "b", CosmosNode: 2},
		utils.ConstellationDevice{DeviceName: "c", CosmosNode: 2},
		utils.ConstellationDevice{DeviceName: "gone", CosmosNode: 2, Blocked: true},
	)
	if got := countKnownManagers(); got != 3 {
		t.Fatalf("countKnownManagers() = %d, want 3 (blocked managers must not count)", got)
	}
	if got := oplogTargetReplicas(); got != 3 {
		t.Fatalf("three managers: target %d replicas, want 3", got)
	}

	// a non-HA constellation never scales up, however many nodes it grows
	config := utils.ReadConfigFromFile()
	config.ConstellationConfig.NATSReplicas = 1
	utils.SetBaseMainConfig(config)
	if got := oplogTargetReplicas(); got != 1 {
		t.Fatalf("non-HA: target %d replicas, want 1", got)
	}
}

// Reform's manager-blocking writes are the leave/teardown class: this node has
// no quorum to publish through, which is the whole reason it is reforming.
func TestUnitForceReformBlocksOtherManagersLocally(t *testing.T) {
	setupTestEnv(t, nil)

	for _, d := range []utils.ConstellationDevice{
		{DeviceName: "self", IP: "192.168.201.1", CosmosNode: 2},
		{DeviceName: "dead1", IP: "192.168.201.2", CosmosNode: 2},
		{DeviceName: "dead2", IP: "192.168.201.3", CosmosNode: 2},
		{DeviceName: "agent1", IP: "192.168.201.4", CosmosNode: 1},
	} {
		if err := utils.CommitMutationLocal(utils.Mutation{Table: "devices", Op: "insert", Doc: d}); err != nil {
			t.Fatal("seed device:", err)
		}
	}

	// attached, so anything on the publish ladder would try to publish
	oplogAttached.Store(true)
	defer oplogAttached.Store(false)

	blocked, err := blockDeadManagers("self")
	if err != nil {
		t.Fatal("blockDeadManagers:", err)
	}
	sort.Strings(blocked)
	if strings.Join(blocked, ",") != "dead1,dead2" {
		t.Fatalf("blocked %v, want [dead1 dead2]", blocked)
	}

	// the survivor and the agents are untouched: reform fences the lost managers,
	// it does not tear the constellation down
	active, err := utils.ListDevices(false)
	if err != nil {
		t.Fatal("ListDevices:", err)
	}
	names := []string{}
	for _, d := range active {
		names = append(names, d.DeviceName)
	}
	sort.Strings(names)
	if strings.Join(names, ",") != "agent1,self" {
		t.Fatalf("surviving devices %v, want [agent1 self]", names)
	}

	// local writes, so the log position must not have moved
	if seq := utils.GetLastAppliedSeq(); seq != 0 {
		t.Fatalf("reform block moved the log position to %d", seq)
	}
}

// withFBL gives the package a non-nil licence object for the duration of a test.
// utils.FBL is nil until InitFBL runs, and the creation policy reads AgentMode
// off it; the E2E harness does the same with an unreachable URL.
func withFBL(t *testing.T, agentMode bool) {
	t.Helper()
	prev := utils.FBL
	utils.FBL = utils.NewFirebaseApiSdk("http://127.0.0.1:9")
	utils.FBL.AgentMode = agentMode
	t.Cleanup(func() { utils.FBL = prev })
}

func mayCreate(waitingSince time.Time) bool {
	return oplogMayCreateStream("cosmos-oplog-e1", waitingSince) == nil
}

// An install upgrading from a release with no op-log holds NONE of the three kv
// markers, and nothing will ever grant it a formation licence. Creation must
// therefore fall back to the designated-creator election rather than refuse
// forever: a permanent refusal leaves every node in the constellation read-only
// with no operator action that recovers it.
//
// The fallback window is real and deliberate — such an install is read-only for
// up to kvCreatorFallbackAfter before its first log exists. That is the price of
// not letting an un-materialized node create a stream while a peer that holds
// the real state might still answer, and it is paid once per install.
func TestUnitOplogUpgradedInstallCanStillCreateItsFirstStream(t *testing.T) {
	setupTestEnv(t, nil)
	withFBL(t, false)
	writeNebulaYML(t, map[string]interface{}{
		"cstln_device_name": "aaa", "cstln_ip": "192.168.201.1/24", "cstln_api_key": "k",
	})
	deviceCacheMux.Lock()
	CachedDevices = map[string]utils.ConstellationDevice{
		"aaa": {DeviceName: "aaa", IP: "192.168.201.1/24", CosmosNode: 2},
		"zzz": {DeviceName: "zzz", IP: "192.168.201.2/24", CosmosNode: 2},
	}
	deviceCacheMux.Unlock()

	// the exact state of an upgraded install: no licence, never materialized
	if utils.IsFormationWriter() || utils.IsOplogBootstrapped() {
		t.Fatal("test premise wrong: node should hold neither marker")
	}

	// this node IS the designated creator ("aaa" sorts lowest), so it must not be
	// held back — it is the one that has to seed the log
	if _, isSelf := designatedKVCreator(); !isSelf {
		t.Fatal("test premise wrong: expected to be the designated creator")
	}

	// deferred at first — a peer holding materialized state may still answer
	if mayCreate(time.Now()) {
		t.Fatal("an un-materialized node created the log before giving peers a chance to answer")
	}

	// but it MUST get through, or the upgrade bricks: no formation writer exists
	// on this install and no later event creates one
	if !mayCreate(time.Now().Add(-kvCreatorFallbackAfter - time.Second)) {
		t.Fatal("the designated creator of an upgraded install is refused even after the " +
			"fallback window, so no node ever creates the log and the constellation " +
			"stays read-only permanently")
	}
}

// The guards in createOplogStream must DEFER, never veto: each one has to time
// out, or an install with no formation writer anywhere waits forever.
func TestUnitOplogStreamCreationGuardsAreDeferrals(t *testing.T) {
	setupTestEnv(t, nil)
	withFBL(t, false)
	writeNebulaYML(t, map[string]interface{}{
		"cstln_device_name": "zzz", "cstln_ip": "192.168.201.2/24", "cstln_api_key": "k",
	})
	deviceCacheMux.Lock()
	CachedDevices = map[string]utils.ConstellationDevice{
		"aaa": {DeviceName: "aaa", IP: "192.168.201.1/24", CosmosNode: 2},
		"zzz": {DeviceName: "zzz", IP: "192.168.201.2/24", CosmosNode: 2},
	}
	deviceCacheMux.Unlock()

	// not the designated creator, and nothing materialized: defers while fresh
	if mayCreate(time.Now()) {
		t.Fatal("a non-designated, unmaterialized node created the log immediately")
	}

	// ...but the same node must get through once the fallback window has passed
	if !mayCreate(time.Now().Add(-kvCreatorFallbackAfter - time.Second)) {
		t.Fatal("the guard never times out, so a constellation whose designated creator " +
			"is gone would never get a log at all")
	}
}

// The split-brain guard. A node that has materialized from a log must NEVER
// create one, however long it waits and whatever the election says.
//
// The chain this closes: after a reform takes the cluster to epoch 3, two
// revived managers still holding epoch 2 have quorum of the old meta group, so
// they attach to their own stream and answer snapshot requests. A replacement
// manager asks, loses the reply race to one of them, and installs epoch 2 —
// becoming "bootstrapped at epoch 2". If being bootstrapped let it create, it
// would then create cosmos-oplog-e2 on the LIVE cluster, the publish fence would
// be breached, and both streams would accept writes with neither side wrong.
func TestUnitOplogMaterializedNodeNeverRecreatesAnAbandonedEpoch(t *testing.T) {
	setupTestEnv(t, nil)
	withFBL(t, false)
	writeNebulaYML(t, map[string]interface{}{
		"cstln_device_name": "aaa", "cstln_ip": "192.168.201.1/24", "cstln_api_key": "k",
	})
	deviceCacheMux.Lock()
	CachedDevices = map[string]utils.ConstellationDevice{
		"aaa": {DeviceName: "aaa", IP: "192.168.201.1/24", CosmosNode: 2},
		"zzz": {DeviceName: "zzz", IP: "192.168.201.2/24", CosmosNode: 2},
	}
	deviceCacheMux.Unlock()

	// it adopted a peer's epoch, exactly as installOplogSnapshot leaves a node
	if err := utils.SetOplogState(2, 0); err != nil {
		t.Fatal("SetOplogState:", err)
	}
	if err := utils.MarkOplogBootstrapped(2); err != nil {
		t.Fatal("MarkOplogBootstrapped:", err)
	}

	// it is the designated creator AND has waited out the fallback — every
	// condition that would otherwise admit it
	if _, isSelf := designatedKVCreator(); !isSelf {
		t.Fatal("test premise wrong: expected to be the designated creator")
	}
	if mayCreate(time.Now().Add(-kvCreatorFallbackAfter - time.Second)) {
		t.Fatal("a node that materialized from a log recreated its epoch's stream — " +
			"this resurrects an abandoned epoch on the live cluster and breaches the publish fence")
	}

	// ...but the formation licence still admits it, because that is what a
	// force-reform grants and reform must be able to seed its new epoch
	if err := utils.SetFormationWriter(utils.GetOplogEpoch()); err != nil {
		t.Fatal("SetFormationWriter:", err)
	}
	if !mayCreate(time.Now()) {
		t.Fatal("the formation licence no longer admits creation, so a reform could never seed its epoch")
	}
}

// An agent never creates the log, however long it waits.
func TestUnitOplogAgentNeverCreatesTheLog(t *testing.T) {
	setupTestEnv(t, nil)
	withFBL(t, true)

	if mayCreate(time.Now().Add(-kvCreatorFallbackAfter - time.Second)) {
		t.Fatal("an agent created the op-log stream")
	}
}

// Reform fences by RENAMING the log, so a node left behind at the old epoch must
// not be able to keep dialing the managers that were blocked — those IPs become
// the cluster route list, and routes must never point at non-voters.
func TestUnitManagerIPsExcludeBlockedDevices(t *testing.T) {
	setupTestEnv(t, nil)

	for _, d := range []utils.ConstellationDevice{
		{DeviceName: "self", IP: "192.168.201.1", CosmosNode: 2},
		{DeviceName: "dead", IP: "192.168.201.2", CosmosNode: 2, Blocked: true},
		{DeviceName: "live", IP: "192.168.201.3", CosmosNode: 2},
	} {
		if err := utils.CommitMutationLocal(utils.Mutation{Table: "devices", Op: "insert", Doc: d}); err != nil {
			t.Fatal("seed device:", err)
		}
	}

	// the cache holds only live managers (refreshDeviceCache drops blocked ones),
	// so the blocked IP can only arrive the way it does in production: as a frozen
	// enrollment seed in nebula.yml that nothing ever rewrites
	deviceCacheMux.Lock()
	CachedDevices = map[string]utils.ConstellationDevice{
		"live": {DeviceName: "live", IP: "192.168.201.3", CosmosNode: 2},
	}
	deviceCacheMux.Unlock()
	writeNebulaYML(t, map[string]interface{}{
		"cstln_device_name":   "self",
		"cstln_ip":            "192.168.201.1/24",
		"cstln_api_key":       "k",
		"cstln_nats_managers": []interface{}{"192.168.201.2", "192.168.201.3"},
	})

	got := getManagerIPs("192.168.201.1")
	if strings.Join(got, ",") != "192.168.201.3" {
		t.Fatalf("getManagerIPs = %v, want only the live manager — a blocked manager "+
			"reachable through a stale enrollment seed means the reform isolated nothing", got)
	}
}

// An IP is not retired along with its device. GetNextAvailableIP allocates out of
// the UNBLOCKED devices, so the replacement manager enrolled after a reform
// normally takes the dead manager's address back — and the schema allows the
// overlap deliberately (UNIQUE(ip) WHERE blocked=0). Subtracting blocked IPs
// blindly would therefore drop the LIVE replacement out of the cluster route
// list and JetStream would never elect, turning the reform into the outage it
// was meant to repair.
func TestUnitManagerIPsKeepsAnIPReusedByALiveDevice(t *testing.T) {
	setupTestEnv(t, nil)

	for _, d := range []utils.ConstellationDevice{
		{DeviceName: "self", IP: "192.168.201.1", CosmosNode: 2},
		{DeviceName: "dead", IP: "192.168.201.2", CosmosNode: 2, Blocked: true},
		// the replacement, enrolled at the address the blocked manager used to hold
		{DeviceName: "replacement", IP: "192.168.201.2", CosmosNode: 2},
	} {
		if err := utils.CommitMutationLocal(utils.Mutation{Table: "devices", Op: "insert", Doc: d}); err != nil {
			t.Fatal("seed device:", err)
		}
	}

	deviceCacheMux.Lock()
	CachedDevices = map[string]utils.ConstellationDevice{
		"replacement": {DeviceName: "replacement", IP: "192.168.201.2", CosmosNode: 2},
	}
	deviceCacheMux.Unlock()
	writeNebulaYML(t, map[string]interface{}{
		"cstln_device_name": "self", "cstln_ip": "192.168.201.1/24", "cstln_api_key": "k",
	})

	got := getManagerIPs("192.168.201.1")
	if strings.Join(got, ",") != "192.168.201.2" {
		t.Fatalf("getManagerIPs = %v, want the replacement at 192.168.201.2 — dropping an "+
			"address a live manager holds means the reformed cluster can never elect", got)
	}
}

// The snapshot exchange is fenced on both sides. A responder behind the asker
// must stay silent: core request/reply takes the FIRST reply and gives up, so a
// fast stale answer starves the peers that could actually help, and the asker's
// next move — creating a stream for its own epoch — is how an abandoned epoch's
// log gets resurrected on a live cluster.
func TestUnitOplogSnapshotRequestCarriesEpoch(t *testing.T) {
	setupTestEnv(t, nil)

	// what requestOplogSnapshot puts on the wire
	if err := utils.SetOplogState(7, 3); err != nil {
		t.Fatal("SetOplogState:", err)
	}
	body, err := json.Marshal(oplogSnapshotRequest{Epoch: utils.GetOplogEpoch()})
	if err != nil {
		t.Fatal("marshal request:", err)
	}

	var req oplogSnapshotRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatal("unmarshal request:", err)
	}
	if req.Epoch != 7 {
		t.Fatalf("request carried epoch %d, want 7", req.Epoch)
	}

	// the responder's rule: decline only when the asker is ahead of us
	declines := func(theirs, ours uint64) bool { return theirs > ours }
	if !declines(req.Epoch, 6) {
		t.Fatal("a node behind the asker answered anyway, starving the asker's one reply")
	}
	if declines(req.Epoch, 7) || declines(req.Epoch, 8) {
		t.Fatal("a node at or ahead of the asker declined, so nobody would ever serve it")
	}

	// a legacy or unreadable body decodes to epoch 0, which every node still serves
	var legacy oplogSnapshotRequest
	json.Unmarshal([]byte("{}"), &legacy)
	if declines(legacy.Epoch, 1) {
		t.Fatal("a legacy empty request body was declined")
	}
}

// Reproduces the E2E finding that a force-reform fences nothing: a blocked
// manager revived, rejoined the cluster, adopted the new epoch and wrote, and
// its writes landed on the survivor. Blocking revokes CLIENT credentials, but a
// manager's client dials its own server and cluster routes carry no auth at all.
//
// The node self-fences instead, off the very state it just adopted: the snapshot
// contains its own device record marked blocked.
func TestUnitOplogBlockedNodeCannotWrite(t *testing.T) {
	setupTestEnv(t, func(c *utils.Config) { c.ConstellationConfig.ThisDeviceName = "mantwo" })
	writeNebulaYML(t, map[string]interface{}{
		"cstln_device_name": "mantwo", "cstln_ip": "192.168.201.2/24", "cstln_api_key": "k",
	})

	if err := utils.CommitMutationLocal(utils.Mutation{Table: "devices", Op: "insert",
		Doc: utils.ConstellationDevice{DeviceName: "mantwo", IP: "192.168.201.2", CosmosNode: 2}}); err != nil {
		t.Fatal("seed device:", err)
	}

	// healthy and attached: the state in which a revived manager is most dangerous,
	// because every other check says it may publish
	oplogAttached.Store(true)
	defer oplogAttached.Store(false)
	if got := oplogWriteMode(); got != oplogPublish {
		t.Fatalf("test premise wrong: mode %v, want publish", got)
	}

	// the reform's block arrives, exactly as an adopted snapshot delivers it
	if err := utils.CommitMutationLocal(utils.Mutation{
		Table:  "devices",
		Op:     "updateMany",
		Filter: map[string]interface{}{"DeviceName": "mantwo"},
		Doc:    map[string]interface{}{"Blocked": true},
	}); err != nil {
		t.Fatal("block self:", err)
	}

	if !isSelfBlocked() {
		t.Fatal("node cannot see its own block")
	}
	if got := oplogWriteMode(); got != oplogReadOnly {
		t.Fatalf("blocked node: mode %v, want read-only — it would rejoin and write", got)
	}

	err := publishOp(utils.Mutation{Table: "users", Op: "insert", Doc: utils.User{Nickname: "stale-write"}})
	if !errors.Is(err, utils.ErrReadOnly) {
		t.Fatalf("blocked node's write returned %v, want ErrReadOnly", err)
	}
	if count, _ := utils.CountUsers(); count != 0 {
		t.Fatalf("blocked node wrote locally anyway: %d users", count)
	}
}

// "Read-only member" must also mean "not a source of truth for others".
//
// Self-fencing deliberately leaves a blocked node ATTACHED so it keeps
// materializing, which means it sails past the snapshot router's attached gate
// unless refused explicitly. That matters more than the write fence: a snapshot
// carries every registered domain, including the CA private key, the auth
// signing keypair, api_tokens, openid client secrets and rclone credentials — so
// a removed machine answering here hands the cluster's secret material to every
// node that enrolls afterwards.
func TestUnitOplogBlockedNodeDoesNotServeSnapshots(t *testing.T) {
	setupTestEnv(t, func(c *utils.Config) { c.ConstellationConfig.ThisDeviceName = "removed" })
	writeNebulaYML(t, map[string]interface{}{
		"cstln_device_name": "removed", "cstln_ip": "192.168.201.2/24", "cstln_api_key": "k",
	})

	if err := utils.CommitMutationLocal(utils.Mutation{Table: "devices", Op: "insert",
		Doc: utils.ConstellationDevice{DeviceName: "removed", IP: "192.168.201.2", CosmosNode: 2}}); err != nil {
		t.Fatal("seed device:", err)
	}

	// attached and healthy — every other gate in the router says "serve"
	oplogAttached.Store(true)
	defer oplogAttached.Store(false)
	oplogHalted.Store(false)

	if serves := oplogMayServeSnapshot(); !serves {
		t.Fatal("test premise wrong: an attached healthy node should serve")
	}

	if err := utils.CommitMutationLocal(utils.Mutation{
		Table:  "devices",
		Op:     "updateMany",
		Filter: map[string]interface{}{"DeviceName": "removed"},
		Doc:    map[string]interface{}{"Blocked": true},
	}); err != nil {
		t.Fatal("block self:", err)
	}

	if serves := oplogMayServeSnapshot(); serves {
		t.Fatal("a removed node still serves snapshots — it would hand the CA key and " +
			"every other secret to each node that enrolls after it was fenced")
	}
}

// The fence fails OPEN everywhere it is unsure. A false positive takes a healthy
// node read-only over a transient DB problem; a false negative is only the
// status quo. An ABSENT record in particular is a node mid-enrollment, not a
// removed one.
func TestUnitOplogSelfBlockFailsOpen(t *testing.T) {
	setupTestEnv(t, func(c *utils.Config) { c.ConstellationConfig.ThisDeviceName = "solo" })

	// no device record at all yet
	if isSelfBlocked() {
		t.Fatal("a node with no device record read as blocked — enrollment would never complete")
	}

	// present and active
	if err := utils.CommitMutationLocal(utils.Mutation{Table: "devices", Op: "insert",
		Doc: utils.ConstellationDevice{DeviceName: "solo", IP: "192.168.201.9", CosmosNode: 2}}); err != nil {
		t.Fatal("seed device:", err)
	}
	if isSelfBlocked() {
		t.Fatal("an active device read as blocked")
	}

	// constellation disabled: GetCurrentDeviceName errors, and that must not fence
	config := utils.ReadConfigFromFile()
	config.ConstellationConfig.Enabled = false
	utils.SetBaseMainConfig(config)
	if isSelfBlocked() {
		t.Fatal("an unidentifiable node read as blocked")
	}
}

// Reproduces the E2E finding that a node healthy at the moment of a reform is
// the one node that can never recover. oplogAttached is a latch, and a
// subscription whose stream is deleted goes quiet rather than erroring — so
// without a liveness probe the supervisor keeps taking the attached branch and
// oplogAttach, which holds every recovery path including epoch adoption, is
// never called again.
func TestUnitOplogAttachedNodeNoticesEpochMoved(t *testing.T) {
	setupTestEnv(t, nil)

	oplogAttached.Store(true)
	oplogAttachedEpoch.Store(1)
	defer oplogAttached.Store(false)

	// same epoch, no JetStream context: nothing to detach on
	oplogCheckAttachedStream()
	if !oplogAttached.Load() {
		t.Fatal("detached while still attached to the current epoch")
	}

	// a reform moves this node's epoch out from under a live subscription
	if err := utils.SetOplogState(2, 0); err != nil {
		t.Fatal("SetOplogState:", err)
	}
	oplogStreamCheckMu.Lock()
	oplogStreamCheckedAt = time.Time{} // the probe is paced; this is the next tick
	oplogStreamCheckMu.Unlock()

	oplogCheckAttachedStream()
	if oplogAttached.Load() {
		t.Fatal("stayed attached after its epoch moved — the supervisor would never " +
			"call oplogAttach again and the node could never adopt the new epoch")
	}
}

// A consumer can die WITHOUT producing an error anywhere: an ordered consumer
// lost during a quorum outage leaves the stream healthy, the epoch unchanged and
// StreamInfo answering normally, so every error-based check passes while nothing
// is ever delivered again. Observed on a survivor that stayed at seq 1 while its
// peers advanced, publishing into a log it would never read back — 503
// apply-timeout forever, reporting itself attached and writable throughout.
//
// So the detection keys on progress instead. This is the third bug in the same
// family: #1 was "the stream went away", #2 "the epoch moved", this is "the
// stream stayed and the consumer didn't".
func TestUnitOplogStalledConsumerIsDetected(t *testing.T) {
	setupTestEnv(t, nil)

	// caught up: nothing to detect however many times we look
	if err := utils.CommitOplogSeq(5); err != nil {
		t.Fatal("CommitOplogSeq:", err)
	}
	for i := 0; i < oplogStallProbes+2; i++ {
		if oplogNoteApplyProgress(5) {
			t.Fatal("a caught-up consumer was called stalled")
		}
	}

	// behind, but applying: every probe moves us on, so the count keeps resetting.
	// This is the case that must never trip — a healthy node working through a
	// backlog looks exactly like a stalled one for a single sample.
	for seq := uint64(6); seq <= 12; seq++ {
		if err := utils.CommitOplogSeq(seq); err != nil {
			t.Fatal("CommitOplogSeq:", err)
		}
		if oplogNoteApplyProgress(100) {
			t.Fatalf("a consumer that advanced to seq %d was called stalled", seq)
		}
	}

	// behind and frozen: the signature of a dead consumer
	stalled := false
	for i := 0; i < oplogStallProbes; i++ {
		stalled = oplogNoteApplyProgress(100)
	}
	if !stalled {
		t.Fatalf("a consumer stuck at seq %d while the log is at 100 was not detected after %d probes",
			utils.GetLastAppliedSeq(), oplogStallProbes)
	}

	// it must take the full window, not fire on the first frozen sample
	oplogResetStallTracking()
	if oplogNoteApplyProgress(100) {
		t.Fatal("declared stalled on a single probe — one lagging sample is not evidence")
	}

	// and a fresh subscription is never judged on the dead one's history
	for i := 0; i < oplogStallProbes-1; i++ {
		oplogNoteApplyProgress(100)
	}
	oplogResetStallTracking()
	if oplogNoteApplyProgress(100) {
		t.Fatal("a rebuilt consumer inherited the previous one's stall count")
	}
}

// nats.go reports an inactive consumer out of band long before three probes can
// prove it, so the async error handler arms the detector. It must stay a HINT:
// the probe still requires its own evidence, or a transient library report would
// tear down a healthy subscription.
func TestUnitOplogStallHintNeedsEvidence(t *testing.T) {
	setupTestEnv(t, nil)

	if err := utils.CommitOplogSeq(5); err != nil {
		t.Fatal("CommitOplogSeq:", err)
	}
	oplogResetStallTracking()

	// armed, but caught up: there is nothing being missed, so nothing happens
	oplogArmStallDetector()
	if oplogNoteApplyProgress(5) {
		t.Fatal("detached a caught-up consumer on a library hint alone")
	}

	// armed, behind, but still applying: progress overrides the hint
	oplogResetStallTracking()
	oplogArmStallDetector()
	if err := utils.CommitOplogSeq(6); err != nil {
		t.Fatal("CommitOplogSeq:", err)
	}
	if oplogNoteApplyProgress(100) {
		t.Fatal("detached a consumer that was still advancing")
	}

	// armed, behind AND frozen: now the hint pays off — one probe instead of three
	oplogResetStallTracking()
	oplogArmStallDetector()
	if !oplogNoteApplyProgress(100) {
		t.Fatal("hint did not shorten the window for a genuinely stalled consumer")
	}

	// and the hint never pushes the counter backwards past an in-progress stall
	oplogResetStallTracking()
	for i := 0; i < oplogStallProbes-1; i++ {
		oplogNoteApplyProgress(100)
	}
	oplogArmStallDetector()
	if !oplogNoteApplyProgress(100) {
		t.Fatal("arming reset progress already accumulated toward a stall")
	}
}

func TestUnitOplogDeviceReactionDiff(t *testing.T) {
	pre := []utils.ConstellationDevice{{
		DeviceName: "d1", IP: "192.168.201.5", IsLighthouse: true, Tags: []string{"old"},
	}}

	// resubmitting the same topology values is not a topology change: this is
	// what keeps a tag-only edit from bouncing the whole mesh
	sameTopology := map[string]interface{}{
		"IsLighthouse": true,
		"IsRelay":      false,
		"Tags":         []string{"new"},
	}
	if utils.DeviceFieldsChanged(pre, sameTopology, deviceGlobalFields) {
		t.Error("tag-only edit misread as a topology change")
	}

	if !utils.DeviceFieldsChanged(pre, map[string]interface{}{"IP": "192.168.201.6"}, deviceGlobalFields) {
		t.Error("IP change not detected")
	}
	if !utils.DeviceFieldsChanged(pre, map[string]interface{}{"Blocked": true}, deviceGlobalFields) {
		t.Error("block not detected")
	}
	if utils.DeviceFieldsChanged(pre, map[string]interface{}{"Nickname": "renamed"}, deviceGlobalFields) {
		t.Error("nickname counted as topology")
	}

	// LB toggles never restart anyone; exit-node toggles are self-only
	if utils.DeviceFieldsChanged(pre, map[string]interface{}{"IsLoadBalancer": true}, deviceGlobalFields) {
		t.Error("LB toggle misread as a global topology change")
	}
	if utils.DeviceFieldsChanged(pre, map[string]interface{}{"IsExitNode": true}, deviceGlobalFields) {
		t.Error("exit-node toggle misread as a global topology change")
	}
	if !utils.DeviceFieldsChanged(pre, map[string]interface{}{"IsExitNode": true}, deviceSelfRestartFields) {
		t.Error("exit-node change not detected as self-restart")
	}
}

func TestUnitStaggerRank(t *testing.T) {
	devices := []utils.ConstellationDevice{
		{DeviceName: "charlie", CosmosNode: 2},
		{DeviceName: "alice", CosmosNode: 1},
		{DeviceName: "phone", CosmosNode: 0}, // nebula-only client: holds no slot
		{DeviceName: "bob", CosmosNode: 2},
	}

	// ranks follow name order among cosmos servers only
	for name, want := range map[string]int{"alice": 0, "bob": 1, "charlie": 2} {
		if got := staggerRank(name, devices); got != want {
			t.Errorf("staggerRank(%s) = %d, want %d", name, got, want)
		}
	}

	// a node absent from the table restarts soonest
	if got := staggerRank("deleted-node", devices); got != 0 {
		t.Errorf("staggerRank(deleted-node) = %d, want 0", got)
	}
	if got := staggerRank("phone", devices); got != 0 {
		t.Errorf("staggerRank(phone) = %d, want 0", got)
	}
}
