#!/usr/bin/env bash
# Vikunja Proxmox Deployment - Bootstrap Installer
# Purpose: Downloads full installer package and executes it
# Usage: bash <(curl -fsSL https://raw.githubusercontent.com/aroige/vikunja/main/deploy/proxmox/vikunja-install-bootstrap.sh)

set -euo pipefail

# Bootstrap version
readonly BOOTSTRAP_VERSION="1.0.0"

# GitHub repository configuration
readonly GITHUB_OWNER="${VIKUNJA_GITHUB_OWNER:-aroige}"
readonly GITHUB_REPO="${VIKUNJA_GITHUB_REPO:-vikunja}"
readonly GITHUB_BRANCH="${VIKUNJA_GITHUB_BRANCH:-main}"
readonly BASE_URL="https://raw.githubusercontent.com/${GITHUB_OWNER}/${GITHUB_REPO}/${GITHUB_BRANCH}/deploy/proxmox"

# Installation directory
readonly INSTALL_DIR="/tmp/vikunja-installer-$$"

# Color codes for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Cleanup function
cleanup() {
    if [[ -d "${INSTALL_DIR}" ]]; then
        rm -rf "${INSTALL_DIR}"
    fi
}

# Set trap for cleanup
trap cleanup EXIT INT TERM

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

# Download a file from GitHub
download_file() {
    local remote_path="$1"
    local local_path="$2"
    local url="${BASE_URL}/${remote_path}"
    
    if ! curl -fsSL "$url" -o "$local_path"; then
        log_error "Failed to download: $remote_path"
        return 1
    fi
}

# Main bootstrap function
main() {
    cat <<'EOF'
╔══════════════════════════════════════════════════════════════╗
║                                                              ║
║              Vikunja Proxmox LXC Deployment                  ║
║                   Bootstrap Installer                        ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
EOF
    
    log_info "Bootstrap version: ${BOOTSTRAP_VERSION}"
    log_info "Repository: ${GITHUB_OWNER}/${GITHUB_REPO}@${GITHUB_BRANCH}"
    echo
    
    # Check prerequisites
    log_info "Checking prerequisites..."
    
    if ! command -v curl &> /dev/null; then
        log_error "curl is not installed. Please install curl and try again."
        exit 1
    fi
    
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root"
        exit 1
    fi
    
    if ! command -v pct &> /dev/null; then
        log_error "This script must be run on a Proxmox VE host"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
    echo
    
    # Create temporary directory
    log_info "Creating temporary installation directory..."
    mkdir -p "${INSTALL_DIR}"/{lib,templates}
    log_success "Created: ${INSTALL_DIR}"
    echo
    
    # Download main installer
    log_info "Downloading installation scripts..."
    
    local files=(
        "vikunja-install-main.sh"
    )
    
    for file in "${files[@]}"; do
        log_info "  → ${file}"
        if ! download_file "$file" "${INSTALL_DIR}/${file}"; then
            log_error "Bootstrap failed: Could not download $file"
            exit 1
        fi
    done
    
    # Download library files
    log_info "Downloading library modules..."
    
    local lib_files=(
        "lib/common.sh"
        "lib/proxmox-api.sh"
        "lib/lxc-setup.sh"
        "lib/service-setup.sh"
        "lib/nginx-setup.sh"
        "lib/health-check.sh"
    )
    
    for file in "${lib_files[@]}"; do
        log_info "  → ${file}"
        if ! download_file "$file" "${INSTALL_DIR}/${file}"; then
            log_error "Bootstrap failed: Could not download $file"
            exit 1
        fi
    done
    
    # Download template files
    log_info "Downloading configuration templates..."
    
    local template_files=(
        "templates/deployment-config.yaml"
        "templates/vikunja-backend.service"
        "templates/vikunja-mcp.service"
        "templates/nginx-vikunja.conf"
        "templates/health-check.sh"
    )
    
    for file in "${template_files[@]}"; do
        log_info "  → ${file}"
        if ! download_file "$file" "${INSTALL_DIR}/${file}"; then
            log_error "Bootstrap failed: Could not download $file"
            exit 1
        fi
    done
    
    log_success "All files downloaded successfully"
    echo
    
    # Make installer executable
    chmod +x "${INSTALL_DIR}/vikunja-install-main.sh"
    
    # Execute the real installer
    log_info "Launching main installer..."
    echo
    echo "═══════════════════════════════════════════════════════════════"
    echo
    
    cd "${INSTALL_DIR}"
    exec ./vikunja-install-main.sh "$@"
}

# Run main function
main "$@"
