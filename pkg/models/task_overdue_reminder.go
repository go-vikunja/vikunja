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
	"slices"
	"time"

	"code.vikunja.io/api/pkg/events"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"

	"code.vikunja.io/api/pkg/log"

	"code.vikunja.io/api/pkg/user"

	"xorm.io/builder"
	"xorm.io/xorm"
)

func overdueSlot(window cron.Window, timezone, clock string) (time.Time, error) {
	if timezone == "" {
		timezone = config.GetTimeZone().String()
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}
	if clock == "" {
		clock = config.DefaultSettingsOverdueTaskRemindersTime.GetString()
	}
	if clock == "" {
		clock = "09:00"
	}
	daily, err := time.Parse("15:04", clock)
	if err != nil {
		return time.Time{}, err
	}
	local := window.To.In(location)
	slot := time.Date(local.Year(), local.Month(), local.Day(), daily.Hour(), daily.Minute(), 0, 0, location)
	if slot.After(window.To) {
		local = local.AddDate(0, 0, -1)
		slot = time.Date(local.Year(), local.Month(), local.Day(), daily.Hour(), daily.Minute(), 0, 0, location)
	}
	if !slot.After(window.From) {
		return time.Time{}, nil
	}
	return slot, nil
}

func getOverdueTasksInWindow(s *xorm.Session, window cron.Window, cond builder.Cond) (map[int64]*userWithTasks, error) {
	var tasks []*Task
	err := s.Join("INNER", "projects", "projects.id = tasks.project_id").
		Where(builder.NotNull{"due_date"}).And(builder.Lt{"due_date": window.To}).
		And(builder.Eq{"tasks.done": false, "projects.is_archived": false}).And(builder.And(builder.IsNull{"tasks.deleted_at"}, builder.IsNull{"projects.deleted_at"})).Find(&tasks)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.ID)
	}
	users, err := getTaskUsersForTasks(s, ids, cond)
	if err != nil {
		return nil, err
	}
	result := make(map[int64]*userWithTasks)
	for _, tu := range users {
		slot, err := overdueSlot(window, tu.User.Timezone, tu.User.OverdueTasksRemindersTime)
		if err != nil {
			log.Warningf("Invalid overdue reminder time for user %d: %v", tu.User.ID, err)
			continue
		}
		if slot.IsZero() || !tu.Task.DueDate.Before(slot) {
			continue
		}
		ut := result[tu.User.ID]
		if ut == nil {
			ut = &userWithTasks{user: tu.User, assigned: make(map[int64]*Task), followed: make(map[int64]*Task), slot: slot}
			result[tu.User.ID] = ut
		}
		if tu.IsAssignee {
			ut.assigned[tu.Task.ID] = tu.Task
		} else {
			ut.followed[tu.Task.ID] = tu.Task
		}
	}
	return result, nil
}

type userWithTasks struct {
	slot     time.Time
	user     *user.User
	assigned map[int64]*Task
	followed map[int64]*Task
}

func (ut *userWithTasks) allTasks() []*Task {
	return append(mapToSlice(ut.assigned), mapToSlice(ut.followed)...)
}

func RegisterOverdueReminderCron() {
	if err := cron.Register("overdue-reminders", "* * * * *", enqueueOverdueReminders); err != nil {
		log.Fatalf("Register overdue reminders: %v", err)
	}
}

func mapToSlice(m map[int64]*Task) []*Task {
	tasks := make([]*Task, 0, len(m))
	for _, t := range m {
		tasks = append(tasks, t)
	}
	return tasks
}

// hasNonWebhookListener catches listeners registered by plugins.
func hasNonWebhookListener(eventName string) bool {
	return slices.ContainsFunc(events.ListenersFor(eventName), func(l events.Listener) bool {
		_, isWebhook := l.(*WebhookListener)
		return !isWebhook
	})
}

// overdueEventsWithListeners lets the cron skip dispatching events nobody handles,
// since each dispatched event costs the webhook listener a db session.
func overdueEventsWithListeners(s *xorm.Session) (single, batch bool, err error) {
	singleName := (&TaskOverdueEvent{}).Name()
	batchName := (&TasksOverdueEvent{}).Name()

	single = hasNonWebhookListener(singleName)
	batch = hasNonWebhookListener(batchName)
	if single && batch {
		return
	}

	webhooks := []*Webhook{}
	err = s.Cols("events").Find(&webhooks)
	if err != nil {
		return
	}

	for _, w := range webhooks {
		single = single || slices.Contains(w.Events, singleName)
		batch = batch || slices.Contains(w.Events, batchName)
	}
	return
}
