---
date: "2019-03-29:00:00+02:00"
title: "Database Migrations"
draft: false
type: "doc"
menu:
  sidebar:
    parent: "development"
---

# Database Migrations

Vikunja runs all database migrations automatically on each start if needed.
Additionally, they can also be run directly by using the `migrate` command.

We use [xormigrate](https://github.com/techknowlogick/xormigrate) to handle migrations, 
which is based on gormigrate.

{{< table_of_contents >}}

## Add a new migration

All migrations are stored in `pkg/migrations` and files should have the same name as their id.

Each migration should have a function to apply and roll it back, as well as a numeric id (the datetime)
and a more in-depth description of what the migration actually does.

To easily get a new id, run the following on any unix system: 

{{< highlight bash >}}
date +%Y%m%d%H%M%S
{{< /highlight >}}

New migrations should be added via the `init()` function to the `migrations` variable.
All migrations are sorted before being executed, since `init()` does not guarantee the order.

When you're adding a new struct, you also need to add it to the `models.GetTables()` function
to ensure it will be created on new installations.

### Generating a new migration stub

You can easily generate a pre-filled migration stub by running `mage dev:make-migration`.
It will ask you for a table name and generate an empty migration similar to the example shown below.

### Example

{{< highlight golang >}}
package migration

import (
	"github.com/go-xorm/xorm"
	"src.techknowlogick.com/xormigrate"
)

// Used for rollback
type teamMembersMigration20190328074430 struct {
	Updated int64 `xorm:"updated"`
}

func (teamMembersMigration20190328074430) TableName() string {
	return "team_members"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20190328074430",
		Description: "Remove updated from team_members",
		Migrate: func(tx *xorm.Engine) error {
			return dropTableColum(tx, "team_members", "updated")
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.Sync2(teamMembersMigration20190328074430{})
		},
	})
}
{{< /highlight >}}

You should always copy the changed parts of the struct you're changing when adding migraitons.