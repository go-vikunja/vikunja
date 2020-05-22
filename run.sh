#!/bin/bash

# This shell script sets the api url based on an environment variable and starts nginx in foreground.

if [ -z "$VIKUNJA_API_URL" ]; then
  VIKUNJA_API_URL="/api/v1"
fi

# Escape the variable to prevent sed from complaining
VIKUNJA_API_URL=$(echo $VIKUNJA_API_URL |sed 's/\//\\\//g')

sed -i "s/http\:\/\/localhost\:3456\/api\/v1/$VIKUNJA_API_URL/g" /usr/share/nginx/html/index.html

# Set the uid and gid of the nginx run user
usermod --non-unique --uid ${PUID} nginx
groupmod --non-unique --gid ${PGID} nginx

nginx -g "daemon off;"
