#!/bin/bash

# Fix the config to contain proper values. No-ops on upgrades: the package
# manager keeps the existing config, so the placeholders are already gone.
NEW_SECRET=$(head -c 512 /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c 32)
sed -i "s/<jwt-secret>/$NEW_SECRET/g" /etc/vikunja/config.yml
sed -i "s/<rootpath>/\/opt\/vikunja\//g" /etc/vikunja/config.yml
sed -i "s/path: \"\.\/vikunja.db\"/path: \"\\/opt\/vikunja\/vikunja.db\"/g" /etc/vikunja/config.yml

systemctl enable vikunja.service

# Nothing to reload or restart in chroots and containers without systemd.
if [ -d /run/systemd/system ]; then
	# Pick up changes to vikunja.service itself.
	systemctl daemon-reload || true
	# try-restart is a no-op while the unit is stopped, so fresh installs stay
	# stopped and upgrades pick up the new binary.
	systemctl try-restart vikunja.service || true
fi
