# Web frontend for Vikunja

> The todo app to organize your life.

[![Build Status](https://drone.kolaente.de/api/badges/vikunja/frontend/status.svg)](https://drone.kolaente.de/vikunja/frontend)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](LICENSE)
[![Download](https://img.shields.io/badge/download-v0.20.3-brightgreen.svg)](https://dl.vikunja.io)
[![Translation](https://badges.crowdin.net/vikunja/localized.svg)](https://crowdin.com/project/vikunja)

This is the web frontend for Vikunja, written in Vue.js.

Take a look at [our roadmap](https://my.vikunja.cloud/share/UrdhKPqumxDXUbYpEGJLSIyNTwAnbBzVlwdDpRbv/auth) (hosted on Vikunja!) for a list of things we're currently working on!

## Security Reports

If you find any security-related issues you don't want to disclose publicly, please use [the contact information on our website](https://vikunja.io/contact/#security).

## Docker

There is a [docker image available](https://hub.docker.com/r/vikunja/api) with support for http/2 and aggressive caching enabled.
In order to build it from sources run the command below. (Docker >= v19.03)

```shell
export DOCKER_BUILDKIT=1
docker build -t vikunja/frontend .
```

Refer to Refer [to multi-platform documentation](https://docs.docker.com/build/building/multi-platform/) in order to build for the different platform.

## Project setup

```shell
pnpm install
```

### Compiles and hot-reloads for development

```shell
pnpm run serve
```

### Compiles and minifies for production

```shell
pnpm run build
```

### Lints and fixes files

```shell
pnpm run lint
```

