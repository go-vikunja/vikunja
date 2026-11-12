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
	"context"
	"encoding/json"
	"fmt"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/jobs"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/notifications"
	"github.com/ThreeDotsLabs/watermill/message"
	"xorm.io/builder"
	"xorm.io/xorm"
)

type reminderJob struct {
	TaskID int64     `json:"task_id"`
	UserID int64     `json:"user_id"`
	At     time.Time `json:"at"`
}

func init() {
	jobs.Register("reminder.mail", "bulk-mail", handleTaskReminder)
	jobs.Register("reminder.notification", "notifications", handleTaskReminder)
	jobs.Register("reminder.webhook", "webhooks", handleTaskReminder)
	jobs.Register("overdue.mail", "bulk-mail", handleOverdueReminder)
	jobs.Register("overdue.webhook", "webhooks", handleOverdueReminder)
}

func reminderMailEnabled() bool {
	return config.ServiceEnableEmailReminders.GetBool() && config.MailerEnabled.GetBool()
}

func enqueueReminder(s *xorm.Session, kind, queue string, payload reminderJob) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%d/%d/%d", payload.TaskID, payload.UserID, payload.At.Unix())
	return jobs.Enqueue(s, jobs.Request{Queue: queue, Kind: kind, Payload: data, UniqueKey: key})
}

func enqueueTaskReminders(_ context.Context, s *xorm.Session, window cron.Window) error {
	reminders, err := getTaskRemindersInWindow(s, window, nil)
	if err != nil {
		return err
	}
	for _, reminder := range reminders {
		itemErr, txErr := jobs.Isolate(s, func() error {
			payload := reminderJob{TaskID: reminder.Task.ID, UserID: reminder.User.ID, At: reminder.TaskReminder.Reminder}
			if reminderMailEnabled() && reminder.User.EmailRemindersEnabled {
				if err := enqueueReminder(s, "reminder.mail", "bulk-mail", payload); err != nil {
					return err
				}
			}
			if err := enqueueReminder(s, "reminder.notification", "notifications", payload); err != nil {
				return err
			}
			if config.WebhooksEnabled.GetBool() || hasNonWebhookListener((&TaskReminderFiredEvent{}).Name()) {
				return enqueueReminder(s, "reminder.webhook", "webhooks", payload)
			}
			return nil
		})
		if txErr != nil {
			return txErr
		}
		if itemErr != nil {
			log.Errorf("Enqueue task reminder for task %d/user %d: %v", reminder.Task.ID, reminder.User.ID, itemErr)
		}
	}
	return nil
}

func enqueueOverdueReminders(_ context.Context, s *xorm.Session, window cron.Window) error {
	single, batch, err := overdueEventsWithListeners(s)
	if err != nil {
		return err
	}
	if !reminderMailEnabled() && !single && !batch {
		return nil
	}
	var cond builder.Cond
	if !single && !batch {
		cond = builder.Eq{"users.overdue_tasks_reminders_enabled": true}
	}
	users, err := getOverdueTasksInWindow(s, window, cond)
	if err != nil {
		return err
	}
	for _, ut := range users {
		itemErr, txErr := jobs.Isolate(s, func() error {
			payload := reminderJob{UserID: ut.user.ID, At: ut.slot}
			if reminderMailEnabled() && ut.user.OverdueTasksRemindersEnabled {
				if err := enqueueReminder(s, "overdue.mail", "bulk-mail", payload); err != nil {
					return err
				}
			}
			if single || batch {
				return enqueueReminder(s, "overdue.webhook", "webhooks", payload)
			}
			return nil
		})
		if txErr != nil {
			return txErr
		}
		if itemErr != nil {
			log.Errorf("Enqueue overdue reminder for user %d: %v", ut.user.ID, itemErr)
		}
	}
	return nil
}

func handleTaskReminder(msg *message.Message) error {
	payload := reminderJob{}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return err
	}
	s := jobs.NewSession(msg)
	defer s.Close()
	defer events.CleanupPending(s)
	reminders, err := getTaskRemindersInWindow(s, cron.Window{From: payload.At.Add(-time.Second), To: payload.At}, builder.Eq{"users.id": payload.UserID}, payload.TaskID)
	if err != nil {
		return err
	}
	job, _ := jobs.FromMessage(msg)
	for _, reminder := range reminders {
		if reminder.Task.ID != payload.TaskID || reminder.User.ID != payload.UserID || !reminder.TaskReminder.Reminder.Equal(payload.At) {
			continue
		}
		switch job.Kind {
		case "reminder.mail":
			if !reminderMailEnabled() || !reminder.User.EmailRemindersEnabled {
				return nil
			}
			s.Close()
			return notifications.SendMailNow(msg.Context(), reminder.User, reminder)
		case "reminder.notification":
			if err := notifications.NotifyDB(reminder.User, reminder, s); err != nil {
				return err
			}
		case "reminder.webhook":
			if !config.WebhooksEnabled.GetBool() && !hasNonWebhookListener((&TaskReminderFiredEvent{}).Name()) {
				return nil
			}
			if err := events.DispatchOnCommit(s, &TaskReminderFiredEvent{Task: reminder.Task, User: reminder.User, Project: reminder.Project, Reminder: reminder.TaskReminder}); err != nil {
				return err
			}
		}
		if err := jobs.Commit(s, msg); err != nil {
			return err
		}
		events.DispatchPending(msg.Context(), s)
		return nil
	}
	return nil
}

func handleOverdueReminder(msg *message.Message) error {
	payload := reminderJob{}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return err
	}
	s := jobs.NewSession(msg)
	defer s.Close()
	defer events.CleanupPending(s)
	users, err := getOverdueTasksInWindow(s, cron.Window{From: payload.At.Add(-time.Second), To: payload.At}, builder.Eq{"users.id": payload.UserID})
	if err != nil {
		return err
	}
	ut := users[payload.UserID]
	if ut == nil {
		return nil
	}
	tasks := ut.allTasks()
	ids := make([]int64, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.ID)
	}
	projects, err := GetProjectsMapSimpleByTaskIDs(s, ids)
	if err != nil {
		return err
	}
	job, _ := jobs.FromMessage(msg)
	if job.Kind == "overdue.mail" {
		if !reminderMailEnabled() || !ut.user.OverdueTasksRemindersEnabled {
			return nil
		}
		var n notifications.Notification = &UndoneTasksOverdueNotification{User: ut.user, Assigned: ut.assigned, Followed: ut.followed, Projects: projects}
		if len(tasks) == 1 {
			n = &UndoneTaskOverdueNotification{User: ut.user, Task: tasks[0], Project: projects[tasks[0].ProjectID]}
		}
		s.Close()
		return notifications.SendMailNow(msg.Context(), ut.user, n)
	}
	single, batch, err := overdueEventsWithListeners(s)
	if err != nil {
		return err
	}
	if err := dispatchOverdueReminders(s, ut, single, batch); err != nil {
		return err
	}
	if err := jobs.Commit(s, msg); err != nil {
		return err
	}
	events.DispatchPending(msg.Context(), s)
	return nil
}

func dispatchOverdueReminders(s *xorm.Session, ut *userWithTasks, single, batch bool) error {
	tasks := ut.allTasks()
	ids := make([]int64, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.ID)
	}
	projects, err := GetProjectsMapSimpleByTaskIDs(s, ids)
	if err != nil {
		return err
	}
	if single {
		for _, task := range tasks {
			if err := events.DispatchOnCommit(s, &TaskOverdueEvent{Task: task, User: ut.user, Project: projects[task.ProjectID]}); err != nil {
				return err
			}
		}
	}
	if batch {
		return events.DispatchOnCommit(s, &TasksOverdueEvent{Tasks: tasks, User: ut.user, Projects: projects})
	}
	return nil
}
