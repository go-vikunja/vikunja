#!/bin/bash

# The shipped config is fully commented out, so the secret must be live yaml.
# Skipping an existing service section avoids a duplicate key on upgrades.
if ! grep -q '^service:' /etc/vikunja/config.yml; then
	NEW_SECRET=$(head -c 512 /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c 32)
	printf '\nservice:\n  secret: "%s"\n' "$NEW_SECRET" >> /etc/vikunja/config.yml
fi

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
