# Vikunja Proxmox LXC Deployment

**Automated deployment and lifecycle management for Vikunja on Proxmox Virtual Environment**

This deployment system provides single-command installation, zero-downtime updates, and comprehensive lifecycle management for Vikunja running in LXC containers on Proxmox.

## Features

- 🚀 **Single-Command Deployment**: Install Vikunja in under 10 minutes
- 🔄 **Zero-Downtime Updates**: Blue-green deployment with automatic rollback
- 🗄️ **Flexible Database Options**: SQLite, PostgreSQL, or MySQL
- 🔍 **Health Monitoring**: Component-level health checks and status reporting
- 💾 **Backup & Restore**: Encrypted backups with easy restoration
- ⚙️ **Configuration Management**: Reconfigure without redeployment
- 🔒 **Secure by Default**: Unprivileged LXC containers, SSL/TLS support
- 📦 **Multi-Instance Support**: Run multiple Vikunja instances on one Proxmox cluster

## Quick Start

**Prerequisites**: Proxmox VE 7.0+, root access, 2 CPU cores, 4GB RAM, 20GB disk available

**Install Vikunja** (one command):

**Stable version (once merged):**
```bash
bash <(curl -fsSL https://raw.githubusercontent.com/aroige/vikunja/main/deploy/proxmox/vikunja-install.sh)
```

**Development/testing version (current branch):**
```bash
VIKUNJA_GITHUB_BRANCH="004-proxmox-deployment" bash <(curl -fsSL https://raw.githubusercontent.com/aroige/vikunja/004-proxmox-deployment/deploy/proxmox/vikunja-install.sh)
```

**See the [Quickstart Guide](../../specs/004-proxmox-deployment/quickstart.md) for complete setup instructions.**

## Commands

### Installation (Run from Proxmox Host)
```bash
# One-command installation - no persistent files left on host
bash <(curl -fsSL https://raw.githubusercontent.com/aroige/vikunja/main/deploy/proxmox/vikunja-install.sh)

# For development/testing, specify branch:
VIKUNJA_GITHUB_BRANCH="004-proxmox-deployment" bash <(curl -fsSL https://raw.githubusercontent.com/aroige/vikunja/004-proxmox-deployment/deploy/proxmox/vikunja-install.sh)
```

### Updates (Run Inside Container)

**Primary method** - SSH into the container:
```bash
# SSH into container
ssh root@<container-ip>

# Run update (updates both Vikunja and deployment scripts)
vikunja-update              # Update to latest main branch
vikunja-update --help       # Show all options
```

**Alternative method** - From Proxmox host:
```bash
# Find container ID
pct list | grep vikunja

# Enter container console
pct enter <container-id>

# Then run update
vikunja-update
```

**Quick trigger from host** (without entering container):
```bash
pct exec <container-id> /opt/vikunja-deploy/vikunja-update.sh
```

### Management (Coming Soon)
The following commands will be available inside the container:
```bash
vikunja-manage status           # Show deployment status
vikunja-manage reconfigure      # Change configuration  
vikunja-manage backup           # Create backup
vikunja-manage restore          # Restore from backup
vikunja-manage logs             # View logs
vikunja-manage restart          # Restart services
```

**Note**: Management commands are planned but not yet implemented. Currently, you can use systemd commands directly:
```bash
# Inside container
systemctl status vikunja-backend-blue
systemctl restart vikunja-backend-blue
journalctl -u vikunja-backend-blue -f
```

### Troubleshooting Tools (Inside Container)

**Configuration Validator** - Detect and diagnose configuration issues:
```bash
# Check Vikunja database configuration
vikunja-config-check

# Output shows:
# - Config file locations checked
# - Environment variables detected
# - Systemd service configuration
# - Database type and connection details
# - Recommendations for missing config
```

This tool is especially useful when:
- Import/export operations use the wrong database
- Seeing "No config file found" warnings
- Verifying configuration after installation or changes
- Troubleshooting database connection issues

See the [Database Configuration Troubleshooting Guide](docs/TROUBLESHOOTING.md#database-importexport-configuration-issues) for detailed examples.

## Documentation

- **[Quickstart Guide](../../specs/004-proxmox-deployment/quickstart.md)** - Complete setup instructions
- **[Architecture](docs/ARCHITECTURE.md)** - Blue-green deployment pattern (coming soon)
- **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common issues and solutions (coming soon)
- **[Development](docs/DEVELOPMENT.md)** - Testing and contribution guidelines (coming soon)

## Architecture

This deployment system uses a **container-centric self-sufficient pattern** with **zero host footprint**:

### Key Design Principles

- **Container is fully portable**: Everything needed lives inside the container at `/opt/vikunja-deploy/`
- **Zero host pollution**: Bootstrap downloads to `/tmp` and cleans up automatically
- **Self-updating scripts**: Update script pulls latest version of deployment tools before updating Vikunja
- **Configuration travels with container**: All config stored in container at `/etc/vikunja/deploy-config.yaml`
- **Migration-friendly**: Container can be migrated between Proxmox hosts without losing management capability

### Components

- **Bootstrap Installer**: Temporary single-command installation (no persistent files on host)
- **LXC Containers**: Lightweight, secure unprivileged containers on Proxmox
- **Blue-Green Deployment**: Zero-downtime updates with automatic rollback
- **Systemd Services**: vikunja-backend.service, vikunja-mcp.service (inside container)
- **Nginx Reverse Proxy**: SSL termination, upstream switching for blue-green (inside container)
- **Management Scripts**: Deployment and update tools (inside container at `/opt/vikunja-deploy/`)
- **Configuration**: YAML deployment config (inside container at `/etc/vikunja/deploy-config.yaml`)

### How It Works

**Installation (from Proxmox host):**
1. **Bootstrap**: Run curl command - downloads installer to `/tmp` on Proxmox host
2. **Container Creation**: Creates LXC container and provisions it
3. **Script Installation**: Copies all management scripts **into the container** at `/opt/vikunja-deploy/`
4. **Cleanup**: Bootstrap temp files removed from host

**Updates (from inside container):**
1. **Self-Update**: Update script first updates itself from GitHub
2. **Vikunja Update**: Then updates Vikunja application code
3. **Blue-Green Deploy**: Zero-downtime deployment with health checks

**Result:** Container is self-contained and portable. No permanent files on Proxmox host. Can be migrated freely between hosts.

This architecture enables true portability while maintaining the convenience of single-command deployment.

**Advanced Usage**: You can customize the installation source using environment variables:

```bash
# Install from a specific branch or fork
export VIKUNJA_GITHUB_OWNER="yourname"
export VIKUNJA_GITHUB_REPO="vikunja"
export VIKUNJA_GITHUB_BRANCH="feature-branch"

bash <(curl -fsSL https://raw.githubusercontent.com/${VIKUNJA_GITHUB_OWNER}/${VIKUNJA_GITHUB_REPO}/${VIKUNJA_GITHUB_BRANCH}/deploy/proxmox/vikunja-install.sh)
```

**Local Installation**: If you prefer to run the installer from a local git clone:

```bash
git clone https://github.com/aroige/vikunja.git
cd vikunja/deploy/proxmox
./vikunja-install-main.sh
```

## Project Structure

```
deploy/proxmox/
├── vikunja-install.sh           # Bootstrap installer (curl-able entry point)
├── vikunja-install-main.sh      # Main installation script (downloaded by bootstrap)
├── vikunja-update.sh            # Update script
├── vikunja-manage.sh            # Management commands
├── lib/                         # Shared library functions
│   ├── common.sh                # Logging, validation, error handling
│   ├── proxmox-api.sh           # Proxmox CLI wrappers
│   ├── lxc-setup.sh             # Container provisioning, root access setup
│   ├── service-setup.sh         # Systemd service management
│   ├── nginx-setup.sh           # Nginx configuration
│   ├── blue-green.sh            # Blue-green deployment logic
│   ├── backup-restore.sh        # Backup/restore operations
│   └── health-check.sh          # Health monitoring
├── templates/                   # Configuration templates
│   ├── deployment-config.yaml   # Deployment settings template
│   ├── vikunja-backend.service  # Systemd unit template
│   ├── vikunja-mcp.service      # MCP systemd unit template
│   ├── nginx-vikunja.conf       # Nginx site configuration
│   └── health-check.sh          # Health check script (deployed to container)
├── tests/                       # Integration tests
│   ├── integration/             # Full deployment cycle tests
│   └── fixtures/                # Test fixtures and mocks
└── docs/                        # Documentation
    ├── ARCHITECTURE.md          # System design and patterns
    ├── TROUBLESHOOTING.md       # Common issues
    └── DEVELOPMENT.md           # Development guide
```

## Root Access Configuration

The installer provides flexible root access configuration for the LXC container with professional security features:

### Interactive Mode

During installation, you'll be prompted to choose root access method:

1. **Password only** - Simple password authentication (less secure, convenient for testing)
2. **SSH key only** - Key-based authentication (recommended for production) ⭐
3. **Both password and SSH key** - Flexible access with moderate security
4. **Auto-generated password, no SSH** - Random password generated, access via `pct enter`

### Non-Interactive Mode (CLI Options)

```bash
# SSH key authentication (recommended for production)
vikunja-install.sh --non-interactive \
  --root-ssh-key ~/.ssh/id_ed25519.pub \
  --disable-root-password \
  --domain vikunja.example.com \
  --ip-address 192.168.1.100/24 \
  --gateway 192.168.1.1

# Password-only authentication
vikunja-install.sh --non-interactive \
  --root-password "SecurePassword123!" \
  --enable-root-password \
  --domain vikunja.example.com \
  --ip-address 192.168.1.100/24 \
  --gateway 192.168.1.1

# Both password and SSH key
vikunja-install.sh --non-interactive \
  --root-password "SecurePassword123!" \
  --root-ssh-key ~/.ssh/id_rsa.pub \
  --enable-root-password \
  --domain vikunja.example.com \
  --ip-address 192.168.1.100/24 \
  --gateway 192.168.1.1
```

### Security Features

The root access configuration includes professional security hardening:

- ✅ **Cryptographically secure password generation** - Uses OpenSSL for 32-character random passwords
- ✅ **SSH public key injection** - Validates and injects SSH keys into authorized_keys with proper permissions
- ✅ **SSH key format validation** - Validates key format before injection (ssh-rsa, ssh-ed25519, ecdsa-sha2-, etc.)
- ✅ **Automatic permission management** - Sets correct ownership and permissions (700 for .ssh, 600 for authorized_keys)
- ✅ **SSH daemon hardening** - Disables password authentication when keys are used, disables empty passwords, disables X11 forwarding
- ✅ **Console access preserved** - Root password always set for emergency console access via `pct enter`
- ✅ **Flexible authentication modes** - Support for password-only, key-only, or both authentication methods

### Security Best Practices

**For Production Deployments:**
- ✅ **Use SSH key authentication only** (`--disable-root-password`)
- ✅ **Use Ed25519 keys** (strongest, fastest): `ssh-keygen -t ed25519`
- ✅ **Protect SSH private keys** with strong passphrases
- ✅ **Rotate keys regularly** and audit authorized_keys
- ✅ **Use certificate authorities** for SSH at scale

**For Development/Testing:**
- ⚠️ **Password authentication acceptable** for local testing
- ⚠️ **Auto-generated passwords** are secure but must be saved
- ⚠️ **Console access via `pct enter`** always available regardless of SSH config

### Access Methods After Installation

**SSH Access** (if configured):
```bash
ssh root@<container-ip>
```

**Console Access** (always available):
```bash
pct enter <container-id>
```

**From Proxmox Web UI**:
Navigate to container → Console

### CLI Options Reference

| Option | Description | Default |
|--------|-------------|---------|
| `--root-password PASS` | Set root password | Auto-generated secure random |
| `--root-ssh-key FILE` | Path to SSH public key file | None |
| `--enable-root-password` | Enable SSH password authentication | Auto (enabled if no key) |
| `--disable-root-password` | Disable SSH password authentication | Auto (disabled if key provided) |

### Supported SSH Key Types

- ✅ RSA keys (minimum 2048 bits): `ssh-rsa ...`
- ✅ Ed25519 keys (recommended): `ssh-ed25519 ...`
- ✅ ECDSA keys: `ecdsa-sha2-nistp256 ...`, `ecdsa-sha2-nistp384 ...`, `ecdsa-sha2-nistp521 ...`
- ✅ FIDO/U2F keys: `sk-ssh-ed25519@openssh.com ...`, `sk-ecdsa-sha2-nistp256@openssh.com ...`

### Troubleshooting Root Access

**Problem**: Cannot SSH to container

**Solutions**:
1. Verify SSH service is running: `pct exec <id> systemctl status sshd`
2. Check firewall rules don't block port 22
3. Verify SSH key was injected correctly: `pct exec <id> cat /root/.ssh/authorized_keys`
4. Check SSH daemon configuration: `pct exec <id> cat /etc/ssh/sshd_config | grep -E '(PermitRootLogin|PasswordAuthentication)'`
5. Use console access as fallback: `pct enter <id>`

**Problem**: Lost root password

**Solutions**:
1. Use console access: `pct enter <container-id>` (always works)
2. Reset password from host: `pct exec <id> bash -c "echo 'root:newpassword' | chpasswd"`
3. Inject new SSH key from host (see vikunja-manage.sh reconfigure)

## Requirements

**Proxmox Host**:
- Proxmox VE 7.0 or later
- Internet connectivity
- Root access via SSH

**Per Vikunja Instance**:
- 2 CPU cores (minimum)
- 4GB RAM (minimum)
- 20GB disk space (minimum)
- Unique container ID (100-999)
- Available ports: 8080 (backend), 3456 (MCP), 80/443 (nginx)

**External** (optional):
- PostgreSQL or MySQL server (if not using SQLite)
- Domain name with DNS configured
- SSL certificate (can use self-signed)

## MCP HTTP Transport

The Vikunja MCP Server supports **remote client connections** via HTTP transport, enabling integration with AI workflow tools like n8n, Claude Desktop, and other MCP clients over the network.

### Features

- **HTTP Streamable Protocol**: Modern bidirectional HTTP protocol (primary)
- **SSE Transport**: Server-Sent Events for backward compatibility (deprecated)
- **Token-Based Authentication**: Vikunja API tokens for secure access
- **Rate Limiting**: Prevent abuse with configurable per-token limits (100 req/15min default)
- **Session Management**: Automatic cleanup of idle/expired sessions
- **Health Monitoring**: `/health` endpoint for monitoring and load balancing

### Configuration

**Environment Variables** (configured via systemd service):

```bash
# HTTP Transport (disabled by default for stdio mode)
MCP_HTTP_ENABLED=false          # Enable HTTP transport (true/false)
MCP_HTTP_PORT=3100              # HTTP server port
MCP_HTTP_HOST=127.0.0.1         # Bind address (use 0.0.0.0 for network access)

# Redis (required when HTTP transport enabled)
REDIS_URL=redis://localhost:6379

# Authentication
AUTH_TOKEN_CACHE_TTL=300        # Token cache TTL in seconds (5 minutes)

# Rate Limiting
RATE_LIMIT_POINTS=100           # Max requests per window
RATE_LIMIT_DURATION=900         # Window duration in seconds (15 minutes)

# Session Management
SESSION_IDLE_TIMEOUT=1800       # Session timeout in seconds (30 minutes)
SESSION_CLEANUP_INTERVAL=300    # Cleanup interval in seconds (5 minutes)
```

### Enabling HTTP Transport

**Default Installation**: HTTP transport is **disabled** by default. The MCP server runs in stdio mode for local Claude Desktop integration.

**To Enable HTTP Transport**:

1. **Edit the deployment configuration** (inside container):
   ```bash
   vim /etc/vikunja/deploy-config.yaml
   ```
   
   Add or update:
   ```yaml
   mcp_http_enabled: "true"
   mcp_http_port: 3100
   mcp_http_host: "0.0.0.0"  # Use 0.0.0.0 for network access
   redis_url: "redis://localhost:6379"
   ```

2. **Regenerate and restart the service**:
   ```bash
   # Regenerate systemd unit with new config
   vikunja-manage reconfigure

   # Or manually restart
   systemctl daemon-reload
   systemctl restart vikunja-mcp-blue
   ```

3. **Verify HTTP transport is running**:
   ```bash
   curl http://localhost:3100/health
   ```
   
   Expected response:
   ```json
   {
     "status": "healthy",
     "timestamp": "2025-10-23T16:00:00.000Z",
     "uptime": 12345,
     "checks": {
       "vikunja": { "status": "healthy" },
       "redis": { "status": "healthy" }
     },
     "sessions": {
       "active": 0,
       "total": 0
     }
   }
   ```

### Client Connection Examples

**n8n MCP Integration**:
```json
{
  "mcpServers": {
    "vikunja": {
      "url": "http://vikunja-server:3100/mcp",
      "transport": "http-streamable",
      "headers": {
        "Authorization": "Bearer your-vikunja-api-token"
      }
    }
  }
}
```

**Claude Desktop** (HTTP mode):
```json
{
  "mcpServers": {
    "vikunja": {
      "url": "http://vikunja-server:3100/mcp",
      "transport": "http-streamable",
      "auth": {
        "type": "bearer",
        "token": "your-vikunja-api-token"
      }
    }
  }
}
```

**SSE Transport** (deprecated, for backward compatibility):
```json
{
  "mcpServers": {
    "vikunja": {
      "url": "http://vikunja-server:3100/sse",
      "transport": "sse",
      "token": "your-vikunja-api-token"
    }
  }
}
```

### Security Considerations

1. **Network Access**: By default, `MCP_HTTP_HOST=127.0.0.1` (local only). Use `0.0.0.0` for network access.

2. **Reverse Proxy**: For production, use nginx/Caddy with HTTPS:
   ```nginx
   location /mcp {
       proxy_pass http://localhost:3100;
       proxy_http_version 1.1;
       proxy_set_header Connection "";
       proxy_buffering off;
   }
   ```

3. **Firewall**: Ensure port 3100 is **not exposed** to the internet unless behind a reverse proxy with HTTPS.

4. **Token Security**: Use dedicated API tokens with minimal required permissions.

5. **Rate Limiting**: Configured per-token to prevent abuse (100 requests per 15 minutes default).

### Testing HTTP Transport

**Health Check**:
```bash
curl http://localhost:3100/health
```

**List Available Tools** (requires authentication):
```bash
curl -X POST http://localhost:3100/mcp \
  -H "Authorization: Bearer your-vikunja-api-token" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'
```

**Expected Response**:
```json
{
  "jsonrpc": "2.0",
  "result": {
    "tools": [
      {"name": "get_project_tasks", "description": "Get tasks for a project", ...},
      {"name": "create_task", "description": "Create a new task", ...},
      ...
    ]
  },
  "id": 1
}
```

### Monitoring

**Check Active Sessions**:
```bash
curl http://localhost:3100/health | jq '.sessions'
```

**View Logs**:
```bash
journalctl -u vikunja-mcp-blue -f
```

**Redis Connection**:
```bash
redis-cli ping  # Should return "PONG"
redis-cli info clients
```

### Troubleshooting

**Problem**: "Connection refused" on port 3100

**Solutions**:
- Verify `MCP_HTTP_ENABLED=true` in service config
- Check service status: `systemctl status vikunja-mcp-blue`
- View logs: `journalctl -u vikunja-mcp-blue`
- Verify port binding: `netstat -tlnp | grep 3100`

**Problem**: "Redis connection failed"

**Solutions**:
- Check Redis is running: `systemctl status redis-server`
- Verify Redis port: `redis-cli ping`
- Check REDIS_URL in service config

**Problem**: "Unauthorized" errors

**Solutions**:
- Verify Vikunja API token is valid
- Check token has required permissions in Vikunja UI
- Verify token is sent in `Authorization: Bearer <token>` header

**Problem**: Rate limit errors (429)

**Solutions**:
- Check `RATE_LIMIT_POINTS` and `RATE_LIMIT_DURATION` configuration
- View rate limit info in 429 response headers (`Retry-After`)
- Monitor requests per token in logs

### Ports Summary

| Port | Service | Default Binding | Purpose |
|------|---------|-----------------|---------|
| 8080 | Vikunja Backend | 127.0.0.1 | API + Web UI |
| 3456 | MCP Server (stdio) | N/A | Local Claude Desktop |
| 3100 | MCP Server (HTTP) | 127.0.0.1 | Remote MCP clients |
| 6379 | Redis | 127.0.0.1 | Session/token cache |
| 80/443 | Nginx | 0.0.0.0 | Reverse proxy |

**Note**: HTTP transport is **opt-in**. The default installation uses stdio mode for local Claude Desktop integration.

## Performance Targets

- **Initial Deployment**: <10 minutes
- **Updates**: <5 minutes (including database migrations)
- **Downtime During Updates**: <5 seconds
- **Health Checks**: <10 seconds
- **Backups**: <5 minutes (for 10k tasks + 1GB attachments)
- **Rollback**: <2 minutes

## Troubleshooting

### Update Workflow Issues

**Problem**: "Command not found: vikunja-update"

**Solution**: The update script runs **inside the container**, not on the Proxmox host. Make sure you've SSH'd into the container or used `pct enter <container-id>`. The command is available at `/usr/local/bin/vikunja-update` inside the container.

**Problem**: How do I access my container to run updates?

**Solutions**:
```bash
# Method 1: SSH (recommended)
ssh root@<container-ip>

# Method 2: Console access from Proxmox host
pct enter <container-id>

# Method 3: Direct command execution from Proxmox host
pct exec <container-id> /opt/vikunja-deploy/vikunja-update.sh
```

**Problem**: Where are the deployment scripts stored?

**Answer**: All deployment scripts are stored **inside the LXC container** at:
- Scripts: `/opt/vikunja-deploy/` (inside container)
- Configuration: `/etc/vikunja/deploy-config.yaml` (inside container)
- Shortcuts: `/usr/local/bin/vikunja-*` (inside container)

**Nothing is stored permanently on the Proxmox host** - this makes the container fully portable and migration-friendly.

**Problem**: I migrated my container to another Proxmox host, can I still update it?

**Solution**: Yes! Since all scripts and configuration are inside the container, it remains fully functional after migration. Just SSH into the container and run `vikunja-update` as usual.

**Problem**: How do I update the deployment scripts themselves?

**Answer**: The update script automatically self-updates before updating Vikunja. It pulls the latest version of deployment scripts from GitHub, so you're always using the current tooling.

### Bootstrap Installation Issues

**Problem**: Download fails with "Failed to download" error

**Solutions**:
- Verify internet connectivity on the Proxmox host
- Check firewall rules allow HTTPS (port 443) to raw.githubusercontent.com
- Ensure curl is installed: `apt-get install -y curl`
- Try with explicit branch: Set `VIKUNJA_GITHUB_BRANCH` environment variable

**Problem**: "This script must be run as root" error

**Solution**: Run with sudo or as root user: `sudo bash <(curl -fsSL ...)`

**Problem**: "This script must be run on a Proxmox VE host" error

**Solution**: This installer only works on Proxmox VE hosts. For other platforms, see the main Vikunja documentation.

### General Issues

For more troubleshooting guidance, see:
- [Troubleshooting Guide](docs/TROUBLESHOOTING.md) (coming soon)
- [Vikunja Documentation](https://vikunja.io/docs)
- [GitHub Issues](https://github.com/go-vikunja/vikunja/issues)

## Support & Contributing

- **Issues**: Report bugs or request features via GitHub Issues
- **Contributions**: See [DEVELOPMENT.md](docs/DEVELOPMENT.md) for contribution guidelines
- **Vikunja Project**: https://vikunja.io
- **Vikunja Documentation**: https://vikunja.io/docs

## License

This deployment system is part of the Vikunja project and is licensed under the same terms as Vikunja (AGPLv3).

See the main [LICENSE](../../LICENSE) file for details.
