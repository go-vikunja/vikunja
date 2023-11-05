---
date: "2022-09-21:00:00+02:00"
title: "Build from sources"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "setup"
---

# Build Vikunja from source

To completely build Vikunja from source, you need to build the api and frontend.

{{< table_of_contents >}}

## API

The Vikunja API has no other dependencies than go itself.
That means compiling it boils down to these steps:

1. Make sure [Go](https://golang.org/doc/install) is properly installed on your system. You'll need at least Go `1.21`.
2. Make sure [Mage](https://magefile.org) is properly installed on your system.
3. Clone the repo with `git clone https://code.vikunja.io/api` and switch into the directory.
4. Run `mage build` in the source of this repo. This will build a binary in the root of the repo which will be able to run on your system.

### Build for different architectures

To build for other platforms and architectures than the one you're currently on, simply run `mage release:release` or `mage release:{linux|windows|darwin}`.

More options are available, please refer to the [magefile docs]({{< ref "../development/mage.md">}}) for more details.

## Frontend

The code for the frontend is located at [code.vikunja.io/frontend](https://code.vikunja.io/frontend).

1. Make sure you have [pnpm](https://pnpm.io/installation) properly installed on your system.
2. Clone the repo with `git clone https://code.vikunja.io/frontend` and switch into the directory.
3. Install all dependencies with `pnpm install`
4. Build the frontend with `pnpm run build`. This will result in a static js bundle in the `dist/` folder which you can deploy.
