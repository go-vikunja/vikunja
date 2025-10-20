#!/usr/bin/env bash
# Test Runner: Execute all integration tests
# Purpose: Run complete integration test suite with reporting

set -euo pipefail

# ============================================================================
# Configuration
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Test suite results
TOTAL_PASSED=0
TOTAL_FAILED=0
SUITE_START_TIME=$(date +%s)

# ============================================================================
# Helper Functions
# ============================================================================

log_header() {
    echo ""
    echo -e "${CYAN}=================================================${NC}"
    echo -e "${CYAN}  $*${NC}"
    echo -e "${CYAN}=================================================${NC}"
    echo ""
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $*"
}

run_test() {
    local test_file="$1"
    local test_name
    test_name=$(basename "$test_file" .sh)
    
    log_header "Running: $test_name"
    
    local start_time
    start_time=$(date +%s)
    
    if bash "$test_file"; then
        local end_time
        end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        log_success "$test_name completed in ${duration}s"
        return 0
    else
        local end_time
        end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        log_error "$test_name failed after ${duration}s"
        return 1
    fi
}

# ============================================================================
# Test Discovery
# ============================================================================

discover_tests() {
    local test_dir="$1"
    
    # Send logs to stderr to avoid interfering with output capture
    echo -e "${BLUE}[INFO]${NC} Discovering tests in: $test_dir" >&2
    
    # Find all test-*.sh files
    local test_files=()
    while IFS= read -r -d '' test_file; do
        test_files+=("$test_file")
    done < <(find "$test_dir" -name "test-*.sh" -type f -print0 | sort -z)
    
    if [[ ${#test_files[@]} -eq 0 ]]; then
        echo -e "${YELLOW}[WARNING]${NC} No tests found" >&2
        return 1
    fi
    
    echo -e "${BLUE}[INFO]${NC} Found ${#test_files[@]} test(s)" >&2
    
    printf '%s\n' "${test_files[@]}"
}

# ============================================================================
# Main Test Execution
# ============================================================================

main() {
    log_header "Vikunja Proxmox Deployment - Integration Test Suite"
    
    log_info "Project root: $PROJECT_ROOT"
    log_info "Test directory: $SCRIPT_DIR/integration"
    
    # Discover tests
    local test_files
    mapfile -t test_files < <(discover_tests "$SCRIPT_DIR/integration")
    
    if [[ ${#test_files[@]} -eq 0 ]]; then
        log_error "No tests to run"
        exit 1
    fi
    
    # Run each test
    local passed=0
    local failed=0
    local failed_tests=()
    
    for test_file in "${test_files[@]}"; do
        if run_test "$test_file"; then
            ((passed++))
        else
            ((failed++))
            failed_tests+=("$(basename "$test_file" .sh)")
        fi
    done
    
    # Calculate total time
    local suite_end_time
    suite_end_time=$(date +%s)
    local total_duration=$((suite_end_time - SUITE_START_TIME))
    
    # Print final summary
    log_header "Test Suite Results"
    
    echo -e "${GREEN}Passed:${NC} $passed"
    echo -e "${RED}Failed:${NC} $failed"
    echo -e "${BLUE}Total:${NC}  $((passed + failed))"
    echo -e "${CYAN}Duration:${NC} ${total_duration}s"
    echo ""
    
    if [[ $failed -eq 0 ]]; then
        log_success "All tests passed! ✓"
        echo ""
        exit 0
    else
        log_error "Some tests failed:"
        for test in "${failed_tests[@]}"; do
            echo -e "  ${RED}✗${NC} $test"
        done
        echo ""
        exit 1
    fi
}

# ============================================================================
# Usage
# ============================================================================

show_usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Run the complete integration test suite for Vikunja Proxmox deployment.

OPTIONS:
    -h, --help          Show this help message
    -v, --verbose       Enable verbose output (sets MOCK_DEBUG=1)
    -t, --test NAME     Run only a specific test (e.g., test-fresh-install)

EXAMPLES:
    # Run all tests
    $(basename "$0")
    
    # Run with verbose output
    $(basename "$0") --verbose
    
    # Run specific test
    $(basename "$0") --test test-fresh-install

ENVIRONMENT:
    MOCK_DEBUG=1        Enable mock API debug logging

EOF
}

# ============================================================================
# Argument Parsing
# ============================================================================

SPECIFIC_TEST=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)
            show_usage
            exit 0
            ;;
        -v|--verbose)
            export MOCK_DEBUG=1
            log_info "Verbose mode enabled"
            shift
            ;;
        -t|--test)
            SPECIFIC_TEST="$2"
            log_info "Running specific test: $SPECIFIC_TEST"
            shift 2
            ;;
        *)
            log_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Run specific test if requested
if [[ -n "$SPECIFIC_TEST" ]]; then
    test_file="$SCRIPT_DIR/integration/${SPECIFIC_TEST}.sh"
    
    if [[ ! -f "$test_file" ]]; then
        log_error "Test not found: $test_file"
        exit 1
    fi
    
    log_header "Running Single Test: $SPECIFIC_TEST"
    
    if run_test "$test_file"; then
        log_success "Test passed!"
        exit 0
    else
        log_error "Test failed!"
        exit 1
    fi
fi

# Run main test suite
main "$@"
