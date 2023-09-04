#!/bin/sh

set -e

cat package.json | sed "s/\"version\": \".*\"/\"version\": \"$VERSION\"/" > tmp.json
mv tmp.json package.json

