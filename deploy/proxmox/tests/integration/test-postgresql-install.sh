#!/usr/bin/env bash
# Integration Test: PostgreSQL Installation
# Purpose: Test deployment with PostgreSQL database

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
TEST_NAME="PostgreSQL Installation"
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

assert_success() {
    local message="$1"
    
    if [[ $? -eq 0 ]]; then
        log_pass "$message"
        return 0
    else
        log_fail "$message"
        return 1
    fi
}

# ============================================================================
# Test Configuration
# ============================================================================

TEST_CONFIG_FILE="$(mktemp)"
cat > "$TEST_CONFIG_FILE" <<'EOF'
deployment:
  name: "test-vikunja-pg"
  environment: "production"
  version: "main"

proxmox:
  node: "pve"
  storage: "local-lvm"
  ostemplate: "local:vztmpl/debian-12-standard_12.2-1_amd64.tar.zst"

resources:
  cores: 4
  memory: 4096
  swap: 1024
  disk: 20

network:
  bridge: "vmbr0"
  ip_address: "192.168.1.200/24"
  gateway: "192.168.1.1"
  nameserver: "8.8.8.8"
  domain: "vikunja-prod.local"

database:
  type: "postgresql"
  host: "localhost"
  port: 5432
  name: "vikunja"
  user: "vikunja"
  password: "secure_password_123"

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
  enabled: true
  schedule: "0 2 * * *"
  retention_days: 30

admin:
  email: "admin@example.com"
EOF

# ============================================================================
# Test Cases
# ============================================================================

test_postgresql_installation() {
    log_test "Testing PostgreSQL installation..."
    
    # Source libraries
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/common.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/proxmox-api.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/lxc-setup.sh"
    
    # Load config
    load_config "$TEST_CONFIG_FILE" || return 1
    
    # Get next container ID
    local ct_id
    ct_id=$(get_next_container_id)
    log_info "Using container ID: $ct_id"
    
    # Create container
    create_container "$ct_id" "test-vikunja-pg" || return 1
    pct_start "$ct_id" || return 1
    
    # Install PostgreSQL
    if pct_exec "$ct_id" -- bash -c "apt-get update && apt-get install -y postgresql"; then
        log_pass "PostgreSQL installed successfully"
    else
        log_fail "PostgreSQL installation failed"
        return 1
    fi
    
    # Store container ID
    echo "$ct_id" > "${MOCK_STATE_DIR}/test_pg_container_id"
}

test_postgresql_database_setup() {
    log_test "Testing PostgreSQL database setup..."
    
    # Source libraries
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/common.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/proxmox-api.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/lxc-setup.sh"
    
    # Load config
    load_config "$TEST_CONFIG_FILE" || return 1
    
    # Get container ID
    local ct_id
    ct_id=$(cat "${MOCK_STATE_DIR}/test_pg_container_id" 2>/dev/null || get_next_container_id)
    
    # Setup PostgreSQL
    if setup_postgresql "$ct_id"; then
        log_pass "PostgreSQL database configured successfully"
    else
        log_fail "PostgreSQL database setup failed"
        return 1
    fi
}

test_postgresql_connection() {
    log_test "Testing PostgreSQL connection..."
    
    # Source libraries
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/common.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/proxmox-api.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/lxc-setup.sh"
    
    # Load config
    load_config "$TEST_CONFIG_FILE" || return 1
    
    # Get container ID
    local ct_id
    ct_id=$(cat "${MOCK_STATE_DIR}/test_pg_container_id" 2>/dev/null || get_next_container_id)
    
    # Test database connection
    if test_db_connection "$ct_id"; then
        log_pass "Database connection successful"
    else
        log_fail "Database connection failed"
        return 1
    fi
}

test_postgresql_environment_variables() {
    log_test "Testing PostgreSQL environment variables in systemd..."
    
    # Source libraries
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/common.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/proxmox-api.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/service-setup.sh"
    
    # Load config
    load_config "$TEST_CONFIG_FILE" || return 1
    
    # Get container ID
    local ct_id
    ct_id=$(cat "${MOCK_STATE_DIR}/test_pg_container_id" 2>/dev/null || get_next_container_id)
    
    # Generate systemd unit
    local unit_file="/tmp/vikunja-backend-blue-pg.service"
    
    if generate_systemd_unit "$ct_id" "backend" "blue" > "$unit_file"; then
        log_pass "Systemd unit generated"
        
        # Verify PostgreSQL environment variables
        if grep -q "VIKUNJA_DATABASE_TYPE=postgres" "$unit_file"; then
            log_pass "PostgreSQL database type set"
        else
            log_fail "PostgreSQL database type not set"
        fi
        
        if grep -q "VIKUNJA_DATABASE_HOST=localhost" "$unit_file"; then
            log_pass "Database host set"
        else
            log_fail "Database host not set"
        fi
        
        if grep -q "VIKUNJA_DATABASE_DATABASE=vikunja" "$unit_file"; then
            log_pass "Database name set"
        else
            log_fail "Database name not set"
        fi
    else
        log_fail "Systemd unit generation failed"
        return 1
    fi
    
    # Cleanup
    rm -f "$unit_file"
}

test_postgresql_connection_failure() {
    log_test "Testing PostgreSQL connection failure handling..."
    
    # Source libraries
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/common.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/proxmox-api.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/lxc-setup.sh"
    
    # Load config
    load_config "$TEST_CONFIG_FILE" || return 1
    
    # Get container ID
    local ct_id
    ct_id=$(cat "${MOCK_STATE_DIR}/test_pg_container_id" 2>/dev/null || get_next_container_id)
    
    # Simulate connection failure
    mock_fail_db
    
    # Test database connection (should fail gracefully)
    if test_db_connection "$ct_id"; then
        log_fail "Database connection should have failed"
    else
        log_pass "Database connection failure handled gracefully"
    fi
    
    # Reset mock state
    export MOCK_DB_CONNECTION_SUCCESS=true
}

test_cleanup() {
    log_test "Testing cleanup..."
    
    # Get container ID
    local ct_id
    ct_id=$(cat "${MOCK_STATE_DIR}/test_pg_container_id" 2>/dev/null)
    
    if [[ -n "$ct_id" ]]; then
        if pct destroy "$ct_id"; then
            log_pass "Container destroyed successfully"
        else
            log_fail "Container destruction failed"
        fi
    fi
    
    # Cleanup test files
    rm -f "$TEST_CONFIG_FILE"
    rm -f "${MOCK_STATE_DIR}/test_pg_container_id"
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
    test_postgresql_installation || true
    test_postgresql_database_setup || true
    test_postgresql_connection || true
    test_postgresql_environment_variables || true
    test_postgresql_connection_failure || true
    test_cleanup || true
    
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
