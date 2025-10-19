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
CPU_CORES="2"
MEMORY_MB="4096"
DISK_GB="20"

# Derived configuration
BRIDGE="vmbr0"
TEMPLATE="local:vztmpl/debian-12-standard_12.2-1_amd64.tar.zst"
REPO_URL="https://github.com/go-vikunja/vikunja.git"
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
    --database TYPE         Database type: sqlite|postgresql|mysql (default: sqlite)
    --db-host HOST          Database host (required for postgresql/mysql)
    --db-port PORT          Database port (default: 5432 for pg, 3306 for mysql)
    --db-name NAME          Database name (default: vikunja)
    --db-user USER          Database username
    --db-pass PASS          Database password
    --domain DOMAIN         Domain name (e.g., vikunja.example.com)
    --ip-address IP/CIDR    IP address with CIDR (e.g., 192.168.1.100/24)
    --gateway IP            Gateway IP address
    --email EMAIL           Administrator email
    --cpu-cores N           CPU cores (default: 2)
    --memory-mb MB          Memory in MB (default: 4096)
    --disk-gb GB            Disk size in GB (default: 20)

EXAMPLES:
    # Interactive installation (recommended)
    $0
    
    # Non-interactive installation with SQLite
    $0 --non-interactive --instance-id production \\
       --domain vikunja.example.com --ip-address 192.168.1.100/24 \\
       --gateway 192.168.1.1 --email admin@example.com
    
    # Non-interactive installation with PostgreSQL
    $0 --non-interactive --instance-id prod --database postgresql \\
       --db-host 192.168.1.50 --db-port 5432 --db-name vikunja \\
       --db-user vikunja --db-pass secret --domain vikunja.example.com \\
       --ip-address 192.168.1.100/24 --gateway 192.168.1.1 \\
       --email admin@example.com

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
    case "${db_choice:-1}" in
        1) DATABASE_TYPE="sqlite" ;;
        2) DATABASE_TYPE="postgresql" ;;
        3) DATABASE_TYPE="mysql" ;;
        *) DATABASE_TYPE="sqlite" ;;
    esac
    
    # Database connection details (if not SQLite)
    if [[ "$DATABASE_TYPE" != "sqlite" ]]; then
        read -p "Database host: " DATABASE_HOST
        
        if [[ "$DATABASE_TYPE" == "postgresql" ]]; then
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
        postgresql|mysql)
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
    if ! setup_nodejs "$CONTAINER_ID"; then
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
        postgresql)
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
    
    # Generate and start backend service (blue)
    if ! generate_systemd_unit "$CONTAINER_ID" "vikunja-backend-blue" "blue" \
        "$BACKEND_PORT_BLUE" "$WORKING_DIR"; then
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
    if ! generate_systemd_unit "$CONTAINER_ID" "vikunja-mcp-blue" "blue" \
        "$MCP_PORT_BLUE" "$WORKING_DIR"; then
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
        "${WORKING_DIR}/frontend/dist"; then
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
    
    # Update state
    set_state "$INSTANCE_ID" "container_id" "$CONTAINER_ID"
    set_state "$INSTANCE_ID" "status" "running"
    set_state "$INSTANCE_ID" "active_color" "$ACTIVE_COLOR"
    update_deployed_version "$INSTANCE_ID" "$commit"
    
    # Save configuration (would be YAML in production)
    local config_data="instance_id=${INSTANCE_ID}
container_id=${CONTAINER_ID}
database_type=${DATABASE_TYPE}
domain=${DOMAIN}
ip_address=${IP_ADDRESS}
"
    save_config "$INSTANCE_ID" "$config_data"
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
    echo "  ./vikunja-manage.sh status      # Check deployment status"
    echo "  ./vikunja-update.sh             # Update to latest version"
    echo "  ./vikunja-manage.sh backup      # Create backup"
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
