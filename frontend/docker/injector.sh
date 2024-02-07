#!/usr/bin/env sh
set -e

echo "info: API URL is $VIKUNJA_API_URL"
echo "info: Sentry enabled: $VIKUNJA_SENTRY_ENABLED"

# Escape the variable to prevent sed from complaining
VIKUNJA_API_URL="$(echo "$VIKUNJA_API_URL" | sed -r 's/([:;])/\\\1/g')"
VIKUNJA_SENTRY_DSN="$(echo "$VIKUNJA_SENTRY_DSN" | sed -r 's/([:;])/\\\1/g')"

sed -ri "s:^(\s*window.API_URL\s*=)\s*.+:\1 '${VIKUNJA_API_URL}':g" /usr/share/nginx/html/index.html
sed -ri "s:^(\s*window.SENTRY_ENABLED\s*=)\s*.+:\1 ${VIKUNJA_SENTRY_ENABLED}:g" /usr/share/nginx/html/index.html
sed -ri "s:^(\s*window.SENTRY_DSN\s*=)\s*.+:\1 '${VIKUNJA_SENTRY_DSN}':g" /usr/share/nginx/html/index.html
sed -ri "s:^(\s*window.PROJECT_INFINITE_NESTING_ENABLED\s*=)\s*.+:\1 '${VIKUNJA_PROJECT_INFINITE_NESTING_ENABLED}':g" /usr/share/nginx/html/index.html
sed -ri "s:^(\s*window.ALLOW_ICON_CHANGES\s*=)\s*.+:\1 ${VIKUNJA_ALLOW_ICON_CHANGES}:g" /usr/share/nginx/html/index.html
sed -ri "s:^(\s*window.CUSTOM_LOGO_URL\s*=)\s*.+:\1 ${VIKUNJA_CUSTOM_LOGO_URL}:g" /usr/share/nginx/html/index.html

date -uIseconds | xargs echo 'info: started at'
