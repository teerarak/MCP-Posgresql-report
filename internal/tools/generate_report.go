package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/teerarak/mcp-postgresql-report/internal/db"
	"github.com/teerarak/mcp-postgresql-report/internal/report"
)

// GenerateReportTool returns the MCP tool definition for generate_report.
func GenerateReportTool() mcp.Tool {
	return mcp.NewTool("generate_report",
		mcp.WithDescription("Execute a SQL query and generate a formatted report with title and metadata."),
		mcp.WithString("sql",
			mcp.Required(),
			mcp.Description("SQL SELECT statement for the report data"),
		),
		mcp.WithString("title",
			mcp.Description("Report title"),
		),
		mcp.WithString("description",
			mcp.Description("Report description or notes"),
		),
		mcp.WithString("format",
			mcp.Description("Output format: table (default), csv, json"),
			mcp.Enum("table", "csv", "json"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of rows (default: 100, max: 10000)"),
		),
	)
}

// GenerateReportHandler returns the handler for generate_report.
func GenerateReportHandler(manager *db.Manager, errLog *log.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := manager.EnsureConnected(ctx); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		pool, err := manager.Pool()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		sqlQuery := req.GetString("sql", "")
		if sqlQuery == "" {
			return mcp.NewToolResultError("sql is required"), nil
		}

		title := req.GetString("title", "Query Report")
		description := req.GetString("description", "")

		rowLimit := req.GetInt("limit", 100)
		if rowLimit <= 0 || rowLimit > 10000 {
			rowLimit = 100
		}

		outFormat := report.Format(req.GetString("format", "table"))

		result, toolErr := runQuery(ctx, pool, sqlQuery, rowLimit, errLog)
		if toolErr != nil {
			return toolErr, nil
		}

		generatedAt := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")

		if outFormat == report.FormatJSON {
			return buildJSONReport(result, title, description, generatedAt)
		}

		return buildTextReport(result, outFormat, title, description, generatedAt)
	}
}

func buildTextReport(result *db.QueryResult, outFormat report.Format, title, description, generatedAt string) (*mcp.CallToolResult, error) {
	sep := "=============================="
	header := fmt.Sprintf("%s\nReport: %s\nGenerated: %s\n", sep, title, generatedAt)
	if description != "" {
		header += fmt.Sprintf("Description: %s\n", description)
	}
	header += fmt.Sprintf("%s\nRows returned: %d\n\n", sep, len(result.Rows))

	formatted, err := report.FormatResult(result, outFormat)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	footer := "\n" + sep
	return mcp.NewToolResultText(header + formatted + footer), nil
}

func buildJSONReport(result *db.QueryResult, title, description, generatedAt string) (*mcp.CallToolResult, error) {
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

	envelope := map[string]any{
		"report": map[string]any{
			"title":        title,
			"description":  description,
			"generated_at": generatedAt,
			"row_count":    len(result.Rows),
			"data":         records,
		},
	}

	out, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("marshal JSON: %v", err)), nil
	}
	return mcp.NewToolResultText(string(out)), nil
}
