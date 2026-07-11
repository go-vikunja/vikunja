// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type projectViewProjectScope20260711120000 struct {
	ProjectScope string `xorm:"varchar(20) not null default 'current' 'project_scope'"`
	// XORM splits an IDs suffix as i_ds during Sync unless the migration field uses Ids.
	IncludedProjectIds []int64 `xorm:"json null"`
}

func (projectViewProjectScope20260711120000) TableName() string {
	return "project_views"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260711120000",
		Description: "Add child project scope to project views",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync(projectViewProjectScope20260711120000{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
