#!/usr/bin/env bash
# Vikunja Update Script - Zero-Downtime Blue-Green Deployment
# Purpose: Update Vikunja deployment with automatic rollback on failures
# Feature: 004-proxmox-deployment - User Story 2
#
# Usage: vikunja-update.sh [OPTIONS] <instance-id>
#
# Exit Codes:
#   0  - Update successful
#   1  - General update error
#   2  - Invalid arguments
#   3  - Instance not found
#   5  - Update already in progress (locked)
#   10 - Health check failed (post-update)
#   11 - Rollback successful (update failed but recovered)
#   12 - Rollback failed (manual intervention needed)
#   20 - No updates available

set -euo pipefail

# Script metadata
readonly SCRIPT_VERSION="1.0.0"
readonly SCRIPT_NAME="$(basename "$0")"

# Resolve script directory (handle symlinks)
SCRIPT_SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SCRIPT_SOURCE" ]; do
    SCRIPT_DIR="$(cd -P "$(dirname "$SCRIPT_SOURCE")" && pwd)"
    SCRIPT_SOURCE="$(readlink "$SCRIPT_SOURCE")"
    [[ $SCRIPT_SOURCE != /* ]] && SCRIPT_SOURCE="$SCRIPT_DIR/$SCRIPT_SOURCE"
done
readonly SCRIPT_DIR="$(cd -P "$(dirname "$SCRIPT_SOURCE")" && pwd)"

# Source library functions
# shellcheck source=lib/common.sh
source "${SCRIPT_DIR}/lib/common.sh"
# shellcheck source=lib/proxmox-api.sh
source "${SCRIPT_DIR}/lib/proxmox-api.sh"
# shellcheck source=lib/lxc-setup.sh
source "${SCRIPT_DIR}/lib/lxc-setup.sh"
# shellcheck source=lib/blue-green.sh
source "${SCRIPT_DIR}/lib/blue-green.sh"
# shellcheck source=lib/backup-restore.sh
source "${SCRIPT_DIR}/lib/backup-restore.sh"
# shellcheck source=lib/nginx-setup.sh
source "${SCRIPT_DIR}/lib/nginx-setup.sh"
# shellcheck source=lib/health-check.sh
source "${SCRIPT_DIR}/lib/health-check.sh"
# shellcheck source=lib/service-setup.sh
source "${SCRIPT_DIR}/lib/service-setup.sh"

#===============================================================================
# Self-Update Function
#===============================================================================

# Update deployment scripts from GitHub
# Returns: 0 if updated/current, 1 if failed
update_deployment_scripts() {
    local github_owner="${VIKUNJA_GITHUB_OWNER:-aroige}"
    local github_repo="${VIKUNJA_GITHUB_REPO:-vikunja}"
    local github_branch="${VIKUNJA_GITHUB_BRANCH:-main}"
    local base_url="https://raw.githubusercontent.com/${github_owner}/${github_repo}/${github_branch}/deploy/proxmox"
    local deploy_dir="/opt/vikunja-deploy"
    local temp_dir="/tmp/vikunja-deploy-update-$$"
    
    log_debug "Checking for deployment script updates..."
    log_debug "Source: ${github_owner}/${github_repo}@${github_branch}"
    
    # Create temp directory
    mkdir -p "${temp_dir}/lib" "${temp_dir}/templates"
    
    # Download current version of update script
    if ! curl -fsSL "${base_url}/vikunja-update.sh" -o "${temp_dir}/vikunja-update.sh"; then
        log_debug "Failed to download update script"
        rm -rf "$temp_dir"
        return 1
    fi
    
    # Check if there are differences
    if [[ -f "${deploy_dir}/vikunja-update.sh" ]]; then
        if diff -q "${deploy_dir}/vikunja-update.sh" "${temp_dir}/vikunja-update.sh" >/dev/null 2>&1; then
            log_debug "Deployment scripts are current"
            rm -rf "$temp_dir"
            return 0
        fi
    fi
    
    log_info "Downloading latest deployment scripts..."
    
    # Download all library files
    local lib_files=(
        "lib/common.sh"
        "lib/proxmox-api.sh"
        "lib/lxc-setup.sh"
        "lib/service-setup.sh"
        "lib/nginx-setup.sh"
        "lib/health-check.sh"
        "lib/blue-green.sh"
        "lib/backup-restore.sh"
    )
    
    for file in "${lib_files[@]}"; do
        if ! curl -fsSL "${base_url}/${file}" -o "${temp_dir}/${file}"; then
            log_warn "Failed to download ${file}"
        fi
    done
    
    # Download templates
    local template_files=(
        "templates/deployment-config.yaml"
        "templates/vikunja-backend.service"
        "templates/vikunja-mcp.service"
        "templates/nginx-vikunja.conf"
        "templates/health-check.sh"
    )
    
    for file in "${template_files[@]}"; do
        if ! curl -fsSL "${base_url}/${file}" -o "${temp_dir}/${file}"; then
            log_debug "Failed to download ${file} (may not exist)"
        fi
    done
    
    # Backup current scripts
    if [[ -d "${deploy_dir}" ]]; then
        cp -r "${deploy_dir}" "${deploy_dir}.backup-$(date +%s)"
    fi
    
    # Install new scripts
    cp -r "${temp_dir}"/* "${deploy_dir}/"
    chmod +x "${deploy_dir}/vikunja-update.sh"
    chmod +x "${deploy_dir}"/lib/*.sh 2>/dev/null || true
    
    rm -rf "$temp_dir"
    
    log_success "Deployment scripts updated"
    return 0
}

#===============================================================================
# Configuration
#===============================================================================

# Defaults
GIT_BRANCH="main"
GIT_COMMIT=""
FORCE_UPDATE=false
SKIP_BACKUP=false
SKIP_MIGRATIONS=false
ROLLBACK_ON_FAILURE=true
NO_HEALTH_CHECK=false
INSTANCE_ID=""

# Repo configuration
REPO_PATH="/opt/vikunja"

#===============================================================================
# Help Text
#===============================================================================

show_help() {
    cat << EOF
${COLOR_BOLD:-}Vikunja Update Script${COLOR_RESET:-}
Zero-downtime updates using blue-green deployment

${COLOR_BOLD:-}USAGE:${COLOR_RESET:-}
    $SCRIPT_NAME [OPTIONS] [instance-id]

${COLOR_BOLD:-}ARGUMENTS:${COLOR_RESET:-}
    <instance-id>           Instance to update (optional when running in container)

${COLOR_BOLD:-}OPTIONS:${COLOR_RESET:-}
    --git-branch <BRANCH>   Git branch to update from (default: main)
    --git-commit <HASH>     Specific commit to deploy (default: latest)
    --force                 Force update even if no new commits
    --skip-backup           Skip pre-update backup (NOT RECOMMENDED)
    --skip-migrations       Skip database migrations (dangerous)
    --rollback-on-failure   Automatically rollback on any failure (default: true)
    --no-health-check       Skip health checks (dangerous)
    
    -h, --help              Show this help message
    -v, --verbose           Verbose output
    -d, --debug             Debug output
    --version               Show version

${COLOR_BOLD:-}EXAMPLES:${COLOR_RESET:-}
    # When running inside container (no instance-id needed)
    vikunja-update

    # Update to specific branch
    vikunja-update --git-branch develop
    
    # Update to specific commit
    vikunja-update --git-commit abc123def
    
    # Force update even if no changes
    vikunja-update --force

    # When running from Proxmox host (instance-id required)
    vikunja-update vikunja-main

${COLOR_BOLD:-}EXIT CODES:${COLOR_RESET:-}
    0   - Update successful
    1   - General error
    2   - Invalid arguments
    3   - Instance not found
    5   - Update already in progress (locked)
    10  - Health check failed
    11  - Rollback successful (update failed but recovered)
    12  - Rollback failed (manual intervention needed)
    20  - No updates available

${COLOR_BOLD:-}DOCUMENTATION:${COLOR_RESET:-}
    See: deploy/proxmox/docs/README.md

EOF
}

#===============================================================================
# Argument Parsing
#===============================================================================

parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            -h|--help)
                show_help
                exit 0
                ;;
            --version)
                echo "$SCRIPT_VERSION"
                exit 0
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -d|--debug)
                DEBUG=true
                shift
                ;;
            --git-branch)
                GIT_BRANCH="$2"
                shift 2
                ;;
            --git-commit)
                GIT_COMMIT="$2"
                shift 2
                ;;
            --force)
                FORCE_UPDATE=true
                shift
                ;;
            --skip-backup)
                SKIP_BACKUP=true
                log_warn "Skipping backup - recovery will not be possible!"
                shift
                ;;
            --skip-migrations)
                SKIP_MIGRATIONS=true
                log_warn "Skipping migrations - may cause database inconsistencies!"
                shift
                ;;
            --no-health-check)
                NO_HEALTH_CHECK=true
                log_warn "Skipping health checks - may deploy broken code!"
                shift
                ;;
            -*)
                log_error "Unknown option: $1"
                show_help
                exit 2
                ;;
            *)
                if [[ -z "$INSTANCE_ID" ]]; then
                    INSTANCE_ID="$1"
                else
                    log_error "Multiple instance IDs provided: $INSTANCE_ID and $1"
                    exit 2
                fi
                shift
                ;;
        esac
    done
    
    # Instance ID validation will happen in main() after detecting environment
    # (required on host, optional in container)
}

#===============================================================================
# Main Update Orchestration
#===============================================================================

main() {
    local update_start
    local update_end
    local update_duration
    local downtime_start
    local downtime_end
    local downtime_seconds
    
    update_start=$(date +%s)
    
    log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_info "Vikunja Update"
    log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    #---------------------------------------------------------------------------
    # Step 0: Detect execution environment and self-update deployment scripts
    #---------------------------------------------------------------------------
    
    log_info "[0/9] Detecting execution environment..."
    
    # Detect if running inside container or on Proxmox host
    local running_in_container=false
    
    # Check multiple indicators for LXC container
    if [[ -f "/.dockerenv" ]] || \
       grep -qa "lxc" /proc/1/cgroup 2>/dev/null || \
       grep -qa "container=lxc" /proc/1/environ 2>/dev/null || \
       [[ -d /opt/vikunja ]] && [[ -d /opt/vikunja-deploy ]] && ! command -v pct &>/dev/null; then
        running_in_container=true
        log_info "Running inside LXC container"
        
        # When in container, we can self-update deployment scripts
        log_info "Checking for deployment script updates..."
        
        if update_deployment_scripts; then
            log_success "Deployment scripts are up to date"
        else
            log_warn "Failed to update deployment scripts, continuing with current version"
        fi
        
        # Instance ID defaults to container hostname if not provided
        if [[ -z "$INSTANCE_ID" ]]; then
            INSTANCE_ID=$(hostname)
            log_info "Using instance ID: $INSTANCE_ID"
        fi
    else
        running_in_container=false
        log_info "Running on Proxmox host"
        
        # Verify we have pct command
        if ! command -v pct &>/dev/null; then
            log_error "Not running on Proxmox VE and not in container"
            log_error "This script must run either:"
            log_error "  1. Inside the Vikunja LXC container, or"
            log_error "  2. On the Proxmox VE host"
            exit 1
        fi
        
        # Instance ID is required when running from host
        if [[ -z "$INSTANCE_ID" ]]; then
            log_error "Instance ID is required when running from Proxmox host"
            log_error "Usage: $SCRIPT_NAME <instance-id>"
            exit 2
        fi
    fi
    
    # Check if running as root
    if [[ $EUID -ne 0 ]]; then
        log_error "Must run as root"
        exit 1
    fi
    
    #---------------------------------------------------------------------------
    # Step 1: Pre-flight Checks
    #---------------------------------------------------------------------------
    
    log_info "[1/9] Pre-flight checks..."
    
    # Load configuration based on execution environment
    local container_id
    if [[ "$running_in_container" == true ]]; then
        # When in container, container_id is self
        container_id=$(hostname | grep -oE '[0-9]+$' || echo "self")
        # Config is local
        local config_file="/etc/vikunja/deploy-config.yaml"
    else
        # When on host, load from host config
        local config_file="/etc/vikunja-deploy/${INSTANCE_ID}/config.yaml"
    fi
    
    # Check if instance exists
    local config_file="/etc/vikunja/${INSTANCE_ID}.yaml"
    if [[ ! -f "$config_file" ]]; then
        log_error "Instance not found: $INSTANCE_ID"
        log_error "Config file missing: $config_file"
        exit 3
    fi
    
    # Load instance configuration
    local container_id
    container_id=$(grep "container_id:" "$config_file" | awk '{print $2}' | tr -d '"')
    
    if [[ -z "$container_id" ]]; then
        log_error "Container ID not found in config: $config_file"
        exit 3
    fi
    
    # Verify container is running
    if ! pct status "$container_id" | grep -q "running"; then
        log_error "Container $container_id is not running"
        exit 1
    fi
    
    # Acquire lock
    if ! acquire_lock "$INSTANCE_ID"; then
        log_error "Update already in progress for $INSTANCE_ID"
        exit 5
    fi
    
    # Ensure lock is released on exit
    trap 'release_lock "$INSTANCE_ID"' EXIT
    
    # Load additional configuration
    local db_type
    local db_config
    db_type=$(grep "database_type:" "$config_file" | awk '{print $2}' | tr -d '"')
    
    case "$db_type" in
        sqlite)
            db_config=$(grep "database_path:" "$config_file" | awk '{print $2}' | tr -d '"')
            ;;
        postgresql|postgres|mysql)
            local db_host db_port db_name db_user db_password
            db_host=$(grep "database_host:" "$config_file" | awk '{print $2}' | tr -d '"')
            db_port=$(grep "database_port:" "$config_file" | awk '{print $2}' | tr -d '"')
            db_name=$(grep "database_name:" "$config_file" | awk '{print $2}' | tr -d '"')
            db_user=$(grep "database_user:" "$config_file" | awk '{print $2}' | tr -d '"')
            db_password=$(grep "database_password:" "$config_file" | awk '{print $2}' | tr -d '"')
            db_config="${db_host}:${db_port}:${db_name}:${db_user}:${db_password}"
            ;;
    esac
    
    log_success "Pre-flight checks passed"
    
    #---------------------------------------------------------------------------
    # Step 2: Check for Updates
    #---------------------------------------------------------------------------
    
    log_info "[2/8] Checking for updates..."
    
    # Get current version
    local current_version
    current_version=$(get_commit_hash "$container_id" "$REPO_PATH" || echo "unknown")
    log_info "Current version: $current_version"
    
    # Fetch latest changes
    if ! check_for_updates "$container_id" "$REPO_PATH"; then
        log_error "Failed to check for updates"
        exit 1
    fi
    
    # Get latest version
    local latest_version
    if [[ -n "$GIT_COMMIT" ]]; then
        latest_version="$GIT_COMMIT"
    else
        latest_version=$(pct_exec "$container_id" "git -C $REPO_PATH rev-parse origin/$GIT_BRANCH" || echo "unknown")
    fi
    
    log_info "Latest version: $latest_version"
    
    # Check if update is needed
    if [[ "$current_version" == "$latest_version" && "$FORCE_UPDATE" == false ]]; then
        log_info "Already at latest version - no update needed"
        exit 20
    fi
    
    # Show changes
    local commit_count
    commit_count=$(pct_exec "$container_id" "git -C $REPO_PATH rev-list ${current_version}..${latest_version} --count 2>/dev/null" || echo "0")
    log_info "Changes: $commit_count commits"
    
    log_success "Update available"
    
    #---------------------------------------------------------------------------
    # Step 3: Pre-Update Backup
    #---------------------------------------------------------------------------
    
    local backup_file=""
    
    if [[ "$SKIP_BACKUP" == false ]]; then
        log_info "[3/8] Creating pre-update backup..."
        
        backup_file=$(create_pre_migration_backup "$INSTANCE_ID" "$container_id" "$db_type" "$db_config")
        
        if [[ -z "$backup_file" ]]; then
            log_error "Backup creation failed"
            exit 1
        fi
        
        # Verify backup integrity
        if ! verify_backup_integrity "$container_id" "$backup_file" "$db_type"; then
            log_error "Backup integrity check failed"
            exit 1
        fi
        
        log_success "Backup created: $backup_file"
    else
        log_warn "[3/8] Skipping backup (--skip-backup flag)"
    fi
    
    #---------------------------------------------------------------------------
    # Step 4: Determine Blue-Green Color & Pull Updates
    #---------------------------------------------------------------------------
    
    log_info "[4/8] Preparing update..."
    
    # Get current active color
    local active_color
    active_color=$(get_active_color "$INSTANCE_ID")
    log_info "Current active color: $active_color"
    
    # Determine inactive color for deployment
    local target_color
    target_color=$(determine_inactive_color "$INSTANCE_ID" "$container_id")
    log_info "Deploying to color: $target_color"
    
    # Pull latest code
    if ! pull_latest "$container_id" "$REPO_PATH"; then
        log_error "Failed to pull latest code"
        exit 1
    fi
    
    # Checkout specific commit if specified
    if [[ -n "$GIT_COMMIT" ]]; then
        if ! checkout_commit "$container_id" "$REPO_PATH" "$GIT_COMMIT"; then
            log_error "Failed to checkout commit $GIT_COMMIT"
            exit 1
        fi
    fi
    
    log_success "Code updated to $latest_version"
    
    #---------------------------------------------------------------------------
    # Step 5: Build New Version
    #---------------------------------------------------------------------------
    
    log_info "[5/8] Building new version..."
    
    # Build backend
    if ! build_backend "$container_id" "$REPO_PATH" "$target_color"; then
        log_error "Backend build failed"
        exit 1
    fi
    
    # Build frontend (shared, not color-specific)
    if ! build_frontend "$container_id" "$REPO_PATH"; then
        log_error "Frontend build failed"
        exit 1
    fi
    
    # Build MCP server
    if ! build_mcp "$container_id" "$REPO_PATH" "$target_color"; then
        log_error "MCP server build failed"
        exit 1
    fi
    
    log_success "Build completed"
    
    #---------------------------------------------------------------------------
    # Step 6: Run Migrations
    #---------------------------------------------------------------------------
    
    if [[ "$SKIP_MIGRATIONS" == false ]]; then
        log_info "[6/8] Running database migrations..."
        
        local pre_migration_count
        pre_migration_count=$(check_migration_status "$container_id" "$db_type" "$db_config" || echo "0")
        log_debug "Pre-migration count: $pre_migration_count"
        
        if ! run_migrations "$container_id" "$db_type" "$db_config"; then
            log_error "Migration failed - initiating rollback"
            
            # Restore from backup
            if [[ -n "$backup_file" ]] && [[ "$SKIP_BACKUP" == false ]]; then
                log_warn "Restoring database from backup..."
                restore_from_backup "$container_id" "$backup_file" "$db_type" "$db_config" || log_error "Backup restoration failed"
            fi
            
            exit 1
        fi
        
        local post_migration_count
        post_migration_count=$(check_migration_status "$container_id" "$db_type" "$db_config" || echo "0")
        local new_migrations=$((post_migration_count - pre_migration_count))
        
        if [[ $new_migrations -gt 0 ]]; then
            log_success "Applied $new_migrations new migration(s)"
        else
            log_info "No new migrations to apply"
        fi
    else
        log_warn "[6/8] Skipping migrations (--skip-migrations flag)"
    fi
    
    #---------------------------------------------------------------------------
    # Step 7: Start Services on Target Color & Health Check
    #---------------------------------------------------------------------------
    
    log_info "[7/8] Starting services on $target_color..."
    
    # Start services on target color
    if ! start_services_on_color "$INSTANCE_ID" "$container_id" "$target_color"; then
        log_error "Failed to start services on $target_color"
        perform_rollback "$INSTANCE_ID" "$container_id" "$active_color" "$backup_file" "$db_type" "$db_config"
        exit 11
    fi
    
    # Wait for services to be ready
    sleep 5
    
    # Perform health check on target color
    if [[ "$NO_HEALTH_CHECK" == false ]]; then
        log_info "Running health checks on $target_color..."
        
        if ! check_component_health "$container_id" "$target_color"; then
            log_error "Health check failed on $target_color"
            perform_rollback "$INSTANCE_ID" "$container_id" "$active_color" "$backup_file" "$db_type" "$db_config"
            exit 11
        fi
        
        log_success "Health checks passed"
    else
        log_warn "Skipping health checks (--no-health-check flag)"
    fi
    
    #---------------------------------------------------------------------------
    # Step 8: Switch Traffic (Zero-Downtime)
    #---------------------------------------------------------------------------
    
    log_info "[8/8] Switching traffic to $target_color..."
    
    downtime_start=$(date +%s%3N)  # Milliseconds
    
    # Switch nginx upstream to target color
    if ! switch_traffic "$INSTANCE_ID" "$container_id" "$target_color"; then
        log_error "Failed to switch traffic"
        perform_rollback "$INSTANCE_ID" "$container_id" "$active_color" "$backup_file" "$db_type" "$db_config"
        exit 11
    fi
    
    downtime_end=$(date +%s%3N)
    downtime_seconds=$(((downtime_end - downtime_start) / 1000))
    
    log_success "Traffic switched to $target_color"
    
    # Stop old services
    log_info "Stopping old services on $active_color..."
    stop_services_on_color "$INSTANCE_ID" "$container_id" "$active_color" || log_warn "Failed to stop all old services"
    
    # Cleanup old backups
    cleanup_old_backups "$INSTANCE_ID" "$container_id" 5 || log_warn "Backup cleanup failed"
    
    #---------------------------------------------------------------------------
    # Update Summary
    #---------------------------------------------------------------------------
    
    update_end=$(date +%s)
    update_duration=$((update_end - update_start))
    
    log_success "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_success "UPDATE COMPLETED SUCCESSFULLY"
    log_success "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_info "Instance:         $INSTANCE_ID"
    log_info "Previous version: $current_version"
    log_info "New version:      $latest_version"
    log_info "Active color:     $target_color"
    log_info "Commits applied:  $commit_count"
    log_info "Update time:      ${update_duration}s"
    log_info "Downtime:         ${downtime_seconds}s"
    log_info ""
    log_info "Rollback available: vikunja-update.sh --rollback $INSTANCE_ID"
    log_success "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    exit 0
}

#===============================================================================
# Rollback Helper
#===============================================================================

perform_rollback() {
    local instance_id="$1"
    local container_id="$2"
    local rollback_color="$3"
    local backup_file="$4"
    local db_type="$5"
    local db_config="$6"
    
    if [[ "$ROLLBACK_ON_FAILURE" == false ]]; then
        log_warn "Automatic rollback disabled - manual intervention required"
        return 1
    fi
    
    log_warn "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_warn "INITIATING AUTOMATIC ROLLBACK"
    log_warn "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # Rollback to previous color
    if ! rollback_to_color "$instance_id" "$container_id" "$rollback_color"; then
        log_error "Rollback failed - manual intervention required"
        exit 12
    fi
    
    # Restore database if backup exists
    if [[ -n "$backup_file" ]] && [[ "$SKIP_BACKUP" == false ]]; then
        log_info "Restoring database from backup..."
        if ! restore_from_backup "$container_id" "$backup_file" "$db_type" "$db_config"; then
            log_error "Database restore failed - manual intervention required"
            exit 12
        fi
        log_success "Database restored"
    fi
    
    log_success "Rollback completed - system restored to previous state"
}

#===============================================================================
# Script Entry Point
#===============================================================================

parse_arguments "$@"
main
