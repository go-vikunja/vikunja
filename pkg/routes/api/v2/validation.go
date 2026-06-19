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

package apiv2

import (
	"reflect"
	"sort"

	"code.vikunja.io/api/pkg/models"

	"github.com/asaskevich/govalidator"
)

// validateInputBody runs govalidator over the request body so v2 enforces the
// `valid:` tag rules (required, url, …) that Huma's schema validation doesn't,
// matching v1. The payload sits in an input field named Body by convention;
// inputs without one (read/list/delete) validate to nil.
func validateInputBody(in any) error {
	v := reflect.Indirect(reflect.ValueOf(in))
	if v.Kind() != reflect.Struct {
		return nil
	}
	body := v.FieldByName("Body")
	if !body.IsValid() || !body.CanInterface() {
		return nil
	}
	// Only struct bodies carry `valid:` tags. Binary/primitive bodies (e.g. the
	// avatar endpoint's []byte) would make govalidator.ValidateStruct error out.
	if reflect.Indirect(body).Kind() != reflect.Struct {
		return nil
	}
	if _, err := govalidator.ValidateStruct(body.Interface()); err != nil {
		byField := govalidator.ErrorsByField(err)
		fields := make([]string, 0, len(byField))
		for field, msg := range byField {
			fields = append(fields, field+": "+msg)
		}
		// Map iteration order is non-deterministic; sort for a stable errors[].
		sort.Strings(fields)
		return models.InvalidFieldError(fields)
	}
	return nil
}
