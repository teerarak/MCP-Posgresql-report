package tools

import (
	"log"

	"github.com/mark3labs/mcp-go/server"
	"github.com/teerarak/mcp-postgresql-report/internal/db"
)

// RegisterAll registers all MCP tools into the server.
func RegisterAll(s *server.MCPServer, manager *db.Manager, errLog *log.Logger) {
	s.AddTool(ConnectDatabaseTool(), ConnectDatabaseHandler(manager, errLog))
	s.AddTool(ListTablesTool(), ListTablesHandler(manager, errLog))
	s.AddTool(DescribeTableTool(), DescribeTableHandler(manager, errLog))
	s.AddTool(QueryDataTool(), QueryDataHandler(manager, errLog))
	s.AddTool(GenerateReportTool(), GenerateReportHandler(manager, errLog))
}
