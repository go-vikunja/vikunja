// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package integrations

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/web/handler"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTask(t *testing.T) {
	testHandler := webHandlerTest{
		user: &testuser1,
		strFunc: func() handler.CObject {
			return &models.Task{}
		},
		t: t,
	}
	// Only run specific nested tests:
	// ^TestTask$/^Update$/^Update_task_items$/^Removing_Assignees_null$
	t.Run("Update", func(t *testing.T) {
		t.Run("Update task items", func(t *testing.T) {
			t.Run("Text", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "1"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
				assert.NotContains(t, rec.Body.String(), `"text":"task #1"`)
			})
			t.Run("Description", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "1"}, `{"description":"Dolor sit amet"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"description":"Dolor sit amet"`)
				assert.NotContains(t, rec.Body.String(), `"description":"Lorem Ipsum"`)
			})
			t.Run("Description to empty", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "1"}, `{"description":""}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.NotContains(t, rec.Body.String(), `"description":"Lorem Ipsum"`)
			})
			t.Run("Done", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "1"}, `{"done":true}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"done":true`)
				assert.NotContains(t, rec.Body.String(), `"done":false`)
			})
			t.Run("Undone", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "2"}, `{"done":false}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"done":false`)
				assert.NotContains(t, rec.Body.String(), `"done":true`)
			})
			t.Run("Due date", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "1"}, `{"due_date": "2020-02-10T10:00:00Z"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"due_date":"2020-02-10T10:00:00Z"`)
				assert.NotContains(t, rec.Body.String(), `"due_date":0`)
			})
			t.Run("Due date unset", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "5"}, `{"due_date": null}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"due_date":null`)
				assert.NotContains(t, rec.Body.String(), `"due_date":"2020-02-10T10:00:00Z"`)
			})
			t.Run("Reminders", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "1"}, `{"reminder_dates": ["2020-02-10T10:00:00Z","2020-02-11T10:00:00Z"]}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"reminder_dates":["2020-02-10T10:00:00Z","2020-02-11T10:00:00Z"]`)
				assert.NotContains(t, rec.Body.String(), `"reminder_dates": null`)
			})
			t.Run("Reminders unset to empty array", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "27"}, `{"reminder_dates": []}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"reminder_dates":null`)
				assert.NotContains(t, rec.Body.String(), `"reminder_dates":[1543626724,1543626824]`)
			})
			t.Run("Reminders unset to null", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "27"}, `{"reminder_dates": null}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"reminder_dates":null`)
				assert.NotContains(t, rec.Body.String(), `"reminder_dates":[1543626724,1543626824]`)
			})
			t.Run("Repeat after", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "1"}, `{"repeat_after":3600}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"repeat_after":3600`)
				assert.NotContains(t, rec.Body.String(), `"repeat_after":0`)
			})
			t.Run("Repeat after unset", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "28"}, `{"repeat_after":0}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"repeat_after":0`)
				assert.NotContains(t, rec.Body.String(), `"repeat_after":3600`)
			})
			t.Run("Repeat after update done", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "28"}, `{"done":true}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"done":false`)
				assert.NotContains(t, rec.Body.String(), `"done":true`)
			})
			t.Run("Assignees", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "1"}, `{"assignees":[{"id":1}]}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"assignees":[{"id":1`)
				assert.NotContains(t, rec.Body.String(), `"assignees":[]`)
			})
			t.Run("Removing Assignees empty array", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "30"}, `{"assignees":[]}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"assignees":null`)
				assert.NotContains(t, rec.Body.String(), `"assignees":[{"id":1`)
			})
			t.Run("Removing Assignees null", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "30"}, `{"assignees":null}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"assignees":null`)
				assert.NotContains(t, rec.Body.String(), `"assignees":[{"id":1`)
			})
			t.Run("Priority", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "1"}, `{"priority":100}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"priority":100`)
				assert.NotContains(t, rec.Body.String(), `"priority":0`)
			})
			t.Run("Priority to 0", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "3"}, `{"priority":0}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"priority":0`)
				assert.NotContains(t, rec.Body.String(), `"priority":100`)
			})
			t.Run("Start date", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "1"}, `{"start_date":"2020-02-10T10:00:00Z"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"start_date":"2020-02-10T10:00:00Z"`)
				assert.NotContains(t, rec.Body.String(), `"start_date":0`)
			})
			t.Run("Start date unset", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "7"}, `{"start_date":null}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"start_date":null`)
				assert.NotContains(t, rec.Body.String(), `"start_date":"2020-02-10T10:00:00Z"`)
			})
			t.Run("End date", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "1"}, `{"end_date":"2020-02-10T12:00:00Z"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"end_date":"2020-02-10T12:00:00Z"`)
				assert.NotContains(t, rec.Body.String(), `"end_date":""`)
			})
			t.Run("End date unset", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "8"}, `{"end_date":null}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"end_date":null`)
				assert.NotContains(t, rec.Body.String(), `"end_date":"2020-02-10T10:00:00Z"`)
			})
			t.Run("Color", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "1"}, `{"hex_color":"f0f0f0"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"hex_color":"f0f0f0"`)
				assert.NotContains(t, rec.Body.String(), `"hex_color":""`)
			})
			t.Run("Color unset", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "31"}, `{"hex_color":""}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"hex_color":""`)
				assert.NotContains(t, rec.Body.String(), `"hex_color":"f0f0f0"`)
			})
			t.Run("Percent Done", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "1"}, `{"percent_done":0.1}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"percent_done":0.1`)
				assert.NotContains(t, rec.Body.String(), `"percent_done":0,`)
			})
			t.Run("Percent Done unset", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "33"}, `{"percent_done":0}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"percent_done":0,`)
				assert.NotContains(t, rec.Body.String(), `"percent_done":0.1`)
			})
		})

		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "99999"}, `{"text":"Lorem Ipsum"}`)
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeTaskDoesNotExist)
		})
		t.Run("Rights check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "14"}, `{"text":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "15"}, `{"text":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "16"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "17"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "18"}, `{"text":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "19"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "20"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})

			t.Run("Shared Via NamespaceTeam readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "21"}, `{"text":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceTeam write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "22"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})
			t.Run("Shared Via NamespaceTeam admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "23"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})

			t.Run("Shared Via NamespaceUser readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "24"}, `{"text":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceUser write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "25"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})
			t.Run("Shared Via NamespaceUser admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"listtask": "26"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})
		})
	})
	t.Run("Delete", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"listtask": "1"})
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testDeleteWithUser(nil, map[string]string{"listtask": "99999"})
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeTaskDoesNotExist)
		})
		t.Run("Rights check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"listtask": "14"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"listtask": "15"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"listtask": "16"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"listtask": "17"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"listtask": "18"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"listtask": "19"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"listtask": "20"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})

			t.Run("Shared Via NamespaceTeam readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"listtask": "21"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceTeam write", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"listtask": "22"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared Via NamespaceTeam admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"listtask": "23"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})

			t.Run("Shared Via NamespaceUser readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"listtask": "24"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceUser write", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"listtask": "25"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared Via NamespaceUser admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"listtask": "26"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
		})
	})
	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "1"}, `{"text":"Lorem Ipsum"}`)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "9999"}, `{"text":"Lorem Ipsum"}`)
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeListDoesNotExist)
		})
		t.Run("Rights check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "20"}, `{"text":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "6"}, `{"text":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "7"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "8"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "9"}, `{"text":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "10"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "11"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})

			t.Run("Shared Via NamespaceTeam readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "12"}, `{"text":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceTeam write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "13"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})
			t.Run("Shared Via NamespaceTeam admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "14"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})

			t.Run("Shared Via NamespaceUser readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "15"}, `{"text":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceUser write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "16"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})
			t.Run("Shared Via NamespaceUser admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"list": "17"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})
		})
	})
}
