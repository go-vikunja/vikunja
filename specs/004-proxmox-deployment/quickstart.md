# Quickstart Guide: Vikunja on Proxmox

**Deploy Vikunja to Proxmox LXC in Under 10 Minutes**

---

## Prerequisites

Before starting, ensure you have:

- ✅ **Proxmox VE 7.0+** installed and accessible
- ✅ **Root SSH access** to your Proxmox host
- ✅ **Minimum resources**: 2 CPU cores, 4GB RAM, 20GB disk available
- ✅ **Internet connectivity** on Proxmox node
- ✅ **Domain name** pointed to your Proxmox node IP (optional but recommended)
- ✅ **SSL certificate** and key (optional, can use self-signed)

---

## Quick Installation

### Option 1: One-Line Install (Recommended)

SSH into your Proxmox host and run:

**For stable release (once merged to main):**
```bash
bash <(curl -fsSL https://raw.githubusercontent.com/aroige/vikunja/main/deploy/proxmox/vikunja-install.sh)
```

**For development/testing (branch 004-proxmox-deployment):**
```bash
export VIKUNJA_GITHUB_BRANCH="004-proxmox-deployment"
bash <(curl -fsSL https://raw.githubusercontent.com/aroige/vikunja/${VIKUNJA_GITHUB_BRANCH}/deploy/proxmox/vikunja-install.sh)
```

Or as a one-liner:
```bash
VIKUNJA_GITHUB_BRANCH="004-proxmox-deployment" bash <(curl -fsSL https://raw.githubusercontent.com/aroige/vikunja/004-proxmox-deployment/deploy/proxmox/vikunja-install.sh)
```

Follow the interactive prompts. Installation completes in 5-10 minutes.

### Option 2: Clone Repository and Run

```bash
git clone https://github.com/aroige/vikunja.git
cd vikunja/deploy/proxmox
./vikunja-install-main.sh
```

> **Note**: The installer requires multiple library files. For single-command installation, use Option 1 above.

---

## Interactive Setup

The installer will prompt you for the following information:

### 1. Basic Configuration

```
Instance ID [vikunja-main]: 
```
**What it is**: Unique identifier for this Vikunja installation  
**Example**: `production`, `vikunja-team`, `my-tasks`  
**Tip**: Use descriptive names if deploying multiple instances

```
Container ID (100-999) [auto-select]: 
```
**What it is**: LXC container ID in Proxmox  
**Default**: Automatically finds the next available ID  
**Tip**: Accept default unless you have specific requirements

### 2. Database Selection

```
Database type (sqlite/postgresql/mysql) [sqlite]: 
```

**SQLite** (Recommended for single-user or small teams):
- ✅ No external database needed
- ✅ Simplest setup
- ⚠️ Single-file database, less concurrent performance

**PostgreSQL** (Recommended for teams):
- ✅ Best concurrent performance
- ✅ Advanced features
- ℹ️ Requires external PostgreSQL server

**MySQL** (Alternative for teams):
- ✅ Good performance
- ℹ️ Requires external MySQL server

**If you select PostgreSQL or MySQL**, you'll be prompted for:
```
Database host [localhost]: 192.168.1.50
Database port [5432]: 
Database name [vikunja]: 
Database user [vikunja]: 
Database password: ********
```

### 3. Network Configuration

```
Public domain [vikunja.local]: vikunja.example.com
```
**What it is**: The domain name users will use to access Vikunja  
**Tip**: Configure DNS A record: `vikunja.example.com → your-proxmox-ip`

```
Static IP address (CIDR) [192.168.1.100/24]: 
```
**What it is**: IP address for the container  
**Format**: `IP/subnet` (e.g., `192.168.1.100/24`)  
**Tip**: Choose an IP in your local network range

```
Gateway [192.168.1.1]: 
```
**What it is**: Network gateway for internet access  
**Default**: Usually your router's IP

### 4. Resource Allocation

```
CPU cores [2]: 
```
**Recommended**: 2 cores minimum, 4 for larger teams

```
Memory (MB) [4096]: 
```
**Recommended**: 4096MB (4GB) minimum, 8192MB (8GB) for larger teams

```
Disk size (GB) [20]: 
```
**Recommended**: 20GB minimum, more if you expect many file attachments

### 5. SSL Configuration

```
SSL certificate path (or 'skip'): /root/ssl/cert.pem
SSL key path (or 'skip'): /root/ssl/key.pem
```

**Options**:
- Provide paths to your SSL certificate and key
- Type `skip` to use HTTP only (not recommended for production)
- Generate self-signed certificate: 
  ```bash
  openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout /tmp/vikunja-key.pem \
    -out /tmp/vikunja-cert.pem
  ```

### 6. Admin Contact

```
Administrator email: admin@example.com
```
**What it is**: Email for system notifications and alerts

### 7. Confirmation

```
┌─────────────────────────────────────────────────────────────┐
│              Deployment Configuration Summary                │
├─────────────────────────────────────────────────────────────┤
│ Instance ID:    production                                   │
│ Container ID:   100                                          │
│ Database:       PostgreSQL (192.168.1.50:5432)               │
│ Domain:         vikunja.example.com                          │
│ IP Address:     192.168.1.100/24                             │
│ Resources:      2 CPU, 4096MB RAM, 20GB disk                 │
│ SSL:            Enabled                                      │
│ Admin Email:    admin@example.com                            │
└─────────────────────────────────────────────────────────────┘

Confirm configuration? [y/N]: y
```

Review carefully and type `y` to proceed.

---

## Installation Progress

The installer will show real-time progress:

```
[1/10] Creating LXC container...                       ✓ (15s)
[2/10] Starting container...                           ✓ (5s)
[3/10] Installing system dependencies...               ✓ (45s)
[4/10] Installing Go runtime...                        ✓ (30s)
[5/10] Installing Node.js runtime...                   ✓ (20s)
[6/10] Cloning Vikunja repository...                   ✓ (25s)
[7/10] Building backend...                             ✓ (90s)
[8/10] Building frontend...                            ✓ (60s)
[9/10] Configuring services...                         ✓ (15s)
[10/10] Running health checks...                       ✓ (10s)
```

**Total time**: 5-10 minutes depending on internet speed and system resources.

---

## Post-Installation

### Success Message

```
┌─────────────────────────────────────────────────────────────┐
│           Vikunja Deployment Successful                      │
├─────────────────────────────────────────────────────────────┤
│ Instance ID:    production                                   │
│ Container ID:   100                                          │
│ IP Address:     192.168.1.100                                │
│ Domain:         vikunja.example.com                          │
│ Version:        v0.23.0 (abc123def)                          │
│ Database:       PostgreSQL (192.168.1.50:5432)               │
│                                                              │
│ Access URL:     https://vikunja.example.com                  │
│ Deployment Time: 8m 32s                                      │
└─────────────────────────────────────────────────────────────┘
```

### Next Steps

#### 1. Configure DNS

Point your domain to the Proxmox node IP:

```
vikunja.example.com  A  YOUR-PROXMOX-IP
```

**Verify DNS propagation**:
```bash
nslookup vikunja.example.com
```

#### 2. Access Vikunja

Open your browser and navigate to:
```
https://vikunja.example.com
```

You'll see the Vikunja registration page.

#### 3. Create Admin Account

Register the first user - this account will have admin privileges:

1. Click "Register" on the login page
2. Enter your email and password
3. Verify your email (if SMTP configured)
4. Log in and start creating tasks!

#### 4. Verify Installation

Check the status of your deployment:

```bash
vikunja-manage.sh status production
```

Expected output:
```
Overall Status:    ✓ Healthy
Uptime:            5m 12s
Version:           v0.23.0
Components:        ✓ Backend  ✓ Frontend  ✓ MCP  ✓ Database
```

---

## Common Use Cases

### Updating Vikunja

When new features are released:

```bash
vikunja-update.sh production
```

**Update process**:
- ✅ Pulls latest code from main branch
- ✅ Creates automatic backup
- ✅ Runs database migrations
- ✅ Zero-downtime deployment (<5 seconds)
- ✅ Automatic rollback on failure

**Duration**: 3-5 minutes

### Creating Backups

Before major changes or on a schedule:

```bash
vikunja-manage.sh backup production
```

**Backup includes**:
- ✅ Database (complete dump)
- ✅ Configuration files
- ✅ Uploaded task attachments
- ✅ Metadata and checksums

**Location**: `/var/backups/vikunja/production/`

### Checking Status

Monitor your deployment health:

```bash
# One-time status check
vikunja-manage.sh status production

# Continuous monitoring (refreshes every 5 seconds)
vikunja-manage.sh status production --watch
```

### Viewing Logs

Troubleshoot issues:

```bash
# View all logs
vikunja-manage.sh logs production

# Follow logs in real-time
vikunja-manage.sh logs production --follow

# View specific component
vikunja-manage.sh logs production --component backend

# Last 50 lines
vikunja-manage.sh logs production --lines 50
```

### Restarting Services

If needed:

```bash
# Restart all services
vikunja-manage.sh restart production

# Restart specific component
vikunja-manage.sh restart production --component backend

# Graceful restart (zero downtime)
vikunja-manage.sh restart production --graceful
```

---

## Advanced Configuration

### Non-Interactive Installation

For automation or scripted deployments:

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
  --cpu 4 \
  --memory 8192 \
  --disk 50 \
  --ssl-cert /root/ssl/cert.pem \
  --ssl-key /root/ssl/key.pem \
  --admin-email admin@company.com \
  --yes
```

### Multiple Instances

Deploy multiple Vikunja instances on the same Proxmox cluster:

```bash
# Production instance
vikunja-install.sh --instance-id production --container-id 100 --ip 192.168.1.100/24

# Staging instance
vikunja-install.sh --instance-id staging --container-id 101 --ip 192.168.1.101/24

# Development instance
vikunja-install.sh --instance-id dev --container-id 102 --ip 192.168.1.102/24
```

Each instance is completely isolated with its own:
- Container ID
- IP address
- Domain
- Database
- Configuration

### Custom Git Branch

Deploy from a specific branch or commit:

```bash
vikunja-install.sh --git-branch develop
vikunja-update.sh production --git-commit abc123def
```

### Reconfiguring After Deployment

Change settings without redeployment:

```bash
# Interactive reconfiguration wizard
vikunja-manage.sh reconfigure production --interactive

# Change specific settings
vikunja-manage.sh reconfigure production --domain new-domain.com
vikunja-manage.sh reconfigure production --admin-email new-admin@example.com
```

---

## Troubleshooting

### Installation Failed

1. **Check logs**:
   ```bash
   cat /var/log/vikunja-deploy/production.log
   ```

2. **Common issues**:
   - **Insufficient disk space**: Free up space or increase `--disk` allocation
   - **Port conflicts**: Container ID or ports already in use
   - **Network issues**: Check firewall, internet connectivity
   - **Database connection**: Verify external database credentials and accessibility

3. **Retry after fixing**:
   ```bash
   vikunja-install.sh  # Use same configuration
   ```

### Update Failed

1. **Automatic rollback**: Update script automatically rolls back on failure

2. **Manual rollback** (if needed):
   ```bash
   vikunja-update.sh production --rollback
   ```

3. **Check what went wrong**:
   ```bash
   vikunja-manage.sh logs production --component backend --lines 200
   ```

### Services Not Starting

1. **Check status**:
   ```bash
   vikunja-manage.sh status production
   ```

2. **Restart services**:
   ```bash
   vikunja-manage.sh restart production
   ```

3. **Enter container for debugging**:
   ```bash
   pct enter 100
   systemctl status vikunja-backend
   systemctl status vikunja-mcp
   journalctl -u vikunja-backend -n 50
   ```

### Cannot Access Web Interface

1. **Verify DNS**:
   ```bash
   nslookup vikunja.example.com
   ping vikunja.example.com
   ```

2. **Check nginx**:
   ```bash
   pct enter 100
   systemctl status nginx
   nginx -t  # Test configuration
   ```

3. **Check SSL certificates**:
   ```bash
   openssl x509 -in /etc/vikunja/ssl/cert.pem -noout -dates
   ```

4. **Firewall rules**:
   ```bash
   # On Proxmox host
   iptables -L -n | grep 443
   ```

### Database Connection Issues

1. **Test database connectivity** (from container):
   ```bash
   pct enter 100
   
   # PostgreSQL
   psql -h 192.168.1.50 -U vikunja -d vikunja
   
   # MySQL
   mysql -h 192.168.1.50 -u vikunja -p vikunja
   ```

2. **Check database logs**:
   ```bash
   vikunja-manage.sh logs production --component backend | grep -i database
   ```

---

## Getting Help

### Documentation

- **Full documentation**: `deploy/proxmox/docs/README.md`
- **Architecture**: `deploy/proxmox/docs/ARCHITECTURE.md`
- **Troubleshooting guide**: `deploy/proxmox/docs/TROUBLESHOOTING.md`

### Support

- **GitHub Issues**: https://github.com/go-vikunja/vikunja/issues
- **Community Forum**: https://community.vikunja.io
- **Matrix Chat**: #vikunja:matrix.org

### Command Reference

```bash
# Get help for any command
vikunja-install.sh --help
vikunja-update.sh --help
vikunja-manage.sh --help
vikunja-manage.sh status --help
```

---

## Maintenance Schedule

### Daily
- ✅ Monitor status: `vikunja-manage.sh status production --watch`

### Weekly
- ✅ Check for updates: `vikunja-update.sh production`
- ✅ Review logs: `vikunja-manage.sh logs production --since "1 week ago"`

### Monthly
- ✅ Create backup: `vikunja-manage.sh backup production --encrypt`
- ✅ Verify backup: Test restoration to a dev instance
- ✅ Review disk usage: Check container disk space

### Quarterly
- ✅ Review SSL certificates: Renew if expiring soon
- ✅ Update Proxmox: Keep Proxmox VE updated
- ✅ Security audit: Review access logs and permissions

---

## Uninstalling

If you need to remove Vikunja:

```bash
# Create final backup
vikunja-manage.sh backup production

# Uninstall (keeps backup)
vikunja-manage.sh uninstall production --keep-data

# Complete removal (deletes everything)
vikunja-manage.sh uninstall production --force
```

**What gets removed**:
- ✅ LXC container
- ✅ State and lock files
- ✅ Logs
- ❌ Backups (preserved with `--keep-data`)
- ❌ External database (not touched)

---

## Success! 🎉

You now have a fully functional Vikunja installation with:

- ✅ **Easy updates**: Single command, zero downtime
- ✅ **Automatic backups**: Before every update
- ✅ **Health monitoring**: Real-time status checks
- ✅ **Rollback capability**: Safe update with automatic recovery
- ✅ **Production-ready**: SSL, systemd services, proper logging

**Start organizing your tasks at**: `https://vikunja.example.com`

---

**Need more features?** Check out the [full documentation](deploy/proxmox/docs/) for advanced configuration, multiple instances, and integration guides.
