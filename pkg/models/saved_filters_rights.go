// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import "code.vikunja.io/web"

// CanRead checks if a user has the right to read a saved filter
func (s *SavedFilter) CanRead(auth web.Auth) (bool, int, error) {
	can, err := s.canDoFilter(auth)
	return can, int(RightAdmin), err
}

// CanDelete checks if a user has the right to delete a saved filter
func (s *SavedFilter) CanDelete(auth web.Auth) (bool, error) {
	return s.canDoFilter(auth)
}

// CanUpdate checks if a user has the right to update a saved filter
func (s *SavedFilter) CanUpdate(auth web.Auth) (bool, error) {
	// A normal check would replace the passed struct which in our case would override the values we want to update.
	sf := &SavedFilter{ID: s.ID}
	return sf.canDoFilter(auth)
}

// CanCreate checks if a user has the right to update a saved filter
func (s *SavedFilter) CanCreate(auth web.Auth) (bool, error) {
	if _, is := auth.(*LinkSharing); is {
		return false, nil
	}

	return true, nil
}

// Helper function to check saved filter rights sind they all have the same logic
func (s *SavedFilter) canDoFilter(auth web.Auth) (can bool, err error) {
	// Link shares can't view or modify saved filters, therefore we can error out right away
	if _, is := auth.(*LinkSharing); is {
		return false, ErrSavedFilterNotAvailableForLinkShare{LinkShareID: auth.GetID(), SavedFilterID: s.ID}
	}

	sf, err := getSavedFilterSimpleByID(s.ID)
	if err != nil {
		return false, err
	}

	// Only owners are allowed to do something with a saved filter
	if sf.OwnerID != auth.GetID() {
		return false, nil
	}

	*s = *sf

	return true, nil
}
