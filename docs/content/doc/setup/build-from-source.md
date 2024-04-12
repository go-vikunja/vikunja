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

To fully build Vikunja from source files, you need to build the api and frontend.

{{< table_of_contents >}}

## General Preparations

1. Make sure you have git installed
2. Clone the repo with `git clone https://code.vikunja.io/vikunja` and switch into the directory.
3. Check out the version you want to build with `git checkout VERSION` - replace `VERSION` with the version want to use. If you don't do this, you'll build the [latest unstable build]({{< ref "versions.md">}}), which might contain bugs.

## Frontend

The code for the frontend is located in the `frontend/` sub folder of the main repo.

1. Make sure you have [pnpm](https://pnpm.io/installation) properly installed on your system.
2. Install all dependencies with `pnpm install`
3. Build the frontend with `pnpm run build`. This will result in a static js bundle in the `dist/` folder.
4. You can either deploy that static js bundle directly, or read on to learn how to bundle it all up in a static binary with the api.

## API

The Vikunja API has no other dependencies than go itself.
That means compiling it boils down to these steps:

1. Make sure [Go](https://golang.org/doc/install) is properly installed on your system. You'll need at least Go `1.21`.
2. Make sure [Mage](https://magefile.org) is properly installed on your system.
3. If you did not build the frontend in the steps before, you need to either do that or create a dummy index file with `mkdir -p frontend/dist && touch frontend/dist/index.html`.
4. Run `mage build` in the source of the main repo. This will build a binary in the root of the repo which will be able to run on your system.

### Build for different architectures

To build for other platforms and architectures than the one you're currently on, simply run `mage release` or `mage release:{linux|windows|darwin}`.

More options are available, please refer to the [magefile docs]({{< ref "../development/mage.md">}}) for more details.
