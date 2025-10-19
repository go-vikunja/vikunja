# Phase 0: Research & Technology Decisions

**Feature**: Proxmox LXC Automated Deployment  
**Date**: 2025-10-19  
**Status**: Complete

## Overview

This document consolidates research findings for implementing automated Vikunja deployment on Proxmox LXC containers. All technical unknowns from the planning phase have been resolved with concrete decisions.

---

## Research Tasks

### 1. Proxmox VE API Integration

**Question**: How to programmatically create and manage LXC containers via Proxmox API vs. direct CLI commands?

**Decision**: Use **Proxmox CLI tools (`pct`, `pvesh`)** instead of REST API

**Rationale**:
- CLI tools (`pct create`, `pct start`, `pct exec`) are simpler for Bash scripting
- No authentication complexity (runs as root on Proxmox host)
- More reliable for one-time deployments vs. continuous API polling
- Better error messages and debugging
- Direct access to container filesystem via `pct exec`

**Alternatives Considered**:
- **Proxmox REST API**: Requires token management, TLS certificates, more complex error handling. Better for web dashboards but overkill for CLI automation.
- **Ansible Proxmox modules**: Adds Python dependency and complexity. Not needed for single-script deployment.

**Implementation Notes**:
- Use `pct` for container lifecycle (create, start, stop, destroy)
- Use `pct exec` for running commands inside container
- Use `pvesh get /nodes/{node}/status` for resource checking
- Error handling: check exit codes, parse stderr for specific errors

---

### 2. Blue-Green Deployment Pattern for Zero Downtime

**Question**: How to achieve <5 seconds downtime during updates with health checks and rollback?

**Decision**: **Port-based blue-green with nginx upstream switching**

**Rationale**:
- Backend runs on two port sets: blue (e.g., 3456) and green (e.g., 3457)
- Nginx upstream block points to active port
- Update process: start new version on inactive port → health check → update nginx config → reload nginx (seamless)
- Nginx reload is zero-downtime (new workers accept connections, old workers drain)
- Rollback: point nginx back to old port and restart old service

**Alternatives Considered**:
- **Container cloning**: Too slow (>5 seconds), wastes resources
- **Haproxy**: Additional dependency, nginx already needed for SSL
- **DNS switching**: Too slow (DNS caching delays)

**Implementation Notes**:
```bash
# Ports allocation
BLUE_BACKEND_PORT=3456
GREEN_BACKEND_PORT=3457
BLUE_MCP_PORT=8456
GREEN_MCP_PORT=8457

# Nginx upstream switching
upstream backend {
    server 127.0.0.1:${ACTIVE_BACKEND_PORT};
}

# Update process
1. Determine inactive color (if blue active, deploy to green)
2. Start new backend/MCP on green ports
3. Health check green ports (HTTP 200, /health endpoint)
4. Update nginx config: ACTIVE_PORT=green
5. Reload nginx: systemctl reload nginx
6. Stop blue services (keep blue binaries for rollback)
7. Mark green as active in state file
```

**Rollback Strategy**:
```bash
1. Detect health check failure on green
2. Keep blue services running
3. Revert nginx config: ACTIVE_PORT=blue
4. Reload nginx
5. Stop green services
6. Log failure details
```

---

### 3. Database Migration Execution During Updates

**Question**: How to run Vikunja migrations safely with automatic backup and rollback?

**Decision**: **Use Vikunja's built-in migration system with pre-backup snapshots**

**Rationale**:
- Vikunja binary has `migrate` command: `./vikunja migrate`
- Migrations are idempotent and ordered by timestamp
- Built-in rollback tracking (migration table)
- No need to reimplement migration logic

**Alternatives Considered**:
- **Manual SQL execution**: Error-prone, breaks with Vikunja updates
- **Third-party tools (Flyway, Liquibase)**: Unnecessary duplication

**Implementation Notes**:
```bash
# Before update
1. Create database backup: 
   - SQLite: cp database.db backup/database.db.$(date +%s)
   - PostgreSQL: pg_dump -Fc -f backup/vikunja_$(date +%s).dump
   - MySQL: mysqldump --single-transaction > backup/vikunja_$(date +%s).sql

2. Test backup integrity:
   - SQLite: sqlite3 backup.db "PRAGMA integrity_check;"
   - PostgreSQL: pg_restore --list backup.dump
   - MySQL: mysql < backup.sql (dry-run to test DB)

3. Run migrations:
   - ./vikunja migrate
   - Check exit code ($? == 0)
   - Verify migration table updated

4. On failure:
   - Stop vikunja services
   - Restore database from backup
   - Rollback to previous version
   - Restart services
```

**Backup Strategy**:
- Keep last 5 backups (automatic cleanup)
- Store in `/var/backups/vikunja/` with timestamp
- Compress with gzip (except SQLite - already compressed)

---

### 4. Health Check Implementation

**Question**: What endpoints and checks are needed to verify all three components (backend, frontend, MCP) are healthy?

**Decision**: **Multi-layer health checks with HTTP endpoints and process monitoring**

**Rationale**:
- Backend: has `/health` endpoint returning JSON with database status
- Frontend: static files, verify nginx serves 200 on `/` 
- MCP: has health endpoint (standard Node.js pattern)
- Process checks: verify systemd service status

**Alternatives Considered**:
- **Ping only**: Insufficient, doesn't check app logic
- **Full e2e tests**: Too slow for update cycle (<5 min requirement)

**Implementation Notes**:
```bash
check_backend_health() {
    local port=$1
    local response=$(curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1:${port}/health)
    if [ "$response" = "200" ]; then
        # Verify JSON response contains "status": "ok"
        local health=$(curl -s http://127.0.0.1:${port}/health | jq -r '.status')
        [ "$health" = "ok" ]
    else
        return 1
    fi
}

check_frontend_health() {
    # Frontend served by nginx
    local response=$(curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1:80/)
    [ "$response" = "200" ]
}

check_mcp_health() {
    local port=$1
    local response=$(curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1:${port}/health)
    [ "$response" = "200" ]
}

check_database_connection() {
    # Test via backend /health endpoint (includes DB check)
    local health=$(curl -s http://127.0.0.1:${BACKEND_PORT}/health | jq -r '.database')
    [ "$health" = "connected" ]
}

full_health_check() {
    check_backend_health $ACTIVE_BACKEND_PORT && \
    check_frontend_health && \
    check_mcp_health $ACTIVE_MCP_PORT && \
    systemctl is-active vikunja-backend && \
    systemctl is-active vikunja-mcp
}
```

**Timeout Strategy**:
- Initial check: immediate (fail fast on obvious errors)
- Retry logic: 3 attempts with 2-second intervals (allows app startup)
- Total timeout: 10 seconds per component (30 seconds total)
- Update constraint: must complete health checks in <1 minute

---

### 5. Configuration File Format and Validation

**Question**: How to structure the YAML configuration file with validation and defaults?

**Decision**: **Single YAML file with environment-specific sections and schema validation**

**Rationale**:
- YAML is human-readable and editable
- Supports comments for documentation
- Standard format for infrastructure tools
- Easy to parse in Bash (using `yq` or `shyaml`)

**Alternatives Considered**:
- **JSON**: No comments, harder to edit manually
- **TOML**: Less common, requires additional parser
- **Env files**: Flat structure, no nesting, harder to organize

**Implementation Notes**:

```yaml
# /etc/vikunja/deployment-config.yaml
deployment:
  version: "1.0.0"
  instance_id: "vikunja-main"
  created_at: "2025-10-19T10:30:00Z"
  
proxmox:
  node: "pve"
  container_id: 100
  template: "debian-12-standard_12.0-1_amd64.tar.zst"
  
resources:
  cpu_cores: 2
  memory_mb: 4096
  disk_size_gb: 20
  
network:
  bridge: "vmbr0"
  ip_address: "192.168.1.100/24"
  gateway: "192.168.1.1"
  domain: "vikunja.example.com"
  
database:
  type: "postgresql"  # sqlite | postgresql | mysql
  # For external DB:
  host: "192.168.1.50"
  port: 5432
  name: "vikunja"
  user: "vikunja"
  password_file: "/etc/vikunja/secrets/db-password"  # Never store plain text
  
services:
  backend:
    port_blue: 3456
    port_green: 3457
    active_color: "blue"  # Updated during deployments
  mcp:
    port_blue: 8456
    port_green: 8457
    active_color: "blue"
  frontend:
    port: 80
    ssl_port: 443
    ssl_cert: "/etc/vikunja/ssl/cert.pem"
    ssl_key: "/etc/vikunja/ssl/key.pem"
    
git:
  repository: "https://github.com/go-vikunja/vikunja.git"
  branch: "main"
  deployed_commit: "abc123def456"  # Updated during deployments
  
backup:
  directory: "/var/backups/vikunja"
  retention_count: 5
  last_backup: "2025-10-19T09:00:00Z"
```

**Validation Rules**:
```bash
validate_config() {
    # Check required fields
    [ -n "$(yq e '.deployment.instance_id' config.yaml)" ] || error "Missing instance_id"
    
    # Validate database type
    local db_type=$(yq e '.database.type' config.yaml)
    [[ "$db_type" =~ ^(sqlite|postgresql|mysql)$ ]] || error "Invalid database type"
    
    # Validate network config
    validate_ip_address "$(yq e '.network.ip_address' config.yaml)"
    validate_domain "$(yq e '.network.domain' config.yaml)"
    
    # Validate resources
    local cpu=$(yq e '.resources.cpu_cores' config.yaml)
    [ "$cpu" -ge 1 ] && [ "$cpu" -le 32 ] || error "CPU cores must be 1-32"
    
    local mem=$(yq e '.resources.memory_mb' config.yaml)
    [ "$mem" -ge 2048 ] || error "Memory must be >= 2048 MB"
}
```

---

### 6. Lock File Mechanism for Concurrent Operations

**Question**: How to prevent concurrent deployments/updates from corrupting state?

**Decision**: **Atomic file-based locking with timeout and stale lock detection**

**Rationale**:
- Simple, no external dependencies
- Works across SSH sessions
- Atomic operations using `mkdir` (atomic in POSIX)
- Timeout prevents permanent locks from crashes

**Alternatives Considered**:
- **flock**: Not available on all systems, NFS issues
- **Redis/database locking**: Requires external service
- **PID files**: Can become stale, not atomic

**Implementation Notes**:
```bash
LOCK_DIR="/var/lock/vikunja-deploy"
LOCK_TIMEOUT=3600  # 1 hour

acquire_lock() {
    local max_attempts=30
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        if mkdir "$LOCK_DIR" 2>/dev/null; then
            # Lock acquired
            echo $$ > "$LOCK_DIR/pid"
            echo $(date +%s) > "$LOCK_DIR/timestamp"
            echo "$(whoami)@$(hostname)" > "$LOCK_DIR/owner"
            trap release_lock EXIT INT TERM
            return 0
        fi
        
        # Check if lock is stale
        if [ -f "$LOCK_DIR/timestamp" ]; then
            local lock_time=$(cat "$LOCK_DIR/timestamp")
            local now=$(date +%s)
            local age=$((now - lock_time))
            
            if [ $age -gt $LOCK_TIMEOUT ]; then
                log_warn "Removing stale lock (age: ${age}s)"
                release_lock_force
                continue
            fi
        fi
        
        log_info "Waiting for lock (attempt $attempt/$max_attempts)..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    # Show who holds the lock
    if [ -f "$LOCK_DIR/owner" ]; then
        local owner=$(cat "$LOCK_DIR/owner")
        error "Cannot acquire lock. Held by: $owner"
    fi
    error "Cannot acquire lock after $max_attempts attempts"
}

release_lock() {
    if [ -d "$LOCK_DIR" ]; then
        rm -rf "$LOCK_DIR"
    fi
}

release_lock_force() {
    # For stale lock cleanup
    rm -rf "$LOCK_DIR"
}
```

---

### 7. Systemd Service Configuration Best Practices

**Question**: What systemd service settings ensure reliable startup, restart, and dependencies?

**Decision**: **Type=simple with dependency ordering and automatic restart policies**

**Rationale**:
- `Type=simple`: Process runs in foreground (Vikunja default)
- Restart policies: always restart on failure (resilience)
- Dependencies: ensure network and database before backend
- Resource limits: prevent runaway processes

**Alternatives Considered**:
- **Type=forking**: Requires daemon mode, more complex
- **Type=notify**: Requires sd_notify support in app

**Implementation Notes**:

```ini
# /etc/systemd/system/vikunja-backend.service
[Unit]
Description=Vikunja Backend API Server
After=network-online.target postgresql.service mysql.service
Wants=network-online.target
PartOf=vikunja.target

[Service]
Type=simple
User=vikunja
Group=vikunja
WorkingDirectory=/opt/vikunja/backend

# Environment
Environment="VIKUNJA_SERVICE_ROOTPATH=/opt/vikunja/backend"
Environment="VIKUNJA_DATABASE_TYPE=postgresql"
EnvironmentFile=-/etc/vikunja/environment

# Execution
ExecStart=/opt/vikunja/backend/vikunja web
ExecReload=/bin/kill -HUP $MAINPID

# Restart policy
Restart=always
RestartSec=5s
StartLimitInterval=60s
StartLimitBurst=3

# Resource limits
MemoryMax=512M
MemoryHigh=384M
CPUQuota=200%

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/vikunja/backend/files /var/log/vikunja

[Install]
WantedBy=multi-user.target vikunja.target
```

**Service Dependencies**:
```bash
# Create target for grouped management
cat > /etc/systemd/system/vikunja.target <<EOF
[Unit]
Description=Vikunja Task Management Suite
Requires=vikunja-backend.service vikunja-mcp.service nginx.service
After=vikunja-backend.service vikunja-mcp.service nginx.service

[Install]
WantedBy=multi-user.target
EOF

# Commands
systemctl start vikunja.target    # Start all services
systemctl stop vikunja.target     # Stop all services
systemctl status vikunja.target   # Check all statuses
```

---

### 8. Nginx Reverse Proxy Configuration with SSL

**Question**: How to configure nginx for SSL termination, domain routing, and WebSocket support?

**Decision**: **Nginx with standard reverse proxy config, SSL termination, and WebSocket upgrades**

**Rationale**:
- Nginx is standard for reverse proxying
- Built-in SSL support
- Efficient static file serving (frontend)
- WebSocket support (needed for real-time features)

**Implementation Notes**:

```nginx
# /etc/nginx/sites-available/vikunja
upstream vikunja_backend {
    server 127.0.0.1:3456;  # Blue/green port updated dynamically
    keepalive 32;
}

upstream vikunja_mcp {
    server 127.0.0.1:8456;  # Blue/green port updated dynamically
}

server {
    listen 80;
    listen [::]:80;
    server_name vikunja.example.com;
    
    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name vikunja.example.com;
    
    # SSL Configuration
    ssl_certificate /etc/vikunja/ssl/cert.pem;
    ssl_certificate_key /etc/vikunja/ssl/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;
    
    # Security headers
    add_header Strict-Transport-Security "max-age=31536000" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    
    # Frontend static files
    location / {
        root /opt/vikunja/frontend/dist;
        try_files $uri $uri/ /index.html;
        
        # Caching
        expires 7d;
        add_header Cache-Control "public, immutable";
    }
    
    # API backend
    location /api/ {
        proxy_pass http://vikunja_backend;
        proxy_http_version 1.1;
        
        # Headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket support
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
    
    # MCP server
    location /mcp/ {
        proxy_pass http://vikunja_mcp/;
        proxy_http_version 1.1;
        
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # Health check endpoint (no auth)
    location /health {
        proxy_pass http://vikunja_backend/health;
        access_log off;
    }
    
    # File uploads (larger limits)
    client_max_body_size 100M;
}
```

**Dynamic Port Switching**:
```bash
update_nginx_upstream() {
    local active_color=$1  # "blue" or "green"
    local backend_port=${active_color}_BACKEND_PORT
    local mcp_port=${active_color}_MCP_PORT
    
    sed -i "s|server 127.0.0.1:[0-9]*;  # Backend|server 127.0.0.1:${!backend_port};  # Backend|" \
        /etc/nginx/sites-available/vikunja
    
    sed -i "s|server 127.0.0.1:[0-9]*;  # MCP|server 127.0.0.1:${!mcp_port};  # MCP|" \
        /etc/nginx/sites-available/vikunja
    
    nginx -t && systemctl reload nginx
}
```

---

## Technology Stack Summary

| Component | Technology | Version | Rationale |
|-----------|-----------|---------|-----------|
| **Scripting** | Bash | 4.0+ | Standard on Proxmox, no additional dependencies |
| **Container** | LXC | (Proxmox default) | Lightweight, fast startup, Proxmox native |
| **OS Template** | Debian | 12 (Bookworm) | Stable, long-term support, wide compatibility |
| **Reverse Proxy** | Nginx | 1.22+ | SSL termination, static files, WebSocket support |
| **Service Manager** | Systemd | 250+ | Standard on Debian 12, reliable service management |
| **Config Format** | YAML | 1.2 | Human-readable, standard format |
| **Config Parser** | yq | 4.x | YAML parsing in Bash scripts |
| **JSON Parser** | jq | 1.6+ | Health check response parsing |
| **Locking** | mkdir | (POSIX) | Atomic operations, no dependencies |
| **Deployment Pattern** | Blue-Green | N/A | Zero-downtime updates, fast rollback |

---

## Dependencies Installation

All dependencies installed automatically during deployment:

```bash
# Inside LXC container (Debian 12)
apt-get update
apt-get install -y \
    curl wget git \
    nginx \
    jq yq \
    sqlite3 postgresql-client mysql-client \
    build-essential \
    ca-certificates

# Go installation (for backend)
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# Node.js installation (for frontend build and MCP server)
# Note: Updated to Node.js 22 to match Vikunja frontend requirements
# - Frontend .nvmrc specifies: 22.18.0
# - Vite 7.1.10 requires: Node.js 20.19+ or 22.12+
# - Node.js 22 LTS until April 2027 (vs Node.js 18 EOL April 2025)
curl -fsSL https://deb.nodesource.com/setup_22.x | bash -
apt-get install -y nodejs

# Verify installations
go version    # go1.21.5
node --version  # v22.x
nginx -v      # nginx/1.22.x
```

### 6.7 User Registration Configuration

**Finding**: The `/api/v1/register` endpoint returns 404 (Unauthorized) when `VIKUNJA_SERVICE_ENABLEREGISTRATION` is not explicitly set.

**Root Cause**: 
- Backend checks `config.ServiceEnableRegistration.GetBool()` in `pkg/routes/api/v1/user_register.go:51`
- Returns `echo.ErrNotFound` if false
- Default value in `config-raw.json` is `"true"`, but environment variable must be set explicitly

**Solution**:
- Add `Environment="VIKUNJA_SERVICE_ENABLEREGISTRATION=true"` to backend systemd service
- This enables new user registration for fresh deployments
- Can be changed to `false` in production environments after initial admin setup

**Impact**: Without this setting, users cannot register accounts via the web interface.

### 6.8 Critical Configuration Variables Missing/Incorrect

**Finding**: Multiple critical configuration issues discovered:
1. Using non-existent `VIKUNJA_SERVICE_FRONTENDURL` instead of `VIKUNJA_SERVICE_PUBLICURL`
2. Missing `VIKUNJA_SERVICE_ROOTPATH` (binary location)
3. Missing database configuration (`VIKUNJA_DATABASE_TYPE`, `VIKUNJA_DATABASE_PATH`)

**Root Cause**:
1. **FrontendURL**: `pkg/routes/api/v1/info.go:92` shows `FrontendURL: config.ServicePublicURL.GetString()`
   - There is NO `ServiceFrontendURL` config key in Vikunja
   - Correct env var: `VIKUNJA_SERVICE_PUBLICURL` (maps to `service.publicurl`)
   - We invented a variable name that doesn't exist

2. **RootPath**: Binary needs working directory for assets, logs, config lookup
   - Config: `service.rootpath` → env: `VIKUNJA_SERVICE_ROOTPATH`
   - Default: `<rootpath>` (needs explicit setting)
   - Used by binary to find assets and config files

3. **Database**: SQLite needs explicit path to avoid current-directory issues
   - Config: `database.type` → env: `VIKUNJA_DATABASE_TYPE` (default: "sqlite")
   - Config: `database.path` → env: `VIKUNJA_DATABASE_PATH` (default: "./vikunja.db")
   - Relative path `./vikunja.db` depends on launch directory (unreliable)

**Solution**:
Set correct environment variables in backend systemd service:
```bash
Environment="VIKUNJA_SERVICE_PUBLICURL=http://192.168.50.64/"
Environment="VIKUNJA_SERVICE_INTERFACE=:3456"
Environment="VIKUNJA_SERVICE_ROOTPATH=/opt/vikunja"
Environment="VIKUNJA_SERVICE_ENABLEREGISTRATION=true"
Environment="VIKUNJA_DATABASE_TYPE=sqlite"
Environment="VIKUNJA_DATABASE_PATH=/opt/vikunja/vikunja.db"
```

**Key Details**:
- `PUBLICURL` must have trailing slash (config auto-adds if missing, line 589 of config.go)
- `ROOTPATH` tells binary where it lives (for asset loading)
- Database path should be absolute to prevent issues

**Impact**: 
- Without PUBLICURL: Frontend gets wrong API URL (was getting 127.0.0.1)
- Without ROOTPATH: Binary may not find assets or config files
- Without DATABASE_PATH: SQLite database created in unpredictable location

### 6.9 Database Configuration Not Passed to Service Generation

**Finding**: Main script collects database configuration for PostgreSQL/MySQL but doesn't pass it to systemd service generation. Service always gets SQLite configuration hardcoded.

**Root Cause**:
1. `generate_systemd_unit()` function signature only accepted 6 parameters (up to frontend_url)
2. Database configuration hardcoded in service template:
   ```bash
   Environment="VIKUNJA_DATABASE_TYPE=sqlite"
   Environment="VIKUNJA_DATABASE_PATH=${working_dir}/vikunja.db"
   ```
3. Main script calls `generate_systemd_unit()` without passing `DATABASE_*` variables
4. PostgreSQL/MySQL deployments would fail to connect to database

**Solution**:
1. **Extended function signature** to accept database parameters:
   ```bash
   generate_systemd_unit ct_id service_type color port working_dir frontend_url \
       db_type db_host db_port db_name db_user db_pass
   ```

2. **Conditional environment variable generation**:
   - SQLite: Set `VIKUNJA_DATABASE_PATH=${working_dir}/vikunja.db`
   - PostgreSQL/MySQL: Set HOST, PORT, DATABASE, USER, PASSWORD

3. **Updated main script** to pass database configuration:
   ```bash
   generate_systemd_unit "$CONTAINER_ID" "backend" "blue" \
       "$BACKEND_PORT_BLUE" "$WORKING_DIR" "$frontend_url" \
       "$DATABASE_TYPE" "$DATABASE_HOST" "$DATABASE_PORT" \
       "$DATABASE_NAME" "$DATABASE_USER" "$DATABASE_PASS"
   ```

4. **Default port assignment** in validation section:
   - PostgreSQL: 5432 (if not provided)
   - MySQL: 3306 (if not provided)

**Environment Variables by Database Type**:

**SQLite**:
```bash
VIKUNJA_DATABASE_TYPE=sqlite
VIKUNJA_DATABASE_PATH=/opt/vikunja/vikunja.db
```

**PostgreSQL**:
```bash
VIKUNJA_DATABASE_TYPE=postgresql
VIKUNJA_DATABASE_HOST=db.example.com
VIKUNJA_DATABASE_PORT=5432
VIKUNJA_DATABASE_DATABASE=vikunja
VIKUNJA_DATABASE_USER=vikunja_user
VIKUNJA_DATABASE_PASSWORD=secure_password
```

**MySQL**:
```bash
VIKUNJA_DATABASE_TYPE=mysql
VIKUNJA_DATABASE_HOST=db.example.com
VIKUNJA_DATABASE_PORT=3306
VIKUNJA_DATABASE_DATABASE=vikunja
VIKUNJA_DATABASE_USER=vikunja_user
VIKUNJA_DATABASE_PASSWORD=secure_password
```

**Impact**: Without this fix, PostgreSQL/MySQL deployments would always try to use SQLite, causing database connection failures and data loss.

**Files Modified**:
- `deploy/proxmox/lib/service-setup.sh` (lines 20-60)
- `deploy/proxmox/vikunja-install-main.sh` (lines 360-385, 630-650)

**Verification**: T042R0 (pre-testing code review), T042R17-T042R21 (database configuration tests)

### 6.10 Service Enable/Start Functions Fail on Success

**Finding**: The `enable_service` function reports failure even though `systemctl enable` succeeds and creates the symlink. Service file is correct but deployment stops at "Failed to enable backend service".

**Root Cause**:
The `tee >(log_debug)` process substitution in the service management functions was causing false failures:
```bash
pct_exec "$ct_id" systemctl enable "$service_name" 2>&1 | tee >(log_debug) || return 1
```

Even though `systemctl enable` succeeded (symlink created), the pipe with `tee >(log_debug)` was returning non-zero exit code, causing the function to return failure.

**Evidence**:
```
✓ Unit file created: /etc/systemd/system/vikunja-backend-blue.service
Created symlink /etc/systemd/system/multi-user.target.wants/vikunja-backend-blue.service → ...
✗ Failed to enable backend service
```

The symlink message proves systemctl succeeded, but the function reported failure.

**Solution**:
Replaced piped output handling with command substitution and proper error checking:
```bash
# Before (BROKEN):
pct_exec "$ct_id" systemctl enable "$service_name" 2>&1 | tee >(log_debug) || return 1

# After (FIXED):
local output
if ! output=$(pct_exec "$ct_id" systemctl enable "$service_name" 2>&1); then
    log_error "Failed to enable service: ${output}"
    return 1
fi
[[ -n "$output" ]] && log_debug "$output"
```

**Functions Fixed** (complete list):
- `enable_service()` - daemon-reload and enable operations
- `start_service()` - start operation and status checking
- `stop_service()` - stop operation
- `restart_service()` - restart operation
- `graceful_restart_backend()` - graceful backend restart
- `graceful_restart_mcp()` - graceful MCP restart
- `reload_nginx_config()` - nginx config reload (duplicate of reload_nginx)
- `reload_nginx()` - nginx configuration test and reload
- `enable_site()` - nginx site symlink creation
- `update_nginx_upstream()` - nginx configuration update

**Files Modified**:
- `deploy/proxmox/lib/service-setup.sh` (10 function fixes)
- `deploy/proxmox/lib/nginx-setup.sh` (4 function fixes)

**Note**: The proxmox-api.sh file still has `tee >(log_debug)` calls but these are for `pct` commands run on the Proxmox host (not inside containers), which may behave differently. We'll monitor if they cause issues.

**Impact**: Without this fix, deployment fails at multiple steps (service enable, service start, nginx reload) even though the actual systemctl/nginx operations succeed.

---

## Next Steps

✅ All research complete - proceed to **Phase 1: Design & Contracts**

Phase 1 will generate:
1. `data-model.md` - State tracking entities
2. `contracts/` - Script interfaces and CLI commands
3. `quickstart.md` - User-facing installation guide
4. Agent context update (Copilot memory)
