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

func TestTask(t *testing.T) {
	testHandler := webHandlerTest{
		user: &testuser1,
		strFunc: func() handler.CObject {
			return &models.Task{}
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
			return &models.Task{}
		},
		t: t,
	}
	// Only run specific nested tests:
	// ^TestTask$/^Update$/^Update_task_items$/^Removing_Assignees_null$
	t.Run("Update", func(t *testing.T) {
		t.Run("Update task items", func(t *testing.T) {
			t.Run("Title", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "1"}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
				assert.NotContains(t, rec.Body.String(), `"title":"task #1"`)
			})
			t.Run("Description", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "1"}, `{"description":"Dolor sit amet"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"description":"Dolor sit amet"`)
				assert.NotContains(t, rec.Body.String(), `"description":"Lorem Ipsum"`)
			})
			t.Run("Description to empty", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "1"}, `{"description":""}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.NotContains(t, rec.Body.String(), `"description":"Lorem Ipsum"`)
			})
			t.Run("Done", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "1"}, `{"done":true}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"done":true`)
				assert.NotContains(t, rec.Body.String(), `"done":false`)
			})
			t.Run("Undone", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "2"}, `{"done":false}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"done":false`)
				assert.NotContains(t, rec.Body.String(), `"done":true`)
			})
			t.Run("Due date", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "1"}, `{"due_date": "2020-02-10T10:00:00Z"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"due_date":"2020-02-10T10:00:00Z"`)
				assert.NotContains(t, rec.Body.String(), `"due_date":0`)
			})
			t.Run("Due date unset", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "5"}, `{"due_date": null}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"due_date":"0001-01-01T00:00:00Z"`)
				assert.NotContains(t, rec.Body.String(), `"due_date":"2020-02-10T10:00:00Z"`)
			})
			t.Run("Reminders", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "1"}, `{"reminders": [{"reminder": "2020-02-10T10:00:00Z"},{"reminder": "2020-02-11T10:00:00Z"}]}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"reminders":[`)
				assert.Contains(t, rec.Body.String(), `{"reminder":"2020-02-10T10:00:00Z"`)
				assert.Contains(t, rec.Body.String(), `{"reminder":"2020-02-11T10:00:00Z"`)
				assert.NotContains(t, rec.Body.String(), `"reminders":null`)
			})
			t.Run("Reminders unset to empty array", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "27"}, `{"reminders": []}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"reminders":null`)
				assert.NotContains(t, rec.Body.String(), `{"Reminder":"2020-02-10T10:00:00Z"`)
			})
			t.Run("Reminders unset to null", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "27"}, `{"reminders": null}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"reminders":null`)
				assert.NotContains(t, rec.Body.String(), `{"Reminder":"2020-02-10T10:00:00Z"`)
			})
			t.Run("Repeat after", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "1"}, `{"repeat_after":3600}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"repeat_after":3600`)
				assert.NotContains(t, rec.Body.String(), `"repeat_after":0`)
			})
			t.Run("Repeat after unset", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "28"}, `{"repeat_after":0}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"repeat_after":0`)
				assert.NotContains(t, rec.Body.String(), `"repeat_after":3600`)
			})
			t.Run("Repeat after update done", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "28"}, `{"done":true}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"done":false`)
				assert.NotContains(t, rec.Body.String(), `"done":true`)
			})
			t.Run("Assignees", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "1"}, `{"assignees":[{"id":1}]}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"assignees":[{"id":1`)
				assert.NotContains(t, rec.Body.String(), `"assignees":[]`)
			})
			t.Run("Removing Assignees empty array", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "30"}, `{"assignees":[]}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"assignees":null`)
				assert.NotContains(t, rec.Body.String(), `"assignees":[{"id":1`)
			})
			t.Run("Removing Assignees null", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "30"}, `{"assignees":null}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"assignees":null`)
				assert.NotContains(t, rec.Body.String(), `"assignees":[{"id":1`)
			})
			t.Run("Priority", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "1"}, `{"priority":100}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"priority":100`)
				assert.NotContains(t, rec.Body.String(), `"priority":0`)
			})
			t.Run("Priority to 0", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "3"}, `{"priority":0}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"priority":0`)
				assert.NotContains(t, rec.Body.String(), `"priority":100`)
			})
			t.Run("Start date", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "1"}, `{"start_date":"2020-02-10T10:00:00Z"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"start_date":"2020-02-10T10:00:00Z"`)
				assert.NotContains(t, rec.Body.String(), `"start_date":0`)
			})
			t.Run("Start date unset", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "7"}, `{"start_date":"0001-01-01T00:00:00Z"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"start_date":"0001-01-01T00:00:00Z"`)
				assert.NotContains(t, rec.Body.String(), `"start_date":"2020-02-10T10:00:00Z"`)
			})
			t.Run("End date", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "1"}, `{"end_date":"2020-02-10T12:00:00Z"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"end_date":"2020-02-10T12:00:00Z"`)
				assert.NotContains(t, rec.Body.String(), `"end_date":""`)
			})
			t.Run("End date unset", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "8"}, `{"end_date":"0001-01-01T00:00:00Z"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"end_date":"0001-01-01T00:00:00Z"`)
				assert.NotContains(t, rec.Body.String(), `"end_date":"2020-02-10T10:00:00Z"`)
			})
			t.Run("Color", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "1"}, `{"hex_color":"f0f0f0"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"hex_color":"f0f0f0"`)
				assert.NotContains(t, rec.Body.String(), `"hex_color":""`)
			})
			t.Run("Color unset", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "31"}, `{"hex_color":""}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"hex_color":""`)
				assert.NotContains(t, rec.Body.String(), `"hex_color":"f0f0f0"`)
			})
			t.Run("Percent Done", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "1"}, `{"percent_done":0.1}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"percent_done":0.1`)
				assert.NotContains(t, rec.Body.String(), `"percent_done":0,`)
			})
			t.Run("Percent Done unset", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "33"}, `{"percent_done":0}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"percent_done":0,`)
				assert.NotContains(t, rec.Body.String(), `"percent_done":0.1`)
			})
		})

		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "99999"}, `{"title":"Lorem Ipsum"}`)
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeTaskDoesNotExist)
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "14"}, `{"title":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "15"}, `{"title":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "16"}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "17"}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "18"}, `{"title":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "19"}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "20"}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "21"}, `{"title":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "22"}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "23"}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "24"}, `{"title":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "25"}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "26"}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
		})
		t.Run("Move to other project", func(t *testing.T) {
			t.Run("normal", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "1"}, `{"project_id":7}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"project_id":7`)
				assert.NotContains(t, rec.Body.String(), `"project_id":1`)
			})
			t.Run("Forbidden", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "1"}, `{"project_id":20}`)
				require.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrorCodeGenericForbidden)
			})
			t.Run("Read Only", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"projecttask": "1"}, `{"project_id":6}`)
				require.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrorCodeGenericForbidden)
			})
		})
	})
	t.Run("Delete", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"projecttask": "1"})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testDeleteWithUser(nil, map[string]string{"projecttask": "99999"})
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeTaskDoesNotExist)
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"projecttask": "14"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"projecttask": "15"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"projecttask": "16"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"projecttask": "17"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"projecttask": "18"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"projecttask": "19"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"projecttask": "20"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"projecttask": "21"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"projecttask": "22"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"projecttask": "23"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"projecttask": "24"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"projecttask": "25"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"projecttask": "26"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
		})
	})
	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "1"}, `{"title":"Lorem Ipsum"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "9999"}, `{"title":"Lorem Ipsum"}`)
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeProjectDoesNotExist)
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "20"}, `{"title":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "6"}, `{"title":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "7"}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "8"}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "9"}, `{"title":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "10"}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "11"}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "12"}, `{"title":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "13"}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "14"}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "15"}, `{"title":"Lorem Ipsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "16"}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"project": "17"}, `{"title":"Lorem Ipsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
		})
		t.Run("Link Share", func(t *testing.T) {
			rec, err := testHandlerLinkShareWrite.testCreateWithLinkShare(nil, map[string]string{"project": "2"}, `{"title":"Lorem Ipsum"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			db.AssertExists(t, "tasks", map[string]interface{}{
				"project_id":    2,
				"title":         "Lorem Ipsum",
				"created_by_id": -2,
			}, false)
		})
	})
}
