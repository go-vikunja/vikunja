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

func TestNormalizeHex(t *testing.T) {
	t.Run("no prefix", func(t *testing.T) {
		normalized := NormalizeHex("fff000")

		assert.Equal(t, "fff000", normalized)
	})
	t.Run("with # prefix", func(t *testing.T) {
		normalized := NormalizeHex("#fff000")

		assert.Equal(t, "fff000", normalized)
	})
	t.Run("longer than 6 chars with #", func(t *testing.T) {
		normalized := NormalizeHex("#fff000SA")

		assert.Equal(t, "fff000", normalized)
	})
	t.Run("longer than 6 chars", func(t *testing.T) {
		normalized := NormalizeHex("fff000SA")

		assert.Equal(t, "fff000", normalized)
	})
	t.Run("shorter than 6 chars with #", func(t *testing.T) {
		normalized := NormalizeHex("#fff")

		assert.Equal(t, "fff", normalized)
	})
	t.Run("shorter than 6 chars", func(t *testing.T) {
		normalized := NormalizeHex("fff")

		assert.Equal(t, "fff", normalized)
	})
}
