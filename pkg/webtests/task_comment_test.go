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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskComments(t *testing.T) {
	testHandler := webHandlerTest{
		user: &testuser1,
		strFunc: func() handler.CObject {
			return &models.TaskComment{}
		},
		t: t,
	}
	testHandlerLinkShareWrite := webHandlerTest{
		linkShare: &models.LinkSharing{
			ID:          2,
			Hash:        "test2",
			ProjectID:   2,
			Permission:  models.PermissionWrite,
			SharingType: models.SharingTypeWithoutPassword,
			SharedByID:  1,
		},
		strFunc: func() handler.CObject {
			return &models.TaskComment{}
		},
		t: t,
	}
	// Only run specific nested tests:
	// ^TestTaskComments$/^Update$/^Update_task_items$/^Removing_Assignees_null$
	t.Run("Update", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "1", "commentid": "1"}, `{"comment":"Lorem Ipsum"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "99999", "commentid": "9999"}, `{"comment":"Lorem Ipsum"}`)
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeTaskDoesNotExist)
		})
		t.Run("Permissions check", func(t *testing.T) {
			// Only the own comments can be updated
			t.Run("Forbidden", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "14", "commentid": "2"}, `{"comment":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "15", "commentid": "3"}, `{"comment":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "16", "commentid": "4"}, `{"comment":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "17", "commentid": "5"}, `{"comment":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "18", "commentid": "6"}, `{"comment":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "19", "commentid": "7"}, `{"comment":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "20", "commentid": "8"}, `{"comment":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "21", "commentid": "9"}, `{"comment":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "22", "commentid": "10"}, `{"comment":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "23", "commentid": "11"}, `{"comment":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "24", "commentid": "12"}, `{"comment":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "25", "commentid": "13"}, `{"comment":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "26", "commentid": "14"}, `{"comment":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
		})
	})
	t.Run("Delete", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "1", "commentid": "1"})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "99999", "commentid": "9999"})
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeTaskDoesNotExist)
		})
		t.Run("Permissions check", func(t *testing.T) {
			// Only the own comments can be deleted
			t.Run("Forbidden", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "14", "commentid": "2"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "15", "commentid": "3"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "16", "commentid": "4"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "17", "commentid": "5"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "18", "commentid": "6"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "19", "commentid": "7"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "20", "commentid": "8"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "21", "commentid": "9"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "22", "commentid": "10"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "23", "commentid": "11"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "24", "commentid": "12"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "25", "commentid": "13"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "26", "commentid": "14"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
		})
	})
	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "1"}, `{"comment":"Lorem Ipsum"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "9999"}, `{"comment":"Lorem Ipsum"}`)
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeTaskDoesNotExist)
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "34"}, `{"comment":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "15"}, `{"comment":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "16"}, `{"comment":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "17"}, `{"comment":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "18"}, `{"comment":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "19"}, `{"comment":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "20"}, `{"comment":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "21"}, `{"comment":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "22"}, `{"comment":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "23"}, `{"comment":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "24"}, `{"comment":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "25"}, `{"comment":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "26"}, `{"comment":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})
		})
		t.Run("Link Share", func(t *testing.T) {
			rec, err := testHandlerLinkShareWrite.testCreateWithLinkShare(nil, map[string]string{"task": "13"}, `{"comment":"Lorem Ipsum"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			db.AssertExists(t, "task_comments", map[string]interface{}{
				"task_id":   13,
				"comment":   "Lorem Ipsum",
				"author_id": -2,
			}, false)
		})
	})
}
