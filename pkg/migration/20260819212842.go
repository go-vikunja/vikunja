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
	"slices"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type projects20260819212842 struct {
	ID              int64 `xorm:"bigint autoincr not null unique pk"`
	ParentProjectID int64 `xorm:"bigint INDEX null"`
	IsArchived      bool  `xorm:"not null default false"`
}

func (projects20260819212842) TableName() string {
	return "projects"
}

const archivedBackfillBatch20260819212842 = 500

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260819212842",
		Description: "Backfill is_archived for descendants of archived projects",
		Migrate:     backfillArchivedDescendants20260819212842,
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}

// Before archiving cascaded to descendants (#775), children of archived
// projects kept is_archived=false and were only treated as archived by
// walking the parent chain at read time. Materialize that state so the
// column alone is authoritative.
func backfillArchivedDescendants20260819212842(tx *xorm.Engine) error {
	all := []*projects20260819212842{}
	if err := tx.Cols("id", "parent_project_id", "is_archived").Find(&all); err != nil {
		return err
	}

	children := make(map[int64][]int64, len(all))
	archivedRoots := []int64{}
	for _, p := range all {
		if p.ParentProjectID > 0 {
			children[p.ParentProjectID] = append(children[p.ParentProjectID], p.ID)
		}
		if p.IsArchived {
			archivedRoots = append(archivedRoots, p.ID)
		}
	}

	visited := make(map[int64]bool, len(archivedRoots))
	toArchive := []int64{}
	stack := slices.Clone(archivedRoots)
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if visited[id] {
			continue
		}
		visited[id] = true
		for _, c := range children[id] {
			if !visited[c] {
				toArchive = append(toArchive, c)
				stack = append(stack, c)
			}
		}
	}

	for chunk := range slices.Chunk(toArchive, archivedBackfillBatch20260819212842) {
		_, err := tx.In("id", chunk).
			And("is_archived = ?", false).
			Cols("is_archived").
			Update(&projects20260819212842{IsArchived: true})
		if err != nil {
			return err
		}
	}
	return nil
}
