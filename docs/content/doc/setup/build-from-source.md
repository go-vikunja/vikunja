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

Vikunja being a go application, has no other dependencies than go itself. 
All libraries are bundeled inside the repo in the `vendor/` folder, so all it boils down to are these steps:

1. Make sure [Go](https://golang.org/doc/install) is properly installed on your system. You'll need at least Go `1.9`.
2. Make sure [Make](https://www.gnu.org/software/make/) is properly installed on your system.
3. Clone the repo with `git clone https://code.vikunja.io/api`
3. Run `make build` in the source of this repo. This will build a binary in the root of the repo which will be able to run on your system.

*Note:* Static ressources such as email templates are built into the binary.
For these to work, you may need to run `make generate` before building the vikunja binary.
When builing entirely with `make`, you dont need to do this, `make generate` will be run automatically when running `make build`.

# Build for different architectures

To build for other platforms and architectures than the one you're currently on, simply run `make release` or `make release-{linux|windows|darwin}`.

More options are available, please refer to the [makefile docs]({{< ref "../development/make.md">}}) for more details.