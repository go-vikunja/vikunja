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

* [dump](#dump)
* [help](#help)
* [migrate](#migrate)
* [restore](#restore)
* [testmail](#testmail)
* [user](#user)
* [version](#version)
* [web](#web)

If you don't specify a command, the [`web`](#web) command will be executed.

All commands use the same standard [config file]({{< ref "../setup/config.md">}}).

### `dump`

Creates a zip file with all vikunja-related files.
This includes config, version, all files and the full database.

Usage:
{{< highlight bash >}}
$ vikunja dump
{{< /highlight >}}

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
 
### `restore`

Restores a previously created dump from a zip file, see `dump`.

Usage:
{{< highlight bash >}}
$ vikunja restore <path to dump zip file>
{{< /highlight >}}

### `testmail`

Sends a test mail using the configured smtp connection.

Usage:
{{< highlight bash >}}
$ vikunja testmail <email to send the test mail to>
{{< /highlight >}}

### `user`

Bundles a few commands to manage users.

#### `user change-status`

Enable or disable a user. Will toggle the current status if no flag (`--enable` or `--disable`) is provided.

Usage:
{{< highlight bash >}}
$ vikunja user change-status <user id> <flags>
{{< /highlight >}}

Flags:
* `-d`, `--disable`: Disable the user.
* `-e`, `--enable`: Enable the user.

#### `user create`

Create a new user.

Usage:
{{< highlight bash >}}
$ vikunja user create <flags>
{{< /highlight >}}

Flags:
* `-a`, `--avatar-provider`: The avatar provider of the new user. Optional.
* `-e`, `--email`: The email address of the new user.
* `-p`, `--password`: The password of the new user. You will be asked to enter it if not provided through the flag.
* `-u`, `--username`: The username of the new user.

#### `user list`

Shows a list of all users.

Usage:
{{< highlight bash >}}
$ vikunja user list
{{< /highlight >}}

#### `user reset-password`

Reset a users password, either through mailing them a reset link or directly.

Usage:
{{< highlight bash >}}
$ vikunja user reset-password <flags>
{{< /highlight >}}

Flags:
* `-d`, `--direct`: If provided, reset the password directly instead of sending the user a reset mail.
* `-p`, `--password`: The new password of the user. Only used in combination with --direct. You will be asked to enter it if not provided through the flag.

#### `user update`

Update an existing user.

Usage:
{{< highlight bash >}}
$ vikunja user update <user id>
{{< /highlight >}}

Flags:
* `-a`, `--avatar-provider`: The new avatar provider of the new user.
* `-e`, `--email`: The new email address of the user.
* `-u`, `--username`: The new username of the user.

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
