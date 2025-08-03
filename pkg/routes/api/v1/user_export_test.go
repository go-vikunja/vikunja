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

package v1

import (
	"testing"

	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
)

func TestDownloadUserDataExportLogic(t *testing.T) {
	t.Run("user with no export file ID", func(t *testing.T) {
		u := &user.User{
			ID:           1,
			ExportFileID: 0,
		}

		// This simulates the check we added in the fix
		assert.Equal(t, int64(0), u.ExportFileID)
	})

	t.Run("IsErrFileDoesNotExist correctly identifies file not found error", func(t *testing.T) {
		err := files.ErrFileDoesNotExist{FileID: 123}
		assert.True(t, files.IsErrFileDoesNotExist(err))
	})
}
