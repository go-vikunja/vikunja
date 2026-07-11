// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package migration

import (
	"fmt"

	"code.vikunja.io/api/pkg/config"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type projectTaskScope20260711120000 struct {
	TaskScope          string  `xorm:"varchar(20) not null default 'current' 'task_scope'"`
	IncludedProjectIDs []int64 `xorm:"json null 'included_project_ids'"`
}

func (projectTaskScope20260711120000) TableName() string {
	return "projects"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260711120000",
		Description: "Add child project task scope to projects",
		Migrate: func(tx *xorm.Engine) error {
			if err := tx.Sync(projectTaskScope20260711120000{}); err != nil {
				return err
			}

			// Sync ignores the explicit column name when the Go field ends in IDs.
			switch config.DatabaseType.GetString() {
			case "sqlite", "postgres":
				_, err := tx.Exec(`ALTER TABLE projects RENAME COLUMN included_project_i_ds TO included_project_ids`)
				return err
			case "mysql":
				_, err := tx.Exec("ALTER TABLE projects CHANGE included_project_i_ds included_project_ids JSON NULL")
				return err
			default:
				return fmt.Errorf("unsupported database type %q", config.DatabaseType.GetString())
			}
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
