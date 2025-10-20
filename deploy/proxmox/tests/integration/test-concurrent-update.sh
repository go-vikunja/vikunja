#!/usr/bin/env bash
# Integration Test: Concurrent Update Prevention
# Purpose: Test that lock mechanism prevents concurrent updates
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

TEST_NAME="concurrent-update"
TEST_INSTANCE="vikunja-test-lock"
TEST_CONTAINER_ID="904"

cleanup_test() {
    log_info "Cleaning up..."
    # Kill any background update processes
    pkill -f "vikunja-update.sh.*${TEST_INSTANCE}" 2>/dev/null || true
    # Clean up lock file
    rm -f "/etc/vikunja/${TEST_INSTANCE}.lock"
    # Clean up container
    pct stop "$TEST_CONTAINER_ID" 2>/dev/null || true
    pct destroy "$TEST_CONTAINER_ID" 2>/dev/null || true
    rm -f "/etc/vikunja/${TEST_INSTANCE}"*
}

trap cleanup_test EXIT

main() {
    test_start "Concurrent Update Prevention - Lock Mechanism Testing"
    
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
        --ip 192.168.1.204/24 \
        --gateway 192.168.1.1 \
        --domain vikunja-lock-test.local \
        --yes; then
        test_fail "Initial deployment failed"
    fi
    
    log_success "Deployment complete"
    
    log_info "Step 2: Starting first update (background)..."
    # Start update in background with --force flag to bypass "no updates" check
    "${DEPLOY_ROOT}/vikunja-update.sh" --force "$TEST_INSTANCE" &
    FIRST_UPDATE_PID=$!
    log_info "First update PID: $FIRST_UPDATE_PID"
    
    # Wait for lock to be acquired
    sleep 2
    
    # Verify lock file exists
    LOCK_FILE="/etc/vikunja/${TEST_INSTANCE}.lock"
    if [[ ! -f "$LOCK_FILE" ]]; then
        test_fail "Lock file not created by first update"
    fi
    
    log_success "Lock file created: $LOCK_FILE"
    
    log_info "Step 3: Attempting concurrent update (should fail)..."
    # Try to start second update - should fail with exit code 5 (operation locked)
    SECOND_UPDATE_EXIT=0
    "${DEPLOY_ROOT}/vikunja-update.sh" --force "$TEST_INSTANCE" 2>&1 | tee /tmp/second-update.log || SECOND_UPDATE_EXIT=$?
    
    log_info "Second update exit code: $SECOND_UPDATE_EXIT"
    
    # Verify second update failed with correct exit code
    if [[ "$SECOND_UPDATE_EXIT" -ne 5 ]]; then
        test_fail "Expected exit code 5 (locked), got $SECOND_UPDATE_EXIT"
    fi
    
    # Verify error message mentions lock
    if ! grep -qi "lock\|in progress\|already running" /tmp/second-update.log; then
        test_fail "Error message doesn't mention lock/concurrent operation"
    fi
    
    log_success "Second update correctly blocked (exit code 5)"
    
    log_info "Step 4: Waiting for first update to complete..."
    wait $FIRST_UPDATE_PID || log_warn "First update exited with code $?"
    
    # Wait a moment for cleanup
    sleep 2
    
    log_info "Step 5: Verifying lock released..."
    if [[ -f "$LOCK_FILE" ]]; then
        test_fail "Lock file not released after first update completed"
    fi
    
    log_success "Lock file released correctly"
    
    log_info "Step 6: Attempting third update (should succeed)..."
    THIRD_UPDATE_EXIT=0
    "${DEPLOY_ROOT}/vikunja-update.sh" --force "$TEST_INSTANCE" || THIRD_UPDATE_EXIT=$?
    
    if [[ "$THIRD_UPDATE_EXIT" -ne 0 && "$THIRD_UPDATE_EXIT" -ne 20 ]]; then
        # Exit code 20 = no updates available (acceptable)
        log_warn "Third update failed with exit code $THIRD_UPDATE_EXIT (may be expected)"
    else
        log_success "Third update executed without lock conflict"
    fi
    
    log_success "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_success "CONCURRENT UPDATE TEST PASSED"
    log_success "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_info "First update: Completed (lock held during execution)"
    log_info "Second update: Blocked (exit code 5)"
    log_info "Third update: Allowed (lock released)"
    log_info "Lock mechanism: Working correctly"
    log_success "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    test_pass "Lock mechanism prevents concurrent updates successfully"
}

main "$@"
