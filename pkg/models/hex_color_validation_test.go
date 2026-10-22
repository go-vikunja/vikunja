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
	"testing"

	"code.vikunja.io/api/pkg/utils"

	"github.com/asaskevich/govalidator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHexColorValidation(t *testing.T) {
	entities := map[string]func(hex string) any{
		"task":    func(hex string) any { return &Task{Title: "test", HexColor: hex} },
		"label":   func(hex string) any { return &Label{Title: "test", HexColor: hex} },
		"project": func(hex string) any { return &Project{Title: "test", HexColor: hex} },
	}

	valid := []string{"", "fff", "#fff", "f0f0f0", "#F0F0F0"}
	// Each of these passed the old length-only check.
	invalid := []string{"1234567", "f0f0f0f", "#f0f0f", "ff00", "zzzzzz", "äöüäöüä"}

	for name, build := range entities {
		t.Run(name, func(t *testing.T) {
			for _, hex := range valid {
				_, err := govalidator.ValidateStruct(build(hex))
				require.NoError(t, err, "%q should be accepted", hex)
				assert.LessOrEqual(t, len(utils.NormalizeHex(hex)), 6)
			}
			for _, hex := range invalid {
				_, err := govalidator.ValidateStruct(build(hex))
				require.Error(t, err, "%q should be rejected", hex)
			}
		})
	}
}
