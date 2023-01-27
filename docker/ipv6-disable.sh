#!/usr/bin/env sh
set -e

if [ ! -f "/proc/net/if_inet6" ]; then
  echo "info: IPv6 is not available! Removing IPv6 listen configuration"
  find /etc/nginx/conf.d -name '*.conf' -type f | \
  while IFS= read -r CONFIG; do
    sed -r '/^\s*listen\s*\[::\]:.+$/d' "$CONFIG" > "$CONFIG.temp"
    if ! diff -U 5 "$CONFIG" "$CONFIG.temp" > "$CONFIG.diff"; then
      echo "info: Removing IPv6 lines from $CONFIG" | \
      cat - "$CONFIG.diff"
      echo "# IPv6 is disabled because /proc/net/if_inet6 was not found" | \
      cat - "$CONFIG.temp" > "$CONFIG"
    else
      echo "info: Skipping $CONFIG because it does not have IPv6 listen"
    fi
    rm -f "$CONFIG.temp" "$CONFIG.diff"
  done
fi
