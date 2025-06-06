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

// Right defines the rights users/teams can have for projects
type Right int

// define unknown right
const (
	RightUnknown = -1
	// Can read projects in a
	RightRead Right = iota - 1
	// Can write in a like projects and tasks. Cannot create new projects.
	RightWrite
	// Can manage a project, can do everything
	RightAdmin
)

func (r Right) isValid() error {
	if r != RightAdmin && r != RightRead && r != RightWrite {
		return ErrInvalidRight{r}
	}

	return nil
}

// MarshalJSON marshals the enum as a quoted json string
func (r Right) MarshalJSON() ([]byte, error) {
	if r == RightUnknown {
		return []byte(`null`), nil
	}
	return []byte(strconv.Itoa(int(r))), nil
}

// UnmarshalJSON unmarshals a quoted json string to the enum value
func (r *Right) UnmarshalJSON(data []byte) error {
	var s int
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case -1:
		*r = RightUnknown
	case 0:
		*r = RightRead
	case 1:
		*r = RightWrite
	case 2:
		*r = RightAdmin
	default:
		return fmt.Errorf("invalid Right %q", s)
	}
	return nil
}
