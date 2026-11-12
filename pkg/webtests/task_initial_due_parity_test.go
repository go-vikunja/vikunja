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
	"net/http"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parityRow is one admitted application row: an initial_due_time selection under the
// first-occurrence policy. want is the due date both sides must store; zero
// means neither side selects one and both give wantReason.
type parityRow struct {
	name, zone, rule, hhmm string
	now                    time.Time
	want                   time.Time
	wantReason             string
}

// parityRows is the admitted set: P-1, P-2, P-3, P-4, P-5, P-7a, the library's two
// initial-search rows and the no-date rows P-7b / P-7c, as named in the table below.
// Deliberately absent: the phase fixtures (application policy, TestInitialDuePhaseFixtures in
// pkg/models/task_initial_due_test.go), the anchored library controls (task_initial_due_oracle_test.go),
// every row with a timezone override, and the BYSETPOS gap limit.
func parityRows(t *testing.T) []parityRow {
	berlin, ny := occZone(t, occBerlin), occZone(t, occNewYork)
	b := func(d, h, m, sec int) time.Time { return time.Date(2026, 10, d, h, m, sec, 0, berlin) }
	n := func(d, h, m int) time.Time { return time.Date(2023, 3, d, h, m, 0, 0, ny) }
	utc := func(d, h, m int) time.Time { return time.Date(2023, 3, d, h, m, 0, 0, time.UTC) }

	return []parityRow{
		{name: "P-1 DAILY 09:00, now 10:00", zone: occBerlin, rule: "FREQ=DAILY", hhmm: "09:00", now: b(28, 10, 0, 0), want: b(29, 9, 0, 0)},
		{name: "P-2 now exactly 09:00:00", zone: occBerlin, rule: "FREQ=DAILY", hhmm: "09:00", now: b(28, 9, 0, 0), want: b(29, 9, 0, 0)},
		{name: "P-3 NY gap seed DAILY 02:30", zone: occNewYork, rule: "FREQ=DAILY", hhmm: "02:30", now: n(12, 1, 0), want: n(13, 2, 30)},
		{name: "P-4 NY gap seed MINUTELY;INTERVAL=15", zone: occNewYork, rule: "FREQ=MINUTELY;INTERVAL=15", hhmm: "02:30", now: n(12, 1, 0), want: n(12, 3, 0)},
		{name: "P-5 COUNT=1, seed before now", zone: occBerlin, rule: "FREQ=DAILY;COUNT=1", hhmm: "09:00", now: b(28, 10, 0, 0), want: b(29, 9, 0, 0)},
		// 2026-10-29 09:00 CET is 08:00Z.
		{name: "P-7a UNTIL = first candidate", zone: occBerlin, rule: "FREQ=DAILY;UNTIL=20261029T080000Z", hhmm: "09:00", now: b(28, 10, 0, 0), want: b(29, 9, 0, 0)},
		{name: "ny-daily-nominal-gap-seed", zone: occNewYork, rule: "FREQ=DAILY;COUNT=1", hhmm: "02:30", now: utc(12, 5, 0), want: utc(13, 6, 30)},
		{name: "ny-minutely-nominal-gap-seed", zone: occNewYork, rule: "FREQ=MINUTELY;INTERVAL=15;COUNT=2", hhmm: "02:30", now: utc(12, 5, 0), want: utc(12, 7, 0)},
		{name: "P-7b now <= UNTIL < first candidate", zone: occBerlin, rule: "FREQ=DAILY;UNTIL=20261029T075959Z", hhmm: "09:00", now: b(28, 10, 0, 0), wantReason: "no_occurrence"},
		{name: "P-7c UNTIL < now", zone: occBerlin, rule: "FREQ=DAILY;UNTIL=20261027T000000Z", hhmm: "09:00", now: b(28, 10, 0, 0), wantReason: "until_passed"},
	}
}

// parityAssertPreview asserts a preview answer against the row and returns its due date (zero
// when none was selected).
func parityAssertPreview(t *testing.T, row parityRow, answer occPreviewAnswer) time.Time {
	t.Helper()
	if row.want.IsZero() {
		answer.assertNoDueDate(t, row.wantReason)
		return time.Time{}
	}
	answer.assertDue(t, row.want, models.DueDateSourceGenerated)
	return answer.DueDate.UTC()
}

// parityAssertWrite asserts a write response against the row and the stored row, and returns the
// saved due date (zero when none was selected).
func parityAssertWrite(t *testing.T, row parityRow, body map[string]any) time.Time {
	t.Helper()
	id := int64(body["id"].(float64))
	saved := initialDueDate(t, body)
	assert.True(t, saved.Equal(zoneStoredTask(t, id).DueDate.UTC()), "response due_date %s is the stored one", saved)
	assert.NotContains(t, body, "initial_due_time", "write-only")
	if row.want.IsZero() {
		assert.True(t, saved.IsZero(), "no due date saved, got %s", saved)
		assert.NotContains(t, body, "due_date_source")
		assert.Equal(t, row.wantReason, body["reason"])
		return time.Time{}
	}
	assert.True(t, row.want.Equal(saved), "saved due_date %s, want %s", saved, row.want)
	assert.Equal(t, models.DueDateSourceGenerated, body["due_date_source"])
	assert.NotContains(t, body, "reason")
	return saved
}

// TestInitialDueParity: for every admitted application row, the due date the preview shows equals the due date the write saves, under one pinned engine
// clock, the zone resolved from the task/project context, and identical inputs. No request sends a
// timezone override. Each row runs two legs: preview with project_id, then create in that project;
// and preview with task_id of an undated task, then PATCH initial_due_time on it. Each leg also
// asserts the row's expected value, so two equally wrong answers cannot pass.
//
// Zones: user 1 (owner of project 1 and the caller) stores the row's zone, which differs from the
// server's (GMT in the test config), so a side that fell back to the server zone would differ.
//
// Black-box boundary: the HTTP API through the v2 router, including AutoPatch's
// GET → merge → PUT, with real database and auth. The engine clock is substituted
// (models.PinEngineClockForTests, an outside-boundary dependency) and the user's zone is set
// through the database, as a fixture. Everything else is real; the test asserts only responses
// and persisted rows. It pins a package clock, so it never runs in parallel.
func TestInitialDueParity(t *testing.T) {
	for _, row := range parityRows(t) {
		t.Run(row.name, func(t *testing.T) {
			occRequireDistinctZones(t, row.zone)

			t.Run("project_id preview == create", func(t *testing.T) {
				e := zoneSetup(t)
				occPinClock(t, row.now)
				zoneSetUser(t, 1, row.zone)
				token := humaTokenFor(t, &testuser1)

				preview := occPreviewDecode(t, occPreview(t, e, token, map[string]any{
					"repeats_rrule": row.rule, "initial_due_time": row.hhmm, "project_id": 1, "count": 2,
				}), row.zone, row.now)
				previewed := parityAssertPreview(t, row, preview)

				created := ruleTextDo(t, e, http.MethodPost, initialDueCreatePath(1), marshal(t, map[string]any{
					"title": "parity " + row.name, "repeats_rrule": row.rule, "initial_due_time": row.hhmm,
				}), token, "", http.StatusCreated)
				saved := parityAssertWrite(t, row, created)

				assert.True(t, previewed.Equal(saved), "preview %s == create %s", previewed, saved)
				if row.want.IsZero() {
					assert.Equal(t, preview.Reason, created["reason"], "same reason on both sides")
				}
			})

			t.Run("task_id preview == PATCH", func(t *testing.T) {
				e := zoneSetup(t)
				occPinClock(t, row.now)
				zoneSetUser(t, 1, row.zone)
				token := humaTokenFor(t, &testuser1)

				id := initialDueCreate(t, e, token, 1, marshal(t, map[string]any{
					"title": "parity undated " + row.name, "repeats_rrule": row.rule,
				}))
				require.True(t, zoneStoredTask(t, id).DueDate.IsZero(), "the task starts undated")

				preview := occPreviewDecode(t, occPreview(t, e, token, map[string]any{
					"repeats_rrule": row.rule, "initial_due_time": row.hhmm, "task_id": id, "count": 2,
				}), row.zone, row.now)
				previewed := parityAssertPreview(t, row, preview)

				patched := ruleTextDo(t, e, http.MethodPatch, initialDueTaskPath(id),
					marshal(t, map[string]any{"initial_due_time": row.hhmm}), token, initialDueMerge, http.StatusOK)
				saved := parityAssertWrite(t, row, patched)

				assert.True(t, previewed.Equal(saved), "preview %s == PATCH %s", previewed, saved)
				if row.want.IsZero() {
					assert.Equal(t, preview.Reason, patched["reason"], "same reason on both sides")
				}
			})
		})
	}
}
