# Deployment Guide

This guide covers deploying the Vikunja MCP Server in various environments.

## Table of Contents
- [Quick Start (Docker)](#quick-start-docker)
- [Production Deployment](#production-deployment)
- [Proxmox LXC Container](#proxmox-lxc-container)
- [Systemd Service](#systemd-service)
- [Monitoring & Maintenance](#monitoring--maintenance)

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

### Create LXC Container

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

# Install Node.js 20 LTS
curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt install -y nodejs

# Install Redis (or use external Redis)
apt install -y redis-server

# Install Git
apt install -y git

# Create MCP user
useradd -r -m -s /bin/bash vikunja-mcp
```

### Deploy MCP Server

```bash
# Switch to MCP user
su - vikunja-mcp

# Clone/copy project
cd /opt
git clone <your-repo> vikunja-mcp-server
cd vikunja-mcp-server

# Install dependencies
npm ci --only=production

# Build TypeScript
npm run build

# Create config directory
sudo mkdir -p /etc/vikunja-mcp
sudo chown vikunja-mcp:vikunja-mcp /etc/vikunja-mcp

# Create environment file
nano /etc/vikunja-mcp/config.env
```

Add configuration:
```env
VIKUNJA_API_URL=http://192.168.1.100:3456
VIKUNJA_API_TOKEN=your-token-here
MCP_PORT=3457
REDIS_HOST=localhost
REDIS_PORT=6379
RATE_LIMIT_DEFAULT=100
RATE_LIMIT_BURST=120
LOG_LEVEL=info
```

---

## Systemd Service

### Create Service File

```bash
sudo nano /etc/systemd/system/vikunja-mcp.service
```

Add configuration:
```ini
[Unit]
Description=Vikunja MCP Server
After=network.target redis.service
Wants=redis.service

[Service]
Type=simple
User=vikunja-mcp
Group=vikunja-mcp
WorkingDirectory=/opt/vikunja-mcp-server
EnvironmentFile=/etc/vikunja-mcp/config.env
ExecStart=/usr/bin/node /opt/vikunja-mcp-server/dist/index.js
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
