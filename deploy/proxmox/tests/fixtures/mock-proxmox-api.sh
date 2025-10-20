#!/usr/bin/env bash
# Mock Proxmox API for Integration Testing
# Purpose: Simulate Proxmox VE environment for CI/CD testing
# Usage: source this file before running tests

set -euo pipefail

# ============================================================================
# Mock Configuration
# ============================================================================

# Mock state directory
export MOCK_STATE_DIR="${MOCK_STATE_DIR:-/tmp/vikunja-test-$$}"
mkdir -p "$MOCK_STATE_DIR"

# Mock containers database
export MOCK_CONTAINERS_DB="${MOCK_STATE_DIR}/containers.db"
touch "$MOCK_CONTAINERS_DB"

# Mock network state
export MOCK_NETWORK_READY=true
export MOCK_BUILD_SUCCESS=true
export MOCK_SERVICE_START_SUCCESS=true

# Debug mode
export MOCK_DEBUG="${MOCK_DEBUG:-0}"

# ============================================================================
# Mock Logging
# ============================================================================

mock_log() {
    if [[ "${MOCK_DEBUG}" == "1" ]]; then
        echo "[MOCK] $*" >&2
    fi
}

# ============================================================================
# Mock Proxmox Commands
# ============================================================================

# Mock pct (Proxmox Container Toolkit)
pct() {
    local cmd="$1"
    shift
    
    mock_log "pct $cmd $*"
    
    case "$cmd" in
        create)
            local ct_id="$1"
            local template="$2"
            shift 2
            
            # Check if container already exists
            if grep -q "^${ct_id}:" "$MOCK_CONTAINERS_DB" 2>/dev/null; then
                echo "Container $ct_id already exists" >&2
                return 1
            fi
            
            # Create container record
            echo "${ct_id}:stopped:${template}:$(date +%s)" >> "$MOCK_CONTAINERS_DB"
            mock_log "Created container $ct_id from $template"
            return 0
            ;;
            
        start)
            local ct_id="$1"
            
            # Check if container exists
            if ! grep -q "^${ct_id}:" "$MOCK_CONTAINERS_DB" 2>/dev/null; then
                echo "Container $ct_id does not exist" >&2
                return 1
            fi
            
            # Update status to running
            sed -i "s/^${ct_id}:stopped:/${ct_id}:running:/" "$MOCK_CONTAINERS_DB"
            mock_log "Started container $ct_id"
            return 0
            ;;
            
        stop)
            local ct_id="$1"
            
            # Update status to stopped
            sed -i "s/^${ct_id}:running:/${ct_id}:stopped:/" "$MOCK_CONTAINERS_DB"
            mock_log "Stopped container $ct_id"
            return 0
            ;;
            
        status)
            local ct_id="$1"
            
            # Get container status
            local status
            status=$(grep "^${ct_id}:" "$MOCK_CONTAINERS_DB" 2>/dev/null | cut -d: -f2 || echo "")
            
            if [[ -z "$status" ]]; then
                echo "Configuration file 'nodes/localhost/lxc/${ct_id}.conf' does not exist" >&2
                return 1
            fi
            
            echo "status: $status"
            return 0
            ;;
            
        exec)
            local ct_id="$1"
            shift
            # Skip "--" separator if present
            if [[ "$1" == "--" ]]; then
                shift
            fi
            
            mock_log "Executing in container $ct_id: $*"
            
            # Mock command execution based on common patterns
            case "$1" in
                apt-get)
                    # Simulate apt-get operations
                    if [[ "$2" == "update" ]]; then
                        echo "Reading package lists..." >&2
                        return 0
                    elif [[ "$2" == "install" ]]; then
                        echo "Installing packages..." >&2
                        return 0
                    fi
                    ;;
                    
                wget|curl)
                    # Simulate downloads
                    mock_log "Simulating download: $*"
                    return 0
                    ;;
                    
                git)
                    if [[ "$2" == "clone" ]]; then
                        # Simulate git clone
                        local target_dir="${*: -1}"
                        mkdir -p "${MOCK_STATE_DIR}/container-${ct_id}${target_dir}"
                        echo "abc123def"  # Return mock commit hash
                        return 0
                    elif [[ "$2" == "-C" ]] && [[ "$4" == "rev-parse" ]]; then
                        # Return mock commit hash
                        echo "abc123def"
                        return 0
                    fi
                    ;;
                    
                bash)
                    # Execute bash commands (simplified simulation)
                    if [[ "$2" == "-c" ]]; then
                        local bash_cmd="$3"
                        
                        # Simulate Go installation
                        if [[ "$bash_cmd" =~ "go.tar.gz" ]]; then
                            mock_log "Simulating Go installation"
                            return 0
                        fi
                        
                        # Simulate Node.js installation
                        if [[ "$bash_cmd" =~ "nodejs" ]]; then
                            mock_log "Simulating Node.js installation"
                            return 0
                        fi
                        
                        # Simulate database operations
                        if [[ "$bash_cmd" =~ "psql" ]] || [[ "$bash_cmd" =~ "mysql" ]]; then
                            if [[ "${MOCK_DB_CONNECTION_SUCCESS:-true}" == "true" ]]; then
                                return 0
                            else
                                return 1
                            fi
                        fi
                        
                        # Simulate mage build
                        if [[ "$bash_cmd" =~ "mage build" ]]; then
                            if [[ "${MOCK_BUILD_SUCCESS}" == "true" ]]; then
                                # Create mock binary
                                local work_dir=$(echo "$bash_cmd" | grep -oP 'cd \K[^ ]+' | head -1)
                                mkdir -p "${MOCK_STATE_DIR}/container-${ct_id}${work_dir}"
                                touch "${MOCK_STATE_DIR}/container-${ct_id}${work_dir}/vikunja"
                                return 0
                            else
                                return 1
                            fi
                        fi
                        
                        # Simulate pnpm operations
                        if [[ "$bash_cmd" =~ "pnpm" ]]; then
                            if [[ "${MOCK_BUILD_SUCCESS}" == "true" ]]; then
                                return 0
                            else
                                return 1
                            fi
                        fi
                    fi
                    ;;
                    
                test)
                    # Simulate file existence checks
                    if [[ "$2" == "-f" ]] || [[ "$2" == "-d" ]]; then
                        local path="$3"
                        
                        # Check in mock container filesystem
                        if [[ -e "${MOCK_STATE_DIR}/container-${ct_id}${path}" ]]; then
                            return 0
                        fi
                        
                        # Always succeed for certain paths
                        if [[ "$path" =~ vikunja$ ]] || [[ "$path" =~ dist$ ]]; then
                            return 0
                        fi
                        
                        return 1
                    fi
                    ;;
                    
                mkdir|rm|tar|chmod|chown)
                    # Always succeed for file operations
                    mock_log "Simulating file operation: $*"
                    return 0
                    ;;
                    
                systemctl)
                    # Mock systemctl operations
                    case "$2" in
                        daemon-reload|enable|start|stop|restart|reload|reload-or-restart)
                            if [[ "${MOCK_SERVICE_START_SUCCESS}" == "true" ]]; then
                                return 0
                            else
                                return 1
                            fi
                            ;;
                        is-active)
                            if [[ "$3" == "--quiet" ]]; then
                                return 0
                            else
                                echo "active"
                                return 0
                            fi
                            ;;
                        status)
                            echo "● $3 - Mock Service"
                            echo "   Loaded: loaded"
                            echo "   Active: active (running)"
                            return 0
                            ;;
                    esac
                    ;;
                    
                curl)
                    # Mock health check endpoints
                    if [[ "$*" =~ "/health" ]] || [[ "$*" =~ "localhost" ]]; then
                        return 0
                    fi
                    ;;
                    
                ss)
                    # Mock socket statistics - return all Vikunja service ports
                    if [[ "$*" =~ "-tuln" ]]; then
                        echo "tcp   LISTEN 0  128  127.0.0.1:3456  0.0.0.0:*"
                        echo "tcp   LISTEN 0  128  127.0.0.1:3457  0.0.0.0:*"
                        echo "tcp   LISTEN 0  128  127.0.0.1:3458  0.0.0.0:*"
                        echo "tcp   LISTEN 0  128  127.0.0.1:3459  0.0.0.0:*"
                        return 0
                    fi
                    ;;
                    
                ip)
                    # Mock IP address information
                    if [[ "$*" =~ "addr show" ]]; then
                        echo "inet 192.168.1.100/24"
                        return 0
                    fi
                    ;;
                    
                nginx)
                    # Mock nginx operations
                    if [[ "$2" == "-t" ]]; then
                        echo "nginx: configuration file test is successful"
                        return 0
                    fi
                    ;;
                    
                /usr/local/go/bin/go)
                    if [[ "$2" == "version" ]]; then
                        echo "go version go1.21.5 linux/amd64"
                        return 0
                    fi
                    ;;
                    
                node)
                    if [[ "$2" == "--version" ]]; then
                        echo "v18.19.0"
                        return 0
                    fi
                    # Mock Node.js execution
                    return 0
                    ;;
                    
                npm)
                    # Mock npm operations
                    return 0
                    ;;
                    
                cat)
                    # Mock reading systemd files
                    if [[ "$2" == ">" ]]; then
                        # Writing file - just succeed
                        return 0
                    fi
                    ;;
            esac
            
            # Default: succeed
            mock_log "Default success for: $*"
            return 0
            ;;
            
        set)
            local ct_id="$1"
            shift
            
            # Just succeed - we're not tracking detailed config
            mock_log "Setting container $ct_id config: $*"
            return 0
            ;;
            
        destroy)
            local ct_id="$1"
            
            # Remove container from database
            sed -i "/^${ct_id}:/d" "$MOCK_CONTAINERS_DB"
            
            # Remove mock filesystem
            rm -rf "${MOCK_STATE_DIR}/container-${ct_id}"
            
            mock_log "Destroyed container $ct_id"
            return 0
            ;;
            
        push|pull)
            # Mock file transfers
            mock_log "File transfer: pct $cmd $*"
            return 0
            ;;
            
        config)
            local ct_id="$1"
            # Return mock config
            echo "net0: name=eth0,bridge=vmbr0,ip=192.168.1.100/24"
            return 0
            ;;
            
        *)
            mock_log "Unknown pct command: $cmd"
            return 1
            ;;
    esac
}

# Mock pvesh (Proxmox VE Shell)
pvesh() {
    local action="$1"
    local path="$2"
    shift 2
    
    mock_log "pvesh $action $path $*"
    
    case "$action" in
        get)
            case "$path" in
                /nodes)
                    echo '{"data":[{"node":"pve","status":"online"}]}'
                    return 0
                    ;;
                /nodes/*)
                    echo '{"data":{"memory":{"free":8589934592,"total":17179869184},"cpu":0.05}}'
                    return 0
                    ;;
                /cluster/resources)
                    # Return list of existing containers
                    echo '{"data":[{"vmid":100,"type":"lxc"},{"vmid":101,"type":"lxc"}]}'
                    return 0
                    ;;
            esac
            ;;
    esac
    
    return 0
}

# Mock tput for color support
tput() {
    case "$1" in
        sgr0) echo -n "" ;;
        bold) echo -n "" ;;
        setaf) echo -n "" ;;
        *) echo -n "" ;;
    esac
}

# ============================================================================
# Mock System Commands
# ============================================================================

# Mock ping
ping() {
    mock_log "ping $*"
    return 0
}

# Mock ss (socket statistics)
ss() {
    if [[ "$*" =~ "-tuln" ]]; then
        # Simulate listening ports for Vikunja services
        cat <<EOF
Netid  State   Recv-Q  Send-Q    Local Address:Port      Peer Address:Port
tcp    LISTEN  0       128             0.0.0.0:22             0.0.0.0:*
tcp    LISTEN  0       128             0.0.0.0:80             0.0.0.0:*
tcp    LISTEN  0       128       127.0.0.1:3456             0.0.0.0:*
tcp    LISTEN  0       128       127.0.0.1:3457             0.0.0.0:*
tcp    LISTEN  0       128       127.0.0.1:3458             0.0.0.0:*
tcp    LISTEN  0       128       127.0.0.1:3459             0.0.0.0:*
EOF
        return 0
    fi
    return 0
}

# Export mock functions
export -f pct
export -f pvesh
export -f tput
export -f ping
export -f ss

# ============================================================================
# Mock Proxmox Environment
# ============================================================================

# Create mock Proxmox version file
mkdir -p /tmp/mock-pve
export MOCK_PVE_VERSION="/tmp/mock-pve/.version"
echo "7.4-1" > "$MOCK_PVE_VERSION"

# Mock /etc/pve/.version check
if [[ ! -e /etc/pve/.version ]]; then
    # Create symlink if we have permissions, otherwise just document
    if [[ -w /etc ]]; then
        sudo mkdir -p /etc/pve 2>/dev/null || true
        sudo ln -sf "$MOCK_PVE_VERSION" /etc/pve/.version 2>/dev/null || true
    fi
fi

# ============================================================================
# Helper Functions for Tests
# ============================================================================

# Reset mock state
mock_reset() {
    rm -rf "$MOCK_STATE_DIR"
    mkdir -p "$MOCK_STATE_DIR"
    touch "$MOCK_CONTAINERS_DB"
    
    export MOCK_NETWORK_READY=true
    export MOCK_BUILD_SUCCESS=true
    export MOCK_SERVICE_START_SUCCESS=true
    export MOCK_DB_CONNECTION_SUCCESS=true
    
    mock_log "Mock state reset"
}

# Set mock failure scenario
mock_fail_build() {
    export MOCK_BUILD_SUCCESS=false
}

mock_fail_service() {
    export MOCK_SERVICE_START_SUCCESS=false
}

mock_fail_db() {
    export MOCK_DB_CONNECTION_SUCCESS=false
}

# Get mock container count
mock_container_count() {
    wc -l < "$MOCK_CONTAINERS_DB" 2>/dev/null || echo "0"
}

# Check if mock container exists
mock_container_exists() {
    local ct_id="$1"
    grep -q "^${ct_id}:" "$MOCK_CONTAINERS_DB" 2>/dev/null
}

# Cleanup mock state
mock_cleanup() {
    rm -rf "$MOCK_STATE_DIR"
    rm -f "$MOCK_PVE_VERSION"
}

# Export helper functions
export -f mock_reset
export -f mock_fail_build
export -f mock_fail_service
export -f mock_fail_db
export -f mock_container_count
export -f mock_container_exists
export -f mock_cleanup

mock_log "Mock Proxmox API loaded (state dir: $MOCK_STATE_DIR)"
