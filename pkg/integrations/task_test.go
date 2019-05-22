//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2019 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package integrations

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/web/handler"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestListTask(t *testing.T) {
	testHandler := webHandlerTest{
		user: &testuser1,
		strFunc: func() handler.CObject {
			return &models.ListTask{}
		},
		t: t,
	}
	// Only run specific nested tests:
	// ^TestListTask$/^Update$/^Update_task_items$/^Removing_Assignees_null$
	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testReadAll(nil, nil)
			assert.NoError(t, err)
			// Not using assert.Equal to avoid having the tests break every time we add new fixtures
			assert.Contains(t, rec.Body.String(), `task #1`)
			assert.Contains(t, rec.Body.String(), `task #2`)
			assert.Contains(t, rec.Body.String(), `task #3`)
			assert.Contains(t, rec.Body.String(), `task #4`)
			assert.Contains(t, rec.Body.String(), `task #5`)
			assert.Contains(t, rec.Body.String(), `task #6`)
			assert.Contains(t, rec.Body.String(), `task #7`)
			assert.Contains(t, rec.Body.String(), `task #8`)
			assert.Contains(t, rec.Body.String(), `task #9`)
			assert.Contains(t, rec.Body.String(), `task #10`)
			assert.Contains(t, rec.Body.String(), `task #11`)
			assert.Contains(t, rec.Body.String(), `task #12`)
			assert.NotContains(t, rec.Body.String(), `task #13`)
			assert.NotContains(t, rec.Body.String(), `task #14`)
			// TODO: add more tasks, since the whole point of this is to get all tasks in all lists where the user
			//  has at least read access
		})
		t.Run("Search", func(t *testing.T) {
			rec, err := testHandler.testReadAll(url.Values{"s": []string{"task #6"}}, nil)
			assert.NoError(t, err)
			assert.NotContains(t, rec.Body.String(), `task #1`)
			assert.NotContains(t, rec.Body.String(), `task #2`)
			assert.NotContains(t, rec.Body.String(), `task #3`)
			assert.NotContains(t, rec.Body.String(), `task #4`)
			assert.NotContains(t, rec.Body.String(), `task #5`)
			assert.Contains(t, rec.Body.String(), `task #6`)
			assert.NotContains(t, rec.Body.String(), `task #7`)
			assert.NotContains(t, rec.Body.String(), `task #8`)
			assert.NotContains(t, rec.Body.String(), `task #9`)
			assert.NotContains(t, rec.Body.String(), `task #10`)
			assert.NotContains(t, rec.Body.String(), `task #11`)
			assert.NotContains(t, rec.Body.String(), `task #12`)
			assert.NotContains(t, rec.Body.String(), `task #13`)
			assert.NotContains(t, rec.Body.String(), `task #14`)
		})
		t.Run("Sort Order", func(t *testing.T) {
			// should equal priority desc
			t.Run("by priority", func(t *testing.T) {
				rec, err := testHandler.testReadAll(url.Values{"sort": []string{"priority"}}, nil)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `[{"id":3,"text":"task #3 high prio","description":"","done":false,"doneAt":0,"dueDate":0,"reminderDates":null,"listID":1,"repeatAfter":0,"parentTaskID":0,"priority":100,"startDate":0,"endDate":0,"assignees":null,"labels":null,"hexColor":"","subtasks":null,"created":1543626724,"updated":1543626724,"createdBy":{"id":0,"username":"","created":0,"updated":0}},{"id":4,"text":"task #4 low prio","description":"","done":false,"doneAt":0,"dueDate":0,"reminderDates":null,"listID":1,"repeatAfter":0,"parentTaskID":0,"priority":1`)
			})
			t.Run("by priority desc", func(t *testing.T) {
				rec, err := testHandler.testReadAll(url.Values{"sort": []string{"prioritydesc"}}, nil)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `[{"id":3,"text":"task #3 high prio","description":"","done":false,"doneAt":0,"dueDate":0,"reminderDates":null,"listID":1,"repeatAfter":0,"parentTaskID":0,"priority":100,"startDate":0,"endDate":0,"assignees":null,"labels":null,"hexColor":"","subtasks":null,"created":1543626724,"updated":1543626724,"createdBy":{"id":0,"username":"","created":0,"updated":0}},{"id":4,"text":"task #4 low prio","description":"","done":false,"doneAt":0,"dueDate":0,"reminderDates":null,"listID":1,"repeatAfter":0,"parentTaskID":0,"priority":1`)
			})
			t.Run("by priority asc", func(t *testing.T) {
				rec, err := testHandler.testReadAll(url.Values{"sort": []string{"priorityasc"}}, nil)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `{"id":4,"text":"task #4 low prio","description":"","done":false,"doneAt":0,"dueDate":0,"reminderDates":null,"listID":1,"repeatAfter":0,"parentTaskID":0,"priority":1,"startDate":0,"endDate":0,"assignees":null,"labels":null,"hexColor":"","subtasks":null,"created":1543626724,"updated":1543626724,"createdBy":{"id":0,"username":"","created":0,"updated":0}},{"id":3,"text":"task #3 high prio","description":"","done":false,"doneAt":0,"dueDate":0,"reminderDates":null,"listID":1,"repeatAfter":0,"parentTaskID":0,"priority":100,"startDate":0,"endDate":0,"assignees":null,"labels":null,"hexColor":"","subtasks":null,"created":1543626724,"updated":1543626724,"createdBy":{"id":0,"username":"","created":0,"updated":0}}]`)
			})
			// should equal duedate desc
			t.Run("by duedate", func(t *testing.T) {
				rec, err := testHandler.testReadAll(url.Values{"sort": []string{"dueadate"}}, nil)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `[{"id":5,"text":"task #5 higher due date","description":"","done":false,"doneAt":0,"dueDate":1543636724,"reminderDates":null,"listID":1,"repeatAfter":0,"parentTaskID":0,"priority":0,"startDate":0,"endDate":0,"assignees":null,"labels":null,"hexColor":"","subtasks":null,"created":1543626724,"updated":1543626724,"createdBy":{"id":0,"username":"","created":0,"updated":0}},{"id":6,"text":"task #6 lower due date"`)
			})
			t.Run("by duedate desc", func(t *testing.T) {
				rec, err := testHandler.testReadAll(url.Values{"sort": []string{"dueadatedesc"}}, nil)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `[{"id":5,"text":"task #5 higher due date","description":"","done":false,"doneAt":0,"dueDate":1543636724,"reminderDates":null,"listID":1,"repeatAfter":0,"parentTaskID":0,"priority":0,"startDate":0,"endDate":0,"assignees":null,"labels":null,"hexColor":"","subtasks":null,"created":1543626724,"updated":1543626724,"createdBy":{"id":0,"username":"","created":0,"updated":0}},{"id":6,"text":"task #6 lower due date"`)
			})
			t.Run("by duedate asc", func(t *testing.T) {
				rec, err := testHandler.testReadAll(url.Values{"sort": []string{"duedateasc"}}, nil)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `{"id":6,"text":"task #6 lower due date","description":"","done":false,"doneAt":0,"dueDate":1543616724,"reminderDates":null,"listID":1,"repeatAfter":0,"parentTaskID":0,"priority":0,"startDate":0,"endDate":0,"assignees":null,"labels":null,"hexColor":"","subtasks":null,"created":1543626724,"updated":1543626724,"createdBy":{"id":0,"username":"","created":0,"updated":0}},{"id":5,"text":"task #5 higher due date","description":"","done":false,"doneAt":0,"dueDate":1543636724,"reminderDates":null,"listID":1,"repeatAfter":0,"parentTaskID":0,"priority":0,"startDate":0,"endDate":0,"assignees":null,"labels":null,"hexColor":"","subtasks":null,"created":1543626724,"updated":1543626724,"createdBy":{"id":0,"username":"","created":0,"updated":0}}]`)
			})
			t.Run("invalid parameter", func(t *testing.T) {
				// Invalid parameter should not sort at all
				rec, err := testHandler.testReadAll(url.Values{"sort": []string{"loremipsum"}}, nil)
				assert.NoError(t, err)
				assert.NotContains(t, rec.Body.String(), `[{"id":3,"text":"task #3 high prio","description":"","done":false,"dueDate":0,"reminderDates":null,"repeatAfter":0,"parentTaskID":0,"priority":100,"startDate":0,"endDate":0,"assignees":null,"labels":null,"hexColor":"","subtasks":null,"created":1543626724,"updated":1543626724,"createdBy":{"id":0,"username":"","email":"","created":0,"updated":0}},{"id":4,"text":"task #4 low prio","description":"","done":false,"dueDate":0,"reminderDates":null,"repeatAfter":0,"parentTaskID":0,"priority":1`)
				assert.NotContains(t, rec.Body.String(), `{"id":4,"text":"task #4 low prio","description":"","done":false,"dueDate":0,"reminderDates":null,"repeatAfter":0,"parentTaskID":0,"priority":1,"startDate":0,"endDate":0,"assignees":null,"labels":null,"hexColor":"","subtasks":null,"created":1543626724,"updated":1543626724,"createdBy":{"id":0,"username":"","email":"","created":0,"updated":0}},{"id":3,"text":"task #3 high prio","description":"","done":false,"dueDate":0,"reminderDates":null,"repeatAfter":0,"parentTaskID":0,"priority":100,"startDate":0,"endDate":0,"assignees":null,"labels":null,"subtasks":null,"created":1543626724,"updated":1543626724,"createdBy":{"id":0,"username":"","email":"","created":0,"updated":0}}]`)
				assert.NotContains(t, rec.Body.String(), `[{"id":5,"text":"task #5 higher due date","description":"","done":false,"dueDate":1543636724,"reminderDates":null,"repeatAfter":0,"parentTaskID":0,"priority":0,"startDate":0,"endDate":0,"assignees":null,"labels":null,"hexColor":"","subtasks":null,"created":1543626724,"updated":1543626724,"createdBy":{"id":0,"username":"","email":"","created":0,"updated":0}},{"id":6,"text":"task #6 lower due date"`)
				assert.NotContains(t, rec.Body.String(), `{"id":6,"text":"task #6 lower due date","description":"","done":false,"dueDate":1543616724,"reminderDates":null,"repeatAfter":0,"parentTaskID":0,"priority":0,"startDate":0,"endDate":0,"assignees":null,"labels":null,"hexColor":"","subtasks":null,"created":1543626724,"updated":1543626724,"createdBy":{"id":0,"username":"","email":"","created":0,"updated":0}},{"id":5,"text":"task #5 higher due date","description":"","done":false,"dueDate":1543636724,"reminderDates":null,"repeatAfter":0,"parentTaskID":0,"priority":0,"startDate":0,"endDate":0,"assignees":null,"labels":null,"subtasks":null,"created":1543626724,"updated":1543626724,"createdBy":{"id":0,"username":"","email":"","created":0,"updated":0}}]`)
			})
		})
		t.Run("Date range", func(t *testing.T) {
			t.Run("start and end date", func(t *testing.T) {
				rec, err := testHandler.testReadAll(url.Values{"startdate": []string{"1540000000"}, "enddate": []string{"1544700001"}}, nil)
				assert.NoError(t, err)
				assert.NotContains(t, rec.Body.String(), `task #1`)
				assert.NotContains(t, rec.Body.String(), `task #2`)
				assert.NotContains(t, rec.Body.String(), `task #3`)
				assert.NotContains(t, rec.Body.String(), `task #4`)
				assert.Contains(t, rec.Body.String(), `task #5`)
				assert.Contains(t, rec.Body.String(), `task #6`)
				assert.Contains(t, rec.Body.String(), `task #7`)
				assert.Contains(t, rec.Body.String(), `task #8`)
				assert.Contains(t, rec.Body.String(), `task #9`)
				assert.NotContains(t, rec.Body.String(), `task #10`)
				assert.NotContains(t, rec.Body.String(), `task #11`)
				assert.NotContains(t, rec.Body.String(), `task #12`)
				assert.NotContains(t, rec.Body.String(), `task #13`)
				assert.NotContains(t, rec.Body.String(), `task #14`)
			})
			t.Run("start date only", func(t *testing.T) {
				rec, err := testHandler.testReadAll(url.Values{"startdate": []string{"1540000000"}}, nil)
				assert.NoError(t, err)
				assert.NotContains(t, rec.Body.String(), `task #1`)
				assert.NotContains(t, rec.Body.String(), `task #2`)
				assert.NotContains(t, rec.Body.String(), `task #3`)
				assert.NotContains(t, rec.Body.String(), `task #4`)
				assert.Contains(t, rec.Body.String(), `task #5`)
				assert.Contains(t, rec.Body.String(), `task #6`)
				assert.Contains(t, rec.Body.String(), `task #7`)
				assert.Contains(t, rec.Body.String(), `task #8`)
				assert.Contains(t, rec.Body.String(), `task #9`)
				assert.NotContains(t, rec.Body.String(), `task #10`)
				assert.NotContains(t, rec.Body.String(), `task #11`)
				assert.NotContains(t, rec.Body.String(), `task #12`)
				assert.NotContains(t, rec.Body.String(), `task #13`)
				assert.NotContains(t, rec.Body.String(), `task #14`)
			})
			t.Run("end date only", func(t *testing.T) {
				rec, err := testHandler.testReadAll(url.Values{"enddate": []string{"1544700001"}}, nil)
				assert.NoError(t, err)
				// If no start date but an end date is specified, this should be null
				// since we don't have any tasks in the fixtures with an end date >
				// the current date.
				assert.Equal(t, "null\n", rec.Body.String())
			})
		})
	})
	t.Run("Update", func(t *testing.T) {
		t.Run("Update task items", func(t *testing.T) {
			t.Run("Text", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "1"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
				assert.NotContains(t, rec.Body.String(), `"text":"task #1"`)
			})
			t.Run("Description", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "1"}, `{"description":"Dolor sit amet"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"description":"Dolor sit amet"`)
				assert.NotContains(t, rec.Body.String(), `"description":"Lorem Ipsum"`)
			})
			t.Run("Description to empty", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "1"}, `{"description":""}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.NotContains(t, rec.Body.String(), `"description":"Lorem Ipsum"`)
			})
			t.Run("Done", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "1"}, `{"done":true}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"done":true`)
				assert.NotContains(t, rec.Body.String(), `"done":false`)
			})
			t.Run("Undone", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "2"}, `{"done":false}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"done":false`)
				assert.NotContains(t, rec.Body.String(), `"done":true`)
			})
			t.Run("Due date", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "1"}, `{"dueDate": 123456}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"dueDate":123456`)
				assert.NotContains(t, rec.Body.String(), `"dueDate":0`)
			})
			t.Run("Due date unset", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "5"}, `{"dueDate": 0}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"dueDate":0`)
				assert.NotContains(t, rec.Body.String(), `"dueDate":1543636724`)
			})
			t.Run("Reminders", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "1"}, `{"reminderDates": [1555508227,1555511000]}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"reminderDates":[1555508227,1555511000]`)
				assert.NotContains(t, rec.Body.String(), `"reminderDates": null`)
			})
			t.Run("Reminders unset to empty array", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "27"}, `{"reminderDates": []}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"reminderDates":null`)
				assert.NotContains(t, rec.Body.String(), `"reminderDates":[1543626724,1543626824]`)
			})
			t.Run("Reminders unset to null", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "27"}, `{"reminderDates": null}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"reminderDates":null`)
				assert.NotContains(t, rec.Body.String(), `"reminderDates":[1543626724,1543626824]`)
			})
			t.Run("Repeat after", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "1"}, `{"repeatAfter":3600}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"repeatAfter":3600`)
				assert.NotContains(t, rec.Body.String(), `"repeatAfter":0`)
			})
			t.Run("Repeat after unset", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "28"}, `{"repeatAfter":0}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"repeatAfter":0`)
				assert.NotContains(t, rec.Body.String(), `"repeatAfter":3600`)
			})
			t.Run("Repeat after update done", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "28"}, `{"done":true}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"done":false`)
				assert.NotContains(t, rec.Body.String(), `"done":true`)
			})
			t.Run("Parent task", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "1"}, `{"parentTaskID":2}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"parentTaskID":2`)
				assert.NotContains(t, rec.Body.String(), `"parentTaskID":0`)
			})
			t.Run("Parent task same task", func(t *testing.T) {
				_, err := testHandler.testUpdate(nil, map[string]string{"listtask": "1"}, `{"parentTaskID":1}`)
				assert.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrCodeParentTaskCannotBeTheSame)
			})
			t.Run("Parent task unset", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "29"}, `{"parentTaskID":0}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"parentTaskID":0`)
				assert.NotContains(t, rec.Body.String(), `"parentTaskID":1`)
			})
			t.Run("Assignees", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "1"}, `{"assignees":[{"id":1}]}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"assignees":[{"id":1`)
				assert.NotContains(t, rec.Body.String(), `"assignees":[]`)
			})
			t.Run("Removing Assignees empty array", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "30"}, `{"assignees":[]}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"assignees":null`)
				assert.NotContains(t, rec.Body.String(), `"assignees":[{"id":1`)
			})
			t.Run("Removing Assignees null", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "30"}, `{"assignees":null}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"assignees":null`)
				assert.NotContains(t, rec.Body.String(), `"assignees":[{"id":1`)
			})
			t.Run("Priority", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "1"}, `{"priority":100}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"priority":100`)
				assert.NotContains(t, rec.Body.String(), `"priority":0`)
			})
			t.Run("Priority to 0", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "3"}, `{"priority":0}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"priority":0`)
				assert.NotContains(t, rec.Body.String(), `"priority":100`)
			})
			t.Run("Start date", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "1"}, `{"startDate":1234567}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"startDate":1234567`)
				assert.NotContains(t, rec.Body.String(), `"startDate":0`)
			})
			t.Run("Start date unset", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "7"}, `{"startDate":0}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"startDate":0`)
				assert.NotContains(t, rec.Body.String(), `"startDate":1544600000`)
			})
			t.Run("End date", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "1"}, `{"endDate":123456}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"endDate":123456`)
				assert.NotContains(t, rec.Body.String(), `"endDate":0`)
			})
			t.Run("End date unset", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "8"}, `{"endDate":0}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"endDate":0`)
				assert.NotContains(t, rec.Body.String(), `"endDate":1544700000`)
			})
			t.Run("Color", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "1"}, `{"hexColor":"f0f0f0"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"hexColor":"f0f0f0"`)
				assert.NotContains(t, rec.Body.String(), `"hexColor":""`)
			})
			t.Run("Color unset", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "31"}, `{"hexColor":""}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"hexColor":""`)
				assert.NotContains(t, rec.Body.String(), `"hexColor":"f0f0f0"`)
			})
		})

		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testUpdate(nil, map[string]string{"listtask": "99999"}, `{"text":"Lorem Ipsum"}`)
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeListTaskDoesNotExist)
		})
		t.Run("Rights check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				_, err := testHandler.testUpdate(nil, map[string]string{"listtask": "14"}, `{"text":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testUpdate(nil, map[string]string{"listtask": "15"}, `{"text":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "16"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "17"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testUpdate(nil, map[string]string{"listtask": "18"}, `{"text":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "19"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "20"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})

			t.Run("Shared Via NamespaceTeam readonly", func(t *testing.T) {
				_, err := testHandler.testUpdate(nil, map[string]string{"listtask": "21"}, `{"text":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceTeam write", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "22"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})
			t.Run("Shared Via NamespaceTeam admin", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "23"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})

			t.Run("Shared Via NamespaceUser readonly", func(t *testing.T) {
				_, err := testHandler.testUpdate(nil, map[string]string{"listtask": "24"}, `{"text":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceUser write", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "25"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})
			t.Run("Shared Via NamespaceUser admin", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"listtask": "26"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})
		})
	})
	t.Run("Delete", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testDelete(nil, map[string]string{"listtask": "1"})
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testDelete(nil, map[string]string{"listtask": "99999"})
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeListTaskDoesNotExist)
		})
		t.Run("Rights check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"listtask": "14"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"listtask": "15"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testDelete(nil, map[string]string{"listtask": "16"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testDelete(nil, map[string]string{"listtask": "17"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"listtask": "18"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testDelete(nil, map[string]string{"listtask": "19"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testDelete(nil, map[string]string{"listtask": "20"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})

			t.Run("Shared Via NamespaceTeam readonly", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"listtask": "21"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceTeam write", func(t *testing.T) {
				rec, err := testHandler.testDelete(nil, map[string]string{"listtask": "22"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared Via NamespaceTeam admin", func(t *testing.T) {
				rec, err := testHandler.testDelete(nil, map[string]string{"listtask": "23"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})

			t.Run("Shared Via NamespaceUser readonly", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"listtask": "24"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceUser write", func(t *testing.T) {
				rec, err := testHandler.testDelete(nil, map[string]string{"listtask": "25"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
			t.Run("Shared Via NamespaceUser admin", func(t *testing.T) {
				rec, err := testHandler.testDelete(nil, map[string]string{"listtask": "26"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
			})
		})
	})
	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testCreate(nil, map[string]string{"list": "1"}, `{"text":"Lorem Ipsum"}`)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testCreate(nil, map[string]string{"list": "9999"}, `{"text":"Lorem Ipsum"}`)
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeListDoesNotExist)
		})
		t.Run("Rights check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user3
				_, err := testHandler.testCreate(nil, map[string]string{"list": "2"}, `{"text":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testCreate(nil, map[string]string{"list": "6"}, `{"text":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testCreate(nil, map[string]string{"list": "7"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testCreate(nil, map[string]string{"list": "8"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testCreate(nil, map[string]string{"list": "9"}, `{"text":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testCreate(nil, map[string]string{"list": "10"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testCreate(nil, map[string]string{"list": "11"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})

			t.Run("Shared Via NamespaceTeam readonly", func(t *testing.T) {
				_, err := testHandler.testCreate(nil, map[string]string{"list": "12"}, `{"text":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceTeam write", func(t *testing.T) {
				rec, err := testHandler.testCreate(nil, map[string]string{"list": "13"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})
			t.Run("Shared Via NamespaceTeam admin", func(t *testing.T) {
				rec, err := testHandler.testCreate(nil, map[string]string{"list": "14"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})

			t.Run("Shared Via NamespaceUser readonly", func(t *testing.T) {
				_, err := testHandler.testCreate(nil, map[string]string{"list": "15"}, `{"text":"Lorem Ipsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceUser write", func(t *testing.T) {
				rec, err := testHandler.testCreate(nil, map[string]string{"list": "16"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})
			t.Run("Shared Via NamespaceUser admin", func(t *testing.T) {
				rec, err := testHandler.testCreate(nil, map[string]string{"list": "17"}, `{"text":"Lorem Ipsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"text":"Lorem Ipsum"`)
			})
		})
	})
}
