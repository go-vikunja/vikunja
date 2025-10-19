# Integration Tests

Comprehensive integration test suite for the Vikunja Proxmox LXC deployment system.

## Overview

These tests use a **mock Proxmox API** to simulate the complete deployment workflow without requiring an actual Proxmox VE host. This enables:

- ✅ CI/CD pipeline integration
- ✅ Local development testing
- ✅ Fast test execution
- ✅ Consistent test environment

## Quick Start

```bash
# Run all tests
./tests/run-tests.sh

# Run with verbose output
./tests/run-tests.sh --verbose

# Run specific test
./tests/run-tests.sh --test test-fresh-install
```

## Test Structure

```
tests/
├── run-tests.sh              # Test runner orchestrator
├── fixtures/
│   └── mock-proxmox-api.sh   # Mock Proxmox API (pct, pvesh, etc.)
└── integration/
    ├── test-fresh-install.sh        # SQLite deployment test (13 test cases)
    ├── test-postgresql-install.sh   # PostgreSQL deployment test (6 test cases)
    └── test-validation-errors.sh    # Input validation test (11 test cases)
```

## Test Coverage

### test-fresh-install.sh (SQLite Deployment)

Tests the complete deployment workflow with SQLite database:

1. ✅ Prerequisites check (libraries exist, mock API available)
2. ✅ Configuration loading (YAML parsing)
3. ✅ Container creation (LXC provisioning)
4. ✅ Network configuration (IP, gateway, nameserver)
5. ✅ Dependency installation (system packages)
6. ✅ Runtime setup (Go 1.21.5, Node.js 22)
7. ✅ Repository cloning (backend, frontend, MCP)
8. ✅ SQLite database setup
9. ✅ Application builds (backend, frontend, MCP)
10. ✅ Systemd service configuration
11. ✅ Nginx configuration (HTTP/HTTPS, WebSocket, file uploads)
12. ✅ Health checks (backend, MCP, frontend)
13. ✅ Cleanup (container destruction)

### test-postgresql-install.sh (PostgreSQL Deployment)

Tests deployment with PostgreSQL database:

1. ✅ PostgreSQL installation
2. ✅ Database setup (create user, database, grant permissions)
3. ✅ Connection validation
4. ✅ Environment variables in systemd units
5. ✅ Connection failure handling
6. ✅ Cleanup

### test-validation-errors.sh (Input Validation)

Tests input validation and error handling:

1. ✅ IP address validation (valid/invalid formats)
2. ✅ Domain validation (FQDN, localhost, invalid)
3. ✅ Port validation (0-65535 range)
4. ✅ Email validation (RFC compliance)
5. ✅ Invalid configuration rejection
6. ✅ Port conflict detection
7. ✅ Resource availability checks
8. ✅ Build failure handling
9. ✅ Service start failure handling
10. ✅ Lock management (acquire, release, stale detection)
11. ✅ State management (set, get, update)

## Mock Proxmox API

The mock API simulates the following commands:

### Proxmox Commands
- `pct create/start/stop/status/exec/destroy/push/pull/config/set` - Container management
- `pvesh get` - API queries (nodes, resources, cluster info)

### System Commands
- `apt-get update/install` - Package management
- `git clone` - Repository operations
- `systemctl daemon-reload/enable/start/stop/restart/status/is-active` - Service management
- `nginx -t` - Configuration validation
- `curl` - Health check endpoints
- `ss -tuln` - Port availability
- `ip addr show` - Network information

### Build Tools
- `mage build` - Backend compilation
- `pnpm install/build` - Frontend builds
- `node` - MCP server execution

## Mock Configuration

Control mock behavior with environment variables:

```bash
# Enable debug logging
export MOCK_DEBUG=1

# Simulate failures
export MOCK_BUILD_SUCCESS=false
export MOCK_SERVICE_START_SUCCESS=false
export MOCK_DB_CONNECTION_SUCCESS=false
```

## Mock State Management

The mock API maintains state in temporary directories:

- **Container database**: Tracks created containers and their status
- **Mock filesystem**: Simulates container filesystem for file checks
- **Network state**: Simulates network availability

State is automatically cleaned up after each test run.

## Writing New Tests

### Basic Test Structure

```bash
#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Source mock API
source "$SCRIPT_DIR/../fixtures/mock-proxmox-api.sh"

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0

# Test helpers
log_pass() {
    echo -e "${GREEN}[PASS]${NC} $*"
    ((TESTS_PASSED++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $*"
    ((TESTS_FAILED++))
}

# Test case
test_my_feature() {
    log_test "Testing my feature..."
    
    # Source libraries
    source "$PROJECT_ROOT/lib/common.sh"
    
    # Your test logic here
    if my_function; then
        log_pass "Feature works"
    else
        log_fail "Feature failed"
    fi
}

# Main execution
main() {
    mock_reset  # Reset state
    test_my_feature || true
    
    if [[ $TESTS_FAILED -eq 0 ]]; then
        mock_cleanup
        exit 0
    else
        mock_cleanup
        exit 1
    fi
}

main "$@"
```

### Test Assertions

```bash
# Equality check
assert_equals "expected" "$actual" "Message"

# Success check
some_command
assert_success "Command succeeded"

# File existence
assert_file_exists "/path/to/file" "File should exist"

# Container existence
assert_container_exists "$ct_id" "Container should exist"
```

### Mock Helpers

```bash
# Reset mock state
mock_reset

# Simulate failures
mock_fail_build
mock_fail_service
mock_fail_db

# Check mock state
mock_container_count     # Returns number of containers
mock_container_exists 100  # Check if CT 100 exists

# Cleanup
mock_cleanup
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Run integration tests
        run: |
          cd deploy/proxmox
          ./tests/run-tests.sh
```

### GitLab CI Example

```yaml
integration-tests:
  image: ubuntu:22.04
  script:
    - cd deploy/proxmox
    - ./tests/run-tests.sh
```

## Troubleshooting

### Tests fail with "pct: command not found"

The mock API should be sourced automatically. If this error occurs:

```bash
# Verify mock is loaded
source tests/fixtures/mock-proxmox-api.sh
type pct  # Should show "pct is a function"
```

### Tests hang or timeout

Check for infinite loops in wait functions:

```bash
# Run with debug logging
MOCK_DEBUG=1 ./tests/run-tests.sh --test test-fresh-install
```

### Mock state persists between tests

Ensure each test calls `mock_reset` at the beginning:

```bash
main() {
    mock_reset  # Required!
    # ... tests ...
}
```

## Performance

Typical test execution times:

- **test-fresh-install.sh**: ~2-3 seconds (13 test cases)
- **test-postgresql-install.sh**: ~1-2 seconds (6 test cases)
- **test-validation-errors.sh**: ~1-2 seconds (11 test cases)
- **Full suite**: ~5-10 seconds (30 test cases)

## Future Enhancements

Planned test additions:

- [ ] Blue-green deployment test (User Story 2)
- [ ] Rollback test (User Story 2)
- [ ] Backup and restore test (User Story 3)
- [ ] Multi-database test (SQLite → PostgreSQL migration)
- [ ] SSL/TLS certificate test
- [ ] Network failure scenarios
- [ ] Concurrent deployment test

## Contributing

When adding new tests:

1. Follow the naming convention: `test-<feature>.sh`
2. Include test documentation in this README
3. Use the mock API for all Proxmox operations
4. Ensure tests clean up after themselves
5. Add color-coded output for readability
6. Include both positive and negative test cases

## License

Same as the main Vikunja project (AGPLv3).
