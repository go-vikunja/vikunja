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

	"code.vikunja.io/api/pkg/i18n"

	"github.com/stretchr/testify/assert"
)

func TestHumanizeDuration(t *testing.T) {

	i18n.Init()

	t.Run("one part", func(t *testing.T) {
		d := 1 * time.Hour
		dur := HumanizeDuration(d, "en")

		assert.Equal(t, "one hour", dur)
	})
	t.Run("amount > 1", func(t *testing.T) {
		d := 2 * time.Hour
		dur := HumanizeDuration(d, "en")

		assert.Equal(t, "2 hours", dur)
	})
	t.Run("2 parts", func(t *testing.T) {
		d := 2*time.Hour + 48*time.Hour
		dur := HumanizeDuration(d, "en")

		assert.Equal(t, "2 days and 2 hours", dur)
	})
	t.Run("multiple parts", func(t *testing.T) {
		d := 2*time.Hour + 24*15*time.Hour
		dur := HumanizeDuration(d, "en")

		assert.Equal(t, "2 weeks, one day and 2 hours", dur)
	})
	t.Run("years", func(t *testing.T) {
		day := 24 * time.Hour
		d := 2*time.Hour + 365*day + 14*day
		dur := HumanizeDuration(d, "en")

		assert.Equal(t, "one year, 2 weeks and 2 hours", dur)
	})
	t.Run("ignore seconds", func(t *testing.T) {
		d := 2*time.Hour + 48*time.Hour + 23*time.Second
		dur := HumanizeDuration(d, "en")

		assert.Equal(t, "2 days and 2 hours", dur)
	})
}
