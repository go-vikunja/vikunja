---
date: "2019-02-12:00:00+02:00"
title: "What to backup"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "setup"
---

# What to backup

Vikunja does not store any data outside of the database.
So, all you need to backup are the contents of that database and maybe the config file.

## MySQL

To create a backup from mysql use the `mysqldump` command:

{{< highlight bash >}}
mysqldump -u <user> -p -h <db-host> <database> > vkunja-backup.sql
{{< /highlight >}}

You will be prompted for the password of the mysql user.

To restore it, simply pipe it back into the `mysql` command:

{{< highlight bash >}}
mysql -u <user> -p -h <db-host> <database> < vkunja-backup.sql
{{< /highlight >}}

## SQLite

To backup sqllite databases, it is enough to copy the database elsewhere.
