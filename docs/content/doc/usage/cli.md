---
date: "2019-03-31:00:00+01:00"
title: "CLI"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "usage"
---

# Command line interface

You can interact with Vikunja using its `cli` interface. 
The following commands are available:

* [help](#help)
* [migrate](#migrate)
* [version](#version)
* [web](#web)

If you don't specify a command, the [`web`](#web) command will be executed.

All commands use the same standard [config file]({{< ref "../setup/config.md">}}).

### `help`

Shows more detailed help about any command.

Usage:

{{< highlight bash >}}
$ vikunja help [command]
{{< /highlight >}}

### `migrate`

Run all database migrations which didn't already run.

Usage:
{{< highlight bash >}}
$ vikunja migrate [flags]
$ vikunja migrate [command]
{{< /highlight >}}

#### `migrate list`

Shows a list with all database migrations.

Usage:
{{< highlight bash >}}
$ vikunja migrate list
{{< /highlight >}}

#### `migrate rollback`

Roll migrations back until a certain point.

Usage:
{{< highlight bash >}}
$ vikunja migrate rollback [flags]    
{{< /highlight >}}

Flags:
* `-n`, `--name` string: The id of the migration you want to roll back until.
 

### `version`

Prints the version of Vikunja.
This is either the semantic version (something like `0.7`) or version + git commit hash.

Usage:
{{< highlight bash >}}
$ vikunja version    
{{< /highlight >}}

### `web`

Starts Vikunja's REST api server.

Usage:
{{< highlight bash >}}
$ vikunja web    
{{< /highlight >}}