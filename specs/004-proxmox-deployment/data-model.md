# Data Model: Proxmox LXC Deployment

**Feature**: Proxmox LXC Automated Deployment  
**Date**: 2025-10-19  
**Purpose**: Define state tracking and configuration entities

## Overview

This data model describes the configuration and state entities managed by the Proxmox deployment system. These are stored as YAML files and state tracking files within the deployed container and Proxmox host.

---

## Entity: Deployment Configuration

**Purpose**: Persistent configuration for a Vikunja instance deployment

**Storage**: `/etc/vikunja/deployment-config.yaml` (inside container)

**Schema**:

```yaml
deployment:
  version: string              # Config schema version (e.g., "1.0.0")
  instance_id: string          # Unique identifier (e.g., "vikunja-main")
  created_at: datetime         # ISO 8601 timestamp
  updated_at: datetime         # Last modification timestamp
  
proxmox:
  node: string                 # Proxmox node name (e.g., "pve")
  container_id: integer        # LXC container ID (100-999)
  template: string             # Container template name
  
resources:
  cpu_cores: integer           # Allocated CPU cores (1-32)
  memory_mb: integer           # RAM in megabytes (min 2048)
  disk_size_gb: integer        # Disk size in GB (min 20)
  
network:
  bridge: string               # Proxmox bridge (e.g., "vmbr0")
  ip_address: string           # Static IP with CIDR (e.g., "192.168.1.100/24")
  gateway: string              # Gateway IP
  domain: string               # Public domain name
  
database:
  type: enum                   # "sqlite" | "postgresql" | "mysql"
  host: string?                # External DB host (null for SQLite)
  port: integer?               # External DB port
  name: string                 # Database name
  user: string?                # DB username (not for SQLite)
  password_file: string?       # Path to password file (not stored in config)
  
services:
  backend:
    port_blue: integer         # Blue backend port (default: 3456)
    port_green: integer        # Green backend port (default: 3457)
    active_color: enum         # "blue" | "green"
    version_blue: string?      # Semantic version on blue
    version_green: string?     # Semantic version on green
  mcp:
    port_blue: integer         # Blue MCP port (default: 8456)
    port_green: integer        # Green MCP port (default: 8457)
    active_color: enum         # "blue" | "green"
    version_blue: string?      # Version on blue
    version_green: string?     # Version on green
  frontend:
    port: integer              # HTTP port (default: 80)
    ssl_port: integer          # HTTPS port (default: 443)
    ssl_cert: string           # Path to SSL certificate
    ssl_key: string            # Path to SSL private key
    
git:
  repository: string           # Git repository URL
  branch: string               # Deployment branch (default: "main")
  deployed_commit: string      # Current deployed commit hash
  
backup:
  directory: string            # Backup storage path
  retention_count: integer     # Number of backups to keep
  last_backup: datetime?       # Last backup timestamp
  
admin:
  email: string                # Administrator email for notifications
```

**Validation Rules**:
- `instance_id`: Must be unique across Proxmox cluster, alphanumeric + hyphens
- `container_id`: Must be available (not in use)
- `cpu_cores`: Range 1-32
- `memory_mb`: Minimum 2048 (2GB), recommended 4096 (4GB)
- `disk_size_gb`: Minimum 20GB
- `ip_address`: Valid IP with CIDR notation
- `domain`: Valid FQDN format
- `database.type`: One of sqlite, postgresql, mysql
- `port_blue` != `port_green`: Ports must not conflict
- All ports: Range 1024-65535 (unprivileged)

**Relationships**:
- References: Proxmox node (external), container ID (external), SSL certificates (files)
- Referenced by: Deployment State, Backup Archives

---

## Entity: Deployment State

**Purpose**: Track current deployment status and operation locks

**Storage**: `/var/lib/vikunja-deploy/state` (on Proxmox host)

**Schema**:

```bash
# /var/lib/vikunja-deploy/state/<instance_id>.state
{
  "instance_id": "vikunja-main",
  "container_id": 100,
  "status": "running",              # "deploying" | "running" | "updating" | "stopped" | "failed"
  "active_color": "blue",           # Current active deployment color
  "last_operation": "update",       # Last operation type
  "last_operation_time": "2025-10-19T10:30:00Z",
  "last_operation_status": "success",  # "success" | "failure" | "in_progress"
  "deployed_version": "v0.23.0",    # Current Vikunja version
  "deployed_commit": "abc123def",   # Git commit hash
  "uptime_start": "2025-10-19T09:00:00Z"
}
```

**State Transitions**:
```
Initial Deployment:
  null → deploying → running

Update:
  running → updating → running (success)
  running → updating → failed → running (rollback)

Stop/Start:
  running → stopped → running

Uninstall:
  (any) → stopped → null (deleted)
```

**Validation Rules**:
- `status`: Must be one of defined enum values
- `active_color`: Must match `deployment-config.yaml` services.*.active_color
- State file must be locked during operations (see Lock File entity)

**Relationships**:
- References: Deployment Configuration (via instance_id)
- Referenced by: Lock File, Operation Log

---

## Entity: Lock File

**Purpose**: Prevent concurrent operations on the same deployment

**Storage**: `/var/lock/vikunja-deploy/<instance_id>.lock` (directory-based lock)

**Schema**:

```bash
# Directory: /var/lock/vikunja-deploy/<instance_id>.lock/
pid           # Process ID of lock holder
timestamp     # Lock acquisition time (Unix epoch)
owner         # User@hostname who acquired lock
operation     # Operation type: "deploy" | "update" | "backup" | "restore" | "uninstall"
```

**Lock Lifecycle**:
1. **Acquire**: `mkdir` (atomic) creates lock directory
2. **Hold**: Write PID, timestamp, owner, operation to files
3. **Release**: `rmdir` removes lock directory
4. **Stale Detection**: If timestamp > 3600 seconds, lock is stale and can be forcibly removed

**Validation Rules**:
- Lock directory must be atomic (mkdir is atomic on POSIX filesystems)
- Timeout: 3600 seconds (1 hour) before considering lock stale
- PID validation: Check if process still exists before declaring stale

**Relationships**:
- References: Deployment State (via instance_id)
- Prevents concurrent modifications to: Deployment Configuration, Deployment State

---

## Entity: Backup Archive

**Purpose**: Point-in-time snapshot of Vikunja instance

**Storage**: `/var/backups/vikunja/<instance_id>/<timestamp>.tar.gz`

**Archive Contents**:
```
vikunja-backup-<timestamp>/
├── metadata.json              # Backup metadata
├── deployment-config.yaml     # Deployment configuration
├── database/
│   ├── vikunja.db            # SQLite database (if SQLite)
│   └── vikunja.sql.gz        # PostgreSQL/MySQL dump (compressed)
├── files/                     # Task attachments
│   └── <file_id>/...
└── environment               # Environment variables (secrets redacted)
```

**Metadata Schema**:

```json
{
  "backup_version": "1.0.0",
  "instance_id": "vikunja-main",
  "created_at": "2025-10-19T10:00:00Z",
  "vikunja_version": "v0.23.0",
  "vikunja_commit": "abc123def456",
  "database_type": "postgresql",
  "database_size_bytes": 52428800,
  "files_count": 1234,
  "files_size_bytes": 1073741824,
  "total_size_bytes": 1126170624,
  "checksum_sha256": "abc123...",
  "hostname": "vikunja-main.example.com",
  "backup_trigger": "manual"
}
```

**Validation Rules**:
- `backup_version`: Semantic version of backup format
- `checksum_sha256`: SHA-256 of entire archive (for integrity verification)
- `total_size_bytes`: Must match actual archive size
- Archive must be compressed with gzip
- Retention: Keep last `backup.retention_count` archives (from config)

**Restoration Requirements**:
- Verify checksum before extracting
- Check `vikunja_version` compatibility
- Validate database dump integrity before importing
- Stop services before restoration
- Create backup of current state before restoring

**Relationships**:
- References: Deployment Configuration (via instance_id)
- Contains: Database snapshot, File attachments, Configuration

---

## Entity: Operation Log

**Purpose**: Audit trail of all deployment operations

**Storage**: `/var/log/vikunja-deploy/<instance_id>.log`

**Log Entry Format**:

```
[2025-10-19T10:30:00Z] [INFO] [deploy] Starting deployment of vikunja-main
[2025-10-19T10:30:15Z] [INFO] [deploy] Created LXC container ID 100
[2025-10-19T10:31:00Z] [INFO] [deploy] Installed Go runtime v1.21.5
[2025-10-19T10:32:00Z] [INFO] [deploy] Cloned repository commit abc123def
[2025-10-19T10:33:00Z] [INFO] [deploy] Built backend successfully
[2025-10-19T10:34:00Z] [INFO] [deploy] Started backend service on port 3456 (blue)
[2025-10-19T10:34:30Z] [INFO] [deploy] Health check passed: backend
[2025-10-19T10:35:00Z] [SUCCESS] [deploy] Deployment completed in 5m 0s
```

**Log Levels**:
- `DEBUG`: Detailed execution steps
- `INFO`: Normal operational messages
- `WARN`: Non-critical issues (e.g., stale locks removed)
- `ERROR`: Operation failures (with error codes)
- `SUCCESS`: Operation completion

**Log Rotation**:
- Max size: 10MB per log file
- Rotation: Keep 5 rotated logs
- Compression: gzip after rotation
- Tool: logrotate

**Validation Rules**:
- Timestamp: ISO 8601 format with timezone
- Operation: Must match Lock File operation types
- Structured format: `[timestamp] [level] [operation] message`

**Relationships**:
- References: Deployment State, Lock File (via instance_id and operation)
- Used by: Status checks, troubleshooting, audit

---

## Entity: Health Check Results

**Purpose**: Cache health check results for status reporting

**Storage**: `/var/lib/vikunja-deploy/health/<instance_id>.json`

**Schema**:

```json
{
  "instance_id": "vikunja-main",
  "timestamp": "2025-10-19T10:35:00Z",
  "overall_status": "healthy",
  "components": {
    "backend": {
      "status": "healthy",
      "response_time_ms": 45,
      "http_code": 200,
      "version": "v0.23.0",
      "uptime_seconds": 3600,
      "checks": {
        "database": "connected",
        "redis": "connected",
        "filesystem": "writable"
      }
    },
    "frontend": {
      "status": "healthy",
      "response_time_ms": 12,
      "http_code": 200,
      "nginx_status": "active"
    },
    "mcp": {
      "status": "healthy",
      "response_time_ms": 34,
      "http_code": 200,
      "version": "v1.0.0",
      "uptime_seconds": 3600
    },
    "database": {
      "status": "healthy",
      "type": "postgresql",
      "size_mb": 50,
      "connections": 5
    }
  },
  "resources": {
    "cpu_percent": 15.2,
    "memory_used_mb": 512,
    "memory_available_mb": 3584,
    "disk_used_gb": 5,
    "disk_available_gb": 15
  },
  "warnings": [],
  "errors": []
}
```

**Health Status Enum**:
- `healthy`: All checks passed
- `degraded`: Some non-critical checks failed (warnings present)
- `unhealthy`: Critical checks failed (errors present)
- `unknown`: Unable to determine status

**Validation Rules**:
- `timestamp`: Must be recent (<60 seconds for cached results)
- `overall_status`: Computed from component statuses
- `response_time_ms`: >0 for healthy components
- `http_code`: 200 for healthy, 5xx for unhealthy, 4xx for degraded

**Relationships**:
- References: Deployment Configuration (via instance_id)
- Used by: Status command, health monitoring, update decision logic

---

## State Machine: Deployment Lifecycle

```
┌─────────────────────────────────────────────────────────────┐
│                   Deployment Lifecycle                       │
└─────────────────────────────────────────────────────────────┘

[null]
  │
  │ vikunja-install.sh
  ↓
[deploying] ──(failure)──> [failed] ──(retry)──> [deploying]
  │                           │
  │ (success)                 │ (uninstall)
  ↓                           ↓
[running] <─────────────── [null]
  │   ↑
  │   │ (rollback success)
  │   │
  │ vikunja-update.sh
  ↓   │
[updating] ──(failure)──> [failed]
  │
  │ (success)
  ↓
[running]
  │
  │ vikunja-manage.sh stop
  ↓
[stopped]
  │
  │ vikunja-manage.sh start
  ↓
[running]
  │
  │ vikunja-manage.sh uninstall
  ↓
[null]
```

**State Definitions**:
- `null`: No deployment exists
- `deploying`: Initial deployment in progress
- `running`: Instance operational and healthy
- `updating`: Update/upgrade in progress
- `stopped`: Services stopped (intentional)
- `failed`: Operation failed (manual intervention needed)

---

## File System Layout

```
Proxmox Host:
/var/lib/vikunja-deploy/
├── state/
│   ├── vikunja-main.state          # Deployment state
│   └── vikunja-test.state          # Another instance
├── health/
│   ├── vikunja-main.json           # Health cache
│   └── vikunja-test.json
└── instances.json                  # Registry of all instances

/var/lock/vikunja-deploy/
├── vikunja-main.lock/              # Lock directory
│   ├── pid
│   ├── timestamp
│   ├── owner
│   └── operation
└── vikunja-test.lock/

/var/log/vikunja-deploy/
├── vikunja-main.log                # Operation log
└── vikunja-main.log.1.gz           # Rotated log

Inside Container:
/etc/vikunja/
├── deployment-config.yaml          # Main configuration
├── environment                     # Environment variables
└── secrets/
    └── db-password                 # Database password

/opt/vikunja/
├── backend/
│   ├── blue/                       # Blue deployment
│   │   └── vikunja                 # Binary
│   └── green/                      # Green deployment
│       └── vikunja                 # Binary
├── frontend/
│   └── dist/                       # Static files
└── mcp-server/
    ├── blue/                       # Blue deployment
    └── green/                      # Green deployment

/var/backups/vikunja/
└── vikunja-main/
    ├── 1729335600.tar.gz           # Backup archive
    ├── 1729338000.tar.gz
    └── latest -> 1729338000.tar.gz # Symlink to latest

/var/log/vikunja/
├── backend-blue.log
├── backend-green.log
├── mcp-blue.log
├── mcp-green.log
└── nginx-access.log
```

---

## Summary

This data model defines 6 primary entities:

1. **Deployment Configuration**: Persistent YAML configuration
2. **Deployment State**: Current status tracking (JSON)
3. **Lock File**: Concurrency control (directory-based)
4. **Backup Archive**: Point-in-time snapshots (tar.gz)
5. **Operation Log**: Audit trail (structured logs)
6. **Health Check Results**: Cached health status (JSON)

All entities follow validation rules and maintain referential relationships. The state machine ensures valid transitions during deployment lifecycle operations.
