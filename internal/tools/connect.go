package tools

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/teerarak/mcp-postgresql-report/internal/config"
	"github.com/teerarak/mcp-postgresql-report/internal/db"
)

// ConnectDatabaseTool returns the MCP tool definition for connect_database.
func ConnectDatabaseTool() mcp.Tool {
	return mcp.NewTool("connect_database",
		mcp.WithDescription("Connect to a PostgreSQL database. If not called, connection is established automatically from environment variables (POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB)."),
		mcp.WithString("host", mcp.Description("PostgreSQL host (overrides POSTGRES_HOST)")),
		mcp.WithString("port", mcp.Description("PostgreSQL port (overrides POSTGRES_PORT)")),
		mcp.WithString("user", mcp.Description("PostgreSQL username (overrides POSTGRES_USER)")),
		mcp.WithString("password", mcp.Description("PostgreSQL password (overrides POSTGRES_PASSWORD)")),
		mcp.WithString("database", mcp.Description("PostgreSQL database name (overrides POSTGRES_DB)")),
		mcp.WithString("ssl_mode",
			mcp.Description("SSL mode: disable, require, verify-ca, verify-full"),
			mcp.Enum("disable", "require", "verify-ca", "verify-full"),
		),
	)
}

// ConnectDatabaseHandler returns the handler for connect_database.
func ConnectDatabaseHandler(manager *db.Manager, errLog *log.Logger) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		cfg := config.Load()

		// Override with any provided arguments
		if v := req.GetString("host", ""); v != "" {
			cfg.Host = v
		}
		if v := req.GetString("port", ""); v != "" {
			cfg.Port = v
		}
		if v := req.GetString("user", ""); v != "" {
			cfg.User = v
		}
		if v := req.GetString("password", ""); v != "" {
			cfg.Password = v
		}
		if v := req.GetString("database", ""); v != "" {
			cfg.DBName = v
		}
		if v := req.GetString("ssl_mode", ""); v != "" {
			cfg.SSLMode = v
		}

		if cfg.DBName == "" {
			return mcp.NewToolResultError("database name is required (set 'database' argument or POSTGRES_DB env var)"), nil
		}

		if err := manager.Connect(ctx, cfg); err != nil {
			errLog.Printf("connect_database error: %v", err)
			return mcp.NewToolResultError(fmt.Sprintf("connection failed: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf(
			"Successfully connected to PostgreSQL database '%s' at %s:%s",
			cfg.DBName, cfg.Host, cfg.Port,
		)), nil
	}
}
