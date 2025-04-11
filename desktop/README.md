# Vikunja desktop

[![Build Status](https://github.com/go-vikunja/vikunja/actions/workflows/ci.yml/badge.svg)](https://github.com/go-vikunja/vikunja/actions/workflows/ci.yml)
[![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](LICENSE)
[![Download](https://img.shields.io/badge/download-v0.22.1-brightgreen.svg)](https://dl.vikunja.io)

The Vikunja frontend all repackaged as an electron app to run as a desktop app!

## Dev

As this repo does not contain any code, only a thin wrapper around electron, you will need to do this to get the 
actual frontend bundle and build the app:

```bash
rm -rf frontend vikunja-frontend-master.zip 
wget https://dl.vikunja.io/frontend/vikunja-frontend-master.zip
unzip vikunja-frontend-master.zip -d frontend
sed -i 's/\/api\/v1//g' frontend/index.html # Make sure to trigger the "enter the Vikunja url" prompt
```

## Building for release

1. Run the snippet from above, but with a valid frontend version instead of `master`
2. Change the version in `package.json` (That's the one that will be used by electron-builder`
3. `yarn install`
4. `yarn dist --linux --windows`

