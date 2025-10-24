# Quickstart Guide: HTTP Transport for MCP Server

**Feature**: HTTP Transport for MCP Server  
**Date**: October 22, 2025  
**Audience**: Developers, System Administrators

## Overview

This guide helps you deploy and use the Vikunja MCP Server with HTTP transport (HTTP Streamable and SSE) for remote client connections.

**What you'll learn**:
- Deploy MCP server to Proxmox LXC container
- Configure HTTP transport and authentication
- Connect MCP clients (n8n, Claude Desktop)
- Monitor and troubleshoot the server

**Prerequisites**:
- Vikunja API server running (v0.20+)
- Redis server (optional, recommended for production)
- Node.js 22+ installed
- pnpm package manager
- Basic Linux command line knowledge

---

## Quick Start (5 Minutes)

### 1. Install Dependencies

```bash
# On Proxmox LXC container or server
curl -fsSL https://deb.nodesource.com/setup_22.x | bash -
apt-get install -y nodejs redis-server
npm install -g pnpm pm2

# Verify installations
node --version  # Should be v22.x.x
redis-cli ping  # Should return PONG
```

### 2. Clone and Build

```bash
# Clone Vikunja repository (or use existing installation)
cd /opt/vikunja
git pull origin main

# Navigate to MCP server
cd mcp-server

# Install dependencies
pnpm install

# Build TypeScript
pnpm build
```

### 3. Configure Environment

```bash
# Copy example environment file
cp .env.example .env

# Edit configuration
nano .env
```

**Required configuration** (`.env`):

```bash
# Vikunja API connection
VIKUNJA_API_URL=http://localhost:3456

# HTTP transport settings
MCP_HTTP_ENABLED=true
MCP_HTTP_PORT=3010

# Redis (optional, recommended)
REDIS_URL=redis://localhost:6379

# Rate limiting
RATE_LIMIT_POINTS=100        # Requests per window
RATE_LIMIT_DURATION=900      # Window in seconds (15 min)

# Logging
LOG_LEVEL=info               # debug | info | warn | error
NODE_ENV=production
```

### 4. Start Server

**Option A: Direct (development)**:
```bash
pnpm dev  # Runs with hot reload
```

**Option B: PM2 (production)**:
```bash
pm2 start dist/index.js --name vikunja-mcp
pm2 save
pm2 startup  # Enable auto-start on boot
```

**Option C: Systemd (recommended for production)**:
```bash
# Create systemd service
sudo cp vikunja-mcp.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable vikunja-mcp
sudo systemctl start vikunja-mcp

# Check status
sudo systemctl status vikunja-mcp
```

### 5. Verify Deployment

```bash
# Check health endpoint
curl http://localhost:3010/health

# Expected response:
# {
#   "status": "healthy",
#   "version": "1.1.0",
#   "checks": {
#     "redis": {"status": "healthy"},
#     "vikunja_api": {"status": "healthy"}
#   }
# }
```

**🎉 Server is running!** Continue to client setup below.

---

## Client Setup

### Option 1: n8n Workflow Automation

**1. Install MCP Agent Node** (if not already installed):
- n8n v1.0+ includes built-in MCP support
- Or install community node: `n8n-nodes-langchain.mcpagent`

**2. Add MCP Server to n8n**:

In n8n workflow editor:
1. Add "MCP Agent" node
2. Create new MCP Server credential
3. Configure:
   - **Transport**: HTTP Streamable (recommended)
   - **URL**: `http://192.168.50.64:3010/mcp`
   - **Authentication**: Bearer token
   - **Token**: Your Vikunja API token (from Settings → API Tokens)

**3. Test Connection**:
```json
// In MCP Agent node
{
  "action": "list_tools",
  "server": "vikunja-mcp"
}

// Expected output:
{
  "tools": [
    {"name": "get_tasks", "description": "Get tasks from a project"},
    {"name": "create_task", "description": "Create a new task"},
    // ... 15+ tools
  ]
}
```

**4. Example Workflow**:
```
Webhook Trigger → MCP Agent (get_tasks) → Process Data → Send Email
```

---

### Option 2: Claude Desktop

**1. Configure MCP Server** in `claude_desktop_config.json`:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`  
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`  
**Linux**: `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "vikunja": {
      "transport": {
        "type": "http",
        "url": "http://192.168.50.64:3010/mcp",
        "headers": {
          "Authorization": "Bearer tk_YOUR_VIKUNJA_API_TOKEN"
        }
      }
    }
  }
}
```

**2. Restart Claude Desktop**

**3. Test in Conversation**:
```
You: "Using Vikunja MCP, list my tasks from project 1"

Claude: I'll use the get_tasks tool...
[Lists your tasks from Vikunja]
```

---

### Option 3: Custom MCP Client (TypeScript)

```typescript
import { Client } from '@modelcontextprotocol/sdk/client/index.js';
import { HTTPStreamableTransport } from '@modelcontextprotocol/sdk/client/http.js';

// Create transport
const transport = new HTTPStreamableTransport({
  url: 'http://192.168.50.64:3010/mcp',
  headers: {
    'Authorization': 'Bearer tk_YOUR_VIKUNJA_API_TOKEN'
  }
});

// Create client
const client = new Client({
  name: 'my-app',
  version: '1.0.0'
});

// Connect
await client.connect(transport);

// List tools
const tools = await client.listTools();
console.log('Available tools:', tools.tools.map(t => t.name));

// Call tool
const result = await client.callTool({
  name: 'get_tasks',
  arguments: { project_id: 1 }
});
console.log('Tasks:', result.content);

// Disconnect
await client.close();
```

---

## Production Deployment

### Proxmox LXC Automated Deployment

Use the automated deployment script:

```bash
cd /opt/vikunja/deploy/proxmox
./deploy.sh

# Script will:
# 1. Install Node.js 22+ and PM2
# 2. Build MCP server
# 3. Setup systemd service
# 4. Configure Redis
# 5. Start services
# 6. Verify health checks
```

**Deployment script updates** (already included in 006-mcp-http-transport):
- Port configuration (3010 for MCP HTTP)
- Health check validation
- Service dependencies (Redis, Vikunja API)
- Summary with connection details

---

### Reverse Proxy (nginx) with TLS

**1. Install nginx**:
```bash
apt-get install -y nginx certbot python3-certbot-nginx
```

**2. Configure nginx** (`/etc/nginx/sites-available/vikunja-mcp`):

```nginx
upstream vikunja_mcp {
    server 127.0.0.1:3010;
}

server {
    listen 443 ssl http2;
    server_name vikunja.example.com;

    ssl_certificate /etc/letsencrypt/live/vikunja.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/vikunja.example.com/privkey.pem;

    # MCP HTTP endpoints
    location /mcp {
        proxy_pass http://vikunja_mcp;
        proxy_http_version 1.1;
        
        # Required for SSE
        proxy_set_header Connection '';
        proxy_buffering off;
        chunked_transfer_encoding off;
        
        # Standard headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts for long-lived connections
        proxy_read_timeout 1h;
        proxy_send_timeout 1h;
    }

    # Health check endpoint
    location /mcp/health {
        proxy_pass http://vikunja_mcp/health;
        proxy_http_version 1.1;
    }
}

# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name vikunja.example.com;
    return 301 https://$server_name$request_uri;
}
```

**3. Enable and test**:
```bash
# Enable site
ln -s /etc/nginx/sites-available/vikunja-mcp /etc/nginx/sites-enabled/
nginx -t  # Test configuration
systemctl reload nginx

# Get SSL certificate
certbot --nginx -d vikunja.example.com

# Test
curl https://vikunja.example.com/mcp/health
```

**4. Update client URLs**:
```
Old: http://192.168.50.64:3010/mcp
New: https://vikunja.example.com/mcp
```

---

## Monitoring

### Health Checks

**Automated monitoring** (add to cron or systemd timer):

```bash
#!/bin/bash
# /usr/local/bin/check-mcp-health.sh

HEALTH_URL="http://localhost:3010/health"
RESPONSE=$(curl -s $HEALTH_URL)
STATUS=$(echo $RESPONSE | jq -r '.status')

if [ "$STATUS" != "healthy" ]; then
    echo "MCP server unhealthy: $RESPONSE"
    # Send alert (email, Slack, PagerDuty, etc.)
    systemctl restart vikunja-mcp
fi
```

**Prometheus metrics** (future enhancement):
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'vikunja-mcp'
    static_configs:
      - targets: ['localhost:3010']
    metrics_path: '/metrics'
```

---

### Logs

**PM2 logs**:
```bash
pm2 logs vikunja-mcp
pm2 logs vikunja-mcp --lines 100
pm2 logs vikunja-mcp --err  # Errors only
```

**Systemd logs**:
```bash
journalctl -u vikunja-mcp -f  # Follow logs
journalctl -u vikunja-mcp --since "1 hour ago"
journalctl -u vikunja-mcp --priority=err
```

**Log rotation** (systemd):
```ini
# /etc/systemd/system/vikunja-mcp.service
[Service]
StandardOutput=journal
StandardError=journal
SyslogIdentifier=vikunja-mcp
```

---

### Performance Monitoring

**Active sessions**:
```bash
curl -s http://localhost:3010/health | jq '.sessions'
# Output: {"active": 12, "total_created": 145}
```

**Redis monitoring**:
```bash
redis-cli
> INFO stats
> KEYS vikunja:mcp:*
> TTL vikunja:mcp:token:abc123
```

**Resource usage**:
```bash
# PM2
pm2 monit

# System
top -p $(pgrep -f vikunja-mcp)
```

---

## Troubleshooting

### Issue: Connection Refused

**Symptoms**: Client cannot connect, "Connection refused" error

**Solutions**:
```bash
# 1. Check if server is running
systemctl status vikunja-mcp
pm2 list

# 2. Check if port is open
netstat -tlnp | grep 3010
curl http://localhost:3010/health

# 3. Check firewall
ufw status
ufw allow 3010/tcp  # If needed

# 4. Check logs
journalctl -u vikunja-mcp -n 50
```

---

### Issue: Authentication Failed

**Symptoms**: 401 Unauthorized, "Authentication failed"

**Solutions**:
```bash
# 1. Verify token is valid
curl -H "Authorization: Bearer tk_YOUR_TOKEN" \
     http://localhost:3456/api/v1/user

# 2. Check token in client config
# Ensure "Bearer " prefix in header, not in token value

# 3. Check Redis cache (if using)
redis-cli
> KEYS vikunja:mcp:token:*
> DEL vikunja:mcp:token:*  # Clear cache to force revalidation

# 4. Check logs for authentication attempts
journalctl -u vikunja-mcp | grep -i auth
```

---

### Issue: Rate Limit Exceeded

**Symptoms**: 429 Too Many Requests, "Rate limit exceeded"

**Solutions**:
```bash
# 1. Check current limit in .env
cat .env | grep RATE_LIMIT

# 2. Increase limit (if legitimate usage)
nano .env
# RATE_LIMIT_POINTS=200  # Increase from 100
systemctl restart vikunja-mcp

# 3. Reset rate limit for token
redis-cli
> KEYS vikunja:mcp:ratelimit:*
> DEL vikunja:mcp:ratelimit:HASH  # Replace HASH

# 4. Check who's making requests
journalctl -u vikunja-mcp | grep "rate limit"
```

---

### Issue: Redis Connection Failed

**Symptoms**: "Redis connection refused", degraded performance

**Solutions**:
```bash
# 1. Check Redis is running
systemctl status redis-server
redis-cli ping  # Should return PONG

# 2. Check Redis URL in .env
cat .env | grep REDIS_URL
# Should match redis-server config

# 3. Server can run without Redis (in-memory fallback)
# Edit .env, comment out REDIS_URL
# REDIS_URL=redis://localhost:6379
systemctl restart vikunja-mcp

# 4. Fix Redis connection
systemctl start redis-server
systemctl restart vikunja-mcp
```

---

### Issue: SSE Transport Not Working

**Symptoms**: SSE connection drops, EventSource errors

**Solutions**:
```bash
# 1. Use HTTP Streamable instead (recommended)
# Update client to use POST /mcp instead of GET /sse

# 2. If must use SSE, check nginx config
# Ensure proxy_buffering off and chunked_transfer_encoding off

# 3. Check browser console for EventSource errors
# F12 → Console → Look for "EventSource failed"

# 4. Verify token in URL query param
# http://server:3010/sse?token=tk_YOUR_TOKEN
```

---

## Security Best Practices

### 1. Use HTTPS in Production

```bash
# Always use reverse proxy with TLS
# Never expose HTTP transport directly to internet
nginx + Let's Encrypt (see "Reverse Proxy" section above)
```

### 2. Rotate API Tokens Regularly

```bash
# In Vikunja UI:
# Settings → API Tokens → Revoke old tokens → Create new

# Update all MCP clients with new token
# Old connections will fail after revocation
```

### 3. Monitor Authentication Attempts

```bash
# Set up alerts for failed auth attempts
journalctl -u vikunja-mcp | grep "Authentication failed" | wc -l
# If > 10 in short period, investigate
```

### 4. Rate Limit Aggressively

```bash
# Start conservative, increase if needed
RATE_LIMIT_POINTS=50   # Instead of 100
RATE_LIMIT_DURATION=900
```

### 5. Use Firewall Rules

```bash
# Only allow MCP port from known IPs
ufw allow from 192.168.1.0/24 to any port 3010
ufw deny 3010  # Block all others
```

---

## Migration from stdio to HTTP

If you're using MCP server via stdio (local only), migrate to HTTP for remote access:

### Before (stdio):

```json
// claude_desktop_config.json
{
  "mcpServers": {
    "vikunja": {
      "command": "node",
      "args": ["/opt/vikunja/mcp-server/dist/index.js"]
    }
  }
}
```

### After (HTTP):

```json
{
  "mcpServers": {
    "vikunja": {
      "transport": {
        "type": "http",
        "url": "https://vikunja.example.com/mcp",
        "headers": {
          "Authorization": "Bearer tk_YOUR_TOKEN"
        }
      }
    }
  }
}
```

**Benefits of HTTP transport**:
- Remote access (not just localhost)
- Multiple clients can connect simultaneously
- No need to run MCP server on client machine
- Centralized logging and monitoring

---

## Performance Tuning

### For High Concurrency (>50 clients)

```bash
# Increase Node.js max connections
node --max-http-header-size=16384 dist/index.js

# Increase file descriptor limits
ulimit -n 65536

# PM2 cluster mode (future enhancement)
# pm2 start dist/index.js -i max --name vikunja-mcp

# Redis persistence for rate limits
# Edit /etc/redis/redis.conf
# save 900 1
# save 300 10
```

### For Low Latency

```bash
# Use local Redis (same server)
REDIS_URL=redis://localhost:6379

# Increase token cache TTL (if acceptable security tradeoff)
# Edit src/config/schema.ts
# TOKEN_CACHE_TTL=600  # 10 minutes instead of 5

# Use keep-alive connections to Vikunja API
# Already configured in VikunjaClient
```

---

## Next Steps

1. ✅ Server deployed and running
2. ⏭️ Connect your first MCP client (n8n or Claude Desktop)
3. ⏭️ Set up monitoring and alerts
4. ⏭️ Configure TLS with reverse proxy (production)
5. ⏭️ Explore all 15+ available MCP tools
6. ⏭️ Build workflows that integrate Vikunja tasks

**Need help?** 
- Documentation: `/opt/vikunja/mcp-server/README.md`
- API Reference: `/opt/vikunja/specs/006-mcp-http-transport/contracts/`
- Logs: `journalctl -u vikunja-mcp -f`

**Enjoy using Vikunja MCP Server! 🚀**
