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

package webtests

import (
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/stretchr/testify/require"
)

func TestTaskAttachmentIDOR(t *testing.T) {
	t.Run("Cannot read attachment from inaccessible task via accessible task ID", func(t *testing.T) {
		// Attachment 4 belongs to task 34 (owned by user 13, inaccessible to testuser1).
		// Task 1 is accessible to testuser1.
		// Requesting GET /tasks/1/attachments/4 should fail because the attachment
		// does not belong to task 1.
		testHandler := webHandlerTest{
			user: &testuser1,
			strFunc: func() handler.CObject {
				return &models.TaskAttachment{}
			},
			t: t,
		}

		_, err := testHandler.testReadOneWithUser(nil, map[string]string{
			"task":       "1", // task accessible to testuser1
			"attachment": "4", // attachment belonging to task 34, NOT accessible to testuser1
		})
		require.Error(t, err)
		assertHandlerErrorCode(t, err, models.ErrCodeTaskAttachmentDoesNotExist)
	})
}
