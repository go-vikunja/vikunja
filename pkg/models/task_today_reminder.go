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
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"

	"xorm.io/builder"
	"xorm.io/xorm"
)

func getTasksForDailyReminder(s *xorm.Session, now time.Time) (usersWithTasks map[int64]*userWithTasks, err error) {
	now = utils.GetTimeWithoutSeconds(now)
	nextMinute := now.Add(1 * time.Minute)

	var tasks []*Task
	err = s.
		Where("due_date is not null AND due_date < ? AND projects.is_archived = false", nextMinute.Add(time.Hour*38).Format(dbTimeFormat)).
		Join("LEFT", "projects", "projects.id = tasks.project_id").
		And("done = false").
		Find(&tasks)
	if err != nil {
		return
	}

	if len(tasks) == 0 {
		return
	}

	var taskIDs []int64
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.ID)
	}

	users, err := getTaskUsersForTasks(s, taskIDs, builder.Or(
		builder.Eq{"users.overdue_tasks_reminders_enabled": true},
		builder.Eq{"users.today_tasks_reminders_enabled": true},
	))
	if err != nil {
		return
	}

	if len(users) == 0 {
		return
	}

	uts := make(map[int64]*userWithTasks)
	tzs := make(map[string]*time.Location)
	for _, t := range users {
		if t.User.Timezone == "" {
			t.User.Timezone = config.GetTimeZone().String()
		}

		tz, exists := tzs[t.User.Timezone]
		if !exists {
			tz, err = time.LoadLocation(t.User.Timezone)
			if err != nil {
				return
			}
			tzs[t.User.Timezone] = tz
		}

		// If it is time for that current user, add the task to their project of overdue tasks
		tm, err := time.Parse("15:04", t.User.OverdueTasksRemindersTime)
		if err != nil {
			return nil, err
		}
		reminderTime := time.Date(now.Year(), now.Month(), now.Day(), tm.Hour(), tm.Minute(), 0, 0, tz)
		isTimeForReminder := reminderTime.After(now) || reminderTime.Equal(now.In(tz))
		wasTimeForReminder := reminderTime.Before(nextMinute)
		taskDue := t.Task.DueDate.In(tz)
		endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, tz)
		if isTimeForReminder && wasTimeForReminder {
			_, exists := uts[t.User.ID]
			if !exists {
				uts[t.User.ID] = &userWithTasks{
					user:     t.User,
					overdue:  make(map[int64]*Task),
					dueToday: make(map[int64]*Task),
				}
			}

			if t.User.OverdueTasksRemindersEnabled && reminderTime.After(taskDue) {
				uts[t.User.ID].overdue[t.Task.ID] = t.Task
				continue
			}
			if t.User.TodayTasksRemindersEnabled && taskDue.After(reminderTime) && taskDue.Before(endOfDay) {
				uts[t.User.ID].dueToday[t.Task.ID] = t.Task
			}
		}
	}

	return uts, nil
}

type userWithTasks struct {
	user     *user.User
	overdue  map[int64]*Task
	dueToday map[int64]*Task
}

// RegisterOverdueReminderCron registers a function which checks once a day for overdue tasks and tasks due today and sends reminders.
func RegisterOverdueReminderCron() {
	if !config.ServiceEnableEmailReminders.GetBool() {
		return
	}

	if !config.MailerEnabled.GetBool() {
		log.Info("Mailer is disabled, not sending overdue per mail")
		return
	}

	err := cron.Schedule("* * * * *", func() {
		s := db.NewSession()
		defer s.Close()

		now := time.Now()
		uts, err := getTasksForDailyReminder(s, now)
		if err != nil {
			log.Errorf("[Daily Tasks Reminder] Could not get tasks for daily reminder: %s", err)
			return
		}

		log.Debugf("[Daily Tasks Reminder] Sending reminders to %d users", len(uts))

		taskIDs := []int64{}
		for _, ut := range uts {
			for _, t := range ut.overdue {
				taskIDs = append(taskIDs, t.ID)
			}
			for _, t := range ut.dueToday {
				taskIDs = append(taskIDs, t.ID)
			}
		}

		projects, err := GetProjectsMapSimpleByTaskIDs(s, taskIDs)
		if err != nil {
			log.Errorf("[Daily Tasks Reminder] Could not get projects for tasks: %s", err)
			return
		}

		for _, ut := range uts {
			n := &DailyTasksReminderNotification{
				User:         ut.user,
				OverdueTasks: ut.overdue,
				DueToday:     ut.dueToday,
				Projects:     projects,
			}

			if len(ut.overdue) == 0 && len(ut.dueToday) == 0 {
				continue
			}

			err = notifications.Notify(ut.user, n)
			if err != nil {
				log.Errorf("[Daily Tasks Reminder] Could not notify user %d: %s", ut.user.ID, err)
				return
			}

			log.Debugf("[Daily Tasks Reminder] Sent reminder email to user %d (overdue: %d, today: %d)", ut.user.ID, len(ut.overdue), len(ut.dueToday))
		}
	})
	if err != nil {
		log.Fatalf("Could not register daily tasks reminder cron: %s", err)
	}
}
