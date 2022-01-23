---
date: "2019-02-12:00:00+02:00"
title: "Magefile"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "development"
---

# Mage

Vikunja uses [Mage](https://magefile.org/) to script common development tasks and even releasing.
Mage is a pure go solution which allows for greater flexibility and things like better parallelization.

This document explains what tasks are available and what they do.

{{< table_of_contents >}}

## Installation

To use mage, you'll need to install the mage cli.
To install it, run the following command:

```
go install github.com/magefile/mage
```

## Categories

There are multiple categories of subcommands in the magefile:

* `build`: Contains commands to build a single binary
* `check`: Contains commands to statically check the source code 
* `release`: Contains commands to release Vikunja with everything that's required
* `test`: Contains commands to run all kinds of tests
* `dev`: Contains commands to run development tasks
* `misc`: Commands which do not belong in either of the other categories

## CI

These tasks are automatically run in our CI every time someone pushes to main or you update a pull request:

* `mage check:lint`
* `mage check:fmt`
* `mage check:ineffassign`
* `mage check:misspell`
* `mage check:goconst`
* `mage build:generate`
* `mage build:build`

## Build

### Build Vikunja

{{< highlight bash >}}
mage build:build
{{< /highlight >}}

or

{{< highlight bash >}}
mage build
{{< /highlight >}}

Builds a `vikunja`-binary in the root directory of the repo for the platform it is run on.

### clean

{{< highlight bash >}}
mage build:clean
{{< /highlight >}}

Cleans all build and executable files

## Check

All check sub-commands exit with a status code of 1 if the check fails.

Various code-checks are available:

* `mage check:all`: Runs fmt-check, lint, got-swag, misspell-check, ineffasign-check, gocyclo-check, static-check, gosec-check, goconst-check all in parallel
* `mage check:fmt`: Checks if the code is properly formatted with go fmt
* `mage check:go-sec`: Checks the source code for potential security issues by scanning the Go AST using the [gosec tool](https://github.com/securego/gosec)
* `mage check:goconst`: Checks for repeated strings that could be replaced by a constant using [goconst](https://github.com/jgautheron/goconst/)
* `mage check:gocyclo`: Checks for the cyclomatic complexity of the source code using [gocyclo](https://github.com/fzipp/gocyclo)
* `mage check:got-swag`: Checks if the swagger docs need to be re-generated from the code annotations
* `mage check:ineffassign`: Checks the source code for ineffectual assigns using [ineffassign](https://github.com/gordonklaus/ineffassign)
* `mage check:lint`: Runs golint on all packages
* `mage check:misspell`: Checks the source code for misspellings
* `mage check:static`: Statically analyzes the source code about a range of different problems using [staticcheck](https://staticcheck.io/docs/)

## Release

### Build Releases

{{< highlight bash >}}
mage release
{{< /highlight >}}

Builds binaries for all platforms and zips them with a copy of the `templates/` folder.
All built zip files are stored into `dist/zips/`. Binaries are stored in `dist/binaries/`,
binaries bundled with `templates` are stored in `dist/releases/`.

All cross-platform binaries built using this series of commands are built with the help of 
[xgo](https://github.com/techknowlogick/xgo). The mage command will automatically install the
binary to be able to use it.

`mage release:release` is a shortcut to execute `mage release:dirs release:windows release:linux release:darwin release:copy release:check release:os-package release:zip`.

* `mage release:dirs` creates all directories needed
* `mage release:windows`/`release:linux`/`release:darwin` execute xgo to build for their respective platforms
* `mage release:copy` bundles binaries with a copy of the `LICENSE` and sample config files to then be zipped
* `mage release:check` creates sha256 checksums for each binary which will be included in the zip file
* `mage release:os-package` bundles a binary with the `sha256` checksum file, a sample `config.yml` and a copy of the license in a folder for each architecture
* `mage release:compress` compresses all build binaries with `upx` to save space
* `mage release:zip` paclages a zip file for the files created by `release:os-package`

### Build os packages

{{< highlight bash >}}
mage release:packages
{{< /highlight >}}

Will build `.deb`, `.rpm` and `.apk` packages to `dist/os-packages`.

### Make a debian repo

{{< highlight bash >}}
mage release:reprepro
{{< /highlight >}}

Takes an already built debian package and creates a debian repo structure around it.

Used to be run inside a [docker container](https://git.kolaente.de/konrad/reprepro-docker) in the CI process when releasing.

## Test

### unit

{{< highlight bash >}}
mage test:unit
{{< /highlight >}}

Runs all tests except integration tests.

### coverage

{{< highlight bash >}}
mage test:coverage
{{< /highlight >}}

Runs all tests except integration tests and generates a `coverage.html` file to inspect the code coverage.

### integration

{{< highlight bash >}}
mage test:integration
{{< /highlight >}}

Runs all integration tests.

## Dev

### Create a new migration 

{{< highlight bash >}}
mage dev:create-migration
{{< /highlight >}}

Creates a new migration with the current date. 
Will ask for the name of the struct you want to create a migration for.

See also [migration docs]({{< ref "mage.md" >}}).

## Misc

### Format the code

{{< highlight bash >}}
mage fmt
{{< /highlight >}}

Formats all source code using `go fmt`.

### Generate swagger definitions from code comments

{{< highlight bash >}}
mage do-the-swag
{{< /highlight >}}

Generates swagger definitions from the comment annotations in the code.
