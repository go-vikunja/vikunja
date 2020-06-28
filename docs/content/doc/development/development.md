---
date: "2019-02-12:00:00+02:00"
title: "Development"
toc: true
draft: false
type: "doc"
menu:
  sidebar:
    parent: "development"
    name: "Development"
---

# Development

We use go modules to vendor libraries for Vikunja, so you'll need at least go `1.11` to use these.
If you don't intend to add new dependencies, go `1.9` and above should be fine.

To contribute to Vikunja, fork the project and work on the master branch.

A lot of developing tasks are automated using a Makefile, so make sure to [take a look at it]({{< ref "make.md">}}).

## Libraries

We keep all libraries used for Vikunja around in the `vendor/` folder to still be able to build the project even if
some maintainers take their libraries down like [it happened in the past](https://github.com/jteeuwen/go-bindata/issues/5).

## Tests

See [testing]({{< ref "test.md">}}).

#### Development using go modules

If you're able to use go modules, you can clone the project wherever you want to and work from there.

#### Development-setup without go modules

Some internal packages are referenced using their respective package URL. This can become problematic. 
To “trick” the Go tool into thinking this is a clone from the official repository, download the source code 
into `$GOPATH/code.vikunja.io/api`. Fork the Vikunja repository, it should then be possible to switch the source directory on the command line.

{{< highlight bash >}}
cd $GOPATH/src/code.vikunja.io/api
{{< /highlight >}}

To be able to create pull requests, the forked repository should be added as a remote to the Vikunja sources, otherwise changes can’t be pushed.

{{< highlight bash >}}
git remote rename origin upstream
git remote add origin git@git.kolaente.de:<USERNAME>/api.git
git fetch --all --prune
{{< /highlight >}}

This should provide a working development environment for Vikunja. Take a look at the Makefile to get an overview about 
the available tasks. The most common tasks should be `make test` which will start our test environment and `make build` 
which will build a vikunja binary into the working directory. Writing test cases is not mandatory to contribute, but it 
is highly encouraged and helps developers sleep at night.

That’s it! You are ready to hack on Vikunja. Test changes, push them to the repository, and open a pull request.

## Static assets

Each Vikunja release contains all static assets directly compiled into the binary.
To prevent this during development, use the `dev` tag when developing.

See the [make docs](make.md#statically-compile-all-templates-into-the-binary) about how to compile with static assets for a release.
