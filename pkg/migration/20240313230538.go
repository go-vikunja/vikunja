// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type projectViewBucketConfiguration20240313230538 struct {
	Title  string
	Filter string
}

type projectView20240313230538 struct {
	ID        int64   `xorm:"autoincr not null unique pk" json:"id" param:"view"`
	Title     string  `xorm:"varchar(255) not null" json:"title" valid:"runelength(1|250)"`
	ProjectID int64   `xorm:"not null index" json:"project_id" param:"project"`
	ViewKind  int     `xorm:"not null" json:"view_kind"`
	Filter    string  `xorm:"text null default null" query:"filter" json:"filter"`
	Position  float64 `xorm:"double null" json:"position"`

	BucketConfigurationMode int                                             `xorm:"default 0" json:"bucket_configuration_mode"`
	BucketConfiguration     []*projectViewBucketConfiguration20240313230538 `xorm:"json" json:"bucket_configuration"`

	Updated time.Time `xorm:"updated not null" json:"updated"`
	Created time.Time `xorm:"created not null" json:"created"`
}

func (projectView20240313230538) TableName() string {
	return "project_views"
}

type projects20240313230538 struct {
	ID int64 `xorm:"autoincr not null unique pk" json:"id" param:"view"`
}

func (projects20240313230538) TableName() string {
	return "projects"
}

type filters20240313230538 struct {
	ID int64 `xorm:"autoincr not null unique pk" json:"id" param:"view"`
}

func (filters20240313230538) TableName() string {
	return "saved_filters"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20240313230538",
		Description: "Add project views table",
		Migrate: func(tx *xorm.Engine) error {
			err := tx.Sync2(projectView20240313230538{})
			if err != nil {
				return err
			}

			projects := []*projects20240313230538{}
			err = tx.Find(&projects)
			if err != nil {
				return err
			}

			createView := func(projectID int64, kind int, title string, position float64) error {
				view := &projectView20240313230538{
					Title:     title,
					ProjectID: projectID,
					ViewKind:  kind,
					Position:  position,
				}

				if kind == 3 {
					view.BucketConfigurationMode = 1
				}

				_, err := tx.Insert(view)
				return err
			}

			for _, project := range projects {
				err = createView(project.ID, 0, "List", 100)
				if err != nil {
					return err
				}
				err = createView(project.ID, 1, "Gantt", 200)
				if err != nil {
					return err
				}
				err = createView(project.ID, 2, "Table", 300)
				if err != nil {
					return err
				}
				err = createView(project.ID, 3, "Kanban", 400)
				if err != nil {
					return err
				}
			}

			filters := []*filters20240313230538{}
			err = tx.Find(&filters)
			if err != nil {
				return err
			}

			for _, filter := range filters {
				err = createView(filter.ID*-1-1, 0, "List", 100)
				if err != nil {
					return err
				}
				err = createView(filter.ID*-1-1, 1, "Gantt", 200)
				if err != nil {
					return err
				}
				err = createView(filter.ID*-1-1, 2, "Table", 300)
				if err != nil {
					return err
				}
				err = createView(filter.ID*-1-1, 3, "Kanban", 400)
				if err != nil {
					return err
				}
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
