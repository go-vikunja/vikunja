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

package feeds

import (
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTasksCalendar(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u1 := &user.User{
		ID:       1,
		Username: "user1",
	}

	// Fixtures: project 22 (user 1) holds task 36 (open, due 2018-10-30) and
	// task 38 (done, same due date).
	t.Run("project feed leaves done tasks out by default", func(t *testing.T) {
		ical, err := BuildTasksCalendar(s, u1, 22, false)
		require.NoError(t, err)

		assert.True(t, strings.HasPrefix(ical, "BEGIN:VCALENDAR\r\n"))
		assert.Contains(t, ical, "X-WR-CALNAME:Test22 archived individually\r\n")
		assert.Contains(t, ical, "SUMMARY:task #36\r\n")
		assert.Contains(t, ical, "DTSTART:20181030T222524Z\r\n")
		assert.NotContains(t, ical, "task #37 done with due date")
	})

	t.Run("include done", func(t *testing.T) {
		ical, err := BuildTasksCalendar(s, u1, 22, true)
		require.NoError(t, err)

		assert.Contains(t, ical, "SUMMARY:task #36\r\n")
		assert.Contains(t, ical, "SUMMARY:task #37 done with due date\r\n")
	})

	t.Run("tasks without dates are skipped", func(t *testing.T) {
		ical, err := BuildTasksCalendar(s, u1, 1, false)
		require.NoError(t, err)

		assert.Contains(t, ical, "SUMMARY:task #5 higher due date\r\n")
		assert.Contains(t, ical, "SUMMARY:task #9 with start and end date\r\n")
		assert.NotContains(t, ical, "SUMMARY:task #1\r\n")
		assert.NotContains(t, ical, "task #8 with end date")
	})

	t.Run("user feed spans the readable, non-archived projects", func(t *testing.T) {
		ical, err := BuildTasksCalendar(s, u1, 0, false)
		require.NoError(t, err)

		assert.Contains(t, ical, "X-WR-CALNAME:Vikunja\r\n")
		assert.Contains(t, ical, "SUMMARY:task #5 higher due date\r\n")
		assert.NotContains(t, ical, "SUMMARY:task #36\r\n", "project 22 is archived")
		assert.NotContains(t, ical, "Title Caldav Test", "project 36 belongs to user 15")
	})

	t.Run("forbidden for a project the user cannot read", func(t *testing.T) {
		_, err := BuildTasksCalendar(s, &user.User{ID: 13}, 1, false)
		require.Error(t, err)
		assert.True(t, models.IsErrUserDoesNotHaveAccessToProject(err), "got %v", err)
	})

	t.Run("nonexistent project", func(t *testing.T) {
		_, err := BuildTasksCalendar(s, u1, 9999999, false)
		require.Error(t, err)
		assert.True(t, models.IsErrProjectDoesNotExist(err), "got %v", err)
	})
}
