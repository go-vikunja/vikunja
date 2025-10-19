#!/usr/bin/env bash
# Vikunja Proxmox Deployment - Health Check Functions
# Provides: Component health monitoring and validation
# Required by: vikunja-install.sh, vikunja-update.sh, vikunja-manage.sh

set -euo pipefail

# Prevent multiple sourcing
if [[ -n "${VIKUNJA_HEALTH_CHECK_LIB_LOADED:-}" ]]; then
    return 0
fi
readonly VIKUNJA_HEALTH_CHECK_LIB_LOADED=1

# Common and proxmox-api functions are sourced by main script before this library

# ============================================================================
# Health Check Functions (T030)
# ============================================================================

# Check health of all components
# Usage: check_component_health ct_id backend_port mcp_port
# Returns: 0 if all healthy, 1 if any unhealthy
check_component_health() {
    local ct_id="$1"
    local backend_port="$2"
    local mcp_port="$3"
    
    log_info "Checking component health"
    
    local all_healthy=true
    
    # Check backend
    if ! check_backend_health "$ct_id" "$backend_port"; then
        log_error "Backend health check failed"
        all_healthy=false
    fi
    
    # Check MCP server
    if ! check_mcp_health "$ct_id" "$mcp_port"; then
        log_warn "MCP health check failed (non-critical)"
    fi
    
    # Check frontend (via nginx)
    if ! check_frontend_health "$ct_id"; then
        log_error "Frontend health check failed"
        all_healthy=false
    fi
    
    if [[ "$all_healthy" == "true" ]]; then
        log_success "All components healthy"
        return 0
    fi
    
    return 1
}

# Check backend health
# Usage: check_backend_health ct_id port
# Returns: 0 if healthy, 1 if not
check_backend_health() {
    local ct_id="$1"
    local port="${2:-${BACKEND_BLUE_PORT:-3456}}"
    
    log_debug "Checking backend health on port ${port}"
    
    # Try to connect to health endpoint
    if pct_exec "$ct_id" curl -sf "http://localhost:${port}/health" >/dev/null 2>&1; then
        log_debug "Backend is healthy"
        return 0
    fi
    
    # Fallback: check if port is listening
    if pct_exec "$ct_id" ss -tuln | grep -q ":${port} "; then
        log_debug "Backend port is listening"
        return 0
    fi
    
    log_debug "Backend is not responding"
    return 1
}

# Check MCP server health
# Usage: check_mcp_health ct_id port
# Returns: 0 if healthy, 1 if not
check_mcp_health() {
    local ct_id="$1"
    local port="${2:-${MCP_BLUE_PORT:-3457}}"
    
    log_debug "Checking MCP server health on port ${port}"
    
    # Check if port is listening
    if pct_exec "$ct_id" ss -tuln | grep -q ":${port} "; then
        log_debug "MCP server is listening"
        return 0
    fi
    
    log_debug "MCP server is not responding"
    return 1
}

# Check frontend health
# Usage: check_frontend_health ct_id
# Returns: 0 if healthy, 1 if not
check_frontend_health() {
    local ct_id="$1"
    
    log_debug "Checking frontend health"
    
    # Check if nginx is serving the frontend
    if pct_exec "$ct_id" curl -sf "http://localhost/" >/dev/null 2>&1; then
        log_debug "Frontend is accessible"
        return 0
    fi
    
    log_debug "Frontend is not accessible"
    return 1
}

# Check database connection
# Usage: check_database_connection ct_id db_type db_host db_port db_name db_user db_pass
# Returns: 0 if connected, 1 if not
check_database_connection() {
    local ct_id="$1"
    local db_type="$2"
    local db_host="$3"
    local db_port="$4"
    local db_name="$5"
    local db_user="$6"
    local db_pass="$7"
    
    log_debug "Checking ${db_type} database connection"
    
    case "$db_type" in
        sqlite)
            # For SQLite, just check if file exists
            if pct_exec "$ct_id" test -f "$db_name"; then
                return 0
            fi
            ;;
        postgresql)
            if pct_exec "$ct_id" bash -c \
                "PGPASSWORD='${db_pass}' psql -h ${db_host} -p ${db_port} -U ${db_user} -d ${db_name} -c 'SELECT 1'" \
                >/dev/null 2>&1; then
                return 0
            fi
            ;;
        mysql)
            if pct_exec "$ct_id" bash -c \
                "mysql -h ${db_host} -P ${db_port} -u ${db_user} -p'${db_pass}' ${db_name} -e 'SELECT 1'" \
                >/dev/null 2>&1; then
                return 0
            fi
            ;;
    esac
    
    return 1
}

# ============================================================================
# Health Check Validation Functions (for T053 - User Story 2)
# ============================================================================

# Wait for component to become healthy
# Usage: wait_for_healthy ct_id component port [timeout]
# Returns: 0 if healthy, 1 on timeout
wait_for_healthy() {
    local ct_id="$1"
    local component="$2"
    local port="$3"
    local timeout="${4:-60}"
    local waited=0
    
    log_info "Waiting for ${component} to become healthy (timeout: ${timeout}s)"
    
    while (( waited < timeout )); do
        case "$component" in
            backend)
                if check_backend_health "$ct_id" "$port"; then
                    log_success "${component} is healthy"
                    return 0
                fi
                ;;
            mcp)
                if check_mcp_health "$ct_id" "$port"; then
                    log_success "${component} is healthy"
                    return 0
                fi
                ;;
            frontend)
                if check_frontend_health "$ct_id"; then
                    log_success "${component} is healthy"
                    return 0
                fi
                ;;
        esac
        
        sleep 2
        ((waited += 2))
    done
    
    log_error "${component} failed to become healthy after ${timeout}s"
    return 1
}

# Retry operation with timeout
# Usage: retry_with_timeout command timeout
# Returns: 0 if successful, 1 on timeout
retry_with_timeout() {
    local -a cmd=("$@")
    local timeout="${cmd[-1]}"
    unset 'cmd[-1]'
    local waited=0
    
    while (( waited < timeout )); do
        if "${cmd[@]}"; then
            return 0
        fi
        sleep 2
        ((waited += 2))
    done
    
    return 1
}

# Full health check with detailed output
# Usage: full_health_check ct_id backend_port mcp_port
# Returns: 0 if all healthy, 1 if any issues
full_health_check() {
    local ct_id="$1"
    local backend_port="$2"
    local mcp_port="$3"
    
    echo ""
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║           Component Health Check                          ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    
    local all_ok=true
    
    # Backend
    echo -n "Backend (port ${backend_port}): "
    if check_backend_health "$ct_id" "$backend_port"; then
        echo "${COLOR_GREEN}✓ Healthy${COLOR_RESET}"
    else
        echo "${COLOR_RED}✗ Unhealthy${COLOR_RESET}"
        all_ok=false
    fi
    
    # MCP
    echo -n "MCP Server (port ${mcp_port}): "
    if check_mcp_health "$ct_id" "$mcp_port"; then
        echo "${COLOR_GREEN}✓ Healthy${COLOR_RESET}"
    else
        echo "${COLOR_YELLOW}⚠ Unhealthy${COLOR_RESET}"
    fi
    
    # Frontend
    echo -n "Frontend (nginx): "
    if check_frontend_health "$ct_id"; then
        echo "${COLOR_GREEN}✓ Healthy${COLOR_RESET}"
    else
        echo "${COLOR_RED}✗ Unhealthy${COLOR_RESET}"
        all_ok=false
    fi
    
    echo ""
    
    if [[ "$all_ok" == "true" ]]; then
        return 0
    fi
    
    return 1
}

log_debug "Health check library loaded"
