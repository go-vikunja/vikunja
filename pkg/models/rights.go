// Vikunja is a todo-list application to facilitate your life.
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

// Right defines the rights users/teams can have for lists/namespaces
type Right int

// define unknown right
const (
	RightUnknown = -1
)

// Enumerate all the team rights
const (
	// Can read lists in a
	RightRead Right = iota
	// Can write in a like lists and todo tasks. Cannot create new lists.
	RightWrite
	// Can manage a list/namespace, can do everything
	RightAdmin
)

func (r Right) isValid() error {
	if r != RightAdmin && r != RightRead && r != RightWrite {
		return ErrInvalidRight{r}
	}

	return nil
}
