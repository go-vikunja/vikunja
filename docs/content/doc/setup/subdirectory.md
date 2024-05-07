---
title: "Running Vikunja in a subdirectory"
date: 2022-09-23T12:15:04+02:00
draft: false
menu:
  sidebar:
    parent: "setup"
---

# Running Vikunja in a subdirectory

Running Vikunja in a subdirectory is not supported out of the box.
However, you can still run it in a subdirectory but need to build the frontend yourself.

## Frontend

First, make sure you're able to build the frontend from source.
Check [the guide about building from source]({{< ref "build-from-source.md">}}#frontend) about that.

### Dynamically set with build command

Run the build with the `VIKUNJA_FRONTEND_BASE` variable specified.

```
VIKUNJA_FRONTEND_BASE=/SUBPATH/ pnpm run build
```

Where `SUBPATH` is the subdirectory you want to run Vikunja on.

### Set via .env.local

* Copy `.env.local.example` to `.env.local`
* Uncomment `VIKUNJA_FRONTEND_BASE` and set `/subpath/` to the desired path.

After saving, build Vikunja as normal.

```
pnpm run build
```

Once you have the frontend built, you can proceed to build the binary as outlined in [building from source]({{< ref "build-from-source.md">}}#api).

## API

If you're not using a reverse proxy you're good to go.
Simply configure the api url in the frontend as you normally would.

If you're using a reverse proxy you'll need to adjust the paths so that the api is available at `/SUBPATH/api/v1`.
You can check if everything is working correctly by opening `/SUBPATH/api/v1/info` in a browser.
