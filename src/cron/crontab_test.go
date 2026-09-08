package cron

import "testing"

func TestNormalizeCrontab(t *testing.T) {
	tests := []struct {
		in   string
		want string
		ok   bool
	}{
		// 6-field seconds-first (Cosmos UI format) stays as-is
		{"0 0 4 * * *", "0 0 4 * * *", true},
		{"0 0 2 * * *", "0 0 2 * * *", true},
		// 5-field standard gets a seconds column prepended, keeping semantics
		{"* * * * *", "0 * * * * *", true},
		{"0 2 * * *", "0 0 2 * * *", true},
		{"*/5 * * * *", "0 */5 * * * *", true},
		// Descriptors pass through
		{"@daily", "@daily", true},
		{"@every 10m", "@every 10m", true},
		// Timezone prefix preserved, 5-field normalized
		{"CRON_TZ=Europe/Paris 0 2 * * *", "CRON_TZ=Europe/Paris 0 0 2 * * *", true},
		{"TZ=UTC 0 2 * * *", "TZ=UTC 0 0 2 * * *", true},
		// Whitespace tidied
		{"  0  0  4  *  *  *  ", "0 0 4 * * *", true},
		// Invalid
		{"", "", false},
		{"0 2 * *", "", false},         // 4 fields
		{"abc def ghi jkl mno pqr", "", false},
		{"* * * * * * *", "", false},   // 7 fields
	}
	for _, tt := range tests {
		got, err := NormalizeCrontab(tt.in)
		if tt.ok && err != nil {
			t.Errorf("NormalizeCrontab(%q) unexpected error: %v", tt.in, err)
			continue
		}
		if tt.ok && got != tt.want {
			t.Errorf("NormalizeCrontab(%q) = %q, want %q", tt.in, got, tt.want)
		}
		if !tt.ok && err == nil {
			t.Errorf("NormalizeCrontab(%q) expected error, got %q", tt.in, got)
		}
	}
}

func TestValidCrontab(t *testing.T) {
	valid := []string{"* * * * *", "0 0 4 * * *", "*/15 2 * * 1-5", "@weekly", "45 23 * * 6", "0 0 4 */2 * *"}
	for _, c := range valid {
		if !ValidCrontab(c) {
			t.Errorf("ValidCrontab(%q) = false, want true", c)
		}
	}
	invalid := []string{"", "not a cron", "* * *", "61 2 * * *", "0 0 99 * * *"}
	for _, c := range invalid {
		if ValidCrontab(c) {
			t.Errorf("ValidCrontab(%q) = true, want false", c)
		}
	}
}
