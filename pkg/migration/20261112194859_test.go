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

package migration

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"github.com/stretchr/testify/require"
)

func TestMoveWebhookDelivery20261112194859(t *testing.T) {
	engine, err := db.CreateTestEngine()
	require.NoError(t, err)
	table := jobs20261112124932{}
	require.NoError(t, engine.DropTables(table))
	t.Cleanup(func() { require.NoError(t, engine.DropTables(table)) })
	require.NoError(t, engine.Sync2(table))
	rows := []jobs20261112124932{
		{Queue: "webhooks", Kind: "webhook.delivery.webhook.delivery.listener", State: "available"},
		{Queue: "webhooks", Kind: "webhook.delivery.webhook.delivery.listener", State: "running"},
		{Queue: "webhooks", Kind: "webhook.delivery.webhook.delivery.listener", State: "failed"},
		{Queue: "webhooks", Kind: "webhook.delivery.webhook.delivery.listener", State: "completed"},
		{Queue: "webhooks", Kind: "task.created.webhook.listener", State: "available"},
		{Queue: "different", Kind: "webhook.delivery.webhook.delivery.listener", State: "available"},
	}
	for i := range rows {
		key := rows[i].Queue + rows[i].Kind + rows[i].State
		rows[i].UniqueKey = &key
		rows[i].Payload = "{}"
		rows[i].Version = 1
		rows[i].Attempt = 2
		rows[i].MaxAttempts = 5
		rows[i].RunAt = time.Now().UTC().Truncate(time.Second)
		_, err = engine.Insert(&rows[i])
		require.NoError(t, err)
	}
	require.NoError(t, moveWebhookDelivery20261112194859(engine))
	require.NoError(t, moveWebhookDelivery20261112194859(engine))
	for i, before := range rows {
		var after jobs20261112124932
		found, err := engine.ID(before.ID).Get(&after)
		require.NoError(t, err)
		require.True(t, found)
		if i < 3 {
			before.Queue = "webhook-delivery"
		}
		after.RunAt = after.RunAt.UTC()
		require.Equal(t, before, after, "only the queue of pending deliveries may change")
	}
	duplicate := rows[0]
	duplicate.ID = 0
	_, err = engine.Insert(&duplicate)
	require.Error(t, err, "migration must preserve live-key uniqueness")
}
