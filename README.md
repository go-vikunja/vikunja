# Vikunja API

> The Todo-app to organize your life.

[![Build Status](https://drone.kolaente.de/api/badges/vikunja/api/status.svg)](https://drone.kolaente.de/vikunja/api)
[![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](LICENSE)
[![Download](https://img.shields.io/badge/download-v0.5-brightgreen.svg)](https://storage.kolaente.de/minio/vikunja/)
[![Docker Pulls](https://img.shields.io/docker/pulls/vikunja/api.svg)](https://hub.docker.com/r/vikunja/api/)
[![Swagger Docs](https://img.shields.io/badge/swagger-docs-brightgreen.svg)](https://try.vikunja.io/api/v1/swagger)
[![Go Report Card](https://goreportcard.com/badge/git.kolaente.de/vikunja/api)](https://goreportcard.com/report/git.kolaente.de/vikunja/api)
[![cover.run](https://cover.run/go/code.vikunja.io/api.svg?style=flat&tag=golang-1.10)](https://cover.run/go?tag=golang-1.10&repo=code.vikunja.io%2Fapi) 

## Features

* Create TODO lists with tasks
  * Reminder for tasks
* Namespaces: A "group" which bundels multiple lists
* Share lists and namespaces with teams and users with granular permissions

Try it under [try.vikunja.io](https://try.vikunja.io)!

### Roadmap

> I know, it's still a long way to go. I'm currently working on a lot of "basic" features, the exiting things will come later. Don't worry, they'll come.

* [x] Prioritize tasks
* [x] Subtasks
* [x] Repeating tasks
* [x] Get tasks via caldav
* [x] Get all your tasks for an interval (day/month/period)
* [ ] Labels for todo lists and tasks
* [ ] Assign users to tasks
* [ ] Attachments on tasks
* [ ] More sharing features
  * [x] Share with individual users
  * [ ] Share via a world-readable link with or without password, like Nextcloud
* [ ] Read-only websocket to notify multiple clients of updates when something was changed
* [ ] "Smart Lists" - Create lists based on filters
* [ ] IMAP-Integration - Send an email to Vikunja to create a new task
* [ ] Webhooks - Trigger other events when an action is done (like completing a task)
* [ ] Performace statistics - Get an overview and beautiful charts about what you got done this month
* [ ] Activity feeds - Get a quick overview about who did what
* [ ] Bulk-edit multiple tasks at once
* [ ] Team-efforts - Requiring a task to be marked as done by multiple members until it's done
* [ ] Global limits for namespaces/lists/tasks
* [ ] Disable registration, making an instance "invite-only" 

See (Featurecreep.md)[Featurecreep.md] for even more! (mostly ideas, for now)

* [ ] [Mobile apps](https://code.vikunja.io/app) (seperate repo) *In Progress*
* [ ] [Webapp](https://code.vikunja.io/frontend) (seperate repo) *In Progress*

## Development

We use go modules to vendor libraries for Vikunja, so you'll need at least go `1.11`.

To contribute to Vikunja, fork the project and work on the master branch.

Some internal packages are referenced using their respective package URL. This can become problematic. To “trick” the Go tool into thinking this is a clone from the official repository, download the source code into `$GOPATH/code.vikunja.io/api`. Fork the Vikunja repository, it should then be possible to switch the source directory on the command line.

```bash
cd $GOPATH/src/code.vikunja.io/api
```

To be able to create pull requests, the forked repository should be added as a remote to the Vikunja sources, otherwise changes can’t be pushed.

```bash
git remote rename origin upstream
git remote add origin git@git.kolaente.de:<USERNAME>/api.git
git fetch --all --prune
```

This should provide a working development environment for Vikunja. Take a look at the Makefile to get an overview about the available tasks. The most common tasks should be `make test` which will start our test environment and `make build` which will build a vikunja binary into the working directory. Writing test cases is not mandatory to contribute, but it is highly encouraged and helps developers sleep at night.

That’s it! You are ready to hack on Vikunja. Test changes, push them to the repository, and open a pull request.
