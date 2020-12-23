// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package integrations

import (
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/web/handler"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestTaskComments(t *testing.T) {
	testHandler := webHandlerTest{
		user: &testuser1,
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
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "99999", "commentid": "9999"}, `{"comment":"Lorem Ipsum"}`)
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeTaskDoesNotExist)
		})
		t.Run("Rights check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "14", "commentid": "2"}, `{"comment":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "15", "commentid": "3"}, `{"comment":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "16", "commentid": "4"}, `{"comment":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "17", "commentid": "5"}, `{"comment":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "18", "commentid": "6"}, `{"comment":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "19", "commentid": "7"}, `{"comment":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "20", "commentid": "8"}, `{"comment":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})

			t.Run("Shared Via NamespaceTeam readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "21", "commentid": "9"}, `{"comment":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceTeam write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "22", "commentid": "10"}, `{"comment":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})
			t.Run("Shared Via NamespaceTeam admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "23", "commentid": "11"}, `{"comment":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})

			t.Run("Shared Via NamespaceUser readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "24", "commentid": "12"}, `{"comment":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceUser write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "25", "commentid": "13"}, `{"comment":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})
			t.Run("Shared Via NamespaceUser admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"task": "26", "commentid": "14"}, `{"comment":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})
		})
	})
	t.Run("Delete", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "1", "commentid": "1"})
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "99999", "commentid": "9999"})
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeTaskDoesNotExist)
		})
		t.Run("Rights check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "14", "commentid": "2"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "15", "commentid": "3"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "16", "commentid": "4"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "17", "commentid": "5"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "18", "commentid": "6"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "19", "commentid": "7"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "20", "commentid": "8"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})

			t.Run("Shared Via NamespaceTeam readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "21", "commentid": "9"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceTeam write", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "22", "commentid": "10"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared Via NamespaceTeam admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "23", "commentid": "11"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})

			t.Run("Shared Via NamespaceUser readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "24", "commentid": "12"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceUser write", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "25", "commentid": "13"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared Via NamespaceUser admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"task": "26", "commentid": "14"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
		})
	})
	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "1"}, `{"comment":"Lorem Ipsum"}`)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "9999"}, `{"comment":"Lorem Ipsum"}`)
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeTaskDoesNotExist)
		})
		t.Run("Rights check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "34"}, `{"comment":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "15"}, `{"comment":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "16"}, `{"comment":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "17"}, `{"comment":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "18"}, `{"comment":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "19"}, `{"comment":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "20"}, `{"comment":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})

			t.Run("Shared Via NamespaceTeam readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "21"}, `{"comment":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceTeam write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "22"}, `{"comment":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})
			t.Run("Shared Via NamespaceTeam admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "23"}, `{"comment":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})

			t.Run("Shared Via NamespaceUser readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "24"}, `{"comment":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceUser write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "25"}, `{"comment":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})
			t.Run("Shared Via NamespaceUser admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"task": "26"}, `{"comment":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})
		})
	})
}
