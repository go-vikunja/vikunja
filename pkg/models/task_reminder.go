// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
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

	"code.vikunja.io/api/pkg/db"
	"xorm.io/xorm"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/mail"
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

func getTaskUsersForTasks(s *xorm.Session, taskIDs []int64) (taskUsers []*taskUser, err error) {
	// Get all creators of tasks
	creators := make(map[int64]*user.User, len(taskIDs))
	err = s.
		Select("users.id, users.username, users.email, users.name").
		Join("LEFT", "tasks", "tasks.created_by_id = users.id").
		In("tasks.id", taskIDs).
		Where("users.email_reminders_enabled = true").
		GroupBy("tasks.id, users.id, users.username, users.email, users.name").
		Find(&creators)
	if err != nil {
		return
	}

	assignees, err := getRawTaskAssigneesForTasks(s, taskIDs)
	if err != nil {
		return
	}

	taskMap := make(map[int64]*Task, len(taskIDs))
	err = s.In("id", taskIDs).Find(&taskMap)
	if err != nil {
		return
	}

	for _, taskID := range taskIDs {
		taskUsers = append(taskUsers, &taskUser{
			Task: taskMap[taskID],
			User: creators[taskMap[taskID].CreatedByID],
		})
	}

	for _, assignee := range assignees {
		if !assignee.EmailRemindersEnabled { // Can't filter that through a query directly since we're using another function
			continue
		}
		taskUsers = append(taskUsers, &taskUser{
			Task: taskMap[assignee.TaskID],
			User: &assignee.User,
		})
	}

	return
}

func getTasksWithRemindersInTheNextMinute(s *xorm.Session, now time.Time) (taskIDs []int64, err error) {

	tz := config.GetTimeZone()
	const dbFormat = `2006-01-02 15:04:05`

	// By default, time.Now() includes nanoseconds which we don't save. That results in getting the wrong dates,
	// so we make sure the time we use to get the reminders don't contain nanoseconds.
	now = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location()).In(tz)
	nextMinute := now.Add(1 * time.Minute)

	log.Debugf("[Task Reminder Cron] Looking for reminders between %s and %s to send...", now, nextMinute)

	reminders := []*TaskReminder{}
	err = s.
		Where("reminder >= ? and reminder < ?", now.Format(dbFormat), nextMinute.Format(dbFormat)).
		Find(&reminders)
	if err != nil {
		return
	}

	log.Debugf("[Task Reminder Cron] Found %d reminders", len(reminders))

	if len(reminders) == 0 {
		return
	}

	// We're sending a reminder to everyone who is assigned to the task or has created it.
	for _, r := range reminders {
		taskIDs = append(taskIDs, r.TaskID)
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

	s := db.NewSession()

	err := cron.Schedule("* * * * *", func() {

		now := time.Now()
		taskIDs, err := getTasksWithRemindersInTheNextMinute(s, now)
		if err != nil {
			log.Errorf("[Task Reminder Cron] Could not get tasks with reminders in the next minute: %s", err)
			return
		}

		users, err := getTaskUsersForTasks(s, taskIDs)
		if err != nil {
			log.Errorf("[Task Reminder Cron] Could not get task users to send them reminders: %s", err)
			return
		}

		log.Debugf("[Task Reminder Cron] Sending reminders to %d users", len(users))

		for _, u := range users {
			data := map[string]interface{}{
				"User": u.User,
				"Task": u.Task,
			}

			mail.SendMailWithTemplate(u.User.Email, `Reminder for "`+u.Task.Title+`"`, "reminder-email", data)
			log.Debugf("[Task Reminder Cron] Sent reminder email for task %d to user %d", u.Task.ID, u.User.ID)
		}
	})
	if err != nil {
		log.Fatalf("Could not register reminder cron: %s", err)
	}
}
