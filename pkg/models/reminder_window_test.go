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

package models

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/cron"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// Retain the historical selector fixtures while testing the new window API separately.
func getTasksWithRemindersDueAndTheirUsers(s *xorm.Session, now time.Time) ([]*ReminderDueNotification, error) {
	return getTaskRemindersInWindow(s, cron.Window{From: now.Add(-time.Second), To: now.Add(time.Minute - time.Second)}, nil)
}
func getUndoneOverdueTasks(s *xorm.Session, now time.Time, cond builder.Cond) (map[int64]*userWithTasks, error) {
	return getOverdueTasksInWindow(s, cron.Window{From: now.Add(-time.Second), To: now.Add(time.Minute - time.Second)}, cond)
}

func TestOverdueSlotUsesUserLocalDate(t *testing.T) {
	to := time.Date(2026, 11, 11, 22, 2, 0, 0, time.UTC)
	slot, err := overdueSlot(cron.Window{From: to.Add(-5 * time.Minute), To: to}, "Australia/Sydney", "09:00")
	require.NoError(t, err)
	assert.True(t, slot.Equal(to.Add(-2*time.Minute)))
	slot, err = overdueSlot(cron.Window{From: to.Add(-2 * time.Minute), To: to}, "Australia/Sydney", "09:00")
	require.NoError(t, err)
	assert.True(t, slot.IsZero(), "left boundary must not repeat")
}
