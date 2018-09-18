# Vikunja API

> The Todo-app to organize your life.

[![Build Status](https://drone.kolaente.de/api/badges/vikunja/api/status.svg)](https://drone.kolaente.de/vikunja/api)
[![License: LGPL v3](https://img.shields.io/badge/License-LGPL%20v3-blue.svg)](LICENSE)
[![Download](https://img.shields.io/badge/download-v0.1-brightgreen.svg)](https://storage.kolaente.de/minio/vikunja/)

## Features

* Create TODO lists with tasks
  * Reminder for tasks
* Namespaces: A "group" which bundels multiple lists
* Share lists and namespaces with teams and users with granular permissions

Try it under [try.vikunja.io](https://try.vikunja.io)!

### Roadmap

* [ ] Labels for todo lists and tasks
* [ ] Prioritize tasks
* [ ] More sharing features
  * [ ] Share with individual users
  * [ ] Share via a world-readable link with or without password, like Nextcloud

* [ ] Mobile apps (seperate repo)
* [ ] Webapp (seperate repo)
* [ ] "Native" clients (will probably be something with electron)

## Development

To contribute to Vikunja, fork the project and work on the master branch.

Some internal packages are referenced using their respective package URL. This can become problematic. To “trick” the Go tool into thinking this is a clone from the official repository, download the source code into `$GOPATH/code.vikunja.io/api`. Fork the Vikunja repository, it should then be possible to switch the source directory on the command line.

```bash
cd $GOPATH/src/code.vikunja.io/api
```

To be able to create pull requests, the forked repository should be added as a remote to the Vikunja sources, otherwise changes can’t be pushed.

```bash
git remote rename origin upstream
git remote add origin git@git.kolaente.de:<USERNAME>/vikunja.git
git fetch --all --prune
```

This should provide a working development environment for Vikunja. Take a look at the Makefile to get an overview about the available tasks. The most common tasks should be `make test` which will start our test environment and `make build` which will build a vikunja binary into the working directory. Writing test cases is not mandatory to contribute, but it is highly encouraged and helps developers sleep at night.

That’s it! You are ready to hack on Vikunja. Test changes, push them to the repository, and open a pull request.
