# Vikunja desktop

[![Build Status](https://drone.kolaente.de/api/badges/vikunja/desktop/status.svg)](https://drone.kolaente.de/vikunja/desktop)

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

TODO
