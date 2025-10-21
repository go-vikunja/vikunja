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
# Configuration Detection and Validation (for T053 - User Story 2)
# ============================================================================

# Detect Vikunja configuration and validate database settings
# Usage: detect_vikunja_configuration ct_id [vikunja_binary_path]
# Returns: 0 if properly configured, 1 if missing/invalid config
# Outputs: JSON with detected configuration or error details
detect_vikunja_configuration() {
    local ct_id="$1"
    local vikunja_bin="${2:-/opt/vikunja/vikunja}"
    
    log_info "Detecting Vikunja configuration..."
    
    # Run a simple healthcheck command and capture output
    local health_output
    health_output=$(pct_exec "$ct_id" "$vikunja_bin" health 2>&1 || true)
    
    # Check for "No config file found" warning
    if echo "$health_output" | grep -q "No config file found"; then
        log_warn "Vikunja is using default/environment configuration (no config.yml found)"
        
        # Check if environment variables are set
        local has_env_config=false
        if pct_exec "$ct_id" bash -c '[[ -n "${VIKUNJA_DATABASE_TYPE:-}" ]]' 2>/dev/null; then
            has_env_config=true
            log_info "Found VIKUNJA_DATABASE_* environment variables"
        fi
        
        if [[ "$has_env_config" == false ]]; then
            log_error "No config file AND no environment variables detected"
            log_error "Vikunja will use default SQLite database!"
            return 1
        fi
    fi
    
    # Try to determine actual database configuration
    local db_type=""
    local db_info=""
    
    # Check systemd service for environment variables
    if pct_exec "$ct_id" test -f /etc/systemd/system/vikunja-api-blue.service 2>/dev/null; then
        db_type=$(pct_exec "$ct_id" grep -oP 'Environment="VIKUNJA_DATABASE_TYPE=\K[^"]+' /etc/systemd/system/vikunja-api-blue.service 2>/dev/null || echo "")
    elif pct_exec "$ct_id" test -f /etc/systemd/system/vikunja-api-green.service 2>/dev/null; then
        db_type=$(pct_exec "$ct_id" grep -oP 'Environment="VIKUNJA_DATABASE_TYPE=\K[^"]+' /etc/systemd/system/vikunja-api-green.service 2>/dev/null || echo "")
    fi
    
    # Check config file if it exists
    if pct_exec "$ct_id" test -f /opt/vikunja/config.yml 2>/dev/null; then
        db_type=$(pct_exec "$ct_id" grep -A5 '^database:' /opt/vikunja/config.yml | grep -oP 'type:\s*\K\S+' 2>/dev/null || echo "$db_type")
        log_info "Found config.yml with database type: ${db_type:-unknown}"
    elif pct_exec "$ct_id" test -f /etc/vikunja/config.yml 2>/dev/null; then
        db_type=$(pct_exec "$ct_id" grep -A5 '^database:' /etc/vikunja/config.yml | grep -oP 'type:\s*\K\S+' 2>/dev/null || echo "$db_type")
        log_info "Found /etc/vikunja/config.yml with database type: ${db_type:-unknown}"
    fi
    
    # Validate detected configuration
    if [[ -z "$db_type" ]]; then
        log_error "Could not detect database type"
        return 1
    fi
    
    log_success "Detected database type: $db_type"
    
    # For PostgreSQL/MySQL, verify we can extract connection details
    if [[ "$db_type" =~ ^(postgres|postgresql|mysql)$ ]]; then
        log_debug "Validating ${db_type} connection details..."
        # This will be validated by check_database_connection later
    fi
    
    return 0
}

# Prompt user for missing database configuration
# Usage: prompt_database_configuration
# Returns: 0 if provided, 1 if cancelled
# Outputs configuration to stdout as KEY=VALUE pairs
prompt_database_configuration() {
    log_warn "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_warn "DATABASE CONFIGURATION REQUIRED"
    log_warn "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_info ""
    log_info "Vikunja requires database configuration. This can be provided via:"
    log_info "  1. Config file (/opt/vikunja/config.yml or /etc/vikunja/config.yml)"
    log_info "  2. Environment variables (VIKUNJA_DATABASE_*)"
    log_info ""
    log_info "The deployment scripts set environment variables in systemd services."
    log_info "This prompt will help you configure the database settings."
    log_info ""
    
    # Database type
    echo -n "Database type (sqlite/postgresql/mysql) [postgresql]: " >&2
    read -r db_type
    db_type="${db_type:-postgresql}"
    
    case "$db_type" in
        sqlite)
            echo -n "SQLite database path [/opt/vikunja/vikunja.db]: " >&2
            read -r db_path
            db_path="${db_path:-/opt/vikunja/vikunja.db}"
            echo "DATABASE_TYPE=sqlite"
            echo "DATABASE_PATH=$db_path"
            return 0
            ;;
        postgresql|postgres)
            db_type="postgres"
            ;;
        mysql)
            ;;
        *)
            log_error "Invalid database type: $db_type"
            return 1
            ;;
    esac
    
    # PostgreSQL/MySQL configuration
    echo -n "Database host [localhost]: " >&2
    read -r db_host
    db_host="${db_host:-localhost}"
    
    echo -n "Database port [5432 for PostgreSQL, 3306 for MySQL]: " >&2
    read -r db_port
    if [[ -z "$db_port" ]]; then
        db_port=$([[ "$db_type" == "postgres" ]] && echo "5432" || echo "3306")
    fi
    
    echo -n "Database name [vikunja]: " >&2
    read -r db_name
    db_name="${db_name:-vikunja}"
    
    echo -n "Database user [vikunja]: " >&2
    read -r db_user
    db_user="${db_user:-vikunja}"
    
    echo -n "Database password: " >&2
    read -rs db_password
    echo "" >&2
    
    if [[ -z "$db_password" ]]; then
        log_error "Database password is required"
        return 1
    fi
    
    echo "DATABASE_TYPE=$db_type"
    echo "DATABASE_HOST=$db_host"
    echo "DATABASE_PORT=$db_port"
    echo "DATABASE_NAME=$db_name"
    echo "DATABASE_USER=$db_user"
    echo "DATABASE_PASSWORD=$db_password"
    
    return 0
}

# Validate import/export configuration before operation
# Usage: validate_operation_config ct_id operation [config_file]
# Returns: 0 if valid, 1 if invalid, prompts user if missing
validate_operation_config() {
    local ct_id="$1"
    local operation="$2"  # "import" or "export"
    local config_file="${3:-}"
    
    log_info "Validating configuration for ${operation} operation..."
    
    # First, try to detect existing configuration
    if detect_vikunja_configuration "$ct_id"; then
        log_success "Configuration validated"
        return 0
    fi
    
    # Configuration is missing or invalid
    log_warn "Configuration missing or invalid for ${operation} operation"
    log_info ""
    log_info "For ${operation} operations, Vikunja needs to know which database to use."
    log_info "Without proper configuration, it will default to SQLite in the current directory."
    log_info ""
    
    # Prompt user
    echo -n "Would you like to configure the database now? (y/N): " >&2
    read -r response
    
    if [[ ! "$response" =~ ^[Yy] ]]; then
        log_warn "Skipping configuration - operation may fail or use wrong database!"
        return 1
    fi
    
    # Get configuration from user
    local config_values
    if ! config_values=$(prompt_database_configuration); then
        log_error "Configuration cancelled or invalid"
        return 1
    fi
    
    # Save configuration to file or display for manual setup
    log_info ""
    log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_info "Configuration collected. Choose setup method:"
    log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_info ""
    log_info "1. Create config.yml file (recommended for manual operations)"
    log_info "2. Set environment variables (recommended for service deployment)"
    log_info "3. Show configuration only (manual setup)"
    echo -n "Select option (1-3) [1]: " >&2
    read -r setup_choice
    setup_choice="${setup_choice:-1}"
    
    case "$setup_choice" in
        1)
            create_config_file "$ct_id" "$config_values"
            ;;
        2)
            show_env_export_commands "$config_values"
            ;;
        3)
            show_config_values "$config_values"
            ;;
        *)
            log_error "Invalid choice"
            return 1
            ;;
    esac
    
    return 0
}

# Helper: Create config.yml from configuration values
create_config_file() {
    local ct_id="$1"
    local config_values="$2"
    local config_path="/opt/vikunja/config.yml"
    
    log_info "Creating config.yml at $config_path..."
    
    # Parse configuration values
    local db_type db_host db_port db_name db_user db_password db_path
    eval "$(echo "$config_values" | sed 's/^/local /')"
    
    # Generate config file content
    local config_content="service:
  timezone: UTC

database:"
    
    if [[ "$DATABASE_TYPE" == "sqlite" ]]; then
        config_content+="
  type: sqlite
  path: $DATABASE_PATH"
    else
        config_content+="
  type: $DATABASE_TYPE
  host: $DATABASE_HOST
  database: $DATABASE_NAME
  user: $DATABASE_USER
  password: \"$DATABASE_PASSWORD\""
        
        if [[ -n "${DATABASE_PORT:-}" ]]; then
            config_content+="
  port: $DATABASE_PORT"
        fi
    fi
    
    # Write config file
    pct_exec "$ct_id" bash -c "cat > $config_path" <<< "$config_content"
    
    log_success "Config file created at $config_path"
    log_info "You can now run Vikunja commands without environment variables"
}

# Helper: Show environment variable export commands
show_env_export_commands() {
    local config_values="$1"
    
    log_info ""
    log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_info "Environment Variable Configuration"
    log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_info ""
    log_info "Run these commands in your shell before executing Vikunja:"
    log_info ""
    
    while IFS='=' read -r key value; do
        echo "export VIKUNJA_${key}=\"${value}\""
    done <<< "$config_values"
    
    log_info ""
    log_info "Or add them to systemd service files in the [Service] section:"
    log_info ""
    
    while IFS='=' read -r key value; do
        echo "Environment=\"VIKUNJA_${key}=${value}\""
    done <<< "$config_values"
}

# Helper: Show raw configuration values
show_config_values() {
    local config_values="$1"
    
    log_info ""
    log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_info "Database Configuration"
    log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_info ""
    echo "$config_values"
    log_info ""
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
