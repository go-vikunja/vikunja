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

// CanCreate checks if a user can create a new bucket
func (b *Bucket) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	pv, err := GetProjectViewByIDAndProject(s, b.ProjectViewID, b.ProjectID)
	if err != nil {
		return false, err
	}

	return canWriteBucketProject(s, pv.ProjectID, a)
}

// CanUpdate checks if a user can update an existing bucket
func (b *Bucket) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	return b.canDoBucket(s, a)
}

// CanDelete checks if a user can delete an existing bucket
func (b *Bucket) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return b.canDoBucket(s, a)
}

// canDoBucket checks if the bucket exists and if the user has the permission to act on it
func (b *Bucket) canDoBucket(s *xorm.Session, a web.Auth) (bool, error) {
	bb, err := getBucketByID(s, b.ID)
	if err != nil {
		return false, err
	}
	pv, err := GetProjectViewByIDAndProject(s, bb.ProjectViewID, b.ProjectID)
	if err != nil {
		return false, err
	}

	return canWriteBucketProject(s, pv.ProjectID, a)
}

// Not Project.CanUpdate: that swallows the archived error to allow un-archiving, buckets must not get that.
func canWriteBucketProject(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
	if isInstanceAdmin(s, a) {
		return true, nil
	}

	if fid := GetSavedFilterIDFromProjectID(projectID); fid > 0 {
		sf := &SavedFilter{ID: fid}
		return sf.CanUpdate(s, a)
	}

	p := &Project{ID: projectID}
	return p.CanWrite(s, a)
}
