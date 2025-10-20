#!/usr/bin/env bash
# Vikunja Proxmox Deployment - Proxmox API Wrapper Functions
# Provides: CLI wrappers for pct, pvesh commands
# Required by: vikunja-install.sh, vikunja-update.sh, vikunja-manage.sh

set -euo pipefail

# Prevent multiple sourcing
if [[ -n "${VIKUNJA_PROXMOX_API_LIB_LOADED:-}" ]]; then
    return 0
fi
readonly VIKUNJA_PROXMOX_API_LIB_LOADED=1

# Common functions are sourced by main script before this library

# ============================================================================
# Proxmox Container (pct) Wrapper Functions
# ============================================================================

# Create LXC container
# Usage: pct_create container_id template [options...]
# Returns: 0 on success, 1 on failure
pct_create() {
    local ct_id="$1"
    local template="$2"
    shift 2
    local -a opts=("$@")
    
    log_debug "Creating container ${ct_id} from template ${template}"
    
    if ! pct create "$ct_id" "$template" "${opts[@]}" 2>&1 | tee >(log_debug); then
        log_error "Failed to create container ${ct_id}"
        return 1
    fi
    
    log_debug "Container ${ct_id} created successfully"
    return 0
}

# Start LXC container
# Usage: pct_start container_id
# Returns: 0 on success, 1 on failure
pct_start() {
    local ct_id="$1"
    
    log_debug "Starting container ${ct_id}"
    
    if ! pct start "$ct_id" 2>&1 | tee >(log_debug); then
        log_error "Failed to start container ${ct_id}"
        return 1
    fi
    
    # Wait for container to be fully started
    local max_wait=30
    local waited=0
    while ! pct_is_running "$ct_id"; do
        if (( waited >= max_wait )); then
            log_error "Container ${ct_id} failed to start within ${max_wait}s"
            return 1
        fi
        sleep 1
        ((waited++))
    done
    
    log_debug "Container ${ct_id} started successfully"
    return 0
}

# Stop LXC container
# Usage: pct_stop container_id [timeout]
# Returns: 0 on success, 1 on failure
pct_stop() {
    local ct_id="$1"
    local timeout="${2:-60}"
    
    log_debug "Stopping container ${ct_id}"
    
    if ! pct stop "$ct_id" --timeout "$timeout" 2>&1 | tee >(log_debug); then
        log_error "Failed to stop container ${ct_id}"
        return 1
    fi
    
    log_debug "Container ${ct_id} stopped successfully"
    return 0
}

# Execute command in container
# Usage: pct_exec container_id command [args...]
# Returns: exit code of command
pct_exec() {
    local ct_id="$1"
    shift
    local -a cmd=("$@")
    
    log_debug "Executing in container ${ct_id}: ${cmd[*]}"
    
    pct exec "$ct_id" -- "${cmd[@]}"
    return $?
}

# Check if container is running
# Usage: pct_is_running container_id
# Returns: 0 if running, 1 if not
pct_is_running() {
    local ct_id="$1"
    
    local status
    status=$(pct status "$ct_id" 2>/dev/null | awk '{print $2}')
    
    if [[ "$status" == "running" ]]; then
        return 0
    fi
    
    return 1
}

# Check if container exists
# Usage: pct_exists container_id
# Returns: 0 if exists, 1 if not
pct_exists() {
    local ct_id="$1"
    
    if pct status "$ct_id" >/dev/null 2>&1; then
        return 0
    fi
    
    return 1
}

# Destroy container
# Usage: pct_destroy container_id [--purge]
# Returns: 0 on success, 1 on failure
pct_destroy() {
    local ct_id="$1"
    local purge="${2:---purge}"
    
    log_debug "Destroying container ${ct_id}"
    
    # Stop container if running
    if pct_is_running "$ct_id"; then
        pct_stop "$ct_id" || log_warn "Failed to stop container before destroy"
    fi
    
    if ! pct destroy "$ct_id" "$purge" 2>&1 | tee >(log_debug); then
        log_error "Failed to destroy container ${ct_id}"
        return 1
    fi
    
    log_debug "Container ${ct_id} destroyed successfully"
    return 0
}

# Push file to container
# Usage: pct_push container_id local_file remote_path
# Returns: 0 on success, 1 on failure
pct_push() {
    local ct_id="$1"
    local local_file="$2"
    local remote_path="$3"
    
    log_debug "Pushing ${local_file} to container ${ct_id}:${remote_path}"
    
    if ! pct push "$ct_id" "$local_file" "$remote_path" 2>&1 | tee >(log_debug); then
        log_error "Failed to push file to container ${ct_id}"
        return 1
    fi
    
    return 0
}

# Pull file from container
# Usage: pct_pull container_id remote_path local_file
# Returns: 0 on success, 1 on failure
pct_pull() {
    local ct_id="$1"
    local remote_path="$2"
    local local_file="$3"
    
    log_debug "Pulling ${remote_path} from container ${ct_id} to ${local_file}"
    
    if ! pct pull "$ct_id" "$remote_path" "$local_file" 2>&1 | tee >(log_debug); then
        log_error "Failed to pull file from container ${ct_id}"
        return 1
    fi
    
    return 0
}

# ============================================================================
# Proxmox VE API (pvesh) Wrapper Functions
# ============================================================================

# Get resource information via pvesh
# Usage: pvesh_get path
# Returns: JSON output from API
pvesh_get() {
    local path="$1"
    
    log_debug "Getting API data from: ${path}"
    
    if ! pvesh get "$path" --output-format json 2>/dev/null; then
        log_error "Failed to query API path: ${path}"
        return 1
    fi
    
    return 0
}

# Get node information
# Usage: get_node_info
# Returns: Node name
get_node_info() {
    pvesh get /nodes --output-format json 2>/dev/null | \
        grep -oP '"node"\s*:\s*"\K[^"]+' | head -1
}

# Get next available container ID
# Usage: get_next_container_id [start_id]
# Returns: Next available ID
get_next_container_id() {
    local start_id="${1:-100}"
    
    log_debug "Finding next available container ID starting from ${start_id}"
    
    # Get all existing container IDs
    local -a existing_ids
    mapfile -t existing_ids < <(pvesh get /cluster/resources --type vm --output-format json 2>/dev/null | \
        grep -oP '"vmid"\s*:\s*\K[0-9]+' | sort -n)
    
    # Find first available ID
    local next_id=$start_id
    for id in "${existing_ids[@]}"; do
        if (( id == next_id )); then
            ((next_id++))
        elif (( id > next_id )); then
            break
        fi
    done
    
    # Ensure ID is in valid range
    if (( next_id > 999 )); then
        log_error "No available container IDs in range 100-999"
        return 1
    fi
    
    echo "$next_id"
    return 0
}

# Check if port is in use on host
# Usage: port_in_use port
# Returns: 0 if in use, 1 if free
port_in_use() {
    local port="$1"
    
    if ss -tuln | grep -q ":${port} "; then
        return 0
    fi
    
    return 1
}

# Check resource availability on node
# Usage: check_resources_available cores memory_mb disk_gb
# Returns: 0 if available, 1 if not
check_resources_available() {
    local req_cores="$1"
    local req_memory_mb="$2"
    local req_disk_gb="$3"
    
    log_debug "Checking resource availability: ${req_cores} cores, ${req_memory_mb}MB RAM, ${req_disk_gb}GB disk"
    
    # Get node stats
    local node
    node=$(get_node_info)
    
    if [[ -z "$node" ]]; then
        log_error "Failed to get node information"
        return 1
    fi
    
    local node_status
    node_status=$(pvesh_get "/nodes/${node}/status") || return 1
    
    # Parse available resources (simplified - in production use jq)
    local avail_memory
    avail_memory=$(echo "$node_status" | grep -oP '"memory".*?"free"\s*:\s*\K[0-9]+' || echo "0")
    avail_memory=$((avail_memory / 1024 / 1024))  # Convert to MB
    
    if (( avail_memory < req_memory_mb )); then
        log_error "Insufficient memory: ${avail_memory}MB available, ${req_memory_mb}MB required"
        return 1
    fi
    
    log_debug "Resource check passed"
    return 0
}

# Get container IP address
# Usage: get_container_ip container_id
# Returns: IP address or empty string
get_container_ip() {
    local ct_id="$1"
    
    local config
    config=$(pct config "$ct_id" 2>/dev/null) || return 1
    
    # Extract IP from net0 configuration
    echo "$config" | grep -oP 'ip=\K[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+' | head -1
}

# ============================================================================
# Helper Functions
# ============================================================================

# Wait for container network to be ready
# Usage: wait_for_container_network container_id [timeout]
# Returns: 0 if network ready, 1 on timeout
wait_for_container_network() {
    local ct_id="$1"
    local timeout="${2:-60}"
    local waited=0
    
    log_debug "Waiting for container ${ct_id} network to be ready"
    
    while (( waited < timeout )); do
        if pct_exec "$ct_id" ip addr show | grep -q "inet "; then
            log_debug "Container network is ready"
            return 0
        fi
        sleep 2
        ((waited += 2))
    done
    
    log_error "Container network not ready after ${timeout}s"
    return 1
}

# Wait for container to respond to ping
# Usage: wait_for_ping ip_address [timeout]
# Returns: 0 if responding, 1 on timeout
wait_for_ping() {
    local ip="$1"
    local timeout="${2:-30}"
    local waited=0
    
    log_debug "Waiting for ${ip} to respond to ping"
    
    while (( waited < timeout )); do
        if ping -c 1 -W 1 "$ip" >/dev/null 2>&1; then
            log_debug "Host is responding to ping"
            return 0
        fi
        sleep 2
        ((waited += 2))
    done
    
    log_error "Host not responding after ${timeout}s"
    return 1
}

log_debug "Proxmox API library loaded"
