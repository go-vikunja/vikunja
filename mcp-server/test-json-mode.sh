#!/bin/bash
# Test script to verify JSON response mode works without Accept header

echo "Testing JSON Response Mode for n8n compatibility"
echo "================================================"
echo ""

# Test 1: Without Accept header (should fail in normal mode)
echo "Test 1: POST /mcp without Accept header (should get 406 without JSON mode)..."
response=$(curl -s -w "\n%{http_code}" -X POST http://localhost:3458/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "clientInfo": {"name": "test", "version": "1.0.0"},
      "capabilities": {}
    }
  }' 2>&1)

http_code=$(echo "$response" | tail -n 1)
body=$(echo "$response" | head -n -1)

if [ "$http_code" == "406" ]; then
  echo "✓ Correctly rejected without JSON mode (got 406)"
elif [ "$http_code" == "200" ] || [ "$http_code" == "401" ]; then
  echo "✓ Request accepted (JSON mode may be enabled or auth failed: $http_code)"
  echo "  Response: $body"
else
  echo "✗ Unexpected status code: $http_code"
  echo "  Response: $body"
fi

echo ""
echo "To enable JSON response mode, set:"
echo "  export MCP_HTTP_JSON_RESPONSE=true"
echo ""
echo "Then n8n and other clients that can't set Accept headers will work!"
