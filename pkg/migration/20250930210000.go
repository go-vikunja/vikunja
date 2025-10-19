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

type taskLegacyCreator struct {
	ID          int64 `xorm:"bigint autoincr not null unique pk"`
	ProjectID   int64 `xorm:"bigint not null"`
	CreatedByID int64 `xorm:"bigint not null"`
}

func (taskLegacyCreator) TableName() string {
	return "tasks"
}

type projectOwnerLookup struct {
	ID      int64 `xorm:"bigint autoincr not null unique pk"`
	OwnerID int64 `xorm:"bigint not null"`
}

func (projectOwnerLookup) TableName() string {
	return "projects"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20250930210000",
		Description: "backfill legacy tasks with missing created_by_id",
		Migrate: func(tx *xorm.Engine) error {
			return backfillLegacyTaskCreators(tx)
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}

func backfillLegacyTaskCreators(tx *xorm.Engine) error {
	session := tx.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	legacyTasks := make([]*taskLegacyCreator, 0)
	if err := session.Where("created_by_id = ?", 0).Find(&legacyTasks); err != nil {
		_ = session.Rollback()
		return err
	}

	if len(legacyTasks) == 0 {
		return session.Commit()
	}

	projectIDSet := make(map[int64]struct{}, len(legacyTasks))
	for _, task := range legacyTasks {
		projectIDSet[task.ProjectID] = struct{}{}
	}

	projectIDs := make([]int64, 0, len(projectIDSet))
	for id := range projectIDSet {
		projectIDs = append(projectIDs, id)
	}

	projectOwners := make(map[int64]int64, len(projectIDs))
	if len(projectIDs) > 0 {
		ownerRows := make([]*projectOwnerLookup, 0, len(projectIDs))
		if err := session.In("id", projectIDs).Find(&ownerRows); err != nil {
			_ = session.Rollback()
			return err
		}
		for _, row := range ownerRows {
			projectOwners[row.ID] = row.OwnerID
		}
	}

	defaultUserID, err := findFallbackUserID(session)
	if err != nil {
		_ = session.Rollback()
		return err
	}

	if defaultUserID == 0 {
		// No users in the system, nothing we can backfill safely.
		return session.Commit()
	}

	for _, task := range legacyTasks {
		ownerID := projectOwners[task.ProjectID]
		newCreatorID := ownerID
		if newCreatorID == 0 {
			newCreatorID = defaultUserID
		}
		if newCreatorID == 0 {
			continue
		}

		update := &taskLegacyCreator{CreatedByID: newCreatorID}
		if _, err := session.Table("tasks").Where("id = ?", task.ID).Cols("created_by_id").Update(update); err != nil {
			_ = session.Rollback()
			return err
		}
	}

	return session.Commit()
}

func findFallbackUserID(session *xorm.Session) (int64, error) {
	type userRow struct {
		ID int64 `xorm:"bigint autoincr not null unique pk"`
	}

	var row userRow
	has, err := session.Table("users").OrderBy("id ASC").Limit(1).Get(&row)
	if err != nil {
		return 0, err
	}
	if !has {
		return 0, nil
	}
	return row.ID, nil
}
