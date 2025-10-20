# Troubleshooting Guide: Vikunja Proxmox Deployment

**Common Issues and Solutions**

This guide helps you diagnose and resolve issues with the Vikunja Proxmox deployment system.

---

## Table of Contents

- [Bootstrap Installation Issues](#bootstrap-installation-issues)
- [Pre-Flight Check Failures](#pre-flight-check-failures)
- [Container Creation Issues](#container-creation-issues)
- [Build Failures](#build-failures)
- [Service Startup Issues](#service-startup-issues)
- [Update and Rollback Issues](#update-and-rollback-issues)
- [Network and Connectivity Issues](#network-and-connectivity-issues)
- [Database Connection Issues](#database-connection-issues)
- [Health Check Failures](#health-check-failures)
- [Performance Issues](#performance-issues)
- [Backup and Restore Issues](#backup-and-restore-issues)

---

## Bootstrap Installation Issues

The bootstrap installer (`vikunja-install.sh`) downloads all required files before executing the main installer. Common issues:

### Issue: "Failed to download" Error

**Symptom**:
```
[ERROR] Bootstrap failed: Could not download vikunja-install-main.sh
```

**Possible Causes**:
1. No internet connectivity on Proxmox host
2. GitHub is unreachable (DNS or firewall issue)
3. Incorrect branch/repository specified
4. File does not exist in repository

**Solutions**:

**Check Internet Connectivity**:
```bash
# Test basic connectivity
ping -c 3 8.8.8.8

# Test DNS resolution
nslookup raw.githubusercontent.com

# Test HTTPS to GitHub
curl -I https://raw.githubusercontent.com
```

**Check Firewall Rules**:
```bash
# Verify HTTPS (port 443) is allowed
iptables -L -n | grep 443

# Test direct download
curl -fsSL https://raw.githubusercontent.com/aroige/vikunja/main/deploy/proxmox/vikunja-install-main.sh
```

If this fails, check your firewall/proxy settings.

**Verify Repository and Branch**:
```bash
# Check if branch exists
git ls-remote https://github.com/aroige/vikunja.git | grep -E 'refs/heads/(main|004-proxmox-deployment)'

# Try with explicit branch
export VIKUNJA_GITHUB_BRANCH="main"
bash <(curl -fsSL https://raw.githubusercontent.com/aroige/vikunja/${VIKUNJA_GITHUB_BRANCH}/deploy/proxmox/vikunja-install.sh)
```

**Use Local Installation (Workaround)**:
If bootstrap continues to fail, clone the repository locally:

```bash
git clone https://github.com/aroige/vikunja.git
cd vikunja/deploy/proxmox
./vikunja-install-main.sh
```

This bypasses the bootstrap download step.

---

### Issue: "curl is not installed" Error

**Symptom**:
```
[ERROR] curl is not installed. Please install curl and try again.
```

**Solution**:
Install curl on your Proxmox host:

```bash
apt-get update
apt-get install -y curl
```

Then retry the installation.

---

### Issue: "This script must be run as root" Error

**Symptom**:
```
[ERROR] This script must be run as root
```

**Solution**:
Run the installer with root privileges:

```bash
# Option 1: Use sudo
sudo bash <(curl -fsSL https://raw.githubusercontent.com/aroige/vikunja/main/deploy/proxmox/vikunja-install.sh)

# Option 2: Switch to root
su -
bash <(curl -fsSL https://raw.githubusercontent.com/aroige/vikunja/main/deploy/proxmox/vikunja-install.sh)
```

---

### Issue: "This script must be run on a Proxmox VE host" Error

**Symptom**:
```
[ERROR] This script must be run on a Proxmox VE host
```

**Cause**: The installer detects Proxmox by checking for the `pct` command (Proxmox Container Toolkit).

**Solution**:
This installer is designed **exclusively for Proxmox VE**. If you want to run Vikunja on other platforms:

- **Docker**: See https://vikunja.io/docs/docker/
- **Kubernetes**: See https://vikunja.io/docs/kubernetes/
- **Manual Installation**: See https://vikunja.io/docs/installing/

If you **are** on a Proxmox host but seeing this error:

```bash
# Verify Proxmox installation
which pct
dpkg -l | grep proxmox-ve

# If pct is missing, reinstall Proxmox Container Toolkit
apt-get install -y proxmox-ve
```

---

### Issue: Bootstrap Downloads to /tmp but Runs Out of Space

**Symptom**:
```
[ERROR] No space left on device
```

**Cause**: `/tmp` is too small to hold installer files (~50KB) and build artifacts.

**Solution**:

**Check /tmp Space**:
```bash
df -h /tmp
```

**Clean Up /tmp**:
```bash
# Remove old Vikunja installer directories
rm -rf /tmp/vikunja-installer-*

# Remove other temporary files
apt-get clean
```

**Increase /tmp Size** (if using tmpfs):
```bash
# Check if /tmp is tmpfs
mount | grep /tmp

# Increase tmpfs size (example: 1GB)
mount -o remount,size=1G /tmp
```

**Use Different Temporary Directory**:
The bootstrap uses `/tmp` by default, but you can modify this by downloading and editing the bootstrap script:

```bash
curl -fsSL https://raw.githubusercontent.com/aroige/vikunja/main/deploy/proxmox/vikunja-install.sh -o /root/vikunja-install.sh
# Edit the script and change INSTALL_DIR variable
nano /root/vikunja-install.sh
# Change: readonly INSTALL_DIR="/tmp/vikunja-installer-$$"
# To:     readonly INSTALL_DIR="/var/tmp/vikunja-installer-$$"
bash /root/vikunja-install.sh
```

---

### Issue: Download Works but Installer Immediately Exits

**Symptom**:
Bootstrap downloads all files but installer exits without output or with cryptic error.

**Possible Causes**:
1. Permission issues on `/tmp/vikunja-installer-*`
2. Bash version incompatibility (needs Bash 4.0+)
3. Missing library file during download

**Solutions**:

**Check Bash Version**:
```bash
bash --version
# Should show: GNU bash, version 4.0 or higher
```

If Bash is too old:
```bash
apt-get update
apt-get install -y bash
```

**Check Downloaded Files**:
```bash
# Find the most recent installer directory
ls -lt /tmp/ | grep vikunja-installer

# List contents
ls -lR /tmp/vikunja-installer-<PID>/

# Verify all files were downloaded:
# - vikunja-install-main.sh
# - lib/*.sh (6 files)
# - templates/* (5 files)
```

If any files are missing, the download failed partially. Check network stability and retry.

**Enable Debug Output**:
```bash
# Run bootstrap with debug
set -x
bash <(curl -fsSL https://raw.githubusercontent.com/aroige/vikunja/main/deploy/proxmox/vikunja-install.sh)
```

This shows exactly what the script is doing.

---

### Issue: Custom Branch/Fork Not Found

**Symptom**:
```
[ERROR] Failed to download: vikunja-install-main.sh
```

When using custom `VIKUNJA_GITHUB_OWNER`, `VIKUNJA_GITHUB_REPO`, or `VIKUNJA_GITHUB_BRANCH`.

**Solution**:

**Verify Environment Variables**:
```bash
echo "Owner: $VIKUNJA_GITHUB_OWNER"
echo "Repo: $VIKUNJA_GITHUB_REPO"
echo "Branch: $VIKUNJA_GITHUB_BRANCH"

# Construct URL manually
echo "URL: https://raw.githubusercontent.com/${VIKUNJA_GITHUB_OWNER}/${VIKUNJA_GITHUB_REPO}/${VIKUNJA_GITHUB_BRANCH}/deploy/proxmox/vikunja-install-main.sh"
```

**Test URL in Browser**:
Copy the constructed URL and paste into a web browser. If you get "404 Not Found", the file doesn't exist at that location.

**Check Branch Exists**:
```bash
git ls-remote https://github.com/${VIKUNJA_GITHUB_OWNER}/${VIKUNJA_GITHUB_REPO}.git | grep ${VIKUNJA_GITHUB_BRANCH}
```

**Common Mistakes**:
- Branch name has typo (e.g., `004-proxmox-deployment` vs `proxmox-deployment`)
- Files not yet pushed to remote repository
- Repository is private (bootstrap can only download from public repos)

---

## Pre-Flight Check Failures

The installer runs pre-flight checks before creating the container.

### Issue: "Insufficient resources" Error

**Symptom**:
```
[ERROR] Insufficient resources: Need 4GB RAM, have 2GB available
```

**Solution**:

**Check Available Resources**:
```bash
# Memory
free -h

# CPU cores
nproc

# Disk space
df -h
```

**Free Up Resources**:
```bash
# Stop unused VMs/containers
pct list
qm list

# Stop a container
pct stop <container-id>

# Stop a VM
qm stop <vm-id>
```

**Reduce Resource Requirements**:
When prompted during installation, allocate fewer resources:
- Minimum: 2 CPU cores, 4GB RAM, 20GB disk
- For testing: 1 CPU core, 2GB RAM (performance will suffer)

---

### Issue: "Port already in use" Error

**Symptom**:
```
[ERROR] Port 8080 already in use by container 101
```

**Solution**:

**Check Port Usage**:
```bash
# Find what's using the port
netstat -tulpn | grep :8080
ss -tulpn | grep :8080

# Check Proxmox containers
pct list
```

**Options**:
1. **Stop Conflicting Service**: If port is used by another Vikunja instance or service you don't need
   ```bash
   pct stop <container-id>
   # or
   systemctl stop <service-name>
   ```

2. **Use Different Instance ID**: Vikunja supports multiple instances with different ports
   ```bash
   # Each instance gets unique ports (8080+offset)
   vikunja-install.sh --instance-id vikunja-team
   ```

3. **Manually Specify Container Ports**: Advanced option (modify deployment config)

---

### Issue: "Domain name already in use" Error

**Symptom**:
```
[WARN] Domain vikunja.example.com already in use by another instance
```

**Solution**:

**Check Existing Instances**:
```bash
vikunja-manage.sh list
```

**Options**:
1. **Use Different Domain**: Each instance needs a unique domain
   ```
   Domain [vikunja.local]: vikunja-team.example.com
   ```

2. **Uninstall Existing Instance**: If you want to replace it
   ```bash
   vikunja-manage.sh uninstall --instance vikunja-main
   ```

3. **Use Subdomain or Port**: 
   - `vikunja-prod.example.com`
   - `vikunja-dev.example.com`
   - `vikunja.example.com:8443` (custom port)

---

## Container Creation Issues

### Issue: LXC Container Fails to Create

**Symptom**:
```
[ERROR] Failed to create container 100
```

**Solutions**:

**Check Container ID Conflicts**:
```bash
# List existing containers
pct list

# Check if ID is in use
pct status <container-id>
```

**Check Proxmox Storage**:
```bash
# List available storage
pvesm status

# Check specific storage pool
pvesm list <storage-pool>
```

**Check Debian Template**:
```bash
# List available templates
pveam available | grep debian-12

# Download template if missing
pveam download local debian-12-standard_12.0-1_amd64.tar.zst
```

**Check Logs**:
```bash
journalctl -xe | grep pct
tail -f /var/log/pve/tasks/active
```

---

### Issue: Container Starts but Network Not Working

**Symptom**:
Container is running but cannot access internet or is unreachable.

**Solutions**:

**Check Container Network Config**:
```bash
# View container config
pct config <container-id>

# Should show net0 line:
# net0: name=eth0,bridge=vmbr0,ip=192.168.1.100/24,gw=192.168.1.1
```

**Test from Container**:
```bash
# Enter container
pct enter <container-id>

# Test network
ip addr
ip route
ping -c 3 8.8.8.8
ping -c 3 google.com
```

**Common Issues**:
- **Incorrect Gateway**: Verify gateway IP is correct for your network
- **Bridge Not Created**: Check `vmbr0` exists on Proxmox host
  ```bash
  ip addr show vmbr0
  ```
- **Firewall Blocking**: Check Proxmox firewall settings
  ```bash
  pve-firewall status
  ```

---

## Build Failures

### Issue: Backend Build Fails

**Symptom**:
```
[ERROR] Backend build failed: go build error
```

**Solutions**:

**Check Go Version**:
```bash
pct exec <container-id> -- go version
# Should be 1.21 or higher
```

**Check Disk Space in Container**:
```bash
pct exec <container-id> -- df -h
```

**Check Build Logs**:
```bash
pct exec <container-id> -- cat /opt/vikunja-<instance>/logs/build.log
```

**Common Causes**:
- **Out of Memory**: Increase container RAM to 4GB minimum
- **Out of Disk Space**: Increase container disk to 20GB minimum
- **Network Issues**: Go modules failing to download
  ```bash
  pct exec <container-id> -- go env GOPROXY
  # Should be: https://proxy.golang.org,direct
  ```

**Retry Build**:
```bash
# Enter container
pct enter <container-id>

# Navigate to source
cd /opt/vikunja-<instance>/vikunja

# Clean and rebuild
make clean
mage build
```

---

### Issue: Frontend Build Fails

**Symptom**:
```
[ERROR] Frontend build failed: npm error
```

**Solutions**:

**Check Node.js Version**:
```bash
pct exec <container-id> -- node --version
pct exec <container-id> -- pnpm --version
# Node 18+ and pnpm required
```

**Check npm/pnpm Cache**:
```bash
pct exec <container-id> -- pnpm cache clean
```

**Increase Memory**:
Frontend builds are memory-intensive. Increase to 8GB if build fails:
```bash
pct set <container-id> --memory 8192
pct reboot <container-id>
```

**Retry Build**:
```bash
pct enter <container-id>
cd /opt/vikunja-<instance>/vikunja/frontend
rm -rf node_modules dist
pnpm install
pnpm build
```

---

### Issue: MCP Server Build Fails

**Symptom**:
```
[ERROR] MCP server build failed
```

**Solution**:

**Check TypeScript Version**:
```bash
pct exec <container-id> -- pnpm --version
```

**Rebuild**:
```bash
pct enter <container-id>
cd /opt/vikunja-<instance>/vikunja/mcp-server
pnpm clean
pnpm build
```

---

## Service Startup Issues

### Issue: Backend Service Won't Start

**Symptom**:
```
[ERROR] vikunja-backend-blue.service failed to start
```

**Solutions**:

**Check Service Status**:
```bash
pct exec <container-id> -- systemctl status vikunja-backend-blue
```

**Check Logs**:
```bash
pct exec <container-id> -- journalctl -u vikunja-backend-blue -n 50
```

**Common Causes**:
- **Port Already in Use**: Check with `netstat -tulpn | grep 8080`
- **Database Connection Failed**: Check database credentials in config
- **Missing Config File**: Check `/opt/vikunja-<instance>/config/config.yml` exists
- **Permission Issues**: Ensure vikunja user owns files
  ```bash
  pct exec <container-id> -- chown -R vikunja:vikunja /opt/vikunja-<instance>
  ```

**Restart Service**:
```bash
pct exec <container-id> -- systemctl restart vikunja-backend-blue
```

---

### Issue: Nginx Won't Start

**Symptom**:
```
[ERROR] nginx.service failed to start
```

**Solutions**:

**Test Nginx Config**:
```bash
pct exec <container-id> -- nginx -t
```

**Check Logs**:
```bash
pct exec <container-id> -- tail -f /var/log/nginx/error.log
```

**Common Issues**:
- **Port 80/443 In Use**: Check with `netstat -tulpn | grep :80`
- **SSL Certificate Invalid**: Check certificate paths in nginx config
- **Syntax Error in Config**: Run `nginx -t` to validate

**Restart Nginx**:
```bash
pct exec <container-id> -- systemctl restart nginx
```

---

## Update and Rollback Issues

### Issue: Update Hangs or Times Out

**Symptom**:
Update starts but never completes, or times out waiting for health checks.

**Solution**:

**Check Update Lock**:
```bash
ls -l /opt/vikunja-<instance>/state/locks/update.lock
```

If lock exists and process is dead:
```bash
rm /opt/vikunja-<instance>/state/locks/update.lock
```

**Check Service Status**:
```bash
systemctl status vikunja-backend-green
journalctl -u vikunja-backend-green -n 50
```

**Manual Rollback**:
If update is stuck, manually roll back:
```bash
vikunja-manage.sh rollback
```

---

### Issue: Automatic Rollback Fails

**Symptom**:
Update fails and automatic rollback also fails, leaving deployment in inconsistent state.

**Solution**:

**Restore from Backup**:
```bash
vikunja-manage.sh restore /opt/vikunja-<instance>/backups/<most-recent>.tar.gz
```

**Manual Recovery**:
```bash
# Stop all services
systemctl stop vikunja-backend-green vikunja-backend-blue
systemctl stop vikunja-mcp-green vikunja-mcp-blue

# Restore database from pre-migration backup
cd /opt/vikunja-<instance>/backups
tar -xzf pre-migration-*.tar.gz

# Start blue services
systemctl start vikunja-backend-blue vikunja-mcp-blue

# Update nginx to point to blue
# Edit /etc/nginx/sites-available/vikunja-<instance>
# Change upstream ports back to blue (8080, 3456)
systemctl reload nginx
```

---

## Network and Connectivity Issues

### Issue: Cannot Access Vikunja Web Interface

**Symptom**:
Installation succeeds but cannot access `https://vikunja.example.com` in browser.

**Solutions**:

**Check DNS Resolution**:
```bash
nslookup vikunja.example.com
# Should resolve to your Proxmox host IP
```

**Check Firewall**:
```bash
# On Proxmox host
iptables -L -n | grep -E '(80|443)'

# Allow HTTP/HTTPS if needed
iptables -A INPUT -p tcp --dport 80 -j ACCEPT
iptables -A INPUT -p tcp --dport 443 -j ACCEPT
```

**Check Nginx Status**:
```bash
pct exec <container-id> -- systemctl status nginx
pct exec <container-id> -- curl -I http://127.0.0.1:80
```

**Check SSL Certificate**:
```bash
pct exec <container-id> -- openssl s_client -connect vikunja.example.com:443
```

**Test from Proxmox Host**:
```bash
curl -I http://<container-ip>:80
curl -k -I https://<container-ip>:443
```

**Browser Shows "Connection Refused"**:
- Verify firewall allows traffic to Proxmox host
- Verify NAT/port forwarding if Proxmox is behind router
- Check browser isn't blocking self-signed certificate (use `https://` explicitly)

---

## Database Connection Issues

### Issue: "Database connection failed" Error

**Symptom**:
```
[ERROR] Database connection failed: dial tcp 192.168.1.50:5432: connection refused
```

**Solutions**:

**For SQLite**:
```bash
# Check database file exists
ls -l /opt/vikunja-<instance>/data/vikunja.db

# Check permissions
chown vikunja:vikunja /opt/vikunja-<instance>/data/vikunja.db
```

**For PostgreSQL**:
```bash
# Test connection from container
pct enter <container-id>
psql -h <db-host> -U <db-user> -d <db-name>

# Check pg_hba.conf allows connections from container IP
# On PostgreSQL server:
sudo nano /etc/postgresql/*/main/pg_hba.conf
# Add: host all all <container-ip>/32 md5
sudo systemctl restart postgresql
```

**For MySQL**:
```bash
# Test connection
pct enter <container-id>
mysql -h <db-host> -u <db-user> -p<db-password> <db-name>

# Check MySQL allows remote connections
# On MySQL server:
sudo nano /etc/mysql/mysql.conf.d/mysqld.cnf
# Ensure: bind-address = 0.0.0.0
sudo systemctl restart mysql

# Grant permissions
mysql -u root -p
GRANT ALL PRIVILEGES ON vikunja.* TO 'vikunja'@'%' IDENTIFIED BY 'password';
FLUSH PRIVILEGES;
```

---

## Health Check Failures

### Issue: Health Checks Fail but Services Are Running

**Symptom**:
```
[ERROR] Health check failed: Backend API unhealthy
```

But `systemctl status vikunja-backend-blue` shows running.

**Solutions**:

**Test Health Check Manually**:
```bash
pct enter <container-id>
/opt/vikunja-<instance>/scripts/health-check.sh 8080 8082 3456
```

**Test Individual Components**:
```bash
# Backend API
curl http://127.0.0.1:8080/api/v1/info

# Frontend
curl http://127.0.0.1:8082/

# MCP server
nc -zv 127.0.0.1 3456
```

**Common Causes**:
- Services still starting (wait 30 seconds)
- Port binding failed (check with `netstat -tulpn`)
- Backend crashed after start (check logs)

---

## Performance Issues

### Issue: Slow Build Times (>15 minutes)

**Solution**:
- Increase container CPU to 4 cores during build
- Check internet speed for git/npm downloads
- Use local git mirror or npm cache proxy

### Issue: High Memory Usage

**Solution**:
- Restart services periodically
- Tune Vikunja database cache settings
- Use external database instead of SQLite for large datasets

### Issue: Slow Frontend Loading

**Solution**:
- Enable gzip compression in nginx
- Use CDN for static assets (advanced)
- Check nginx access logs for bottlenecks

---

## Backup and Restore Issues

### Issue: Backup Fails with "No space left on device"

**Solution**:
```bash
# Check disk space
df -h

# Clean old backups
vikunja-manage.sh backup --cleanup

# Manually remove old backups
rm /opt/vikunja-<instance>/backups/vikunja-backup-*.tar.gz
```

### Issue: Restore Fails with "Database restore failed"

**Solution**:
- Ensure database service is running
- Check database credentials match backup
- Manually restore database:
  ```bash
  tar -xzf backup.tar.gz
  # For PostgreSQL:
  psql -U vikunja -d vikunja < database_dump.sql
  # For SQLite:
  cp database/vikunja.db /opt/vikunja-<instance>/data/
  ```

---

## Getting Help

If you've tried these solutions and still have issues:

1. **Enable Debug Logging**:
   ```bash
   vikunja-install.sh --debug
   vikunja-manage.sh --debug status
   ```

2. **Collect Diagnostic Information**:
   ```bash
   # System info
   uname -a
   pveversion
   
   # Container info
   pct config <container-id>
   
   # Service status
   systemctl status vikunja-*
   
   # Logs
   journalctl -u vikunja-backend-blue -n 100
   ```

3. **Check Documentation**:
   - [Vikunja Documentation](https://vikunja.io/docs)
   - [Proxmox VE Documentation](https://pve.proxmox.com/pve-docs/)
   - [Architecture Guide](ARCHITECTURE.md)

4. **Report Issues**:
   - [Vikunja GitHub Issues](https://github.com/go-vikunja/vikunja/issues)
   - Include system info, logs, and steps to reproduce

---

**Last Updated**: 2025-10-19  
**Version**: 1.0.0  
**Maintainer**: Vikunja Deployment System
