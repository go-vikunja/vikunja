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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserDataExportStatus(t *testing.T) {
	t.Run("no export", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		status, err := GetUserDataExportStatus(&user.User{ID: 15})
		require.NoError(t, err)
		assert.Nil(t, status)
	})

	t.Run("with export", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		status, err := GetUserDataExportStatus(&user.User{ID: 1, ExportFileID: 1})
		require.NoError(t, err)
		require.NotNil(t, status)
		assert.Equal(t, int64(1), status.ID)
	})

	t.Run("export points at a missing file", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		// A dangling ExportFileID must read as "no export" rather than erroring,
		// matching the download path which 404s the same case.
		status, err := GetUserDataExportStatus(&user.User{ID: 15, ExportFileID: 9999})
		require.NoError(t, err)
		assert.Nil(t, status)
	})
}
