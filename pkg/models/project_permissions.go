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
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// CanWrite return whether the user can write on that project or not
func (p *Project) CanWrite(s *xorm.Session, a web.Auth) (bool, error) {

	// Favorites and saved filters aggregate tasks from real projects, they have no row of their own to write to.
	if p.ID < 1 {
		return false, nil
	}

	if isInstanceAdmin(s, a) {
		return true, nil
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

	canWrite, err = originalProject.checkPermission(s, u, PermissionWrite, PermissionAdmin)
	if err != nil {
		return false, err
	}
	return canWrite, errIsArchived
}

// projectReadPermission holds the read access of one auth subject for one project.
// project is nil for pseudo projects without a row of their own, so callers know not to overwrite theirs.
type projectReadPermission struct {
	canRead       bool
	maxPermission int
	project       *Project
}

// checkReadPermissionsForProjects resolves read access of a single auth subject for many projects at
// once. Unless it returns an error, the result has an entry for every requested id.
func checkReadPermissionsForProjects(s *xorm.Session, a web.Auth, projectIDs []int64) (map[int64]*projectReadPermission, error) {
	permissions := make(map[int64]*projectReadPermission, len(projectIDs))

	if len(projectIDs) == 0 {
		return permissions, nil
	}

	// Resolve pseudo ids before the instance admin branch below: they have no row to look up.
	projectIDsWithRow := make([]int64, 0, len(projectIDs))
	for _, projectID := range projectIDs {
		switch {
		case projectID == FavoritesPseudoProject.ID:
			owner, err := user.GetFromAuth(a)
			if err != nil {
				return nil, err
			}

			favorites := FavoritesPseudoProject
			favorites.Owner = owner
			permissions[projectID] = &projectReadPermission{
				canRead:       true,
				maxPermission: int(PermissionRead),
				project:       &favorites,
			}
		case GetSavedFilterIDFromProjectID(projectID) > 0:
			sf := &SavedFilter{ID: GetSavedFilterIDFromProjectID(projectID)}
			canRead, maxPermission, err := sf.CanRead(s, a)
			if err != nil {
				return nil, err
			}

			permissions[projectID] = &projectReadPermission{
				canRead:       canRead,
				maxPermission: maxPermission,
			}
		default:
			projectIDsWithRow = append(projectIDsWithRow, projectID)
		}
	}

	if len(projectIDsWithRow) == 0 {
		return permissions, nil
	}

	projects, err := requireProjectsByIDs(s, projectIDsWithRow)
	if err != nil {
		return nil, err
	}

	if isInstanceAdmin(s, a) {
		for _, projectID := range projectIDsWithRow {
			permissions[projectID] = &projectReadPermission{
				canRead:       true,
				maxPermission: int(PermissionAdmin),
				project:       projects[projectID],
			}
		}
		return permissions, nil
	}

	if shareAuth, is := a.(*LinkSharing); is {
		for _, projectID := range projectIDsWithRow {
			permissions[projectID] = &projectReadPermission{
				canRead: projectID == shareAuth.ProjectID &&
					(shareAuth.Permission == PermissionRead || shareAuth.Permission == PermissionWrite || shareAuth.Permission == PermissionAdmin),
				maxPermission: int(shareAuth.Permission),
				project:       projects[projectID],
			}
		}
		return permissions, nil
	}

	projectPermissions, err := checkPermissionsForProjects(s, &user.User{ID: a.GetID()}, projectIDsWithRow)
	if err != nil {
		return nil, err
	}

	for _, projectID := range projectIDsWithRow {
		permission := &projectReadPermission{project: projects[projectID]}
		if pp, has := projectPermissions[projectID]; has {
			permission.canRead = true
			permission.maxPermission = int(pp)
		}
		permissions[projectID] = permission
	}

	return permissions, nil
}

// requireProjectsByIDs loads all given projects, failing with the same error GetProjectSimpleByID
// gives for the first id without a row.
func requireProjectsByIDs(s *xorm.Session, projectIDs []int64) (map[int64]*Project, error) {
	projects, err := GetProjectsMapByIDs(s, projectIDs)
	if err != nil {
		return nil, err
	}

	for _, projectID := range projectIDs {
		if _, has := projects[projectID]; !has {
			return nil, ErrProjectDoesNotExist{ID: projectID}
		}
	}

	return projects, nil
}

// CanRead checks if a user has read access to a project
func (p *Project) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	permissions, err := checkReadPermissionsForProjects(s, a, []int64{p.ID})
	if err != nil {
		return false, 0, err
	}

	permission, has := permissions[p.ID]
	if !has {
		return false, 0, nil
	}

	if permission.project != nil {
		*p = *permission.project
	}

	return permission.canRead, permission.maxPermission, nil
}

// CanUpdate checks if the user can update a project
func (p *Project) CanUpdate(s *xorm.Session, a web.Auth) (canUpdate bool, err error) {
	// The favorite project can't be edited
	if p.ID == FavoritesPseudoProject.ID {
		return false, nil
	}

	// Ahead of the admin bypass: a filter's pseudo project is the filter, and only its owner may update it.
	fid := GetSavedFilterIDFromProjectID(p.ID)
	if fid > 0 {
		sf, err := GetSavedFilterSimpleByID(s, fid)
		if err != nil {
			return false, err
		}

		return sf.CanUpdate(s, a)
	}

	if isInstanceAdmin(s, a) {
		return true, nil
	}

	// Get the project
	ol, err := GetProjectSimpleByID(s, p.ID)
	if err != nil {
		return false, err
	}

	// Check if we're moving the project to a different parent project.
	// If that is the case, we need to verify permissions to do so.
	//
	// The reparent Admin gate for GHSA-2vq4-854f-5c72 / CVE-2026-35595
	// lives in UpdateProject, not here: CanUpdate is reused by
	// permission-check-only callers (buckets, webhooks, task ops) that
	// pass stub &Project{ID: ...} values with ParentProjectID=0 and never
	// commit a reparent, which would spuriously trip the gate.
	// Only a real new parent (> 0) needs a write check here; detach-to-root
	// (explicit 0) is gated for Admin in UpdateProject instead.
	if p.ParentProjectID != nil && *p.ParentProjectID > 0 && *p.ParentProjectID != ol.parentID() {
		newProject := &Project{ID: *p.ParentProjectID}
		can, err := newProject.CanWrite(s, a)
		if err != nil {
			return false, err
		}
		if !can {
			return false, ErrGenericForbidden{}
		}
	}

	canUpdate, err = p.CanWrite(s, a)
	// Un-archiving an archived project is allowed here; whether its parent
	// still is archived is checked in UpdateProject.
	archivedErr := ErrProjectIsArchived{}
	is := errors.As(err, &archivedErr)
	if is && !p.IsArchived && archivedErr.ProjectID == p.ID {
		err = nil
	}
	return canUpdate, err
}

// CanDelete checks if the user can delete a project
func (p *Project) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	// IsAdmin covers the instance admin bypass, but only after denying pseudo projects.
	return p.IsAdmin(s, a)
}

// CanCreate checks if the user can create a project
func (p *Project) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	if isInstanceAdmin(s, a) {
		return true, nil
	}
	if pid := p.parentID(); pid > 0 {
		parent := &Project{ID: pid}
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
	// Pseudo projects have no ACL of their own, so nobody is their admin.
	if IsPseudoProjectID(p.ID) {
		return false, nil
	}

	if isInstanceAdmin(s, a) {
		return true, nil
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
	is, err := originalProject.checkPermission(s, u, PermissionAdmin)
	return is, err
}

// Little helper function to check if a user is project owner
func (p *Project) isOwner(u *user.User) bool {
	return p.OwnerID == u.ID
}

// Checks n different permissions for any given user
func (p *Project) checkPermission(s *xorm.Session, u *user.User, permissions ...Permission) (bool, error) {
	projectPermissions, err := checkPermissionsForProjects(s, u, []int64{p.ID})
	if err != nil {
		return false, err
	}
	permission, has := projectPermissions[p.ID]
	if !has {
		return false, nil
	}

	for _, r := range permissions {
		if r == permission {
			return true, nil
		}
	}

	return false, nil
}

// checkPermissionsForProjects returns the effective permission of the user on each of
// the given projects. Projects the user cannot access are absent from the result.
func checkPermissionsForProjects(s *xorm.Session, u *user.User, projectIDs []int64) (map[int64]Permission, error) {
	permissions := make(map[int64]Permission, len(projectIDs))
	if len(projectIDs) < 1 {
		return permissions, nil
	}

	access, err := getProjectAccessForUser(s, u.ID)
	if err != nil {
		return nil, err
	}
	for _, id := range projectIDs {
		if p, has := access.permission(id); has {
			permissions[id] = p
		}
	}
	return permissions, nil
}
