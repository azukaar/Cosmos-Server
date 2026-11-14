package cron

import (
	"fmt"
	"strings"

	"github.com/robfig/cron/v3"
)

// Cosmos uses a DUAL-FORMAT crontab scheme:
//
//   - Standard 5-field crontabs:  "minute hour dom month dow"  (e.g. "0 2 * * *")
//   - Optional 6-field with seconds: "second minute hour dom month dow" (e.g. "0 0 2 * * *")
//   - Descriptors: @hourly, @daily, @weekly, @monthly, @yearly, @every 10m
//   - Timezone prefix: "CRON_TZ=Europe/Paris 0 2 * * *" (or "TZ=...")
//
// When a crontab has exactly 6 fields, it is interpreted as seconds-based
// (the Cosmos UI generates schedules this way). A plain 5-field crontab is
// interpreted with the STANDARD semantic (minutes first, as in system cron).
// This keeps compatibility with users who type standard 5-field expressions.
//
// All stored crontabs are NORMALIZED to the canonical 6-field form
// "<sec> <min> <hour> <dom> <month> <dow>" so the whole system (UI preview,
// API, scheduler) agrees on the same representation.

// crontab6Parser parses the 6-field seconds-based format.
var crontab6Parser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

// crontab5Parser parses the standard 5-field format.
var crontab5Parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

// parseCrontab parses a crontab in either the 5-field or 6-field format.
// Descriptors and TZ= prefixes are handled by the underlying robfig/cron parser.
func parseCrontab(spec string) (cron.Schedule, error) {
	trimmed := strings.TrimSpace(spec)
	if trimmed == "" {
		return nil, fmt.Errorf("crontab is empty")
	}

	_, body := splitTZPrefix(trimmed)

	// Descriptors (@daily, @every 1h, ...) are only handled by the 5-field parser.
	if strings.HasPrefix(body, "@") {
		return crontab5Parser.Parse(trimmed)
	}

	fields := strings.Fields(body)
	if len(fields) == 6 {
		return crontab6Parser.Parse(trimmed)
	}
	if len(fields) == 5 {
		return crontab5Parser.Parse(trimmed)
	}
	return nil, fmt.Errorf("crontab must have 5 or 6 fields, got %d", len(fields))
}

// ValidCrontab reports whether the given crontab is syntactically valid in
// either supported format.
func ValidCrontab(crontab string) bool {
	_, err := parseCrontab(crontab)
	return err == nil
}

// NormalizeCrontab converts a user-supplied crontab into the canonical
// 6-field seconds-first representation, so the UI/API/scheduler all share the
// same string. It returns an error for invalid crontabs.
func NormalizeCrontab(crontab string) (string, error) {
	if crontab == "" {
		return "", fmt.Errorf("crontab is empty")
	}

	trimmed := strings.TrimSpace(crontab)

	// Descriptors are already canonical, keep them as-is.
	if strings.HasPrefix(trimmed, "@") {
		return trimmed, nil
	}

	if _, err := parseCrontab(trimmed); err != nil {
		return "", err
	}

	prefix, rest := splitTZPrefix(trimmed)
	fields := strings.Fields(rest)

	if len(fields) == 5 {
		// Insert a "0" seconds column: "m h dom mon dow" -> "0 m h dom mon dow".
		if prefix != "" {
			return tidyCrontab(prefix + " 0 " + rest), nil
		}
		return tidyCrontab("0 " + trimmed), nil
	}

	// Already 6-field seconds-first: just tidy whitespace (keep TZ prefix).
	return tidyCrontab(trimmed), nil
}

// splitTZPrefix splits a crontab into an optional "TZ=..." / "CRON_TZ=..."
// prefix and the remaining schedule body.
func splitTZPrefix(crontab string) (string, string) {
	trimmed := strings.TrimSpace(crontab)
	if idx := strings.Index(trimmed, " "); idx != -1 {
		prefix := trimmed[:idx]
		if strings.HasPrefix(prefix, "TZ=") || strings.HasPrefix(prefix, "CRON_TZ=") {
			return prefix, strings.TrimSpace(trimmed[idx+1:])
		}
	}
	return "", trimmed
}

// tidyCrontab collapses repeated whitespace (preserving any leading TZ= prefix).
func tidyCrontab(crontab string) string {
	prefix, rest := splitTZPrefix(crontab)
	joined := strings.Join(strings.Fields(rest), " ")
	if prefix != "" {
		return prefix + " " + joined
	}
	return joined
}
