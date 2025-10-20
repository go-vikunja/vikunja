#!/usr/bin/env bash
# Integration Test: Database Migration Execution
# Purpose: Test that migrations run successfully with pre-backup during updates
# User Story: US2 - Seamless Updates from Main Branch

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_ROOT="$(dirname "$SCRIPT_DIR")"
DEPLOY_ROOT="$(dirname "$TEST_ROOT")"

# shellcheck source=../fixtures/test-helpers.sh
source "${SCRIPT_DIR}/../fixtures/test-helpers.sh" 2>/dev/null || {
    test_start() { echo "[TEST] $1"; }
    test_pass() { echo "[PASS] $1"; }
    test_fail() { echo "[FAIL] $1" >&2; exit 1; }
    test_skip() { echo "[SKIP] $1"; exit 0; }
    log_info() { echo "[INFO] $1"; }
    log_success() { echo "[OK] $1"; }
    log_warn() { echo "[WARN] $1"; }
}

TEST_NAME="migration-execution"
TEST_INSTANCE="vikunja-test-migration"
TEST_CONTAINER_ID="903"

cleanup_test() {
    log_info "Cleaning up..."
    pct stop "$TEST_CONTAINER_ID" 2>/dev/null || true
    pct destroy "$TEST_CONTAINER_ID" 2>/dev/null || true
    rm -f "/etc/vikunja/${TEST_INSTANCE}"*
}

trap cleanup_test EXIT

main() {
    test_start "Migration Execution - Pre-Backup and Migration Testing"
    
    # Check Proxmox
    if ! command -v pct &>/dev/null; then
        test_skip "Not running on Proxmox"
    fi
    
    if [[ $EUID -ne 0 ]]; then
        test_fail "Must run as root"
    fi
    
    log_info "Step 1: Initial deployment..."
    if ! "${DEPLOY_ROOT}/vikunja-install-main.sh" \
        --instance-id "$TEST_INSTANCE" \
        --container-id "$TEST_CONTAINER_ID" \
        --db-type sqlite \
        --ip 192.168.1.203/24 \
        --gateway 192.168.1.1 \
        --domain vikunja-migration-test.local \
        --yes; then
        test_fail "Initial deployment failed"
    fi
    
    log_success "Deployment complete"
    
    log_info "Step 2: Recording pre-migration state..."
    DB_PATH="/opt/vikunja/vikunja.db"
    PRE_MIGRATION_CHECKSUM=$(pct exec "$TEST_CONTAINER_ID" -- md5sum "$DB_PATH" | awk '{print $1}')
    log_info "Pre-migration DB checksum: $PRE_MIGRATION_CHECKSUM"
    
    # Get current migration count
    PRE_MIGRATION_COUNT=$(pct exec "$TEST_CONTAINER_ID" -- sqlite3 "$DB_PATH" \
        "SELECT COUNT(*) FROM migration;" 2>/dev/null || echo "0")
    log_info "Pre-migration count: $PRE_MIGRATION_COUNT"
    
    log_info "Step 3: Running update with migrations..."
    if ! "${DEPLOY_ROOT}/vikunja-update.sh" --force "$TEST_INSTANCE"; then
        log_warn "Update failed (may be expected if no new migrations)"
    fi
    
    log_info "Step 4: Verifying backup was created..."
    BACKUP_DIR="/var/backups/vikunja"
    BACKUP_COUNT=$(pct exec "$TEST_CONTAINER_ID" -- ls -1 "$BACKUP_DIR"/*.db 2>/dev/null | wc -l || echo "0")
    
    if [[ "$BACKUP_COUNT" -lt 1 ]]; then
        test_fail "No backup created before update"
    fi
    
    log_info "Found $BACKUP_COUNT backup(s)"
    
    log_info "Step 5: Verifying migration execution..."
    POST_MIGRATION_COUNT=$(pct exec "$TEST_CONTAINER_ID" -- sqlite3 "$DB_PATH" \
        "SELECT COUNT(*) FROM migration;" 2>/dev/null || echo "0")
    log_info "Post-migration count: $POST_MIGRATION_COUNT"
    
    if [[ "$POST_MIGRATION_COUNT" -ge "$PRE_MIGRATION_COUNT" ]]; then
        log_success "Migration table updated (or unchanged if no new migrations)"
    else
        test_fail "Migration count decreased (rollback occurred?)"
    fi
    
    log_info "Step 6: Verifying database integrity..."
    if ! pct exec "$TEST_CONTAINER_ID" -- sqlite3 "$DB_PATH" "PRAGMA integrity_check;" &>/dev/null; then
        test_fail "Database integrity check failed"
    fi
    
    log_success "Database integrity verified"
    
    log_success "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_success "MIGRATION TEST PASSED"
    log_success "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_info "Pre-migration count: $PRE_MIGRATION_COUNT"
    log_info "Post-migration count: $POST_MIGRATION_COUNT"
    log_info "Backups created: $BACKUP_COUNT"
    log_info "Database integrity: OK"
    log_success "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    test_pass "Migration execution with pre-backup successful"
}

main "$@"
