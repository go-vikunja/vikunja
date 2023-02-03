#!/usr/bin/env sh
set -e

DEFAULT_CONF_FILE="etc/nginx/conf.d/default.conf"

if [ -f "/proc/net/if_inet6" ]; then
    echo "info: IPv6 available."
    exit 0
fi

echo "info: IPv6 not available!"
echo "info: Removing IPv6 lines from /$DEFAULT_CONF_FILE"
sed -i 's/\(listen\s*\[::\].*\)$/#\1 # Disabled IPv6/' /${DEFAULT_CONF_FILE}
