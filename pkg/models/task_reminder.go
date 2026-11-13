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
	"fmt"
	"time"

	"code.vikunja.io/api/pkg/cron"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// ReminderRelation represents the date attribute of the task which a period based reminder relates to
type ReminderRelation string

// All valid ReminderRelations
const (
	ReminderRelationDueDate   ReminderRelation = `due_date`
	ReminderRelationStartDate ReminderRelation = `start_date`
	ReminderRelationEndDate   ReminderRelation = `end_date`
)

// TaskReminder holds a reminder on a task.
// If RelativeTo and the assciated date field are defined, then the attribute Reminder will be computed.
// If RelativeTo is missing, than Reminder must be given.
type TaskReminder struct {
	ID     int64 `xorm:"bigint autoincr not null unique pk" json:"-"`
	TaskID int64 `xorm:"bigint not null INDEX" json:"-"`
	// The absolute time when the user wants to be reminded of the task.
	Reminder time.Time `xorm:"DATETIME not null INDEX 'reminder'" json:"reminder"`
	Created  time.Time `xorm:"created not null" json:"-"`
	// A period in seconds relative to another date argument. Negative values mean the reminder triggers before the date. Default: 0, tiggers when RelativeTo is due.
	RelativePeriod int64 `xorm:"bigint null" json:"relative_period"`
	// The name of the date field to which the relative period refers to.
	RelativeTo ReminderRelation `xorm:"varchar(50) null" json:"relative_to"`
}

// TableName returns a pretty table name
func (TaskReminder) TableName() string {
	return "task_reminders"
}

type taskUser struct {
	Task       *Task      `xorm:"extends"`
	User       *user.User `xorm:"extends"`
	IsAssignee bool
}

//nolint:gocyclo
func getTaskUsersForTasks(s *xorm.Session, taskIDs []int64, cond builder.Cond) (taskUsers []*taskUser, err error) {
	if len(taskIDs) == 0 {
		return
	}

	taskUsers = []*taskUser{}
	taskMap := make(map[int64]*Task, len(taskIDs))
	err = s.In("id", taskIDs).Find(&taskMap)
	if err != nil {
		return
	}

	notBot := builder.Or(
		builder.IsNull{"users.bot_owner_id"},
		builder.Eq{"users.bot_owner_id": 0},
	)

	subscriptions, err := GetSubscriptionsForEntities(s, SubscriptionEntityTask, taskIDs)
	if err != nil {
		return nil, err
	}

	subscriberIDs := []int64{}
	for _, subs := range subscriptions {
		for _, sub := range subs {
			subscriberIDs = append(subscriberIDs, sub.UserID)
		}
	}

	if len(subscriberIDs) == 0 {
		return
	}

	subscriberCond := []builder.Cond{
		builder.In("id", subscriberIDs),
		builder.Eq{"users.status": user.StatusActive},
		notBot,
	}
	if cond != nil {
		subscriberCond = append(subscriberCond, cond)
	}

	subscribers, err := user.GetUsersByCond(s, builder.And(subscriberCond...))
	if err != nil {
		return nil, err
	}

	assigneeRows := []*TaskAssginee{}
	err = s.In("task_id", taskIDs).Find(&assigneeRows)
	if err != nil {
		return nil, err
	}
	assignees := make(map[int64]map[int64]struct{})
	for _, a := range assigneeRows {
		if assignees[a.TaskID] == nil {
			assignees[a.TaskID] = make(map[int64]struct{})
		}
		assignees[a.TaskID][a.UserID] = struct{}{}
	}

	for taskID, subs := range subscriptions {
		for _, sub := range subs {
			u, has := subscribers[sub.UserID]
			if !has {
				continue
			}
			task, has := taskMap[taskID]
			if !has {
				continue
			}
			_, isAssignee := assignees[taskID][sub.UserID]
			taskUsers = append(taskUsers, &taskUser{Task: task, User: u, IsAssignee: isAssignee})
		}
	}

	return
}

func getTaskRemindersInWindow(s *xorm.Session, window cron.Window, cond builder.Cond, taskIDsFilter ...int64) ([]*ReminderDueNotification, error) {
	reminders := []*TaskReminder{}
	query := s.Join("INNER", "tasks", "tasks.id = task_reminders.task_id").
		Join("INNER", "projects", "projects.id = tasks.project_id").
		Where(builder.Gt{"reminder": window.From}).And(builder.Lte{"reminder": window.To}).
		And(builder.Eq{"tasks.done": false}).And(builder.And(builder.IsNull{"tasks.deleted_at"}, builder.IsNull{"projects.deleted_at"}))
	if len(taskIDsFilter) > 0 {
		query.In("task_reminders.task_id", taskIDsFilter)
	}
	err := query.Find(&reminders)
	if err != nil {
		return nil, err
	}
	taskIDs := make([]int64, 0, len(reminders))
	for _, reminder := range reminders {
		taskIDs = append(taskIDs, reminder.TaskID)
	}
	users, err := getTaskUsersForTasks(s, taskIDs, cond)
	if err != nil {
		return nil, err
	}
	projects, err := GetProjectsMapSimpleByTaskIDs(s, taskIDs)
	if err != nil {
		return nil, err
	}
	perTask := make(map[int64][]*taskUser)
	for _, u := range users {
		perTask[u.Task.ID] = append(perTask[u.Task.ID], u)
	}
	seen := make(map[string]bool)
	result := make([]*ReminderDueNotification, 0)
	for _, reminder := range reminders {
		for _, u := range perTask[reminder.TaskID] {
			key := fmt.Sprintf("%d/%d/%d", reminder.TaskID, u.User.ID, reminder.Reminder.Unix())
			project := projects[u.Task.ProjectID]
			if seen[key] || project == nil || project.IsArchived {
				continue
			}
			seen[key] = true
			u.Task.setIdentifier(project)
			result = append(result, &ReminderDueNotification{User: u.User, Task: u.Task, Project: project, TaskReminder: reminder})
		}
	}
	return result, nil
}

func RegisterReminderCron() {
	if err := cron.Register("task-reminders", "* * * * *", enqueueTaskReminders); err != nil {
		log.Fatalf("Register task reminders: %v", err)
	}
}
