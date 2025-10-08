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
	"code.vikunja.io/api/pkg/user"

	"xorm.io/xorm"
)

// UserExportServiceProvider defines the interface for the user export service
type UserExportServiceProvider interface {
	ExportUserData(s *xorm.Session, u *user.User) error
}

// Function variable for dependency injection (set by service layer)
var ExportUserDataFunc func(s *xorm.Session, u *user.User) error

// ExportUserData delegates to the service layer
func ExportUserData(s *xorm.Session, u *user.User) error {
	if ExportUserDataFunc != nil {
		return ExportUserDataFunc(s, u)
	}

	// This should never happen if services are properly initialized
	panic("ExportUserDataFunc not initialized - services.InitializeDependencies() must be called")
}
