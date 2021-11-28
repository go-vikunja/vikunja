---
date: "2019-02-12:00:00+02:00"
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

1. Make sure [Go](https://golang.org/doc/install) is properly installed on your system. You'll need at least Go `1.17`.
2. Make sure [Mage](https://magefile) is properly installed on your system.
3. Clone the repo with `git clone https://code.vikunja.io/api` and switch into the directory.
3. Run `mage build:build` in the source of this repo. This will build a binary in the root of the repo which will be able to run on your system.

*Note:* Static ressources such as email templates are built into the binary.
For these to work, you may need to run `mage build:generate` before building the vikunja binary.
When builing entirely with `mage`, you dont need to do this, `mage build:generate` will be run automatically when running `mage build:build`.

### Build for different architectures

To build for other platforms and architectures than the one you're currently on, simply run `mage release:release` or `mage release:{linux|windows|darwin}`.

More options are available, please refer to the [magefile docs]({{< ref "../development/mage.md">}}) for more details.

## Frontend

The code for the frontend is located at [code.vikunja.io/frontend](https://code.vikunja.io/frontend).

You need to have yarn v1 and nodejs in version 16 installed.

1. Make sure [yarn v1](https://yarnpkg.com/getting-started/install) is properly installed on your system.
3. Clone the repo with `git clone https://code.vikunja.io/frontend` and switch into the directory.
3. Install all dependencies with `yarn install`
4. Build the frontend with `yarn build`. This will result in a js bundle in the `dist/` folder which you can deploy.
