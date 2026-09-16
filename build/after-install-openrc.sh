#!/bin/sh

# Fix the config to contain proper values. No-ops on upgrades: the package
# manager keeps the existing config, so the placeholders are already gone.
NEW_SECRET=$(head -c 512 /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c 32)
sed -i "s/<jwt-secret>/$NEW_SECRET/g" /etc/vikunja/config.yml
sed -i "s/<rootpath>/\/opt\/vikunja\//g" /etc/vikunja/config.yml
sed -i "s/path: \"\.\/vikunja.db\"/path: \"\\/opt\/vikunja\/vikunja.db\"/g" /etc/vikunja/config.yml

rc-update add vikunja default

# Only restart when already running, so fresh installs stay stopped and
# upgrades pick up the new binary.
if rc-service vikunja status >/dev/null 2>&1; then
	rc-service vikunja restart || true
fi
