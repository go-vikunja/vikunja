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

### Installation
```bash
vikunja-install.sh              # Interactive installation
vikunja-install.sh --help       # Show all options
```

### Updates
```bash
vikunja-update.sh               # Update to latest main branch
vikunja-update.sh --check       # Check for updates without installing
```

### Management
```bash
vikunja-manage.sh status        # Show deployment status
vikunja-manage.sh reconfigure   # Change configuration
vikunja-manage.sh backup        # Create backup
vikunja-manage.sh restore       # Restore from backup
vikunja-manage.sh logs          # View logs
vikunja-manage.sh restart       # Restart services
vikunja-manage.sh stop          # Stop services
vikunja-manage.sh start         # Start services
vikunja-manage.sh uninstall     # Remove deployment
vikunja-manage.sh list          # List all instances
```

## Documentation

- **[Quickstart Guide](../../specs/004-proxmox-deployment/quickstart.md)** - Complete setup instructions
- **[Architecture](docs/ARCHITECTURE.md)** - Blue-green deployment pattern (coming soon)
- **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common issues and solutions (coming soon)
- **[Development](docs/DEVELOPMENT.md)** - Testing and contribution guidelines (coming soon)

## Architecture

This deployment system uses:
- **Bootstrap Installer**: Single-command curl-based installation that downloads all required components
- **LXC Containers**: Lightweight, secure unprivileged containers on Proxmox
- **Blue-Green Deployment**: Zero-downtime updates with automatic rollback
- **Systemd Services**: vikunja-backend.service, vikunja-mcp.service
- **Nginx Reverse Proxy**: SSL termination, upstream switching for blue-green
- **State Management**: YAML configuration, lock files, version tracking

### How the Bootstrap Installer Works

The single-command installation uses a three-stage bootstrap pattern:

1. **Stage 1 (Bootstrap)**: You run the curl command, which executes a lightweight bootstrap script
2. **Stage 2 (Download)**: The bootstrap downloads all required files (main installer, libraries, templates) to a temporary directory (`/tmp/vikunja-installer-<PID>`)
3. **Stage 3 (Execute)**: The bootstrap launches the full installer with all dependencies available locally

This pattern enables single-command installation while maintaining modular code architecture. It matches industry-standard patterns used by Docker, Kubernetes, and other infrastructure tools.

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
│   ├── lxc-setup.sh             # Container provisioning
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

## Performance Targets

- **Initial Deployment**: <10 minutes
- **Updates**: <5 minutes (including database migrations)
- **Downtime During Updates**: <5 seconds
- **Health Checks**: <10 seconds
- **Backups**: <5 minutes (for 10k tasks + 1GB attachments)
- **Rollback**: <2 minutes

## Troubleshooting

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
