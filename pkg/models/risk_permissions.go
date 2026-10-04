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
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// Permissions of risks follow the project they belong to. A link share is not a person and has no
// access to risks.

func isLinkShareAuth(a web.Auth) bool {
	_, ok := a.(*LinkSharing)
	return ok
}

// CanRead loads the risk and checks read access to its project. It leaves the risk in r, so that
// ReadOne does not read it again.
func (r *Risk) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	if isLinkShareAuth(a) {
		return false, 0, nil
	}

	existing, err := getRiskByID(s, r.ID)
	if err != nil {
		return false, 0, err
	}
	project, err := GetProjectSimpleByID(s, existing.ProjectID)
	if err != nil {
		return false, 0, err
	}
	can, maxPermission, err := project.CanRead(s, a)
	if err != nil || !can {
		return false, maxPermission, err
	}

	*r = *existing
	r.loaded, r.project = existing, project
	return true, maxPermission, nil
}

// CanCreate needs write access to the project in r.ProjectID.
func (r *Risk) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	if isLinkShareAuth(a) {
		return false, nil
	}
	project, err := GetProjectSimpleByID(s, r.ProjectID)
	if err != nil {
		return false, err
	}
	can, err := project.CanWrite(s, a)
	if err != nil || !can {
		return false, err
	}
	r.project = project
	return true, nil
}

// canModify is the check for everything that changes an existing risk: write access to its project.
// The risk is loaded here, so that Update and Delete do not read it again.
func (r *Risk) canModify(s *xorm.Session, a web.Auth) (bool, error) {
	if isLinkShareAuth(a) {
		return false, nil
	}
	existing, err := getRiskByID(s, r.ID)
	if err != nil {
		return false, err
	}
	project, err := GetProjectSimpleByID(s, existing.ProjectID)
	if err != nil {
		return false, err
	}
	can, err := project.CanWrite(s, a)
	if err != nil || !can {
		return false, err
	}
	r.loaded, r.project = existing, project
	return true, nil
}

// CanUpdate needs write access to the project of the risk.
func (r *Risk) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	return r.canModify(s, a)
}

// CanDelete needs write access to the project of the risk. Delete reports the risk it removed, so it
// is filled in here too.
func (r *Risk) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	can, err := r.canModify(s, a)
	if err != nil || !can {
		return false, err
	}
	loaded, project := r.loaded, r.project
	*r = *loaded
	r.loaded, r.project = loaded, project
	return true, nil
}
