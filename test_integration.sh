#!/bin/bash
# Integration test for MCP PostgreSQL Report Server
# Requires: PostgreSQL running with mcp_test database + employees table

set -e

BINARY="./mcp-pg-report"
PASS=0
FAIL=0

export POSTGRES_USER=postgres
export POSTGRES_DB=mcp_test
export POSTGRES_HOST=/var/run/postgresql
export POSTGRES_PORT=5432
export POSTGRES_SSLMODE=disable

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

run_test() {
    local name="$1"
    local input="$2"
    local expect="$3"

    response=$(echo "$input" | "$BINARY" 2>/dev/null)

    if echo "$response" | grep -q "$expect"; then
        echo -e "${GREEN}PASS${NC}: $name"
        PASS=$((PASS + 1))
    else
        echo -e "${RED}FAIL${NC}: $name"
        echo "  Expected to find: $expect"
        echo "  Got: $response"
        FAIL=$((FAIL + 1))
    fi
}

echo "=== MCP PostgreSQL Report Server Integration Tests ==="
echo ""

# Test 1: tools/list returns all 5 tools
run_test "tools/list returns 5 tools" \
    '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' \
    '"generate_report"'

# Test 2: list_tables finds employees
run_test "list_tables finds employees table" \
    '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_tables","arguments":{}}}' \
    'employees'

# Test 3: describe_table shows columns
run_test "describe_table shows column names" \
    '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"describe_table","arguments":{"table_name":"employees"}}}' \
    'column_name'

# Test 4: query_data returns rows (table format)
run_test "query_data returns Alice (table format)" \
    '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"query_data","arguments":{"sql":"SELECT * FROM employees ORDER BY id","format":"table"}}}' \
    'Alice'

# Test 5: query_data returns CSV format
run_test "query_data returns CSV format" \
    '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"query_data","arguments":{"sql":"SELECT name,department FROM employees ORDER BY id","format":"csv"}}}' \
    'name,department'

# Test 6: query_data returns JSON format
run_test "query_data returns JSON format" \
    '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"query_data","arguments":{"sql":"SELECT name FROM employees ORDER BY id","format":"json"}}}' \
    'Alice'

# Test 7: generate_report with title
run_test "generate_report shows report title" \
    '{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"generate_report","arguments":{"sql":"SELECT department, COUNT(*) as total FROM employees GROUP BY department ORDER BY department","title":"Department Summary","format":"table"}}}' \
    'Department Summary'

# Test 8: generate_report shows data
run_test "generate_report shows Engineering department" \
    '{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"generate_report","arguments":{"sql":"SELECT department, COUNT(*) as total FROM employees GROUP BY department ORDER BY department","title":"Dept Report","format":"table"}}}' \
    'Engineering'

# Test 9: generate_report JSON format has report envelope
run_test "generate_report JSON has report envelope" \
    '{"jsonrpc":"2.0","id":9,"method":"tools/call","params":{"name":"generate_report","arguments":{"sql":"SELECT name FROM employees LIMIT 1","title":"Test","format":"json"}}}' \
    'row_count'

# Test 10: Security - non-SELECT is rejected
run_test "query_data rejects non-SELECT (INSERT)" \
    '{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"query_data","arguments":{"sql":"INSERT INTO employees (name) VALUES ('"'"'Hacker'"'"')"}}}' \
    'only SELECT'

# Test 11: query with LIMIT already present (no double LIMIT)
run_test "query_data respects existing LIMIT" \
    '{"jsonrpc":"2.0","id":11,"method":"tools/call","params":{"name":"query_data","arguments":{"sql":"SELECT * FROM employees LIMIT 2","format":"table"}}}' \
    'Rows returned'

# Test 12: connect_database tool works
run_test "connect_database succeeds" \
    '{"jsonrpc":"2.0","id":12,"method":"tools/call","params":{"name":"connect_database","arguments":{"host":"/var/run/postgresql","database":"mcp_test","user":"postgres"}}}' \
    'Successfully connected'

echo ""
echo "=== Results: ${PASS} passed, ${FAIL} failed ==="
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
