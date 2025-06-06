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
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type favorites20210709211508 struct {
	EntityID int64 `xorm:"bigint not null pk"`
	UserID   int64 `xorm:"bigint not null pk"`
	Kind     int   `xorm:"int not null pk"`
}

func (favorites20210709211508) TableName() string {
	return "favorites"
}

type task20210709211508 struct {
	ID          int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"listtask"`
	IsFavorite  bool  `xorm:"default false" json:"is_favorite"`
	CreatedByID int64 `xorm:"bigint not null" json:"-"` // ID of the user who put that task on the list
}

func (task20210709211508) TableName() string {
	return "tasks"
}

type list20210709211508 struct {
	ID         int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"listtask"`
	IsFavorite bool  `xorm:"default false" json:"is_favorite"`
	OwnerID    int64 `xorm:"bigint not null" json:"-"` // ID of the user who put that task on the list
}

func (list20210709211508) TableName() string {
	return "lists"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20210709211508",
		Description: "Move favorites to new table",
		Migrate: func(tx *xorm.Engine) error {
			err := tx.Sync2(favorites20210709211508{})
			if err != nil {
				return err
			}

			// Migrate all existing favorites
			tasks := []*task20210709211508{}
			err = tx.Where("is_favorite = ?", true).Find(&tasks)
			if err != nil {
				return err
			}

			for _, task := range tasks {
				fav := &favorites20210709211508{
					EntityID: task.ID,
					UserID:   task.CreatedByID,
					Kind:     1,
				}
				_, err = tx.Insert(fav)
				if err != nil {
					return err
				}
			}

			lists := []*list20210709211508{}
			err = tx.Where("is_favorite = ?", true).Find(&lists)
			if err != nil {
				return err
			}

			for _, list := range lists {
				fav := &favorites20210709211508{
					EntityID: list.ID,
					UserID:   list.OwnerID,
					Kind:     2,
				}
				_, err = tx.Insert(fav)
				if err != nil {
					return err
				}
			}

			err = dropTableColum(tx, "tasks", "is_favorite")
			if err != nil {
				return err
			}

			return dropTableColum(tx, "lists", "is_favorite")
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
