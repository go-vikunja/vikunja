#!/usr/bin/env sh
set -e

if [ -n "$PUID" ] && [ "$PUID" -ne 0 ] && \
   [ -n "$PGID" ] && [ "$PGID" -ne 0 ] ; then
  echo "info: creating the new user vikunja with $PUID:$PGID"
  addgroup -g "$PGID" vikunja
  adduser -s /bin/sh -D -G vikunja -u "$PUID" vikunja -h /app/vikunja -H
  chown -R vikunja:vikunja ./
  su -pc /app/vikunja/vikunja - vikunja "$@"
else
  echo "info: creation of non-root user is skipped"
  exec /app/vikunja/vikunja "$@"
fi

