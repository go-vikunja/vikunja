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

package migration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSchemaPlacement(t *testing.T) {
	t.Run("fresh install", func(t *testing.T) {
		require.NoError(t, validateSchemaPlacement("public", nil))
	})
	t.Run("data in active schema", func(t *testing.T) {
		require.NoError(t, validateSchemaPlacement("public", []string{"public"}))
	})
	t.Run("data in active schema with leftovers elsewhere", func(t *testing.T) {
		require.NoError(t, validateSchemaPlacement("vikunja", []string{"public", "vikunja"}))
	})
	t.Run("data only in another schema", func(t *testing.T) {
		err := validateSchemaPlacement("public", []string{"vikunja"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "vikunja")
		assert.Contains(t, err.Error(), `"public"`)
		assert.Contains(t, err.Error(), "database.schema")
	})
	t.Run("configured schema does not exist", func(t *testing.T) {
		err := validateSchemaPlacement("", []string{"vikunja"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})
}
