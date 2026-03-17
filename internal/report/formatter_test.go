package report_test

import (
	"strings"
	"testing"

	"github.com/teerarak/mcp-postgresql-report/internal/db"
	"github.com/teerarak/mcp-postgresql-report/internal/report"
)

var sampleResult = &db.QueryResult{
	Columns: []string{"id", "name", "department"},
	Rows: [][]string{
		{"1", "Alice", "Engineering"},
		{"2", "Bob", "Marketing"},
		{"3", "Charlie", "Engineering"},
	},
}

func TestFormatTable(t *testing.T) {
	out, err := report.FormatResult(sampleResult, report.FormatTable)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must contain column headers
	for _, col := range sampleResult.Columns {
		if !strings.Contains(out, col) {
			t.Errorf("expected column %q in output, got:\n%s", col, out)
		}
	}

	// Must contain data
	for _, row := range sampleResult.Rows {
		for _, val := range row {
			if !strings.Contains(out, val) {
				t.Errorf("expected value %q in output, got:\n%s", val, out)
			}
		}
	}

	// Must have separator row
	if !strings.Contains(out, "---") {
		t.Errorf("expected separator row in table output:\n%s", out)
	}
}

func TestFormatCSV(t *testing.T) {
	out, err := report.FormatResult(sampleResult, report.FormatCSV)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	// Header + 3 data rows = 4 lines
	if len(lines) != 4 {
		t.Errorf("expected 4 lines (header+3 rows), got %d:\n%s", len(lines), out)
	}

	// First line is header
	if lines[0] != "id,name,department" {
		t.Errorf("expected header line 'id,name,department', got %q", lines[0])
	}

	// Check a data line
	if !strings.Contains(lines[1], "Alice") {
		t.Errorf("expected Alice in first data row, got %q", lines[1])
	}
}

func TestFormatJSON(t *testing.T) {
	out, err := report.FormatResult(sampleResult, report.FormatJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must be a JSON array
	if !strings.HasPrefix(strings.TrimSpace(out), "[") {
		t.Errorf("expected JSON array, got:\n%s", out)
	}

	// Must contain all column names as keys
	for _, col := range sampleResult.Columns {
		if !strings.Contains(out, `"`+col+`"`) {
			t.Errorf("expected JSON key %q in output:\n%s", col, out)
		}
	}

	// Must contain data values
	if !strings.Contains(out, "Alice") || !strings.Contains(out, "Marketing") {
		t.Errorf("expected data values in JSON output:\n%s", out)
	}
}

func TestFormatTableEmpty(t *testing.T) {
	empty := &db.QueryResult{
		Columns: []string{"id", "name"},
		Rows:    nil,
	}

	out, err := report.FormatResult(empty, report.FormatTable)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still have header
	if !strings.Contains(out, "id") || !strings.Contains(out, "name") {
		t.Errorf("expected header in empty table output:\n%s", out)
	}
}

func TestFormatCSVEmpty(t *testing.T) {
	empty := &db.QueryResult{
		Columns: []string{"id", "name"},
		Rows:    nil,
	}

	out, err := report.FormatResult(empty, report.FormatCSV)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 line (header only) for empty CSV, got %d", len(lines))
	}
}

func TestFormatJSONEmpty(t *testing.T) {
	empty := &db.QueryResult{
		Columns: []string{"id"},
		Rows:    nil,
	}

	out, err := report.FormatResult(empty, report.FormatJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	trimmed := strings.TrimSpace(out)
	if trimmed != "[]" {
		t.Errorf("expected empty JSON array '[]', got %q", trimmed)
	}
}

func TestFormatNullValues(t *testing.T) {
	withNull := &db.QueryResult{
		Columns: []string{"id", "name", "notes"},
		Rows: [][]string{
			{"1", "Alice", "NULL"},
			{"2", "Bob", "some note"},
		},
	}

	for _, fmt := range []report.Format{report.FormatTable, report.FormatCSV, report.FormatJSON} {
		out, err := report.FormatResult(withNull, fmt)
		if err != nil {
			t.Errorf("format %s: unexpected error: %v", fmt, err)
			continue
		}
		if !strings.Contains(out, "NULL") {
			t.Errorf("format %s: expected 'NULL' in output:\n%s", fmt, out)
		}
	}
}

func TestFormatTableNoColumns(t *testing.T) {
	noCol := &db.QueryResult{}
	out, err := report.FormatResult(noCol, report.FormatTable)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "no columns") {
		t.Errorf("expected 'no columns' message, got: %q", out)
	}
}
