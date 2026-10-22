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
	"strconv"
	"strings"

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
	body = reflect.Indirect(body)
	if body.Kind() != reflect.Struct {
		return nil
	}

	fields := append(structErrors(body.Interface(), ""), pointerSliceErrors(body)...)
	if len(fields) == 0 {
		return nil
	}
	// Map iteration order is non-deterministic; sort for a stable errors[].
	sort.Strings(fields)
	return models.InvalidFieldError(fields)
}

// structErrors returns govalidator failures as "<prefix><field>: <message>", the shape invalidFieldDetails expects.
func structErrors(s any, prefix string) []string {
	_, err := govalidator.ValidateStruct(s)
	if err == nil {
		return nil
	}
	byField := govalidator.ErrorsByField(err)
	fields := make([]string, 0, len(byField))
	for field, msg := range byField {
		fields = append(fields, prefix+field+": "+msg)
	}
	return fields
}

// pointerSliceErrors validates elements of []*T body fields, which govalidator walks past without ever applying their `valid:` tags.
func pointerSliceErrors(body reflect.Value) []string {
	var fields []string
	t := body.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		// readOnly fields are accepted on write so clients can round-trip a GET response; the models ignore them.
		if f.PkgPath != "" || f.Tag.Get("readOnly") == "true" || f.Tag.Get("valid") == "-" {
			continue
		}
		if f.Type.Kind() != reflect.Slice || f.Type.Elem().Kind() != reflect.Pointer || f.Type.Elem().Elem().Kind() != reflect.Struct {
			continue
		}
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "-" {
			continue
		}
		if name == "" {
			name = f.Name
		}
		v := body.Field(i)
		if !v.CanInterface() {
			continue
		}
		for j := 0; j < v.Len(); j++ {
			el := v.Index(j)
			if el.IsNil() {
				continue
			}
			// "tasks[1]." matches how Huma locates its own array-element errors.
			fields = append(fields, structErrors(el.Interface(), name+"["+strconv.Itoa(j)+"].")...)
		}
	}
	return fields
}
