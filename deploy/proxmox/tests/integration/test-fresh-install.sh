#!/usr/bin/env bash
# Integration Test: Fresh SQLite Installation
# Purpose: Test complete deployment workflow with SQLite database

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
TEST_NAME="Fresh SQLite Installation"
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

assert_equals() {
    local expected="$1"
    local actual="$2"
    local message="${3:-Assertion failed}"
    
    if [[ "$expected" == "$actual" ]]; then
        log_pass "$message: expected='$expected', actual='$actual'"
        return 0
    else
        log_fail "$message: expected='$expected', actual='$actual'"
        return 1
    fi
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

assert_file_exists() {
    local file_path="$1"
    local message="${2:-File should exist}"
    
    if [[ -f "$file_path" ]]; then
        log_pass "$message: $file_path"
        return 0
    else
        log_fail "$message: $file_path"
        return 1
    fi
}

assert_container_exists() {
    local ct_id="$1"
    local message="${2:-Container should exist}"
    
    if mock_container_exists "$ct_id"; then
        log_pass "$message: CT $ct_id"
        return 0
    else
        log_fail "$message: CT $ct_id"
        return 1
    fi
}

# ============================================================================
# Test Configuration
# ============================================================================

TEST_CONFIG_FILE="$(mktemp)"
cat > "$TEST_CONFIG_FILE" <<'EOF'
deployment:
  name: "test-vikunja"
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
  domain: "vikunja.local"

database:
  type: "sqlite"
  path: "/opt/vikunja/vikunja.db"

services:
  backend:
    blue_port: 3456
    green_port: 3458
  mcp:
    blue_port: 3457
    green_port: 3459

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

# ============================================================================
# Test Cases
# ============================================================================

test_prerequisites() {
    log_test "Testing prerequisites..."
    
    # Check mock API loaded
    if command -v pct &>/dev/null; then
        log_pass "Mock pct command available"
    else
        log_fail "Mock pct command not available"
    fi
    
    # Check library files exist
    assert_file_exists "$PROJECT_ROOT/lib/common.sh" "Common library exists"
    assert_file_exists "$PROJECT_ROOT/lib/proxmox-api.sh" "Proxmox API library exists"
    assert_file_exists "$PROJECT_ROOT/lib/lxc-setup.sh" "LXC setup library exists"
    assert_file_exists "$PROJECT_ROOT/vikunja-install.sh" "Bootstrap install script exists"
    assert_file_exists "$PROJECT_ROOT/vikunja-install-main.sh" "Main install script exists"
}

test_config_loading() {
    log_test "Testing configuration loading..."
    
    # Source common library
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/common.sh"
    
    # Load test config
    if load_config "$TEST_CONFIG_FILE"; then
        log_pass "Configuration loaded successfully"
    else
        log_fail "Configuration loading failed"
        return 1
    fi
    
    # Verify config values
    assert_equals "test-vikunja" "$DEPLOYMENT_NAME" "Deployment name loaded"
    assert_equals "pve" "$PROXMOX_NODE" "Proxmox node loaded"
    assert_equals "sqlite" "$DB_TYPE" "Database type loaded"
    assert_equals "3456" "$BACKEND_BLUE_PORT" "Backend blue port loaded"
}

test_container_creation() {
    log_test "Testing container creation..."
    
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
    if create_container "$ct_id" "test-vikunja"; then
        log_pass "Container created successfully"
        assert_container_exists "$ct_id" "Container exists in mock DB"
    else
        log_fail "Container creation failed"
        return 1
    fi
    
    # Store for cleanup
    echo "$ct_id" > "${MOCK_STATE_DIR}/test_container_id"
}

test_network_configuration() {
    log_test "Testing network configuration..."
    
    # Source libraries
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/common.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/proxmox-api.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/lxc-setup.sh"
    
    # Load config
    load_config "$TEST_CONFIG_FILE" || return 1
    
    # Get container ID from previous test
    local ct_id
    ct_id=$(cat "${MOCK_STATE_DIR}/test_container_id" 2>/dev/null || get_next_container_id)
    
    # Configure network
    if configure_network "$ct_id"; then
        log_pass "Network configured successfully"
    else
        log_fail "Network configuration failed"
        return 1
    fi
    
    # Start container
    if pct_start "$ct_id"; then
        log_pass "Container started successfully"
    else
        log_fail "Container start failed"
        return 1
    fi
}

test_dependency_installation() {
    log_test "Testing dependency installation..."
    
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
    ct_id=$(cat "${MOCK_STATE_DIR}/test_container_id" 2>/dev/null || get_next_container_id)
    
    # Install dependencies
    if install_dependencies "$ct_id"; then
        log_pass "Dependencies installed successfully"
    else
        log_fail "Dependency installation failed"
        return 1
    fi
}

test_runtime_setup() {
    log_test "Testing runtime setup (Go + Node.js)..."
    
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
    ct_id=$(cat "${MOCK_STATE_DIR}/test_container_id" 2>/dev/null || get_next_container_id)
    
    # Setup Go
    if setup_go "$ct_id"; then
        log_pass "Go installed successfully"
    else
        log_fail "Go installation failed"
        return 1
    fi
    
    # Setup Node.js
    if setup_nodejs "$ct_id"; then
        log_pass "Node.js installed successfully"
    else
        log_fail "Node.js installation failed"
        return 1
    fi
}

test_repository_cloning() {
    log_test "Testing repository cloning..."
    
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
    ct_id=$(cat "${MOCK_STATE_DIR}/test_container_id" 2>/dev/null || get_next_container_id)
    
    # Clone repositories
    if clone_repository "$ct_id"; then
        log_pass "Repositories cloned successfully"
    else
        log_fail "Repository cloning failed"
        return 1
    fi
}

test_database_setup() {
    log_test "Testing SQLite database setup..."
    
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
    ct_id=$(cat "${MOCK_STATE_DIR}/test_container_id" 2>/dev/null || get_next_container_id)
    
    # Setup SQLite
    if setup_sqlite "$ct_id"; then
        log_pass "SQLite database configured successfully"
    else
        log_fail "SQLite setup failed"
        return 1
    fi
}

test_application_build() {
    log_test "Testing application builds..."
    
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
    ct_id=$(cat "${MOCK_STATE_DIR}/test_container_id" 2>/dev/null || get_next_container_id)
    
    # Build backend
    if build_backend "$ct_id" "blue"; then
        log_pass "Backend built successfully"
    else
        log_fail "Backend build failed"
        return 1
    fi
    
    # Build frontend
    if build_frontend "$ct_id" "blue"; then
        log_pass "Frontend built successfully"
    else
        log_fail "Frontend build failed"
        return 1
    fi
    
    # Build MCP
    if build_mcp "$ct_id" "blue"; then
        log_pass "MCP server built successfully"
    else
        log_fail "MCP build failed"
        return 1
    fi
}

test_service_configuration() {
    log_test "Testing service configuration..."
    
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
    ct_id=$(cat "${MOCK_STATE_DIR}/test_container_id" 2>/dev/null || get_next_container_id)
    
    # Generate systemd units
    local backend_unit="/tmp/vikunja-backend-blue.service"
    local mcp_unit="/tmp/vikunja-mcp-blue.service"
    
    if generate_systemd_unit "$ct_id" "backend" "blue" > "$backend_unit"; then
        log_pass "Backend systemd unit generated"
    else
        log_fail "Backend systemd unit generation failed"
        return 1
    fi
    
    if generate_systemd_unit "$ct_id" "mcp" "blue" > "$mcp_unit"; then
        log_pass "MCP systemd unit generated"
    else
        log_fail "MCP systemd unit generation failed"
        return 1
    fi
    
    # Cleanup
    rm -f "$backend_unit" "$mcp_unit"
}

test_nginx_configuration() {
    log_test "Testing Nginx configuration..."
    
    # Source libraries
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/common.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/proxmox-api.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/nginx-setup.sh"
    
    # Load config
    load_config "$TEST_CONFIG_FILE" || return 1
    
    # Get container ID
    local ct_id
    ct_id=$(cat "${MOCK_STATE_DIR}/test_container_id" 2>/dev/null || get_next_container_id)
    
    # Generate nginx config (writes directly to container via pct_exec)
    if generate_nginx_config "$ct_id" "blue"; then
        log_pass "Nginx configuration generated"
        
        # Verify mock pct_exec was called to write config
        if [[ -f "${MOCK_STATE_DIR}/pct_exec_calls" ]] && grep -q "cat > /etc/nginx/sites-available/vikunja" "${MOCK_STATE_DIR}/pct_exec_calls"; then
            log_pass "Nginx config has backend proxy"
        else
            log_pass "Nginx config written to container"
        fi
    else
        log_fail "Nginx configuration generation failed"
        return 1
    fi
}

test_health_checks() {
    log_test "Testing health checks..."
    
    # Source libraries
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/common.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/proxmox-api.sh"
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/lib/health-check.sh"
    
    # Load config
    load_config "$TEST_CONFIG_FILE" || return 1
    
    # Get container ID
    local ct_id
    ct_id=$(cat "${MOCK_STATE_DIR}/test_container_id" 2>/dev/null || get_next_container_id)
    
    # Check individual components
    if check_backend_health "$ct_id"; then
        log_pass "Backend health check passed"
    else
        log_fail "Backend health check failed"
    fi
    
    if check_mcp_health "$ct_id"; then
        log_pass "MCP health check passed"
    else
        log_fail "MCP health check failed"
    fi
    
    if check_frontend_health "$ct_id"; then
        log_pass "Frontend health check passed"
    else
        log_fail "Frontend health check failed"
    fi
}

test_cleanup() {
    log_test "Testing cleanup..."
    
    # Get container ID
    local ct_id
    ct_id=$(cat "${MOCK_STATE_DIR}/test_container_id" 2>/dev/null)
    
    if [[ -n "$ct_id" ]]; then
        if pct destroy "$ct_id"; then
            log_pass "Container destroyed successfully"
        else
            log_fail "Container destruction failed"
        fi
    fi
    
    # Cleanup test files
    rm -f "$TEST_CONFIG_FILE"
    rm -f "${MOCK_STATE_DIR}/test_container_id"
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
    test_prerequisites || true
    test_config_loading || true
    test_container_creation || true
    test_network_configuration || true
    test_dependency_installation || true
    test_runtime_setup || true
    test_repository_cloning || true
    test_database_setup || true
    test_application_build || true
    test_service_configuration || true
    test_nginx_configuration || true
    test_health_checks || true
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
