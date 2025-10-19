#!/usr/bin/env bash
# Integration Test: Validation and Error Handling
# Purpose: Test input validation and error scenarios

set -euo pipefail

# ============================================================================
# Test Setup
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Source mock API
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../fixtures/mock-proxmox-api.sh"

# Test results
TEST_NAME="Validation and Error Handling"
TESTS_PASSED=0
TESTS_FAILED=0

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ============================================================================
# Test Helpers
# ============================================================================

log_test() {
    echo -e "${BLUE}[TEST]${NC} $*"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $*"
    ((TESTS_PASSED++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $*"
    ((TESTS_FAILED++))
}

log_info() {
    echo -e "${YELLOW}[INFO]${NC} $*"
}

# ============================================================================
# Test Cases
# ============================================================================

test_ip_validation() {
    log_test "Testing IP address validation..."
    
    # Source common library
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/common.sh"
    
    # Valid IPs
    if validate_ip "192.168.1.1"; then
        log_pass "Valid IP accepted: 192.168.1.1"
    else
        log_fail "Valid IP rejected: 192.168.1.1"
    fi
    
    if validate_ip "10.0.0.255"; then
        log_pass "Valid IP accepted: 10.0.0.255"
    else
        log_fail "Valid IP rejected: 10.0.0.255"
    fi
    
    # Invalid IPs
    if ! validate_ip "256.1.1.1"; then
        log_pass "Invalid IP rejected: 256.1.1.1"
    else
        log_fail "Invalid IP accepted: 256.1.1.1"
    fi
    
    if ! validate_ip "192.168.1"; then
        log_pass "Incomplete IP rejected: 192.168.1"
    else
        log_fail "Incomplete IP accepted: 192.168.1"
    fi
    
    if ! validate_ip "not.an.ip.address"; then
        log_pass "Non-numeric IP rejected: not.an.ip.address"
    else
        log_fail "Non-numeric IP accepted: not.an.ip.address"
    fi
}

test_domain_validation() {
    log_test "Testing domain validation..."
    
    # Source common library
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/common.sh"
    
    # Valid domains
    if validate_domain "example.com"; then
        log_pass "Valid domain accepted: example.com"
    else
        log_fail "Valid domain rejected: example.com"
    fi
    
    if validate_domain "sub.example.com"; then
        log_pass "Subdomain accepted: sub.example.com"
    else
        log_fail "Subdomain rejected: sub.example.com"
    fi
    
    if validate_domain "localhost"; then
        log_pass "Localhost accepted: localhost"
    else
        log_fail "Localhost rejected: localhost"
    fi
    
    # Invalid domains
    if ! validate_domain ""; then
        log_pass "Empty domain rejected"
    else
        log_fail "Empty domain accepted"
    fi
    
    if ! validate_domain "invalid domain with spaces"; then
        log_pass "Domain with spaces rejected"
    else
        log_fail "Domain with spaces accepted"
    fi
}

test_port_validation() {
    log_test "Testing port validation..."
    
    # Source common library
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/common.sh"
    
    # Valid ports
    if validate_port "80"; then
        log_pass "Valid port accepted: 80"
    else
        log_fail "Valid port rejected: 80"
    fi
    
    if validate_port "3456"; then
        log_pass "Valid port accepted: 3456"
    else
        log_fail "Valid port rejected: 3456"
    fi
    
    if validate_port "65535"; then
        log_pass "Max port accepted: 65535"
    else
        log_fail "Max port rejected: 65535"
    fi
    
    # Invalid ports
    if ! validate_port "0"; then
        log_pass "Port 0 rejected"
    else
        log_fail "Port 0 accepted"
    fi
    
    if ! validate_port "65536"; then
        log_pass "Port > 65535 rejected"
    else
        log_fail "Port > 65535 accepted"
    fi
    
    if ! validate_port "abc"; then
        log_pass "Non-numeric port rejected"
    else
        log_fail "Non-numeric port accepted"
    fi
}

test_email_validation() {
    log_test "Testing email validation..."
    
    # Source common library
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/common.sh"
    
    # Valid emails
    if validate_email "user@example.com"; then
        log_pass "Valid email accepted: user@example.com"
    else
        log_fail "Valid email rejected: user@example.com"
    fi
    
    if validate_email "admin+tag@sub.example.com"; then
        log_pass "Complex email accepted: admin+tag@sub.example.com"
    else
        log_fail "Complex email rejected: admin+tag@sub.example.com"
    fi
    
    # Invalid emails
    if ! validate_email "invalid"; then
        log_pass "Invalid email rejected: invalid"
    else
        log_fail "Invalid email accepted: invalid"
    fi
    
    if ! validate_email "no@domain"; then
        log_pass "Email without TLD rejected: no@domain"
    else
        log_fail "Email without TLD accepted: no@domain"
    fi
}

test_invalid_config() {
    log_test "Testing invalid configuration handling..."
    
    # Source common library
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/common.sh"
    
    # Create invalid config (missing required fields)
    local invalid_config
    invalid_config="$(mktemp)"
    cat > "$invalid_config" <<'EOF'
deployment:
  name: ""
  environment: "test"

proxmox:
  node: ""

database:
  type: "invalid_type"
EOF
    
    # Try to load invalid config
    if ! load_config "$invalid_config" 2>/dev/null; then
        log_pass "Invalid config rejected"
    else
        log_fail "Invalid config accepted"
    fi
    
    rm -f "$invalid_config"
}

test_port_conflict() {
    log_test "Testing port conflict detection..."
    
    # Source libraries
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/common.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/proxmox-api.sh"
    
    # Mock: simulate port already in use
    export MOCK_PORT_IN_USE=3456
    
    # Check if port is in use
    if port_in_use "3456"; then
        log_pass "Port conflict detected"
    else
        log_fail "Port conflict not detected"
    fi
    
    # Check available port
    if ! port_in_use "9999"; then
        log_pass "Available port correctly identified"
    else
        log_fail "Available port incorrectly flagged as in use"
    fi
    
    unset MOCK_PORT_IN_USE
}

test_resource_availability() {
    log_test "Testing resource availability checks..."
    
    # Source libraries
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/common.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/proxmox-api.sh"
    
    # Check available resources (should pass with mock)
    if check_resources_available "pve" 2048 2; then
        log_pass "Resource availability check passed"
    else
        log_fail "Resource availability check failed"
    fi
}

test_build_failure_handling() {
    log_test "Testing build failure handling..."
    
    # Source libraries
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/common.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/proxmox-api.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/lxc-setup.sh"
    
    # Create test config
    local test_config
    test_config="$(mktemp)"
    cat > "$test_config" <<'EOF'
deployment:
  name: "test-build-failure"
  environment: "test"
  version: "main"
proxmox:
  node: "pve"
  storage: "local-lvm"
  ostemplate: "local:vztmpl/debian-12-standard_12.2-1_amd64.tar.zst"
resources:
  cores: 2
  memory: 2048
  swap: 512
  disk: 10
network:
  bridge: "vmbr0"
  ip_address: "192.168.1.100/24"
  gateway: "192.168.1.1"
  nameserver: "8.8.8.8"
  domain: "test.local"
database:
  type: "sqlite"
  path: "/opt/vikunja/vikunja.db"
services:
  backend:
    blue_port: 3456
    green_port: 3457
  mcp:
    blue_port: 8456
    green_port: 8457
git:
  backend_repo: "https://github.com/go-vikunja/vikunja.git"
  backend_branch: "main"
  frontend_repo: "https://github.com/go-vikunja/frontend.git"
  frontend_branch: "main"
  mcp_repo: "https://github.com/go-vikunja/mcp-server.git"
  mcp_branch: "main"
backup:
  enabled: false
admin:
  email: "test@example.com"
EOF
    
    load_config "$test_config" || return 1
    
    # Get container ID
    local ct_id
    ct_id=$(get_next_container_id)
    
    # Create container
    create_container "$ct_id" "test-build-failure" || return 1
    pct_start "$ct_id" || return 1
    
    # Simulate build failure
    mock_fail_build
    
    # Try to build (should fail gracefully)
    if ! build_backend "$ct_id" "blue" 2>/dev/null; then
        log_pass "Build failure handled gracefully"
    else
        log_fail "Build failure not detected"
    fi
    
    # Cleanup
    pct destroy "$ct_id" || true
    rm -f "$test_config"
    
    # Reset mock
    export MOCK_BUILD_SUCCESS=true
}

test_service_start_failure() {
    log_test "Testing service start failure handling..."
    
    # Source libraries
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/common.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/proxmox-api.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/service-setup.sh"
    
    # Get container ID
    local ct_id
    ct_id=$(get_next_container_id)
    
    # Simulate service start failure
    mock_fail_service
    
    # Try to start service (should fail gracefully)
    if ! start_service "$ct_id" "vikunja-backend-blue" 2>/dev/null; then
        log_pass "Service start failure handled gracefully"
    else
        log_fail "Service start failure not detected"
    fi
    
    # Reset mock
    export MOCK_SERVICE_START_SUCCESS=true
}

test_lock_management() {
    log_test "Testing lock management..."
    
    # Source common library
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/common.sh"
    
    local lock_name="test-deployment"
    
    # Acquire lock
    if acquire_lock "$lock_name"; then
        log_pass "Lock acquired successfully"
    else
        log_fail "Lock acquisition failed"
    fi
    
    # Try to acquire same lock (should fail)
    if ! acquire_lock "$lock_name" 2>/dev/null; then
        log_pass "Duplicate lock prevented"
    else
        log_fail "Duplicate lock allowed"
    fi
    
    # Release lock
    if release_lock "$lock_name"; then
        log_pass "Lock released successfully"
    else
        log_fail "Lock release failed"
    fi
    
    # Acquire again (should succeed)
    if acquire_lock "$lock_name"; then
        log_pass "Lock re-acquired after release"
        release_lock "$lock_name"
    else
        log_fail "Lock re-acquisition failed"
    fi
}

test_state_management() {
    log_test "Testing state management..."
    
    # Source common library
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/common.sh"
    
    # Set state
    if set_state "test-deployment" "deploying"; then
        log_pass "State set successfully"
    else
        log_fail "State setting failed"
    fi
    
    # Get state
    local state
    state=$(get_state "test-deployment")
    
    if [[ "$state" == "deploying" ]]; then
        log_pass "State retrieved correctly: $state"
    else
        log_fail "State mismatch: expected 'deploying', got '$state'"
    fi
    
    # Update state
    if set_state "test-deployment" "deployed"; then
        log_pass "State updated successfully"
    else
        log_fail "State update failed"
    fi
    
    # Verify update
    state=$(get_state "test-deployment")
    if [[ "$state" == "deployed" ]]; then
        log_pass "Updated state retrieved correctly: $state"
    else
        log_fail "Updated state mismatch: expected 'deployed', got '$state'"
    fi
}

# ============================================================================
# Test Execution
# ============================================================================

main() {
    echo ""
    echo "================================================="
    echo "  Integration Test: $TEST_NAME"
    echo "================================================="
    echo ""
    
    # Reset mock state
    mock_reset
    
    # Run tests
    test_ip_validation || true
    test_domain_validation || true
    test_port_validation || true
    test_email_validation || true
    test_invalid_config || true
    test_port_conflict || true
    test_resource_availability || true
    test_build_failure_handling || true
    test_service_start_failure || true
    test_lock_management || true
    test_state_management || true
    
    # Print summary
    echo ""
    echo "================================================="
    echo "  Test Results"
    echo "================================================="
    echo -e "${GREEN}Passed:${NC} $TESTS_PASSED"
    echo -e "${RED}Failed:${NC} $TESTS_FAILED"
    echo ""
    
    if [[ $TESTS_FAILED -eq 0 ]]; then
        echo -e "${GREEN}✓ All tests passed!${NC}"
        mock_cleanup
        exit 0
    else
        echo -e "${RED}✗ Some tests failed${NC}"
        mock_cleanup
        exit 1
    fi
}

# Run tests if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
