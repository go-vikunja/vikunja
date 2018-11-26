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

// TeamRight defines the rights teams can have for lists/namespaces
type TeamRight int

// define unknown team right
const (
	TeamRightUnknown = -1
)

// Enumerate all the team rights
const (
	// Can read lists in a Team
	TeamRightRead TeamRight = iota
	// Can write tasks in a Team like lists and todo tasks. Cannot create new lists.
	TeamRightWrite
	// Can manage a list/namespace, can do everything
	TeamRightAdmin
)

func (r TeamRight) isValid() error {
	if r != TeamRightAdmin && r != TeamRightRead && r != TeamRightWrite {
		return ErrInvalidTeamRight{r}
	}

	return nil
}
