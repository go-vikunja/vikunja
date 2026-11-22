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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookDeliveryListener_SkipsDisabledQueuedWebhook(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	var requests atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	previousClient := webhookClient
	webhookClient = ts.Client()
	t.Cleanup(func() { webhookClient = previousClient })

	s := db.NewSession()
	defer s.Close()
	webhook := &Webhook{
		ProjectID:   1,
		CreatedByID: 1,
		TargetURL:   ts.URL,
		Events:      []string{"task.updated"},
	}
	_, err := s.Insert(webhook)
	require.NoError(t, err)
	payload, err := json.Marshal(&WebhookDeliveryEvent{
		WebhookID: webhook.ID,
		Payload:   &WebhookPayload{EventName: "task.updated"},
	})
	require.NoError(t, err)
	_, err = s.ID(webhook.ID).Cols("is_disabled").Update(&Webhook{IsDisabled: true})
	require.NoError(t, err)
	require.NoError(t, s.Commit())

	require.NoError(t, (&WebhookDeliveryListener{}).Handle(message.NewMessage("disabled", payload)))
	assert.Zero(t, requests.Load())
}

func TestWebhookListener_SkipsDisabledDeletedProject(t *testing.T) {
	f := setupAccessLoss(t)
	_, err := f.s.ID(f.child.ID).Delete(&Project{})
	require.NoError(t, err)
	_, err = f.s.ID(f.webhookC).Cols("is_disabled", "events").Update(&Webhook{
		IsDisabled: true,
		Events:     []string{"project.deleted"},
	})
	require.NoError(t, err)
	active := &Webhook{
		ProjectID:   f.child.ID,
		CreatedByID: f.owner.ID,
		TargetURL:   "https://example.com/active",
		Events:      []string{"project.deleted"},
	}
	_, err = f.s.Insert(active)
	require.NoError(t, err)
	require.NoError(t, f.s.Commit())

	matches, err := projectWebhooksForEvent(f.s, f.child.ID, "project.deleted", nil)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	assert.Equal(t, active.ID, matches[0].ID)
}

func TestWebhookListener_TeamReinstatementDoesNotEnable(t *testing.T) {
	f := setupAccessLoss(t)
	team := f.shareViaTeam(t)
	require.NoError(t, removeUserFromTeamsByIDs(f.s, f.member, []int64{team.ID}))
	require.NoError(t, (&TeamMember{
		TeamID:   team.ID,
		Username: f.member.Username,
	}).Create(f.s, f.owner))
	require.NoError(t, f.s.Commit())

	events.ClearDispatchedEvents()
	events.TestListener(t, &TaskUpdatedEvent{
		Task: f.task,
		Doer: f.owner,
	}, &WebhookListener{EventName: "task.updated"})
	assert.Empty(t, events.GetDispatchedEvents((&WebhookDeliveryEvent{}).Name()))
	db.AssertExists(t, "webhooks", map[string]interface{}{
		"id":            f.webhookC,
		"is_disabled":   true,
		"created_by_id": f.member.ID,
	}, false)
}

func TestWebhookListener_SkipsDisabledUserWebhook(t *testing.T) {
	f := setupAccessLoss(t)
	RegisterUserDirectedEventForWebhook(&TaskReminderFiredEvent{})
	disabled := &Webhook{
		UserID:      f.owner.ID,
		CreatedByID: f.owner.ID,
		TargetURL:   "https://example.com/disabled",
		Events:      []string{"task.reminder.fired"},
		IsDisabled:  true,
	}
	_, err := f.s.Insert(disabled)
	require.NoError(t, err)
	active := &Webhook{
		UserID:      f.owner.ID,
		CreatedByID: f.owner.ID,
		TargetURL:   "https://example.com/active",
		Events:      []string{"task.reminder.fired"},
	}
	_, err = f.s.Insert(active)
	require.NoError(t, err)
	require.NoError(t, f.s.Commit())
	events.ClearDispatchedEvents()
	events.TestListener(t, &TaskReminderFiredEvent{
		Task:    f.task,
		User:    f.owner,
		Project: f.child,
	}, &WebhookListener{EventName: "task.reminder.fired"})
	deliveries := webhookDeliveriesByID(t)
	assert.Contains(t, deliveries, active.ID)
	assert.NotContains(t, deliveries, disabled.ID)
}
