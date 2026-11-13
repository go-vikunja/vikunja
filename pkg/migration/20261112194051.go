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

type jobs20261112194051 struct {
	ID int64 `xorm:"pk"`
}

func (jobs20261112194051) TableName() string { return "jobs" }

func removeCompletedJobs20261112194051(tx *xorm.Engine) error {
	var afterID int64
	for {
		var rows []jobs20261112194051
		if err := tx.Where("state = ? AND id > ?", "completed", afterID).OrderBy("id").Limit(500).Find(&rows); err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		ids := make([]int64, len(rows))
		for i, row := range rows {
			ids[i] = row.ID
		}
		if _, err := tx.In("id", ids).Where("state = ?", "completed").Delete(new(jobs20261112194051)); err != nil {
			return err
		}
		afterID = rows[len(rows)-1].ID
	}
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261112194051",
		Description: "Remove completed job history",
		Migrate:     removeCompletedJobs20261112194051,
		Rollback:    func(*xorm.Engine) error { return nil },
	})
}
