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

* `mage lint`
* `mage build:build`

## Build

### Build Vikunja

```
mage build
```

Builds a `vikunja`-binary in the root directory of the repo for the platform it is run on.

### clean

```
mage build:clean
```

Cleans all build and executable files

## Check

All check sub-commands exit with a status code of 1 if the check fails.

Various code-checks are available:

* `mage check:all`: Runs golangci and swagger documentation check
* `mage lint`: Checks if the code follows the rules as defined in the `.golangci.yml` config file.
* `mage lint:fix`: Fixes all code style issues which are easily fixable.

## Release

### Build Releases

```
mage release
```

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
* `mage release:zip` packages a zip file for the files created by `release:os-package`

### Build os packages

```
mage release:packages
```

Will build `.deb`, `.rpm` and `.apk` packages to `dist/os-packages`.

### Make a debian repo

```
mage release:reprepro
```

Takes an already built debian package and creates a debian repo structure around it.

Used to be run inside a [docker container](https://git.kolaente.de/konrad/reprepro-docker) in the CI process when releasing.

## Test

### unit

```
mage test:unit
```

Runs all tests except integration tests.

### coverage

```
mage test:coverage
```

Runs all tests except integration tests and generates a `coverage.html` file to inspect the code coverage.

### integration

```
mage test:integration
```

Runs all integration tests.

## Dev

### Create a new migration 

```
mage dev:create-migration
```

Creates a new migration with the current date.
Will ask for the name of the struct you want to create a migration for.

See also [migration docs]({{< ref "mage.md" >}}).

## Misc

### Format the code

```
mage fmt
```

Formats all source code using `go fmt`.

### Generate swagger definitions from code comments

```
mage do-the-swag
```

Generates swagger definitions from the comment annotations in the code.
