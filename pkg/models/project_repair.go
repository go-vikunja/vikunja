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

package models

import (
	"code.vikunja.io/api/pkg/log"

	"xorm.io/xorm"
)

// RepairOrphanedProjectsResult holds the result of a repair operation.
type RepairOrphanedProjectsResult struct {
	Found    int
	Repaired int
}

// RepairOrphanedProjects finds projects whose parent_project_id references a
// project that no longer exists and sets their parent_project_id to 0,
// making them top-level projects.
// If dryRun is true, it reports what would be fixed without making changes.
func RepairOrphanedProjects(s *xorm.Session, dryRun bool) (*RepairOrphanedProjectsResult, error) {
	result := &RepairOrphanedProjectsResult{}

	var orphans []*Project
	err := s.SQL(`SELECT p.* FROM projects p
		LEFT JOIN projects parent ON p.parent_project_id = parent.id
		WHERE p.parent_project_id > 0 AND parent.id IS NULL`).
		Find(&orphans)
	if err != nil {
		return nil, err
	}

	result.Found = len(orphans)

	if dryRun {
		for _, p := range orphans {
			log.Infof("[dry-run] Would re-parent project %d (%s) from non-existent parent %d to top level",
				p.ID, p.Title, p.ParentProjectID)
		}
		return result, nil
	}

	for _, p := range orphans {
		log.Infof("Re-parenting project %d (%s) from non-existent parent %d to top level",
			p.ID, p.Title, p.ParentProjectID)
		_, err = s.Where("id = ?", p.ID).
			Cols("parent_project_id").
			Update(&Project{ParentProjectID: 0})
		if err != nil {
			return result, err
		}
		result.Repaired++
	}

	return result, nil
}

