#!/usr/bin/env bash
# Vikunja Health Check Script
# Deployed to container at /usr/local/bin/vikunja-health-check
# Called by: monitoring systems, manual checks

set -euo pipefail

# Configuration
BACKEND_PORT="${BACKEND_PORT:-3456}"
MCP_PORT="${MCP_PORT:-8456}"
FRONTEND_PORT="${FRONTEND_PORT:-80}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Exit codes
EXIT_OK=0
EXIT_WARNING=1
EXIT_CRITICAL=2
EXIT_UNKNOWN=3

# Health check functions
check_backend() {
    if curl -sf "http://localhost:${BACKEND_PORT}/health" >/dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Backend (port ${BACKEND_PORT}): Healthy"
        return 0
    else
        echo -e "${RED}✗${NC} Backend (port ${BACKEND_PORT}): Unhealthy"
        return 1
    fi
}

check_mcp() {
    if ss -tuln | grep -q ":${MCP_PORT} "; then
        echo -e "${GREEN}✓${NC} MCP Server (port ${MCP_PORT}): Healthy"
        return 0
    else
        echo -e "${YELLOW}⚠${NC} MCP Server (port ${MCP_PORT}): Not running"
        return 1
    fi
}

check_frontend() {
    if curl -sf "http://localhost:${FRONTEND_PORT}/" >/dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Frontend (port ${FRONTEND_PORT}): Healthy"
        return 0
    else
        echo -e "${RED}✗${NC} Frontend (port ${FRONTEND_PORT}): Unhealthy"
        return 1
    fi
}

check_database() {
    # Check if backend can connect to database
    if curl -sf "http://localhost:${BACKEND_PORT}/api/v1/info" >/dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Database: Connected"
        return 0
    else
        echo -e "${RED}✗${NC} Database: Connection failed"
        return 1
    fi
}

# Main execution
main() {
    echo "Vikunja Health Check"
    echo "===================="
    echo ""
    
    local exit_code=$EXIT_OK
    
    # Run checks
    if ! check_backend; then
        exit_code=$EXIT_CRITICAL
    fi
    
    if ! check_mcp; then
        # MCP is non-critical
        if [[ $exit_code -eq $EXIT_OK ]]; then
            exit_code=$EXIT_WARNING
        fi
    fi
    
    if ! check_frontend; then
        exit_code=$EXIT_CRITICAL
    fi
    
    if ! check_database; then
        exit_code=$EXIT_CRITICAL
    fi
    
    echo ""
    case $exit_code in
        $EXIT_OK)
            echo -e "${GREEN}Status: All systems healthy${NC}"
            ;;
        $EXIT_WARNING)
            echo -e "${YELLOW}Status: Non-critical issues detected${NC}"
            ;;
        $EXIT_CRITICAL)
            echo -e "${RED}Status: Critical issues detected${NC}"
            ;;
    esac
    
    exit $exit_code
}

main "$@"
