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

package v2

import "time"

// CaldavToken represents a caldav token for a user.
type CaldavToken struct {
	ID      int64     `json:"id"`
	Token   string    `json:"token,omitempty"`
	Created time.Time `json:"created"`
	Links   *CaldavTokenLinks `json:"_links"`
}

// CaldavTokenLinks represents the links for a caldav token.
type CaldavTokenLinks struct {
	Self *Link `json:"self"`
}
