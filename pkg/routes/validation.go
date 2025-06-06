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

package routes

import (
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"

	"github.com/asaskevich/govalidator"
)

// CustomValidator is a dummy struct to use govalidator with echo
type CustomValidator struct{}

func init() {
	govalidator.TagMap["time"] = func(str string) bool {
		return govalidator.IsTime(str, "15:04")
	}

	// Custom validator for database TEXT fields that adapts to the database being used
	govalidator.TagMap["dbtext"] = func(str string) bool {
		// Get the current database dialect
		dialect := strings.ToLower(config.DatabaseType.GetString())

		// Default limit for MySQL and unknown databases (65KB safely under TEXT limit)
		maxLength := 65000

		// For databases that support larger text fields
		if dialect == "postgres" || dialect == "sqlite" || dialect == "sqlite3" {
			maxLength = 1048576 // ~1MB limit for PostgreSQL and SQLite
		}

		return len(str) <= maxLength
	}
}

// Validate validates stuff
func (cv *CustomValidator) Validate(i interface{}) error {
	if _, err := govalidator.ValidateStruct(i); err != nil {

		var errs []string
		for field, e := range govalidator.ErrorsByField(err) {
			errs = append(errs, field+": "+e)
		}

		return models.InvalidFieldError(errs)
	}
	return nil
}
