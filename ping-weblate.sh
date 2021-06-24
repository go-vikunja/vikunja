#!/bin/sh

set -e

# Shell script because yaml doesn't understand the header is a string literal and not a yaml symbol

curl -d operation=pull -H "Authorization: Token $WEBLATE_TOKEN" https://hosted.weblate.org/api/projects/vikunja/repository/
curl -d operation=push -H "Authorization: Token $WEBLATE_TOKEN" https://hosted.weblate.org/api/projects/vikunja/repository/
