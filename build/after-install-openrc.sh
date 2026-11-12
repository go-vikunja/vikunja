#!/bin/sh

# The shipped config is fully commented out, so the secret must be live yaml.
# Skipping an existing service section avoids a duplicate key on upgrades.
if ! grep -q '^service:' /etc/vikunja/config.yml; then
	NEW_SECRET=$(head -c 512 /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c 32)
	printf '\nservice:\n  secret: "%s"\n' "$NEW_SECRET" >> /etc/vikunja/config.yml
fi

sed -i "s/<rootpath>/\/opt\/vikunja\//g" /etc/vikunja/config.yml
sed -i "s/path: \"\.\/vikunja.db\"/path: \"\\/opt\/vikunja\/vikunja.db\"/g" /etc/vikunja/config.yml

rc-update add vikunja default

# Only restart when already running, so fresh installs stay stopped and
# upgrades pick up the new binary.
if rc-service vikunja status >/dev/null 2>&1; then
	rc-service vikunja restart || true
fi
