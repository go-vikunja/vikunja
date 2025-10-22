#!/usr/bin/env bash
# Test MCP HTTP Transport Configuration
# Purpose: Verify that systemd units have correct TRANSPORT_TYPE=http environment variable
# Usage: ./test-http-transport.sh CONTAINER_ID

set -euo pipefail

CONTAINER_ID="${1:-}"

if [[ -z "$CONTAINER_ID" ]]; then
    echo "Error: Container ID required"
    echo "Usage: $0 CONTAINER_ID"
    exit 1
fi

echo "Testing MCP HTTP Transport Configuration..."
echo "Container ID: $CONTAINER_ID"
echo ""

# Test blue environment
echo "=== Blue Environment ==="
if pct exec "$CONTAINER_ID" -- systemctl cat vikunja-mcp-blue | grep -q "TRANSPORT_TYPE=http"; then
    echo "✓ TRANSPORT_TYPE=http found in vikunja-mcp-blue"
else
    echo "✗ TRANSPORT_TYPE=http NOT found in vikunja-mcp-blue"
    exit 1
fi

# Test green environment
echo ""
echo "=== Green Environment ==="
if pct exec "$CONTAINER_ID" -- systemctl cat vikunja-mcp-green | grep -q "TRANSPORT_TYPE=http"; then
    echo "✓ TRANSPORT_TYPE=http found in vikunja-mcp-green"
else
    echo "✗ TRANSPORT_TYPE=http NOT found in vikunja-mcp-green"
    exit 1
fi

echo ""
echo "✓ All tests passed - HTTP transport is configured correctly"
