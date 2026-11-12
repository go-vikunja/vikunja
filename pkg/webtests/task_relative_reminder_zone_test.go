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
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A whole-day relative reminder counts calendar days in the saver's zone on every write
// path. The fixture is due Sun 2027-03-14 20:00 EDT with a "1 day before" reminder: Sat 20:00 EST
// (2027-03-14T01:00Z); the elapsed value, Sat 19:00 EST, is 00:00Z.
const (
	rrDue          = "2027-03-15T00:00:00Z" // Sun 2027-03-14 20:00 EDT
	rrDayBefore    = `[{"relative_to":"due_date","relative_period":-86400}]`
	rrMovedDue     = "2027-03-22T00:00:00Z" // Sun 2027-03-21 20:00 EDT
	rrMovedWantRem = "2027-03-21T00:00:00Z" // Sat 2027-03-20 20:00 EDT
)

var rrWantReminder = time.Date(2027, 3, 14, 1, 0, 0, 0, time.UTC) // Sat 2027-03-13 20:00 EST

// rrDo sends a request, requires a 2xx status and returns the decoded task.
func rrDo(t *testing.T, e *echo.Echo, method, path, body, token, contentType string) map[string]any {
	t.Helper()
	rec := humaRequest(t, e, method, path, body, token, contentType)
	require.GreaterOrEqual(t, rec.Code, 200, rec.Body.String())
	require.Less(t, rec.Code, 300, rec.Body.String())
	task := map[string]any{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &task))
	return task
}

func rrTaskID(t *testing.T, task map[string]any) int64 {
	t.Helper()
	id, ok := task["id"].(float64)
	require.True(t, ok, "the response carries the task id")
	return int64(id)
}

// rrAssertReminder checks the response's only reminder and the stored row against want, and that
// the reminder object keeps its JSON shape.
func rrAssertReminder(t *testing.T, task map[string]any, want time.Time) {
	t.Helper()
	reminders, ok := task["reminders"].([]any)
	require.True(t, ok, "reminders: %v", task["reminders"])
	require.Len(t, reminders, 1)
	reminder, ok := reminders[0].(map[string]any)
	require.True(t, ok)
	keys := make([]string, 0, len(reminder))
	for k := range reminder {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	assert.Equal(t, []string{"relative_period", "relative_to", "reminder"}, keys, "the reminder's JSON shape is unchanged")
	assert.InDelta(t, -86400, reminder["relative_period"], 0, "relative_period keeps its meaning")

	got, err := time.Parse(time.RFC3339, reminder["reminder"].(string))
	require.NoError(t, err)
	assert.True(t, got.Equal(want), "response reminder %v, want %v", got.UTC(), want)
	assert.Equal(t, []time.Time{want}, zoneStoredReminders(t, rrTaskID(t, task)), "stored reminder")
}

func rrMustTime(t *testing.T, s string) time.Time {
	t.Helper()
	at, err := time.Parse(time.RFC3339, s)
	require.NoError(t, err)
	return at
}

func TestRelativeReminderZone_V1(t *testing.T) {
	e := zoneSetup(t)
	zoneSetUser(t, 1, zoneNewYork)
	token := humaTokenFor(t, &testuser1)

	created := rrDo(t, e, http.MethodPut, "/api/v1/projects/1/tasks",
		`{"title":"v1 relative","due_date":"`+rrDue+`","reminders":`+rrDayBefore+`}`, token, "")
	rrAssertReminder(t, created, rrWantReminder)

	// Re-sent with a new title.
	created["title"] = "v1 relative renamed"
	updated := rrDo(t, e, http.MethodPost, fmt.Sprintf("/api/v1/tasks/%d", rrTaskID(t, created)), marshal(t, created), token, "")
	rrAssertReminder(t, updated, rrWantReminder)
}

func TestRelativeReminderZone_V2(t *testing.T) {
	e := zoneSetup(t)
	zoneSetUser(t, 1, zoneNewYork)
	token := humaTokenFor(t, &testuser1)

	created := rrDo(t, e, http.MethodPost, "/api/v2/projects/1/tasks",
		`{"title":"v2 relative","due_date":"`+rrDue+`","reminders":`+rrDayBefore+`}`, token, "")
	rrAssertReminder(t, created, rrWantReminder)
	path := fmt.Sprintf("/api/v2/tasks/%d", rrTaskID(t, created))

	// A merge patch of the title only.
	rrAssertReminder(t, rrDo(t, e, http.MethodPatch, path, `{"title":"v2 renamed"}`, token, zoneMergePatch), rrWantReminder)

	// A due-date change recomputes the reminder on the new date's calendar.
	moved := rrDo(t, e, http.MethodPatch, path, `{"due_date":"`+rrMovedDue+`"}`, token, zoneMergePatch)
	rrAssertReminder(t, moved, rrMustTime(t, rrMovedWantRem))
}

// The server zone is the positive control: without a user zone (GMT) the reminder is the elapsed
// value, so the New York runs above discriminate.
func TestRelativeReminderZone_ServerZoneControl(t *testing.T) {
	e := zoneSetup(t)
	created := rrDo(t, e, http.MethodPost, "/api/v2/projects/1/tasks",
		`{"title":"server zone","due_date":"`+rrDue+`","reminders":`+rrDayBefore+`}`, humaTokenFor(t, &testuser1), "")
	rrAssertReminder(t, created, time.Date(2027, 3, 14, 0, 0, 0, 0, time.UTC))
}

// A kanban completion recomputes the reminder in the completion's zone. 2030 dates, as in
// task_recurrence_zone_test.go, so the due date stays in the future: due Sat 2030-03-09 20:00 EST,
// New York's clocks spring forward on Sun 2030-03-10.
func TestRelativeReminderZone_KanbanCompletion(t *testing.T) {
	e := zoneSetup(t)
	zoneSetUser(t, 1, zoneNewYork)
	zoneSeedRule(t, 1, zoneDailyRule, time.Date(2030, 3, 10, 1, 0, 0, 0, time.UTC)) // Sat 20:00 EST
	s := db.NewSession()
	_, err := s.Insert(&models.TaskReminder{TaskID: 1, RelativeTo: models.ReminderRelationDueDate, RelativePeriod: -86400,
		Reminder: time.Date(2030, 3, 9, 1, 0, 0, 0, time.UTC)})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	require.NoError(t, s.Close())
	token := humaTokenFor(t, &testuser1)

	zoneDo(t, e, http.MethodPut, "/api/v2/projects/1/views/4/buckets/3/tasks", `{"task_id":1}`, token, "")
	task := zoneStoredTask(t, 1)
	assert.False(t, task.Done, "reopened")
	assert.Equal(t, time.Date(2030, 3, 11, 0, 0, 0, 0, time.UTC), task.DueDate.UTC(), "Sun 20:00 EDT")
	assert.Equal(t, []time.Time{time.Date(2030, 3, 10, 1, 0, 0, 0, time.UTC)}, zoneStoredReminders(t, 1),
		"Sat 20:00 EST; elapsed would be 00:00Z")

	zoneDo(t, e, http.MethodPut, "/api/v2/projects/1/views/4/buckets/3/tasks", `{"task_id":1}`, token, "")
	assert.Equal(t, time.Date(2030, 3, 12, 0, 0, 0, 0, time.UTC), zoneStoredTask(t, 1).DueDate.UTC(), "Mon 20:00 EDT")
	assert.Equal(t, []time.Time{time.Date(2030, 3, 11, 0, 0, 0, 0, time.UTC)}, zoneStoredReminders(t, 1), "Sun 20:00 EDT")
}
