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

{{< table_of_contents >}}

## General

To contribute to Vikunja, fork the project and work on the main branch.
Once you feel like your changes are ready, open a PR in the respective repo.
A maintainer will take a look and give you feedback. Once everyone is happy, the PR gets merged and released.

If you plan to do a bigger change, it is better to open an issue for discussion first.

## API

The code for the api is located at [code.vikunja.io/api](https://code.vikunja.io/api).

We use go modules to manage third-party libraries for Vikunja, so you'll need at least go `1.17` to use these.

A lot of developing tasks are automated using a Magefile, so make sure to [take a look at it]({{< ref "mage.md">}}).

Make sure to check the other doc articles for specific development tasks like [testing]({{< ref "test.md">}}),
[database migrations]({{< ref "db-migrations.md" >}}) and the [project structure]({{< ref "structure.md" >}}).

## Frontend requirements

The code for the frontend is located at [code.vikunja.io/frontend](https://code.vikunja.io/frontend).

You need to have yarn v1 and nodejs in version 16 installed.

## Git flow

The `main` branch is the latest and bleeding edge branch with all changes. Unstable releases are automatically 
created from this branch.

A release gets tagged from the main branch with the version name as tag name.

Backports and point-releases should go to a `release/version` branch, based on the tag they are building on top of.

## Conventional commits

We're using [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) because they greatly simplify 
generating release notes.

It is not required to use them when creating a PR, but appreciated.
