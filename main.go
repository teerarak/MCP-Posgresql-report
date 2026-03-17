package main

import (
	"context"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/teerarak/mcp-postgresql-report/internal/config"
	"github.com/teerarak/mcp-postgresql-report/internal/db"
	"github.com/teerarak/mcp-postgresql-report/internal/tools"
)

func main() {
	// Log to stderr — stdout is reserved for the MCP stdio protocol
	errLog := log.New(os.Stderr, "[mcp-pg] ", log.LstdFlags)

	manager := db.NewManager()
	defer manager.Close()

	// Attempt eager connection from env vars; non-fatal if vars are absent
	cfg := config.Load()
	if cfg.DBName != "" {
		if err := manager.Connect(context.Background(), cfg); err != nil {
			errLog.Printf("Warning: initial DB connection failed: %v", err)
		} else {
			errLog.Printf("Connected to database '%s' at %s:%s", cfg.DBName, cfg.Host, cfg.Port)
		}
	} else {
		errLog.Printf("No POSTGRES_DB set — use connect_database tool to connect")
	}

	s := server.NewMCPServer(
		"postgresql-report",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	tools.RegisterAll(s, manager, errLog)

	errLog.Println("Starting MCP server (stdio transport)...")
	if err := server.ServeStdio(s); err != nil {
		errLog.Fatalf("Server error: %v", err)
	}
}
