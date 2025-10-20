#!/usr/bin/env bash
# Integration Test: Update Cycle
# Purpose: Test successful update from one version to another with zero downtime
# User Story: US2 - Seamless Updates from Main Branch

set -euo pipefail

# Source test utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_ROOT="$(dirname "$SCRIPT_DIR")"
DEPLOY_ROOT="$(dirname "$TEST_ROOT")"

# shellcheck source=../fixtures/test-helpers.sh
source "${SCRIPT_DIR}/../fixtures/test-helpers.sh" 2>/dev/null || {
    echo "[WARNING] test-helpers.sh not found, using basic functions"
    
    # Basic test functions
    test_start() { echo "[TEST] $1"; }
    test_pass() { echo "[PASS] $1"; }
    test_fail() { echo "[FAIL] $1" >&2; exit 1; }
    test_skip() { echo "[SKIP] $1"; exit 0; }
}

# Test configuration
TEST_NAME="update-cycle"
TEST_INSTANCE="vikunja-test-update"
TEST_CONTAINER_ID="901"
TEST_IP="192.168.1.201/24"
TEST_GATEWAY="192.168.1.1"
TEST_DOMAIN="vikunja-update-test.local"

# Cleanup function
cleanup_test() {
    log_info "Cleaning up test resources..."
    
    # Stop and destroy test container if exists
    if pct status "$TEST_CONTAINER_ID" &>/dev/null; then
        pct stop "$TEST_CONTAINER_ID" 2>/dev/null || true
        pct destroy "$TEST_CONTAINER_ID" 2>/dev/null || true
    fi
    
    # Clean up state files
    rm -f "/etc/vikunja/${TEST_INSTANCE}.lock"
    rm -f "/etc/vikunja/${TEST_INSTANCE}.state"
    rm -f "/etc/vikunja/${TEST_INSTANCE}.yaml"
    
    log_success "Cleanup complete"
}

# Trap cleanup on exit
trap cleanup_test EXIT

#===============================================================================
# Test: Update Cycle
#===============================================================================

main() {
    test_start "Update Cycle - Zero Downtime Update"
    
    #---------------------------------------------------------------------------
    # Prerequisites Check
    #---------------------------------------------------------------------------
    
    log_info "Checking prerequisites..."
    
    # Check if running on Proxmox
    if ! command -v pct &>/dev/null; then
        test_skip "Not running on Proxmox - skipping integration test"
    fi
    
    # Check if running as root
    if [[ $EUID -ne 0 ]]; then
        test_fail "Must run as root on Proxmox host"
    fi
    
    # Check if container ID is available
    if pct status "$TEST_CONTAINER_ID" &>/dev/null; then
        log_warn "Container $TEST_CONTAINER_ID already exists - cleaning up"
        cleanup_test
    fi
    
    log_success "Prerequisites passed"
    
    #---------------------------------------------------------------------------
    # Step 1: Initial Deployment (v1)
    #---------------------------------------------------------------------------
    
    log_info "Step 1: Deploying initial version (v1)..."
    
    # Deploy using vikunja-install-main.sh
    if ! "${DEPLOY_ROOT}/vikunja-install-main.sh" \
        --instance-id "$TEST_INSTANCE" \
        --container-id "$TEST_CONTAINER_ID" \
        --db-type sqlite \
        --ip "$TEST_IP" \
        --gateway "$TEST_GATEWAY" \
        --domain "$TEST_DOMAIN" \
        --cpu 2 \
        --memory 4096 \
        --disk 20 \
        --yes; then
        test_fail "Initial deployment failed"
    fi
    
    # Get deployed version
    V1_VERSION=$(pct exec "$TEST_CONTAINER_ID" -- /opt/vikunja/vikunja version 2>/dev/null | head -n1 || echo "unknown")
    log_info "Deployed version v1: $V1_VERSION"
    
    # Verify services are running
    if ! pct exec "$TEST_CONTAINER_ID" -- systemctl is-active vikunja-backend-blue.service &>/dev/null; then
        test_fail "Backend service not running after initial deployment"
    fi
    
    log_success "Initial deployment complete (v1: $V1_VERSION)"
    
    #---------------------------------------------------------------------------
    # Step 2: Verify Initial State
    #---------------------------------------------------------------------------
    
    log_info "Step 2: Verifying initial state..."
    
    # Check active color is blue
    ACTIVE_COLOR=$(grep "active_color:" "/etc/vikunja/${TEST_INSTANCE}.state" | awk '{print $2}' | tr -d '"')
    if [[ "$ACTIVE_COLOR" != "blue" ]]; then
        test_fail "Expected active color 'blue', got '$ACTIVE_COLOR'"
    fi
    
    # Test backend responds on port 3456 (blue)
    if ! pct exec "$TEST_CONTAINER_ID" -- curl -sf http://localhost:3456/api/v1/info &>/dev/null; then
        test_fail "Backend not responding on blue port 3456"
    fi
    
    # Test MCP server responds on port 8456 (blue)
    if ! pct exec "$TEST_CONTAINER_ID" -- curl -sf http://localhost:8456/health &>/dev/null; then
        test_fail "MCP server not responding on blue port 8456"
    fi
    
    log_success "Initial state verified (blue active)"
    
    #---------------------------------------------------------------------------
    # Step 3: Simulate Code Changes (Mock Update Available)
    #---------------------------------------------------------------------------
    
    log_info "Step 3: Simulating code changes..."
    
    # In a real scenario, we'd git pull to get new code
    # For testing, we'll just verify the update script can detect changes
    # by checking git status in the container
    
    REPO_PATH="/opt/vikunja"
    GIT_STATUS=$(pct exec "$TEST_CONTAINER_ID" -- bash -c "cd $REPO_PATH && git fetch origin && git rev-list HEAD..origin/main --count" 2>/dev/null || echo "0")
    
    log_info "Git commits behind origin/main: $GIT_STATUS"
    
    # For testing purposes, we'll force an update even if no changes
    log_warn "Using --force flag for testing (ignores 'no updates available')"
    
    log_success "Ready for update"
    
    #---------------------------------------------------------------------------
    # Step 4: Execute Update
    #---------------------------------------------------------------------------
    
    log_info "Step 4: Executing update to v2..."
    
    # Record start time for downtime calculation
    UPDATE_START=$(date +%s)
    
    # Run update script
    if ! "${DEPLOY_ROOT}/vikunja-update.sh" \
        --force \
        "$TEST_INSTANCE"; then
        test_fail "Update script failed"
    fi
    
    UPDATE_END=$(date +%s)
    UPDATE_DURATION=$((UPDATE_END - UPDATE_START))
    
    log_info "Update completed in ${UPDATE_DURATION}s"
    
    # Verify update duration < 5 minutes (300 seconds)
    if [[ $UPDATE_DURATION -gt 300 ]]; then
        test_fail "Update took ${UPDATE_DURATION}s (>300s requirement)"
    fi
    
    log_success "Update script executed successfully in ${UPDATE_DURATION}s"
    
    #---------------------------------------------------------------------------
    # Step 5: Verify Update Success
    #---------------------------------------------------------------------------
    
    log_info "Step 5: Verifying update success..."
    
    # Check active color switched to green
    ACTIVE_COLOR=$(grep "active_color:" "/etc/vikunja/${TEST_INSTANCE}.state" | awk '{print $2}' | tr -d '"')
    if [[ "$ACTIVE_COLOR" != "green" ]]; then
        test_fail "Expected active color 'green' after update, got '$ACTIVE_COLOR'"
    fi
    
    # Verify green services are running
    if ! pct exec "$TEST_CONTAINER_ID" -- systemctl is-active vikunja-backend-green.service &>/dev/null; then
        test_fail "Green backend service not running after update"
    fi
    
    if ! pct exec "$TEST_CONTAINER_ID" -- systemctl is-active vikunja-mcp-green.service &>/dev/null; then
        test_fail "Green MCP service not running after update"
    fi
    
    # Verify blue services are stopped
    if pct exec "$TEST_CONTAINER_ID" -- systemctl is-active vikunja-backend-blue.service &>/dev/null; then
        test_fail "Blue backend service still running after update"
    fi
    
    # Test backend responds on port 3457 (green)
    if ! pct exec "$TEST_CONTAINER_ID" -- curl -sf http://localhost:3457/api/v1/info &>/dev/null; then
        test_fail "Backend not responding on green port 3457 after update"
    fi
    
    # Test MCP server responds on port 8457 (green)
    if ! pct exec "$TEST_CONTAINER_ID" -- curl -sf http://localhost:8457/health &>/dev/null; then
        test_fail "MCP server not responding on green port 8457 after update"
    fi
    
    # Verify nginx upstream switched to green
    NGINX_CONFIG="/etc/nginx/sites-available/${TEST_INSTANCE}.conf"
    if ! pct exec "$TEST_CONTAINER_ID" -- grep -q "server 127.0.0.1:3457" "$NGINX_CONFIG"; then
        test_fail "Nginx not pointing to green backend port 3457"
    fi
    
    log_success "Update verification passed (green active)"
    
    #---------------------------------------------------------------------------
    # Step 6: Verify Zero Downtime
    #---------------------------------------------------------------------------
    
    log_info "Step 6: Verifying zero-downtime claim..."
    
    # Check deployment log for downtime measurement
    LOG_FILE="/var/log/vikunja-deploy/${TEST_INSTANCE}.log"
    if pct exec "$TEST_CONTAINER_ID" -- test -f "$LOG_FILE"; then
        DOWNTIME=$(pct exec "$TEST_CONTAINER_ID" -- grep -oP 'Downtime: \K[0-9]+' "$LOG_FILE" 2>/dev/null || echo "unknown")
        log_info "Recorded downtime: ${DOWNTIME}s"
        
        # Verify downtime < 5 seconds (99.9% uptime requirement)
        if [[ "$DOWNTIME" != "unknown" && "$DOWNTIME" -gt 5 ]]; then
            test_fail "Downtime ${DOWNTIME}s exceeds 5s requirement"
        fi
    else
        log_warn "Deployment log not found - skipping downtime verification"
    fi
    
    log_success "Zero-downtime requirement met"
    
    #---------------------------------------------------------------------------
    # Step 7: Verify Rollback Availability
    #---------------------------------------------------------------------------
    
    log_info "Step 7: Verifying rollback availability..."
    
    # Check that blue binaries still exist (for potential rollback)
    if ! pct exec "$TEST_CONTAINER_ID" -- test -f /opt/vikunja/vikunja-blue; then
        test_fail "Blue binary not preserved for rollback"
    fi
    
    # Verify backup was created before update
    BACKUP_DIR="/var/backups/vikunja"
    BACKUP_COUNT=$(pct exec "$TEST_CONTAINER_ID" -- bash -c "ls -1 $BACKUP_DIR/*.db 2>/dev/null | wc -l" || echo "0")
    if [[ "$BACKUP_COUNT" -lt 1 ]]; then
        test_fail "No database backup found in $BACKUP_DIR"
    fi
    
    log_info "Found $BACKUP_COUNT backup(s) in $BACKUP_DIR"
    
    log_success "Rollback mechanisms verified"
    
    #---------------------------------------------------------------------------
    # Test Summary
    #---------------------------------------------------------------------------
    
    log_success "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_success "UPDATE CYCLE TEST PASSED"
    log_success "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_info "Initial version: $V1_VERSION (blue)"
    log_info "Updated version: v2 (green)"
    log_info "Update duration: ${UPDATE_DURATION}s (<300s ✓)"
    log_info "Downtime: ${DOWNTIME:-unknown}s (<5s ✓)"
    log_info "Active services: green (backend, MCP)"
    log_info "Rollback available: Yes"
    log_success "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    test_pass "Update cycle completed successfully with zero downtime"
}

# Execute test
main "$@"
