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
	"errors"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// CanWrite return whether the user can write on that project or not
func (p *Project) CanWrite(s *xorm.Session, a web.Auth) (bool, error) {

	// The favorite project can't be edited
	if p.ID == FavoritesPseudoProject.ID {
		return false, nil
	}

	// Get the project and check the permission
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
			(shareAuth.Permission == PermissionWrite || shareAuth.Permission == PermissionAdmin), errIsArchived
	}

	u := &user.User{ID: a.GetID()}

	// Check if the user is either owner or can write to the project
	if originalProject.isOwner(u) {
		canWrite = true
	}

	if canWrite {
		return canWrite, errIsArchived
	}

	canWrite, _, err = originalProject.checkPermission(s, u, PermissionWrite, PermissionAdmin)
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
		return true, int(PermissionRead), nil
	}

	// Saved Filter Projects need a special case
	if GetSavedFilterIDFromProjectID(p.ID) > 0 {
		sf := &SavedFilter{ID: GetSavedFilterIDFromProjectID(p.ID)}
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
			(shareAuth.Permission == PermissionRead || shareAuth.Permission == PermissionWrite || shareAuth.Permission == PermissionAdmin), int(shareAuth.Permission), nil
	}

	return p.checkPermission(s, &user.User{ID: a.GetID()}, PermissionRead, PermissionWrite, PermissionAdmin)
}

// CanUpdate checks if the user can update a project
func (p *Project) CanUpdate(s *xorm.Session, a web.Auth) (canUpdate bool, err error) {
	// The favorite project can't be edited
	if p.ID == FavoritesPseudoProject.ID {
		return false, nil
	}

	fid := GetSavedFilterIDFromProjectID(p.ID)
	if fid > 0 {
		sf, err := GetSavedFilterSimpleByID(s, fid)
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

// IsAdmin returns whether the user has admin permissions on the project or not
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
		return originalProject.ID == shareAuth.ProjectID && shareAuth.Permission == PermissionAdmin, nil
	}

	u := &user.User{ID: a.GetID()}

	// Check all the things
	// Check if the user is either owner or can write to the project
	// Owners are always admins
	if originalProject.isOwner(u) {
		return true, nil
	}
	is, _, err := originalProject.checkPermission(s, u, PermissionAdmin)
	return is, err
}

// Little helper function to check if a user is project owner
func (p *Project) isOwner(u *user.User) bool {
	return p.OwnerID == u.ID
}

// Checks n different permissions for any given user
func (p *Project) checkPermission(s *xorm.Session, u *user.User, permissions ...Permission) (bool, int, error) {
	projectPermissions, err := checkPermissionsForProjects(s, u, []int64{p.ID})
	if err != nil {
		return false, 0, err
	}
	permission, has := projectPermissions[p.ID]
	if !has {
		return false, 0, nil
	}

	for _, r := range permissions {
		if r == permission.MaxPermission {
			return true, int(permission.MaxPermission), nil
		}
	}

	return false, 0, nil
}

type projectPermission struct {
	ID            int64 `xorm:"pk autoincr"`
	MaxPermission Permission
}

func checkPermissionsForProjects(s *xorm.Session, u *user.User, projectIDs []int64) (projectPermissionMap map[int64]*projectPermission, err error) {
	projectPermissionMap = make(map[int64]*projectPermission)

	if len(projectIDs) < 1 {
		return
	}

	args := []interface{}{
		u.ID,
		u.ID,
		u.ID,
		u.ID,
		u.ID,
		u.ID,
		u.ID,
	}

	err = s.SQL(`
WITH RECURSIVE
    project_hierarchy AS (
        -- Base case: Start with the specified projects
        SELECT id,
               parent_project_id,
               0  AS level,
               id AS original_project_id
        FROM projects
        WHERE id IN (`+utils.JoinInt64Slice(projectIDs, ", ")+`)

        UNION ALL

        -- Recursive case: Traverse up the hierarchy
        SELECT p.id,
               p.parent_project_id,
               ph.level + 1,
               ph.original_project_id
        FROM projects p
                 INNER JOIN project_hierarchy ph ON p.id = ph.parent_project_id),

    project_permissions AS (SELECT ph.id,
                                   ph.original_project_id,
                                   CASE
                                       WHEN p.owner_id = ? THEN 2
                                       WHEN COALESCE(ul.permission, 0) > COALESCE(tl.permission, 0) THEN ul.permission
                                       ELSE COALESCE(tl.permission, 0)
                                       END AS project_permission,
            CASE
                WHEN p.owner_id = ? THEN 1  -- Direct project ownership
                ELSE ph.level + 1  -- Derived from parent project
            END AS priority
                            FROM project_hierarchy ph
                                LEFT JOIN projects p
                            ON ph.id = p.id
                                LEFT JOIN users_projects ul ON ul.project_id = ph.id AND ul.user_id = ?
                                LEFT JOIN team_projects tl ON tl.project_id = ph.id
                                LEFT JOIN team_members tm ON tm.team_id = tl.team_id AND tm.user_id = ?
                            WHERE p.owner_id = ? OR ul.user_id = ? OR tm.user_id = ?)

SELECT ph.original_project_id AS id,
       COALESCE(MAX(pp.project_permission), -1) AS max_permission
FROM project_hierarchy ph
         LEFT JOIN (SELECT *,
                           ROW_NUMBER() OVER (PARTITION BY original_project_id ORDER BY priority) AS rn
                    FROM project_permissions) pp ON ph.id = pp.id AND pp.rn = 1
GROUP BY ph.original_project_id`, args...).
		Find(&projectPermissionMap)
	return
}
