package tools

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/teerarak/mcp-postgresql-report/internal/db"
	"github.com/teerarak/mcp-postgresql-report/internal/report"
)

// DescribeTableTool returns the MCP tool definition for describe_table.
func DescribeTableTool() mcp.Tool {
	return mcp.NewTool("describe_table",
		mcp.WithDescription("Show column definitions and indexes of a PostgreSQL table."),
		mcp.WithString("table_name",
			mcp.Required(),
			mcp.Description("Name of the table to describe"),
		),
		mcp.WithString("schema", mcp.Description("Schema name (default: public)")),
	)
}

// DescribeTableHandler returns the handler for describe_table.
func DescribeTableHandler(manager *db.Manager, errLog *log.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := manager.EnsureConnected(ctx); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		pool, err := manager.Pool()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		tableName := req.GetString("table_name", "")
		if tableName == "" {
			return mcp.NewToolResultError("table_name is required"), nil
		}
		schema := req.GetString("schema", "public")

		// Query columns
		const colQuery = `
SELECT column_name, data_type, is_nullable, column_default, character_maximum_length
FROM information_schema.columns
WHERE table_schema = $1 AND table_name = $2
ORDER BY ordinal_position`

		colResult, err := db.FetchRows(ctx, pool, colQuery, schema, tableName)
		if err != nil {
			errLog.Printf("describe_table columns error: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("columns query error: %v", err)), nil
		}

		if len(colResult.Rows) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("Table '%s.%s' not found or has no columns", schema, tableName)), nil
		}

		colFormatted, err := report.FormatResult(colResult, report.FormatTable)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Query indexes
		const idxQuery = `
SELECT indexname, indexdef
FROM pg_indexes
WHERE schemaname = $1 AND tablename = $2
ORDER BY indexname`

		idxResult, err := db.FetchRows(ctx, pool, idxQuery, schema, tableName)
		if err != nil {
			errLog.Printf("describe_table indexes error: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("indexes query error: %v", err)), nil
		}

		out := fmt.Sprintf("Table: %s.%s\n\n### Columns\n\n%s", schema, tableName, colFormatted)

		if len(idxResult.Rows) > 0 {
			idxFormatted, err := report.FormatResult(idxResult, report.FormatTable)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out += fmt.Sprintf("\n### Indexes\n\n%s", idxFormatted)
		}

		return mcp.NewToolResultText(out), nil
	}
}
