#!/bin/sh

set -e

sed -i "s/\${version}/$VERSION/g" package.json

sed -i "s/\"version\": \".*\"/\"version\": \"$VERSION\"/" package.json

