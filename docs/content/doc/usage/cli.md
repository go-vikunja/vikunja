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

You can interact with Vikunja using its `cli` interface.<br />
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

## Using the cli in docker

When running Vikunja in docker, you'll need to execute all commands in the `api` container.
Instead of running the `vikunja` binary directly, run it like this:

```sh
docker exec <name of the vikunja api container> /app/vikunja/vikunja <subcommand>
```

### `dump`

Creates a zip file with all vikunja-related files.
This includes config, version, all files and the full database.

Usage:
```
$ vikunja dump
```

### `help`

Shows more detailed help about any command.

Usage:

```
$ vikunja help [command]
```

### `migrate`

Run all database migrations which didn't already run.

Usage:
```
$ vikunja migrate [flags]
$ vikunja migrate [command]
```

#### `migrate list`

Shows a list with all database migrations.

Usage:
```
$ vikunja migrate list
```

#### `migrate rollback`

Roll migrations back until a certain point.

Usage:
```
$ vikunja migrate rollback [flags]
```

Flags:
* `-n`, `--name` string: The id of the migration you want to roll back until.

### `restore`

Restores a previously created dump from a zip file, see `dump`.

Usage:
```
$ vikunja restore <path to dump zip file>
```

### `testmail`

Sends a test mail using the configured smtp connection.

Usage:
```
$ vikunja testmail <email to send the test mail to>
```

### `user`

Bundles a few commands to manage users.

#### `user change-status`

Enable or disable a user. Will toggle the current status if no flag (`--enable` or `--disable`) is provided.

Usage:
```
$ vikunja user change-status <user id> <flags>
```

Flags:
* `-d`, `--disable`: Disable the user.
* `-e`, `--enable`: Enable the user.

#### `user create`

Create a new user.

Usage:
```
$ vikunja user create <flags>
```

Flags:
* `-a`, `--avatar-provider`: The avatar provider of the new user. Optional.
* `-e`, `--email`: The email address of the new user.
* `-p`, `--password`: The password of the new user. You will be asked to enter it if not provided through the flag.
* `-u`, `--username`: The username of the new user.

#### `user delete`

Start the user deletion process.
If called without the `--now` flag, this command will only trigger an email to the user in order for them to confirm and start the deletion process (this is the same behavoir as if the user requested their deletion via the web interface).
With the flag the user is deleted **immediately**.

**USE WITH CAUTION.**

```
$ vikunja user delete <id> <flags>
```

Flags:
* `-n`, `--now` If provided, deletes the user immediately instead of emailing them first.

#### `user list`

Shows a list of all users.

Usage:
```
$ vikunja user list
```

#### `user reset-password`

Reset a users password, either through mailing them a reset link or directly.

Usage:
```
$ vikunja user reset-password <flags>
```

Flags:
* `-d`, `--direct`: If provided, reset the password directly instead of sending the user a reset mail.
* `-p`, `--password`: The new password of the user. Only used in combination with --direct. You will be asked to enter it if not provided through the flag.

#### `user update`

Update an existing user.

Usage:
```
$ vikunja user update <user id>
```

Flags:
* `-a`, `--avatar-provider`: The new avatar provider of the new user.
* `-e`, `--email`: The new email address of the user.
* `-u`, `--username`: The new username of the user.

### `version`

Prints the version of Vikunja.
This is either the semantic version (something like `0.7`) or version + git commit hash.

Usage:
```
$ vikunja version
```

### `web`

Starts Vikunja's REST api server.

Usage:
```
$ vikunja web
```
