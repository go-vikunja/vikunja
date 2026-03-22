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
	"xorm.io/xorm"
)

// GetProjectIDsForToken returns the list of project IDs a token is scoped to.
// Returns nil if the token has no project scope (user-wide token).
func GetProjectIDsForToken(s *xorm.Session, token *APIToken) ([]int64, error) {
	if token.ProjectID == 0 {
		return nil, nil
	}

	if !token.IncludeSubProjects {
		return []int64{token.ProjectID}, nil
	}

	// Use recursive CTE to get all descendant project IDs
	var descendantIDs []int64
	err := s.SQL(
		`WITH RECURSIVE descendant_ids (id) AS (
    SELECT id
    FROM projects
    WHERE id = ?
    UNION ALL
    SELECT p.id
    FROM projects p
    INNER JOIN descendant_ids di ON p.parent_project_id = di.id
)
SELECT id FROM descendant_ids`,
		token.ProjectID,
	).Find(&descendantIDs)
	if err != nil {
		return nil, err
	}

	return descendantIDs, nil
}

// ProjectScopeContains checks if the given project ID is within the token's project scope.
// If scopedProjectIDs is nil, the token is unscoped and all projects are allowed.
func ProjectScopeContains(scopedProjectIDs []int64, projectID int64) bool {
	if scopedProjectIDs == nil {
		return true
	}

	for _, id := range scopedProjectIDs {
		if id == projectID {
			return true
		}
	}

	return false
}
