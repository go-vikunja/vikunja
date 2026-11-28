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
	"fmt"
	"net/http"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// initial_due_time on the v2 task writes (rows W-1..W-7d, W-10..W-12).
//
// Boundary: the HTTP API through the v2 router, including AutoPatch's GET → merge
// → PUT. The engine clock is substituted (models.PinEngineClockForTests, an outside-boundary
// dependency); users' zones and seeded rows are set up through the database, as fixtures are.
// Everything else is real. The tests assert only responses and persisted rows. The tests pin a
// package clock, so none of them runs in parallel.
//
// The clock: Tue 2030-01-15 12:00Z, 07:00 in New York (EST), 21:00 in Tokyo, 12:00 on the server
// (the test config's zone). 09:00 gives a different first instance in each zone.
var (
	initialDueNow     = time.Date(2030, 1, 15, 12, 0, 0, 0, time.UTC)
	initialDueNY      = time.Date(2030, 1, 15, 14, 0, 0, 0, time.UTC) // 09:00 EST today
	initialDueNYFri   = time.Date(2030, 1, 18, 14, 0, 0, 0, time.UTC) // 09:00 EST, Friday
	initialDueTokyo   = time.Date(2030, 1, 16, 0, 0, 0, 0, time.UTC)  // 09:00 JST tomorrow; today's is past
	initialDueTokyoTZ = "Asia/Tokyo"
)

const (
	initialDueDaily  = "FREQ=DAILY"
	initialDueFriday = "FREQ=WEEKLY;BYDAY=FR"
	initialDueMerge  = "application/merge-patch+json"
)

// initialDueServer is the first 09:00 after now in the server zone: the control a zone-resolution
// row must differ from.
func initialDueServer() time.Time {
	zone := config.GetTimeZone()
	local := initialDueNow.In(zone)
	at := time.Date(local.Year(), local.Month(), local.Day(), 9, 0, 0, 0, zone)
	if !at.After(initialDueNow) {
		at = at.AddDate(0, 0, 1)
	}
	return at.UTC()
}

// initialDueSetup loads the fixtures, pins the engine clock and gives user 1 the zone userZone
// ("" for none).
func initialDueSetup(t *testing.T, userZone string) (*echo.Echo, string) {
	t.Helper()
	e := zoneSetup(t)
	t.Cleanup(models.PinEngineClockForTests(initialDueNow))
	zoneSetUser(t, 1, userZone)
	return e, humaTokenFor(t, &testuser1)
}

func initialDueTaskPath(id int64) string { return fmt.Sprintf("/api/v2/tasks/%d", id) }

func initialDueCreatePath(projectID int64) string {
	return fmt.Sprintf("/api/v2/projects/%d/tasks", projectID)
}

// initialDueCreate creates a task over v2 without initial_due_time and returns its id.
func initialDueCreate(t *testing.T, e *echo.Echo, token string, projectID int64, body string) int64 {
	t.Helper()
	created := ruleTextDo(t, e, http.MethodPost, initialDueCreatePath(projectID), body, token, "", http.StatusCreated)
	return int64(created["id"].(float64))
}

// initialDuePut reads a task over v2, lets edit change the body, and PUTs it back.
func initialDuePut(t *testing.T, e *echo.Echo, token string, id int64, edit func(map[string]any), want int) map[string]any {
	t.Helper()
	task := ruleTextRead(t, e, token, initialDueTaskPath(id))
	edit(task)
	return ruleTextDo(t, e, http.MethodPut, initialDueTaskPath(id), marshal(t, task), token, "", want)
}

func initialDueDate(t *testing.T, body map[string]any) time.Time {
	t.Helper()
	raw, ok := body["due_date"].(string)
	require.True(t, ok, "due_date in %v", body)
	at, err := time.Parse(time.RFC3339, raw)
	require.NoError(t, err)
	return at.UTC()
}

// initialDueGenerated asserts a write response that seeded want, and the stored row.
func initialDueGenerated(t *testing.T, body map[string]any, want time.Time) {
	t.Helper()
	assert.Equal(t, want, initialDueDate(t, body), "response due_date")
	assert.Equal(t, models.DueDateSourceGenerated, body["due_date_source"])
	assert.NotContains(t, body, "reason")
	assert.NotContains(t, body, "initial_due_time", "write-only: never returned")
	id := int64(body["id"].(float64))
	assert.Equal(t, want, zoneStoredTask(t, id).DueDate.UTC(), "stored due_date")
}

// initialDueUnchanged asserts the stored row still has the due date, rule and done flag it had.
func initialDueUnchanged(t *testing.T, id int64, before *models.Task) {
	t.Helper()
	after := zoneStoredTask(t, id)
	assert.True(t, before.DueDate.Equal(after.DueDate), "due_date unchanged: %v → %v", before.DueDate, after.DueDate)
	assert.Equal(t, before.Repeats, after.Repeats, "rule unchanged")
	assert.Equal(t, before.Done, after.Done, "done unchanged")
	assert.True(t, before.Updated.Equal(after.Updated), "row not rewritten")
}

// W-1, W-2, W-3: create.
func TestInitialDue_Create(t *testing.T) {
	t.Run("W-1 undated DAILY at 09:00 is generated in the caller's zone", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		created := ruleTextDo(t, e, http.MethodPost, initialDueCreatePath(1),
			`{"title":"w1","repeats_rrule":"FREQ=DAILY","initial_due_time":"09:00"}`, token, "", http.StatusCreated)
		initialDueGenerated(t, created, initialDueNY)

		read := ruleTextRead(t, e, token, initialDueTaskPath(int64(created["id"].(float64))))
		for _, field := range []string{"initial_due_time", "due_date_source", "reason"} {
			assert.NotContains(t, read, field, "GET never carries %s", field)
		}
	})

	t.Run("W-1 control: without a user zone the server zone decides", func(t *testing.T) {
		e, token := initialDueSetup(t, "")
		created := ruleTextDo(t, e, http.MethodPost, initialDueCreatePath(1),
			`{"title":"w1 server","repeats_rrule":"FREQ=DAILY","initial_due_time":"09:00"}`, token, "", http.StatusCreated)
		require.NotEqual(t, initialDueNY, initialDueServer(), "the fixture discriminates")
		initialDueGenerated(t, created, initialDueServer())
	})

	t.Run("W-2 an explicit 03:30 on the spring-forward day is stored verbatim", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		// 2030-03-10 is the US spring-forward Sunday; 03:30 EDT is 07:30Z.
		created := ruleTextDo(t, e, http.MethodPost, initialDueCreatePath(1),
			`{"title":"w2","due_date":"2030-03-10T07:30:00Z","repeats_rrule":"FREQ=DAILY"}`, token, "", http.StatusCreated)
		want := time.Date(2030, 3, 10, 7, 30, 0, 0, time.UTC)
		assert.Equal(t, want, initialDueDate(t, created))
		assert.Equal(t, models.DueDateSourceExplicit, created["due_date_source"])
		assert.NotContains(t, created, "reason")
		assert.Equal(t, want, zoneStoredTask(t, int64(created["id"].(float64))).DueDate.UTC())
	})

	t.Run("an undated create without initial_due_time has no due_date_source", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		created := ruleTextDo(t, e, http.MethodPost, initialDueCreatePath(1),
			`{"title":"plain","repeats_rrule":"FREQ=DAILY"}`, token, "", http.StatusCreated)
		assert.True(t, initialDueDate(t, created).IsZero())
		assert.NotContains(t, created, "due_date_source")
		assert.NotContains(t, created, "reason")
	})

	t.Run("W-3 due_date and initial_due_time is 400 and nothing is written", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		rec := humaRequest(t, e, http.MethodPost, initialDueCreatePath(1),
			`{"title":"w3 refused","due_date":"2030-02-01T09:00:00Z","repeats_rrule":"FREQ=DAILY","initial_due_time":"09:00"}`, token, "")
		assert.Contains(t, ruleTextRefused(t, rec), "due_date")
		db.AssertMissing(t, "tasks", map[string]any{"title": "w3 refused"})
	})

	t.Run("a malformed initial_due_time is 422 from the schema", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		for _, bad := range []string{"24:00", "9:00", "09:60", "09:00:00", "noon"} {
			rec := humaRequest(t, e, http.MethodPost, initialDueCreatePath(1),
				`{"title":"bad time","repeats_rrule":"FREQ=DAILY","initial_due_time":"`+bad+`"}`, token, "")
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "%s: %s", bad, rec.Body.String())
		}
		db.AssertMissing(t, "tasks", map[string]any{"title": "bad time"})
	})
}

// W-4, W-5, W-5b: PATCH (AutoPatch GET → merge → PUT tasks-update).
func TestInitialDue_Patch(t *testing.T) {
	t.Run("W-4 PATCH initial_due_time on an undated task with a rule", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		id := initialDueCreate(t, e, token, 1, `{"title":"w4","repeats_rrule":"FREQ=DAILY"}`)
		patched := ruleTextDo(t, e, http.MethodPatch, initialDueTaskPath(id), `{"initial_due_time":"09:00"}`, token, initialDueMerge, http.StatusOK)
		initialDueGenerated(t, patched, initialDueNY)
		assert.Equal(t, initialDueDaily, storedRule(t, id), "the echoed rule is kept")
	})

	t.Run("W-5 PATCH on a dated task: the echo carries due_date, 400", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		id := initialDueCreate(t, e, token, 1, `{"title":"w5","due_date":"2030-02-01T09:00:00Z","repeats_rrule":"FREQ=DAILY"}`)
		before := zoneStoredTask(t, id)
		rec := humaRequest(t, e, http.MethodPatch, initialDueTaskPath(id), `{"initial_due_time":"09:00"}`, token, initialDueMerge)
		ruleTextRefused(t, rec)
		initialDueUnchanged(t, id, before)
	})

	t.Run("W-5b PATCH with due_date null on a dated task seeds", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		id := initialDueCreate(t, e, token, 1, `{"title":"w5b","due_date":"2030-02-01T09:00:00Z","repeats_rrule":"FREQ=DAILY"}`)
		patched := ruleTextDo(t, e, http.MethodPatch, initialDueTaskPath(id), `{"initial_due_time":"09:00","due_date":null}`, token, initialDueMerge, http.StatusOK)
		initialDueGenerated(t, patched, initialDueNY)
	})

	t.Run("a later PATCH without initial_due_time reports the stored date as explicit", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		id := initialDueCreate(t, e, token, 1, `{"title":"later","repeats_rrule":"FREQ=DAILY"}`)
		ruleTextDo(t, e, http.MethodPatch, initialDueTaskPath(id), `{"initial_due_time":"09:00"}`, token, initialDueMerge, http.StatusOK)
		patched := ruleTextDo(t, e, http.MethodPatch, initialDueTaskPath(id), `{"title":"renamed"}`, token, initialDueMerge, http.StatusOK)
		assert.Equal(t, initialDueNY, initialDueDate(t, patched))
		assert.Equal(t, models.DueDateSourceExplicit, patched["due_date_source"], "unchanged stored date, not generated by this write")
	})
}

// The JSON Patch shapes the frontend editor sends (application/json-patch+json, one add op per
// field). Same boundary as the
// rows above: the HTTP API through the v2 router and AutoPatch, the engine clock substituted.
func TestInitialDue_JSONPatch(t *testing.T) {
	t.Run("add initial_due_time with no due_date op seeds and equals the preview", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		id := initialDueCreate(t, e, token, 1, `{"title":"q3b","repeats_rrule":"FREQ=DAILY"}`)
		require.True(t, zoneStoredTask(t, id).DueDate.IsZero(), "the task starts undated")

		preview := occPreviewDecode(t, occPreview(t, e, token, map[string]any{
			"repeats_rrule": initialDueDaily, "initial_due_time": "09:00", "task_id": id, "count": 2,
		}), zoneNewYork, initialDueNow)
		preview.assertDue(t, initialDueNY, models.DueDateSourceGenerated)

		patched := ruleTextDo(t, e, http.MethodPatch, initialDueTaskPath(id),
			`[{"op":"add","path":"/initial_due_time","value":"09:00"}]`, token, ruleTextJSONPatch, http.StatusOK)
		initialDueGenerated(t, patched, initialDueNY)
		assert.True(t, preview.DueDate.Equal(initialDueDate(t, patched)), "preview %s == PATCH %s", preview.DueDate, initialDueDate(t, patched))
	})

	t.Run("add due_date 0001-01-01T00:00:00Z with initial_due_time counts as empty and seeds", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		id := initialDueCreate(t, e, token, 1, `{"title":"q3a","repeats_rrule":"FREQ=DAILY"}`)

		patched := ruleTextDo(t, e, http.MethodPatch, initialDueTaskPath(id),
			`[{"op":"add","path":"/due_date","value":"0001-01-01T00:00:00Z"},{"op":"add","path":"/initial_due_time","value":"09:00"}]`,
			token, ruleTextJSONPatch, http.StatusOK)
		initialDueGenerated(t, patched, initialDueNY)
	})

	t.Run("JSON Patch initial_due_time on a dated task is still refused", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		id := initialDueCreate(t, e, token, 1, `{"title":"q3 dated","due_date":"2030-02-01T09:00:00Z","repeats_rrule":"FREQ=DAILY"}`)
		before := zoneStoredTask(t, id)
		rec := humaRequest(t, e, http.MethodPatch, initialDueTaskPath(id),
			`[{"op":"add","path":"/initial_due_time","value":"09:00"}]`, token, ruleTextJSONPatch)
		ruleTextRefused(t, rec)
		initialDueUnchanged(t, id, before)
	})
}

// W-6, W-6b: PUT with an accepted rule that has no instance to seed.
func TestInitialDue_NoInstance(t *testing.T) {
	for _, tc := range []struct {
		row, rule, reason string
	}{
		{"W-6 UNTIL before now", "FREQ=DAILY;UNTIL=20291231T000000Z", "until_passed"},
		// now ≤ UNTIL (13:00Z) < the first 09:00 New York candidate (14:00Z).
		{"W-6b UNTIL between now and the first candidate", "FREQ=DAILY;UNTIL=20300115T130000Z", "no_occurrence"},
	} {
		t.Run(tc.row, func(t *testing.T) {
			e, token := initialDueSetup(t, zoneNewYork)
			id := initialDueCreate(t, e, token, 1, `{"title":"`+tc.row+`"}`)
			put := initialDuePut(t, e, token, id, func(task map[string]any) {
				task["repeats_rrule"] = tc.rule
				task["initial_due_time"] = "09:00"
			}, http.StatusOK)
			assert.True(t, initialDueDate(t, put).IsZero(), "no due_date: %v", put["due_date"])
			assert.Equal(t, tc.reason, put["reason"])
			assert.NotContains(t, put, "due_date_source", "absent when there is no due date")
			stored := zoneStoredTask(t, id)
			assert.True(t, stored.DueDate.IsZero())
			assert.Contains(t, stored.Repeats, "UNTIL=", "the task is saved with its rule")
		})
	}
}

// W-7, W-7b, W-7c: the refusals, and reopen-and-seed.
func TestInitialDue_Refusals(t *testing.T) {
	t.Run("W-7 without a rule: create and PATCH are 400", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		rec := humaRequest(t, e, http.MethodPost, initialDueCreatePath(1), `{"title":"w7 refused","initial_due_time":"09:00"}`, token, "")
		assert.Contains(t, ruleTextRefused(t, rec), "repeat rule")
		db.AssertMissing(t, "tasks", map[string]any{"title": "w7 refused"})

		id := initialDueCreate(t, e, token, 1, `{"title":"w7"}`)
		before := zoneStoredTask(t, id)
		rec = humaRequest(t, e, http.MethodPatch, initialDueTaskPath(id), `{"initial_due_time":"09:00"}`, token, initialDueMerge)
		ruleTextRefused(t, rec)
		initialDueUnchanged(t, id, before)
	})

	t.Run("W-7b create with done true is 400", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		rec := humaRequest(t, e, http.MethodPost, initialDueCreatePath(1),
			`{"title":"w7b refused","done":true,"repeats_rrule":"FREQ=DAILY","initial_due_time":"09:00"}`, token, "")
		assert.Contains(t, ruleTextRefused(t, rec), "done")
		db.AssertMissing(t, "tasks", map[string]any{"title": "w7b refused"})
	})

	t.Run("W-7b PUT completing and seeding is 400", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		id := initialDueCreate(t, e, token, 1, `{"title":"w7b put","repeats_rrule":"FREQ=DAILY"}`)
		before := zoneStoredTask(t, id)
		rec := humaRequest(t, e, http.MethodPut, initialDueTaskPath(id), func() string {
			task := ruleTextRead(t, e, token, initialDueTaskPath(id))
			task["done"] = true
			task["initial_due_time"] = "09:00"
			return marshal(t, task)
		}(), token, "")
		ruleTextRefused(t, rec)
		initialDueUnchanged(t, id, before)
	})

	t.Run("W-7b PATCH on an already-done task is 400", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		id := initialDueCreate(t, e, token, 1, `{"title":"w7b done","repeats_rrule":"FREQ=DAILY"}`)
		zoneSeedTask(t, id, []string{"done"}, &models.Task{Done: true})
		before := zoneStoredTask(t, id)
		rec := humaRequest(t, e, http.MethodPatch, initialDueTaskPath(id), `{"initial_due_time":"09:00"}`, token, initialDueMerge)
		ruleTextRefused(t, rec)
		initialDueUnchanged(t, id, before)
	})

	t.Run("W-7c reopen-and-seed is allowed", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		id := initialDueCreate(t, e, token, 1, `{"title":"w7c","repeats_rrule":"FREQ=DAILY"}`)
		zoneSeedTask(t, id, []string{"done"}, &models.Task{Done: true})
		patched := ruleTextDo(t, e, http.MethodPatch, initialDueTaskPath(id), `{"done":false,"initial_due_time":"09:00"}`, token, initialDueMerge, http.StatusOK)
		initialDueGenerated(t, patched, initialDueNY)
		assert.False(t, zoneStoredTask(t, id).Done)
	})

	t.Run("a caller without write access gets the same 403 as without initial_due_time", func(t *testing.T) {
		// User 1 can only read project 3 and its task 32.
		e, token := initialDueSetup(t, zoneNewYork)
		for _, body := range []string{`{"title":"no write"}`, `{"title":"no write","repeats_rrule":"FREQ=DAILY","initial_due_time":"09:00"}`} {
			rec := humaRequest(t, e, http.MethodPost, initialDueCreatePath(3), body, token, "")
			assert.Equal(t, http.StatusForbidden, rec.Code, "%s: %s", body, rec.Body.String())
		}
		db.AssertMissing(t, "tasks", map[string]any{"title": "no write"})

		before := zoneStoredTask(t, 32)
		for _, patch := range []string{`{"title":"no write"}`, `{"repeats_rrule":"FREQ=DAILY","initial_due_time":"09:00"}`} {
			rec := humaRequest(t, e, http.MethodPatch, initialDueTaskPath(32), patch, token, initialDueMerge)
			assert.Equal(t, http.StatusForbidden, rec.Code, "%s: %s", patch, rec.Body.String())
		}
		initialDueUnchanged(t, 32, before)
	})
}

// W-7d: a failure after the selection rolls the whole write back. The failures are real
// validation errors raised after the task row (create) or the reminders (update) were written in
// the transaction, so no inside-boundary collaborator is patched. Each has a positive control
// showing the same write, minus the failing part, succeeds with the generated date.
func TestInitialDue_Rollback(t *testing.T) {
	// 13:00Z: one hour before the generated 09:00 EST due date.
	reminderAt := initialDueNY.Add(-time.Hour)

	t.Run("W-7d create: a reminder that fails after the row insert leaves no task and no reminder", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		rec := humaRequest(t, e, http.MethodPost, initialDueCreatePath(1),
			`{"title":"w7d create","repeats_rrule":"FREQ=DAILY","initial_due_time":"09:00",`+
				`"reminders":[{"relative_to":"due_date","relative_period":-3600},{"relative_period":60}]}`, token, "")
		require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
		assert.InDelta(t, float64(models.ErrCodeReminderRelativeToMissing), ruleTextBody(t, rec)["code"], 0)
		db.AssertMissing(t, "tasks", map[string]any{"title": "w7d create"})

		control := ruleTextDo(t, e, http.MethodPost, initialDueCreatePath(1),
			`{"title":"w7d create control","repeats_rrule":"FREQ=DAILY","initial_due_time":"09:00",`+
				`"reminders":[{"relative_to":"due_date","relative_period":-3600}]}`, token, "", http.StatusCreated)
		initialDueGenerated(t, control, initialDueNY)
		assert.Equal(t, []time.Time{reminderAt}, zoneStoredReminders(t, int64(control["id"].(float64))),
			"the relative reminder is computed from the generated date")
	})

	t.Run("W-7d update: a cover image check after the reminders were rewritten rolls both back", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		id := initialDueCreate(t, e, token, 1, `{"title":"w7d update","repeats_rrule":"FREQ=DAILY"}`)
		zoneSeedReminders(t, id, time.Date(2030, 1, 20, 8, 0, 0, 0, time.UTC))
		before := zoneStoredTask(t, id)
		beforeReminders := zoneStoredReminders(t, id)

		rec := humaRequest(t, e, http.MethodPatch, initialDueTaskPath(id),
			`{"initial_due_time":"09:00","reminders":[{"relative_to":"due_date","relative_period":-3600}],"cover_image_attachment_id":999999}`,
			token, initialDueMerge)
		require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
		assert.InDelta(t, float64(models.ErrCodeAttachmentDoesNotBelongToTask), ruleTextBody(t, rec)["code"], 0)
		initialDueUnchanged(t, id, before)
		assert.Equal(t, beforeReminders, zoneStoredReminders(t, id), "reminders unchanged")

		control := ruleTextDo(t, e, http.MethodPatch, initialDueTaskPath(id),
			`{"initial_due_time":"09:00","reminders":[{"relative_to":"due_date","relative_period":-3600}]}`,
			token, initialDueMerge, http.StatusOK)
		initialDueGenerated(t, control, initialDueNY)
		assert.Equal(t, []time.Time{reminderAt}, zoneStoredReminders(t, id))
	})
}

// W-10: a link share with write access creates a task; a link share has no zone, so the project
// owner's is used.
func TestInitialDue_LinkShare(t *testing.T) {
	e, _ := initialDueSetup(t, zoneNewYork)
	// Share 2 has write access to project 2, owned by user 3.
	zoneSetUser(t, 3, initialDueTokyoTZ)
	token, err := auth.NewLinkShareJWTAuthtoken(&models.LinkSharing{
		ID: 2, Hash: "test2", ProjectID: 2, Permission: models.PermissionWrite,
		SharingType: models.SharingTypeWithoutPassword, SharedByID: 1,
	})
	require.NoError(t, err)

	created := ruleTextDo(t, e, http.MethodPost, initialDueCreatePath(2),
		`{"title":"w10","repeats_rrule":"FREQ=DAILY","initial_due_time":"09:00"}`, token, "", http.StatusCreated)
	initialDueGenerated(t, created, initialDueTokyo)

	id := int64(created["id"].(float64))
	zoneSeedTask(t, id, []string{"due_date"}, &models.Task{})
	patched := ruleTextDo(t, e, http.MethodPatch, initialDueTaskPath(id), `{"initial_due_time":"09:00"}`, token, initialDueMerge, http.StatusOK)
	initialDueGenerated(t, patched, initialDueTokyo)
}

// W-11a, W-11b: an update resolves the zone for the task's real project. User 1 has write access
// to project 10, owned by user 6 (Tokyo); user 1 owns project 1; the server zone is the third.
func TestInitialDue_UpdateProjectZone(t *testing.T) {
	t.Run("W-11a PUT omitting project_id uses the stored task's project", func(t *testing.T) {
		// A zoneless caller falls through to the owner of the project the zone is resolved for:
		// project 0 would fall through again, to the server zone.
		e, token := initialDueSetup(t, "")
		zoneSetUser(t, 6, initialDueTokyoTZ)
		id := initialDueCreate(t, e, token, 10, `{"title":"w11a","repeats_rrule":"FREQ=DAILY"}`)
		put := initialDuePut(t, e, token, id, func(task map[string]any) {
			delete(task, "project_id")
			task["initial_due_time"] = "09:00"
		}, http.StatusOK)
		require.NotEqual(t, initialDueServer(), initialDueTokyo, "the fixture discriminates")
		initialDueGenerated(t, put, initialDueTokyo)
		assert.Equal(t, int64(10), zoneStoredTask(t, id).ProjectID)
	})

	t.Run("W-11a with a zoned caller, owner and server all different, the caller's zone wins", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		zoneSetUser(t, 6, initialDueTokyoTZ)
		id := initialDueCreate(t, e, token, 10, `{"title":"w11a zoned","repeats_rrule":"FREQ=DAILY"}`)
		put := initialDuePut(t, e, token, id, func(task map[string]any) {
			delete(task, "project_id")
			task["initial_due_time"] = "09:00"
		}, http.StatusOK)
		initialDueGenerated(t, put, initialDueNY)
	})

	t.Run("W-11b move-and-seed uses the target project", func(t *testing.T) {
		e, token := initialDueSetup(t, "")
		zoneSetUser(t, 6, initialDueTokyoTZ)
		id := initialDueCreate(t, e, token, 1, `{"title":"w11b","repeats_rrule":"FREQ=DAILY"}`)
		put := initialDuePut(t, e, token, id, func(task map[string]any) {
			task["project_id"] = 10
			task["initial_due_time"] = "09:00"
		}, http.StatusOK)
		initialDueGenerated(t, put, initialDueTokyo)
		assert.Equal(t, int64(10), zoneStoredTask(t, id).ProjectID)
	})
}

// W-12: the selection searches the rule the write stores, whichever way the rule arrived. The
// Friday rule and the daily rule give different first dates, so the date shows which rule was used.
func TestInitialDue_FinalRule(t *testing.T) {
	t.Run("text-only create", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		created := ruleTextDo(t, e, http.MethodPost, initialDueCreatePath(1),
			`{"title":"text","repeats_rrule":"FREQ=WEEKLY;BYDAY=FR","initial_due_time":"09:00"}`, token, "", http.StatusCreated)
		initialDueGenerated(t, created, initialDueNYFri)
		assert.Contains(t, storedRule(t, int64(created["id"].(float64))), "BYDAY=FR")
	})

	t.Run("structured-only create", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		created := ruleTextDo(t, e, http.MethodPost, initialDueCreatePath(1),
			`{"title":"structured","repeat":{"freq":"weekly","interval":1,"by_day":["fr"]},"initial_due_time":"09:00"}`, token, "", http.StatusCreated)
		initialDueGenerated(t, created, initialDueNYFri)
		assert.Contains(t, storedRule(t, int64(created["id"].(float64))), "BYDAY=FR")
	})

	t.Run("PATCH echo of an unchanged rule", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		id := initialDueCreate(t, e, token, 1, `{"title":"echo","repeats_rrule":"FREQ=WEEKLY;BYDAY=FR"}`)
		rule := storedRule(t, id)
		patched := ruleTextDo(t, e, http.MethodPatch, initialDueTaskPath(id), `{"initial_due_time":"09:00"}`, token, initialDueMerge, http.StatusOK)
		initialDueGenerated(t, patched, initialDueNYFri)
		assert.Equal(t, rule, storedRule(t, id), "kept byte-identical")
	})

	t.Run("PATCH replacing the rule", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		id := initialDueCreate(t, e, token, 1, `{"title":"replace","repeats_rrule":"FREQ=DAILY"}`)
		patched := ruleTextDo(t, e, http.MethodPatch, initialDueTaskPath(id),
			`{"initial_due_time":"09:00","repeats_rrule":"FREQ=WEEKLY;BYDAY=FR"}`, token, initialDueMerge, http.StatusOK)
		initialDueGenerated(t, patched, initialDueNYFri)
		assert.Contains(t, storedRule(t, id), "BYDAY=FR")
	})

	t.Run("PATCH clearing the rule with initial_due_time is 400", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		id := initialDueCreate(t, e, token, 1, `{"title":"clear","repeats_rrule":"FREQ=DAILY"}`)
		before := zoneStoredTask(t, id)
		for _, patch := range []string{`{"initial_due_time":"09:00","repeats_rrule":""}`, `{"initial_due_time":"09:00","repeat":null}`} {
			rec := humaRequest(t, e, http.MethodPatch, initialDueTaskPath(id), patch, token, initialDueMerge)
			assert.Contains(t, ruleTextRefused(t, rec), "repeat rule", patch)
			initialDueUnchanged(t, id, before)
		}
	})

	t.Run("PATCH with a stale COUNT echo keeps the stored rule and seeds from it", func(t *testing.T) {
		e, token := initialDueSetup(t, zoneNewYork)
		id := initialDueCreate(t, e, token, 1, `{"title":"stale","repeats_rrule":"FREQ=DAILY;COUNT=5"}`)
		// A completion elsewhere lowered the stored COUNT; this client still holds COUNT=5.
		zoneSeedTask(t, id, []string{"repeats"}, &models.Task{Repeats: "FREQ=DAILY;COUNT=2"})
		patched := ruleTextDo(t, e, http.MethodPatch, initialDueTaskPath(id),
			`{"initial_due_time":"09:00","repeats_rrule":"FREQ=DAILY;COUNT=5"}`, token, initialDueMerge, http.StatusOK)
		initialDueGenerated(t, patched, initialDueNY)
		assert.Equal(t, "FREQ=DAILY;COUNT=2", storedRule(t, id))
		assert.Contains(t, patched["repeats_rrule"], "COUNT=2", "the response shows the kept rule")
	})
}
