---
title: "Typesense"
date: 2023-09-29T12:23:55+02:00
draft: false
menu:
  sidebar:
    parent: "setup"
---

# Use Typesense for enhanced search capabilities

Vikunja supports using [Typesense](https://typesense.org/) for a better search experience.
Typesense allows fast fulltext search including fuzzy matching support. 
It may return different results than what you'd get with a database-only search, but generally, the results are more relevant to what you're looking for.

This document explains how to set up and use Typesense with Vikunja.

## Setup

1. First, install Typesense on your system. Refer to [their documentation](https://typesense.org/docs/guide/install-typesense.html) for specific instructions.
2. Once Typesense is available on your system and reachable by Vikunja, add the relevant configuration keys to your Vikunja config. [Check out the docs article about this]({{< ref "config.md#typesense">}}).
3. Index all tasks currently in Vikunja. To do that, run the `vikunja index` command with the api binary. This may take a while, depending on the size of your instance.
4. Restart the api. From now on, all task changes will be automatically indexed in Typesense.
