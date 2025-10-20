#!/usr/bin/env bash
# Backup and Restore Library
# Purpose: Database and file backup/restore operations
# Feature: 004-proxmox-deployment - User Story 2 & 5

readonly BACKUP_BASE_DIR="/var/backups/vikunja"
readonly MAX_BACKUPS=5  # Keep last 5 backups per instance

#===============================================================================
# create_pre_migration_backup
#
# Creates a backup before running database migrations
#
# Arguments:
#   $1 - Instance ID
#   $2 - Container ID
#   $3 - Database type (sqlite|postgresql|mysql)
#   $4 - Database configuration (path or connection details)
#
# Returns:
#   0 on success
#   1 on error
#
# Output:
#   Prints backup file path to stdout
#===============================================================================
create_pre_migration_backup() {
    local instance_id="$1"
    local container_id="$2"
    local db_type="$3"
    local db_config="$4"
    
    if [[ -z "$instance_id" || -z "$container_id" || -z "$db_type" ]]; then
        log_error "create_pre_migration_backup: Missing required arguments"
        return 1
    fi
    
    log_info "Creating pre-migration backup..."
    
    # Create backup directory in container
    local backup_dir="$BACKUP_BASE_DIR"
    pct_exec "$container_id" "mkdir -p $backup_dir" || {
        log_error "Failed to create backup directory"
        return 1
    }
    
    # Generate timestamped backup filename
    local timestamp
    timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="${backup_dir}/${instance_id}_${timestamp}"
    
    case "$db_type" in
        sqlite)
            backup_file="${backup_file}.db"
            if ! backup_sqlite "$container_id" "$db_config" "$backup_file"; then
                return 1
            fi
            ;;
        postgresql|postgres)
            backup_file="${backup_file}.pgdump"
            if ! backup_postgresql "$container_id" "$db_config" "$backup_file"; then
                return 1
            fi
            ;;
        mysql)
            backup_file="${backup_file}.sql"
            if ! backup_mysql "$container_id" "$db_config" "$backup_file"; then
                return 1
            fi
            ;;
        *)
            log_error "Unsupported database type: $db_type"
            return 1
            ;;
    esac
    
    # Verify backup was created
    if ! pct_exec "$container_id" "test -f $backup_file"; then
        log_error "Backup file not created: $backup_file"
        return 1
    fi
    
    # Get backup size
    local backup_size
    backup_size=$(pct_exec "$container_id" "du -h $backup_file" | awk '{print $1}')
    
    log_success "Backup created: $backup_file ($backup_size)"
    
    # Output backup path for caller
    echo "$backup_file"
    return 0
}

#===============================================================================
# backup_sqlite
#
# Creates a SQLite database backup
#
# Arguments:
#   $1 - Container ID
#   $2 - Database path
#   $3 - Backup file path
#
# Returns:
#   0 on success
#   1 on error
#===============================================================================
backup_sqlite() {
    local container_id="$1"
    local db_path="$2"
    local backup_file="$3"
    
    log_debug "Backing up SQLite database: $db_path"
    
    # Verify database exists
    if ! pct_exec "$container_id" "test -f $db_path"; then
        log_error "Database not found: $db_path"
        return 1
    fi
    
    # Create backup using cp (SQLite files can be copied directly if not in WAL mode)
    # Use sqlite3 .backup command for consistency
    if ! pct_exec "$container_id" "sqlite3 $db_path \".backup '$backup_file'\""; then
        log_error "Failed to backup SQLite database"
        return 1
    fi
    
    log_debug "SQLite backup complete: $backup_file"
    return 0
}

#===============================================================================
# backup_postgresql
#
# Creates a PostgreSQL database backup
#
# Arguments:
#   $1 - Container ID
#   $2 - Database configuration (format: "host:port:database:user:password")
#   $3 - Backup file path
#
# Returns:
#   0 on success
#   1 on error
#===============================================================================
backup_postgresql() {
    local container_id="$1"
    local db_config="$2"
    local backup_file="$3"
    
    log_debug "Backing up PostgreSQL database"
    
    # Parse connection details
    IFS=':' read -r db_host db_port db_name db_user db_password <<< "$db_config"
    
    # Create pg_dump command with custom format (compressed)
    local pgdump_cmd="PGPASSWORD='$db_password' pg_dump -h $db_host -p $db_port -U $db_user -Fc -f $backup_file $db_name"
    
    if ! pct_exec "$container_id" "$pgdump_cmd"; then
        log_error "Failed to backup PostgreSQL database"
        return 1
    fi
    
    log_debug "PostgreSQL backup complete: $backup_file"
    return 0
}

#===============================================================================
# backup_mysql
#
# Creates a MySQL database backup
#
# Arguments:
#   $1 - Container ID
#   $2 - Database configuration (format: "host:port:database:user:password")
#   $3 - Backup file path
#
# Returns:
#   0 on success
#   1 on error
#===============================================================================
backup_mysql() {
    local container_id="$1"
    local db_config="$2"
    local backup_file="$3"
    
    log_debug "Backing up MySQL database"
    
    # Parse connection details
    IFS=':' read -r db_host db_port db_name db_user db_password <<< "$db_config"
    
    # Create mysqldump command with single-transaction for consistency
    local mysqldump_cmd="mysqldump -h $db_host -P $db_port -u $db_user -p'$db_password' --single-transaction --quick $db_name > $backup_file"
    
    if ! pct_exec "$container_id" "$mysqldump_cmd"; then
        log_error "Failed to backup MySQL database"
        return 1
    fi
    
    # Compress backup
    pct_exec "$container_id" "gzip -f $backup_file" || log_warn "Failed to compress backup"
    backup_file="${backup_file}.gz"
    
    log_debug "MySQL backup complete: $backup_file"
    return 0
}

#===============================================================================
# verify_backup_integrity
#
# Verifies the integrity of a backup file
#
# Arguments:
#   $1 - Container ID
#   $2 - Backup file path
#   $3 - Database type (sqlite|postgresql|mysql)
#
# Returns:
#   0 if backup is valid
#   1 if backup is invalid or corrupted
#===============================================================================
verify_backup_integrity() {
    local container_id="$1"
    local backup_file="$2"
    local db_type="$3"
    
    if [[ -z "$container_id" || -z "$backup_file" || -z "$db_type" ]]; then
        log_error "verify_backup_integrity: Missing required arguments"
        return 1
    fi
    
    log_debug "Verifying backup integrity: $backup_file"
    
    # Check file exists
    if ! pct_exec "$container_id" "test -f $backup_file"; then
        log_error "Backup file not found: $backup_file"
        return 1
    fi
    
    # Check file size (must be > 0)
    local file_size
    file_size=$(pct_exec "$container_id" "stat -f%z $backup_file 2>/dev/null || stat -c%s $backup_file 2>/dev/null")
    
    if [[ "$file_size" -eq 0 ]]; then
        log_error "Backup file is empty: $backup_file"
        return 1
    fi
    
    # Database-specific integrity checks
    case "$db_type" in
        sqlite)
            # Check SQLite integrity
            if ! pct_exec "$container_id" "sqlite3 $backup_file 'PRAGMA integrity_check;'" | grep -q "ok"; then
                log_error "SQLite backup integrity check failed"
                return 1
            fi
            ;;
        postgresql|postgres)
            # Check pg_dump file can be listed (validates format)
            if ! pct_exec "$container_id" "pg_restore --list $backup_file &>/dev/null"; then
                log_error "PostgreSQL backup integrity check failed"
                return 1
            fi
            ;;
        mysql)
            # For MySQL, check gzip integrity if compressed
            if [[ "$backup_file" == *.gz ]]; then
                if ! pct_exec "$container_id" "gzip -t $backup_file"; then
                    log_error "MySQL backup gzip integrity check failed"
                    return 1
                fi
            fi
            ;;
    esac
    
    log_debug "Backup integrity verified: $backup_file"
    return 0
}

#===============================================================================
# cleanup_old_backups
#
# Removes old backups, keeping only the most recent N backups
#
# Arguments:
#   $1 - Instance ID
#   $2 - Container ID
#   $3 - Max backups to keep (optional, default: 5)
#
# Returns:
#   0 on success
#===============================================================================
cleanup_old_backups() {
    local instance_id="$1"
    local container_id="$2"
    local max_backups="${3:-$MAX_BACKUPS}"
    
    if [[ -z "$instance_id" || -z "$container_id" ]]; then
        log_error "cleanup_old_backups: Missing required arguments"
        return 1
    fi
    
    log_debug "Cleaning up old backups (keeping last $max_backups)..."
    
    local backup_dir="$BACKUP_BASE_DIR"
    
    # Count existing backups for this instance
    local backup_count
    backup_count=$(pct_exec "$container_id" "ls -1 $backup_dir/${instance_id}_* 2>/dev/null | wc -l" || echo "0")
    
    if [[ "$backup_count" -le "$max_backups" ]]; then
        log_debug "No cleanup needed ($backup_count backups <= $max_backups max)"
        return 0
    fi
    
    # Remove oldest backups
    local to_remove=$((backup_count - max_backups))
    log_info "Removing $to_remove old backup(s)..."
    
    # List backups sorted by modification time (oldest first), remove first N
    pct_exec "$container_id" "ls -1t $backup_dir/${instance_id}_* | tail -n $to_remove | xargs rm -f" || {
        log_warn "Failed to cleanup some old backups"
    }
    
    log_success "Backup cleanup complete"
    return 0
}

#===============================================================================
# restore_from_backup
#
# Restores database from a backup file
#
# Arguments:
#   $1 - Container ID
#   $2 - Backup file path
#   $3 - Database type (sqlite|postgresql|mysql)
#   $4 - Database configuration (path or connection details)
#
# Returns:
#   0 on success
#   1 on error
#===============================================================================
restore_from_backup() {
    local container_id="$1"
    local backup_file="$2"
    local db_type="$3"
    local db_config="$4"
    
    if [[ -z "$container_id" || -z "$backup_file" || -z "$db_type" ]]; then
        log_error "restore_from_backup: Missing required arguments"
        return 1
    fi
    
    log_warn "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_warn "RESTORING DATABASE FROM BACKUP"
    log_warn "Backup: $backup_file"
    log_warn "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # Verify backup integrity first
    if ! verify_backup_integrity "$container_id" "$backup_file" "$db_type"; then
        log_error "Backup integrity check failed - cannot restore"
        return 1
    fi
    
    case "$db_type" in
        sqlite)
            if ! restore_sqlite "$container_id" "$db_config" "$backup_file"; then
                return 1
            fi
            ;;
        postgresql|postgres)
            if ! restore_postgresql "$container_id" "$db_config" "$backup_file"; then
                return 1
            fi
            ;;
        mysql)
            if ! restore_mysql "$container_id" "$db_config" "$backup_file"; then
                return 1
            fi
            ;;
        *)
            log_error "Unsupported database type: $db_type"
            return 1
            ;;
    esac
    
    log_success "Database restored from backup"
    return 0
}

#===============================================================================
# restore_sqlite
#===============================================================================
restore_sqlite() {
    local container_id="$1"
    local db_path="$2"
    local backup_file="$3"
    
    log_info "Restoring SQLite database..."
    
    # Stop services first
    pct_exec "$container_id" "systemctl stop 'vikunja-*'" 2>/dev/null || true
    sleep 2
    
    # Backup current database before overwriting
    pct_exec "$container_id" "mv $db_path ${db_path}.before-restore" 2>/dev/null || true
    
    # Restore from backup
    if ! pct_exec "$container_id" "cp $backup_file $db_path"; then
        log_error "Failed to restore SQLite database"
        # Try to restore original
        pct_exec "$container_id" "mv ${db_path}.before-restore $db_path" 2>/dev/null || true
        return 1
    fi
    
    log_success "SQLite database restored"
    return 0
}

#===============================================================================
# restore_postgresql
#===============================================================================
restore_postgresql() {
    local container_id="$1"
    local db_config="$2"
    local backup_file="$3"
    
    log_info "Restoring PostgreSQL database..."
    
    # Parse connection details
    IFS=':' read -r db_host db_port db_name db_user db_password <<< "$db_config"
    
    # Drop and recreate database (WARNING: destructive)
    local drop_cmd="PGPASSWORD='$db_password' psql -h $db_host -p $db_port -U $db_user -c 'DROP DATABASE IF EXISTS ${db_name};'"
    local create_cmd="PGPASSWORD='$db_password' psql -h $db_host -p $db_port -U $db_user -c 'CREATE DATABASE ${db_name};'"
    
    pct_exec "$container_id" "$drop_cmd" || log_warn "Failed to drop database"
    pct_exec "$container_id" "$create_cmd" || {
        log_error "Failed to create database"
        return 1
    }
    
    # Restore from backup
    local restore_cmd="PGPASSWORD='$db_password' pg_restore -h $db_host -p $db_port -U $db_user -d $db_name $backup_file"
    
    if ! pct_exec "$container_id" "$restore_cmd"; then
        log_error "Failed to restore PostgreSQL database"
        return 1
    fi
    
    log_success "PostgreSQL database restored"
    return 0
}

#===============================================================================
# restore_mysql
#===============================================================================
restore_mysql() {
    local container_id="$1"
    local db_config="$2"
    local backup_file="$3"
    
    log_info "Restoring MySQL database..."
    
    # Parse connection details
    IFS=':' read -r db_host db_port db_name db_user db_password <<< "$db_config"
    
    # Decompress if needed
    local restore_file="$backup_file"
    if [[ "$backup_file" == *.gz ]]; then
        restore_file="${backup_file%.gz}"
        pct_exec "$container_id" "gunzip -c $backup_file > $restore_file" || {
            log_error "Failed to decompress backup"
            return 1
        }
    fi
    
    # Restore from backup
    local restore_cmd="mysql -h $db_host -P $db_port -u $db_user -p'$db_password' $db_name < $restore_file"
    
    if ! pct_exec "$container_id" "$restore_cmd"; then
        log_error "Failed to restore MySQL database"
        return 1
    fi
    
    # Cleanup decompressed file
    if [[ "$restore_file" != "$backup_file" ]]; then
        pct_exec "$container_id" "rm -f $restore_file"
    fi
    
    log_success "MySQL database restored"
    return 0
}

# Export functions
export -f create_pre_migration_backup
export -f verify_backup_integrity
export -f cleanup_old_backups
export -f restore_from_backup
export -f backup_sqlite
export -f backup_postgresql
export -f backup_mysql
export -f restore_sqlite
export -f restore_postgresql
export -f restore_mysql
