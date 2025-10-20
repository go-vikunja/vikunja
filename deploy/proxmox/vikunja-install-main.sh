#!/usr/bin/env bash
# Vikunja Proxmox Deployment - Main Installation Script
# Purpose: Single-command deployment of Vikunja to Proxmox LXC
# Usage: ./vikunja-install.sh [options]

set -euo pipefail

# Script version
readonly VERSION="1.0.0"

# Determine script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source library functions
# shellcheck source=./lib/common.sh
source "${SCRIPT_DIR}/lib/common.sh"
# shellcheck source=./lib/proxmox-api.sh
source "${SCRIPT_DIR}/lib/proxmox-api.sh"
# shellcheck source=./lib/lxc-setup.sh
source "${SCRIPT_DIR}/lib/lxc-setup.sh"
# shellcheck source=./lib/service-setup.sh
source "${SCRIPT_DIR}/lib/service-setup.sh"
# shellcheck source=./lib/nginx-setup.sh
source "${SCRIPT_DIR}/lib/nginx-setup.sh"
# shellcheck source=./lib/health-check.sh
source "${SCRIPT_DIR}/lib/health-check.sh"

# ============================================================================
# Global Configuration Variables
# ============================================================================

# User-provided configuration
INSTANCE_ID=""
CONTAINER_ID=""
DATABASE_TYPE="sqlite"
DATABASE_HOST=""
DATABASE_PORT=""
DATABASE_NAME="vikunja"
DATABASE_USER=""
DATABASE_PASS=""
DOMAIN=""
IP_ADDRESS=""
GATEWAY=""
ADMIN_EMAIL=""
USE_HTTPS="false"
CPU_CORES="2"
MEMORY_MB="4096"
DISK_GB="20"

# Root access configuration
ROOT_PASSWORD=""
ROOT_SSH_KEY_PATH=""
ENABLE_ROOT_PASSWORD_AUTH="auto"  # auto, true, false

# Derived configuration
BRIDGE="vmbr0"
# Auto-detect available Debian 12 template (will be set in pre-flight checks)
TEMPLATE=""
REPO_URL="https://github.com/aroige/vikunja.git"
REPO_BRANCH="main"
WORKING_DIR="/opt/vikunja"

# Ports (blue-green deployment)
BACKEND_PORT_BLUE=3456
BACKEND_PORT_GREEN=3457
MCP_PORT_BLUE=8456
MCP_PORT_GREEN=8457
ACTIVE_COLOR="blue"

# Flags
NON_INTERACTIVE=false
DEBUG=0

# ============================================================================
# Utility Functions
# ============================================================================

# Display help message (T035)
show_help() {
    cat <<EOF
Vikunja Proxmox LXC Deployment Script v${VERSION}

Usage: $0 [OPTIONS]

OPTIONS:
    -h, --help              Show this help message
    -v, --version           Show version information
    -n, --non-interactive   Non-interactive mode (requires all options)
    -d, --debug             Enable debug output
    
CONFIGURATION OPTIONS:
    --instance-id ID        Instance identifier (default: vikunja-main)
    --container-id ID       LXC container ID (100-999, default: auto)
    --database TYPE         Database type: sqlite|postgres|mysql (default: sqlite)
    --db-host HOST          Database host (required for postgres/mysql)
    --db-port PORT          Database port (default: 5432 for pg, 3306 for mysql)
    --db-name NAME          Database name (default: vikunja)
    --db-user USER          Database username
    --db-pass PASS          Database password
    --domain DOMAIN         Domain name (e.g., vikunja.example.com)
    --https                 Use HTTPS URLs (for external reverse proxy setups)
    --ip-address IP/CIDR    IP address with CIDR (e.g., 192.168.1.100/24)
    --gateway IP            Gateway IP address
    --email EMAIL           Administrator email
    --cpu-cores N           CPU cores (default: 2)
    --memory-mb MB          Memory in MB (default: 4096)
    --disk-gb GB            Disk size in GB (default: 20)

ROOT ACCESS OPTIONS:
    --root-password PASS    Root password (default: secure random generated)
    --root-ssh-key FILE     Path to SSH public key file for root access
    --enable-root-password  Enable SSH password authentication (default: auto)
    --disable-root-password Disable SSH password authentication (key-only)

EXAMPLES:
    # Interactive installation (recommended)
    $0
    
    # Non-interactive installation with SQLite
    $0 --non-interactive --instance-id production \\
       --domain vikunja.example.com --ip-address 192.168.1.100/24 \\
       --gateway 192.168.1.1 --email admin@example.com
    
    # Non-interactive installation with PostgreSQL
    $0 --non-interactive --instance-id prod --database postgres \\
       --db-host 192.168.1.50 --db-port 5432 --db-name vikunja \\
       --db-user vikunja --db-pass secret --domain vikunja.example.com \\
       --ip-address 192.168.1.100/24 --gateway 192.168.1.1 \\
       --email admin@example.com
    
    # Non-interactive installation with SSH key authentication
    $0 --non-interactive --instance-id prod --domain vikunja.example.com \\
       --ip-address 192.168.1.100/24 --gateway 192.168.1.1 \\
       --root-ssh-key ~/.ssh/id_ed25519.pub --disable-root-password

For more information, see: https://vikunja.io/docs
EOF
}

# Display version information
show_version() {
    echo "Vikunja Proxmox LXC Deployment Script"
    echo "Version: ${VERSION}"
    echo "Copyright (c) 2025 Vikunja Contributors"
}

# Parse command line arguments (T035)
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -v|--version)
                show_version
                exit 0
                ;;
            -n|--non-interactive)
                NON_INTERACTIVE=true
                shift
                ;;
            -d|--debug)
                DEBUG=1
                shift
                ;;
            --instance-id)
                INSTANCE_ID="$2"
                shift 2
                ;;
            --container-id)
                CONTAINER_ID="$2"
                shift 2
                ;;
            --database)
                DATABASE_TYPE="$2"
                shift 2
                ;;
            --db-host)
                DATABASE_HOST="$2"
                shift 2
                ;;
            --db-port)
                DATABASE_PORT="$2"
                shift 2
                ;;
            --db-name)
                DATABASE_NAME="$2"
                shift 2
                ;;
            --db-user)
                DATABASE_USER="$2"
                shift 2
                ;;
            --db-pass)
                DATABASE_PASS="$2"
                shift 2
                ;;
            --domain)
                DOMAIN="$2"
                shift 2
                ;;
            --ip-address)
                IP_ADDRESS="$2"
                shift 2
                ;;
            --gateway)
                GATEWAY="$2"
                shift 2
                ;;
            --email)
                ADMIN_EMAIL="$2"
                shift 2
                ;;
            --https)
                USE_HTTPS="true"
                shift
                ;;
            --cpu-cores)
                CPU_CORES="$2"
                shift 2
                ;;
            --memory-mb)
                MEMORY_MB="$2"
                shift 2
                ;;
            --disk-gb)
                DISK_GB="$2"
                shift 2
                ;;
            --root-password)
                ROOT_PASSWORD="$2"
                shift 2
                ;;
            --root-ssh-key)
                ROOT_SSH_KEY_PATH="$2"
                shift 2
                ;;
            --enable-root-password)
                ENABLE_ROOT_PASSWORD_AUTH="true"
                shift
                ;;
            --disable-root-password)
                ENABLE_ROOT_PASSWORD_AUTH="false"
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                echo "Run '$0 --help' for usage information."
                exit 1
                ;;
        esac
    done
}

# ============================================================================
# Interactive Prompts (T036)
# ============================================================================

# Prompt for configuration in interactive mode
prompt_configuration() {
    if [[ "$NON_INTERACTIVE" == "true" ]]; then
        return 0
    fi
    
    echo ""
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║     Vikunja Proxmox LXC Deployment Configuration          ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo ""
    
    # Instance ID
    read -p "Instance ID [vikunja-main]: " input
    INSTANCE_ID="${input:-vikunja-main}"
    
    # Container ID
    local suggested_id
    suggested_id=$(get_next_container_id 100) || suggested_id="100"
    read -p "Container ID (100-999) [${suggested_id}]: " input
    CONTAINER_ID="${input:-${suggested_id}}"
    
    # Database type
    echo ""
    echo "Database Options:"
    echo "  1) SQLite (recommended for small teams)"
    echo "  2) PostgreSQL (recommended for production)"
    echo "  3) MySQL"
    read -p "Select database type [1]: " db_choice
    case "$db_choice" in
        1) DATABASE_TYPE="sqlite" ;;
        2) DATABASE_TYPE="postgres" ;;
        3) DATABASE_TYPE="mysql" ;;
        *) DATABASE_TYPE="sqlite" ;;
    esac
    
    # Database connection details (if not SQLite)
    if [[ "$DATABASE_TYPE" != "sqlite" ]]; then
        read -p "Database host: " DATABASE_HOST
        
        if [[ "$DATABASE_TYPE" == "postgres" ]]; then
            read -p "Database port [5432]: " input
            DATABASE_PORT="${input:-5432}"
        else
            read -p "Database port [3306]: " input
            DATABASE_PORT="${input:-3306}"
        fi
        
        read -p "Database name [vikunja]: " input
        DATABASE_NAME="${input:-vikunja}"
        
        read -p "Database user: " DATABASE_USER
        read -sp "Database password: " DATABASE_PASS
        echo ""
    fi
    
    # Network configuration
    echo ""
    read -p "Domain name (e.g., vikunja.example.com): " DOMAIN
    
    # Ask about HTTPS if domain is provided
    if [[ -n "$DOMAIN" ]]; then
        echo ""
        echo "HTTPS Configuration:"
        echo "  If you're using an external reverse proxy with SSL/TLS (e.g., Caddy, Traefik),"
        echo "  or will configure Let's Encrypt, answer 'yes' here."
        read -p "Will you access Vikunja via HTTPS? [y/N]: " use_https
        if [[ "$use_https" =~ ^[Yy] ]]; then
            USE_HTTPS="true"
        else
            USE_HTTPS="false"
        fi
    fi
    
    read -p "IP address with CIDR (e.g., 192.168.1.100/24): " IP_ADDRESS
    read -p "Gateway IP: " GATEWAY
    
    # Resources
    echo ""
    read -p "CPU cores [2]: " input
    CPU_CORES="${input:-2}"
    read -p "Memory (MB) [4096]: " input
    MEMORY_MB="${input:-4096}"
    read -p "Disk size (GB) [20]: " input
    DISK_GB="${input:-20}"
    
    # Admin email
    echo ""
    read -p "Administrator email: " ADMIN_EMAIL
    
    # Root access configuration
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Root Access Configuration"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "Configure root access to the LXC container:"
    echo "  1) Password only (less secure, convenient for testing)"
    echo "  2) SSH key only (recommended for production)"
    echo "  3) Both password and SSH key (flexible, moderate security)"
    echo "  4) Auto-generated password, no SSH (can access via 'pct enter')"
    echo ""
    read -p "Select root access method [2]: " root_access_choice
    
    case "$root_access_choice" in
        1)
            # Password only
            read -sp "Enter root password: " ROOT_PASSWORD
            echo ""
            read -sp "Confirm root password: " root_password_confirm
            echo ""
            if [[ "$ROOT_PASSWORD" != "$root_password_confirm" ]]; then
                log_error "Passwords do not match"
                exit 1
            fi
            ENABLE_ROOT_PASSWORD_AUTH="true"
            ;;
        2|"")
            # SSH key only (default)
            read -p "Path to SSH public key [~/.ssh/id_ed25519.pub]: " input
            ROOT_SSH_KEY_PATH="${input:-$HOME/.ssh/id_ed25519.pub}"
            # Expand tilde
            ROOT_SSH_KEY_PATH="${ROOT_SSH_KEY_PATH/#\~/$HOME}"
            ENABLE_ROOT_PASSWORD_AUTH="false"
            ;;
        3)
            # Both password and SSH key
            read -sp "Enter root password: " ROOT_PASSWORD
            echo ""
            read -sp "Confirm root password: " root_password_confirm
            echo ""
            if [[ "$ROOT_PASSWORD" != "$root_password_confirm" ]]; then
                log_error "Passwords do not match"
                exit 1
            fi
            read -p "Path to SSH public key [~/.ssh/id_ed25519.pub]: " input
            ROOT_SSH_KEY_PATH="${input:-$HOME/.ssh/id_ed25519.pub}"
            ROOT_SSH_KEY_PATH="${ROOT_SSH_KEY_PATH/#\~/$HOME}"
            ENABLE_ROOT_PASSWORD_AUTH="true"
            ;;
        4)
            # Auto-generated password, no SSH
            ROOT_PASSWORD=""  # Will be auto-generated
            ENABLE_ROOT_PASSWORD_AUTH="false"
            log_info "Root password will be auto-generated and displayed after installation"
            ;;
        *)
            log_error "Invalid selection"
            exit 1
            ;;
    esac
    
    echo ""
}

# Display configuration summary and confirm
confirm_configuration() {
    echo ""
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║         Deployment Configuration Summary                  ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo ""
    echo "Instance ID:    ${INSTANCE_ID}"
    echo "Container ID:   ${CONTAINER_ID}"
    echo "Database:       ${DATABASE_TYPE}"
    if [[ "$DATABASE_TYPE" != "sqlite" ]]; then
        echo "DB Host:        ${DATABASE_HOST}:${DATABASE_PORT}"
        echo "DB Name:        ${DATABASE_NAME}"
        echo "DB User:        ${DATABASE_USER}"
    fi
    echo "Domain:         ${DOMAIN}"
    echo "IP Address:     ${IP_ADDRESS}"
    echo "Gateway:        ${GATEWAY}"
    echo "Resources:      ${CPU_CORES} CPU, ${MEMORY_MB}MB RAM, ${DISK_GB}GB disk"
    echo "Admin Email:    ${ADMIN_EMAIL}"
    echo ""
    
    if [[ "$NON_INTERACTIVE" != "true" ]]; then
        read -p "Confirm configuration? [y/N]: " confirm
        if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
            log_error "Deployment cancelled by user"
            exit 0
        fi
    fi
}

# ============================================================================
# Configuration Validation (T037)
# ============================================================================

# Validate all configuration
validate_configuration() {
    log_info "Validating configuration"
    
    local validation_failed=false
    
    # Instance ID validation
    if [[ ! "$INSTANCE_ID" =~ ^[a-zA-Z0-9-]+$ ]]; then
        log_error "Invalid instance ID: must be alphanumeric with hyphens"
        validation_failed=true
    fi
    
    # Container ID validation
    if ! [[ "$CONTAINER_ID" =~ ^[0-9]+$ ]] || (( CONTAINER_ID < 100 || CONTAINER_ID > 999 )); then
        log_error "Invalid container ID: must be between 100-999"
        validation_failed=true
    fi
    
    # Check if container ID is available
    if pct_exists "$CONTAINER_ID"; then
        log_error "Container ID ${CONTAINER_ID} already exists"
        validation_failed=true
    fi
    
    # Database validation
    case "$DATABASE_TYPE" in
        sqlite)
            # SQLite requires no external connection
            ;;
        postgres|mysql)
            if [[ -z "$DATABASE_HOST" ]]; then
                log_error "Database host is required for ${DATABASE_TYPE}"
                validation_failed=true
            fi
            if [[ -z "$DATABASE_USER" ]]; then
                log_error "Database user is required for ${DATABASE_TYPE}"
                validation_failed=true
            fi
            if [[ -z "$DATABASE_PASS" ]]; then
                log_error "Database password is required for ${DATABASE_TYPE}"
                validation_failed=true
            fi
            # Set default port if not provided
            if [[ -z "$DATABASE_PORT" ]]; then
                if [[ "$DATABASE_TYPE" == "postgres" ]]; then
                    DATABASE_PORT=5432
                    log_info "Using default PostgreSQL port: ${DATABASE_PORT}"
                else
                    DATABASE_PORT=3306
                    log_info "Using default MySQL port: ${DATABASE_PORT}"
                fi
            fi
            if ! validate_port "$DATABASE_PORT"; then
                log_error "Invalid database port: ${DATABASE_PORT}"
                validation_failed=true
            fi
            ;;
        *)
            log_error "Invalid database type: ${DATABASE_TYPE}"
            validation_failed=true
            ;;
    esac
    
    # Network validation
    if ! validate_domain "$DOMAIN"; then
        log_error "Invalid domain: ${DOMAIN}"
        validation_failed=true
    fi
    
    local ip_only="${IP_ADDRESS%/*}"
    if ! validate_ip "$ip_only"; then
        log_error "Invalid IP address: ${IP_ADDRESS}"
        validation_failed=true
    fi
    
    if ! validate_ip "$GATEWAY"; then
        log_error "Invalid gateway IP: ${GATEWAY}"
        validation_failed=true
    fi
    
    # Email validation
    if ! validate_email "$ADMIN_EMAIL"; then
        log_error "Invalid email address: ${ADMIN_EMAIL}"
        validation_failed=true
    fi
    
    # Resource validation
    if (( CPU_CORES < 1 || CPU_CORES > 32 )); then
        log_error "Invalid CPU cores: must be 1-32"
        validation_failed=true
    fi
    
    if (( MEMORY_MB < 2048 )); then
        log_error "Insufficient memory: minimum 2048MB required"
        validation_failed=true
    fi
    
    if (( DISK_GB < 20 )); then
        log_error "Insufficient disk space: minimum 20GB required"
        validation_failed=true
    fi
    
    if [[ "$validation_failed" == "true" ]]; then
        error "Configuration validation failed" 1
    fi
    
    log_success "Configuration validated"
}

# ============================================================================
# Pre-Flight Checks (T038)
# ============================================================================

# Perform pre-flight checks before deployment
preflight_checks() {
    log_info "Running pre-flight checks"
    
    # Check if running as root
    if ! check_root; then
        error "This script must be run as root" 2
    fi
    
    # Check if running on Proxmox
    if ! check_proxmox; then
        error "This script must be run on a Proxmox VE host" 3
    fi
    
    # Detect available Debian 12 template
    log_debug "Detecting available Debian 12 template..."
    local template_file
    template_file=$(ls -1 /var/lib/vz/template/cache/debian-12-standard*.tar.* 2>/dev/null | head -1)
    
    if [[ -z "$template_file" ]]; then
        log_error "No Debian 12 template found in /var/lib/vz/template/cache/"
        log_error ""
        log_error "Download a Debian 12 template with:"
        log_error "  pveam update"
        log_error "  pveam available | grep debian-12"
        log_error "  pveam download local debian-12-standard_12.7-1_amd64.tar.zst"
        error "Debian 12 template not found" 3
    fi
    
    # Convert full path to Proxmox storage format
    local template_name
    template_name=$(basename "$template_file")
    TEMPLATE="local:vztmpl/${template_name}"
    log_debug "Using template: ${TEMPLATE}"
    
    # Check if lock exists
    if check_lock "$INSTANCE_ID"; then
        error "Another deployment is in progress for ${INSTANCE_ID}" 4
    fi
    
    # Check resource availability
    if ! check_resources_available "$CPU_CORES" "$MEMORY_MB" "$DISK_GB"; then
        log_warn "Resource availability check could not be completed"
    fi
    
    # Check if required ports are available on host
    for port in $BACKEND_PORT_BLUE $BACKEND_PORT_GREEN $MCP_PORT_BLUE $MCP_PORT_GREEN; do
        if port_in_use "$port"; then
            log_warn "Port ${port} is already in use on host"
        fi
    done
    
    log_success "Pre-flight checks passed"
}

# ============================================================================
# Deployment Orchestration (T039)
# ============================================================================

# Main deployment function
deploy_vikunja() {
    local start_time
    start_time=$(date +%s)
    
    log_info "Starting Vikunja deployment for instance '${INSTANCE_ID}'"
    
    # Acquire lock
    if ! acquire_lock "$INSTANCE_ID"; then
        error "Failed to acquire deployment lock" 5
    fi
    
    # Step 1: Create container
    progress_start "[1/10] Creating LXC container..."
    if ! create_container "$CONTAINER_ID" "$TEMPLATE" "$CPU_CORES" "$MEMORY_MB" "$DISK_GB"; then
        progress_fail "Failed to create container"
        return 1
    fi
    progress_complete "[1/10] LXC container created"
    
    # Step 2: Configure network
    progress_start "[2/10] Configuring network..."
    if ! configure_network "$CONTAINER_ID" "$BRIDGE" "$IP_ADDRESS" "$GATEWAY"; then
        progress_fail "Failed to configure network"
        return 1
    fi
    progress_complete "[2/10] Network configured"
    
    # Step 3: Start container
    progress_start "[3/10] Starting container..."
    if ! pct_start "$CONTAINER_ID"; then
        progress_fail "Failed to start container"
        return 1
    fi
    if ! wait_for_container_network "$CONTAINER_ID" 60; then
        progress_fail "Container network not ready"
        return 1
    fi
    progress_complete "[3/10] Container started"
    
    # Step 3.5: Configure root access
    progress_start "[3.5/10] Configuring root access..."
    local password_auth_setting="$ENABLE_ROOT_PASSWORD_AUTH"
    # Handle "auto" setting
    if [[ "$password_auth_setting" == "auto" ]]; then
        if [[ -n "$ROOT_SSH_KEY_PATH" ]]; then
            password_auth_setting="false"
        else
            password_auth_setting="true"
        fi
    fi
    
    if ! setup_ssh_access "$CONTAINER_ID" "$ROOT_PASSWORD" "$ROOT_SSH_KEY_PATH" "$password_auth_setting"; then
        progress_fail "Failed to configure root access"
        log_warn "Container is running but root access configuration failed"
        log_warn "You can still access the container with: pct enter ${CONTAINER_ID}"
        return 1
    fi
    progress_complete "[3.5/10] Root access configured"
    
    # Step 4: Install dependencies
    progress_start "[4/10] Installing system dependencies..."
    if ! install_dependencies "$CONTAINER_ID"; then
        progress_fail "Failed to install dependencies"
        return 1
    fi
    progress_complete "[4/10] Dependencies installed"
    
    # Step 5: Setup Go runtime
    progress_start "[5/10] Installing Go runtime..."
    if ! setup_go "$CONTAINER_ID"; then
        progress_fail "Failed to install Go"
        return 1
    fi
    progress_complete "[5/10] Go runtime installed"
    
    # Step 6: Setup Node.js runtime
    progress_start "[6/10] Installing Node.js runtime..."
    # Install Node.js 22 (required for Vite 7.x frontend build)
    if ! setup_nodejs "$CONTAINER_ID" 22; then
        progress_fail "Failed to install Node.js"
        return 1
    fi
    progress_complete "[6/10] Node.js runtime installed"
    
    # Step 7: Clone repository
    progress_start "[7/10] Cloning Vikunja repository..."
    local commit
    if ! commit=$(clone_repository "$CONTAINER_ID" "$REPO_URL" "$REPO_BRANCH" "$WORKING_DIR"); then
        progress_fail "Failed to clone repository"
        return 1
    fi
    progress_complete "[7/10] Repository cloned (commit: ${commit})"
    
    # Step 8: Setup database
    progress_start "[8/10] Setting up database..."
    case "$DATABASE_TYPE" in
        sqlite)
            if ! setup_sqlite "$CONTAINER_ID" "${WORKING_DIR}/vikunja.db"; then
                progress_fail "Failed to setup SQLite"
                return 1
            fi
            ;;
        postgres)
            if ! setup_postgresql "$CONTAINER_ID" "$DATABASE_HOST" "$DATABASE_PORT" \
                "$DATABASE_NAME" "$DATABASE_USER" "$DATABASE_PASS"; then
                progress_fail "Failed to setup PostgreSQL"
                return 1
            fi
            ;;
        mysql)
            if ! setup_mysql "$CONTAINER_ID" "$DATABASE_HOST" "$DATABASE_PORT" \
                "$DATABASE_NAME" "$DATABASE_USER" "$DATABASE_PASS"; then
                progress_fail "Failed to setup MySQL"
                return 1
            fi
            ;;
    esac
    progress_complete "[8/10] Database configured"
    
    # Step 9: Build applications
    progress_start "[9/10] Building Vikunja (this may take several minutes)..."
    
    if ! build_backend "$CONTAINER_ID" "$WORKING_DIR"; then
        progress_fail "Failed to build backend"
        return 1
    fi
    
    if ! build_frontend "$CONTAINER_ID" "$WORKING_DIR"; then
        progress_fail "Failed to build frontend"
        return 1
    fi
    
    if ! build_mcp "$CONTAINER_ID" "$WORKING_DIR"; then
        progress_fail "Failed to build MCP server"
        return 1
    fi
    
    progress_complete "[9/10] Build completed"
    
    # Step 10: Configure services
    progress_start "[10/10] Configuring services..."
    
    # Construct frontend URL from IP or domain
    local frontend_url
    local protocol="http"
    if [[ "$USE_HTTPS" == "true" ]]; then
        protocol="https"
    fi
    
    if [[ -n "$DOMAIN" ]]; then
        frontend_url="${protocol}://${DOMAIN}"
    else
        # Extract IP without CIDR notation (IPs are always http)
        frontend_url="http://${IP_ADDRESS%/*}"
    fi
    
    # Generate and start backend service (blue)
    if ! generate_systemd_unit "$CONTAINER_ID" "backend" "blue" \
        "$BACKEND_PORT_BLUE" "$WORKING_DIR" "$frontend_url" \
        "$DATABASE_TYPE" "$DATABASE_HOST" "$DATABASE_PORT" \
        "$DATABASE_NAME" "$DATABASE_USER" "$DATABASE_PASS"; then
        progress_fail "Failed to generate backend service"
        return 1
    fi
    
    if ! enable_service "$CONTAINER_ID" "vikunja-backend-blue"; then
        progress_fail "Failed to enable backend service"
        return 1
    fi
    
    if ! start_service "$CONTAINER_ID" "vikunja-backend-blue"; then
        progress_fail "Failed to start backend service"
        return 1
    fi
    
    # Generate and start MCP service (blue)
    # Note: MCP service doesn't need database config, pass empty values
    if ! generate_systemd_unit "$CONTAINER_ID" "mcp" "blue" \
        "$MCP_PORT_BLUE" "$WORKING_DIR" "$frontend_url" \
        "" "" "" "" "" ""; then
        progress_fail "Failed to generate MCP service"
        return 1
    fi
    
    if ! enable_service "$CONTAINER_ID" "vikunja-mcp-blue"; then
        progress_fail "Failed to enable MCP service"
        return 1
    fi
    
    if ! start_service "$CONTAINER_ID" "vikunja-mcp-blue"; then
        progress_fail "Failed to start MCP service"
        return 1
    fi
    
    # Configure nginx
    if ! generate_nginx_config "$CONTAINER_ID" "$DOMAIN" "$BACKEND_PORT_BLUE" \
        "${WORKING_DIR}/frontend/dist" "" "" "$IP_ADDRESS"; then
        progress_fail "Failed to generate nginx config"
        return 1
    fi
    
    if ! enable_site "$CONTAINER_ID"; then
        progress_fail "Failed to enable nginx site"
        return 1
    fi
    
    if ! reload_nginx "$CONTAINER_ID"; then
        progress_fail "Failed to reload nginx"
        return 1
    fi
    
    progress_complete "[10/10] Services configured"
    
    # Step 11: Install management scripts into container for future updates
    log_info "Installing management scripts into container..."
    
    # Create deployment directory in container
    pct_exec "$CONTAINER_ID" bash -c "mkdir -p /opt/vikunja-deploy/lib /opt/vikunja-deploy/templates"
    
    # Copy all scripts from bootstrap temp directory into container
    pct push "$CONTAINER_ID" "${SCRIPT_DIR}/vikunja-update.sh" "/opt/vikunja-deploy/vikunja-update.sh"
    pct push "$CONTAINER_ID" "${SCRIPT_DIR}/lib/common.sh" "/opt/vikunja-deploy/lib/common.sh"
    pct push "$CONTAINER_ID" "${SCRIPT_DIR}/lib/proxmox-api.sh" "/opt/vikunja-deploy/lib/proxmox-api.sh"
    pct push "$CONTAINER_ID" "${SCRIPT_DIR}/lib/lxc-setup.sh" "/opt/vikunja-deploy/lib/lxc-setup.sh"
    pct push "$CONTAINER_ID" "${SCRIPT_DIR}/lib/service-setup.sh" "/opt/vikunja-deploy/lib/service-setup.sh"
    pct push "$CONTAINER_ID" "${SCRIPT_DIR}/lib/nginx-setup.sh" "/opt/vikunja-deploy/lib/nginx-setup.sh"
    pct push "$CONTAINER_ID" "${SCRIPT_DIR}/lib/health-check.sh" "/opt/vikunja-deploy/lib/health-check.sh"
    pct push "$CONTAINER_ID" "${SCRIPT_DIR}/lib/blue-green.sh" "/opt/vikunja-deploy/lib/blue-green.sh"
    pct push "$CONTAINER_ID" "${SCRIPT_DIR}/lib/backup-restore.sh" "/opt/vikunja-deploy/lib/backup-restore.sh"
    
    # Copy templates (suppress errors if files don't match pattern)
    shopt -s nullglob
    for template in "${SCRIPT_DIR}"/templates/*.service "${SCRIPT_DIR}"/templates/*.conf "${SCRIPT_DIR}"/templates/*.sh "${SCRIPT_DIR}"/templates/*.yaml; do
        if [[ -f "$template" ]]; then
            pct push "$CONTAINER_ID" "$template" "/opt/vikunja-deploy/templates/$(basename "$template")"
        fi
    done
    shopt -u nullglob
    
    # Make scripts executable in container
    pct_exec "$CONTAINER_ID" bash -c "chmod +x /opt/vikunja-deploy/vikunja-update.sh /opt/vikunja-deploy/lib/*.sh"
    
    # Create convenience symlink
    pct_exec "$CONTAINER_ID" bash -c "ln -sf /opt/vikunja-deploy/vikunja-update.sh /usr/local/bin/vikunja-update"
    
    log_success "Management scripts installed in container"
    
    # Wait for services to become healthy
    log_info "Waiting for services to become healthy..."
    if ! wait_for_healthy "$CONTAINER_ID" "backend" "$BACKEND_PORT_BLUE" 60; then
        log_error "Backend service did not become healthy"
        return 1
    fi
    
    if ! wait_for_healthy "$CONTAINER_ID" "frontend" "" 30; then
        log_error "Frontend did not become accessible"
        return 1
    fi
    
    # Save configuration and state
    save_deployment_state "$commit"
    
    local end_time
    end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    log_success "Deployment completed in ${duration}s"
    
    return 0
}

# Save deployment configuration and state
save_deployment_state() {
    local commit="$1"
    
    log_debug "Saving deployment state"
    
    # Save configuration inside the container (for portability)
    local config_yaml="/etc/vikunja/deploy-config.yaml"
    
    # Create config directory in container
    pct_exec "$CONTAINER_ID" bash -c "mkdir -p /etc/vikunja"
    
    # Generate YAML configuration
    local config_content="# Vikunja Deployment Configuration
# Auto-generated during installation
# Do not edit manually

deployment:
  instance_id: \"${INSTANCE_ID}\"
  container_id: \"${CONTAINER_ID}\"
  deployed_at: \"$(date -Iseconds)\"
  deployed_version: \"${commit}\"

database:
  type: \"${DATABASE_TYPE}\"
  host: \"${DATABASE_HOST}\"
  port: \"${DATABASE_PORT}\"
  name: \"${DATABASE_NAME}\"
  user: \"${DATABASE_USER}\"

network:
  domain: \"${DOMAIN}\"
  ip_address: \"${IP_ADDRESS}\"
  use_https: ${USE_HTTPS}

resources:
  cpu_cores: \"${CPU_CORES}\"
  memory_mb: \"${MEMORY_MB}\"
  disk_gb: \"${DISK_GB}\"

services:
  active_color: \"${ACTIVE_COLOR}\"
  backend_port_blue: ${BACKEND_PORT_BLUE}
  backend_port_green: ${BACKEND_PORT_GREEN}
  mcp_port_blue: ${MCP_PORT_BLUE}
  mcp_port_green: ${MCP_PORT_GREEN}

paths:
  working_dir: \"${WORKING_DIR}\"
  repo_url: \"${REPO_URL}\"
  repo_branch: \"${REPO_BRANCH}\"
"
    
    # Write config to container using heredoc
    pct_exec "$CONTAINER_ID" bash -c "cat > $config_yaml" <<< "$config_content"
    
    log_debug "Configuration saved to container: $config_yaml"
    
    # Also save minimal state on host for reference (optional)
    set_state "$INSTANCE_ID" "container_id" "$CONTAINER_ID"
    set_state "$INSTANCE_ID" "status" "running"
    set_state "$INSTANCE_ID" "deployed_version" "$commit"
}

# ============================================================================
# Post-Deployment Summary (T040)
# ============================================================================

# Display post-deployment summary
show_deployment_summary() {
    local commit="$1"
    local duration="$2"
    
    echo ""
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║          Vikunja Deployment Successful!                   ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo ""
    echo "Instance ID:       ${INSTANCE_ID}"
    echo "Container ID:      ${CONTAINER_ID}"
    echo "IP Address:        ${IP_ADDRESS%/*}"
    echo "Domain:            ${DOMAIN}"
    echo "Version:           ${commit}"
    echo "Database:          ${DATABASE_TYPE}"
    echo ""
    echo "Access URL:        http://${DOMAIN}"
    echo "Deployment Time:   ${duration}s"
    echo ""
    
    # Root access section
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║               Container Root Access                        ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo ""
    
    # Determine what access methods are configured
    local has_ssh_key=false
    local has_password_auth=false
    
    if [[ -n "$ROOT_SSH_KEY_PATH" ]]; then
        has_ssh_key=true
    fi
    
    if [[ "$ENABLE_ROOT_PASSWORD_AUTH" == "true" ]] || [[ -z "$ROOT_SSH_KEY_PATH" && "$ENABLE_ROOT_PASSWORD_AUTH" != "false" ]]; then
        has_password_auth=true
    fi
    
    # Display SSH connection command
    if [[ "$has_ssh_key" == "true" || "$has_password_auth" == "true" ]]; then
        echo "SSH Access:        ssh root@${IP_ADDRESS%/*}"
        
        if [[ "$has_ssh_key" == "true" && "$has_password_auth" == "false" ]]; then
            echo "Authentication:    SSH key only (password auth disabled)"
        elif [[ "$has_ssh_key" == "true" && "$has_password_auth" == "true" ]]; then
            echo "Authentication:    SSH key or password"
        elif [[ "$has_password_auth" == "true" ]]; then
            echo "Authentication:    Password only"
        fi
    fi
    
    # Display root password if it was set/generated
    if [[ -n "${LXC_ROOT_PASSWORD:-}" ]]; then
        echo ""
        echo "⚠️  IMPORTANT - Save this root password securely:"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "Root Password:     ${LXC_ROOT_PASSWORD}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        echo "This password is for console/emergency access."
        if [[ "$has_password_auth" == "false" ]]; then
            echo "SSH password authentication is DISABLED (key-only access)."
        fi
    fi
    
    # Alternative access method
    echo ""
    echo "Console Access:    pct enter ${CONTAINER_ID}"
    echo ""
    
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║                    Next Steps                              ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo ""
    echo "1. Configure DNS to point ${DOMAIN} to ${IP_ADDRESS%/*}"
    echo "2. Access Vikunja at http://${DOMAIN}"
    echo "3. Create your admin account (first user)"
    echo "4. Configure SSL/TLS certificates for production use"
    echo ""
    echo "Management commands:"
    echo "  SSH into container:  ssh root@${IP_ADDRESS%/*}"
    echo "  Or from host:        pct enter ${CONTAINER_ID}"
    echo ""
    echo "  Inside container, run:"
    echo "    vikunja-update              # Update Vikunja and deployment scripts"
    echo "    vikunja-update --help       # See all options"
    echo ""
    echo "  Scripts location:    /opt/vikunja-deploy/"
    echo ""
    echo "For more information: https://vikunja.io/docs"
    echo ""
}

# ============================================================================
# Cleanup and Error Handling (T041, T042)
# ============================================================================

# Cleanup function called on failure
cleanup_on_deployment_failure() {
    local exit_code=$?
    
    if [[ $exit_code -ne 0 ]]; then
        log_error "Deployment failed with exit code ${exit_code}"
        
        # Ask if user wants to cleanup
        if [[ "$NON_INTERACTIVE" != "true" ]]; then
            echo ""
            read -p "Remove failed deployment? [y/N]: " cleanup_confirm
            if [[ "$cleanup_confirm" =~ ^[Yy]$ ]]; then
                log_info "Cleaning up failed deployment..."
                
                # Destroy container if it exists
                if [[ -n "$CONTAINER_ID" ]] && pct_exists "$CONTAINER_ID"; then
                    pct_destroy "$CONTAINER_ID" --purge || log_warn "Failed to destroy container"
                fi
                
                log_info "Cleanup completed"
            fi
        fi
        
        # Release lock
        if [[ -n "$INSTANCE_ID" ]]; then
            release_lock "$INSTANCE_ID"
        fi
    fi
}

# ============================================================================
# Main Execution
# ============================================================================

main() {
    # Set up error trap
    trap cleanup_on_deployment_failure EXIT
    
    # Display banner
    echo ""
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║    Vikunja Proxmox LXC Automated Deployment v${VERSION}        ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo ""
    
    # Parse arguments
    parse_arguments "$@"
    
    # Interactive prompts or use provided arguments
    prompt_configuration
    
    # Validate configuration
    validate_configuration
    
    # Display configuration and confirm
    confirm_configuration
    
    # Pre-flight checks
    preflight_checks
    
    # Execute deployment
    local start_time
    start_time=$(date +%s)
    
    if deploy_vikunja; then
        local end_time
        end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        # Get deployed commit
        local commit
        commit=$(get_commit_hash "$CONTAINER_ID" "$WORKING_DIR")
        
        # Show summary
        show_deployment_summary "$commit" "$duration"
        
        # Release lock
        release_lock "$INSTANCE_ID"
        
        log_success "Installation completed successfully!"
        exit 0
    else
        log_error "Deployment failed"
        exit 6
    fi
}

# Run main function
main "$@"
