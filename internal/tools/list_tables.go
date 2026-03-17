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

// ListTablesTool returns the MCP tool definition for list_tables.
func ListTablesTool() mcp.Tool {
	return mcp.NewTool("list_tables",
		mcp.WithDescription("List all tables in the PostgreSQL database schema."),
		mcp.WithString("schema", mcp.Description("Schema name to list tables from (default: public)")),
	)
}

// ListTablesHandler returns the handler for list_tables.
func ListTablesHandler(manager *db.Manager, errLog *log.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := manager.EnsureConnected(ctx); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		pool, err := manager.Pool()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		schema := req.GetString("schema", "public")

		const query = `
SELECT table_name, table_type
FROM information_schema.tables
WHERE table_schema = $1
ORDER BY table_name`

		result, err := db.FetchRows(ctx, pool, query, schema)
		if err != nil {
			errLog.Printf("list_tables error: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("query error: %v", err)), nil
		}

		if len(result.Rows) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No tables found in schema '%s'", schema)), nil
		}

		formatted, err := report.FormatResult(result, report.FormatTable)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Tables in schema '%s':\n\n%s", schema, formatted)), nil
	}
}
