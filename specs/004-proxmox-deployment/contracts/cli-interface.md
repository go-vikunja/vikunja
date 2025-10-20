# Script Contracts: Proxmox Deployment CLI

**Feature**: Proxmox LXC Automated Deployment  
**Date**: 2025-10-19  
**Purpose**: Define command-line interfaces for all deployment scripts

---

## Overview

This document specifies the CLI contracts for the Vikunja Proxmox deployment system. All scripts follow consistent conventions for arguments, options, exit codes, and output formatting.

---

## Global Conventions

### Exit Codes

All scripts use standard exit codes:

```bash
0   - Success
1   - General error
2   - Misuse of command (invalid arguments)
3   - Configuration error
4   - Resource unavailable (ports, disk space, etc.)
5   - Operation locked (another operation in progress)
10  - Health check failed
11  - Rollback succeeded (update failed)
20  - User cancelled operation
```

### Output Format

**Standard Output** (stdout):
- Informational messages
- Progress indicators
- Results and status

**Standard Error** (stderr):
- Error messages
- Warnings
- Debug output (when --debug enabled)

**Colors**:
```bash
GREEN   - Success messages
RED     - Error messages
YELLOW  - Warning messages
BLUE    - Informational messages
CYAN    - Progress indicators
```

### Common Options

All scripts support:
```
-h, --help              Show help message and exit
-v, --verbose           Verbose output
-d, --debug             Debug output with timestamps
-q, --quiet             Suppress non-error output
-y, --yes               Non-interactive mode (use defaults)
-n, --dry-run           Show what would be done without executing
--version               Show script version
```

---

## 1. vikunja-install.sh

**Purpose**: Initial deployment of Vikunja to Proxmox LXC container (Bootstrap script)

**Execution**: Runs on Proxmox host as root

**Architecture**: This script uses a three-stage bootstrap pattern:
1. **Bootstrap Stage**: User runs curl command, executes lightweight bootstrap script
2. **Download Stage**: Bootstrap downloads all required files (main installer, libraries, templates) to `/tmp/vikunja-installer-<PID>/`
3. **Execute Stage**: Bootstrap launches `vikunja-install-main.sh` with all dependencies available

This pattern enables single-command installation while maintaining modular code architecture.

### Synopsis

```bash
# Recommended: One-line curl installation
bash <(curl -fsSL https://raw.githubusercontent.com/go-vikunja/vikunja/main/deploy/proxmox/vikunja-install.sh)

# With environment variable customization (for custom branches/forks)
export VIKUNJA_GITHUB_OWNER="yourname"
export VIKUNJA_GITHUB_REPO="vikunja"
export VIKUNJA_GITHUB_BRANCH="feature-branch"
bash <(curl -fsSL https://raw.githubusercontent.com/${VIKUNJA_GITHUB_OWNER}/${VIKUNJA_GITHUB_REPO}/${VIKUNJA_GITHUB_BRANCH}/deploy/proxmox/vikunja-install.sh)

# Alternative: Local installation (skips bootstrap, requires git clone)
git clone https://github.com/aroige/vikunja.git
cd vikunja/deploy/proxmox
./vikunja-install-main.sh [OPTIONS]
```

### Environment Variables (Bootstrap Customization)

The bootstrap script respects these environment variables:

```bash
VIKUNJA_GITHUB_OWNER    # GitHub repository owner (default: aroige)
VIKUNJA_GITHUB_REPO     # GitHub repository name (default: vikunja)
VIKUNJA_GITHUB_BRANCH   # Git branch to install from (default: main)
```

**Use Case**: Install from a development branch or custom fork:
```bash
export VIKUNJA_GITHUB_BRANCH="004-proxmox-deployment"
bash <(curl -fsSL https://raw.githubusercontent.com/aroige/vikunja/${VIKUNJA_GITHUB_BRANCH}/deploy/proxmox/vikunja-install.sh)
```

### Options

```
-i, --instance-id <ID>          Instance identifier (default: vikunja-main)
-c, --container-id <ID>         LXC container ID (default: auto-select)
--node <NODE>                   Proxmox node name (default: current node)

Database Options:
--db-type <TYPE>                Database type: sqlite | postgresql | mysql (interactive if not set)
--db-host <HOST>                External database host (for postgresql/mysql)
--db-port <PORT>                External database port
--db-name <NAME>                Database name
--db-user <USER>                Database username
--db-password <PASS>            Database password (insecure - use password file)
--db-password-file <FILE>       Path to password file

Network Options:
--ip <IP/CIDR>                  Static IP address (e.g., 192.168.1.100/24)
--gateway <IP>                  Gateway IP address
--domain <DOMAIN>               Public domain name (e.g., vikunja.example.com)
--bridge <BRIDGE>               Proxmox network bridge (default: vmbr0)

Resources:
--cpu <CORES>                   CPU cores (default: 2)
--memory <MB>                   RAM in MB (default: 4096)
--disk <GB>                     Disk size in GB (default: 20)

SSL:
--ssl-cert <FILE>               Path to SSL certificate
--ssl-key <FILE>                Path to SSL private key

Git:
--git-repo <URL>                Git repository URL (default: official Vikunja repo)
--git-branch <BRANCH>           Git branch (default: main)

Other:
--admin-email <EMAIL>           Administrator email for notifications
--config-file <FILE>            Pre-configured deployment config YAML
```

### Interactive Prompts

When options are not provided via CLI, the script prompts interactively:

```
1. Instance ID [vikunja-main]: 
2. Container ID (100-999) [auto-select]: 
3. Database type (sqlite/postgresql/mysql) [sqlite]: 
   → If postgresql/mysql:
     - Database host [localhost]: 
     - Database port [5432/3306]: 
     - Database name [vikunja]: 
     - Database user [vikunja]: 
     - Database password: (hidden input)
4. Public domain [vikunja.local]: 
5. Static IP address (CIDR) [192.168.1.100/24]: 
6. Gateway [192.168.1.1]: 
7. CPU cores [2]: 
8. Memory (MB) [4096]: 
9. Disk size (GB) [20]: 
10. SSL certificate path (or 'skip'): 
11. SSL key path (or 'skip'): 
12. Administrator email: 

Confirm configuration? [y/N]: 
```

### Output

**Success**:
```
┌─────────────────────────────────────────────────────────────┐
│           Vikunja Deployment Successful                      │
├─────────────────────────────────────────────────────────────┤
│ Instance ID:    vikunja-main                                 │
│ Container ID:   100                                          │
│ IP Address:     192.168.1.100                                │
│ Domain:         vikunja.example.com                          │
│ Version:        v0.23.0 (abc123def)                          │
│ Database:       PostgreSQL (192.168.1.50:5432)               │
│ Admin Email:    admin@example.com                            │
│                                                              │
│ Access URL:     https://vikunja.example.com                  │
│ Deployment Time: 8m 32s                                      │
│                                                              │
│ Next steps:                                                  │
│ 1. Configure DNS: vikunja.example.com → 192.168.1.100       │
│ 2. Create admin account: vikunja user add                    │
│ 3. Check status: vikunja-manage.sh status                    │
│ 4. View logs: vikunja-manage.sh logs                         │
└─────────────────────────────────────────────────────────────┘

Configuration saved to: /etc/vikunja/deployment-config.yaml
Log file: /var/log/vikunja-deploy/vikunja-main.log
```

**Failure**:
```
[ERROR] Deployment failed at step: Building backend
Error: Insufficient disk space (available: 10GB, required: 20GB)

Troubleshooting:
- Free up disk space on Proxmox node
- Reduce disk allocation with --disk option
- Check logs: /var/log/vikunja-deploy/vikunja-main.log

Cleanup completed. Safe to retry.
Exit code: 4
```

### Exit Codes

```
0  - Deployment successful
1  - General deployment error
2  - Invalid arguments
3  - Configuration validation failed
4  - Insufficient resources (CPU/RAM/disk)
5  - Container ID already in use
6  - Network configuration error (IP conflict, invalid gateway)
7  - Database connection failed
8  - Git repository clone failed
9  - Build failed (Go/Node.js compilation)
10 - Health check failed after deployment
```

---

## 2. vikunja-update.sh

**Purpose**: Update existing Vikunja deployment to latest version from main branch

**Execution**: Runs on Proxmox host as root

### Synopsis

```bash
vikunja-update.sh [OPTIONS] <instance-id>
```

### Arguments

```
<instance-id>           Instance to update (required)
```

### Options

```
--git-branch <BRANCH>   Git branch to update from (default: main)
--git-commit <HASH>     Specific commit to deploy (default: latest)
--force                 Force update even if no new commits
--skip-backup           Skip pre-update backup (NOT RECOMMENDED)
--skip-migrations       Skip database migrations (dangerous)
--rollback-on-failure   Automatically rollback on any failure (default: true)
--no-health-check       Skip health checks (dangerous)
```

### Output

**Success**:
```
[INFO] Starting update for vikunja-main...
[INFO] Current version: v0.22.0 (xyz789abc)
[INFO] Latest version: v0.23.0 (abc123def)
[INFO] Changes: 47 commits, 156 files changed

[1/8] Creating pre-update backup...                    ✓ (15s)
[2/8] Pulling latest changes...                        ✓ (8s)
[3/8] Building backend...                              ✓ (45s)
[4/8] Building MCP server...                           ✓ (12s)
[5/8] Running database migrations...                   ✓ (3s)
      - Applied 2 new migrations
[6/8] Starting services on green (port 3457)...        ✓ (5s)
[7/8] Health check...                                  ✓ (8s)
      - Backend: healthy (response: 45ms)
      - MCP: healthy (response: 34ms)
      - Database: connected
[8/8] Switching traffic to green...                    ✓ (2s)

Update completed successfully in 4m 38s

Active version: v0.23.0 (abc123def)
Downtime: 3 seconds (99.998% uptime maintained)

Rollback available: vikunja-update.sh --rollback vikunja-main
```

**Failure with Rollback**:
```
[ERROR] Health check failed after update

[ROLLBACK] Reverting to previous version...
[1/4] Switching traffic back to blue...                ✓ (2s)
[2/4] Stopping green services...                       ✓ (3s)
[3/4] Verifying blue health...                         ✓ (5s)
[4/4] Cleaning up failed deployment...                 ✓ (2s)

Rollback completed in 45 seconds
Active version: v0.22.0 (xyz789abc) - STABLE

Failure reason: Backend health check timeout
Check logs: /var/log/vikunja-deploy/vikunja-main.log

Exit code: 10
```

### Exit Codes

```
0  - Update successful
1  - General update error
2  - Invalid arguments
3  - Instance not found
5  - Update already in progress (locked)
10 - Health check failed (post-update)
11 - Rollback successful (update failed but recovered)
12 - Rollback failed (manual intervention needed)
20 - No updates available
```

---

## 3. vikunja-manage.sh

**Purpose**: Management operations (status, backup, restore, reconfigure, uninstall)

**Execution**: Runs on Proxmox host as root

### Synopsis

```bash
vikunja-manage.sh <command> [OPTIONS] <instance-id>
```

### Commands

#### status

Display current status of deployment

```bash
vikunja-manage.sh status [OPTIONS] <instance-id>

Options:
--format <FORMAT>       Output format: text | json | table (default: table)
--watch                 Continuous monitoring mode (refresh every 5s)
```

**Output**:
```
┌─────────────────────────────────────────────────────────────┐
│              Vikunja Status: vikunja-main                    │
├─────────────────────────────────────────────────────────────┤
│ Overall Status:    ✓ Healthy                                │
│ Uptime:            3d 14h 22m                                │
│ Version:           v0.23.0 (abc123def)                       │
│ Active Color:      blue                                      │
├─────────────────────────────────────────────────────────────┤
│ Components                                                   │
├─────────────────────────────────────────────────────────────┤
│ ✓ Backend          healthy  (port 3456, response: 45ms)     │
│ ✓ Frontend         healthy  (nginx active)                  │
│ ✓ MCP Server       healthy  (port 8456, response: 34ms)     │
│ ✓ Database         healthy  (postgresql, 50MB, 5 conns)     │
├─────────────────────────────────────────────────────────────┤
│ Resources                                                    │
├─────────────────────────────────────────────────────────────┤
│ CPU Usage:         15.2% (2 cores)                           │
│ Memory:            512MB / 4096MB (12.5%)                    │
│ Disk:              5GB / 20GB (25%)                          │
├─────────────────────────────────────────────────────────────┤
│ Last Operations                                              │
├─────────────────────────────────────────────────────────────┤
│ Update:            2025-10-16 10:30 (success)                │
│ Backup:            2025-10-19 09:00 (success)                │
│ Health Check:      2025-10-19 10:35 (success)                │
└─────────────────────────────────────────────────────────────┘

Commands:
  vikunja-manage.sh logs vikunja-main       View logs
  vikunja-manage.sh backup vikunja-main     Create backup
  vikunja-manage.sh restart vikunja-main    Restart services
```

#### backup

Create a backup archive

```bash
vikunja-manage.sh backup [OPTIONS] <instance-id>

Options:
--output <PATH>         Backup output path (default: /var/backups/vikunja/)
--compress <LEVEL>      Compression level 1-9 (default: 6)
--encrypt               Encrypt backup (prompts for password)
```

**Output**:
```
[INFO] Creating backup for vikunja-main...
[1/5] Stopping background tasks...                     ✓ (2s)
[2/5] Dumping database (50MB)...                       ✓ (8s)
[3/5] Archiving files (1234 files, 1GB)...             ✓ (35s)
[4/5] Compressing archive...                           ✓ (25s)
[5/5] Verifying backup integrity...                    ✓ (5s)

Backup created successfully

File: /var/backups/vikunja/vikunja-main/1729335600.tar.gz
Size: 487MB (compressed from 1.05GB)
Checksum: sha256:abc123def456...
Duration: 1m 15s

Restore with: vikunja-manage.sh restore vikunja-main --from <backup-file>
```

#### restore

Restore from backup archive

```bash
vikunja-manage.sh restore [OPTIONS] <instance-id>

Options:
--from <FILE>           Backup file to restore (required)
--decrypt               Decrypt backup (prompts for password)
--force                 Skip confirmation prompt
```

**Output**:
```
[WARN] This will replace all current data
Current version: v0.23.0
Backup version:  v0.22.0
Confirm restore? [y/N]: y

[INFO] Restoring vikunja-main from backup...
[1/6] Stopping services...                             ✓ (5s)
[2/6] Creating safety backup...                        ✓ (45s)
[3/6] Verifying backup integrity...                    ✓ (5s)
[4/6] Extracting archive...                            ✓ (30s)
[5/6] Restoring database...                            ✓ (15s)
[6/6] Starting services...                             ✓ (10s)

Restore completed successfully in 2m 10s

Active version: v0.22.0 (restored)
Safety backup: /var/backups/vikunja/vikunja-main/pre-restore-1729336000.tar.gz
```

#### reconfigure

Modify deployment configuration

```bash
vikunja-manage.sh reconfigure [OPTIONS] <instance-id>

Options:
--domain <DOMAIN>       Change public domain
--db-type <TYPE>        Change database type (requires migration)
--db-host <HOST>        Change database host
--admin-email <EMAIL>   Change admin email
--config-file <FILE>    Apply configuration from file
--interactive           Interactive reconfiguration wizard
```

#### logs

View deployment logs

```bash
vikunja-manage.sh logs [OPTIONS] <instance-id>

Options:
--component <NAME>      Component: backend | frontend | mcp | deploy (default: all)
--follow, -f            Follow log output (tail -f)
--lines <N>             Number of lines (default: 100)
--since <TIME>          Show logs since time (e.g., "1 hour ago")
```

#### restart

Restart services

```bash
vikunja-manage.sh restart [OPTIONS] <instance-id>

Options:
--component <NAME>      Restart specific component (default: all)
--graceful              Graceful restart with health checks
```

#### stop / start

Stop or start services

```bash
vikunja-manage.sh stop <instance-id>
vikunja-manage.sh start <instance-id>
```

#### uninstall

Remove deployment

```bash
vikunja-manage.sh uninstall [OPTIONS] <instance-id>

Options:
--keep-data             Keep data files (create final backup)
--force                 Skip confirmation
```

**Output**:
```
[WARN] This will permanently delete vikunja-main
- Container ID: 100
- IP Address: 192.168.1.100
- Database: postgresql (external - not deleted)

Create final backup? [Y/n]: y
Confirm uninstall? [y/N]: y

[INFO] Uninstalling vikunja-main...
[1/5] Creating final backup...                         ✓ (1m 15s)
[2/5] Stopping services...                             ✓ (5s)
[3/5] Removing container...                            ✓ (8s)
[4/5] Cleaning up state files...                       ✓ (1s)
[5/5] Removing logs (--keep-data preserves backup)...  ✓ (2s)

Uninstall completed successfully

Final backup: /var/backups/vikunja/vikunja-main/final-1729337000.tar.gz
```

---

## 4. Library Functions (lib/common.sh)

**Purpose**: Shared utility functions used by all scripts

### Functions

```bash
# Logging
log_info <message>              # Info message (blue)
log_success <message>           # Success message (green)
log_warn <message>              # Warning message (yellow)
log_error <message>             # Error message (red)
log_debug <message>             # Debug message (if --debug)

# Progress
progress_start <message>        # Start progress indicator
progress_update <percent>       # Update progress (0-100)
progress_complete               # Complete progress (✓)
progress_fail <error>           # Fail progress (✗)

# Validation
validate_ip <ip>                # Validate IP address format
validate_domain <domain>        # Validate domain format
validate_port <port>            # Validate port number (1024-65535)
validate_email <email>          # Validate email format
check_root                      # Ensure running as root
check_proxmox                   # Ensure running on Proxmox node

# Lock Management
acquire_lock <instance_id> <operation>     # Acquire deployment lock
release_lock <instance_id>                 # Release lock
check_lock <instance_id>                   # Check lock status

# Configuration
load_config <instance_id>       # Load deployment config
save_config <instance_id>       # Save deployment config
update_config <instance_id> <key> <value>  # Update single value

# State Management
get_state <instance_id>         # Get current state
set_state <instance_id> <status>           # Update state
update_deployed_version <instance_id> <version> <commit>

# Health Checks
check_component_health <component> <port>  # Check single component
full_health_check <instance_id>            # Check all components

# Error Handling
error <message> [exit_code]     # Print error and exit
trap_errors                     # Set up error trapping
cleanup_on_error                # Cleanup function for trap
```

---

## 5. Usage Examples

### Example 1: Fresh Installation (Interactive)

```bash
curl -fsSL https://get.vikunja.io/proxmox | bash
```

User answers prompts, deployment completes in <10 minutes.

### Example 2: Fresh Installation (Non-Interactive)

```bash
vikunja-install.sh \
  --instance-id production \
  --db-type postgresql \
  --db-host 192.168.1.50 \
  --db-name vikunja_prod \
  --db-user vikunja \
  --db-password-file /root/db-pass.txt \
  --domain vikunja.company.com \
  --ip 192.168.1.100/24 \
  --gateway 192.168.1.1 \
  --ssl-cert /root/ssl/cert.pem \
  --ssl-key /root/ssl/key.pem \
  --admin-email admin@company.com \
  --yes
```

### Example 3: Update to Latest

```bash
vikunja-update.sh production
```

Checks for updates, runs migrations, performs blue-green deployment.

### Example 4: Check Status

```bash
vikunja-manage.sh status production --watch
```

Continuous monitoring with 5-second refresh.

### Example 5: Backup Before Major Change

```bash
vikunja-manage.sh backup production --encrypt
vikunja-manage.sh reconfigure production --db-type mysql --interactive
```

### Example 6: Restore After Failure

```bash
vikunja-manage.sh restore production --from /var/backups/vikunja/production/1729335600.tar.gz
```

### Example 7: Multiple Instances

```bash
# Deploy staging instance
vikunja-install.sh --instance-id staging --container-id 101 --ip 192.168.1.101/24

# Deploy development instance
vikunja-install.sh --instance-id dev --container-id 102 --ip 192.168.1.102/24

# List all instances
vikunja-manage.sh list
```

---

## Contract Compliance Checklist

✅ All scripts support `--help`, `--version`, common options  
✅ Consistent exit codes across all scripts  
✅ Color-coded output with stderr for errors  
✅ Interactive and non-interactive modes  
✅ Dry-run capability where applicable  
✅ Validation before destructive operations  
✅ Progress indicators for long operations  
✅ Structured error messages with remediation  
✅ JSON output option for programmatic use  
✅ Lock mechanism prevents concurrent operations  
✅ Comprehensive logging for audit trail
