#!/bin/sh

set -xe

frontend_version=$(git describe --tags --always --abbrev=10)

sed -i "s/\${version}/$frontend_version/g" package.json

sed -i "s/\"version\": \".*\"/\"version\": \"$frontend_version\"/" package.json
