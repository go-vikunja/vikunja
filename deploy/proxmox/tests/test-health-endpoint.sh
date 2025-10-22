#!/usr/bin/env bash
# Test MCP HTTP Health Endpoint
# Purpose: Validate MCP HTTP server responds correctly to health checks
# Usage: ./test-health-endpoint.sh CONTAINER_ID PORT

set -euo pipefail

CONTAINER_ID="${1:-}"
MCP_PORT="${2:-3010}"

if [[ -z "$CONTAINER_ID" ]]; then
    echo "Error: Container ID required"
    echo "Usage: $0 CONTAINER_ID [PORT]"
    exit 1
fi

echo "Testing MCP HTTP Health Endpoint..."
echo "Container ID: $CONTAINER_ID"
echo "MCP Port: $MCP_PORT"
echo ""

# Test 1: Check if port is listening
echo "=== Test 1: Port Listening ==="
if pct exec "$CONTAINER_ID" -- ss -tuln | grep -q ":$MCP_PORT "; then
    echo "✓ Port $MCP_PORT is listening"
else
    echo "✗ Port $MCP_PORT is NOT listening"
    exit 1
fi

# Test 2: Test SSE endpoint returns 401 (no auth token)
echo ""
echo "=== Test 2: SSE Endpoint Authentication ==="
if pct exec "$CONTAINER_ID" -- bash -c "curl -f -s -I http://localhost:$MCP_PORT/sse 2>&1 | grep -q '401\\|Unauthorized'"; then
    echo "✓ SSE endpoint returns 401 Unauthorized as expected"
else
    echo "✗ SSE endpoint did not return expected 401 response"
    exit 1
fi

# Test 3: Verify service is running
echo ""
echo "=== Test 3: Service Status ==="
if [[ "$MCP_PORT" == "3010" ]]; then
    service_name="vikunja-mcp-blue"
elif [[ "$MCP_PORT" == "3011" ]]; then
    service_name="vikunja-mcp-green"
else
    service_name="vikunja-mcp-blue"
fi

if pct exec "$CONTAINER_ID" -- systemctl is-active --quiet "$service_name"; then
    echo "✓ Service $service_name is active"
else
    echo "✗ Service $service_name is NOT active"
    exit 1
fi

echo ""
echo "✓ All health checks passed - MCP HTTP server is working correctly"
