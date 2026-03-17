package report

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/teerarak/mcp-postgresql-report/internal/db"
)

// Format defines supported report output formats.
type Format string

const (
	FormatTable Format = "table"
	FormatCSV   Format = "csv"
	FormatJSON  Format = "json"
)

// FormatResult formats a QueryResult into the requested output format.
func FormatResult(result *db.QueryResult, format Format) (string, error) {
	switch format {
	case FormatCSV:
		return formatCSV(result)
	case FormatJSON:
		return formatJSON(result)
	default:
		return formatTable(result), nil
	}
}

// formatTable produces a pipe-delimited text table using tabwriter.
func formatTable(result *db.QueryResult) string {
	if len(result.Columns) == 0 {
		return "(no columns)"
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)

	// Header row
	fmt.Fprintln(w, "| "+strings.Join(result.Columns, " \t| ")+" \t|")

	// Separator
	seps := make([]string, len(result.Columns))
	for i, col := range result.Columns {
		seps[i] = strings.Repeat("-", len(col))
	}
	fmt.Fprintln(w, "|-"+strings.Join(seps, "-\t|-")+"-\t|")

	// Data rows
	for _, row := range result.Rows {
		fmt.Fprintln(w, "| "+strings.Join(row, " \t| ")+" \t|")
	}

	w.Flush()
	return buf.String()
}

// formatCSV produces RFC 4180 CSV output.
func formatCSV(result *db.QueryResult) (string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if err := w.Write(result.Columns); err != nil {
		return "", fmt.Errorf("write CSV header: %w", err)
	}
	for _, row := range result.Rows {
		if err := w.Write(row); err != nil {
			return "", fmt.Errorf("write CSV row: %w", err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// formatJSON produces indented JSON as an array of objects.
func formatJSON(result *db.QueryResult) (string, error) {
	records := make([]map[string]string, 0, len(result.Rows))
	for _, row := range result.Rows {
		rec := make(map[string]string, len(result.Columns))
		for i, col := range result.Columns {
			if i < len(row) {
				rec[col] = row[i]
			}
		}
		records = append(records, rec)
	}

	out, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal JSON: %w", err)
	}
	return string(out), nil
}
