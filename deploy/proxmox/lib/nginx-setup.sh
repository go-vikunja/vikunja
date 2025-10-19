#!/usr/bin/env bash
# Vikunja Proxmox Deployment - Nginx Configuration Functions
# Provides: Nginx reverse proxy setup and management
# Required by: vikunja-install.sh, vikunja-update.sh

set -euo pipefail

# Prevent multiple sourcing
if [[ -n "${VIKUNJA_NGINX_SETUP_LIB_LOADED:-}" ]]; then
    return 0
fi
readonly VIKUNJA_NGINX_SETUP_LIB_LOADED=1

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=./common.sh
source "${SCRIPT_DIR}/common.sh"
# shellcheck source=./proxmox-api.sh
source "${SCRIPT_DIR}/proxmox-api.sh"

# ============================================================================
# Nginx Configuration Functions (T029)
# ============================================================================

# Generate nginx configuration for Vikunja
# Usage: generate_nginx_config ct_id domain backend_port frontend_dir ssl_cert ssl_key
# Returns: 0 on success, 1 on failure
generate_nginx_config() {
    local ct_id="$1"
    local domain="${2:-${DOMAIN:-vikunja.local}}"
    local backend_port="${3:-${BACKEND_BLUE_PORT:-3456}}"
    local frontend_dir="${4:-/opt/vikunja/frontend/dist}"
    local ssl_cert="${5:-}"
    local ssl_key="${6:-}"
    
    log_info "Generating nginx configuration for ${domain}"
    
    local config_file="/etc/nginx/sites-available/vikunja"
    local use_ssl="false"
    
    # Check if SSL certificates are provided
    if [[ -n "$ssl_cert" ]] && [[ -n "$ssl_key" ]]; then
        use_ssl="true"
    fi
    
    # Generate configuration
    local nginx_config
    if [[ "$use_ssl" == "true" ]]; then
        nginx_config=$(cat <<'EOF'
# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name DOMAIN_PLACEHOLDER;
    return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
    listen 443 ssl http2;
    server_name DOMAIN_PLACEHOLDER;
    
    ssl_certificate SSL_CERT_PLACEHOLDER;
    ssl_certificate_key SSL_KEY_PLACEHOLDER;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    
    client_max_body_size 20M;
    
    # Frontend static files
    location / {
        root FRONTEND_DIR_PLACEHOLDER;
        try_files $uri $uri/ /index.html;
    }
    
    # API backend proxy
    location /api/ {
        proxy_pass http://localhost:BACKEND_PORT_PLACEHOLDER;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # WebSocket support for real-time updates
    location /api/v1/websocket {
        proxy_pass http://localhost:BACKEND_PORT_PLACEHOLDER;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }
}
EOF
)
    else
        nginx_config=$(cat <<'EOF'
server {
    listen 80;
    server_name DOMAIN_PLACEHOLDER;
    
    client_max_body_size 20M;
    
    # Frontend static files
    location / {
        root FRONTEND_DIR_PLACEHOLDER;
        try_files $uri $uri/ /index.html;
    }
    
    # API backend proxy
    location /api/ {
        proxy_pass http://localhost:BACKEND_PORT_PLACEHOLDER;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # WebSocket support
    location /api/v1/websocket {
        proxy_pass http://localhost:BACKEND_PORT_PLACEHOLDER;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }
}
EOF
)
    fi
    
    # Replace placeholders
    nginx_config="${nginx_config//DOMAIN_PLACEHOLDER/$domain}"
    nginx_config="${nginx_config//BACKEND_PORT_PLACEHOLDER/$backend_port}"
    nginx_config="${nginx_config//FRONTEND_DIR_PLACEHOLDER/$frontend_dir}"
    nginx_config="${nginx_config//SSL_CERT_PLACEHOLDER/$ssl_cert}"
    nginx_config="${nginx_config//SSL_KEY_PLACEHOLDER/$ssl_key}"
    
    # Write configuration to container
    pct_exec "$ct_id" bash -c "cat > ${config_file} <<'NGINXEOF'
${nginx_config}
NGINXEOF
" || return 1
    
    log_success "Nginx configuration created"
    return 0
}

# Enable nginx site
# Usage: enable_site ct_id
# Returns: 0 on success, 1 on failure
enable_site() {
    local ct_id="$1"
    
    log_info "Enabling nginx site"
    
    # Remove default site
    pct_exec "$ct_id" rm -f /etc/nginx/sites-enabled/default 2>/dev/null || true
    
    # Enable vikunja site
    pct_exec "$ct_id" ln -sf /etc/nginx/sites-available/vikunja /etc/nginx/sites-enabled/vikunja \
        2>&1 | tee >(log_debug) || return 1
    
    log_success "Site enabled"
    return 0
}

# Reload nginx
# Usage: reload_nginx ct_id
# Returns: 0 on success, 1 on failure
reload_nginx() {
    local ct_id="$1"
    
    log_info "Reloading nginx"
    
    # Test configuration
    if ! pct_exec "$ct_id" nginx -t 2>&1 | tee >(log_debug); then
        log_error "Nginx configuration test failed"
        return 1
    fi
    
    # Reload nginx
    if ! pct_exec "$ct_id" systemctl reload nginx 2>&1 | tee >(log_debug); then
        log_error "Failed to reload nginx"
        return 1
    fi
    
    log_success "Nginx reloaded"
    return 0
}

# ============================================================================
# Blue-Green Nginx Functions (for T052 - User Story 2)
# ============================================================================

# Update nginx upstream to point to active color
# Usage: update_nginx_upstream ct_id active_color blue_port green_port
# Returns: 0 on success, 1 on failure
update_nginx_upstream() {
    local ct_id="$1"
    local active_color="$2"
    local blue_port="$3"
    local green_port="$4"
    
    local active_port
    if [[ "$active_color" == "blue" ]]; then
        active_port="$blue_port"
    else
        active_port="$green_port"
    fi
    
    log_info "Updating nginx upstream to ${active_color} (port ${active_port})"
    
    # Update backend port in configuration
    pct_exec "$ct_id" sed -i "s/proxy_pass http:\/\/localhost:[0-9]\+/proxy_pass http:\/\/localhost:${active_port}/" \
        /etc/nginx/sites-available/vikunja 2>&1 | tee >(log_debug) || return 1
    
    # Test and reload
    if ! test_nginx_config "$ct_id"; then
        log_error "Nginx configuration test failed after upstream update"
        return 1
    fi
    
    reload_nginx "$ct_id" || return 1
    
    log_success "Nginx upstream updated to ${active_color}"
    return 0
}

# Test nginx configuration
# Usage: test_nginx_config ct_id
# Returns: 0 if valid, 1 if invalid
test_nginx_config() {
    local ct_id="$1"
    
    log_debug "Testing nginx configuration"
    
    if pct_exec "$ct_id" nginx -t >/dev/null 2>&1; then
        return 0
    fi
    
    return 1
}

log_debug "Nginx setup library loaded"
