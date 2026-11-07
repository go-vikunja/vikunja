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
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests cannot pin the server's clock, so they use 2030 dates: a normal-mode completion
// steps from a future due date, not from now. Each 2030 date has the weekday and DST position of
// the literal golden value the model tests (pkg/models, under withFixedNow) assert.

const (
	zoneWeeklyRule = "FREQ=WEEKLY;BYDAY=MO,WE,FR"
	zoneDailyRule  = "FREQ=DAILY;INTERVAL=1"
	zoneMergePatch = "application/merge-patch+json"
	zoneNewYork    = "America/New_York"
)

// The shared fixture: due Mon 2030-10-28 20:00 EDT. In New York the next occurrence is Wed 20:00 EDT;
// in the server zone (GMT) the due date is Tue 00:00, so the next one is Wed 00:00Z, a day early.
var (
	zoneWeeklyDue    = time.Date(2030, 10, 29, 0, 0, 0, 0, time.UTC)
	zoneWeeklyNY     = time.Date(2030, 10, 31, 0, 0, 0, 0, time.UTC)
	zoneWeeklyServer = time.Date(2030, 10, 30, 0, 0, 0, 0, time.UTC)
)

func zoneSetUser(t *testing.T, userID int64, zone string) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	_, err := s.ID(userID).Cols("timezone").Update(&user.User{Timezone: zone})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

// zoneSeedTask writes a rule and dates straight onto a task, as a CalDAV client or an earlier
// completion would have stored them.
func zoneSeedTask(t *testing.T, taskID int64, cols []string, task *models.Task) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	_, err := s.ID(taskID).Cols(cols...).Update(task)
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

func zoneSeedRule(t *testing.T, taskID int64, rule string, due time.Time) {
	t.Helper()
	zoneSeedTask(t, taskID, []string{"repeats", "due_date", "done"}, &models.Task{Repeats: rule, DueDate: due})
}

func zoneSeedReminders(t *testing.T, taskID int64, at ...time.Time) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	for _, r := range at {
		_, err := s.Insert(&models.TaskReminder{TaskID: taskID, Reminder: r})
		require.NoError(t, err)
	}
	require.NoError(t, s.Commit())
}

func zoneStoredTask(t *testing.T, taskID int64) *models.Task {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	task := &models.Task{}
	has, err := s.ID(taskID).Get(task)
	require.NoError(t, err)
	require.True(t, has)
	return task
}

func zoneStoredReminders(t *testing.T, taskID int64) []time.Time {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	rows := []*models.TaskReminder{}
	require.NoError(t, s.Where("task_id = ?", taskID).OrderBy("reminder ASC, id ASC").Find(&rows))
	out := make([]time.Time, len(rows))
	for i, r := range rows {
		out[i] = r.Reminder.UTC()
	}
	return out
}

// zoneDo sends a request and checks its status.
func zoneDo(t *testing.T, e *echo.Echo, method, path, body, token, contentType string) {
	t.Helper()
	rec := humaRequest(t, e, method, path, body, token, contentType)
	require.GreaterOrEqual(t, rec.Code, 200, rec.Body.String())
	require.Less(t, rec.Code, 300, rec.Body.String())
}

func zoneV1Echo(t *testing.T, e *echo.Echo, token string, taskID int64, edit func(map[string]any)) {
	t.Helper()
	path := fmt.Sprintf("/api/v1/tasks/%d", taskID)
	rec := humaRequest(t, e, http.MethodGet, path, "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	task := map[string]any{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &task))
	edit(task)
	zoneDo(t, e, http.MethodPost, path, marshal(t, task), token, "")
}

func zoneSetup(t *testing.T) *echo.Echo {
	t.Helper()
	e, err := setupTestEnv()
	require.NoError(t, err)
	return e
}

// Every completion path stores the same next date for the same user.
// The no-zone runs are the positive control: the server zone gives a different date, so the
// fixture discriminates.
func TestRecurrenceZone_CompletionPaths(t *testing.T) {
	paths := []struct {
		name     string
		complete func(t *testing.T, e *echo.Echo, token string)
	}{
		{"V1Post", func(t *testing.T, e *echo.Echo, token string) {
			zoneV1Echo(t, e, token, 1, func(task map[string]any) { task["done"] = true })
		}},
		{"V2Patch", func(t *testing.T, e *echo.Echo, token string) {
			zoneDo(t, e, http.MethodPatch, "/api/v2/tasks/1", `{"done":true}`, token, zoneMergePatch)
		}},
		{"KanbanDoneBucket", func(t *testing.T, e *echo.Echo, token string) {
			zoneDo(t, e, http.MethodPut, "/api/v2/projects/1/views/4/buckets/3/tasks", `{"task_id":1}`, token, "")
		}},
		{"BulkV2", func(t *testing.T, e *echo.Echo, token string) {
			zoneDo(t, e, http.MethodPut, "/api/v2/tasks/bulk", `{"task_ids":[1],"fields":["done"],"values":{"done":true}}`, token, "")
		}},
		{"BulkV1", func(t *testing.T, e *echo.Echo, token string) {
			zoneDo(t, e, http.MethodPost, "/api/v1/tasks/bulk", `{"task_ids":[1],"fields":["done"],"values":{"done":true}}`, token, "")
		}},
	}
	for _, zone := range []struct {
		name, zone string
		want       time.Time
	}{{"NewYork", zoneNewYork, zoneWeeklyNY}, {"ServerZoneControl", "", zoneWeeklyServer}} {
		for _, p := range paths {
			t.Run("TestRecurrenceZone_"+p.name+"_"+zone.name, func(t *testing.T) {
				e := zoneSetup(t)
				zoneSetUser(t, 1, zone.zone)
				zoneSeedRule(t, 1, zoneWeeklyRule, zoneWeeklyDue)

				p.complete(t, e, humaTokenFor(t, &testuser1))

				task := zoneStoredTask(t, 1)
				assert.False(t, task.Done, "reopened")
				assert.Equal(t, zone.want, task.DueDate.UTC())
			})
		}

		// A project move with done: true into project 1 runs both completion sites: view 4's done
		// bucket (updateTaskBucket) and the update itself. The final row equals a single update.
		t.Run("TestRecurrenceZone_ProjectMoveDone_"+zone.name, func(t *testing.T) {
			e := zoneSetup(t)
			zoneSetUser(t, 1, zone.zone)
			s := db.NewSession()
			source := &models.Project{Title: "zone move source"}
			require.NoError(t, models.CreateProject(s, source, &testuser1, true, true))
			task := &models.Task{Title: "moved", ProjectID: source.ID}
			require.NoError(t, task.Create(s, &testuser1))
			require.NoError(t, s.Commit())
			require.NoError(t, s.Close())
			zoneSeedRule(t, task.ID, zoneWeeklyRule, zoneWeeklyDue)

			zoneDo(t, e, http.MethodPatch, fmt.Sprintf("/api/v2/tasks/%d", task.ID), `{"project_id":1,"done":true}`,
				humaTokenFor(t, &testuser1), zoneMergePatch)

			stored := zoneStoredTask(t, task.ID)
			assert.Equal(t, int64(1), stored.ProjectID)
			assert.False(t, stored.Done)
			assert.Equal(t, zone.want, stored.DueDate.UTC())
		})
	}
}

// The JWT carries no zone (pkg/user GetUserFromClaims: id, username, admin); the stored one is
// re-fetched.
func TestRecurrenceZone_JWTUser_UsesStoredZone(t *testing.T) {
	e := zoneSetup(t)
	zoneSetUser(t, 1, zoneNewYork)
	zoneSeedRule(t, 1, zoneWeeklyRule, zoneWeeklyDue)

	zoneDo(t, e, http.MethodPatch, "/api/v2/tasks/1", `{"done":true}`, humaTokenFor(t, &testuser1), zoneMergePatch)

	assert.Equal(t, zoneWeeklyNY, zoneStoredTask(t, 1).DueDate.UTC())
}

// A second member completes on their own calendar. Due Mon 2030-10-28 23:30Z: in
// Berlin that is Tue 00:30 CET, so the next one is Wed 00:30 CET; for the New York owner and for
// the GMT server it is Monday evening, so theirs is Wed 23:30Z.
func TestRecurrenceZone_SecondMemberBerlin(t *testing.T) {
	due := time.Date(2030, 10, 28, 23, 30, 0, 0, time.UTC)
	want := time.Date(2030, 10, 29, 23, 30, 0, 0, time.UTC)
	ownerOrServer := time.Date(2030, 10, 30, 23, 30, 0, 0, time.UTC)
	require.NotEqual(t, want, ownerOrServer, "the expectation differs from both fallbacks")

	e := zoneSetup(t)
	zoneSetUser(t, 1, zoneNewYork)
	zoneSetUser(t, 2, "Europe/Berlin")
	s := db.NewSession()
	_, err := s.Insert(&models.ProjectUser{UserID: 2, ProjectID: 1, Permission: models.PermissionWrite})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	require.NoError(t, s.Close())
	zoneSeedRule(t, 1, zoneWeeklyRule, due)

	zoneDo(t, e, http.MethodPatch, "/api/v2/tasks/1", `{"done":true}`, humaTokenFor(t, &testuser2), zoneMergePatch)

	assert.Equal(t, want, zoneStoredTask(t, 1).DueDate.UTC())
}

// A link share has no zone: the owner of the project (user3, project 2) decides.
func TestRecurrenceZone_LinkShare_UsesOwnerZone(t *testing.T) {
	for _, tc := range []struct {
		name, zone string
		want       time.Time
	}{{"owner in New York", zoneNewYork, zoneWeeklyNY}, {"owner without a zone", "", zoneWeeklyServer}} {
		t.Run(tc.name, func(t *testing.T) {
			e := zoneSetup(t)
			zoneSetUser(t, 3, tc.zone)
			zoneSeedRule(t, 13, zoneWeeklyRule, zoneWeeklyDue)
			token, err := auth.NewLinkShareJWTAuthtoken(&models.LinkSharing{
				ID: 2, Hash: "test2", ProjectID: 2, Permission: models.PermissionWrite,
				SharingType: models.SharingTypeWithoutPassword, SharedByID: 1,
			})
			require.NoError(t, err)

			zoneDo(t, e, http.MethodPatch, "/api/v2/tasks/13", `{"done":true}`, token, zoneMergePatch)

			task := zoneStoredTask(t, 13)
			assert.False(t, task.Done)
			assert.Equal(t, tc.want, task.DueDate.UTC())
		})
	}
}

// A gap occurrence is skipped end to end: daily, due Sat 2030-03-09 02:30 EST. The first completion
// skips Sun 2030-03-10 (02:30 does not exist) and lands on Mon 02:30 EDT; the second gives Tue
// 02:30 EDT. Each chain goes through each save site.
func TestRecurrenceZone_GapSkipChains(t *testing.T) {
	due := time.Date(2030, 3, 9, 7, 30, 0, 0, time.UTC)
	monday := time.Date(2030, 3, 11, 6, 30, 0, 0, time.UTC)
	tuesday := time.Date(2030, 3, 12, 6, 30, 0, 0, time.UTC)
	patch := func(t *testing.T, e *echo.Echo, token string) {
		zoneDo(t, e, http.MethodPatch, "/api/v2/tasks/1", `{"done":true}`, token, zoneMergePatch)
	}
	kanban := func(t *testing.T, e *echo.Echo, token string) {
		zoneDo(t, e, http.MethodPut, "/api/v2/projects/1/views/4/buckets/3/tasks", `{"task_id":1}`, token, "")
	}
	for _, chain := range []struct {
		name          string
		first, second func(*testing.T, *echo.Echo, string)
	}{{"PatchThenPatch", patch, patch}, {"PatchThenKanban", patch, kanban}, {"KanbanThenPatch", kanban, patch}} {
		t.Run(chain.name, func(t *testing.T) {
			e := zoneSetup(t)
			zoneSetUser(t, 1, zoneNewYork)
			zoneSeedRule(t, 1, zoneDailyRule, due)
			token := humaTokenFor(t, &testuser1)

			chain.first(t, e, token)
			assert.Equal(t, monday, zoneStoredTask(t, 1).DueDate.UTC(), "the gap day is skipped")

			chain.second(t, e, token)
			assert.Equal(t, tuesday, zoneStoredTask(t, 1).DueDate.UTC())
		})
	}
}

// Companion dates are not recurrence instances (a documented known limit): a start, an absolute
// reminder and an end in the gap move forward by the gap and keep their shifted time on later
// steps. The due date (Sat 2030-03-09 20:00 EST) never meets the gap.
func TestRecurrenceZone_CompanionsKeepShiftedTime(t *testing.T) {
	e := zoneSetup(t)
	zoneSetUser(t, 1, zoneNewYork)
	zoneSeedTask(t, 1, []string{"repeats", "due_date", "start_date", "end_date", "done"}, &models.Task{
		Repeats:   zoneDailyRule,
		DueDate:   time.Date(2030, 3, 10, 1, 0, 0, 0, time.UTC), // Sat 20:00 EST
		StartDate: time.Date(2030, 3, 9, 7, 30, 0, 0, time.UTC), // Sat 02:30 EST
		EndDate:   time.Date(2030, 3, 9, 7, 45, 0, 0, time.UTC), // Sat 02:45 EST
	})
	zoneSeedReminders(t, 1, time.Date(2030, 3, 9, 7, 15, 0, 0, time.UTC)) // Sat 02:15 EST
	token := humaTokenFor(t, &testuser1)

	for _, day := range []int{10, 11} { // Sun, then Mon: 03:30, 03:15 and 03:45 EDT
		zoneDo(t, e, http.MethodPatch, "/api/v2/tasks/1", `{"done":true}`, token, zoneMergePatch)
		task := zoneStoredTask(t, 1)
		assert.Equal(t, time.Date(2030, 3, day+1, 0, 0, 0, 0, time.UTC), task.DueDate.UTC(), "due 20:00 EDT")
		assert.Equal(t, time.Date(2030, 3, day, 7, 30, 0, 0, time.UTC), task.StartDate.UTC())
		assert.Equal(t, time.Date(2030, 3, day, 7, 45, 0, 0, time.UTC), task.EndDate.UTC())
		assert.Equal(t, []time.Time{time.Date(2030, 3, day, 7, 15, 0, 0, time.UTC)}, zoneStoredReminders(t, 1))
	}
}

// Documented known limit: absolute reminders
// Sat 02:30 and 03:30 EST both land on Sun 03:30 EDT (02:30 does not exist that day), and the
// reminder dedup, unchanged from upstream, keeps one row per instant.
func TestRecurrenceZone_RemindersLandingOnOneInstantMerge(t *testing.T) {
	e := zoneSetup(t)
	zoneSetUser(t, 1, zoneNewYork)
	zoneSeedRule(t, 1, zoneDailyRule, time.Date(2030, 3, 10, 1, 0, 0, 0, time.UTC)) // Sat 20:00 EST
	zoneSeedReminders(t, 1, time.Date(2030, 3, 9, 7, 30, 0, 0, time.UTC), time.Date(2030, 3, 9, 8, 30, 0, 0, time.UTC))

	zoneDo(t, e, http.MethodPatch, "/api/v2/tasks/1", `{"done":true}`, humaTokenFor(t, &testuser1), zoneMergePatch)

	assert.Equal(t, []time.Time{time.Date(2030, 3, 10, 7, 30, 0, 0, time.UTC)}, zoneStoredReminders(t, 1),
		"the two reminders become one row at Sun 03:30 EDT")
}
