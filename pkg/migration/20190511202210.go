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
	"strconv"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/utils"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type listTask20190511202210 struct {
	ID                int64   `xorm:"int(11) autoincr not null unique pk" json:"id" param:"listtask"`
	Text              string  `xorm:"varchar(250) not null" json:"text" valid:"runelength(3|250)" minLength:"3" maxLength:"250"`
	Description       string  `xorm:"varchar(250)" json:"description" valid:"runelength(0|250)" maxLength:"250"`
	Done              bool    `xorm:"INDEX null" json:"done"`
	DoneAtUnix        int64   `xorm:"INDEX null" json:"done_at"`
	DueDateUnix       int64   `xorm:"int(11) INDEX null" json:"due_date"`
	RemindersUnix     []int64 `xorm:"JSON TEXT null" json:"reminder_dates"`
	CreatedByID       int64   `xorm:"int(11) not null" json:"-"` // ID of the user who put that task on the list
	ListID            int64   `xorm:"int(11) INDEX not null" json:"list_id" param:"list"`
	RepeatAfter       int64   `xorm:"int(11) INDEX null" json:"repeat_after"`
	ParentTaskID      int64   `xorm:"int(11) INDEX null" json:"parentTaskID"`
	Priority          int64   `xorm:"int(11) null" json:"priority"`
	StartDateUnix     int64   `xorm:"int(11) INDEX null" json:"start_date" query:"-"`
	EndDateUnix       int64   `xorm:"int(11) INDEX null" json:"end_date" query:"-"`
	HexColor          string  `xorm:"varchar(6) null" json:"hex_color" valid:"runelength(0|6)" maxLength:"6"`
	UID               string  `xorm:"varchar(250) null" json:"-"`
	Sorting           string  `xorm:"-" json:"-" query:"sort"` // Parameter to sort by
	StartDateSortUnix int64   `xorm:"-" json:"-" query:"start_date"`
	EndDateSortUnix   int64   `xorm:"-" json:"-" query:"end_date"`
	Created           int64   `xorm:"created not null" json:"created"`
	Updated           int64   `xorm:"updated not null" json:"updated"`
}

func (listTask20190511202210) TableName() string {
	return "tasks"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20190511202210",
		Description: "Add task uid",
		Migrate: func(tx *xorm.Engine) error {
			err := tx.Sync2(listTask20190511202210{})
			if err != nil {
				return err
			}

			// Get all tasks and generate a random uid for them
			var allTasks []*listTask20190511202210
			err = tx.Find(&allTasks)
			if err != nil {
				return err
			}

			for _, t := range allTasks {
				var err error
				t.UID, err = utils.CryptoRandomString(40)
				if err != nil {
					log.Errorf("Migration 20190511202210: Could not generate random string: %s", err)
					t.UID = "migration-error-uid-" + strconv.FormatInt(t.ID, 10)
				}
				_, err = tx.Where("id = ?", t.ID).Cols("uid").Update(t)
				if err != nil {
					return err
				}
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return dropTableColum(tx, "tasks", "uid")
		},
	})
}
