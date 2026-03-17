package tools

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/teerarak/mcp-postgresql-report/internal/db"
	"github.com/teerarak/mcp-postgresql-report/internal/report"
)

var hasLimitRe = regexp.MustCompile(`(?i)\bLIMIT\b`)

// QueryDataTool returns the MCP tool definition for query_data.
func QueryDataTool() mcp.Tool {
	return mcp.NewTool("query_data",
		mcp.WithDescription("Execute a SELECT SQL query on the PostgreSQL database and return the results."),
		mcp.WithString("sql",
			mcp.Required(),
			mcp.Description("SQL SELECT statement to execute"),
		),
		mcp.WithString("format",
			mcp.Description("Output format: table (default), csv, json"),
			mcp.Enum("table", "csv", "json"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of rows to return (default: 100, max: 10000)"),
		),
	)
}

// QueryDataHandler returns the handler for query_data.
func QueryDataHandler(manager *db.Manager, errLog *log.Logger) server.ToolHandlerFunc {
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

		rowLimit := req.GetInt("limit", 100)
		if rowLimit <= 0 || rowLimit > 10000 {
			rowLimit = 100
		}

		result, toolErr := runQuery(ctx, pool, sqlQuery, rowLimit, errLog)
		if toolErr != nil {
			return toolErr, nil
		}

		outFormat := report.Format(req.GetString("format", "table"))
		formatted, err := report.FormatResult(result, outFormat)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Rows returned: %d\n\n%s", len(result.Rows), formatted)), nil
	}
}

// runQuery validates and executes a SELECT query in a read-only transaction.
// Returns either a QueryResult or a *mcp.CallToolResult error.
func runQuery(ctx context.Context, pool *pgxpool.Pool, sqlQuery string, rowLimit int, errLog *log.Logger) (*db.QueryResult, *mcp.CallToolResult) {
	trimmed := strings.TrimSpace(sqlQuery)
	if !strings.HasPrefix(strings.ToUpper(trimmed), "SELECT") {
		return nil, mcp.NewToolResultError("only SELECT statements are allowed")
	}

	// Append LIMIT if not already present
	if !hasLimitRe.MatchString(sqlQuery) {
		sqlQuery = fmt.Sprintf("%s LIMIT %d", sqlQuery, rowLimit)
	}

	// Execute inside a read-only transaction
	tx, err := pool.Begin(ctx)
	if err != nil {
		errLog.Printf("begin tx error: %v", err)
		return nil, mcp.NewToolResultError(fmt.Sprintf("begin transaction: %v", err))
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, "SET TRANSACTION READ ONLY"); err != nil {
		return nil, mcp.NewToolResultError(fmt.Sprintf("set read-only: %v", err))
	}

	rows, err := tx.Query(ctx, sqlQuery)
	if err != nil {
		errLog.Printf("query error: %v", err)
		return nil, mcp.NewToolResultError(fmt.Sprintf("query error: %v", err))
	}
	defer rows.Close()

	fieldDescs := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescs))
	for i, fd := range fieldDescs {
		columns[i] = string(fd.Name)
	}

	var resultRows [][]string
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return nil, mcp.NewToolResultError(fmt.Sprintf("scan row: %v", err))
		}
		row := make([]string, len(vals))
		for i, v := range vals {
			if v == nil {
				row[i] = "NULL"
			} else {
				row[i] = fmt.Sprintf("%v", v)
			}
		}
		resultRows = append(resultRows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, mcp.NewToolResultError(fmt.Sprintf("rows error: %v", err))
	}

	return &db.QueryResult{Columns: columns, Rows: resultRows}, nil
}
