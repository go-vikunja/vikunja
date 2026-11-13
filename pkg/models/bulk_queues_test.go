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
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/jobs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBulkQueuesPersistAndIsolateInteractiveJobs(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	engine, err := db.CreateDBEngine()
	require.NoError(t, err)
	clearJobs := func() {
		_, err := engine.Where("id > 0").Delete(new(jobs.Job))
		require.NoError(t, err)
	}
	clearJobs()
	t.Cleanup(clearJobs)
	events.Unfake()
	defer events.Fake()
	events.RegisterDurableListener((&TaskCommentCreatedEvent{}).Name(), "notifications", &SendTaskCommentNotification{})
	for _, event := range []events.Event{&TaskReminderFiredEvent{}, &TaskOverdueEvent{}, &TasksOverdueEvent{}, &TaskCommentCreatedEvent{}} {
		RegisterEventForWebhook(event)
	}
	events.RegisterDurableListener((&WebhookDeliveryEvent{}).Name(), "webhook-delivery", &WebhookDeliveryListener{})
	s := db.NewSession()
	defer s.Close()
	defer events.CleanupPending(s)
	defer jobs.DiscardPending(s)
	for i := range 3 {
		require.NoError(t, enqueueReminder(s, "reminder.notification", "bulk-notifications", reminderJob{TaskID: int64(i + 1), UserID: 1, At: time.Now()}))
		require.NoError(t, events.DispatchOnCommit(s, &TaskReminderFiredEvent{}))
	}
	for _, event := range []events.Event{&TaskOverdueEvent{}, &TasksOverdueEvent{}, &TaskCommentCreatedEvent{}, &WebhookDeliveryEvent{}} {
		require.NoError(t, events.DispatchOnCommit(s, event))
	}
	rows := []jobs.Job{}
	require.NoError(t, s.Find(&rows))
	expected := map[string]string{
		"reminder.notification":                               "bulk-notifications",
		"task.reminder.fired.webhook.listener":                "bulk-webhooks",
		"task.overdue.webhook.listener":                       "bulk-webhooks",
		"tasks.overdue.webhook.listener":                      "bulk-webhooks",
		"task.comment.created.webhook.listener":               "webhooks",
		"task.comment.created.task.comment.notification.send": "notifications",
		"webhook.delivery.webhook.delivery.listener":          "webhook-delivery",
	}
	seen := map[string]bool{}
	for _, row := range rows {
		if want, ok := expected[row.Kind]; ok {
			assert.Equal(t, want, row.Queue, row.Kind)
			seen[row.Kind] = true
		}
	}
	for kind := range expected {
		assert.True(t, seen[kind], kind)
	}
	require.NoError(t, s.Commit())
	jobs.WakePending(s)
	subscriber := jobs.NewSubscriber(jobs.NewQueue(s.Engine(), "bulk-isolation"), map[string]jobs.Options{
		"bulk-notifications": {Concurrency: 2, Lease: time.Minute, Timeout: time.Minute},
		"bulk-webhooks":      {Concurrency: 2, Lease: time.Minute, Timeout: time.Minute},
		"notifications":      {Concurrency: 2, Lease: time.Minute, Timeout: time.Minute},
		"webhooks":           {Concurrency: 2, Lease: time.Minute, Timeout: time.Minute},
	})
	defer subscriber.Close()
	receive := func(ch <-chan *message.Message) *message.Message {
		select {
		case msg := <-ch:
			return msg
		case <-time.After(5 * time.Second):
			t.Fatal("queue did not deliver")
			return nil
		}
	}
	for _, queue := range []string{"bulk-notifications", "bulk-webhooks"} {
		ch, err := subscriber.Subscribe(t.Context(), queue)
		require.NoError(t, err)
		for range 2 {
			msg := receive(ch)
			defer msg.Nack()
		}
		running, err := s.Engine().Where("queue = ? AND state = ?", queue, jobs.StateRunning).Count(new(jobs.Job))
		require.NoError(t, err)
		assert.EqualValues(t, 2, running, "bulk handlers remain active")
	}
	for _, queue := range []string{"notifications", "webhooks"} {
		ch, err := subscriber.Subscribe(t.Context(), queue)
		require.NoError(t, err)
		msg := receive(ch)
		job, ok := jobs.FromMessage(msg)
		require.True(t, ok)
		assert.Equal(t, queue, job.Queue)
		assert.Contains(t, job.Kind, "task.comment.created.")
		msg.Ack()
	}
}

func TestBulkQueuesOverdueProducerRouting(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	engine, err := db.CreateDBEngine()
	require.NoError(t, err)
	clearJobs := func() {
		_, err := engine.Where("id > 0").Delete(new(jobs.Job))
		require.NoError(t, err)
	}
	clearJobs()
	t.Cleanup(clearJobs)
	s := db.NewSession()
	defer s.Close()
	defer jobs.DiscardPending(s)
	_, err = s.Insert(&Webhook{TargetURL: "https://example.com", UserID: 1, CreatedByID: 1, Events: []string{(&TaskOverdueEvent{}).Name()}})
	require.NoError(t, err)
	at := time.Date(2018, 12, 1, 9, 0, 0, 0, time.UTC)
	require.NoError(t, enqueueOverdueReminders(t.Context(), s, cron.Window{From: at.Add(-time.Minute), To: at}))
	rows := []jobs.Job{}
	require.NoError(t, s.Where("kind = ?", "overdue.webhook").Find(&rows))
	require.NotEmpty(t, rows)
	for _, row := range rows {
		assert.Equal(t, "bulk-webhooks", row.Queue)
	}
}
