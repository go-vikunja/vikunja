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

We use go modules to manage third-party libraries for Vikunja, so you'll need at least go `1.11` to use these.

To contribute to Vikunja, fork the project and work on the main branch.

A lot of developing tasks are automated using a Magefile, so make sure to [take a look at it]({{< ref "mage.md">}}).

Make sure to check the other doc articles for specific development tasks like [testing]({{< ref "test.md">}}), 
[database migrations]({{< ref "db-migrations.md" >}}) and the [project structure]({{< ref "structure.md" >}}).
