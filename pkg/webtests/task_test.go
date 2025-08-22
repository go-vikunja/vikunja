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
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTask(t *testing.T) {
	th := NewTestHelper(t)
	th.Login(t, &testuser1)

	// Only run specific nested tests:
	// ^TestTask$/^Update$/^Update_task_items$/^Removing_Assignees_null$
	t.Run("Update", func(t *testing.T) {
		t.Run("Update task items", func(t *testing.T) {
			t.Run("Title", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/1", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
				assert.NotContains(t, rec.Body.String(), `"title":"task #1"`)
			})
			t.Run("Description", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/1", strings.NewReader(`{"description":"Dolor sit amet"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"description":"Dolor sit amet"`)
				assert.NotContains(t, rec.Body.String(), `"description":"Lorem Ipsum"`)
			})
			t.Run("Description to empty", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/1", strings.NewReader(`{"description":""}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.NotContains(t, rec.Body.String(), `"description":"Lorem Ipsum"`)
			})
			t.Run("Done", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/1", strings.NewReader(`{"done":true}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"done":true`)
				assert.NotContains(t, rec.Body.String(), `"done":false`)
			})
			t.Run("Undone", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/2", strings.NewReader(`{"done":false}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"done":false`)
				assert.NotContains(t, rec.Body.String(), `"done":true`)
			})
			t.Run("Due date", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/1", strings.NewReader(`{"due_date": "2020-02-10T10:00:00Z"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"due_date":"2020-02-10T10:00:00Z"`)
				assert.NotContains(t, rec.Body.String(), `"due_date":0`)
			})
			t.Run("Due date unset", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/5", strings.NewReader(`{"due_date": null}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"due_date":"0001-01-01T00:00:00Z"`)
				assert.NotContains(t, rec.Body.String(), `"due_date":"2020-02-10T10:00:00Z"`)
			})
			t.Run("Reminders", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/1", strings.NewReader(`{"reminders": [{"reminder": "2020-02-10T10:00:00Z"},{"reminder": "2020-02-11T10:00:00Z"}]}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"reminders":[`)
				assert.Contains(t, rec.Body.String(), `{"reminder":"2020-02-10T10:00:00Z"`)
				assert.Contains(t, rec.Body.String(), `{"reminder":"2020-02-11T10:00:00Z"`)
				assert.NotContains(t, rec.Body.String(), `"reminders":null`)
			})
			t.Run("Reminders unset to empty array", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/27", strings.NewReader(`{"reminders": []}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"reminders":null`)
				assert.NotContains(t, rec.Body.String(), `{"Reminder":"2020-02-10T10:00:00Z"`)
			})
			t.Run("Reminders unset to null", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/27", strings.NewReader(`{"reminders": null}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"reminders":null`)
				assert.NotContains(t, rec.Body.String(), `{"Reminder":"2020-02-10T10:00:00Z"`)
			})
			t.Run("Repeat after", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/1", strings.NewReader(`{"repeat_after":3600}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"repeat_after":3600`)
				assert.NotContains(t, rec.Body.String(), `"repeat_after":0`)
			})
			t.Run("Repeat after unset", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/28", strings.NewReader(`{"repeat_after":0}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"repeat_after":0`)
				assert.NotContains(t, rec.Body.String(), `"repeat_after":3600`)
			})
			t.Run("Repeat after update done", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/28", strings.NewReader(`{"done":true}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"done":false`)
				assert.NotContains(t, rec.Body.String(), `"done":true`)
			})
			t.Run("Assignees", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/1", strings.NewReader(`{"assignees":[{"id":1}]}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"assignees":[{"id":1`)
				assert.NotContains(t, rec.Body.String(), `"assignees":[]`)
			})
			t.Run("Removing Assignees empty array", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/30", strings.NewReader(`{"assignees":[]}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"assignees":null`)
				assert.NotContains(t, rec.Body.String(), `"assignees":[{"id":1`)
			})
			t.Run("Removing Assignees null", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/30", strings.NewReader(`{"assignees":null}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"assignees":null`)
				assert.NotContains(t, rec.Body.String(), `"assignees":[{"id":1`)
			})
			t.Run("Priority", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/1", strings.NewReader(`{"priority":100}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"priority":100`)
				assert.NotContains(t, rec.Body.String(), `"priority":0`)
			})
			t.Run("Priority to 0", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/3", strings.NewReader(`{"priority":0}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"priority":0`)
				assert.NotContains(t, rec.Body.String(), `"priority":100`)
			})
			t.Run("Start date", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/1", strings.NewReader(`{"start_date":"2020-02-10T10:00:00Z"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"start_date":"2020-02-10T10:00:00Z"`)
				assert.NotContains(t, rec.Body.String(), `"start_date":0`)
			})
			t.Run("Start date unset", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/7", strings.NewReader(`{"start_date":"0001-01-01T00:00:00Z"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"start_date":"0001-01-01T00:00:00Z"`)
				assert.NotContains(t, rec.Body.String(), `"start_date":"2020-02-10T10:00:00Z"`)
			})
			t.Run("End date", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/1", strings.NewReader(`{"end_date":"2020-02-10T12:00:00Z"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"end_date":"2020-02-10T12:00:00Z"`)
				assert.NotContains(t, rec.Body.String(), `"end_date":""`)
			})
			t.Run("End date unset", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/8", strings.NewReader(`{"end_date":"0001-01-01T00:00:00Z"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"end_date":"0001-01-01T00:00:00Z"`)
				assert.NotContains(t, rec.Body.String(), `"end_date":"2020-02-10T10:00:00Z"`)
			})
			t.Run("Color", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/1", strings.NewReader(`{"hex_color":"f0f0f0"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"hex_color":"f0f0f0"`)
				assert.NotContains(t, rec.Body.String(), `"hex_color":""`)
			})
			t.Run("Color unset", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/31", strings.NewReader(`{"hex_color":""}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"hex_color":""`)
				assert.NotContains(t, rec.Body.String(), `"hex_color":"f0f0f0"`)
			})
			t.Run("Percent Done", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/1", strings.NewReader(`{"percent_done":0.1}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"percent_done":0.1`)
				assert.NotContains(t, rec.Body.String(), `"percent_done":0,`)
			})
			t.Run("Percent Done unset", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/33", strings.NewReader(`{"percent_done":0}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"percent_done":0,`)
				assert.NotContains(t, rec.Body.String(), `"percent_done":0.1`)
			})
		})

		t.Run("Nonexisting", func(t *testing.T) {
			_, err := th.Request(t, "POST", "/api/v1/tasks/99999", strings.NewReader(`{"title":"Lorem Ipsum"}`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), `"code":404`)
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/tasks/14", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/tasks/15", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/16", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/17", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/tasks/18", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/19", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/20", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/tasks/21", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/22", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/23", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/tasks/24", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/25", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/26", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
		})
		t.Run("Move to other project", func(t *testing.T) {
			t.Run("normal", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/tasks/1", strings.NewReader(`{"project_id":7}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"project_id":7`)
				assert.NotContains(t, rec.Body.String(), `"project_id":1`)
			})
			t.Run("Forbidden", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/tasks/1", strings.NewReader(`{"project_id":20}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Read Only", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/tasks/1", strings.NewReader(`{"project_id":6}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
		})
	})
	t.Run("Delete", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := th.Request(t, "DELETE", "/api/v1/tasks/1", nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := th.Request(t, "DELETE", "/api/v1/tasks/99999", nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), `"code":404`)
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/tasks/14", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/tasks/15", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := th.Request(t, "DELETE", "/api/v1/tasks/16", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := th.Request(t, "DELETE", "/api/v1/tasks/17", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/tasks/18", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := th.Request(t, "DELETE", "/api/v1/tasks/19", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := th.Request(t, "DELETE", "/api/v1/tasks/20", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/tasks/21", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				rec, err := th.Request(t, "DELETE", "/api/v1/tasks/22", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := th.Request(t, "DELETE", "/api/v1/tasks/23", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/tasks/24", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				rec, err := th.Request(t, "DELETE", "/api/v1/tasks/25", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := th.Request(t, "DELETE", "/api/v1/tasks/26", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
		})
	})
	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := th.Request(t, "PUT", "/api/v1/projects/1/tasks", strings.NewReader(`{"title":"Lorem Ipsum"}`))
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := th.Request(t, "PUT", "/api/v1/projects/9999/tasks", strings.NewReader(`{"title":"Lorem Ipsum"}`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), `"code":404`)
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := th.Request(t, "PUT", "/api/v1/projects/20/tasks", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "PUT", "/api/v1/projects/6/tasks", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/projects/7/tasks", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/projects/8/tasks", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := th.Request(t, "PUT", "/api/v1/projects/9/tasks", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/projects/10/tasks", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/projects/11/tasks", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "PUT", "/api/v1/projects/12/tasks", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/projects/13/tasks", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/projects/14/tasks", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := th.Request(t, "PUT", "/api/v1/projects/15/tasks", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/projects/16/tasks", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/projects/17/tasks", strings.NewReader(`{"title":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem Ipsum"`)
			})
		})
		t.Run("Link Share", func(t *testing.T) {
			th.Logout(t)
			th.SetLinkShare(t, &models.LinkSharing{
				ID:          2,
				Hash:        "test2",
				ProjectID:   2,
				Permission:  models.PermissionWrite,
				SharingType: models.SharingTypeWithoutPassword,
				SharedByID:  1,
			})
			rec, err := th.Request(t, "PUT", "/api/v1/projects/2/tasks", strings.NewReader(`{"title":"Lorem Ipsum"}`))
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
