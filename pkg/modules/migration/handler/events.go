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

package handler

import (
	"code.vikunja.io/api/pkg/user"
)

// MigrationRequestedEvent represents a MigrationRequestedEvent event
type MigrationRequestedEvent struct {
	Migrator     interface{} `json:"migrator"`
	User         *user.User  `json:"user"`
	MigratorKind string      `json:"migrator_kind"`
}

// Name defines the name for MigrationRequestedEvent
func (t *MigrationRequestedEvent) Name() string {
	return "migration.requested"
}
