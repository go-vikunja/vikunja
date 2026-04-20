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

package humaapi

import (
	"github.com/danielgtaylor/huma/v2"
)

// vikunjaError is the JSON shape Vikunja's existing error_handler.go emits
// for string-style errors. We preserve it so frontend contracts don't change.
type vikunjaError struct {
	StatusCode int    `json:"-"`
	Code       int    `json:"code,omitempty"`
	Message    string `json:"message"`
}

func (e *vikunjaError) Error() string  { return e.Message }
func (e *vikunjaError) GetStatus() int { return e.StatusCode }

// NewVikunjaError produces an error that serializes to Vikunja's legacy shape
// (`{"message": "..."}` for plain errors; `{"code": X, "message": "..."}` when
// a domain code is supplied). Registered as huma.NewError so every Huma
// handler's error return routes through here.
func NewVikunjaError(status int, msg string, _ ...error) huma.StatusError {
	return &vikunjaError{StatusCode: status, Message: msg}
}

// Install replaces huma.NewError globally. Call once at init.
func Install() {
	huma.NewError = NewVikunjaError
}
