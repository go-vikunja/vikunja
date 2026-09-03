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
	"sort"
	"strconv"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// projectAccess is every project a user can reach, with the effective permission.
// Absent id = no access.
type projectAccess struct {
	userID      int64
	permissions map[int64]Permission
	sortedIDs   []int64
}

func (pa *projectAccess) permission(projectID int64) (Permission, bool) {
	p, has := pa.permissions[projectID]
	return p, has
}

// One row per project and granting ancestor-or-self, so MAX over a project's rows is
// the greatest of its own grant and everything it inherits: a grant on a descendant
// can raise an inherited permission, never lower it. Binds the user id three times.
// tree uses UNION, not UNION ALL: deduplicating (id, permission) terminates on a
// parent_project_id cycle and caps the row count at three per project.
// The recursive step's join implies parent_project_id IS NOT NULL, which is why root
// projects store NULL: the partial index then covers real children only.
const projectAccessCTE = `
WITH RECURSIVE grants (project_id, permission) AS (
    SELECT project_id, MAX(permission)
    FROM (
        SELECT id AS project_id, 2 AS permission FROM projects WHERE owner_id = ?
        UNION ALL
        SELECT project_id, permission FROM users_projects WHERE user_id = ?
        UNION ALL
        SELECT tp.project_id, tp.permission
        FROM team_projects tp
        INNER JOIN team_members tm ON tm.team_id = tp.team_id
        WHERE tm.user_id = ?
    ) direct_grants
    GROUP BY project_id
),
tree (id, permission) AS (
    SELECT p.id, g.permission
    FROM projects p
    INNER JOIN grants g ON g.project_id = p.id
    UNION
    SELECT p.id, t.permission
    FROM projects p
    INNER JOIN tree t ON p.parent_project_id = t.id
)`

const projectAccessQuery = projectAccessCTE + `
SELECT id, MAX(permission) AS permission FROM tree GROUP BY id`

const projectAccessIDsQuery = projectAccessCTE + `
SELECT DISTINCT id FROM tree`

type projectAccessRow struct {
	ID         int64 `xorm:"id"`
	Permission int64 `xorm:"permission"`
}

// Resolves the whole reachable tree in one query, memoized per session.
func getProjectAccessForUser(s *xorm.Session, userID int64) (*projectAccess, error) {
	cacheKey := "project-access-" + strconv.FormatInt(userID, 10)
	if pa, has := db.GetCached[*projectAccess](s, cacheKey); has {
		return pa, nil
	}

	rows := []*projectAccessRow{}
	err := s.SQL(projectAccessQuery, userID, userID, userID).Find(&rows)
	if err != nil {
		return nil, err
	}

	pa := &projectAccess{
		userID:      userID,
		permissions: make(map[int64]Permission, len(rows)),
		sortedIDs:   make([]int64, 0, len(rows)),
	}
	for _, r := range rows {
		permission := Permission(r.Permission)
		// A grant outside the enum (manual db edit, restored dump) is no grant at all.
		if permission.isValid() != nil {
			continue
		}
		pa.permissions[r.ID] = permission
		pa.sortedIDs = append(pa.sortedIDs, r.ID)
	}
	sort.Slice(pa.sortedIDs, func(i, j int) bool { return pa.sortedIDs[i] < pa.sortedIDs[j] })
	db.SetCached(s, cacheKey, pa)
	return pa, nil
}

// Above this many ids the inlined list bloats the statement and defeats server-side
// plan caching, so the subquery form with its three bind parameters wins.
const maxInListSize = 1000

func (pa *projectAccess) cond(column string) builder.Cond {
	if len(pa.sortedIDs) > maxInListSize {
		return pa.condViaSubquery(column)
	}
	// The empty slice matters: an argument-less builder.In is dropped by Where and
	// matches every row, while In with an empty slice renders as 0=1.
	return builder.In(column, pa.sortedIDs)
}

// Keeps the bind parameter count at three whatever the tree size.
func (pa *projectAccess) condViaSubquery(column string) builder.Cond {
	return builder.In(column, builder.Expr(projectAccessIDsQuery, pa.userID, pa.userID, pa.userID))
}

// Includes projects inherited through a shared parent.
func accessibleProjectIDsCond(s *xorm.Session, a web.Auth, column string) (builder.Cond, error) {
	if share, ok := a.(*LinkSharing); ok {
		return builder.Eq{column: share.ProjectID}, nil
	}

	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, err
	}

	access, err := getProjectAccessForUser(s, u.ID)
	if err != nil {
		return nil, err
	}
	return access.cond(column), nil
}

// GetAllParentProjects returns the project itself and every ancestor, keyed by id.
func GetAllParentProjects(s *xorm.Session, projectID int64) (map[int64]*Project, error) {
	cacheKey := "parent-projects-" + strconv.FormatInt(projectID, 10)
	chain, has := db.GetCached[map[int64]*Project](s, cacheKey)
	if !has {
		chain = make(map[int64]*Project)
		err := s.SQL(`WITH RECURSIVE all_projects AS (
		    SELECT
		        p.*
		    FROM
		        projects p
		    WHERE
		        p.id = ?
		    UNION ALL
		    SELECT
		        p.*
		    FROM
		        projects p
		            INNER JOIN all_projects pc ON p.ID = pc.parent_project_id
		)
		SELECT DISTINCT * FROM all_projects`, projectID).Find(&chain)
		if err != nil {
			return nil, err
		}
		db.SetCached(s, cacheKey, chain)
	}

	// The memo outlives this call, so hand out a copy callers cannot corrupt.
	out := make(map[int64]*Project, len(chain))
	for id, p := range chain {
		project := *p
		if p.ParentProjectID != nil {
			parentID := *p.ParentProjectID
			project.ParentProjectID = &parentID
		}
		out[id] = &project
	}
	return out, nil
}
