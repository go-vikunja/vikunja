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

func getUndoneOverdueTasks(s *xorm.Session, now time.Time, cond builder.Cond) (usersWithTasks map[int64]*userWithTasks, err error) {
	now = utils.GetTimeWithoutSeconds(now)
	nextMinute := now.Add(1 * time.Minute)

	var tasks []*Task
	err = s.
		Where("due_date is not null AND due_date < ? AND projects.is_archived = false", nextMinute.Add(time.Hour*14).Format(dbTimeFormat)).
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

	users, err := getTaskUsersForTasks(s, taskIDs, cond)
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
		overdueMailTime := time.Date(now.Year(), now.Month(), now.Day(), tm.Hour(), tm.Minute(), 0, 0, tz)
		isTimeForReminder := overdueMailTime.After(now) || overdueMailTime.Equal(now.In(tz))
		wasTimeForReminder := overdueMailTime.Before(nextMinute)
		taskIsOverdueInUserTimezone := overdueMailTime.After(t.Task.DueDate.In(tz))
		if isTimeForReminder && wasTimeForReminder && taskIsOverdueInUserTimezone {
			_, exists := uts[t.User.ID]
			if !exists {
				uts[t.User.ID] = &userWithTasks{
					user:  t.User,
					tasks: make(map[int64]*Task),
				}
			}
			uts[t.User.ID].tasks[t.Task.ID] = t.Task
		}
	}

	return uts, nil
}

type userWithTasks struct {
	user  *user.User
	tasks map[int64]*Task
}

// RegisterOverdueReminderCron registers a function which checks once a day for tasks that are overdue and not done.
// It handles both email and webhook notifications.
func RegisterOverdueReminderCron() {
	emailEnabled := config.ServiceEnableEmailReminders.GetBool() && config.MailerEnabled.GetBool()
	webhookEnabled := config.WebhooksEnabled.GetBool()

	if !emailEnabled && !webhookEnabled {
		if !config.ServiceEnableEmailReminders.GetBool() || !config.MailerEnabled.GetBool() {
			log.Info("Mailer is disabled, not sending overdue reminders per mail")
		}

		if !config.WebhooksEnabled.GetBool() {
			log.Info("Webhooks are disabled, not sending overdue reminders per webhook")
		}

		return
	}

	if config.ServiceEnableEmailReminders.GetBool() && !config.MailerEnabled.GetBool() {
		log.Info("Mailer is disabled, not sending overdue reminders per mail")
	} else if !config.ServiceEnableEmailReminders.GetBool() && config.MailerEnabled.GetBool() {
		log.Info("Mailer reminder is enabled, but mailer is disabled, not sending overdue reminders per mail")
	}

	err := cron.Schedule("* * * * *", func() {
		s := db.NewSession()
		defer s.Close()

		now := time.Now()

		// Build filter based on enabled notification types
		var filters []builder.Cond
		if emailEnabled {
			filters = append(filters, builder.Eq{"users.overdue_tasks_reminders_enabled": true})
		}
		if webhookEnabled {
			// Include users who have webhook settings configured for task.undone.overdue or all
			filters = append(filters, builder.Exists(
				builder.Select("1").
					From("user_webhook_settings").
					Where(builder.And(
						builder.Expr("user_webhook_settings.user_id = users.id"),
						builder.In("user_webhook_settings.notification_type", "task.undone.overdue", "all"),
						builder.Eq{"user_webhook_settings.enabled": true},
						builder.Neq{"user_webhook_settings.target_url": ""},
					)),
			))
		}

		if len(filters) == 0 {
			return
		}

		uts, err := getUndoneOverdueTasks(s, now, builder.Or(filters...))
		if err != nil {
			log.Errorf("[Undone Overdue Tasks Reminder] Could not get undone overdue tasks in the next minute: %s", err)
			return
		}

		if len(uts) == 0 {
			return
		}

		log.Debugf("[Undone Overdue Tasks Reminder] Sending reminders to %d users", len(uts))

		taskIDs := []int64{}
		for _, ut := range uts {
			for _, t := range ut.tasks {
				taskIDs = append(taskIDs, t.ID)
			}
		}

		projects, err := GetProjectsMapSimpleByTaskIDs(s, taskIDs)
		if err != nil {
			log.Errorf("[Undone Overdue Tasks Reminder] Could not get projects for tasks: %s", err)
			return
		}

		for _, ut := range uts {
			var n notifications.Notification = &UndoneTasksOverdueNotification{
				User:     ut.user,
				Tasks:    ut.tasks,
				Projects: projects,
			}

			if len(ut.tasks) == 1 {
				// We know there's only one entry in the map so this is actually O(1) and we can use it to get the
				// first entry without knowing the key of it.
				for _, t := range ut.tasks {
					n = &UndoneTaskOverdueNotification{
						User:    ut.user,
						Task:    t,
						Project: projects[t.ProjectID],
						Overdue: true,
					}
				}
			}

			err = notifications.Notify(ut.user, n)
			if err != nil {
				log.Errorf("[Undone Overdue Tasks Reminder] Could not notify user %d: %s", ut.user.ID, err)
				return
			}

			log.Debugf("[Undone Overdue Tasks Reminder] Sent reminder for %d tasks to user %d", len(ut.tasks), ut.user.ID)
		}
	})
	if err != nil {
		log.Fatalf("Could not register undone overdue tasks reminder cron: %s", err)
	}
}
