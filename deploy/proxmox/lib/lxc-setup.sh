#!/usr/bin/env bash
# Vikunja Proxmox Deployment - LXC Setup and Provisioning Functions
# Provides: Container creation, provisioning, database setup, build functions
# Required by: vikunja-install.sh, vikunja-update.sh

set -euo pipefail

# Common and proxmox-api functions are sourced by main script before this library

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

# Install system dependencies in container
# Usage: install_dependencies ct_id
# Returns: 0 on success, 1 on failure
install_dependencies() {
    local ct_id="$1"
    
    log_info "Installing system dependencies"
    
    # Update package lists
    pct_exec "$ct_id" apt-get update 2>&1 || return 1
    
    # Install required packages
    local packages=(
        "git" "curl" "wget" "build-essential"
        "ca-certificates" "gnupg" "lsb-release"
        "nginx" "sqlite3" "postgresql-client" "mysql-client"
        "sudo" "systemd" "procps"
    )
    
    pct_exec "$ct_id" apt-get install -y "${packages[@]}" 2>&1 || return 1
    
    log_success "System dependencies installed"
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
    local node_version="${2:-18}"
    
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
    
    # Clone repository
    pct_exec "$ct_id" git clone --depth 1 --branch "$branch" "$repo_url" "$target_dir" \
        2>&1 || return 1
    
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
    
    # Test connection
    if ! test_db_connection "$ct_id" "postgresql" "$host" "$port" "$dbname" "$user" "$password"; then
        log_error "Failed to connect to PostgreSQL database"
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
    local host="$2"
    local port="$3"
    local dbname="$4"
    local user="$5"
    local password="$6"
    
    log_info "Setting up MySQL connection to ${host}:${port}"
    
    # Test connection
    if ! test_db_connection "$ct_id" "mysql" "$host" "$port" "$dbname" "$user" "$password"; then
        log_error "Failed to connect to MySQL database"
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
    
    log_debug "Testing ${type} connection to ${host}:${port}"
    
    case "$type" in
        postgresql)
            pct_exec "$ct_id" bash -c \
                "PGPASSWORD='${password}' psql -h ${host} -p ${port} -U ${user} -d ${dbname} -c 'SELECT 1'" \
                >/dev/null 2>&1
            ;;
        mysql)
            pct_exec "$ct_id" bash -c \
                "mysql -h ${host} -P ${port} -u ${user} -p'${password}' ${dbname} -e 'SELECT 1'" \
                >/dev/null 2>&1
            ;;
        *)
            log_error "Unknown database type: ${type}"
            return 1
            ;;
    esac
    
    return $?
}

# ============================================================================
# Build Functions (T027)
# ============================================================================

# Build Vikunja backend
# Usage: build_backend ct_id source_dir
# Returns: 0 on success, 1 on failure
build_backend() {
    local ct_id="$1"
    local source_dir="$2"
    
    log_info "Building Vikunja backend"
    
    # Run mage build
    pct_exec "$ct_id" bash -c "
        cd ${source_dir} && \
        export PATH=\$PATH:/usr/local/go/bin && \
        export GOPATH=/root/go && \
        go install github.com/magefile/mage@latest && \
        /root/go/bin/mage build
    " 2>&1 || return 1
    
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
    
    log_info "Building Vikunja frontend"
    
    # Run pnpm build
    pct_exec "$ct_id" bash -c "
        cd ${source_dir}/frontend && \
        pnpm install --frozen-lockfile && \
        pnpm build
    " 2>&1 || return 1
    
    # Verify dist directory was created
    if ! pct_exec "$ct_id" test -d "${source_dir}/frontend/dist"; then
        log_error "Frontend dist directory not found after build"
        return 1
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
    
    log_info "Building MCP server"
    
    # Run pnpm build
    pct_exec "$ct_id" bash -c "
        cd ${source_dir}/mcp-server && \
        pnpm install --frozen-lockfile && \
        pnpm build
    " 2>&1 || return 1
    
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

log_debug "LXC setup library loaded"
