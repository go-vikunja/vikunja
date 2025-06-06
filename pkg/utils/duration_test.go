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
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseISO8601Duration(t *testing.T) {
	t.Run("full example", func(t *testing.T) {
		dur := ParseISO8601Duration("P1DT1H1M1S")
		expected, _ := time.ParseDuration("25h1m1s")

		assert.Equal(t, expected, dur)
	})
	t.Run("negative duration", func(t *testing.T) {
		dur := ParseISO8601Duration("-P1DT1H1M1S")
		expected, _ := time.ParseDuration("-25h1m1s")

		assert.Equal(t, expected, dur)
	})
}
