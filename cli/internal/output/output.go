// Package output provides helpers for printing CLI results.
package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// JSON prints v as indented JSON to stdout.
func JSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// Table prints a simple aligned table to stdout.
func Table(headers []string, rows [][]string) {
	if len(rows) == 0 {
		fmt.Println("  (no results)")
		return
	}

	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	printRow := func(cols []string) {
		parts := make([]string, len(cols))
		for i, c := range cols {
			if i < len(widths) {
				parts[i] = fmt.Sprintf("%-*s", widths[i], c)
			} else {
				parts[i] = c
			}
		}
		fmt.Println("  " + strings.Join(parts, "  "))
	}

	printRow(headers)
	sep := make([]string, len(headers))
	for i, w := range widths {
		sep[i] = strings.Repeat("-", w)
	}
	printRow(sep)
	for _, row := range rows {
		printRow(row)
	}
}

// Success prints a success message.
func Success(format string, a ...any) {
	fmt.Printf("  ✓ "+format+"\n", a...)
}

// Errorf prints an error message to stderr.
func Errorf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "  ✗ "+format+"\n", a...)
}
