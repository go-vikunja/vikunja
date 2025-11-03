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
	"encoding/json"
	"fmt"
	"strconv"
)

// Permission defines the permissions users/teams can have for projects
type Permission int

// define unknown permission
const (
	PermissionUnknown = -1
	// Can read projects in a
	PermissionRead Permission = iota - 1
	// Can write in a like projects and tasks. Cannot create new projects.
	PermissionWrite
	// Can manage a project, can do everything
	PermissionAdmin
)

func (r Permission) isValid() error {
	if r != PermissionAdmin && r != PermissionRead && r != PermissionWrite {
		return ErrInvalidPermission{r}
	}

	return nil
}

// MarshalJSON marshals the enum as a quoted json string
func (r Permission) MarshalJSON() ([]byte, error) {
	if r == PermissionUnknown {
		return []byte(`null`), nil
	}
	return []byte(strconv.Itoa(int(r))), nil
}

// UnmarshalJSON unmarshals a quoted json string to the enum value
func (r *Permission) UnmarshalJSON(data []byte) error {
	var s int
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case -1:
		*r = PermissionUnknown
	case 0:
		*r = PermissionRead
	case 1:
		*r = PermissionWrite
	case 2:
		*r = PermissionAdmin
	default:
		return fmt.Errorf("invalid Permission %q", s)
	}
	return nil
}
