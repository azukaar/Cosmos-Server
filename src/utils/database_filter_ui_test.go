package utils

import (
	"encoding/json"
	"testing"
)

// Guards the exact filter strings the Constellation UI ships as EventExplorerStandalone initSearch values.
func TestUIShippedEventFilters(t *testing.T) {
	filters := []string{
		`{"eventId":{"$regex":"^cosmos\\.scheduler\\."}}`,
		`{"eventId":{"$regex":"^cosmos\\.database\\."}}`,
		`{"eventId":{"$regex":"^cosmos\\.seaweedfs\\."}}`,
		`{"eventId":{"$regex":"^cosmos\\.registry\\."}}`,
		`{"$or":[{"eventId":{"$regex":"^cosmos\\.constellation\\."}},{"object":{"$regex":"^device@"}}]}`,
	}

	for _, raw := range filters {
		var f map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &f); err != nil {
			t.Fatalf("%s: bad json: %v", raw, err)
		}
		sql, args, err := TranslateEventFilter(DialectSQLite, f)
		if err != nil {
			t.Fatalf("%s: %v", raw, err)
		}
		t.Logf("%s\n  -> %s %v", raw, sql, args)
	}
}
