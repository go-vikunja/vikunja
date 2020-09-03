---
date: "2020-07-06:00:00+02:00"
title: "UTF-8 Settings"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "setup"
---

# UTF-8 Settings

Vikunja itself is always fully capable of handling utf-8 characters.
However, your database might be not. 
Vikunja itself will work just fine until you want to use non-latin characters in your tasks/lists/etc.

On this page, you will find information about how to fully ensure non-latin characters like aüäß or emojis work 
with your installation.

{{< table_of_contents >}}

## Postgresql & SQLite

Postgresql and SQLite should handle utf-8 just fine - If you discover any issues nonetheless, please 
[drop us a message](https://vikunja.io/contact/).

## MySQL

MySQL is not able to handle utf-8 by default.
To fix this, follow the steps below.

To find out if your db supports utf-8, run the following in a shell or similar, assuming the database 
you're using for vikunja is called `vikunja`:

{{< highlight sql >}}
SELECT default_character_set_name FROM information_schema.SCHEMATA WHERE schema_name = 'vikunja';
{{< /highlight >}}

This will get you a result like the following:

```
+----------------------------+
| default_character_set_name |
+----------------------------+
| latin1                     |
+----------------------------+
1 row in set (0.001 sec)
```

The charset `latin1` means the db is encoded in the `latin1` encoding which does not support utf-8 characters.

(The following guide is based on [this thread from stackoverflow](https://dba.stackexchange.com/a/104866))

### 0. Backup your database

Before attempting any conversion, please [back up your database]({{< ref "backups.md">}}).

### 1. Create a pre-conversion script

Copy the following sql statements in a file called `preAlterTables.sql` and replace all occurences of `vikunja` with 
the name of your database:

{{< highlight sql >}}
use information_schema;
SELECT concat("ALTER DATABASE `",table_schema,"` CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci;") as _sql 
FROM `TABLES` where table_schema like 'vikunja' and TABLE_TYPE='BASE TABLE' group by table_schema;
SELECT concat("ALTER TABLE `",table_schema,"`.`",table_name,"` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;") as _sql  
FROM `TABLES` where table_schema like 'vikunja' and TABLE_TYPE='BASE TABLE' group by table_schema, table_name;
SELECT concat("ALTER TABLE `",table_schema,"`.`",table_name, "` CHANGE `",column_name,"` `",column_name,"` ",data_type,"(",character_maximum_length,") CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",IF(is_nullable="YES"," NULL"," NOT NULL"),";") as _sql 
FROM `COLUMNS` where table_schema like 'vikunja' and data_type in ('varchar','char');
SELECT concat("ALTER TABLE `",table_schema,"`.`",table_name, "` CHANGE `",column_name,"` `",column_name,"` ",data_type," CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",IF(is_nullable="YES"," NULL"," NOT NULL"),";") as _sql 
FROM `COLUMNS` where table_schema like 'vikunja' and data_type in ('text','tinytext','mediumtext','longtext');
{{< /highlight >}}

### 2. Run the pre-conversion script

Running this will create the actual migration script for your particular database structure and save it in a file called `alterTables.sql`:

{{< highlight bash >}}
mysql -uroot < preAlterTables.sql | egrep '^ALTER' > alterTables.sql
{{< /highlight >}}

### 3. Convert the database

At this point converting is just a matter of executing the previously generated sql script:

{{< highlight bash >}}
mysql -uroot < alterTables.sql
{{< /highlight >}}

### 4. Verify it was successfully converted

If everything worked as intended, your db collation should now look like this:

{{< highlight sql >}}
SELECT default_character_set_name FROM information_schema.SCHEMATA WHERE schema_name = 'vikunja';
{{< /highlight >}}

Should get you:

```
+----------------------------+
| default_character_set_name |
+----------------------------+
| utf8mb4                    |
+----------------------------+
1 row in set (0.001 sec)
```
