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

	"code.vikunja.io/api/pkg/utils"
	"xorm.io/builder"

	"code.vikunja.io/api/pkg/notifications"

	"code.vikunja.io/api/pkg/db"
	"xorm.io/xorm"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
)

// TaskReminder holds a reminder on a task
type TaskReminder struct {
	ID       int64     `xorm:"bigint autoincr not null unique pk"`
	TaskID   int64     `xorm:"bigint not null INDEX"`
	Reminder time.Time `xorm:"DATETIME not null INDEX 'reminder'"`
	Created  time.Time `xorm:"created not null"`
}

// TableName returns a pretty table name
func (TaskReminder) TableName() string {
	return "task_reminders"
}

type taskUser struct {
	Task *Task      `xorm:"extends"`
	User *user.User `xorm:"extends"`
}

const dbTimeFormat = `2006-01-02 15:04:05`

func getTaskUsersForTasks(s *xorm.Session, taskIDs []int64, cond builder.Cond) (taskUsers []*taskUser, err error) {
	if len(taskIDs) == 0 {
		return
	}

	// Get all creators of tasks
	creators := make(map[int64]*user.User, len(taskIDs))
	err = s.
		Select("users.id, users.username, users.email, users.name, users.timezone").
		Join("LEFT", "tasks", "tasks.created_by_id = users.id").
		In("tasks.id", taskIDs).
		Where(cond).
		GroupBy("tasks.id, users.id, users.username, users.email, users.name, users.timezone").
		Find(&creators)
	if err != nil {
		return
	}

	taskMap := make(map[int64]*Task, len(taskIDs))
	err = s.In("id", taskIDs).Find(&taskMap)
	if err != nil {
		return
	}

	for _, task := range taskMap {
		u, exists := creators[task.CreatedByID]
		if !exists {
			continue
		}

		taskUsers = append(taskUsers, &taskUser{
			Task: taskMap[task.ID],
			User: u,
		})
	}

	var assignees []*TaskAssigneeWithUser
	err = s.Table("task_assignees").
		Select("task_id, users.*").
		In("task_id", taskIDs).
		Join("INNER", "users", "task_assignees.user_id = users.id").
		Where(cond).
		Find(&assignees)
	if err != nil {
		return
	}

	for _, assignee := range assignees {
		taskUsers = append(taskUsers, &taskUser{
			Task: taskMap[assignee.TaskID],
			User: &assignee.User,
		})
	}

	return
}

func getTasksWithRemindersDueAndTheirUsers(s *xorm.Session, now time.Time) (reminderNotifications []*ReminderDueNotification, err error) {
	now = utils.GetTimeWithoutNanoSeconds(now)
	reminderNotifications = []*ReminderDueNotification{}

	nextMinute := now.Add(1 * time.Minute)

	log.Debugf("[Task Reminder Cron] Looking for reminders between %s and %s to send...", now, nextMinute)

	reminders := []*TaskReminder{}
	err = s.
		Join("INNER", "tasks", "tasks.id = task_reminders.task_id").
		// All reminders from -12h to +14h to include all time zones
		Where("reminder >= ? and reminder < ?", now.Add(time.Hour*-12).Format(dbTimeFormat), nextMinute.Add(time.Hour*14).Format(dbTimeFormat)).
		And("tasks.done = false").
		Find(&reminders)
	if err != nil {
		return
	}

	log.Debugf("[Task Reminder Cron] Found %d reminders", len(reminders))

	if len(reminders) == 0 {
		return
	}

	var taskIDs []int64
	for _, r := range reminders {
		taskIDs = append(taskIDs, r.TaskID)
	}

	if len(taskIDs) == 0 {
		return
	}

	usersWithReminders, err := getTaskUsersForTasks(s, taskIDs, builder.Eq{"users.email_reminders_enabled": true})
	if err != nil {
		return
	}

	usersPerTask := make(map[int64][]*taskUser, len(usersWithReminders))
	for _, ur := range usersWithReminders {
		usersPerTask[ur.Task.ID] = append(usersPerTask[ur.Task.ID], ur)
	}

	// Time zone cache per time zone string to avoid parsing the same time zone over and over again
	tzs := make(map[string]*time.Location)
	// Figure out which reminders are actually due in the time zone of the users
	for _, r := range reminders {

		for _, u := range usersPerTask[r.TaskID] {

			if u.User.Timezone == "" {
				u.User.Timezone = config.GetTimeZone().String()
			}

			// I think this will break once there's more reminders than what we can handle in one minute
			tz, exists := tzs[u.User.Timezone]
			if !exists {
				tz, err = time.LoadLocation(u.User.Timezone)
				if err != nil {
					return
				}
				tzs[u.User.Timezone] = tz
			}

			actualReminder := r.Reminder.In(tz)
			if (actualReminder.After(now) && actualReminder.Before(now.Add(time.Minute))) || actualReminder.Equal(now) {
				reminderNotifications = append(reminderNotifications, &ReminderDueNotification{
					User: u.User,
					Task: u.Task,
				})
			}
		}
	}

	return
}

// RegisterReminderCron registers a cron function which runs every minute to check if any reminders are due the
// next minute to send emails.
func RegisterReminderCron() {
	if !config.ServiceEnableEmailReminders.GetBool() {
		return
	}

	if !config.MailerEnabled.GetBool() {
		log.Info("Mailer is disabled, not sending reminders per mail")
		return
	}

	tz := config.GetTimeZone()

	log.Debugf("[Task Reminder Cron] Timezone is %s", tz)

	err := cron.Schedule("* * * * *", func() {
		s := db.NewSession()
		defer s.Close()

		now := time.Now()
		reminders, err := getTasksWithRemindersDueAndTheirUsers(s, now)
		if err != nil {
			log.Errorf("[Task Reminder Cron] Could not get tasks with reminders in the next minute: %s", err)
			return
		}

		if len(reminders) == 0 {
			return
		}

		log.Debugf("[Task Reminder Cron] Sending %d reminders", len(reminders))

		for _, n := range reminders {
			err = notifications.Notify(n.User, n)
			if err != nil {
				log.Errorf("[Task Reminder Cron] Could not notify user %d: %s", n.User.ID, err)
				return
			}

			log.Debugf("[Task Reminder Cron] Sent reminder email for task %d to user %d", n.Task.ID, n.User.ID)
		}
	})
	if err != nil {
		log.Fatalf("Could not register reminder cron: %s", err)
	}
}
