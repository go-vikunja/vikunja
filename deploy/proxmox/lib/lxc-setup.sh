#!/usr/bin/env bash
# Vikunja Proxmox Deployment - LXC Setup and Provisioning Functions
# 
# This library provides comprehensive functions for:
#   - Container creation and configuration (create_container, configure_network)
#   - System provisioning (install_dependencies, setup_go, setup_nodejs)
#   - Database setup (setup_postgresql, setup_mysql, setup_sqlite)
#   - Repository management (clone_repository, check_for_updates, pull_latest)
#   - Build operations (build_backend, build_frontend, build_mcp)
#   - Input validation and error handling
#
# All functions follow consistent patterns:
#   - Parameter validation at start
#   - Detailed logging (debug, info, warning, error, success)
#   - Comprehensive error handling with helpful troubleshooting messages
#   - Graceful handling of idempotent operations
#   - Timeout protection for long-running operations
#
# Required by: vikunja-install.sh, vikunja-update.sh
# Dependencies: lib/common.sh, lib/proxmox-api.sh (sourced by main script)

set -euo pipefail

# Common and proxmox-api functions are sourced by main script before this library

# ============================================================================
# Helper Functions
# ============================================================================

# Safe timeout wrapper - uses timeout if available, otherwise no timeout
# Usage: safe_timeout seconds command [args...]
# Returns: Command exit code
safe_timeout() {
    local timeout_seconds="$1"
    shift
    
    # Check if timeout command is available
    if command -v timeout >/dev/null 2>&1; then
        timeout "$timeout_seconds" "$@"
    else
        log_warning "timeout command not available, running without timeout protection"
        "$@"
    fi
}

# ============================================================================
# Input Validation Functions
# ============================================================================

# Validate container ID format
# Usage: validate_ct_id ct_id
# Returns: 0 if valid, 1 if invalid
validate_ct_id() {
    local ct_id="$1"
    
    if [[ ! "$ct_id" =~ ^[0-9]+$ ]]; then
        log_error "Invalid container ID: ${ct_id} (must be numeric)"
        return 1
    fi
    
    if [[ "$ct_id" -lt 100 ]] || [[ "$ct_id" -gt 999999999 ]]; then
        log_error "Invalid container ID: ${ct_id} (must be between 100 and 999999999)"
        return 1
    fi
    
    return 0
}

# Validate network configuration
# Usage: validate_network_config ip_cidr gateway
# Returns: 0 if valid, 1 if invalid
validate_network_config() {
    local ip_cidr="$1"
    local gateway="$2"
    
    # Basic CIDR validation
    if [[ ! "$ip_cidr" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+/[0-9]+$ ]]; then
        log_error "Invalid IP CIDR format: ${ip_cidr} (expected: x.x.x.x/xx)"
        return 1
    fi
    
    # Basic gateway validation
    if [[ ! "$gateway" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        log_error "Invalid gateway format: ${gateway} (expected: x.x.x.x)"
        return 1
    fi
    
    return 0
}

# ============================================================================
# LXC Container Creation Functions (T024)
# ============================================================================

# Create LXC container with specified configuration
# Usage: create_container ct_id [template] [cores] [memory_mb] [disk_gb]
# Returns: 0 on success, 1 on failure
create_container() {
    local ct_id="$1"
    local template="${2:-${PROXMOX_OSTEMPLATE}}"
    local cores="${3:-${RESOURCES_CORES:-2}}"
    local memory_mb="${4:-${RESOURCES_MEMORY:-2048}}"
    local disk_gb="${5:-${RESOURCES_DISK:-10}}"
    
    # Validate container ID
    if ! validate_ct_id "$ct_id"; then
        return 1
    fi
    
    log_info "Creating LXC container ${ct_id}"
    log_debug "Template: ${template}"
    log_debug "Cores: ${cores}, Memory: ${memory_mb}MB, Disk: ${disk_gb}GB"
    
    # Check if container already exists
    if pct_exists "$ct_id"; then
        log_error "Container ${ct_id} already exists"
        return 1
    fi
    
    # Verify template exists
    if [[ ! "$template" =~ ^(local|local-lvm|local-zfs): ]]; then
        log_debug "Checking if template file exists: ${template}"
        if ! [[ -f "${template}" ]]; then
            log_error "Template not found: ${template}"
            log_error "Available templates:"
            pveam available | grep debian || true
            log_error ""
            log_error "Download template with: pveam download local debian-12-standard_12.2-1_amd64.tar.zst"
            return 1
        fi
    fi
    
    # Detect storage for rootfs
    local storage="local-lvm"
    
    # Check if local-lvm exists, fallback to other storage
    if ! pvesm status | grep -q "^local-lvm"; then
        log_debug "local-lvm not found, detecting alternative storage..."
        
        # Try to find any LVM-thin, ZFS, or directory storage
        storage=$(pvesm status | awk 'NR>1 && ($2=="lvmthin" || $2=="zfspool" || $2=="dir") {print $1; exit}')
        
        if [[ -z "$storage" ]]; then
            log_error "No suitable storage found for container rootfs"
            log_error "Available storage:"
            pvesm status
            return 1
        fi
        
        log_debug "Using storage: ${storage}"
    fi
    
    # Build pct create options
    local -a opts=(
        --hostname "vikunja-${ct_id}"
        --cores "$cores"
        --memory "$memory_mb"
        --rootfs "${storage}:${disk_gb}"
        --unprivileged 1
        --features "nesting=1"
        --onboot 1
        --start 0
    )
    
    log_debug "Creating container with: pct create ${ct_id} ${template} ${opts[*]}"
    
    # Create container with detailed error output
    local create_output
    if ! create_output=$(pct create "$ct_id" "$template" "${opts[@]}" 2>&1); then
        log_error "Failed to create container ${ct_id}"
        log_error "Error output:"
        echo "$create_output" | while IFS= read -r line; do
            log_error "  ${line}"
        done
        
        # Provide helpful troubleshooting
        log_error ""
        log_error "Troubleshooting:"
        log_error "1. Check template exists: ls -lh /var/lib/vz/template/cache/"
        log_error "2. Check storage status: pvesm status"
        log_error "3. Check disk space: df -h"
        log_error "4. Try manual creation: pct create ${ct_id} ${template} ${opts[*]}"
        
        return 1
    fi
    
    log_success "Container ${ct_id} created"
    return 0
}

# Configure container network
# Usage: configure_network ct_id [bridge] [ip_cidr] [gateway]
# Returns: 0 on success, 1 on failure
configure_network() {
    local ct_id="$1"
    local bridge="${2:-${NETWORK_BRIDGE:-vmbr0}}"
    local ip_cidr="${3:-${NETWORK_IP:-192.168.1.100/24}}"
    local gateway="${4:-${NETWORK_GATEWAY:-192.168.1.1}}"
    
    # Validate inputs
    if ! validate_ct_id "$ct_id"; then
        return 1
    fi
    
    if ! validate_network_config "$ip_cidr" "$gateway"; then
        return 1
    fi
    
    log_info "Configuring network for container ${ct_id}"
    
    # Set network configuration
    pct set "$ct_id" \
        --net0 "name=eth0,bridge=${bridge},ip=${ip_cidr},gw=${gateway},firewall=1" \
        --nameserver "8.8.8.8,8.8.4.4" \
        --searchdomain "local" \
        2>&1
    
    log_success "Network configured"
    return 0
}

# Allocate resources to container
# Usage: allocate_resources ct_id cores memory_mb
# Returns: 0 on success, 1 on failure
allocate_resources() {
    local ct_id="$1"
    local cores="$2"
    local memory_mb="$3"
    
    log_info "Allocating resources: ${cores} cores, ${memory_mb}MB RAM"
    
    pct set "$ct_id" \
        --cores "$cores" \
        --memory "$memory_mb" \
        --swap 512 \
        2>&1
    
    log_success "Resources allocated"
    return 0
}

# ============================================================================
# Container Provisioning Functions (T025)
# ============================================================================

# Wait for container to be ready
# Usage: wait_for_container ct_id [timeout_seconds]
# Returns: 0 if ready, 1 if timeout
wait_for_container() {
    local ct_id="$1"
    local timeout="${2:-60}"
    local elapsed=0
    
    log_debug "Waiting for container ${ct_id} to be ready (timeout: ${timeout}s)..."
    
    while [[ $elapsed -lt $timeout ]]; do
        if pct_exec "$ct_id" test -f /bin/bash 2>/dev/null; then
            log_debug "Container is ready"
            return 0
        fi
        sleep 2
        elapsed=$((elapsed + 2))
    done
    
    log_error "Container did not become ready within ${timeout} seconds"
    return 1
}

# Install system dependencies in container
# Usage: install_dependencies ct_id
# Returns: 0 on success, 1 on failure
install_dependencies() {
    local ct_id="$1"
    
    log_info "Installing system dependencies"
    
    # Wait for container to be ready
    if ! wait_for_container "$ct_id" 60; then
        return 1
    fi
    
    # Update package lists with retries
    log_debug "Updating package lists..."
    local retry_count=0
    while [[ $retry_count -lt 3 ]]; do
        if pct_exec "$ct_id" apt-get update 2>&1; then
            break
        fi
        retry_count=$((retry_count + 1))
        log_warning "apt-get update failed (attempt ${retry_count}/3), retrying..."
        sleep 5
    done
    
    if [[ $retry_count -eq 3 ]]; then
        log_error "Failed to update package lists after 3 attempts"
        return 1
    fi
    
    # Install required packages
    # Note: Using default-mysql-client for Debian 12+ compatibility
    log_debug "Installing required packages..."
    local packages=(
        "git" "curl" "wget" "build-essential"
        "ca-certificates" "gnupg" "lsb-release"
        "nginx" "sqlite3" "postgresql-client" "default-mysql-client"
        "sudo" "systemd" "procps"
    )
    
    # Use DEBIAN_FRONTEND=noninteractive to avoid prompts
    if ! pct_exec "$ct_id" bash -c \
        "DEBIAN_FRONTEND=noninteractive apt-get install -y ${packages[*]}" \
        2>&1; then
        log_error "Failed to install system dependencies"
        log_error "You can try manually with: pct enter ${ct_id}"
        log_error "Then run: apt-get install -y ${packages[*]}"
        return 1
    fi
    
    log_success "System dependencies installed"
    return 0
}

# Setup SSH access and root authentication in container
# Usage: setup_ssh_access ct_id [root_password] [ssh_key_path] [enable_password_auth]
# Parameters:
#   ct_id                  - Container ID
#   root_password          - Root password (optional, secure random if empty)
#   ssh_key_path           - Path to SSH public key file (optional)
#   enable_password_auth   - "true" to enable password auth, "false" to disable (default: "false" if key provided, "true" otherwise)
# Returns: 0 on success, 1 on failure
#
# Security Features:
#   - Generates cryptographically secure random password if not provided
#   - Injects SSH public keys into authorized_keys with proper permissions
#   - Configures SSH hardening (disable root password login when key is used)
#   - Supports password-only, key-only, or both authentication methods
#   - Creates .ssh directory with correct ownership and permissions (700)
#   - Sets authorized_keys permissions to 600
#   - Validates SSH key format before injection
#   - Prevents weak authentication configurations
setup_ssh_access() {
    local ct_id="$1"
    local root_password="${2:-}"
    local ssh_key_path="${3:-}"
    local enable_password_auth="${4:-}"
    
    log_info "Configuring SSH access and root authentication"
    
    # Validate container exists and is running
    if ! pct_exists "$ct_id"; then
        log_error "Container ${ct_id} does not exist"
        return 1
    fi
    
    if ! pct status "$ct_id" | grep -q "running"; then
        log_error "Container ${ct_id} is not running"
        return 1
    fi
    
    # Install openssh-server if not already installed
    log_debug "Ensuring openssh-server is installed"
    pct_exec "$ct_id" bash -c "
        if ! dpkg -l | grep -q openssh-server; then
            apt-get update -qq
            apt-get install -y openssh-server
        fi
    " 2>&1 || {
        log_error "Failed to install openssh-server"
        return 1
    }
    
    # Generate secure random password if not provided
    if [[ -z "$root_password" ]]; then
        log_debug "Generating cryptographically secure root password"
        # Generate 32-character alphanumeric password
        root_password=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-32)
    fi
    
    # Always set root password (even if SSH key is used, for console access)
    log_debug "Setting root password for console/emergency access"
    if ! pct_exec "$ct_id" bash -c "echo 'root:${root_password}' | chpasswd" 2>&1; then
        log_error "Failed to set root password"
        return 1
    fi
    
    # Handle SSH key injection if provided
    local ssh_key_configured=false
    if [[ -n "$ssh_key_path" ]]; then
        log_info "Configuring SSH key authentication"
        
        # Validate SSH key file exists and is readable
        if [[ ! -f "$ssh_key_path" ]]; then
            log_error "SSH key file not found: ${ssh_key_path}"
            return 1
        fi
        
        if [[ ! -r "$ssh_key_path" ]]; then
            log_error "SSH key file not readable: ${ssh_key_path}"
            return 1
        fi
        
        # Read and validate SSH key format
        local ssh_key_content
        ssh_key_content=$(cat "$ssh_key_path")
        
        if [[ -z "$ssh_key_content" ]]; then
            log_error "SSH key file is empty: ${ssh_key_path}"
            return 1
        fi
        
        # Basic SSH key format validation (should start with ssh-rsa, ssh-ed25519, ecdsa-sha2-, etc.)
        if ! echo "$ssh_key_content" | grep -qE '^(ssh-rsa|ssh-ed25519|ecdsa-sha2-|sk-ssh-ed25519@openssh\.com|sk-ecdsa-sha2-)'; then
            log_error "Invalid SSH public key format in: ${ssh_key_path}"
            log_error "Key must start with: ssh-rsa, ssh-ed25519, ecdsa-sha2-, etc."
            return 1
        fi
        
        log_debug "SSH key validated, injecting into container"
        
        # Create .ssh directory with proper permissions
        if ! pct_exec "$ct_id" bash -c "
            mkdir -p /root/.ssh
            chmod 700 /root/.ssh
            chown root:root /root/.ssh
        " 2>&1; then
            log_error "Failed to create /root/.ssh directory"
            return 1
        fi
        
        # Inject SSH key into authorized_keys
        # Use printf to avoid issues with special characters
        if ! pct_exec "$ct_id" bash -c "
            printf '%s\n' '${ssh_key_content}' >> /root/.ssh/authorized_keys
            chmod 600 /root/.ssh/authorized_keys
            chown root:root /root/.ssh/authorized_keys
        " 2>&1; then
            log_error "Failed to inject SSH key into authorized_keys"
            return 1
        fi
        
        ssh_key_configured=true
        log_success "SSH key injected successfully"
    fi
    
    # Determine SSH password authentication setting
    # Default behavior: disable password auth if SSH key is configured, enable otherwise
    if [[ -z "$enable_password_auth" ]]; then
        if [[ "$ssh_key_configured" == "true" ]]; then
            enable_password_auth="false"
            log_debug "SSH key configured - password authentication will be disabled (use --root-password force to enable)"
        else
            enable_password_auth="true"
            log_debug "No SSH key configured - password authentication will be enabled"
        fi
    fi
    
    # Configure SSH daemon
    log_debug "Configuring SSH daemon security settings"
    
    local sshd_config_commands=""
    
    # Always permit root login (method will be determined by password auth setting)
    sshd_config_commands+="sed -i 's/^#*PermitRootLogin.*/PermitRootLogin yes/' /etc/ssh/sshd_config; "
    
    if [[ "$enable_password_auth" == "true" ]]; then
        log_info "Enabling SSH password authentication"
        sshd_config_commands+="sed -i 's/^#*PasswordAuthentication.*/PasswordAuthentication yes/' /etc/ssh/sshd_config; "
    else
        log_info "Disabling SSH password authentication (key-only access)"
        sshd_config_commands+="sed -i 's/^#*PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config; "
        sshd_config_commands+="sed -i 's/^#*PubkeyAuthentication.*/PubkeyAuthentication yes/' /etc/ssh/sshd_config; "
    fi
    
    # Additional security hardening
    sshd_config_commands+="sed -i 's/^#*PermitEmptyPasswords.*/PermitEmptyPasswords no/' /etc/ssh/sshd_config; "
    sshd_config_commands+="sed -i 's/^#*X11Forwarding.*/X11Forwarding no/' /etc/ssh/sshd_config; "
    
    # Apply SSH configuration and restart daemon
    if ! pct_exec "$ct_id" bash -c "$sshd_config_commands systemctl restart sshd" 2>&1; then
        log_error "Failed to configure SSH daemon"
        return 1
    fi
    
    # Log summary
    log_success "SSH access configured successfully"
    log_info "Root authentication summary:"
    log_info "  - Console/emergency access: Password enabled"
    
    if [[ "$ssh_key_configured" == "true" ]]; then
        log_info "  - SSH key authentication: Enabled"
    fi
    
    if [[ "$enable_password_auth" == "true" ]]; then
        log_info "  - SSH password authentication: Enabled"
    else
        log_info "  - SSH password authentication: Disabled (key-only)"
    fi
    
    # Store password for later display (caller's responsibility to show securely)
    # Export for use by deployment summary
    export LXC_ROOT_PASSWORD="$root_password"
    
    return 0
}

# Setup Go runtime in container
# Usage: setup_go ct_id [version]
# Returns: 0 on success, 1 on failure
setup_go() {
    local ct_id="$1"
    local go_version="${2:-1.21.5}"
    
    log_info "Installing Go ${go_version}"
    
    # Download and install Go
    local go_url="https://go.dev/dl/go${go_version}.linux-amd64.tar.gz"
    
    pct_exec "$ct_id" bash -c "
        rm -rf /usr/local/go && \
        wget -q ${go_url} -O /tmp/go.tar.gz && \
        tar -C /usr/local -xzf /tmp/go.tar.gz && \
        rm /tmp/go.tar.gz
    " 2>&1 || return 1
    
    # Add Go to PATH
    pct_exec "$ct_id" bash -c "
        echo 'export PATH=\$PATH:/usr/local/go/bin' >> /etc/profile.d/go.sh && \
        echo 'export GOPATH=/root/go' >> /etc/profile.d/go.sh
    " 2>&1 || return 1
    
    # Verify installation
    local go_ver
    go_ver=$(pct_exec "$ct_id" /usr/local/go/bin/go version)
    log_success "Go installed: ${go_ver}"
    
    return 0
}

# Setup Node.js runtime in container
# Usage: setup_nodejs ct_id [version]
# Returns: 0 on success, 1 on failure
setup_nodejs() {
    local ct_id="$1"
    # Default to Node.js 22 to match Vikunja requirements:
    # - Frontend .nvmrc: 22.18.0
    # - Vite 7.1.10 requires: Node.js 20.19+ or 22.12+
    # - Node.js 22 LTS until April 2027 (better than 18, EOL April 2025)
    local node_version="${2:-22}"
    
    log_info "Installing Node.js ${node_version}"
    
    # Install Node.js from NodeSource
    pct_exec "$ct_id" bash -c "
        curl -fsSL https://deb.nodesource.com/setup_${node_version}.x | bash - && \
        apt-get install -y nodejs
    " 2>&1 || return 1
    
    # Install pnpm
    pct_exec "$ct_id" npm install -g pnpm 2>&1 || return 1
    
    # Verify installation
    local node_ver
    node_ver=$(pct_exec "$ct_id" node --version)
    log_success "Node.js installed: ${node_ver}"
    
    return 0
}

# Clone Vikunja repository in container
# Usage: clone_repository ct_id [repo_url] [branch] [target_dir]
# Returns: 0 on success, 1 on failure
clone_repository() {
    local ct_id="$1"
    local repo_url="${2:-${GIT_BACKEND_REPO}}"
    local branch="${3:-${GIT_BACKEND_BRANCH:-main}}"
    local target_dir="${4:-/opt/vikunja}"
    
    log_info "Cloning Vikunja repository (branch: ${branch})"
    
    # Check if directory already exists
    if pct_exec "$ct_id" test -d "$target_dir"; then
        log_warning "Directory ${target_dir} already exists"
        
        # Check if it's a git repository
        if pct_exec "$ct_id" test -d "${target_dir}/.git"; then
            log_debug "Existing git repository found, pulling latest changes..."
            if ! pct_exec "$ct_id" git -C "$target_dir" pull 2>&1; then
                log_error "Failed to update existing repository"
                return 1
            fi
            local commit
            commit=$(pct_exec "$ct_id" git -C "$target_dir" rev-parse --short HEAD)
            log_success "Repository updated (commit: ${commit})"
            echo "$commit"
            return 0
        else
            log_error "Directory exists but is not a git repository"
            log_error "Remove it manually: pct exec ${ct_id} -- rm -rf ${target_dir}"
            return 1
        fi
    fi
    
    # Clone repository with timeout
    log_debug "Cloning from ${repo_url}..."
    
    # Note: We use bash -c because timeout can't call shell functions directly
    local clone_cmd="pct exec ${ct_id} -- git clone --depth 1 --branch ${branch} ${repo_url} ${target_dir}"
    
    if ! safe_timeout 300 bash -c "${clone_cmd}" 2>&1; then
        log_error "Failed to clone repository (timeout or network error)"
        log_error "Check network connectivity in container: pct enter ${ct_id}"
        log_error "Then run: git clone --branch ${branch} ${repo_url} ${target_dir}"
        return 1
    fi
    
    # Get commit hash
    local commit
    commit=$(pct_exec "$ct_id" git -C "$target_dir" rev-parse --short HEAD)
    
    log_success "Repository cloned (commit: ${commit})"
    echo "$commit"
    
    return 0
}

# ============================================================================
# Database Setup Functions (T026)
# ============================================================================

# Setup SQLite database
# Usage: setup_sqlite ct_id [db_path]
# Returns: 0 on success, 1 on failure
setup_sqlite() {
    local ct_id="$1"
    local db_path="${2:-${DB_PATH:-/opt/vikunja/vikunja.db}}"
    
    log_info "Setting up SQLite database"
    
    # Create database directory
    local db_dir
    db_dir=$(dirname "$db_path")
    pct_exec "$ct_id" mkdir -p "$db_dir" || return 1
    
    # Set permissions
    pct_exec "$ct_id" chown -R root:root "$db_dir" || return 1
    pct_exec "$ct_id" chmod 755 "$db_dir" || return 1
    
    log_success "SQLite database path configured: ${db_path}"
    return 0
}

# Setup PostgreSQL connection
# Usage: setup_postgresql ct_id host port dbname user password
# Returns: 0 on success, 1 on failure
setup_postgresql() {
    local ct_id="$1"
    local host="${2:-${DB_HOST:-localhost}}"
    local port="${3:-${DB_PORT:-5432}}"
    local dbname="${4:-${DB_NAME:-vikunja}}"
    local user="${5:-${DB_USER:-vikunja}}"
    local password="${6:-${DB_PASSWORD:-vikunja}}"
    
    log_info "Setting up PostgreSQL connection to ${host}:${port}"
    
    # First, test if we can connect to PostgreSQL server (using postgres database)
    log_debug "Testing PostgreSQL server connectivity..."
    if ! pct_exec "$ct_id" bash -c \
        "PGPASSWORD='${password}' psql -h ${host} -p ${port} -U ${user} -d postgres -c 'SELECT 1'" \
        >/dev/null 2>&1; then
        log_error "Failed to connect to PostgreSQL server at ${host}:${port}"
        log_error "Please check:"
        log_error "  1. PostgreSQL server is running: systemctl status postgresql"
        log_error "  2. Server allows connections from container IP (pg_hba.conf)"
        log_error "  3. Credentials are correct (user: ${user})"
        log_error "  4. Network connectivity: ping ${host}"
        return 1
    fi
    
    log_debug "PostgreSQL server connectivity OK"
    
    # Check if database exists, create if it doesn't
    log_debug "Checking if database '${dbname}' exists..."
    if ! pct_exec "$ct_id" bash -c \
        "PGPASSWORD='${password}' psql -h ${host} -p ${port} -U ${user} -d postgres -lqt | cut -d \\| -f 1 | grep -qw ${dbname}" \
        >/dev/null 2>&1; then
        
        log_info "Database '${dbname}' does not exist, creating..."
        
        local create_output
        create_output=$(pct_exec "$ct_id" bash -c \
            "PGPASSWORD='${password}' psql -h ${host} -p ${port} -U ${user} -d postgres -c 'CREATE DATABASE ${dbname};'" \
            2>&1)
        local create_exit=$?
        
        # Check if creation failed (but ignore "already exists" error)
        if [[ $create_exit -ne 0 ]] && ! echo "$create_output" | grep -qi "already exists"; then
            log_error "Failed to create database '${dbname}'"
            log_error "Error output:"
            echo "$create_output" | while IFS= read -r line; do
                log_error "  $line"
            done
            log_error ""
            log_error "You may need to create it manually:"
            log_error "  psql -U postgres -c 'CREATE DATABASE ${dbname};'"
            log_error "  psql -U postgres -c 'GRANT ALL PRIVILEGES ON DATABASE ${dbname} TO ${user};'"
            return 1
        fi
        
        if echo "$create_output" | grep -qi "already exists"; then
            log_debug "Database '${dbname}' already exists (concurrent creation or race condition)"
        else
            log_success "Database '${dbname}' created"
        fi
    else
        log_debug "Database '${dbname}' already exists"
    fi
    
    # Grant all privileges on database to user
    log_debug "Granting all privileges on database '${dbname}' to user '${user}'..."
    local grant_output
    grant_output=$(pct_exec "$ct_id" bash -c \
        "PGPASSWORD='${password}' psql -h ${host} -p ${port} -U ${user} -d postgres -c 'GRANT ALL PRIVILEGES ON DATABASE ${dbname} TO ${user};'" \
        2>&1)
    local grant_exit=$?
    
    if [[ $grant_exit -ne 0 ]]; then
        log_warning "Failed to grant database privileges (this may be OK if user already has permissions)"
        log_debug "Grant output: $grant_output"
    else
        log_debug "Database privileges granted successfully"
    fi
    
    # Grant schema permissions (required for PostgreSQL 15+)
    log_debug "Granting schema privileges on '${dbname}'..."
    grant_output=$(pct_exec "$ct_id" bash -c \
        "PGPASSWORD='${password}' psql -h ${host} -p ${port} -U ${user} -d ${dbname} -c 'GRANT ALL ON SCHEMA public TO ${user};'" \
        2>&1)
    grant_exit=$?
    
    if [[ $grant_exit -ne 0 ]]; then
        log_warning "Failed to grant schema privileges (this may be OK if user already has permissions)"
        log_debug "Grant output: $grant_output"
    else
        log_debug "Schema privileges granted successfully"
    fi
    
    # Test final connection to the vikunja database
    if ! test_db_connection "$ct_id" "postgresql" "$host" "$port" "$dbname" "$user" "$password"; then
        log_error "Failed to connect to database '${dbname}'"
        return 1
    fi
    
    log_success "PostgreSQL connection configured"
    return 0
}

# Setup MySQL connection
# Usage: setup_mysql ct_id host port dbname user password
# Returns: 0 on success, 1 on failure
setup_mysql() {
    local ct_id="$1"
    local host="${2:-${DB_HOST:-localhost}}"
    local port="${3:-${DB_PORT:-3306}}"
    local dbname="${4:-${DB_NAME:-vikunja}}"
    local user="${5:-${DB_USER:-vikunja}}"
    local password="${6:-${DB_PASSWORD:-vikunja}}"
    
    log_info "Setting up MySQL connection to ${host}:${port}"
    
    # First, test if we can connect to MySQL server (using mysql database)
    log_debug "Testing MySQL server connectivity..."
    if ! pct_exec "$ct_id" bash -c \
        "mysql -h ${host} -P ${port} -u ${user} -p'${password}' -e 'SELECT 1'" \
        >/dev/null 2>&1; then
        log_error "Failed to connect to MySQL server at ${host}:${port}"
        log_error "Please check:"
        log_error "  1. MySQL server is running: systemctl status mysql"
        log_error "  2. Server allows connections from container IP"
        log_error "  3. Credentials are correct (user: ${user})"
        log_error "  4. Network connectivity: ping ${host}"
        return 1
    fi
    
    log_debug "MySQL server connectivity OK"
    
    # Check if database exists, create if it doesn't
    log_debug "Checking if database '${dbname}' exists..."
    if ! pct_exec "$ct_id" bash -c \
        "mysql -h ${host} -P ${port} -u ${user} -p'${password}' -e 'USE ${dbname}'" \
        >/dev/null 2>&1; then
        
        log_info "Database '${dbname}' does not exist, creating..."
        
        local create_output
        create_output=$(pct_exec "$ct_id" bash -c \
            "mysql -h ${host} -P ${port} -u ${user} -p'${password}' -e 'CREATE DATABASE ${dbname} CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;'" \
            2>&1)
        local create_exit=$?
        
        # Check if creation failed (but ignore "database exists" error)
        if [[ $create_exit -ne 0 ]] && ! echo "$create_output" | grep -qi "database exists"; then
            log_error "Failed to create database '${dbname}'"
            log_error "Error output:"
            echo "$create_output" | while IFS= read -r line; do
                log_error "  $line"
            done
            log_error ""
            log_error "You may need to create it manually:"
            log_error "  mysql -u root -p -e \"CREATE DATABASE ${dbname};\""
            log_error "  mysql -u root -p -e \"GRANT ALL PRIVILEGES ON ${dbname}.* TO '${user}'@'%';\""
            return 1
        fi
        
        if echo "$create_output" | grep -qi "database exists"; then
            log_debug "Database '${dbname}' already exists (concurrent creation or race condition)"
        else
            log_success "Database '${dbname}' created"
        fi
    else
        log_debug "Database '${dbname}' already exists"
    fi
    
    # Grant all privileges on database to user
    log_debug "Granting all privileges on database '${dbname}' to user '${user}'..."
    local grant_output
    grant_output=$(pct_exec "$ct_id" bash -c \
        "mysql -h ${host} -P ${port} -u ${user} -p'${password}' -e 'GRANT ALL PRIVILEGES ON ${dbname}.* TO '\''${user}'\''@'\''%'\''; FLUSH PRIVILEGES;'" \
        2>&1)
    local grant_exit=$?
    
    if [[ $grant_exit -ne 0 ]]; then
        log_warning "Failed to grant privileges (this may be OK if user already has permissions)"
        log_debug "Grant output: $grant_output"
    else
        log_debug "Privileges granted successfully"
    fi
    
    # Test final connection to the vikunja database
    if ! test_db_connection "$ct_id" "mysql" "$host" "$port" "$dbname" "$user" "$password"; then
        log_error "Failed to connect to database '${dbname}'"
        return 1
    fi
    
    log_success "MySQL connection configured"
    return 0
}

# Test database connection
# Usage: test_db_connection ct_id type host port dbname user password
# Returns: 0 if successful, 1 if failed
test_db_connection() {
    local ct_id="$1"
    local type="${2:-postgresql}"
    local host="${3:-${DB_HOST:-localhost}}"
    local port="${4:-${DB_PORT:-5432}}"
    local dbname="${5:-${DB_NAME:-vikunja}}"
    local user="${6:-${DB_USER:-vikunja}}"
    local password="${7:-${DB_PASSWORD:-vikunja}}"
    
    log_debug "Testing ${type} connection to ${host}:${port}/${dbname}"
    
    local test_output
    local exit_code
    
    case "$type" in
        postgresql)
            test_output=$(pct_exec "$ct_id" bash -c \
                "PGPASSWORD='${password}' psql -h ${host} -p ${port} -U ${user} -d ${dbname} -c 'SELECT 1'" \
                2>&1)
            exit_code=$?
            ;;
        mysql)
            test_output=$(pct_exec "$ct_id" bash -c \
                "mysql -h ${host} -P ${port} -u ${user} -p'${password}' ${dbname} -e 'SELECT 1'" \
                2>&1)
            exit_code=$?
            ;;
        *)
            log_error "Unknown database type: ${type}"
            return 1
            ;;
    esac
    
    if [[ $exit_code -ne 0 ]]; then
        log_debug "Connection test failed with output:"
        echo "$test_output" | while IFS= read -r line; do
            log_debug "  $line"
        done
        return 1
    fi
    
    return 0
}

# ============================================================================
# Build Functions (T027)
# ============================================================================

# Check if container has sufficient disk space
# Usage: check_disk_space ct_id required_mb [path]
# Returns: 0 if sufficient, 1 if insufficient
check_disk_space() {
    local ct_id="$1"
    local required_mb="$2"
    local check_path="${3:-/}"
    
    log_debug "Checking disk space in container ${ct_id} (path: ${check_path}, required: ${required_mb}MB)"
    
    # Get available space in MB
    local available_mb
    available_mb=$(pct_exec "$ct_id" df -BM "$check_path" | awk 'NR==2 {print $4}' | sed 's/M//')
    
    if [[ "$available_mb" -lt "$required_mb" ]]; then
        log_error "Insufficient disk space in container"
        log_error "  Path: ${check_path}"
        log_error "  Available: ${available_mb}MB"
        log_error "  Required: ${required_mb}MB"
        log_error "  Shortage: $((required_mb - available_mb))MB"
        log_error ""
        log_error "Increase container disk size or free up space"
        return 1
    fi
    
    log_debug "Disk space check passed: ${available_mb}MB available (${required_mb}MB required)"
    return 0
}

# Build Vikunja backend
# Usage: build_backend ct_id source_dir
# Returns: 0 on success, 1 on failure
build_backend() {
    local ct_id="$1"
    local source_dir="$2"
    
    log_info "Building Vikunja backend (this may take several minutes)"
    
    # Verify source directory exists
    if ! pct_exec "$ct_id" test -d "${source_dir}"; then
        log_error "Source directory not found: ${source_dir}"
        return 1
    fi
    
    # Check disk space (need ~2GB for Go build cache + binary)
    if ! check_disk_space "$ct_id" 2048 "$source_dir"; then
        return 1
    fi
    
    # Install mage and build
    log_debug "Installing mage build tool..."
    if ! pct_exec "$ct_id" bash -c "
        export PATH=\$PATH:/usr/local/go/bin
        export GOPATH=/root/go
        go install github.com/magefile/mage@latest
    " 2>&1; then
        log_error "Failed to install mage"
        return 1
    fi
    
    # Run mage build with timeout
    log_debug "Running mage build (timeout: 10 minutes)..."
    
    local build_cmd="pct exec ${ct_id} -- bash -c 'cd ${source_dir} && export PATH=\$PATH:/usr/local/go/bin:/root/go/bin && export GOPATH=/root/go && mage build'"
    
    if ! safe_timeout 600 bash -c "${build_cmd}" 2>&1; then
        log_error "Backend build failed or timed out"
        log_error "Check build logs in container: pct enter ${ct_id}"
        log_error "Then run: cd ${source_dir} && mage build"
        return 1
    fi
    
    # Verify binary was created
    if ! pct_exec "$ct_id" test -f "${source_dir}/vikunja"; then
        log_error "Backend binary not found after build"
        return 1
    fi
    
    log_success "Backend built successfully"
    return 0
}

# Build Vikunja frontend
# Usage: build_frontend ct_id source_dir
# Returns: 0 on success, 1 on failure
build_frontend() {
    local ct_id="$1"
    local source_dir="$2"
    
    log_info "Building Vikunja frontend (this may take several minutes)"
    
    # Verify frontend directory exists
    if ! pct_exec "$ct_id" test -d "${source_dir}/frontend"; then
        log_error "Frontend directory not found: ${source_dir}/frontend"
        return 1
    fi
    
    # Check disk space (need ~3GB for node_modules + build)
    if ! check_disk_space "$ct_id" 3072 "${source_dir}/frontend"; then
        return 1
    fi
    
    # Run pnpm install and build with timeout
    log_debug "Installing frontend dependencies..."
    
    local install_cmd="pct exec ${ct_id} -- bash -c 'cd ${source_dir}/frontend && pnpm install --frozen-lockfile'"
    
    if ! safe_timeout 600 bash -c "${install_cmd}" 2>&1; then
        log_error "Frontend dependency installation failed or timed out"
        log_error "Check logs in container: pct enter ${ct_id}"
        log_error "Then run: cd ${source_dir}/frontend && pnpm install"
        return 1
    fi
    
    log_debug "Building frontend..."
    
    local build_cmd="pct exec ${ct_id} -- bash -c 'cd ${source_dir}/frontend && pnpm build'"
    
    if ! safe_timeout 600 bash -c "${build_cmd}" 2>&1; then
        log_error "Frontend build failed or timed out"
        log_error "Check logs in container: pct enter ${ct_id}"
        log_error "Then run: cd ${source_dir}/frontend && pnpm build"
        return 1
    fi
    
    # Verify dist directory was created
    if ! pct_exec "$ct_id" test -d "${source_dir}/frontend/dist"; then
        log_error "Frontend dist directory not found after build"
        return 1
    fi
    
    # Inject API URL auto-detection into frontend
    log_debug "Injecting API URL auto-detection into frontend..."
    
    # Create a small JavaScript file for API URL auto-detection
    local api_detect_js="// Auto-detect API URL for same-origin deployments
if (!localStorage.getItem('API_URL')) {
    localStorage.setItem('API_URL', window.location.origin);
}"
    
    # Write the auto-detection script to a file in the dist directory
    local result
    result=$(pct_exec "$ct_id" bash -c "cat > ${source_dir}/frontend/dist/api-detect.js <<'EOF'
${api_detect_js}
EOF
" 2>&1)
    
    if [[ $? -eq 0 ]]; then
        # Inject script tag into index.html before any other scripts
        pct_exec "$ct_id" bash -c "cd ${source_dir}/frontend/dist && \
            sed -i 's|<head>|<head><script src=\"/api-detect.js\"></script>|' index.html" 2>&1
        
        if [[ $? -eq 0 ]]; then
            log_debug "API URL auto-detection injected successfully"
        else
            log_warning "Failed to inject script tag into index.html"
        fi
    else
        log_warning "Failed to create API URL auto-detection script"
        log_debug "Error: ${result}"
    fi
    
    log_success "Frontend built successfully"
    return 0
}

# Build MCP server
# Usage: build_mcp ct_id source_dir
# Returns: 0 on success, 1 on failure
build_mcp() {
    local ct_id="$1"
    local source_dir="$2"
    
    log_info "Building MCP server (this may take a few minutes)"
    
    # Verify MCP directory exists
    if ! pct_exec "$ct_id" test -d "${source_dir}/mcp-server"; then
        log_error "MCP server directory not found: ${source_dir}/mcp-server"
        return 1
    fi
    
    # Check disk space (need ~1GB for node_modules + build)
    if ! check_disk_space "$ct_id" 1024 "${source_dir}/mcp-server"; then
        return 1
    fi
    
    # Run npm install and build with timeout
    log_debug "Installing MCP dependencies..."
    
    # Note: mcp-server uses package-lock.json (npm), not pnpm-lock.yaml
    # Use 'npm ci' for clean install with lockfile
    local install_cmd="pct exec ${ct_id} -- bash -c 'cd ${source_dir}/mcp-server && npm ci'"
    
    if ! safe_timeout 300 bash -c "${install_cmd}" 2>&1; then
        log_error "MCP dependency installation failed or timed out"
        log_error "Check logs in container: pct enter ${ct_id}"
        log_error "Then run: cd ${source_dir}/mcp-server && npm ci"
        return 1
    fi
    
    log_debug "Building MCP server..."
    
    # Use npm run build (consistent with npm ci above)
    local build_cmd="pct exec ${ct_id} -- bash -c 'cd ${source_dir}/mcp-server && npm run build'"
    
    if ! safe_timeout 300 bash -c "${build_cmd}" 2>&1; then
        log_error "MCP build failed or timed out"
        log_error "Check logs in container: pct enter ${ct_id}"
        log_error "Then run: cd ${source_dir}/mcp-server && npm run build"
        return 1
    fi
    
    # Verify dist directory was created
    if ! pct_exec "$ct_id" test -d "${source_dir}/mcp-server/dist"; then
        log_error "MCP dist directory not found after build"
        return 1
    fi
    
    log_success "MCP server built successfully"
    return 0
}

# ============================================================================
# Git Operations Functions (for T048 - User Story 2)
# ============================================================================

# Check for repository updates
# Usage: check_for_updates ct_id repo_dir
# Returns: 0 if updates available, 1 if up to date, 2 on error
check_for_updates() {
    local ct_id="$1"
    local repo_dir="$2"
    
    log_debug "Checking for updates in ${repo_dir}"
    
    # Fetch latest changes
    if ! pct_exec "$ct_id" git -C "$repo_dir" fetch origin 2>&1; then
        log_error "Failed to fetch updates"
        return 2
    fi
    
    # Check if there are differences
    local local_commit
    local_commit=$(pct_exec "$ct_id" git -C "$repo_dir" rev-parse HEAD)
    local remote_commit
    remote_commit=$(pct_exec "$ct_id" git -C "$repo_dir" rev-parse origin/main)
    
    if [[ "$local_commit" != "$remote_commit" ]]; then
        log_info "Updates available: ${local_commit:0:7} → ${remote_commit:0:7}"
        return 0
    fi
    
    log_info "Already up to date"
    return 1
}

# Pull latest changes from repository
# Usage: pull_latest ct_id repo_dir
# Returns: 0 on success, 1 on failure
pull_latest() {
    local ct_id="$1"
    local repo_dir="$2"
    
    log_info "Pulling latest changes"
    
    if ! pct_exec "$ct_id" git -C "$repo_dir" pull origin main 2>&1; then
        log_error "Failed to pull updates"
        return 1
    fi
    
    return 0
}

# Get current commit hash
# Usage: get_commit_hash ct_id repo_dir
# Returns: commit hash
get_commit_hash() {
    local ct_id="$1"
    local repo_dir="$2"
    
    pct_exec "$ct_id" git -C "$repo_dir" rev-parse --short HEAD
}

# Checkout specific commit
# Usage: checkout_commit ct_id repo_dir commit_hash
# Returns: 0 on success, 1 on failure
checkout_commit() {
    local ct_id="$1"
    local repo_dir="$2"
    local commit="$3"
    
    log_info "Checking out commit ${commit}"
    
    if ! pct_exec "$ct_id" git -C "$repo_dir" checkout "$commit" 2>&1; then
        log_error "Failed to checkout commit"
        return 1
    fi
    
    return 0
}

# ============================================================================
# Database Migration Functions (User Story 2 - Updates)
# ============================================================================

# run_migrations - Execute Vikunja database migrations
# Usage: run_migrations ct_id db_type db_config
# Returns: 0 on success, 1 on failure
run_migrations() {
    local ct_id="$1"
    local db_type="$2"
    local db_config="$3"
    
    if [[ -z "$ct_id" || -z "$db_type" ]]; then
        log_error "run_migrations: Missing required arguments"
        return 1
    fi
    
    log_info "Running database migrations..."
    
    local vikunja_binary="/opt/vikunja/vikunja"
    
    # Verify vikunja binary exists
    if ! pct_exec "$ct_id" "test -f $vikunja_binary"; then
        log_error "Vikunja binary not found: $vikunja_binary"
        return 1
    fi
    
    # Run migrations using Vikunja's built-in migrate command
    log_debug "Executing: $vikunja_binary migrate"
    
    if ! pct_exec "$ct_id" "cd /opt/vikunja && $vikunja_binary migrate"; then
        log_error "Database migration failed"
        return 1
    fi
    
    log_success "Database migrations completed"
    return 0
}

# check_migration_status - Verify migration table status
# Usage: check_migration_status ct_id db_type db_config
# Returns: 0 on success, prints migration count to stdout
check_migration_status() {
    local ct_id="$1"
    local db_type="$2"
    local db_config="$3"
    
    if [[ -z "$ct_id" || -z "$db_type" ]]; then
        log_error "check_migration_status: Missing required arguments"
        return 1
    fi
    
    log_debug "Checking migration status..."
    
    local migration_count=0
    
    case "$db_type" in
        sqlite)
            local db_path="$db_config"
            migration_count=$(pct_exec "$ct_id" "sqlite3 $db_path 'SELECT COUNT(*) FROM migration;' 2>/dev/null" || echo "0")
            ;;
        postgresql|postgres)
            IFS=':' read -r db_host db_port db_name db_user db_password <<< "$db_config"
            migration_count=$(pct_exec "$ct_id" "PGPASSWORD='$db_password' psql -h $db_host -p $db_port -U $db_user -d $db_name -t -c 'SELECT COUNT(*) FROM migration;' 2>/dev/null" | tr -d '[:space:]' || echo "0")
            ;;
        mysql)
            IFS=':' read -r db_host db_port db_name db_user db_password <<< "$db_config"
            migration_count=$(pct_exec "$ct_id" "mysql -h $db_host -P $db_port -u $db_user -p'$db_password' -D $db_name -sN -e 'SELECT COUNT(*) FROM migration;' 2>/dev/null" || echo "0")
            ;;
        *)
            log_error "Unsupported database type: $db_type"
            return 1
            ;;
    esac
    
    echo "$migration_count"
    log_debug "Applied migrations: $migration_count"
    return 0
}

log_debug "LXC setup library loaded"
