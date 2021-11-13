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

There are two parts you need to back up: The database and attachment files.

{{< table_of_contents >}}

## Files

To back up attachments and other files, it is enough to copy them [from the attachments folder]({{< ref "config.md" >}}#basepath) to some other place.

## Database

### MySQL

To create a backup from mysql use the `mysqldump` command:

{{< highlight bash >}}
mysqldump -u <user> -p -h <db-host> <database> > vkunja-backup.sql
{{< /highlight >}}

You will be prompted for the password of the mysql user.

To restore it, simply pipe it back into the `mysql` command:

{{< highlight bash >}}
mysql -u <user> -p -h <db-host> <database> < vkunja-backup.sql
{{< /highlight >}}

### PostgreSQL

To create a backup from PostgreSQL use the `pg_dump` command:

{{< highlight bash >}}
pg_dump -U <user> -h <db-host> <database> > vikunja-backup.sql
{{< /highlight >}}

You might be prompted for the password of the database user.

To restore it, simply pipe it back into the `psql` command:

{{< highlight bash >}}
psql -U <user> -h <db-host> <database> < vikunja-backup.sql
{{< /highlight >}}

For more information, please visit the [relevant PostgreSQL documentation](https://www.postgresql.org/docs/12/backup-dump.html).

### SQLite

To back up sqllite databases, it is enough to copy the [database file]({{< ref "config.md" >}}#path) to somwhere else.
