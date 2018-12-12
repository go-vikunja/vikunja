//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import "code.vikunja.io/api/pkg/metrics"

// DeleteUserByID deletes a user by its ID
func DeleteUserByID(id int64, doer *User) error {
	// Check if the id is 0
	if id == 0 {
		return ErrIDCannotBeZero{}
	}

	// Delete the user
	_, err := x.Id(id).Delete(&User{})

	if err != nil {
		return err
	}

	// Update the metrics
	metrics.UpdateCount(-1, metrics.ActiveUsersKey)

	return err
}
