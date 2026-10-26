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
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The server's clock cannot be pinned here, so these tests use 2030 dates with the weekday and
// DST position of the model tests' literal values: a completion steps from a future due date.

func zoneSetCaldavUser(t *testing.T, zone string) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	_, err := s.ID(testuser15.ID).Cols("timezone").Update(&user.User{Timezone: zone})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

func zoneTaskByUID(t *testing.T, uid string) *models.Task {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	task := &models.Task{}
	has, err := s.Where("uid = ?", uid).Get(task)
	require.NoError(t, err)
	require.True(t, has, "task %s", uid)
	return task
}

func zonePUT(t *testing.T, e *echo.Echo, uid, vtodo string) {
	t.Helper()
	rec := caldavPUT(t, e, "/dav/projects/36/"+uid+".ics", vtodo)
	require.GreaterOrEqual(t, rec.Code, 200, rec.Body.String())
	require.Less(t, rec.Code, 300, rec.Body.String())
}

// A CalDAV completion uses the same user's zone as every other path. Due
// Mon 2030-10-28 20:00 EDT: New York lands on Wed 20:00 EDT, the GMT server on Wed 00:00Z.
func TestRecurrenceZone_CalDAVCompletion(t *testing.T) {
	due := time.Date(2030, 10, 29, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name, zone string
		want       time.Time
	}{
		{"user in New York", "America/New_York", time.Date(2030, 10, 31, 0, 0, 0, 0, time.UTC)},
		{"user without a zone (server-zone control)", "", time.Date(2030, 10, 30, 0, 0, 0, 0, time.UTC)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := setupTestEnv(t)
			zoneSetCaldavUser(t, tc.zone)
			const uid = "recurrence-zone-weekly"
			zonePUT(t, e, uid, NewVTodo(uid, "weekly task").Due(due).Rrule("FREQ=WEEKLY;BYDAY=MO,WE,FR").Build())
			zonePUT(t, e, uid, NewVTodo(uid, "weekly task").Due(due).Rrule("FREQ=WEEKLY;BYDAY=MO,WE,FR").Status("COMPLETED").Build())

			task := zoneTaskByUID(t, uid)
			assert.False(t, task.Done, "reopened")
			assert.Equal(t, tc.want, task.DueDate.UTC())

			get := caldavGET(t, e, "/dav/projects/36/"+uid+".ics")
			require.Equal(t, http.StatusOK, get.Code)
			assert.NotContains(t, get.Body.String(), "intended")
			assert.NotContains(t, get.Body.String(), "INTENDED")
		})
	}
}

// Over CalDAV, a PUT re-sending the same DUE keeps the stored intended time, and a PUT that
// changes DUE clears it.
func TestRecurrenceZone_CalDAVIntendedDueTime(t *testing.T) {
	shifted := time.Date(2030, 3, 10, 7, 30, 0, 0, time.UTC) // Sun 03:30 EDT
	const uid = "recurrence-zone-intended"
	setup := func(t *testing.T) *echo.Echo {
		t.Helper()
		e := setupTestEnv(t)
		zoneSetCaldavUser(t, "America/New_York")
		zonePUT(t, e, uid, NewVTodo(uid, "shifted daily").Due(shifted).Rrule("FREQ=DAILY;INTERVAL=1").Build())
		s := db.NewSession()
		intended := 9000
		_, err := s.Where("uid = ?", uid).Cols("intended_due_time").Update(&models.Task{IntendedDueTime: &intended})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		require.NoError(t, s.Close())
		return e
	}

	t.Run("an unchanged DUE keeps it", func(t *testing.T) {
		e := setup(t)
		zonePUT(t, e, uid, NewVTodo(uid, "shifted daily renamed").Due(shifted).Rrule("FREQ=DAILY;INTERVAL=1").Build())

		task := zoneTaskByUID(t, uid)
		assert.Equal(t, "shifted daily renamed", task.Title)
		require.NotNil(t, task.IntendedDueTime)
		assert.Equal(t, 9000, *task.IntendedDueTime)
	})
	t.Run("a changed DUE clears it", func(t *testing.T) {
		e := setup(t)
		edited := time.Date(2030, 3, 10, 7, 20, 0, 0, time.UTC) // Sun 03:20 EDT
		zonePUT(t, e, uid, NewVTodo(uid, "shifted daily").Due(edited).Rrule("FREQ=DAILY;INTERVAL=1").Build())

		task := zoneTaskByUID(t, uid)
		assert.Equal(t, edited, task.DueDate.UTC())
		assert.Nil(t, task.IntendedDueTime)
	})
}
