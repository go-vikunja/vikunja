---
date: "2019-02-12:00:00+02:00"
title: "Makefile"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "development"
---

# Makefile

We scripted a lot of tasks used mostly for developing into the makefile. This documents explains what
taks are available and what they do.

## CI

These tasks are automatically run in our CI every time someone pushes to master or you update a pull request:

* `make lint`
* `make fmt-check`
* `make ineffassign-check`
* `make misspell-check`
* `make goconst-check`
* `make build`

### clean

{{< highlight bash >}}
make clean
{{< /highlight >}}

Clears all builds and binaries.

### test

{{< highlight bash >}}
make test
{{< /highlight >}}

Runs all tests in Vikunja.

### Format the code

{{< highlight bash >}}
make fmt
{{< /highlight >}}

Formats all source code using `go fmt`.

#### Check formatting

{{< highlight bash >}}
make fmt-check
{{< /highlight >}}

Checks if the code needs to be formatted. Fails if it does.

### Build Vikunja

{{< highlight bash >}}
make build
{{< /highlight >}}

Builds a `vikunja`-binary in the root directory of the repo for the platform it is run on.

### Build Releases

{{< highlight bash >}}
make build
{{< /highlight >}}

Builds binaries for all platforms and zips them with a copy of the `templates/` folder.
All built zip files are stored into `dist/zips/`. Binaries are stored in `dist/binaries/`,
binaries bundled with `templates` are stored in `dist/releases/`.

All cross-platform binaries built using this series of commands are built with the help of 
[xgo](https://github.com/karalabe/xgo). The make command will automatically install the
binary to be able to use it.

`make release` is actually just a shortcut to execute `make release-dirs release-windows release-linux release-darwin release-copy release-check release-os-package release-zip`.

* `release-dirs` creates all directories needed
* `release-windows`/`release-linux`/`release-darwin` execute xgo to build for their respective platforms
* `release-copy` bundles binaries with a copy of `templates/` to then be zipped
* `release-check` creates sha256 checksums for each binary which will be included in the zip file
* `release-os-package` bundles a binary with a copy of the `templates/` folder, the `sha256` checksum file, a sample `config.yml` and a copy of the license in a folder for each architecture
* `release-zip` makes a zip file for the files created by `release-os-package`

### Build debian packages

{{< highlight bash >}}
make build-deb
{{< /highlight >}}

Will build a `.deb` package into the current folder. You need to have [fpm](https://fpm.readthedocs.io/en/latest/intro.html) installed to be able to do this.

#### Make a debian repo

{{< highlight bash >}}
make reprepro
{{< /highlight >}}

Takes an already built debian package and creates a debian repo structure around it.

Used to be run inside a [docker container](https://git.kolaente.de/konrad/reprepro-docker) in the CI process when releasing.

### Generate swagger definitions from code comments

{{< highlight bash >}}
make do-the-swag
{{< /highlight >}}

Generates swagger definitions from the comments in the code.

#### Check if swagger generation is needed

{{< highlight bash >}}
make got-swag
{{< /highlight >}}

This command is currently more an experiment, use it with caution.
It may bring up wrong results.

### Code-Checks

* `misspell-check`: Checks for commonly misspelled words
* `ineffassign-check`: Checks for ineffectual assignments in the code using [ineffassign](https://github.com/gordonklaus/ineffassign).
* `gocyclo-check`: Calculates cyclomatic complexities of functions using [gocyclo](https://github.com/fzipp/gocyclo).
* `static-check`: Analyzes the code for bugs, improvements and more using [staticcheck](https://staticcheck.io/docs/).
* `gosec-check`: Inspects source code for security problems by scanning the Go AST using the [gosec tool](https://github.com/securego/gosec).
* `goconst-check`: Finds repeated strings that could be replaced by a constant using [goconst](https://github.com/jgautheron/goconst/).