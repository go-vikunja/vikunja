// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"time"

	"code.vikunja.io/api/pkg/user"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/utils"
	"xorm.io/builder"
	"xorm.io/xorm"
)

func getUndoneOverdueTasks(s *xorm.Session, now time.Time) (taskIDs []int64, err error) {
	now = utils.GetTimeWithoutNanoSeconds(now)

	var tasks []*Task
	err = s.
		Where("due_date is not null and due_date < ?", now.Format(dbTimeFormat)).
		And("done = false").
		Find(&tasks)
	if err != nil {
		return
	}

	for _, task := range tasks {
		taskIDs = append(taskIDs, task.ID)
	}

	return
}

type userWithTasks struct {
	user  *user.User
	tasks []*Task
}

// RegisterOverdueReminderCron registers a function which checks once a day for tasks that are overdue and not done.
func RegisterOverdueReminderCron() {
	if !config.ServiceEnableEmailReminders.GetBool() {
		return
	}

	if !config.MailerEnabled.GetBool() {
		log.Info("Mailer is disabled, not sending overdue per mail")
		return
	}

	err := cron.Schedule("0 8 * * *", func() {
		s := db.NewSession()
		defer s.Close()

		now := time.Now()
		taskIDs, err := getUndoneOverdueTasks(s, now)
		if err != nil {
			log.Errorf("[Undone Overdue Tasks Reminder] Could not get tasks with reminders in the next minute: %s", err)
			return
		}

		users, err := getTaskUsersForTasks(s, taskIDs, builder.Eq{"users.overdue_tasks_reminders_enabled": true})
		if err != nil {
			log.Errorf("[Undone Overdue Tasks Reminder] Could not get task users to send them reminders: %s", err)
			return
		}

		uts := make(map[int64]*userWithTasks)
		for _, t := range users {
			_, exists := uts[t.User.ID]
			if !exists {
				uts[t.User.ID] = &userWithTasks{
					user:  t.User,
					tasks: []*Task{},
				}
			}
			uts[t.User.ID].tasks = append(uts[t.User.ID].tasks, t.Task)
		}

		log.Debugf("[Undone Overdue Tasks Reminder] Sending reminders to %d users", len(users))

		for _, ut := range uts {
			var n notifications.Notification = &UndoneTasksOverdueNotification{
				User:  ut.user,
				Tasks: ut.tasks,
			}

			if len(ut.tasks) == 1 {
				n = &UndoneTaskOverdueNotification{
					User: ut.user,
					Task: ut.tasks[0],
				}
			}

			err = notifications.Notify(ut.user, n)
			if err != nil {
				log.Errorf("[Undone Overdue Tasks Reminder] Could not notify user %d: %s", ut.user.ID, err)
				return
			}

			log.Debugf("[Undone Overdue Tasks Reminder] Sent reminder email for %d tasks to user %d", len(ut.tasks), ut.user.ID)
			continue
		}
	})
	if err != nil {
		log.Fatalf("Could not register undone overdue tasks reminder cron: %s", err)
	}
}
