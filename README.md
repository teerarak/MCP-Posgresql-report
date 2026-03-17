# MCP PostgreSQL Report Server

MCP (Model Context Protocol) server เขียนด้วย Go สำหรับเชื่อมต่อกับ PostgreSQL และดึงข้อมูลออกมาเป็นรายงานในรูปแบบต่างๆ

## Features

- เชื่อมต่อ PostgreSQL ผ่าน environment variables หรือ tool arguments
- แสดงรายชื่อตารางในฐานข้อมูล
- ดู schema ของตาราง (columns + indexes)
- รัน SELECT query และแสดงผล
- สร้างรายงานพร้อม title และ metadata ในรูปแบบ table, CSV, หรือ JSON

## MCP Tools

| Tool | Description |
|------|-------------|
| `connect_database` | เชื่อมต่อ PostgreSQL (ถ้าไม่ call จะใช้ env vars อัตโนมัติ) |
| `list_tables` | แสดงตารางทั้งหมดใน schema |
| `describe_table` | แสดง columns และ indexes ของตาราง |
| `query_data` | รัน SELECT query และ format ผลลัพธ์ |
| `generate_report` | สร้าง report พร้อม header/title/description |

## Build

```bash
go build -o mcp-pg-report .
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_HOST` | `localhost` | PostgreSQL host |
| `POSTGRES_PORT` | `5432` | PostgreSQL port |
| `POSTGRES_USER` | `postgres` | PostgreSQL username |
| `POSTGRES_PASSWORD` | _(empty)_ | PostgreSQL password |
| `POSTGRES_DB` | _(required)_ | Database name |
| `POSTGRES_SSLMODE` | `disable` | SSL mode |

## Claude Desktop Configuration

เพิ่มใน `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "postgresql-report": {
      "command": "/path/to/mcp-pg-report",
      "env": {
        "POSTGRES_HOST": "localhost",
        "POSTGRES_PORT": "5432",
        "POSTGRES_USER": "myuser",
        "POSTGRES_PASSWORD": "mypassword",
        "POSTGRES_DB": "mydb"
      }
    }
  }
}
```

## Usage Examples

### List tables
```
list_tables(schema: "public")
```

### Describe a table
```
describe_table(table_name: "users")
```

### Query data
```
query_data(sql: "SELECT * FROM users WHERE active = true", format: "table", limit: 50)
```

### Generate a report
```
generate_report(
  sql: "SELECT department, COUNT(*) as total FROM employees GROUP BY department",
  title: "Employee Count by Department",
  description: "Summary report as of today",
  format: "table"
)
```

### Output formats

**table** (default):
```
| id  | name  | email           |
|-----|-------|-----------------|
| 1   | Alice | alice@email.com |
| 2   | Bob   | bob@email.com   |
```

**csv**:
```
id,name,email
1,Alice,alice@email.com
2,Bob,bob@email.com
```

**json**:
```json
[
  {"id": "1", "name": "Alice", "email": "alice@email.com"},
  {"id": "2", "name": "Bob", "email": "bob@email.com"}
]
```

## Project Structure

```
.
├── main.go                          # MCP server entrypoint (stdio transport)
├── internal/
│   ├── config/config.go             # Environment variable configuration
│   ├── db/
│   │   ├── pool.go                  # PostgreSQL connection pool manager
│   │   └── queries.go               # Generic query executor → QueryResult
│   ├── report/
│   │   └── formatter.go             # table/CSV/JSON formatters (stdlib only)
│   └── tools/
│       ├── register.go              # Registers all tools into MCP server
│       ├── connect.go               # connect_database
│       ├── list_tables.go           # list_tables
│       ├── describe_table.go        # describe_table
│       ├── query_data.go            # query_data (read-only TX)
│       └── generate_report.go      # generate_report
```
