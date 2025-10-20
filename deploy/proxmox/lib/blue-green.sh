#!/usr/bin/env bash
# Blue-Green Deployment Library
# Purpose: Zero-downtime deployment using port-based blue-green pattern
# Feature: 004-proxmox-deployment - User Story 2

# Port allocation (referenced in research.md Section 2)
readonly BLUE_BACKEND_PORT=3456
readonly GREEN_BACKEND_PORT=3457
readonly BLUE_MCP_PORT=8456
readonly GREEN_MCP_PORT=8457

#===============================================================================
# determine_inactive_color
#
# Determines which color (blue/green) is currently inactive
#
# Arguments:
#   $1 - Instance ID
#   $2 - Container ID
#
# Returns:
#   0 on success
#   1 on error
#
# Output:
#   Prints inactive color to stdout: "blue" or "green"
#===============================================================================
determine_inactive_color() {
    local instance_id="$1"
    local container_id="$2"
    
    if [[ -z "$instance_id" || -z "$container_id" ]]; then
        log_error "determine_inactive_color: Missing required arguments"
        return 1
    fi
    
    local state_file="/etc/vikunja/${instance_id}.state"
    
    if [[ ! -f "$state_file" ]]; then
        log_error "State file not found: $state_file"
        return 1
    fi
    
    # Read active color from state file
    local active_color
    active_color=$(grep "active_color:" "$state_file" | awk '{print $2}' | tr -d '"' || echo "blue")
    
    log_debug "Current active color: $active_color"
    
    # Return the inactive color
    if [[ "$active_color" == "blue" ]]; then
        echo "green"
    else
        echo "blue"
    fi
    
    return 0
}

#===============================================================================
# get_active_color
#
# Gets the currently active color from state file
#
# Arguments:
#   $1 - Instance ID
#
# Returns:
#   0 on success
#   1 on error
#
# Output:
#   Prints active color to stdout: "blue" or "green"
#===============================================================================
get_active_color() {
    local instance_id="$1"
    
    if [[ -z "$instance_id" ]]; then
        log_error "get_active_color: Missing instance ID"
        return 1
    fi
    
    local state_file="/etc/vikunja/${instance_id}.state"
    
    if [[ ! -f "$state_file" ]]; then
        log_error "State file not found: $state_file"
        return 1
    fi
    
    local active_color
    active_color=$(grep "active_color:" "$state_file" | awk '{print $2}' | tr -d '"' || echo "blue")
    
    echo "$active_color"
    return 0
}

#===============================================================================
# get_color_ports
#
# Returns the backend and MCP ports for a given color
#
# Arguments:
#   $1 - Color ("blue" or "green")
#
# Returns:
#   0 on success
#   1 on error
#
# Output:
#   Prints "BACKEND_PORT MCP_PORT" to stdout
#===============================================================================
get_color_ports() {
    local color="$1"
    
    if [[ "$color" == "blue" ]]; then
        echo "$BLUE_BACKEND_PORT $BLUE_MCP_PORT"
    elif [[ "$color" == "green" ]]; then
        echo "$GREEN_BACKEND_PORT $GREEN_MCP_PORT"
    else
        log_error "Invalid color: $color (must be 'blue' or 'green')"
        return 1
    fi
    
    return 0
}

#===============================================================================
# start_services_on_color
#
# Starts backend and MCP services for a specific color
#
# Arguments:
#   $1 - Instance ID
#   $2 - Container ID
#   $3 - Color ("blue" or "green")
#
# Returns:
#   0 on success
#   1 on error
#===============================================================================
start_services_on_color() {
    local instance_id="$1"
    local container_id="$2"
    local color="$3"
    
    if [[ -z "$instance_id" || -z "$container_id" || -z "$color" ]]; then
        log_error "start_services_on_color: Missing required arguments"
        return 1
    fi
    
    log_info "Starting services on $color..."
    
    # Start backend service
    local backend_service="vikunja-backend-${color}.service"
    if ! pct_exec "$container_id" "systemctl start $backend_service"; then
        log_error "Failed to start $backend_service"
        return 1
    fi
    
    log_debug "Started $backend_service"
    
    # Start MCP service
    local mcp_service="vikunja-mcp-${color}.service"
    if ! pct_exec "$container_id" "systemctl start $mcp_service"; then
        log_error "Failed to start $mcp_service"
        return 1
    fi
    
    log_debug "Started $mcp_service"
    
    log_success "Services started on $color"
    return 0
}

#===============================================================================
# stop_services_on_color
#
# Stops backend and MCP services for a specific color
#
# Arguments:
#   $1 - Instance ID
#   $2 - Container ID
#   $3 - Color ("blue" or "green")
#
# Returns:
#   0 on success
#   1 on error (non-fatal - logs warning)
#===============================================================================
stop_services_on_color() {
    local instance_id="$1"
    local container_id="$2"
    local color="$3"
    
    if [[ -z "$instance_id" || -z "$container_id" || -z "$color" ]]; then
        log_error "stop_services_on_color: Missing required arguments"
        return 1
    fi
    
    log_info "Stopping services on $color..."
    
    # Stop backend service
    local backend_service="vikunja-backend-${color}.service"
    if ! pct_exec "$container_id" "systemctl stop $backend_service"; then
        log_warn "Failed to stop $backend_service (may not be running)"
    else
        log_debug "Stopped $backend_service"
    fi
    
    # Stop MCP service
    local mcp_service="vikunja-mcp-${color}.service"
    if ! pct_exec "$container_id" "systemctl stop $mcp_service"; then
        log_warn "Failed to stop $mcp_service (may not be running)"
    else
        log_debug "Stopped $mcp_service"
    fi
    
    log_success "Services stopped on $color"
    return 0
}

#===============================================================================
# switch_traffic
#
# Switches nginx traffic to the specified color (zero-downtime)
#
# Arguments:
#   $1 - Instance ID
#   $2 - Container ID
#   $3 - Target color ("blue" or "green")
#
# Returns:
#   0 on success
#   1 on error
#===============================================================================
switch_traffic() {
    local instance_id="$1"
    local container_id="$2"
    local target_color="$3"
    
    if [[ -z "$instance_id" || -z "$container_id" || -z "$target_color" ]]; then
        log_error "switch_traffic: Missing required arguments"
        return 1
    fi
    
    log_info "Switching traffic to $target_color..."
    
    # Get target ports
    local ports
    ports=$(get_color_ports "$target_color")
    local backend_port
    local mcp_port
    backend_port=$(echo "$ports" | awk '{print $1}')
    mcp_port=$(echo "$ports" | awk '{print $2}')
    
    log_debug "Target ports: backend=$backend_port, mcp=$mcp_port"
    
    # Update nginx upstream configuration
    if ! update_nginx_upstream "$instance_id" "$container_id" "$backend_port"; then
        log_error "Failed to update nginx upstream"
        return 1
    fi
    
    # Update state file with new active color
    local state_file="/etc/vikunja/${instance_id}.state"
    update_state "$instance_id" "active_color" "$target_color"
    
    log_success "Traffic switched to $target_color (backend: $backend_port, mcp: $mcp_port)"
    return 0
}

#===============================================================================
# rollback_to_color
#
# Rolls back to the specified color after a failed deployment
#
# Arguments:
#   $1 - Instance ID
#   $2 - Container ID
#   $3 - Target color to rollback to ("blue" or "green")
#
# Returns:
#   0 on success
#   1 on error
#===============================================================================
rollback_to_color() {
    local instance_id="$1"
    local container_id="$2"
    local target_color="$3"
    
    if [[ -z "$instance_id" || -z "$container_id" || -z "$target_color" ]]; then
        log_error "rollback_to_color: Missing required arguments"
        return 1
    fi
    
    log_warn "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_warn "ROLLBACK: Reverting to $target_color"
    log_warn "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    local failed_color
    if [[ "$target_color" == "blue" ]]; then
        failed_color="green"
    else
        failed_color="blue"
    fi
    
    # Step 1: Verify target color services are running
    log_info "[1/4] Verifying $target_color services..."
    local backend_service="vikunja-backend-${target_color}.service"
    
    local service_status
    service_status=$(pct_exec "$container_id" "systemctl is-active $backend_service" || echo "inactive")
    
    if [[ "$service_status" != "active" ]]; then
        log_warn "$target_color services not running, attempting to start..."
        if ! start_services_on_color "$instance_id" "$container_id" "$target_color"; then
            log_error "Failed to start $target_color services"
            return 1
        fi
        sleep 5  # Give services time to start
    fi
    
    log_success "$target_color services verified"
    
    # Step 2: Switch traffic back to target color
    log_info "[2/4] Switching traffic back to $target_color..."
    if ! switch_traffic "$instance_id" "$container_id" "$target_color"; then
        log_error "Failed to switch traffic during rollback"
        return 1
    fi
    
    log_success "Traffic switched to $target_color"
    
    # Step 3: Stop failed color services
    log_info "[3/4] Stopping $failed_color services..."
    stop_services_on_color "$instance_id" "$container_id" "$failed_color" || log_warn "Could not stop all $failed_color services"
    
    # Step 4: Verify rollback health
    log_info "[4/4] Verifying rollback health..."
    if ! check_component_health "$container_id" "$target_color"; then
        log_error "Health check failed after rollback - manual intervention required"
        return 1
    fi
    
    log_success "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_success "ROLLBACK COMPLETED: $target_color is active"
    log_success "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    return 0
}

#===============================================================================
# cleanup_failed_deployment
#
# Cleans up artifacts from a failed deployment attempt
#
# Arguments:
#   $1 - Instance ID
#   $2 - Container ID
#   $3 - Failed color ("blue" or "green")
#
# Returns:
#   0 on success
#===============================================================================
cleanup_failed_deployment() {
    local instance_id="$1"
    local container_id="$2"
    local failed_color="$3"
    
    if [[ -z "$instance_id" || -z "$container_id" || -z "$failed_color" ]]; then
        log_error "cleanup_failed_deployment: Missing required arguments"
        return 1
    fi
    
    log_info "Cleaning up failed $failed_color deployment..."
    
    # Stop services (if running)
    stop_services_on_color "$instance_id" "$container_id" "$failed_color" || true
    
    # Note: We keep the binaries for forensics and potential retry
    # Cleanup of old binaries happens during successful deployments
    
    log_info "Cleanup complete (binaries preserved for analysis)"
    return 0
}

#===============================================================================
# verify_blue_green_ready
#
# Verifies system is ready for blue-green deployment
#
# Arguments:
#   $1 - Instance ID
#   $2 - Container ID
#
# Returns:
#   0 if ready
#   1 if not ready
#===============================================================================
verify_blue_green_ready() {
    local instance_id="$1"
    local container_id="$2"
    
    if [[ -z "$instance_id" || -z "$container_id" ]]; then
        log_error "verify_blue_green_ready: Missing required arguments"
        return 1
    fi
    
    log_debug "Verifying blue-green readiness..."
    
    # Check state file exists
    local state_file="/etc/vikunja/${instance_id}.state"
    if [[ ! -f "$state_file" ]]; then
        log_error "State file not found: $state_file"
        return 1
    fi
    
    # Check active color is set
    local active_color
    active_color=$(get_active_color "$instance_id")
    if [[ -z "$active_color" ]]; then
        log_error "Active color not set in state file"
        return 1
    fi
    
    # Verify systemd service files exist for both colors
    for color in blue green; do
        local backend_service="vikunja-backend-${color}.service"
        if ! pct_exec "$container_id" "systemctl cat $backend_service &>/dev/null"; then
            log_error "Service not found: $backend_service"
            return 1
        fi
    done
    
    log_debug "Blue-green deployment ready (active: $active_color)"
    return 0
}

# Export functions for use in other scripts
export -f determine_inactive_color
export -f get_active_color
export -f get_color_ports
export -f start_services_on_color
export -f stop_services_on_color
export -f switch_traffic
export -f rollback_to_color
export -f cleanup_failed_deployment
export -f verify_blue_green_ready
