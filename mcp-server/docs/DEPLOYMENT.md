# Deployment Guide

This guide covers deploying the Vikunja MCP Server in various environments.

## Table of Contents
- [Authentication Overview](#authentication-overview)
- [HTTP Transport](#http-transport)
- [Quick Start (Docker)](#quick-start-docker)
- [Production Deployment](#production-deployment)
- [Proxmox LXC Container](#proxmox-lxc-container)
- [Systemd Service](#systemd-service)
- [Monitoring & Maintenance](#monitoring--maintenance)

---

## Authentication Overview

The Vikunja MCP Server supports two authentication methods:

### Method 1: Per-Request Token (Recommended)

**Best for:** Claude Desktop, n8n, and other AI agents where each user has their own Vikunja API token.

The AI agent passes the Vikunja API token with each MCP request via environment variables. This is the **preferred and most secure method** as it allows:
- Multiple users with different permissions
- Token rotation without server restart
- Fine-grained access control per user

**Configuration (Claude Desktop example):**
```json
{
  "mcpServers": {
    "vikunja": {
      "command": "node",
      "args": ["/path/to/vikunja-mcp-server/dist/index.js"],
      "env": {
        "VIKUNJA_API_URL": "http://localhost:3456",
        "VIKUNJA_API_TOKEN": "your-personal-vikunja-token-here",
        "MCP_PORT": "3457",
        "REDIS_HOST": "localhost",
        "REDIS_PORT": "6379"
      }
    }
  }
}
```

The MCP server validates the token on each request and caches user information for 5 minutes.

### Method 2: Server-Level Token

**Best for:** Single-user deployments, testing, or when all requests should use the same Vikunja account.

⚠️ **Note:** The current Proxmox deployment scripts use this method by default since they run as a system service. All MCP requests will use the same Vikunja account.

**When MCP server is deployed as a systemd service:**
- The server runs continuously without per-request tokens
- All API requests are made **unauthenticated** unless you configure a default token
- Client applications must either:
  1. Pass their own `VIKUNJA_API_TOKEN` when invoking the MCP server, OR
  2. Use a shared token configured in the systemd service environment

**Important:** The systemd service configuration in Proxmox deployments does NOT currently set a default `VIKUNJA_API_TOKEN`. You must configure your MCP clients (Claude Desktop, n8n, etc.) to pass the token via environment variables (Method 1 above).

### Why No Default Token in Systemd Service?

The deployment intentionally omits `VIKUNJA_API_TOKEN` from the systemd service to:
1. **Prevent unprotected access** - No default token means no unauthenticated requests succeed
2. **Support multi-user scenarios** - Each client can use their own token
3. **Security best practice** - Tokens should not be stored in system service files

**To add a default token** (single-user deployments only):

Edit `/etc/systemd/system/vikunja-mcp-blue.service` and add:
```ini
Environment="VIKUNJA_API_TOKEN=your-shared-token-here"
```

Then reload and restart:
```bash
systemctl daemon-reload
systemctl restart vikunja-mcp-blue
```

---

## HTTP Transport

The MCP server supports two transport mechanisms for connecting AI agents and automation tools:

### Transport Types

1. **Stdio Transport** (Default) - Traditional stdin/stdout communication
   - **Best for:** Claude Desktop, direct process spawning
   - **Pros:** Simple, secure (no network exposure), low overhead
   - **Cons:** One client per process, requires subprocess management

2. **HTTP/SSE Transport** - Server-Sent Events over HTTP
   - **Best for:** n8n workflows, Python MCP SDK clients, web-based AI agents
   - **Pros:** Multiple concurrent clients, network-accessible, RESTful architecture
   - **Cons:** Requires port exposure, network security considerations

### Configuration

Set the transport type using environment variables:

```bash
# Stdio Transport (default) - no additional configuration needed
TRANSPORT_TYPE=stdio

# HTTP/SSE Transport - requires MCP port configuration
TRANSPORT_TYPE=http
MCP_PORT=3010  # Port for HTTP/SSE endpoint (required for HTTP)

# Optional: CORS configuration for browser-based clients
CORS_ENABLED=true
CORS_ALLOWED_ORIGINS=https://app.example.com,https://n8n.example.com
```

### HTTP/SSE Endpoint

When configured with `TRANSPORT_TYPE=http`, the MCP server exposes:

**Endpoint:** `POST /sse`

**Authentication:** Vikunja API token (two methods)
1. **Authorization header (recommended):**
   ```http
   POST /sse HTTP/1.1
   Host: mcp.example.com:3010
   Authorization: Bearer YOUR_VIKUNJA_API_TOKEN
   ```

2. **Query parameter (fallback):**
   ```http
   POST /sse?token=YOUR_VIKUNJA_API_TOKEN HTTP/1.1
   Host: mcp.example.com:3010
   ```

**Response:** Server-Sent Events (SSE) stream with JSON-RPC 2.0 messages

### Client Examples

#### n8n Workflow

```javascript
// In n8n HTTP Request node:
{
  "method": "POST",
  "url": "http://mcp.example.com:3010/sse",
  "headers": {
    "Authorization": "Bearer {{$node[\"Get Vikunja Token\"].json[\"token\"]}}"
  },
  "options": {
    "responseType": "stream"
  }
}

// Handle SSE events in subsequent nodes
// Each event contains MCP protocol messages
```

#### Python MCP SDK

```python
from mcp import ClientSession, StdioServerParameters
from mcp.client.sse import sse_client
import httpx

async def connect_to_mcp():
    vikunja_token = "your-vikunja-api-token"
    
    # Create SSE client with authentication
    async with httpx.AsyncClient() as http_client:
        async with sse_client(
            http_client=http_client,
            url="http://mcp.example.com:3010/sse",
            headers={"Authorization": f"Bearer {vikunja_token}"}
        ) as (read, write):
            async with ClientSession(read, write) as session:
                # Initialize MCP session
                await session.initialize()
                
                # List available tools
                tools = await session.list_tools()
                print(f"Available tools: {tools}")
                
                # Execute create_task tool
                result = await session.call_tool(
                    "vikunja_create_task",
                    {
                        "title": "Task from Python",
                        "project_id": 1
                    }
                )
                print(f"Task created: {result}")

# Run the async function
import asyncio
asyncio.run(connect_to_mcp())
```

#### curl (Manual Testing)

```bash
# Test connection (expect 401 without token)
curl -N -X POST http://mcp.example.com:3010/sse

# Establish authenticated SSE connection
curl -N -X POST \
  -H "Authorization: Bearer YOUR_VIKUNJA_API_TOKEN" \
  http://mcp.example.com:3010/sse

# Expected SSE stream output:
# event: connected
# data: {"connectionId":"550e8400-e29b-41d4-a716-446655440000","userId":1}
#
# (connection stays open for MCP messages)
```

### Proxmox Deployment with HTTP Transport

The automated Proxmox deployment configures HTTP transport by default (see [Proxmox LXC Container](#proxmox-lxc-container) section).

**Ports:**
- **Blue environment:** 3010 (MCP HTTP), 3456 (Vikunja API), 80/443 (Vikunja Frontend)
- **Green environment:** 3011 (MCP HTTP), 3456 (Vikunja API), 80/443 (Vikunja Frontend)

**Health Check:**
```bash
# MCP HTTP endpoint (expect 401 - proves server running)
curl -i http://mcp.example.com:3010/sse

# With valid token (expect SSE stream)
curl -N -H "Authorization: Bearer YOUR_TOKEN" http://mcp.example.com:3010/sse
```

**Systemd service automatically sets:**
```ini
Environment="TRANSPORT_TYPE=http"
Environment="MCP_PORT=3010"  # or 3011 for green
```

### Troubleshooting HTTP Transport

#### Connection Refused
```bash
# Check if MCP server is running
systemctl status vikunja-mcp-blue

# Verify port is listening
ss -tulpn | grep 3010

# Check firewall rules
ufw status | grep 3010
```

#### 401 Unauthorized
```bash
# Test token validity directly with Vikunja API
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://vikunja.example.com/api/v1/user

# Check MCP server logs for auth errors
journalctl -u vikunja-mcp-blue -n 50 --no-pager
```

#### SSE Connection Drops
```bash
# Check for network/proxy timeouts (SSE connections are long-lived)
# Configure reverse proxy timeout: proxy_read_timeout 3600s;

# Monitor active SSE connections
ss -tn | grep 3010 | wc -l

# Check server logs for connection close events
journalctl -u vikunja-mcp-blue -f | grep "SSE connection"
```

#### CORS Errors (Browser Clients)
```bash
# Enable CORS in MCP server environment
Environment="CORS_ENABLED=true"
Environment="CORS_ALLOWED_ORIGINS=https://your-app.com"

# Restart service
systemctl restart vikunja-mcp-blue
```

### Security Considerations

1. **Network Exposure:** HTTP transport exposes MCP server on the network
   - Use firewall rules to restrict access (e.g., only from n8n server IP)
   - Consider VPN or private network for sensitive deployments

2. **Authentication:** Per-request token validation ensures security
   - Tokens are validated on every SSE connection establishment
   - User context is cached for 5 minutes (Redis)
   - Token rotation is supported (no server restart needed)

3. **HTTPS/TLS:** For production deployments
   - Use reverse proxy (nginx, Caddy) with TLS termination
   - Let's Encrypt certificates recommended
   - Example nginx config:
     ```nginx
     server {
         listen 443 ssl http2;
         server_name mcp.example.com;
         
         ssl_certificate /etc/letsencrypt/live/mcp.example.com/fullchain.pem;
         ssl_certificate_key /etc/letsencrypt/live/mcp.example.com/privkey.pem;
         
         location /sse {
             proxy_pass http://127.0.0.1:3010;
             proxy_http_version 1.1;
             proxy_set_header Connection "";
             proxy_buffering off;
             proxy_cache off;
             proxy_read_timeout 86400s;  # 24-hour SSE connection
         }
     }
     ```

4. **Rate Limiting:** Consider implementing rate limits at reverse proxy level
   - Protect against connection flooding
   - Example nginx: `limit_req_zone $binary_remote_addr zone=mcp:10m rate=10r/s;`

---

## Quick Start (Docker)

The fastest way to get started is with Docker Compose.

### Prerequisites
````
```

---

## Quick Start (Docker)

The fastest way to get started is with Docker Compose.

### Prerequisites
- Docker & Docker Compose installed
- Vikunja instance running
- Valid Vikunja API token

### Steps

1. **Clone or copy the MCP server files:**
   ```bash
   cd /opt
   git clone <your-repo> vikunja-mcp-server
   cd vikunja-mcp-server
   ```

2. **Create environment file:**
   ```bash
   cp .env.example .env
   nano .env
   ```

   Configure your settings:
   ```env
   VIKUNJA_API_URL=http://192.168.1.100:3456
   VIKUNJA_API_TOKEN=your-token-here
   MCP_PORT=3457
   REDIS_HOST=redis
   REDIS_PORT=6379
   RATE_LIMIT_DEFAULT=100
   RATE_LIMIT_BURST=120
   LOG_LEVEL=info
   ```

3. **Start services:**
   ```bash
   docker-compose up -d
   ```

4. **Verify:**
   ```bash
   # Check health
   curl http://localhost:3457/health
   
   # Check logs
   docker logs vikunja-mcp-server
   ```

5. **Configure AI agent** (see [INTEGRATIONS.md](./INTEGRATIONS.md))

---

## Production Deployment

### Architecture

```
┌─────────────┐
│  AI Agents  │
│ (Claude/n8n)│
└──────┬──────┘
       │
       ▼
┌─────────────────┐      ┌──────────┐      ┌──────────┐
│  MCP Server     │─────▶│  Redis   │      │ Vikunja  │
│  (Port 3457)    │      │  (6379)  │      │  (3456)  │
└─────────────────┘      └──────────┘      └──────────┘
```

### Deployment Checklist

- [ ] **Security**
  - [ ] Use non-root user
  - [ ] Configure firewall rules
  - [ ] Enable TLS/SSL (if exposing publicly)
  - [ ] Rotate API tokens regularly
  - [ ] Set proper rate limits

- [ ] **Performance**
  - [ ] Configure Redis persistence
  - [ ] Set appropriate rate limits
  - [ ] Monitor resource usage
  - [ ] Enable connection pooling

- [ ] **Reliability**
  - [ ] Configure systemd/Docker restart policies
  - [ ] Set up health checks
  - [ ] Configure log rotation
  - [ ] Implement backup strategy

- [ ] **Monitoring**
  - [ ] Set up logging aggregation
  - [ ] Configure metrics collection
  - [ ] Set up alerting
  - [ ] Create dashboard

### Environment-Specific Configurations

#### Development
```env
LOG_LEVEL=debug
RATE_LIMIT_DEFAULT=1000
RATE_LIMIT_BURST=1500
```

#### Staging
```env
LOG_LEVEL=info
RATE_LIMIT_DEFAULT=500
RATE_LIMIT_BURST=750
```

#### Production
```env
LOG_LEVEL=warn
RATE_LIMIT_DEFAULT=100
RATE_LIMIT_BURST=120
LOG_FORMAT=json
```

---

## Proxmox LXC Container

For Proxmox environments, deploy as a lightweight LXC container.

### Automated Deployment (Recommended)

The Vikunja project includes automated Proxmox deployment scripts that handle complete installation including the MCP server:

```bash
# Clone the Vikunja repository
git clone https://github.com/aroige/vikunja.git
cd vikunja/deploy/proxmox

# Run the automated deployment
./vikunja-install.sh
```

The automated script will:
- Create and configure an LXC container
- Install Vikunja backend and frontend
- Install Node.js 22 and Redis
- Build and deploy the MCP server
- Configure systemd services with blue-green deployment support
- Set up Nginx reverse proxy

**Important - MCP Authentication After Automated Deployment:**

The automated deployment configures the MCP server systemd service with:
```ini
Environment="VIKUNJA_API_URL=http://127.0.0.1:3456"
Environment="REDIS_HOST=localhost"
Environment="REDIS_PORT=6379"
Environment="MCP_PORT=8456"  # Blue deployment, Green uses 8457
```

**Note:** `VIKUNJA_API_TOKEN` is intentionally NOT set in the systemd service. This means:
1. You must configure your MCP clients (Claude Desktop, n8n, etc.) to pass the token via their configuration
2. Each client can use different tokens for multi-user access
3. To set a shared default token, edit `/etc/systemd/system/vikunja-mcp-blue.service` and add `Environment="VIKUNJA_API_TOKEN=your-token"`

**Connecting to the Deployed MCP Server:**

After deployment, configure your AI agent client to connect. Example for Claude Desktop:

```json
{
  "mcpServers": {
    "vikunja": {
      "command": "ssh",
      "args": [
        "root@<container-ip>",
        "node", "/opt/vikunja/mcp-server/dist/index.js"
      ],
      "env": {
        "VIKUNJA_API_URL": "http://127.0.0.1:3456",
        "VIKUNJA_API_TOKEN": "your-personal-vikunja-token",
        "MCP_PORT": "8456",
        "REDIS_HOST": "localhost",
        "REDIS_PORT": "6379"
      }
    }
  }
}
```

### Manual Deployment

If you prefer manual deployment:

1. **Create Debian 12 container:**
   ```bash
   # In Proxmox shell
   pct create 200 local:vztmpl/debian-12-standard_12.2-1_amd64.tar.zst \
     --hostname vikunja-mcp \
     --memory 2048 \
     --cores 2 \
     --storage local-lvm \
     --rootfs 8 \
     --net0 name=eth0,bridge=vmbr0,ip=dhcp \
     --unprivileged 1 \
     --features nesting=1
   
   # Start container
   pct start 200
   ```

2. **Enter container:**
   ```bash
   pct enter 200
   ```

### Install Dependencies

```bash
# Update system
apt update && apt upgrade -y

# Install Node.js 20 LTS (or 22+ for latest)
curl -fsSL https://deb.nodesource.com/setup_22.x | bash -
apt install -y nodejs

# Install Redis (required for rate limiting)
apt install -y redis-server redis-tools

# Install Git
apt install -y git

# Verify Redis is running
systemctl status redis-server
```

### Deploy MCP Server

```bash
# Clone/copy project
cd /opt
git clone <your-repo> vikunja-mcp-server
cd vikunja-mcp-server

# Install dependencies
npm ci --only=production

# Build TypeScript
npm run build

# Verify build succeeded
ls -la dist/index.js
```

---

## Systemd Service

### Automated Deployment Service File

The Proxmox automated deployment creates a systemd service like this:

```ini
[Unit]
Description=Vikunja MCP Server (blue)
After=network.target redis-server.service
Wants=redis-server.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/vikunja/mcp-server
ExecStart=/usr/bin/node /opt/vikunja/mcp-server/dist/index.js
Environment="MCP_PORT=8456"
Environment="VIKUNJA_API_URL=http://127.0.0.1:3456"
Environment="REDIS_HOST=localhost"
Environment="REDIS_PORT=6379"
Environment="RATE_LIMIT_DEFAULT=100"
Environment="RATE_LIMIT_BURST=120"
Environment="LOG_LEVEL=info"
Environment="LOG_FORMAT=json"
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

**Note:** The service does NOT include `VIKUNJA_API_TOKEN`. See [Authentication Overview](#authentication-overview) for details.

### Manual Service File Creation

For manual deployments with a dedicated user:

```bash
sudo nano /etc/systemd/system/vikunja-mcp.service
```

Add configuration:
```ini
[Unit]
Description=Vikunja MCP Server
After=network.target redis-server.service
Wants=redis-server.service

[Service]
Type=simple
User=vikunja-mcp
Group=vikunja-mcp
WorkingDirectory=/opt/vikunja-mcp-server
ExecStart=/usr/bin/node /opt/vikunja-mcp-server/dist/index.js
Environment="MCP_PORT=3457"
Environment="VIKUNJA_API_URL=http://localhost:3456"
Environment="REDIS_HOST=localhost"
Environment="REDIS_PORT=6379"
Environment="RATE_LIMIT_DEFAULT=100"
Environment="RATE_LIMIT_BURST=120"
Environment="LOG_LEVEL=info"
Environment="LOG_FORMAT=json"
# Optional: Add default token for single-user deployments
# Environment="VIKUNJA_API_TOKEN=your-shared-token-here"
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=vikunja-mcp

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/vikunja-mcp

[Install]
WantedBy=multi-user.target
```

### Enable and Start

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service
sudo systemctl enable vikunja-mcp

# Start service
sudo systemctl start vikunja-mcp

# Check status
sudo systemctl status vikunja-mcp

# View logs
sudo journalctl -u vikunja-mcp -f
```

### Service Management

```bash
# Start
sudo systemctl start vikunja-mcp

# Stop
sudo systemctl stop vikunja-mcp

# Restart
sudo systemctl restart vikunja-mcp

# Status
sudo systemctl status vikunja-mcp

# Logs
sudo journalctl -u vikunja-mcp -f

# Last 100 lines
sudo journalctl -u vikunja-mcp -n 100

# Follow errors only
sudo journalctl -u vikunja-mcp -p err -f
```

---

## Monitoring & Maintenance

### Health Checks

```bash
# Basic health check
curl http://localhost:3457/health

# Expected response
{"status":"healthy","timestamp":"2025-10-17T12:00:00.000Z"}

# Check Vikunja connection
curl -H "Authorization: Bearer your-token" \
  http://localhost:3456/api/v1/info

# Check Redis
redis-cli ping
# Expected: PONG
```

### Log Management

#### Configure Log Rotation

```bash
sudo nano /etc/logrotate.d/vikunja-mcp
```

Add:
```
/var/log/vikunja-mcp/*.log {
    daily
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 vikunja-mcp vikunja-mcp
    sharedscripts
    postrotate
        systemctl reload vikunja-mcp > /dev/null 2>&1 || true
    endscript
}
```

#### View Logs

```bash
# Systemd journal
sudo journalctl -u vikunja-mcp -f

# Application logs (if file logging enabled)
tail -f /var/log/vikunja-mcp/app.log

# Docker logs
docker logs -f vikunja-mcp-server
```

### Performance Monitoring

#### Monitor Resource Usage

```bash
# Container stats (Docker)
docker stats vikunja-mcp-server

# Process stats (systemd)
systemctl status vikunja-mcp
ps aux | grep vikunja-mcp

# Memory usage
free -h

# Disk usage
df -h
```

#### Redis Monitoring

```bash
# Redis info
redis-cli info

# Monitor commands
redis-cli monitor

# Check rate limit keys
redis-cli --scan --pattern "ratelimit:*" | wc -l

# Get key details
redis-cli ttl "ratelimit:your-token"
```

### Backup & Restore

#### Backup Configuration

```bash
# Backup script
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backup/vikunja-mcp"

# Create backup directory
mkdir -p $BACKUP_DIR

# Backup environment config
cp /etc/vikunja-mcp/config.env $BACKUP_DIR/config.env.$DATE

# Backup Redis data (if local)
redis-cli save
cp /var/lib/redis/dump.rdb $BACKUP_DIR/redis_dump.$DATE.rdb

echo "Backup completed: $BACKUP_DIR/$DATE"
```

#### Restore Configuration

```bash
# Restore environment
cp /backup/vikunja-mcp/config.env.20251017_120000 /etc/vikunja-mcp/config.env

# Restore Redis data
sudo systemctl stop redis
cp /backup/vikunja-mcp/redis_dump.20251017_120000.rdb /var/lib/redis/dump.rdb
sudo systemctl start redis

# Restart MCP server
sudo systemctl restart vikunja-mcp
```

### Upgrading

#### Docker Deployment

```bash
# Pull latest image
docker-compose pull

# Restart with new image
docker-compose up -d

# Verify
docker logs vikunja-mcp-server
```

#### Systemd Deployment

```bash
# Stop service
sudo systemctl stop vikunja-mcp

# Backup current version
cd /opt
sudo cp -r vikunja-mcp-server vikunja-mcp-server.backup

# Update code
cd vikunja-mcp-server
git pull

# Install dependencies
npm ci --only=production

# Build
npm run build

# Start service
sudo systemctl start vikunja-mcp

# Check logs
sudo journalctl -u vikunja-mcp -f
```

### Troubleshooting

#### Service Won't Start

1. **Check logs:**
   ```bash
   sudo journalctl -u vikunja-mcp -n 50
   ```

2. **Verify configuration:**
   ```bash
   cat /etc/vikunja-mcp/config.env
   ```

3. **Test manually:**
   ```bash
   sudo -u vikunja-mcp /usr/bin/node /opt/vikunja-mcp-server/dist/index.js
   ```

4. **Check permissions:**
   ```bash
   ls -la /opt/vikunja-mcp-server/
   ls -la /etc/vikunja-mcp/
   ```

#### High Memory Usage

1. **Check Node.js heap:**
   ```bash
   NODE_OPTIONS="--max-old-space-size=512" node dist/index.js
   ```

2. **Monitor Redis memory:**
   ```bash
   redis-cli info memory
   ```

3. **Check for memory leaks:**
   ```bash
   # Enable heapdump
   npm install heapdump
   # Analyze with Chrome DevTools
   ```

#### Rate Limiting Issues

1. **Check Redis connection:**
   ```bash
   redis-cli -h localhost -p 6379 ping
   ```

2. **View rate limit data:**
   ```bash
   redis-cli keys "ratelimit:*"
   redis-cli get "ratelimit:your-token"
   ```

3. **Clear rate limits:**
   ```bash
   redis-cli del "ratelimit:your-token"
   ```

4. **Adjust limits in config:**
   ```env
   RATE_LIMIT_DEFAULT=500
   RATE_LIMIT_BURST=750
   ```

---

## Security Hardening

### Network Security

```bash
# Configure firewall (ufw)
sudo ufw allow 3457/tcp comment 'MCP Server'
sudo ufw enable

# Or iptables
sudo iptables -A INPUT -p tcp --dport 3457 -j ACCEPT
sudo iptables-save > /etc/iptables/rules.v4
```

### TLS/SSL (Production)

Use a reverse proxy (nginx/caddy) for TLS:

```nginx
# /etc/nginx/sites-available/vikunja-mcp
server {
    listen 443 ssl http2;
    server_name mcp.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/mcp.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/mcp.yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://localhost:3457;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

### Token Management

```bash
# Rotate tokens regularly
# 1. Generate new token in Vikunja
# 2. Update config
sudo nano /etc/vikunja-mcp/config.env
# 3. Restart service
sudo systemctl restart vikunja-mcp
```

---

## Next Steps

- **Integration**: See [INTEGRATIONS.md](./INTEGRATIONS.md) for AI agent setup
- **API Reference**: See [API.md](./API.md) for available tools
- **Examples**: See [EXAMPLES.md](./EXAMPLES.md) for usage patterns
