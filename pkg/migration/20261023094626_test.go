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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/require"
)

type webhooksStub20261023094626 struct {
	BasicAuthUser     string `xorm:"null"`
	BasicAuthPassword string `xorm:"null"`
	UserID            int64  `xorm:"bigint null index"`
}

func (webhooksStub20261023094626) TableName() string {
	return "webhooks"
}

func TestRepairWebhooksTable20261023094626(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, x.DropTables("webhooks"))
		require.NoError(t, x.Sync2(&models.Webhook{}))
	})

	createStub := func(t *testing.T) {
		require.NoError(t, x.DropTables("webhooks"))
		require.NoError(t, x.Sync2(webhooksStub20261023094626{}))
	}

	t.Run("recreates stub table", func(t *testing.T) {
		createStub(t)

		require.NoError(t, repairWebhooksTable20261023094626(x))

		_, err := x.Insert(&models.Webhook{TargetURL: "https://example.com", Events: []string{"task.created"}, UserID: 1, CreatedByID: 1})
		require.NoError(t, err)
		webhooks := []*models.Webhook{}
		require.NoError(t, x.Cols("events").Find(&webhooks))
		require.Len(t, webhooks, 1)
		require.Equal(t, []string{"task.created"}, webhooks[0].Events)
	})

	t.Run("leaves healthy table untouched", func(t *testing.T) {
		require.NoError(t, x.DropTables("webhooks"))
		require.NoError(t, x.Sync2(&models.Webhook{}))
		w := &models.Webhook{TargetURL: "https://example.com", Events: []string{"task.created"}, ProjectID: 1, CreatedByID: 1}
		_, err := x.Insert(w)
		require.NoError(t, err)

		require.NoError(t, repairWebhooksTable20261023094626(x))

		has, err := x.ID(w.ID).Exist(&models.Webhook{})
		require.NoError(t, err)
		require.True(t, has)
	})

	t.Run("refuses to drop stub table with rows", func(t *testing.T) {
		createStub(t)
		_, err := x.Insert(&webhooksStub20261023094626{UserID: 1})
		require.NoError(t, err)

		require.Error(t, repairWebhooksTable20261023094626(x))

		exists, err := x.IsTableExist("webhooks")
		require.NoError(t, err)
		require.True(t, exists)
	})
}
