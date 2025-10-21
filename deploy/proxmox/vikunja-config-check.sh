#!/usr/bin/env bash
# Vikunja Configuration Validator
# Purpose: Detect and validate Vikunja database configuration
# Feature: 004-proxmox-deployment - Configuration troubleshooting tool
#
# Usage: vikunja-config-check.sh [vikunja-binary-path]
#
# Exit Codes:
#   0  - Configuration valid
#   1  - Configuration missing or invalid
#   2  - Invalid arguments

set -euo pipefail

# Script metadata
readonly SCRIPT_VERSION="1.0.0"
readonly SCRIPT_NAME="$(basename "$0")"

# Color codes
readonly COLOR_RED='\033[0;31m'
readonly COLOR_GREEN='\033[0;32m'
readonly COLOR_YELLOW='\033[1;33m'
readonly COLOR_BLUE='\033[0;34m'
readonly COLOR_RESET='\033[0m'
readonly COLOR_BOLD='\033[1m'

# Logging functions
log_info() { echo -e "${COLOR_BLUE}[INFO]${COLOR_RESET} $*"; }
log_success() { echo -e "${COLOR_GREEN}[SUCCESS]${COLOR_RESET} $*"; }
log_warn() { echo -e "${COLOR_YELLOW}[WARN]${COLOR_RESET} $*"; }
log_error() { echo -e "${COLOR_RED}[ERROR]${COLOR_RESET} $*"; }

# Configuration
VIKUNJA_BIN="${1:-/opt/vikunja/vikunja}"

#===============================================================================
# Help Text
#===============================================================================

show_help() {
    cat << EOF
${COLOR_BOLD}Vikunja Configuration Validator${COLOR_RESET}
Detect and validate Vikunja database configuration

${COLOR_BOLD}USAGE:${COLOR_RESET}
    $SCRIPT_NAME [vikunja-binary-path]

${COLOR_BOLD}ARGUMENTS:${COLOR_RESET}
    vikunja-binary-path    Path to vikunja binary (default: /opt/vikunja/vikunja)

${COLOR_BOLD}OPTIONS:${COLOR_RESET}
    -h, --help            Show this help message
    --version             Show version

${COLOR_BOLD}EXAMPLES:${COLOR_RESET}
    # Check default location
    $SCRIPT_NAME
    
    # Check specific binary
    $SCRIPT_NAME /usr/local/bin/vikunja

${COLOR_BOLD}DESCRIPTION:${COLOR_RESET}
    This script validates Vikunja's database configuration by checking:
    - Config file locations (./config.yml, /opt/vikunja/config.yml, /etc/vikunja/config.yml)
    - Environment variables (VIKUNJA_DATABASE_*)
    - Systemd service configuration
    - Database connectivity
    
    It helps diagnose common issues like:
    - "No config file found" warnings
    - Import/export using wrong database
    - Database connection failures

EOF
}

#===============================================================================
# Configuration Detection
#===============================================================================

detect_config_file() {
    local config_locations=(
        "./config.yml"
        "/opt/vikunja/config.yml"
        "/etc/vikunja/config.yml"
        "$HOME/.config/vikunja/config.yml"
    )
    
    log_info "Checking for config files..."
    
    local found=false
    for location in "${config_locations[@]}"; do
        if [[ -f "$location" ]]; then
            log_success "Found config file: $location"
            
            # Parse database type
            if grep -q "^database:" "$location"; then
                local db_type
                db_type=$(grep -A10 "^database:" "$location" | grep -oP '^\s*type:\s*\K\S+' || echo "unknown")
                log_info "  Database type: $db_type"
                
                case "$db_type" in
                    sqlite)
                        local db_path
                        db_path=$(grep -A10 "^database:" "$location" | grep -oP '^\s*path:\s*\K\S+' || echo "not specified")
                        log_info "  Database path: $db_path"
                        ;;
                    postgres*|mysql)
                        local db_host db_name
                        db_host=$(grep -A10 "^database:" "$location" | grep -oP '^\s*host:\s*\K\S+' || echo "not specified")
                        db_name=$(grep -A10 "^database:" "$location" | grep -oP '^\s*database:\s*\K\S+' || echo "not specified")
                        log_info "  Database host: $db_host"
                        log_info "  Database name: $db_name"
                        ;;
                esac
            fi
            found=true
        fi
    done
    
    if [[ "$found" == false ]]; then
        log_warn "No config.yml found in standard locations"
        return 1
    fi
    
    return 0
}

detect_env_vars() {
    log_info "Checking for environment variables..."
    
    local env_vars=(
        "VIKUNJA_DATABASE_TYPE"
        "VIKUNJA_DATABASE_HOST"
        "VIKUNJA_DATABASE_DATABASE"
        "VIKUNJA_DATABASE_USER"
        "VIKUNJA_DATABASE_PATH"
    )
    
    local found=false
    for var in "${env_vars[@]}"; do
        if [[ -n "${!var:-}" ]]; then
            log_success "  $var=${!var}"
            found=true
        fi
    done
    
    if [[ "$found" == false ]]; then
        log_warn "No VIKUNJA_DATABASE_* environment variables set"
        return 1
    fi
    
    return 0
}

detect_systemd_config() {
    log_info "Checking systemd service configuration..."
    
    local service_files=(
        "/etc/systemd/system/vikunja-api-blue.service"
        "/etc/systemd/system/vikunja-api-green.service"
        "/etc/systemd/system/vikunja.service"
    )
    
    local found=false
    for service_file in "${service_files[@]}"; do
        if [[ -f "$service_file" ]]; then
            log_success "Found service file: $service_file"
            
            # Extract database environment variables
            if grep -q "VIKUNJA_DATABASE_TYPE" "$service_file"; then
                local db_type
                db_type=$(grep -oP 'Environment="VIKUNJA_DATABASE_TYPE=\K[^"]+' "$service_file" || echo "unknown")
                log_info "  Database type: $db_type"
                
                case "$db_type" in
                    sqlite)
                        local db_path
                        db_path=$(grep -oP 'Environment="VIKUNJA_DATABASE_PATH=\K[^"]+' "$service_file" || echo "not specified")
                        log_info "  Database path: $db_path"
                        ;;
                    postgres*|mysql)
                        local db_host
                        db_host=$(grep -oP 'Environment="VIKUNJA_DATABASE_HOST=\K[^"]+' "$service_file" || echo "not specified")
                        log_info "  Database host: $db_host"
                        ;;
                esac
                found=true
            fi
        fi
    done
    
    if [[ "$found" == false ]]; then
        log_warn "No systemd service files found or no database config in services"
        return 1
    fi
    
    return 0
}

run_vikunja_health() {
    log_info "Running Vikunja health check..."
    
    if [[ ! -f "$VIKUNJA_BIN" ]]; then
        log_error "Vikunja binary not found: $VIKUNJA_BIN"
        return 1
    fi
    
    if [[ ! -x "$VIKUNJA_BIN" ]]; then
        log_error "Vikunja binary is not executable: $VIKUNJA_BIN"
        return 1
    fi
    
    local health_output
    health_output=$("$VIKUNJA_BIN" health 2>&1 || true)
    
    # Check for warning about missing config
    if echo "$health_output" | grep -q "No config file found"; then
        log_warn "Vikunja reports: 'No config file found, using default or config from environment variables'"
        echo "$health_output" | grep "No config file found"
        return 1
    fi
    
    log_success "Vikunja health check completed"
    return 0
}

#===============================================================================
# Main
#===============================================================================

main() {
    # Parse arguments
    case "${1:-}" in
        -h|--help)
            show_help
            exit 0
            ;;
        --version)
            echo "$SCRIPT_VERSION"
            exit 0
            ;;
    esac
    
    echo ""
    log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_info "Vikunja Configuration Validator"
    log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    local config_found=false
    
    # Check config file
    if detect_config_file; then
        config_found=true
    fi
    
    echo ""
    
    # Check environment variables
    if detect_env_vars; then
        config_found=true
    fi
    
    echo ""
    
    # Check systemd services
    detect_systemd_config && config_found=true
    
    echo ""
    
    # Run health check
    local health_ok=true
    run_vikunja_health || health_ok=false
    
    echo ""
    log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_info "Configuration Summary"
    log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    if [[ "$config_found" == true ]]; then
        log_success "Configuration detected ✓"
        
        if [[ "$health_ok" == false ]]; then
            log_warn "However, Vikunja reports missing config file"
            log_warn "This may indicate environment variables are not set in current shell"
            log_warn "Check systemd service files for proper configuration"
        fi
        
        echo ""
        log_info "Recommendations:"
        log_info "  • For manual operations: Create /opt/vikunja/config.yml"
        log_info "  • For systemd services: Configuration is in service files (OK)"
        log_info "  • For import/export: Use config.yml or set env vars before running"
        echo ""
        
        exit 0
    else
        log_error "No configuration detected! ✗"
        echo ""
        log_error "Vikunja will use default SQLite database in current directory"
        log_error "This may cause import/export operations to use wrong database"
        echo ""
        log_info "Solutions:"
        log_info "  1. Create /opt/vikunja/config.yml (see docs/TROUBLESHOOTING.md)"
        log_info "  2. Set VIKUNJA_DATABASE_* environment variables"
        log_info "  3. Configure systemd service files (automatic with deployment scripts)"
        echo ""
        
        exit 1
    fi
}

main "$@"
