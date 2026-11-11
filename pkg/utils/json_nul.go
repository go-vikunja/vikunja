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

package utils

import (
	"bytes"
	"errors"
)

// ErrJSONContainsNUL is returned for JSON that decodes to a string containing NUL.
var ErrJSONContainsNUL = errors.New("text must not contain NUL characters (\\u0000)")

var jsonNULEscape = []byte(`u0000`)

// CheckJSONForNUL rejects JSON containing a \u0000 escape. Raw NUL bytes are
// already invalid JSON, so the escape is the only way to smuggle one in.
func CheckJSONForNUL(data []byte) error {
	for i := 0; ; {
		j := bytes.Index(data[i:], jsonNULEscape)
		if j < 0 {
			return nil
		}
		j += i
		backslashes := 0
		for k := j - 1; k >= 0 && data[k] == '\\'; k-- {
			backslashes++
		}
		// An even run means the backslashes escape each other: literal "\u0000" text.
		if backslashes%2 == 1 {
			return ErrJSONContainsNUL
		}
		i = j + len(jsonNULEscape)
	}
}
