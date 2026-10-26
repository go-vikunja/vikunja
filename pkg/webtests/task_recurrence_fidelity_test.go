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
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaskRecurrenceFidelity proves the recurrence write and read rules through real v1 and v2
// HTTP requests.
// Every subtest reloads the fixtures and works on task #1 (user1, project 1), whose rule is
// stored directly, the way a CalDAV client or the RRULE migration would have stored it.
func TestTaskRecurrenceFidelity(t *testing.T) {
	const (
		easterRule      = "FREQ=YEARLY;BYEASTER=0"
		simpleRule      = "FREQ=DAILY;INTERVAL=1"
		policyRejected  = "FREQ=HOURLY;INTERVAL=1;BYHOUR=9"
		taskPathV1      = "/api/v1/tasks/1"
		taskPathV2      = "/api/v2/tasks/1"
		mergePatch      = "application/merge-patch+json"
		readOnlyMessage = "cannot be edited here"
	)

	setup := func(t *testing.T, rule string) (*echo.Echo, string) {
		t.Helper()
		e, err := setupTestEnv()
		require.NoError(t, err)
		seedStoredRule(t, 1, rule)
		return e, humaTokenFor(t, &testuser1)
	}

	t.Run("v2 PATCH of the title keeps a read-only rule byte-identical", func(t *testing.T) {
		e, token := setup(t, easterRule)

		rec := humaRequest(t, e, http.MethodPatch, taskPathV2, `{"title":"renamed"}`, token, mergePatch)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"title":"renamed"`)
		assert.Equal(t, easterRule, storedRule(t, 1))
	})

	t.Run("v2 PATCH of a different repeat on a read-only rule is refused", func(t *testing.T) {
		e, token := setup(t, easterRule)

		rec := humaRequest(t, e, http.MethodPatch, taskPathV2, `{"repeat":{"freq":"yearly","interval":2}}`, token, mergePatch)
		require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), readOnlyMessage)
		assert.Contains(t, rec.Body.String(), "CalDAV")
		assert.Equal(t, easterRule, storedRule(t, 1), "a refused write leaves the rule alone")
	})

	t.Run("v2 PATCH of repeat null clears a read-only rule", func(t *testing.T) {
		e, token := setup(t, easterRule)

		rec := humaRequest(t, e, http.MethodPatch, taskPathV2, `{"repeat":null}`, token, mergePatch)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Empty(t, storedRule(t, 1))
	})

	t.Run("v2 PATCH of repeat null clears a representable rule", func(t *testing.T) {
		e, token := setup(t, simpleRule)

		rec := humaRequest(t, e, http.MethodPatch, taskPathV2, `{"repeat":null}`, token, mergePatch)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Empty(t, storedRule(t, 1))
	})

	t.Run("v2 PATCH of the title keeps a COUNT the server decremented", func(t *testing.T) {
		e, token := setup(t, "FREQ=DAILY;INTERVAL=1;COUNT=3")
		s := db.NewSession()
		_, err := s.ID(1).Cols("due_date").Update(&models.Task{DueDate: time.Date(2030, 1, 10, 12, 0, 0, 0, time.UTC)})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		require.NoError(t, s.Close())

		rec := humaRequest(t, e, http.MethodPatch, taskPathV2, `{"done":true}`, token, mergePatch)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		afterCompletion := storedRule(t, 1)
		require.Contains(t, afterCompletion, "COUNT=2", "completing the task decrements COUNT")

		rec = humaRequest(t, e, http.MethodPatch, taskPathV2, `{"title":"renamed after completion"}`, token, mergePatch)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, afterCompletion, storedRule(t, 1), "a rename echoes the decremented rule and must keep it")
	})

	// A stale client echoing the COUNT it read before a completion must not reset the decremented
	// COUNT; a deliberate change still saves as sent.
	countSetup := func(t *testing.T) (*echo.Echo, string, map[string]any) {
		t.Helper()
		e, token := setup(t, "FREQ=DAILY;INTERVAL=1;COUNT=3")
		s := db.NewSession()
		_, err := s.ID(1).Cols("due_date").Update(&models.Task{DueDate: time.Date(2030, 1, 10, 12, 0, 0, 0, time.UTC)})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		require.NoError(t, s.Close())

		stale := readTaskV1(t, e, token) // read before the completion: COUNT=3
		rec := humaRequest(t, e, http.MethodPatch, taskPathV2, `{"done":true}`, token, mergePatch)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		require.Contains(t, storedRule(t, 1), "COUNT=2", "completing the task decrements COUNT")
		return e, token, stale
	}

	t.Run("v2 PATCH carrying the stale COUNT keeps the decremented one", func(t *testing.T) {
		e, token, _ := countSetup(t)
		afterCompletion := storedRule(t, 1)

		rec := humaRequest(t, e, http.MethodPatch, taskPathV2,
			`{"title":"renamed","repeat":{"freq":"daily","interval":1,"count":3}}`, token, mergePatch)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"title":"renamed"`)
		assert.Equal(t, afterCompletion, storedRule(t, 1))
	})

	t.Run("v1 POST of a copy read before the completion keeps the decremented COUNT", func(t *testing.T) {
		e, token, stale := countSetup(t)
		afterCompletion := storedRule(t, 1)

		stale["title"] = "renamed by a stale tab"
		rec := humaRequest(t, e, http.MethodPost, taskPathV1, marshal(t, stale), token, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, afterCompletion, storedRule(t, 1))
	})

	t.Run("v2 PATCH of a lower COUNT saves it", func(t *testing.T) {
		e, token, _ := countSetup(t)

		rec := humaRequest(t, e, http.MethodPatch, taskPathV2, `{"repeat":{"freq":"daily","interval":1,"count":1}}`, token, mergePatch)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Contains(t, storedRule(t, 1), "COUNT=1")
	})

	t.Run("v2 PATCH of a changed rule with the old COUNT saves it as sent", func(t *testing.T) {
		e, token, _ := countSetup(t)

		rec := humaRequest(t, e, http.MethodPatch, taskPathV2, `{"repeat":{"freq":"weekly","interval":1,"count":3}}`, token, mergePatch)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		stored := storedRule(t, 1)
		assert.Contains(t, stored, "FREQ=WEEKLY")
		assert.Contains(t, stored, "COUNT=3")
	})

	// An update that replaces the rule and completes the task in one request completes it by the
	// rule it stores, not the one it replaces.
	completeWith := func(t *testing.T, stored, body string) (*echo.Echo, *models.Task) {
		t.Helper()
		e, token := setup(t, stored)
		s := db.NewSession()
		_, err := s.ID(1).Cols("due_date", "done").Update(&models.Task{DueDate: time.Date(2030, 1, 10, 12, 0, 0, 0, time.UTC)})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		require.NoError(t, s.Close())

		rec := humaRequest(t, e, http.MethodPatch, taskPathV2, body, token, mergePatch)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		task := &models.Task{}
		s = db.NewSession()
		defer s.Close()
		has, err := s.ID(1).Get(task)
		require.NoError(t, err)
		require.True(t, has)
		return e, task
	}

	t.Run("completing while replacing the rule steps and decrements the new rule", func(t *testing.T) {
		_, task := completeWith(t, "FREQ=DAILY;INTERVAL=1;COUNT=3",
			`{"done":true,"repeat":{"freq":"weekly","interval":1,"count":5}}`)

		assert.False(t, task.Done)
		assert.Contains(t, task.Repeats, "FREQ=WEEKLY")
		assert.Contains(t, task.Repeats, "COUNT=4")
		assert.Equal(t, time.Date(2030, 1, 17, 12, 0, 0, 0, time.UTC), task.DueDate.UTC(), "advanced by a week")
	})

	t.Run("completing a stored COUNT rule whose decremented form exceeds the column leaves it done", func(t *testing.T) {
		// A pre-existing row (seeded directly): its date-only UNTIL gains a time (+8 characters)
		// when completion writes the decremented rule back, which would not fit varchar(500).
		long := "FREQ=YEARLY;COUNT=3;UNTIL=20400101;BYYEARDAY=1"
		for i := 2; len(long)+len(","+strconv.Itoa(i)) <= 498; i++ {
			long += "," + strconv.Itoa(i)
		}
		require.LessOrEqual(t, len(long), 500)
		_, task := completeWith(t, long, `{"done":true}`)

		assert.True(t, task.Done, "not advanced")
		assert.Equal(t, long, task.Repeats, "the stored rule is unchanged")
		assert.Equal(t, time.Date(2030, 1, 10, 12, 0, 0, 0, time.UTC), task.DueDate.UTC(), "dates unchanged")
	})

	// Completion never stores a date outside [1000-01-01, 9999-12-31], the range of MySQL's
	// DATETIME (the strictest supported database): such a task stays done, its dates unchanged.
	seedTask1 := func(t *testing.T, cols []string, task *models.Task) {
		t.Helper()
		s := db.NewSession()
		_, err := s.ID(1).Cols(cols...).Update(task)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		require.NoError(t, s.Close())
	}
	readTask1 := func(t *testing.T) *models.Task {
		t.Helper()
		s := db.NewSession()
		defer s.Close()
		task := &models.Task{}
		has, err := s.ID(1).Get(task)
		require.NoError(t, err)
		require.True(t, has)
		return task
	}

	t.Run("completing a task whose start would move before the year 1000 leaves it done", func(t *testing.T) {
		e, token := setup(t, "FREQ=DAILY;INTERVAL=1")
		due := time.Date(2030, 10, 23, 12, 0, 0, 0, time.UTC)
		start := time.Date(1000, 10, 23, 12, 0, 0, 0, time.UTC)
		seedTask1(t, []string{"due_date", "start_date", "repeats_from_current_date", "done"},
			&models.Task{DueDate: due, StartDate: start, RepeatsFromCurrentDate: true})

		rec := humaRequest(t, e, http.MethodPatch, taskPathV2, `{"done":true}`, token, mergePatch)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		task := readTask1(t)
		assert.True(t, task.Done, "not advanced: the start would land in the year 996")
		assert.Equal(t, due, task.DueDate.UTC())
		assert.Equal(t, start, task.StartDate.UTC())
	})

	t.Run("completing a task whose relative reminder would move past 9999 leaves it done", func(t *testing.T) {
		e, token := setup(t, "FREQ=DAILY;INTERVAL=1")
		due := time.Date(9999, 12, 29, 12, 0, 0, 0, time.UTC)
		seedTask1(t, []string{"due_date", "done"}, &models.Task{DueDate: due})
		s := db.NewSession()
		_, err := s.Insert(&models.TaskReminder{TaskID: 1, Reminder: due.Add(48 * time.Hour), RelativeTo: models.ReminderRelationDueDate, RelativePeriod: 172800})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		require.NoError(t, s.Close())

		rec := humaRequest(t, e, http.MethodPatch, taskPathV2, `{"done":true}`, token, mergePatch)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		task := readTask1(t)
		assert.True(t, task.Done, "not advanced: the reminder would land on 10000-01-01")
		assert.Equal(t, due, task.DueDate.UTC())
		reminders := []*models.TaskReminder{}
		s = db.NewSession()
		defer s.Close()
		require.NoError(t, s.Where("task_id = ?", 1).Find(&reminders))
		require.Len(t, reminders, 1)
		assert.Equal(t, due.Add(48*time.Hour), reminders[0].Reminder.UTC())
	})

	t.Run("completing while adding a rule to a plain task leaves it done with the rule stored", func(t *testing.T) {
		// Upstream behaviour is kept: no advance and no COUNT decrement.
		_, task := completeWith(t, "", `{"done":true,"repeat":{"freq":"daily","interval":1,"count":3}}`)

		assert.True(t, task.Done)
		assert.Equal(t, "FREQ=DAILY;INTERVAL=1;COUNT=3", task.Repeats)
		assert.Equal(t, time.Date(2030, 1, 10, 12, 0, 0, 0, time.UTC), task.DueDate.UTC(), "not rescheduled")
	})

	t.Run("completing while replacing the rule and its flag steps by the new flag", func(t *testing.T) {
		before := time.Now()
		_, task := completeWith(t, "FREQ=DAILY;INTERVAL=1;COUNT=3",
			`{"done":true,"repeats_from_current_date":true,"repeat":{"freq":"weekly","interval":1,"count":5}}`)

		assert.False(t, task.Done)
		assert.True(t, task.RepeatsFromCurrentDate)
		assert.Contains(t, task.Repeats, "FREQ=WEEKLY")
		assert.Contains(t, task.Repeats, "COUNT=4")
		// From the completion date, not from the stored due date (which would give 2030-01-17).
		assert.WithinRange(t, task.DueDate, before.Add(7*24*time.Hour-time.Minute), time.Now().Add(7*24*time.Hour+time.Minute))
	})

	t.Run("completing with a stale COUNT decrements the stored rule", func(t *testing.T) {
		_, task := completeWith(t, "FREQ=DAILY;INTERVAL=1;COUNT=2",
			`{"done":true,"repeat":{"freq":"daily","interval":1,"count":3}}`)

		assert.False(t, task.Done)
		assert.Equal(t, "FREQ=DAILY;INTERVAL=1;COUNT=1", task.Repeats)
		assert.Equal(t, time.Date(2030, 1, 11, 12, 0, 0, 0, time.UTC), task.DueDate.UTC())
	})

	t.Run("completing while clearing the rule reopens the task once and keeps the rule cleared", func(t *testing.T) {
		_, task := completeWith(t, "FREQ=DAILY;INTERVAL=1;COUNT=3", `{"done":true,"repeat":null}`)

		assert.False(t, task.Done, "completed by the old rule, as a v1 {done: true} always was")
		assert.Empty(t, task.Repeats, "the old rule is not written back over the clear")
		assert.Equal(t, time.Date(2030, 1, 11, 12, 0, 0, 0, time.UTC), task.DueDate.UTC())
	})

	t.Run("an exhausted COUNT rule stays cleared after a later update", func(t *testing.T) {
		e, token := setup(t, "FREQ=DAILY;INTERVAL=1;COUNT=1")

		rec := humaRequest(t, e, http.MethodPatch, taskPathV2, `{"done":true}`, token, mergePatch)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		require.Empty(t, storedRule(t, 1), "completing the last occurrence clears the rule")

		rec = humaRequest(t, e, http.MethodPatch, taskPathV2, `{"title":"renamed after the last one"}`, token, mergePatch)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Empty(t, storedRule(t, 1))
		assert.Contains(t, rec.Body.String(), `"done":true`)
	})

	t.Run("v2 PATCH of a repeat whose BY* lists would expand unbounded is refused fast", func(t *testing.T) {
		e, token := setup(t, simpleRule)
		for _, tc := range []struct {
			name, refusal string
			zeros         int
		}{
			// ~3 KB, ~125M timestamps: refused by the length limit before anything else.
			{"500 zeros per list", "characters long", 500},
			// Short enough for the column, still 70^3 timestamps: refused by the safety bound.
			{"70 zeros per list", "more than once", 70},
		} {
			t.Run(tc.name, func(t *testing.T) {
				zeros := strings.TrimSuffix(strings.Repeat("0,", tc.zeros), ",")
				body := `{"repeat":{"freq":"daily","interval":1,"by_hour":[` + zeros + `],"by_minute":[` + zeros + `],"by_second":[` + zeros + `]}}`

				start := time.Now()
				rec := humaRequest(t, e, http.MethodPatch, taskPathV2, body, token, mergePatch)
				assert.Less(t, time.Since(start), 2*time.Second, "the rule must be refused before it is expanded")
				require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
				assert.Contains(t, rec.Body.String(), tc.refusal)
				assert.Equal(t, simpleRule, storedRule(t, 1))
			})
		}
	})

	t.Run("v2 PATCH of a repeat longer than the stored column is refused, not a server error", func(t *testing.T) {
		e, token := setup(t, simpleRule)
		yearDays := make([]string, 200)
		for i := range yearDays {
			yearDays[i] = strconv.Itoa(i + 1)
		}
		// Within the safety bound (distinct, in-domain values) but over 500 characters as text.
		body := `{"repeat":{"freq":"yearly","interval":1,"by_year_day":[` + strings.Join(yearDays, ",") + `]}}`

		rec := humaRequest(t, e, http.MethodPatch, taskPathV2, body, token, mergePatch)
		require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), "500 characters")
		assert.Equal(t, simpleRule, storedRule(t, 1))
	})

	t.Run("v2 PATCH of a repeat over the INTERVAL, COUNT or BYSETPOS limits is refused", func(t *testing.T) {
		e, token := setup(t, simpleRule)
		for _, tc := range []struct{ repeat, refusal string }{
			{`{"freq":"weekly","interval":52181}`, "INTERVAL"},
			{`{"freq":"weekly","interval":2635249153387078802}`, "INTERVAL"},
			{`{"freq":"secondly","interval":1,"count":2147483647}`, "COUNT"},
			{`{"freq":"daily","interval":1,"count":10001}`, "COUNT"},
			{`{"freq":"daily","interval":1,"by_set_pos":[2,3]}`, "BYSETPOS"},
			{`{"freq":"yearly","interval":1,"by_set_pos":[1,2,3,4,5,6,7,8,9,10,11]}`, "BYSETPOS"},
			{`{"freq":"weekly","interval":1,"by_day":["2mo"]}`, "BYDAY"},
		} {
			if strconv.IntSize == 32 && strings.Contains(tc.repeat, "2635249153387078802") {
				continue // 32-bit: the interval does not fit the API's int at all
			}
			rec := humaRequest(t, e, http.MethodPatch, taskPathV2, `{"repeat":`+tc.repeat+`}`, token, mergePatch)
			require.Equal(t, http.StatusBadRequest, rec.Code, "%s: %s", tc.repeat, rec.Body.String())
			assert.Contains(t, rec.Body.String(), tc.refusal, tc.repeat)
			assert.Equal(t, simpleRule, storedRule(t, 1))
		}
	})

	t.Run("v2 PATCH of a repeat with 10 BYSETPOS values is accepted", func(t *testing.T) {
		e, token := setup(t, simpleRule)

		rec := humaRequest(t, e, http.MethodPatch, taskPathV2, `{"repeat":{"freq":"yearly","interval":1,"by_set_pos":[1,2,3,4,5,6,7,8,9,10]}}`, token, mergePatch)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Contains(t, storedRule(t, 1), "BYSETPOS=1,2,3,4,5,6,7,8,9,10")
	})

	t.Run("v2 PATCH of a normal multi-value repeat is accepted", func(t *testing.T) {
		e, token := setup(t, simpleRule)

		rec := humaRequest(t, e, http.MethodPatch, taskPathV2, `{"repeat":{"freq":"daily","interval":1,"by_hour":[9,17],"by_minute":[0,30]}}`, token, mergePatch)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, "FREQ=DAILY;INTERVAL=1;BYHOUR=9,17;BYMINUTE=0,30", storedRule(t, 1))
	})

	t.Run("v2 PATCH of a different repeat replaces the rule with the request's COUNT", func(t *testing.T) {
		e, token := setup(t, simpleRule)

		rec := humaRequest(t, e, http.MethodPatch, taskPathV2, `{"repeat":{"freq":"weekly","interval":2,"count":5}}`, token, mergePatch)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		stored := storedRule(t, 1)
		assert.Contains(t, stored, "FREQ=WEEKLY")
		assert.Contains(t, stored, "INTERVAL=2")
		assert.Contains(t, stored, "COUNT=5")
	})

	t.Run("v1 POST of the task as read keeps the rule, omitting repeat clears it", func(t *testing.T) {
		e, token := setup(t, easterRule)

		task := readTaskV1(t, e, token)
		task["title"] = "renamed over v1"
		rec := humaRequest(t, e, http.MethodPost, taskPathV1, marshal(t, task), token, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, easterRule, storedRule(t, 1), "an echoed read-only repeat keeps the stored text")

		task = readTaskV1(t, e, token)
		task["repeat"] = map[string]any{"freq": "yearly", "interval": 2}
		rec = humaRequest(t, e, http.MethodPost, taskPathV1, marshal(t, task), token, "")
		require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), readOnlyMessage)
		assert.Equal(t, easterRule, storedRule(t, 1))

		task = readTaskV1(t, e, token)
		delete(task, "repeat")
		rec = humaRequest(t, e, http.MethodPost, taskPathV1, marshal(t, task), token, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Empty(t, storedRule(t, 1), "v1 is a full replace: an absent repeat clears the rule")
	})

	t.Run("v1 POST of repeat null clears the rule", func(t *testing.T) {
		e, token := setup(t, easterRule)

		task := readTaskV1(t, e, token)
		task["repeat"] = nil
		rec := humaRequest(t, e, http.MethodPost, taskPathV1, marshal(t, task), token, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Empty(t, storedRule(t, 1))
	})

	t.Run("v2 GET flags a read-only projection and omits the flag otherwise", func(t *testing.T) {
		e, token := setup(t, easterRule)

		rec := humaRequest(t, e, http.MethodGet, taskPathV2, "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		repeat := responseRepeat(t, rec.Body.Bytes())
		require.NotNil(t, repeat, "a stored rule always has a structured repeat")
		assert.Equal(t, true, repeat["read_only"])

		seedStoredRule(t, 1, simpleRule)
		rec = humaRequest(t, e, http.MethodGet, taskPathV2, "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		repeat = responseRepeat(t, rec.Body.Bytes())
		require.NotNil(t, repeat)
		_, has := repeat["read_only"]
		assert.False(t, has, "read_only is omitted for a representable rule")
	})

	t.Run("a field-masked update never fails on the stored rule", func(t *testing.T) {
		e, token := setup(t, policyRejected)

		rec := humaRequest(t, e, http.MethodPut, "/api/v2/tasks/bulk",
			`{"task_ids":[1],"fields":["title"],"values":{"title":"masked v2"}}`, token, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, policyRejected, storedRule(t, 1))

		rec = humaRequest(t, e, http.MethodPost, "/api/v1/tasks/bulk",
			`{"task_ids":[1],"fields":["title"],"values":{"title":"masked v1"}}`, token, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, policyRejected, storedRule(t, 1))

		stored := &models.Task{}
		s := db.NewSession()
		defer s.Close()
		has, err := s.ID(1).Get(stored)
		require.NoError(t, err)
		require.True(t, has)
		assert.Equal(t, "masked v1", stored.Title)
	})

	t.Run("v2 PATCH of the title never fails on a policy-rejected stored rule", func(t *testing.T) {
		e, token := setup(t, policyRejected)

		rec := humaRequest(t, e, http.MethodPatch, taskPathV2, `{"title":"renamed"}`, token, mergePatch)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, policyRejected, storedRule(t, 1))
	})

	t.Run("duplicating a task copies a policy-rejected stored rule verbatim", func(t *testing.T) {
		// Task 2 (no attachments, so no file fixtures are needed), as stored by a CalDAV client.
		e, token := setup(t, simpleRule)
		seedStoredRule(t, 2, policyRejected)

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/tasks/2/duplicate", ``, token, "")
		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
		var created struct {
			Task struct {
				ID int64 `json:"id"`
			} `json:"duplicated_task"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created), rec.Body.String())
		require.NotZero(t, created.Task.ID, rec.Body.String())
		assert.Equal(t, policyRejected, storedRule(t, created.Task.ID))
	})

	t.Run("the public API still refuses a policy-rejected repeat on create and update", func(t *testing.T) {
		e, token := setup(t, simpleRule)
		rejectedRepeat := `{"freq":"hourly","interval":1,"by_hour":[9]}`

		rec := humaRequest(t, e, http.MethodPatch, taskPathV2, `{"repeat":`+rejectedRepeat+`}`, token, mergePatch)
		require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), "sub-daily")
		assert.Equal(t, simpleRule, storedRule(t, 1))

		rec = humaRequest(t, e, http.MethodPut, "/api/v1/projects/1/tasks", `{"title":"new","repeat":`+rejectedRepeat+`}`, token, "")
		require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), "sub-daily")
	})

	t.Run("bulk create with an invalid repeat names the failing index", func(t *testing.T) {
		e, token := setup(t, "")

		rec := humaRequest(t, e, http.MethodPost, "/api/v2/projects/1/tasks/bulk",
			`{"tasks":[{"title":"ok"},{"title":"ok too"},{"title":"bad","repeat":{"freq":"fortnightly","interval":1}}]}`, token, "")
		require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), "index 2")
		var problem v2ProblemJSON
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &problem))
		assert.Equal(t, models.ErrCodeInvalidTaskInBulkCreation, problem.Code)
		db.AssertMissing(t, "tasks", map[string]any{"title": "ok too"})
	})
}

func seedStoredRule(t *testing.T, taskID int64, rule string) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	_, err := s.ID(taskID).Cols("repeats").Update(&models.Task{Repeats: rule})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

func storedRule(t *testing.T, taskID int64) string {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	task := &models.Task{}
	has, err := s.ID(taskID).Get(task)
	require.NoError(t, err)
	require.True(t, has)
	return task.Repeats
}

func readTaskV1(t *testing.T, e *echo.Echo, token string) map[string]any {
	t.Helper()
	rec := humaRequest(t, e, http.MethodGet, "/api/v1/tasks/1", "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	task := map[string]any{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &task))
	return task
}

func responseRepeat(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var task struct {
		Repeat map[string]any `json:"repeat"`
	}
	require.NoError(t, json.Unmarshal(body, &task))
	return task.Repeat
}

func marshal(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return strings.TrimSpace(string(b))
}
