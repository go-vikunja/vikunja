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

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/builder"
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

	if p.isOwner(&user.User{ID: a.GetID()}) {
		return true, int(RightAdmin), nil
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

	var conds []builder.Cond
	for _, r := range rights {
		// User conditions
		// If the project was shared directly with the user and the user has the right
		conds = append(conds, builder.And(
			builder.Eq{"ul.user_id": a.GetID()},
			builder.Eq{"ul.right": r},
		))

		// Team rights
		// If the project was shared directly with the team and the team has the right
		conds = append(conds, builder.And(
			builder.Eq{"tm2.user_id": a.GetID()},
			builder.Eq{"tl.right": r},
		))
	}

	type allProjectRights struct {
		UserProject *ProjectUser `xorm:"extends"`
		TeamProject *TeamProject `xorm:"extends"`
	}

	r := &allProjectRights{}
	var maxRight = 0
	exists, err := s.
		Select("p.*, ul.right, tl.right").
		Table("projects").
		Alias("p").
		// User stuff
		Join("LEFT", []string{"users_projects", "ul"}, "ul.project_id = p.id").
		// Team stuff
		Join("LEFT", []string{"team_projects", "tl"}, "p.id = tl.project_id").
		Join("LEFT", []string{"team_members", "tm2"}, "tm2.team_id = tl.team_id").
		// The actual condition
		Where(builder.And(
			builder.Or(
				conds...,
			),
			builder.Eq{"p.id": p.ID},
		)).
		Get(r)

	// If there's noting shared for this project, and it has a parent, go up the tree
	if !exists && p.ParentProjectID > 0 {
		parent, err := GetProjectSimpleByID(s, p.ParentProjectID)
		if err != nil {
			return false, 0, err
		}

		return parent.checkRight(s, a, rights...)
	}

	// Figure out the max right and return it
	if int(r.UserProject.Right) > maxRight {
		maxRight = int(r.UserProject.Right)
	}
	if int(r.TeamProject.Right) > maxRight {
		maxRight = int(r.TeamProject.Right)
	}

	return exists, maxRight, err
}
