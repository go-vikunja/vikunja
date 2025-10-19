#!/usr/bin/env bash
# Vikunja Proxmox Deployment - Common Library Functions
# Provides: logging, validation, configuration, state management, error handling
# Required by: All deployment scripts

set -euo pipefail

# Prevent multiple sourcing
if [[ -n "${VIKUNJA_COMMON_LIB_LOADED:-}" ]]; then
    return 0
fi
readonly VIKUNJA_COMMON_LIB_LOADED=1

# ============================================================================
# Color Codes and Output Formatting Constants
# ============================================================================

# Check if terminal supports colors
if [[ -t 1 ]] && command -v tput >/dev/null 2>&1; then
    readonly COLOR_RESET=$(tput sgr0)
    readonly COLOR_BOLD=$(tput bold)
    readonly COLOR_RED=$(tput setaf 1)
    readonly COLOR_GREEN=$(tput setaf 2)
    readonly COLOR_YELLOW=$(tput setaf 3)
    readonly COLOR_BLUE=$(tput setaf 4)
    readonly COLOR_MAGENTA=$(tput setaf 5)
    readonly COLOR_CYAN=$(tput setaf 6)
    readonly COLOR_WHITE=$(tput setaf 7)
else
    readonly COLOR_RESET=""
    readonly COLOR_BOLD=""
    readonly COLOR_RED=""
    readonly COLOR_GREEN=""
    readonly COLOR_YELLOW=""
    readonly COLOR_BLUE=""
    readonly COLOR_MAGENTA=""
    readonly COLOR_CYAN=""
    readonly COLOR_WHITE=""
fi

# Unicode symbols
readonly SYMBOL_CHECK="✓"
readonly SYMBOL_CROSS="✗"
readonly SYMBOL_ARROW="→"
readonly SYMBOL_INFO="ℹ"
readonly SYMBOL_WARN="⚠"
readonly SYMBOL_BULLET="•"

# ============================================================================
# Logging Functions
# ============================================================================

# Log informational message
# Usage: log_info "message"
log_info() {
    local msg="$1"
    echo "${COLOR_BLUE}${SYMBOL_INFO}${COLOR_RESET} ${msg}"
}

# Log success message
# Usage: log_success "message"
log_success() {
    local msg="$1"
    echo "${COLOR_GREEN}${SYMBOL_CHECK}${COLOR_RESET} ${msg}"
}

# Log warning message
# Usage: log_warn "message"
log_warn() {
    local msg="$1"
    echo "${COLOR_YELLOW}${SYMBOL_WARN}${COLOR_RESET} ${msg}" >&2
}

# Log error message
# Usage: log_error "message"
log_error() {
    local msg="$1"
    echo "${COLOR_RED}${SYMBOL_CROSS}${COLOR_RESET} ${msg}" >&2
}

# Log debug message (only if DEBUG=1)
# Usage: log_debug "message"
log_debug() {
    local msg="${1:-}"
    if [[ -n "$msg" ]] && [[ "${DEBUG:-0}" == "1" ]]; then
        echo "${COLOR_CYAN}[DEBUG]${COLOR_RESET} ${msg}" >&2
    fi
}

# ============================================================================
# Progress Indicators
# ============================================================================

# Global variable for current progress operation
PROGRESS_PID=""
PROGRESS_MSG=""

# Start a progress indicator with spinner
# Usage: progress_start "message"
progress_start() {
    local msg="$1"
    PROGRESS_MSG="$msg"
    
    # Start spinner in background
    {
        local spin='⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏'
        local i=0
        while true; do
            printf "\r${COLOR_BLUE}${spin:$i:1}${COLOR_RESET} %s" "$PROGRESS_MSG"
            i=$(( (i + 1) % ${#spin} ))
            sleep 0.1
        done
    } &
    
    PROGRESS_PID=$!
    # Disable job control messages
    disown 2>/dev/null || true
}

# Update progress message
# Usage: progress_update "new message"
progress_update() {
    local msg="$1"
    PROGRESS_MSG="$msg"
}

# Complete progress with success
# Usage: progress_complete "completion message"
progress_complete() {
    local msg="$1"
    
    if [[ -n "$PROGRESS_PID" ]]; then
        kill "$PROGRESS_PID" 2>/dev/null || true
        wait "$PROGRESS_PID" 2>/dev/null || true
        PROGRESS_PID=""
    fi
    
    printf "\r\033[K"  # Clear line
    log_success "$msg"
}

# Complete progress with failure
# Usage: progress_fail "failure message"
progress_fail() {
    local msg="$1"
    
    if [[ -n "$PROGRESS_PID" ]]; then
        kill "$PROGRESS_PID" 2>/dev/null || true
        wait "$PROGRESS_PID" 2>/dev/null || true
        PROGRESS_PID=""
    fi
    
    printf "\r\033[K"  # Clear line
    log_error "$msg"
}

# ============================================================================
# Input Validation Functions
# ============================================================================

# Validate IP address (IPv4)
# Usage: validate_ip "192.168.1.1" || echo "Invalid"
# Returns: 0 if valid, 1 if invalid
validate_ip() {
    local ip="$1"
    local regex='^([0-9]{1,3}\.){3}[0-9]{1,3}$'
    
    if [[ ! $ip =~ $regex ]]; then
        return 1
    fi
    
    # Check each octet is <= 255
    local IFS='.'
    local -a octets=($ip)
    for octet in "${octets[@]}"; do
        if (( octet > 255 )); then
            return 1
        fi
    done
    
    return 0
}

# Validate domain name or hostname
# Usage: validate_domain "example.com" || echo "Invalid"
# Returns: 0 if valid, 1 if invalid
validate_domain() {
    local domain="$1"
    local regex='^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$'
    
    if [[ $domain =~ $regex ]] && [[ ${#domain} -le 253 ]]; then
        return 0
    fi
    
    return 1
}

# Validate port number
# Usage: validate_port "8080" || echo "Invalid"
# Returns: 0 if valid, 1 if invalid
validate_port() {
    local port="$1"
    
    if [[ "$port" =~ ^[0-9]+$ ]] && (( port >= 1 && port <= 65535 )); then
        return 0
    fi
    
    return 1
}

# Validate email address
# Usage: validate_email "user@example.com" || echo "Invalid"
# Returns: 0 if valid, 1 if invalid
validate_email() {
    local email="$1"
    local regex='^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
    
    if [[ $email =~ $regex ]]; then
        return 0
    fi
    
    return 1
}

# ============================================================================
# Privilege Check Functions
# ============================================================================

# Check if running as root
# Usage: check_root || error "Must run as root"
# Returns: 0 if root, 1 if not root
check_root() {
    if [[ $EUID -eq 0 ]]; then
        return 0
    fi
    return 1
}

# Check if running on Proxmox VE
# Usage: check_proxmox || error "Must run on Proxmox"
# Returns: 0 if Proxmox, 1 if not Proxmox
check_proxmox() {
    # Check for Proxmox-specific files and commands
    if [[ -f /etc/pve/.version ]] && command -v pct >/dev/null 2>&1; then
        return 0
    fi
    return 1
}

# ============================================================================
# Lock Management Functions
# ============================================================================

# Lock file directory
readonly LOCK_DIR="/var/lock/vikunja-deploy"

# Acquire deployment lock for instance
# Usage: acquire_lock "instance-id" || error "Lock already held"
# Returns: 0 if lock acquired, 1 if lock exists
acquire_lock() {
    local instance_id="$1"
    local lock_file="${LOCK_DIR}/${instance_id}.lock"
    
    # Create lock directory if needed
    mkdir -p "$LOCK_DIR"
    
    # Check if lock exists and is recent (< 2 hours old)
    if [[ -f "$lock_file" ]]; then
        local lock_age=$(($(date +%s) - $(stat -c %Y "$lock_file" 2>/dev/null || echo 0)))
        if (( lock_age < 7200 )); then
            log_debug "Lock file exists and is recent (${lock_age}s old)"
            return 1
        else
            log_warn "Removing stale lock file (${lock_age}s old)"
            rm -f "$lock_file"
        fi
    fi
    
    # Create lock file with PID and timestamp
    echo "$$:$(date +%s):$(whoami)" > "$lock_file"
    log_debug "Lock acquired: $lock_file"
    
    return 0
}

# Release deployment lock for instance
# Usage: release_lock "instance-id"
release_lock() {
    local instance_id="$1"
    local lock_file="${LOCK_DIR}/${instance_id}.lock"
    
    if [[ -f "$lock_file" ]]; then
        rm -f "$lock_file"
        log_debug "Lock released: $lock_file"
    fi
}

# Check if lock exists for instance
# Usage: check_lock "instance-id" && echo "Locked"
# Returns: 0 if lock exists, 1 if no lock
check_lock() {
    local instance_id="$1"
    local lock_file="${LOCK_DIR}/${instance_id}.lock"
    
    if [[ -f "$lock_file" ]]; then
        local lock_age=$(($(date +%s) - $(stat -c %Y "$lock_file" 2>/dev/null || echo 0)))
        if (( lock_age < 7200 )); then
            return 0
        fi
    fi
    
    return 1
}

# ============================================================================
# Configuration Management Functions
# ============================================================================

# Configuration directory
readonly CONFIG_DIR="/etc/vikunja-deploy"

# Load configuration for instance
# Usage: load_config "instance-id" or load_config "config-file-path"
# Sets global variables from config file
# Returns: 0 if loaded, 1 if not found
load_config() {
    local arg="$1"
    local config_file
    
    # Check if argument is a file path or instance ID
    if [[ -f "$arg" ]]; then
        # Direct file path provided
        config_file="$arg"
    else
        # Instance ID provided
        config_file="${CONFIG_DIR}/${arg}/config.yaml"
    fi
    
    if [[ ! -f "$config_file" ]]; then
        log_debug "Config file not found: $config_file"
        return 1
    fi
    
    log_debug "Loading config: $config_file"
    
    # Simple YAML parsing using grep and sed
    # Extract key-value pairs from YAML and export as environment variables
    
    # Parse deployment section
    DEPLOYMENT_NAME=$(grep -A5 "^deployment:" "$config_file" | grep "name:" | sed 's/.*name: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    DEPLOYMENT_ENV=$(grep -A5 "^deployment:" "$config_file" | grep "environment:" | sed 's/.*environment: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    DEPLOYMENT_VERSION=$(grep -A5 "^deployment:" "$config_file" | grep "version:" | sed 's/.*version: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    
    # Parse proxmox section
    PROXMOX_NODE=$(grep -A5 "^proxmox:" "$config_file" | grep "node:" | sed 's/.*node: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    PROXMOX_STORAGE=$(grep -A5 "^proxmox:" "$config_file" | grep "storage:" | sed 's/.*storage: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    PROXMOX_OSTEMPLATE=$(grep -A5 "^proxmox:" "$config_file" | grep "ostemplate:" | sed 's/.*ostemplate: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    
    # Parse resources section
    RESOURCES_CORES=$(grep -A5 "^resources:" "$config_file" | grep "cores:" | sed 's/.*cores: *\([0-9]*\).*/\1/')
    RESOURCES_MEMORY=$(grep -A5 "^resources:" "$config_file" | grep "memory:" | sed 's/.*memory: *\([0-9]*\).*/\1/')
    RESOURCES_SWAP=$(grep -A5 "^resources:" "$config_file" | grep "swap:" | sed 's/.*swap: *\([0-9]*\).*/\1/')
    RESOURCES_DISK=$(grep -A5 "^resources:" "$config_file" | grep "disk:" | sed 's/.*disk: *\([0-9]*\).*/\1/')
    
    # Parse network section
    NETWORK_BRIDGE=$(grep -A7 "^network:" "$config_file" | grep "bridge:" | sed 's/.*bridge: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    NETWORK_IP=$(grep -A7 "^network:" "$config_file" | grep "ip_address:" | sed 's/.*ip_address: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    NETWORK_GATEWAY=$(grep -A7 "^network:" "$config_file" | grep "gateway:" | sed 's/.*gateway: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    NETWORK_NAMESERVER=$(grep -A7 "^network:" "$config_file" | grep "nameserver:" | sed 's/.*nameserver: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    NETWORK_DOMAIN=$(grep -A7 "^network:" "$config_file" | grep "domain:" | sed 's/.*domain: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    
    # Parse database section
    DB_TYPE=$(grep -A8 "^database:" "$config_file" | grep "type:" | sed 's/.*type: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    DB_PATH=$(grep -A8 "^database:" "$config_file" | grep "path:" | sed 's/.*path: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    DB_HOST=$(grep -A8 "^database:" "$config_file" | grep "host:" | sed 's/.*host: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    DB_PORT=$(grep -A8 "^database:" "$config_file" | grep "port:" | sed 's/.*port: *\([0-9]*\).*/\1/')
    DB_NAME=$(grep -A8 "^database:" "$config_file" | grep "name:" | sed 's/.*name: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    DB_USER=$(grep -A8 "^database:" "$config_file" | grep "user:" | sed 's/.*user: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    DB_PASSWORD=$(grep -A8 "^database:" "$config_file" | grep "password:" | sed 's/.*password: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    
    # Parse services section
    BACKEND_BLUE_PORT=$(grep -A10 "^services:" "$config_file" | grep -A3 "backend:" | grep "blue_port:" | sed 's/.*blue_port: *\([0-9]*\).*/\1/')
    BACKEND_GREEN_PORT=$(grep -A10 "^services:" "$config_file" | grep -A3 "backend:" | grep "green_port:" | sed 's/.*green_port: *\([0-9]*\).*/\1/')
    MCP_BLUE_PORT=$(grep -A10 "^services:" "$config_file" | grep -A3 "mcp:" | grep "blue_port:" | sed 's/.*blue_port: *\([0-9]*\).*/\1/')
    MCP_GREEN_PORT=$(grep -A10 "^services:" "$config_file" | grep -A3 "mcp:" | grep "green_port:" | sed 's/.*green_port: *\([0-9]*\).*/\1/')
    
    # Parse git section
    GIT_BACKEND_REPO=$(grep -A15 "^git:" "$config_file" | grep "backend_repo:" | sed 's/.*backend_repo: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    GIT_BACKEND_BRANCH=$(grep -A15 "^git:" "$config_file" | grep "backend_branch:" | sed 's/.*backend_branch: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    GIT_FRONTEND_REPO=$(grep -A15 "^git:" "$config_file" | grep "frontend_repo:" | sed 's/.*frontend_repo: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    GIT_FRONTEND_BRANCH=$(grep -A15 "^git:" "$config_file" | grep "frontend_branch:" | sed 's/.*frontend_branch: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    GIT_MCP_REPO=$(grep -A15 "^git:" "$config_file" | grep "mcp_repo:" | sed 's/.*mcp_repo: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    GIT_MCP_BRANCH=$(grep -A15 "^git:" "$config_file" | grep "mcp_branch:" | sed 's/.*mcp_branch: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    
    # Parse admin section
    ADMIN_EMAIL=$(grep -A3 "^admin:" "$config_file" | grep "email:" | sed 's/.*email: *"\?\([^"]*\)"\?.*/\1/' | tr -d '"')
    
    # Export all variables
    export DEPLOYMENT_NAME DEPLOYMENT_ENV DEPLOYMENT_VERSION
    export PROXMOX_NODE PROXMOX_STORAGE PROXMOX_OSTEMPLATE
    export RESOURCES_CORES RESOURCES_MEMORY RESOURCES_SWAP RESOURCES_DISK
    export NETWORK_BRIDGE NETWORK_IP NETWORK_GATEWAY NETWORK_NAMESERVER NETWORK_DOMAIN
    export DB_TYPE DB_PATH DB_HOST DB_PORT DB_NAME DB_USER DB_PASSWORD
    export BACKEND_BLUE_PORT BACKEND_GREEN_PORT MCP_BLUE_PORT MCP_GREEN_PORT
    export GIT_BACKEND_REPO GIT_BACKEND_BRANCH GIT_FRONTEND_REPO GIT_FRONTEND_BRANCH
    export GIT_MCP_REPO GIT_MCP_BRANCH
    export ADMIN_EMAIL
    
    log_debug "Config loaded successfully"
    
    return 0
}

# Save configuration for instance
# Usage: save_config "instance-id" "config-data"
# Returns: 0 if saved, 1 on error
save_config() {
    local instance_id="$1"
    local config_data="$2"
    local config_dir="${CONFIG_DIR}/${instance_id}"
    local config_file="${config_dir}/config.yaml"
    
    # Create config directory
    mkdir -p "$config_dir"
    
    # Write config file
    echo "$config_data" > "$config_file"
    chmod 600 "$config_file"
    
    log_debug "Config saved: $config_file"
    
    return 0
}

# Update specific configuration value
# Usage: update_config "instance-id" "key" "value"
# Returns: 0 if updated, 1 on error
update_config() {
    local instance_id="$1"
    local key="$2"
    local value="$3"
    local config_file="${CONFIG_DIR}/${instance_id}/config.yaml"
    
    if [[ ! -f "$config_file" ]]; then
        log_error "Config file not found: $config_file"
        return 1
    fi
    
    # Simple update (would use yq in production)
    # For now, just log
    log_debug "Would update $key=$value in $config_file"
    
    return 0
}

# ============================================================================
# State Management Functions
# ============================================================================

# State directory
readonly STATE_DIR="/var/lib/vikunja-deploy"

# Get state value for instance
# Usage: version=$(get_state "instance-id" "deployed_version")
# Returns: state value or empty string
get_state() {
    local instance_id="$1"
    local key="$2"
    local state_file="${STATE_DIR}/${instance_id}/state"
    
    if [[ ! -f "$state_file" ]]; then
        return 0
    fi
    
    # Read value from state file
    grep "^${key}=" "$state_file" 2>/dev/null | cut -d= -f2- || true
}

# Set state value for instance
# Usage: set_state "instance-id" "key" "value"
# Returns: 0 on success
set_state() {
    local instance_id="$1"
    local key="$2"
    local value="${3:-}"
    local state_dir="${STATE_DIR}/${instance_id}"
    local state_file="${state_dir}/state"
    
    # Create state directory
    mkdir -p "$state_dir"
    
    # Update or add key=value
    if [[ -f "$state_file" ]]; then
        # Remove existing key if present
        grep -v "^${key}=" "$state_file" > "${state_file}.tmp" 2>/dev/null || true
        mv "${state_file}.tmp" "$state_file"
    fi
    
    # Add new value
    echo "${key}=${value}" >> "$state_file"
    
    log_debug "State updated: ${key}=${value}"
    
    return 0
}

# Update deployed version in state
# Usage: update_deployed_version "instance-id" "commit-hash"
# Returns: 0 on success
update_deployed_version() {
    local instance_id="$1"
    local version="$2"
    
    set_state "$instance_id" "deployed_version" "$version"
    set_state "$instance_id" "last_updated" "$(date +%s)"
    
    return 0
}

# ============================================================================
# Error Handling Functions
# ============================================================================

# Exit with error message
# Usage: error "Something went wrong" [exit_code]
# Exits with code (default: 1)
error() {
    local msg="$1"
    local code="${2:-1}"
    
    log_error "$msg"
    exit "$code"
}

# Set up error traps
# Usage: trap_errors
# Sets up trap handlers for cleanup
trap_errors() {
    trap cleanup_on_error ERR EXIT SIGINT SIGTERM
}

# Cleanup function called on error or exit
# Usage: Called automatically by trap_errors
# Override this function in scripts for custom cleanup
cleanup_on_error() {
    local exit_code=$?
    
    # Only run on non-zero exit
    if [[ $exit_code -ne 0 ]]; then
        log_debug "Cleanup triggered (exit code: $exit_code)"
        
        # Stop any running progress indicators
        if [[ -n "${PROGRESS_PID:-}" ]]; then
            kill "$PROGRESS_PID" 2>/dev/null || true
            wait "$PROGRESS_PID" 2>/dev/null || true
            printf "\r\033[K"  # Clear line
        fi
    fi
}

# ============================================================================
# Library Initialization
# ============================================================================

log_debug "Common library loaded"
