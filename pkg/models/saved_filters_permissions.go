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

// CanRead checks if a user has the permission to read a saved filter
func (sf *SavedFilter) CanRead(s *xorm.Session, auth web.Auth) (bool, int, error) {
	can, err := sf.canDoFilter(s, auth)
	return can, int(PermissionAdmin), err
}

// CanDelete checks if a user has the permission to delete a saved filter
func (sf *SavedFilter) CanDelete(s *xorm.Session, auth web.Auth) (bool, error) {
	return sf.canDoFilter(s, auth)
}

// CanUpdate checks if a user has the permission to update a saved filter
func (sf *SavedFilter) CanUpdate(s *xorm.Session, auth web.Auth) (bool, error) {
	// A normal check would replace the passed struct which in our case would override the values we want to update.
	sff := &SavedFilter{ID: sf.ID}
	return sff.canDoFilter(s, auth)
}

// CanCreate checks if a user has the permission to update a saved filter
func (sf *SavedFilter) CanCreate(_ *xorm.Session, auth web.Auth) (bool, error) {
	if _, is := auth.(*LinkSharing); is {
		return false, nil
	}

	return true, nil
}

// Helper function to check saved filter permissions sind they all have the same logic
func (sf *SavedFilter) canDoFilter(s *xorm.Session, auth web.Auth) (can bool, err error) {
	// Link shares can't view or modify saved filters, therefore we can error out right away
	if _, is := auth.(*LinkSharing); is {
		return false, ErrSavedFilterNotAvailableForLinkShare{LinkShareID: auth.GetID(), SavedFilterID: sf.ID}
	}

	sff, err := GetSavedFilterSimpleByID(s, sf.ID)
	if err != nil {
		return false, err
	}

	// Only owners are allowed to do something with a saved filter
	if sff.OwnerID != auth.GetID() {
		return false, nil
	}

	*sf = *sff

	return true, nil
}
