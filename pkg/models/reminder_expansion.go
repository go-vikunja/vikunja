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
	"errors"
	"fmt"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/jobs"
	"github.com/ThreeDotsLabs/watermill/message"
	"xorm.io/builder"
	"xorm.io/xorm"
)

type reminderExpansion struct {
	TaskID int64     `json:"task_id"`
	At     time.Time `json:"at"`
}

func init() {
	jobs.Register("reminder.expand", "reminder-expansion", handleReminderExpansion)
}

func enqueueTaskReminders(ctx context.Context, s *xorm.Session, window cron.Window) error {
	last := TaskReminder{}
	found, err := s.Cols("id").Where(builder.Gt{"reminder": window.From}).And(builder.Lte{"reminder": window.To}).Desc("id").Get(&last)
	if err != nil || !found {
		return err
	}
	const pageSize = 500
	var cursor int64
	for cursor < last.ID {
		if err := ctx.Err(); err != nil {
			return err
		}
		rows := []TaskReminder{}
		err = s.Cols("id", "task_id", "reminder").Where("id > ? AND id <= ?", cursor, last.ID).
			And(builder.Gt{"reminder": window.From}).And(builder.Lte{"reminder": window.To}).OrderBy("id").Limit(pageSize).Find(&rows)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}
		requests := make([]jobs.Request, 0, len(rows))
		for _, row := range rows {
			payload, err := json.Marshal(reminderExpansion{TaskID: row.TaskID, At: row.Reminder})
			if err != nil {
				return err
			}
			requests = append(requests, jobs.Request{Queue: "reminder-expansion", Kind: "reminder.expand", Payload: payload, UniqueKey: fmt.Sprintf("%d/%d", row.TaskID, row.Reminder.Unix())})
		}
		if err := jobs.EnqueueBatch(s, requests); err != nil {
			return err
		}
		cursor = rows[len(rows)-1].ID
	}
	return nil
}

func handleReminderExpansion(msg *message.Message) error {
	payload := reminderExpansion{}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return err
	}
	if payload.TaskID <= 0 || payload.At.IsZero() {
		return errors.New("reminder expansion requires task and timestamp")
	}
	s := jobs.NewSession(msg)
	defer s.Close()
	defer jobs.DiscardPending(s)
	reminders, err := getTaskRemindersInWindow(s, cron.Window{From: payload.At.Add(-time.Second), To: payload.At}, nil, payload.TaskID)
	if err != nil {
		return err
	}
	for _, reminder := range reminders {
		if !reminder.TaskReminder.Reminder.Equal(payload.At) {
			continue
		}
		child := reminderJob{TaskID: reminder.Task.ID, UserID: reminder.User.ID, At: payload.At}
		if reminderMailEnabled() && reminder.User.EmailRemindersEnabled {
			if err := enqueueReminder(s, "reminder.mail", "bulk-mail", child); err != nil {
				return err
			}
		}
		if err := enqueueReminder(s, "reminder.notification", "notifications", child); err != nil {
			return err
		}
		if config.WebhooksEnabled.GetBool() || hasNonWebhookListener((&TaskReminderFiredEvent{}).Name()) {
			if err := enqueueReminder(s, "reminder.webhook", "webhooks", child); err != nil {
				return err
			}
		}
	}
	return jobs.Commit(s, msg)
}
