// Copyright 2019 Vikunja and contriubtors. All rights reserved.
//
// This file is part of Vikunja.
//
// Vikunja is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Vikunja is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Vikunja.  If not, see <https://www.gnu.org/licenses/>.

package migration

import "code.vikunja.io/api/pkg/models"

// Migrator is the basic migrator interface which is shared among all migrators
type Migrator interface {
	// Migrate is the interface used to migrate a user's tasks from another platform to vikunja.
	// The user object is the user who's tasks will be migrated.
	Migrate(user *models.User) error
	// AuthURL returns a url for clients to authenticate against.
	// The use case for this are Oauth flows, where the server token should remain hidden and not
	// known to the frontend.
	AuthURL() string
	// Name holds the name of the migration.
	// This is used to show the name to users and to keep track of users who already migrated.
	Name() string
}
