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
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/jobs"
	"code.vikunja.io/api/pkg/user"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm/contexts"
	"xorm.io/xorm/schemas"
)

func TestEnqueueBatchReminderExpansionWindow(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	resetReminderJobs(t)
	s := db.NewSession()
	defer s.Close()
	defer jobs.DiscardPending(s)
	now := time.Date(2026, 11, 12, 20, 5, 0, 0, time.UTC)
	const count = 1001
	for start := 0; start < count; start += 100 {
		rows := make([]*TaskReminder, 0, 100)
		for i := start; i < min(start+100, count); i++ {
			rows = append(rows, &TaskReminder{TaskID: int64(100000 + i), Reminder: now})
		}
		_, err := s.Insert(&rows)
		require.NoError(t, err)
	}
	// A duplicate in a later page must share the same expansion key.
	_, err := s.Insert(&TaskReminder{TaskID: 100000, Reminder: now}, &TaskReminder{TaskID: 99999, Reminder: now.Add(-time.Minute)}, &TaskReminder{TaskID: 99998, Reminder: now.Add(time.Second)})
	require.NoError(t, err)
	window := cron.Window{From: now.Add(-time.Minute), To: now}
	require.NoError(t, enqueueTaskReminders(t.Context(), s, window))
	rows := []jobs.Job{}
	require.NoError(t, s.Where("kind LIKE ?", "reminder.%").Find(&rows))
	require.Len(t, rows, count, "tick must enqueue task/time expansion without recipient discovery")
	for _, row := range rows {
		assert.Equal(t, "reminder.expand", row.Kind)
		assert.Equal(t, "reminder-expansion", row.Queue)
		assert.Equal(t, 5, row.MaxAttempts)
	}
	require.NoError(t, enqueueTaskReminders(t.Context(), s, window))
	total, err := s.Where("kind LIKE ?", "reminder.%").Count(new(jobs.Job))
	require.NoError(t, err)
	require.EqualValues(t, count, total)
	require.NoError(t, s.Rollback())
	jobs.DiscardPending(s)
	total, err = s.Where("kind LIKE ?", "reminder.%").Count(new(jobs.Job))
	require.NoError(t, err)
	require.Zero(t, total, "tick jobs must roll back with its watermark transaction")
}

type reminderChildFailure struct{ enabled atomic.Bool }

func (h *reminderChildFailure) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
	if h.enabled.Load() && strings.HasPrefix(c.SQL, "INSERT INTO jobs") {
		for _, arg := range c.Args {
			if arg == "reminder.notification" {
				return c.Ctx, errors.New("injected notification enqueue failure")
			}
		}
	}
	return c.Ctx, nil
}
func (*reminderChildFailure) AfterProcess(c *contexts.ContextHook) error { return c.Err }

func TestEnqueueBatchReminderExpansionAtomicity(t *testing.T) {
	for _, scenario := range []string{"success", "child failure", "stale claim", "done", "deleted", "muted", "lost access", "archived"} {
		t.Run(scenario, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			resetReminderJobs(t)
			enableReminderMail(t)
			engine, err := db.CreateDBEngine()
			require.NoError(t, err)
			old := config.WebhooksEnabled.GetBool()
			config.WebhooksEnabled.Set(true)
			t.Cleanup(func() { config.WebhooksEnabled.Set(old) })
			s := db.NewSession()
			_, err = s.ID(1).Cols("email_reminders_enabled").Update(&user.User{EmailRemindersEnabled: true})
			require.NoError(t, err)
			at := time.Date(2018, 12, 1, 1, 12, 4, 0, time.UTC)
			require.NoError(t, enqueueReminder(s, "reminder.expand", "reminder-expansion", reminderJob{TaskID: 27, At: at}))
			require.NoError(t, s.Commit())
			jobs.WakePending(s)
			require.NoError(t, s.Close())
			sub := jobs.NewSubscriber(jobs.NewQueue(engine, "expansion-test"), map[string]jobs.Options{"reminder-expansion": {Concurrency: 1, Lease: time.Minute, Timeout: time.Minute}})
			defer sub.Close()
			out, err := sub.Subscribe(t.Context(), "reminder-expansion")
			require.NoError(t, err)
			var msg *message.Message
			select {
			case msg = <-out:
			case <-time.After(5 * time.Second):
				t.Fatal("expansion was not claimed")
			}
			defer msg.Nack()
			parent, ok := jobs.FromMessage(msg)
			require.True(t, ok)
			switch scenario {
			case "child failure":
				hook := &reminderChildFailure{}
				hook.enabled.Store(true)
				engine.AddHook(hook)
				t.Cleanup(func() { hook.enabled.Store(false) })
			case "stale claim":
				_, err = engine.ID(parent.ID).Cols("expires_at").Update(&jobs.Job{ExpiresAt: time.Now().Add(-time.Minute)})
				require.NoError(t, err)
			case "done":
				_, err = engine.ID(27).Cols("done").Update(&Task{Done: true})
				require.NoError(t, err)
			case "deleted":
				_, err = engine.ID(27).Delete(new(Task))
				require.NoError(t, err)
			case "lost access":
				_, err = engine.ID(1).Cols("owner_id").Update(&Project{OwnerID: 2})
				require.NoError(t, err)
			case "archived":
				_, err = engine.ID(1).Cols("is_archived").Update(&Project{IsArchived: true})
				require.NoError(t, err)
			case "muted":
				_, err = engine.Where("entity_type = ? AND entity_id = ? AND user_id = ?", SubscriptionEntityTask, 27, 1).Cols("muted").Update(&Subscription{Muted: true})
				require.NoError(t, err)
			}
			err = handleReminderExpansion(msg)
			if scenario == "child failure" || scenario == "stale claim" {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			rows := []jobs.Job{}
			require.NoError(t, engine.Where("kind LIKE ? AND kind <> ?", "reminder.%", "reminder.expand").Find(&rows))
			if scenario == "success" {
				require.Len(t, rows, 3)
				for _, row := range rows {
					want := map[string]string{"reminder.mail": "bulk-mail", "reminder.notification": "bulk-notifications", "reminder.webhook": "bulk-webhooks"}
					assert.Equal(t, want[row.Kind], row.Queue)
				}
			} else {
				require.Empty(t, rows, "no partially committed children")
			}
			exists, err := engine.ID(parent.ID).Exist(new(jobs.Job))
			require.NoError(t, err)
			assert.Equal(t, scenario == "child failure" || scenario == "stale claim", exists, "parent completion must share the child transaction")
			if scenario == "child failure" || scenario == "stale claim" {
				msg.Nack()
			} else {
				msg.Ack()
			}
		})
	}
}

func TestRunnerPostgresClockSelectsReminderWindow(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	resetReminderJobs(t)
	session := db.NewReadSession()
	defer session.Close()
	e := session.Engine()
	if e.Dialect().URI().DBType != schemas.POSTGRES {
		t.Skip("PostgreSQL database clock integration")
	}
	require.NoError(t, e.Sync(new(cron.State)))
	var clock struct{ Now time.Time }
	_, err := session.SQL("SELECT clock_timestamp() AT TIME ZONE 'UTC' AS now").Get(&clock)
	require.NoError(t, err)
	minute := clock.Now.Truncate(time.Minute)
	due, future := minute.Add(-30*time.Second), minute.Add(2*time.Minute)
	_, err = session.Insert(&TaskReminder{TaskID: 1, Reminder: due}, &TaskReminder{TaskID: 2, Reminder: future})
	require.NoError(t, err)
	name := t.Name()
	_, err = e.Where("job = ?", name).Delete(new(cron.State))
	require.NoError(t, err)
	t.Cleanup(func() { _, err := e.Where("job = ?", name).Delete(new(cron.State)); require.NoError(t, err) })
	runner := cron.NewRunner(e)
	require.NoError(t, runner.Register(name, clock.Now.Add(3*time.Minute)))
	_, err = e.Where("job = ?", name).Cols("last_run_until").Update(&cron.State{LastRunUntil: minute.Add(-time.Minute)})
	require.NoError(t, err)
	ran, err := runner.Tick(t.Context(), name, clock.Now.Add(3*time.Minute), enqueueTaskReminders)
	require.NoError(t, err)
	require.True(t, ran)
	var rows []jobs.Job
	require.NoError(t, e.Where("kind = ?", "reminder.expand").Find(&rows))
	require.Len(t, rows, 1, "fast caller must not enqueue the future reminder")
	assert.Contains(t, rows[0].Payload, `"task_id":1`)
	ahead := minute.Add(10 * time.Minute)
	_, err = e.Where("job = ?", name).Cols("last_run_until").Update(&cron.State{LastRunUntil: ahead})
	require.NoError(t, err)
	_, err = session.Insert(&TaskReminder{TaskID: 3, Reminder: minute.Add(-time.Second)})
	require.NoError(t, err)
	ran, err = runner.Tick(t.Context(), name, ahead.Add(time.Minute), enqueueTaskReminders)
	require.NoError(t, err)
	require.False(t, ran)
	var state cron.State
	_, err = e.Where("job = ?", name).Get(&state)
	require.NoError(t, err)
	assert.True(t, state.LastRunUntil.Equal(ahead))
	count, err := e.Where("kind = ?", "reminder.expand").Count(new(jobs.Job))
	require.NoError(t, err)
	assert.EqualValues(t, 1, count, "already-ahead cursors require explicit repair; reminders behind them are not replayed")
}
