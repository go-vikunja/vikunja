#!/bin/sh

set -xe

frontend_version=$(sed -n 's/.*"VERSION": "\([^"]*\)".*/\1/p' ./frontend/version.json)

sed -i "s/\${version}/$frontend_version/g" package.json

sed -i "s/\"version\": \".*\"/\"version\": \"$frontend_version\"/" package.json

