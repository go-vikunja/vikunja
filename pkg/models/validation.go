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
	"sort"

	"code.vikunja.io/api/pkg/db"

	"github.com/asaskevich/govalidator"
)

// The custom validators the models' `valid:` tags reference are registered here
// rather than in the HTTP layer, so every entry point that validates a model
// (echo's CustomValidator, /api/v2, MCP) picks them up by importing this package.
func init() {
	govalidator.TagMap["time"] = func(str string) bool {
		return govalidator.IsTime(str, "15:04")
	}

	// MySQL TEXT tops out far below what PostgreSQL and SQLite accept.
	govalidator.TagMap["dbtext"] = func(str string) bool {
		maxLength := 65000
		if dialect := db.GetDialect(); dialect == "postgres" || dialect == "sqlite3" {
			maxLength = 1048576
		}
		return len(str) <= maxLength
	}
}

// ValidateStruct reports every `valid:` tag failure as an InvalidFieldError.
func ValidateStruct(i interface{}) error {
	return ValidateStructFields(i, nil)
}

// ValidateStructFields restricts ValidateStruct to the named fields (nil means all):
// a partial payload must not trip a `required` rule for a field it never sent.
func ValidateStructFields(i interface{}, only map[string]bool) error {
	_, err := govalidator.ValidateStruct(i)
	if err == nil {
		return nil
	}

	var errs []string
	for field, e := range govalidator.ErrorsByField(err) {
		if only != nil && !only[field] {
			continue
		}
		errs = append(errs, field+": "+e)
	}
	if len(errs) == 0 {
		return nil
	}

	// Map iteration order is non-deterministic; sort for a stable errors[].
	sort.Strings(errs)
	return InvalidFieldError(errs)
}
