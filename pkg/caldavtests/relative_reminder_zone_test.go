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

package caldavtests

import (
	"net/http"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Over CalDAV, a whole-day relative alarm PUT by a New York user is stored on New York's
// calendar, while the export keeps -PT24H, so a CalDAV client
// may fire an hour apart from Vikunja on a clock-change weekend.
//
// Known limit, not tested and not to be fixed here: RELATED=END is exported for both due-
// and end-relative reminders and imported as due-relative when DUE exists, so an end-relative
// reminder does not round-trip.

func relativeReminderRows(t *testing.T, taskID int64) []*models.TaskReminder {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	rows := []*models.TaskReminder{}
	require.NoError(t, s.Where("task_id = ?", taskID).Find(&rows))
	return rows
}

func TestRelativeReminderZone_CalDAV(t *testing.T) {
	sunEvening := time.Date(2027, 3, 15, 0, 0, 0, 0, time.UTC) // Sun 2027-03-14 20:00 EDT
	cases := []struct {
		name, uid string
		vtodo     *VTodoBuilder
		relTo     models.ReminderRelation
		period    int64
		want      time.Time
		trigger   string
	}{
		{"start-relative one day before", "rr-zone-start",
			NewVTodo("rr-zone-start", "start").DtStart(sunEvening).AlarmRelativeStart("-P1D"),
			models.ReminderRelationStartDate, -86400,
			time.Date(2027, 3, 14, 1, 0, 0, 0, time.UTC), // Sat 20:00 EST
			"TRIGGER;RELATED=START:-PT24H0M0S"},
		{"due-relative one day before", "rr-zone-due",
			NewVTodo("rr-zone-due", "due").Due(sunEvening).AlarmRelativeEnd("-P1D"),
			models.ReminderRelationDueDate, -86400,
			time.Date(2027, 3, 14, 1, 0, 0, 0, time.UTC), // Sat 20:00 EST
			"TRIGGER;RELATED=END:-PT24H0M0S"},
		{"due-relative one day two hours before", "rr-zone-due-mixed",
			NewVTodo("rr-zone-due-mixed", "due mixed").Due(sunEvening).AlarmRelativeEnd("-P1DT2H"),
			models.ReminderRelationDueDate, -93600,
			time.Date(2027, 3, 13, 23, 0, 0, 0, time.UTC), // Sat 18:00 EST
			"TRIGGER;RELATED=END:-PT26H0M0S"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := setupTestEnv(t)
			zoneSetCaldavUser(t, "America/New_York")
			zonePUT(t, e, tc.uid, tc.vtodo.Build())

			assertStored := func(t *testing.T) {
				t.Helper()
				rows := relativeReminderRows(t, zoneTaskByUID(t, tc.uid).ID)
				require.Len(t, rows, 1)
				assert.Equal(t, tc.relTo, rows[0].RelativeTo)
				assert.Equal(t, tc.period, rows[0].RelativePeriod)
				assert.True(t, rows[0].Reminder.Equal(tc.want), "stored %v, want %v", rows[0].Reminder.UTC(), tc.want)
			}
			assertStored(t)

			// The export is unchanged: whole days stay -PT24H.
			get := caldavGET(t, e, "/dav/projects/36/"+tc.uid+".ics")
			require.Equal(t, http.StatusOK, get.Code)
			assert.Contains(t, get.Body.String(), tc.trigger)

			// Round trip (start- and due-relative only): PUT back what GET returned.
			zonePUT(t, e, tc.uid, get.Body.String())
			assertStored(t)
		})
	}
}
