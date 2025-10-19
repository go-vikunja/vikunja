#!/usr/bin/env bash
# Vikunja Proxmox Deployment - Service Setup Functions
# Provides: Systemd service management
# Required by: vikunja-install.sh, vikunja-update.sh, vikunja-manage.sh

set -euo pipefail

# Prevent multiple sourcing
if [[ -n "${VIKUNJA_SERVICE_SETUP_LIB_LOADED:-}" ]]; then
    return 0
fi
readonly VIKUNJA_SERVICE_SETUP_LIB_LOADED=1

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=./common.sh
source "${SCRIPT_DIR}/common.sh"
# shellcheck source=./proxmox-api.sh
source "${SCRIPT_DIR}/proxmox-api.sh"

# ============================================================================
# Systemd Service Creation Functions (T028)
# ============================================================================

# Generate systemd unit file for backend service
# Usage: generate_systemd_unit ct_id service_name color [port] [working_dir]
# Returns: 0 on success, 1 on failure
generate_systemd_unit() {
    local ct_id="$1"
    local service_type="$2"
    local color="$3"
    local port="${4:-${BACKEND_BLUE_PORT:-3456}}"
    local working_dir="${5:-/opt/vikunja}"
    
    # Construct full service name from type and color
    local service_name="vikunja-${service_type}-${color}"
    
    log_info "Generating systemd unit: ${service_name}"
    
    local unit_file="/etc/systemd/system/${service_name}.service"
    
    # Create unit file content based on service type
    local unit_content
    if [[ "$service_type" == "backend" ]]; then
        unit_content=$(cat <<EOF
[Unit]
Description=Vikunja Backend Service (${color})
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${working_dir}
ExecStart=${working_dir}/vikunja
Environment="VIKUNJA_SERVICE_FRONTENDURL=http://localhost:${port}"
Environment="VIKUNJA_SERVICE_INTERFACE=:${port}"
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF
)
    elif [[ "$service_type" == "mcp" ]]; then
        unit_content=$(cat <<EOF
[Unit]
Description=Vikunja MCP Server (${color})
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${working_dir}/mcp-server
ExecStart=/usr/bin/node ${working_dir}/mcp-server/dist/index.js
Environment="PORT=${port}"
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF
)
    else
        log_error "Unknown service type: ${service_type}"
        return 1
    fi
    
    # Write unit file to container
    pct_exec "$ct_id" bash -c "cat > ${unit_file} <<'UNITEOF'
${unit_content}
UNITEOF
" || return 1
    
    log_success "Unit file created: ${unit_file}"
    return 0
}

# Enable systemd service
# Usage: enable_service ct_id service_name
# Returns: 0 on success, 1 on failure
enable_service() {
    local ct_id="$1"
    local service_name="$2"
    
    log_debug "Enabling service: ${service_name}"
    
    # Reload systemd
    pct_exec "$ct_id" systemctl daemon-reload 2>&1 | tee >(log_debug) || return 1
    
    # Enable service
    pct_exec "$ct_id" systemctl enable "$service_name" 2>&1 | tee >(log_debug) || return 1
    
    log_success "Service enabled: ${service_name}"
    return 0
}

# Start systemd service
# Usage: start_service ct_id service_name
# Returns: 0 on success, 1 on failure
start_service() {
    local ct_id="$1"
    local service_name="$2"
    
    log_info "Starting service: ${service_name}"
    
    # Start service
    if ! pct_exec "$ct_id" systemctl start "$service_name" 2>&1 | tee >(log_debug); then
        log_error "Failed to start service: ${service_name}"
        # Show service status for debugging
        pct_exec "$ct_id" systemctl status "$service_name" 2>&1 | tee >(log_debug)
        return 1
    fi
    
    # Wait for service to be active
    local max_wait=30
    local waited=0
    while (( waited < max_wait )); do
        if pct_exec "$ct_id" systemctl is-active --quiet "$service_name"; then
            log_success "Service started: ${service_name}"
            return 0
        fi
        sleep 1
        ((waited++))
    done
    
    log_error "Service failed to become active: ${service_name}"
    return 1
}

# Stop systemd service
# Usage: stop_service ct_id service_name
# Returns: 0 on success, 1 on failure
stop_service() {
    local ct_id="$1"
    local service_name="$2"
    
    log_info "Stopping service: ${service_name}"
    
    if ! pct_exec "$ct_id" systemctl stop "$service_name" 2>&1 | tee >(log_debug); then
        log_warn "Failed to stop service: ${service_name}"
        return 1
    fi
    
    log_success "Service stopped: ${service_name}"
    return 0
}

# Restart systemd service
# Usage: restart_service ct_id service_name
# Returns: 0 on success, 1 on failure
restart_service() {
    local ct_id="$1"
    local service_name="$2"
    
    log_info "Restarting service: ${service_name}"
    
    if ! pct_exec "$ct_id" systemctl restart "$service_name" 2>&1 | tee >(log_debug); then
        log_error "Failed to restart service: ${service_name}"
        return 1
    fi
    
    log_success "Service restarted: ${service_name}"
    return 0
}

# Check if service is active
# Usage: is_service_active ct_id service_name
# Returns: 0 if active, 1 if not
is_service_active() {
    local ct_id="$1"
    local service_name="$2"
    
    pct_exec "$ct_id" systemctl is-active --quiet "$service_name"
    return $?
}

# Get service status
# Usage: get_service_status ct_id service_name
# Returns: status string (active, inactive, failed, etc.)
get_service_status() {
    local ct_id="$1"
    local service_name="$2"
    
    pct_exec "$ct_id" systemctl is-active "$service_name" 2>/dev/null || echo "inactive"
}

# ============================================================================
# Graceful Restart Functions (for T068 - User Story 3)
# ============================================================================

# Gracefully restart backend service
# Usage: graceful_restart_backend ct_id service_name
# Returns: 0 on success, 1 on failure
graceful_restart_backend() {
    local ct_id="$1"
    local service_name="$2"
    
    log_info "Gracefully restarting backend: ${service_name}"
    
    # Send SIGTERM and wait for graceful shutdown
    if ! pct_exec "$ct_id" systemctl reload-or-restart "$service_name" 2>&1 | tee >(log_debug); then
        log_error "Failed to gracefully restart backend"
        return 1
    fi
    
    # Wait for service to be active
    sleep 5
    if is_service_active "$ct_id" "$service_name"; then
        log_success "Backend restarted gracefully"
        return 0
    fi
    
    log_error "Backend failed to restart"
    return 1
}

# Gracefully restart MCP service
# Usage: graceful_restart_mcp ct_id service_name
# Returns: 0 on success, 1 on failure
graceful_restart_mcp() {
    local ct_id="$1"
    local service_name="$2"
    
    log_info "Gracefully restarting MCP: ${service_name}"
    
    # MCP server handles SIGTERM for graceful shutdown
    if ! pct_exec "$ct_id" systemctl reload-or-restart "$service_name" 2>&1 | tee >(log_debug); then
        log_error "Failed to gracefully restart MCP"
        return 1
    fi
    
    # Wait for service to be active
    sleep 3
    if is_service_active "$ct_id" "$service_name"; then
        log_success "MCP restarted gracefully"
        return 0
    fi
    
    log_error "MCP failed to restart"
    return 1
}

# Reload nginx configuration
# Usage: reload_nginx_config ct_id
# Returns: 0 on success, 1 on failure
reload_nginx_config() {
    local ct_id="$1"
    
    log_info "Reloading nginx configuration"
    
    # Test configuration first
    if ! pct_exec "$ct_id" nginx -t 2>&1 | tee >(log_debug); then
        log_error "Nginx configuration test failed"
        return 1
    fi
    
    # Reload nginx
    if ! pct_exec "$ct_id" systemctl reload nginx 2>&1 | tee >(log_debug); then
        log_error "Failed to reload nginx"
        return 1
    fi
    
    log_success "Nginx configuration reloaded"
    return 0
}

log_debug "Service setup library loaded"
