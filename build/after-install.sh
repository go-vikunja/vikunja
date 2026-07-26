#!/bin/bash

systemctl enable vikunja.service

# Fix the config to contain proper values
NEW_SECRET=$(head -c 512 /dev/urandom | tr -dc 'a-zA-Z0-9' | head -c 32)
sed -i "s/<jwt-secret>/$NEW_SECRET/g" /etc/vikunja/config.yml
sed -i "s/<rootpath>/\/opt\/vikunja\//g" /etc/vikunja/config.yml
sed -i "s/path: \"\.\/vikunja.db\"/path: \"\\/opt\/vikunja\/vikunja.db\"/g" /etc/vikunja/config.yml

STATE_DIR=/var/lib/vikunja
CONFIG=/etc/vikunja/config.yml

# dpkg passes "configure <old-version>", where the version is empty on a first
# install; rpm passes 1 for an install and 2 for an upgrade. Anything we don't
# recognise counts as an upgrade, so an odd invocation never rewrites the config.
is_fresh_install() {
	case "${1:-}" in
	configure) [ -z "${2:-}" ] && return 0 ;;
	1) return 0 ;;
	esac
	return 1
}

if ! getent passwd vikunja >/dev/null 2>&1; then
	NOLOGIN=/usr/sbin/nologin
	[ -x "$NOLOGIN" ] || NOLOGIN=/sbin/nologin
	[ -x "$NOLOGIN" ] || NOLOGIN=/bin/false
	useradd --system --user-group --home-dir "$STATE_DIR" --no-create-home \
		--shell "$NOLOGIN" --comment "Vikunja service" vikunja || true
fi

if is_fresh_install "$@" && ! grep -q "^database:" "$CONFIG"; then
	# Every key in the shipped config is commented out, so the database otherwise
	# defaults to sitting next to the binary in /opt/vikunja - which would mean the
	# service user needs write access to the directory holding the binary it runs.
	cat >>"$CONFIG" <<EOF

# Written by the vikunja package on first install, so the service can run as the
# unprivileged vikunja user. Removing these puts the data back in /opt/vikunja.
database:
  path: "$STATE_DIR/vikunja.db"
files:
  basepath: "$STATE_DIR/files"
EOF
fi

# Hand the data to the service user. Installs from before the switch keep theirs in
# /opt/vikunja, where the built-in default puts it.
for p in "$STATE_DIR" /opt/vikunja/vikunja.db /opt/vikunja/vikunja.db-wal \
	/opt/vikunja/vikunja.db-shm /opt/vikunja/files /opt/vikunja/logs; do
	[ -e "$p" ] && chown -R vikunja:vikunja "$p"
done

# ...and any absolute path the admin set explicitly. Commented lines can't match,
# so this only sees keys that are actually in effect.
grep -hE '^[[:space:]]*(path|basepath|dir):[[:space:]]*"?/' "$CONFIG" 2>/dev/null |
	sed -E 's/^[^:]*:[[:space:]]*"?([^"]*[^"[:space:]])"?[[:space:]]*$/\1/' |
	while read -r p; do
		[ -e "$p" ] && chown -R vikunja:vikunja "$p"
	done

# SQLite needs to create its WAL and journal beside the database file, so the
# directory has to be writable too. Only relevant for pre-switch installs whose
# database still lives next to the binary; see the docs for moving it to /var/lib.
if [ -e /opt/vikunja/vikunja.db ]; then
	chgrp vikunja /opt/vikunja && chmod 2775 /opt/vikunja
	echo "vikunja: the database is still in /opt/vikunja, which the service user can now write." >&2
	echo "vikunja: moving it to $STATE_DIR is recommended - see https://vikunja.io/docs/systemd-hardening" >&2
fi

# The config holds the database password and the signing secret.
chown root:vikunja "$CONFIG" && chmod 640 "$CONFIG"

systemctl daemon-reload || true
