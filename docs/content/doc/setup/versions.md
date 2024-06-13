---
date: "2022-07-07:00:00+02:00"
title: "Versions"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "setup"
---

# Vikunja Versions

Vikunja api is available in two different release flavors.

{{< table_of_contents >}}

## Stable

Stable releases have a fixed version number like `0.20.2` and are published at irregular intervals whenever a new version is ready.
They receive few bugfixes and security patches.

We use [Semantic Versioning](https://semver.org) for these releases.

## Unstable

Unstable versions are build every time a PR is merged or a commit to the main development branch is made.
As such, they contain the current development code and are more likely to have bugs.
There might be multiple new such builds a day.

Versions contain the last stable version, the number of commits since then and the commit the currently running binary was built from.
They look like this: `v0.20.2-75-6049427322`

Since a release is also cut from the main branch at some point, features from unstable will eventually become available in stable releases.
At the point in time of a new version release, the unstable build is the same exact thing.

The demo instance at [try.vikunja.io](https://try.vikunja.io) automatically updates and always runs the last unstable build.

## Switching between versions

First you should create a backup of your current setup!

Switching between versions is the same process as [upgrading]({{< ref install.md >}}#updating).
Simply replace the stable binary with an unstable one or vice-versa.

For installations using docker, it is as simple as using the `unstable` or `latest` tag to switch between versions.

**Note:** While switching from stable to unstable should work without any problem, switching back might work but is not recommended and might break your instance.
To switch from unstable back to stable the best way is to wait for the next stable release after the used unstable build and then upgrade to that.
