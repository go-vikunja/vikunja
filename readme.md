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
sed -i 's/\/fonts/.\/fonts/g' frontend/index.html
sed -i 's/\/js/.\/js/g' frontend/index.html      
sed -i 's/\/css/.\/css/g' frontend/index.html    
sed -i 's/\/images/.\/images/g' frontend/index.html
sed -i 's/\/images/.\/images/g' frontend/js/*
sed -i 's/\/css/.\/css/g' frontend/js/*
sed -i 's/\/js/.\/js/g' frontend/js/*
```

## Building for release

TODO
