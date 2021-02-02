// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package migration

import (
	"code.vikunja.io/api/pkg/user"
)

// Migrator is the basic migrator interface which is shared among all migrators
type Migrator interface {
	// Migrate is the interface used to migrate a user's tasks from another platform to vikunja.
	// The user object is the user who's tasks will be migrated.
	Migrate(user *user.User) error
	// AuthURL returns a url for clients to authenticate against.
	// The use case for this are Oauth flows, where the server token should remain hidden and not
	// known to the frontend.
	AuthURL() string
	// Title holds the name of the migration.
	// This is used to show the name to users and to keep track of users who already migrated.
	Name() string
}
