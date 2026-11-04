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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckJSONForNUL(t *testing.T) {
	cases := map[string]bool{
		`{"a":"b"}`:             false,
		`{"a":"\u0000"}`:        true,
		`{"a":"x\u0000y"}`:      true,
		`{"a":"\\u0000"}`:       false,
		`{"a":"\\\u0000"}`:      true,
		`{"a":"u0000"}`:         false,
		`{"a":"\\u0000\u0000"}`: true,
		`u0000`:                 false,
		`{"\u0000":1}`:          true,
	}
	for in, want := range cases {
		err := CheckJSONForNUL([]byte(in))
		if want {
			assert.ErrorIs(t, err, ErrJSONContainsNUL, in)
		} else {
			assert.NoError(t, err, in)
		}
	}
}
