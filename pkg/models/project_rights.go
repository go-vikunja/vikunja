// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"errors"
	"strings"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// CanWrite return whether the user can write on that project or not
func (p *Project) CanWrite(s *xorm.Session, a web.Auth) (bool, error) {

	// The favorite project can't be edited
	if p.ID == FavoritesPseudoProject.ID {
		return false, nil
	}

	// Get the project and check the right
	originalProject, err := GetProjectSimpleByID(s, p.ID)
	if err != nil {
		return false, err
	}

	// We put the result of the is archived check in a separate variable to be able to return it later without
	// needing to recheck it again
	errIsArchived := originalProject.CheckIsArchived(s)

	var canWrite bool

	// Check if we're dealing with a share auth
	shareAuth, ok := a.(*LinkSharing)
	if ok {
		return originalProject.ID == shareAuth.ProjectID &&
			(shareAuth.Right == RightWrite || shareAuth.Right == RightAdmin), errIsArchived
	}

	// Check if the user is either owner or can write to the project
	if originalProject.isOwner(&user.User{ID: a.GetID()}) {
		canWrite = true
	}

	if canWrite {
		return canWrite, errIsArchived
	}

	canWrite, _, err = originalProject.checkRight(s, a, RightWrite, RightAdmin)
	if err != nil {
		return false, err
	}
	return canWrite, errIsArchived
}

// CanRead checks if a user has read access to a project
func (p *Project) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {

	// The favorite project needs a special treatment
	if p.ID == FavoritesPseudoProject.ID {
		owner, err := user.GetFromAuth(a)
		if err != nil {
			return false, 0, err
		}

		*p = FavoritesPseudoProject
		p.Owner = owner
		return true, int(RightRead), nil
	}

	// Saved Filter Projects need a special case
	if getSavedFilterIDFromProjectID(p.ID) > 0 {
		sf := &SavedFilter{ID: getSavedFilterIDFromProjectID(p.ID)}
		return sf.CanRead(s, a)
	}

	// Check if the user is either owner or can read
	var err error
	originalProject, err := GetProjectSimpleByID(s, p.ID)
	if err != nil {
		return false, 0, err
	}

	*p = *originalProject

	// Check if we're dealing with a share auth
	shareAuth, ok := a.(*LinkSharing)
	if ok {
		return p.ID == shareAuth.ProjectID &&
			(shareAuth.Right == RightRead || shareAuth.Right == RightWrite || shareAuth.Right == RightAdmin), int(shareAuth.Right), nil
	}

	return p.checkRight(s, a, RightRead, RightWrite, RightAdmin)
}

// CanUpdate checks if the user can update a project
func (p *Project) CanUpdate(s *xorm.Session, a web.Auth) (canUpdate bool, err error) {
	// The favorite project can't be edited
	if p.ID == FavoritesPseudoProject.ID {
		return false, nil
	}

	fid := getSavedFilterIDFromProjectID(p.ID)
	if fid > 0 {
		sf, err := getSavedFilterSimpleByID(s, fid)
		if err != nil {
			return false, err
		}

		return sf.CanUpdate(s, a)
	}

	// Get the project
	ol, err := GetProjectSimpleByID(s, p.ID)
	if err != nil {
		return false, err
	}

	// Check if we're moving the project to a different parent project.
	// If that is the case, we need to verify permissions to do so.
	if p.ParentProjectID != 0 && p.ParentProjectID != ol.ParentProjectID {
		newProject := &Project{ID: p.ParentProjectID}
		can, err := newProject.CanWrite(s, a)
		if err != nil {
			return false, err
		}
		if !can {
			return false, ErrGenericForbidden{}
		}
	}

	canUpdate, err = p.CanWrite(s, a)
	// If the project is archived and the user tries to un-archive it, let the request through
	archivedErr := ErrProjectIsArchived{}
	is := errors.As(err, &archivedErr)
	if is && !p.IsArchived && archivedErr.ProjectID == p.ID {
		err = nil
	}
	return canUpdate, err
}

// CanDelete checks if the user can delete a project
func (p *Project) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return p.IsAdmin(s, a)
}

// CanCreate checks if the user can create a project
func (p *Project) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	if p.ParentProjectID != 0 {
		parent := &Project{ID: p.ParentProjectID}
		return parent.CanWrite(s, a)
	}
	// Check if we're dealing with a share auth
	_, is := a.(*LinkSharing)
	if is {
		return false, nil
	}
	return true, nil
}

// IsAdmin returns whether the user has admin rights on the project or not
func (p *Project) IsAdmin(s *xorm.Session, a web.Auth) (bool, error) {
	// The favorite project can't be edited
	if p.ID == FavoritesPseudoProject.ID {
		return false, nil
	}

	originalProject, err := GetProjectSimpleByID(s, p.ID)
	if err != nil {
		return false, err
	}

	// Check if we're dealing with a share auth
	shareAuth, ok := a.(*LinkSharing)
	if ok {
		return originalProject.ID == shareAuth.ProjectID && shareAuth.Right == RightAdmin, nil
	}

	// Check all the things
	// Check if the user is either owner or can write to the project
	// Owners are always admins
	if originalProject.isOwner(&user.User{ID: a.GetID()}) {
		return true, nil
	}
	is, _, err := originalProject.checkRight(s, a, RightAdmin)
	return is, err
}

// Little helper function to check if a user is project owner
func (p *Project) isOwner(u *user.User) bool {
	return p.OwnerID == u.ID
}

// Checks n different rights for any given user
func (p *Project) checkRight(s *xorm.Session, a web.Auth, rights ...Right) (bool, int, error) {
	projectRights, err := checkRightsForProjects(s, a, []int64{p.ID})
	if err != nil {
		return false, 0, err
	}
	right, has := projectRights[p.ID]
	if !has {
		return false, 0, nil
	}

	for _, r := range rights {
		if r == right.MaxRight {
			return true, int(right.MaxRight), nil
		}
	}

	return false, 0, nil
}

type projectRight struct {
	ID       int64 `xorm:"pk autoincr"`
	MaxRight Right
}

func checkRightsForProjects(s *xorm.Session, a web.Auth, projectIDs []int64) (projectRightMap map[int64]*projectRight, err error) {
	projectRightMap = make(map[int64]*projectRight)
	whereIDIn := strings.Repeat("?,", len(projectIDs))[:len(projectIDs)*2-1]
	args := []interface{}{
		a.GetID(),
		a.GetID(),
		a.GetID(),
		int64(0),
	}
	for _, id := range projectIDs {
		args = append(args, id)
	}

	err = s.SQL(`WITH RECURSIVE
    all_projects AS (SELECT p.id,
                            p.parent_project_id,
                            CASE
                                WHEN p.owner_id = 1 THEN 2
                                WHEN COALESCE(ul.right, 0) > COALESCE(tl.right, 0) THEN ul.right
                                ELSE COALESCE(tl.right, 0)
                                END AS initial_right
                     FROM projects p
                              LEFT JOIN team_projects tl ON tl.project_id = p.id
                              LEFT JOIN team_members tm2 ON tm2.team_id = tl.team_id
                              LEFT JOIN users_projects ul ON ul.project_id = p.id
                     WHERE (tm2.user_id = ? OR ul.user_id = ? OR p.owner_id = ?)
                       AND (p.parent_project_id IS NULL OR p.parent_project_id = ? OR
                            ((tm2.user_id IS NOT NULL OR ul.user_id IS NOT NULL) AND
                             p.parent_project_id IS NOT NULL))
                       AND p.id in (`+whereIDIn+`)
                     GROUP BY p.id

                     UNION ALL

                     SELECT p.id, p.parent_project_id, ap.initial_right
                     FROM projects p
                              INNER JOIN all_projects ap ON p.parent_project_id = ap.id),

    project_max_rights AS (SELECT id, MAX(initial_right) AS max_right
                           FROM all_projects
                           GROUP BY id),

    inherited_rights AS (SELECT ap.id,
                                CASE
                                    WHEN COALESCE(pmr.max_right, 0) > COALESCE(parent.max_right, 0)
                                        THEN COALESCE(pmr.max_right, 0)
                                    ELSE COALESCE(parent.max_right, 0)
                                    END AS inherited_right
                         FROM all_projects ap
                                  LEFT JOIN project_max_rights pmr ON ap.id = pmr.id
                                  LEFT JOIN project_max_rights parent ON ap.parent_project_id = parent.id)

SELECT DISTINCT ap.id,
                ir.inherited_right AS max_right
FROM all_projects ap
         LEFT JOIN all_projects np ON ap.parent_project_id = np.id
         LEFT JOIN inherited_rights ir ON ap.id = ir.id`, args...).Find(&projectRightMap)
	return
}
