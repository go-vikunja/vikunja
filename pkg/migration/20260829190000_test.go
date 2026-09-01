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

	"github.com/stretchr/testify/require"
)

// Deliberately without the unique(entity_user) tag the model carries, so the
// duplicates below can be inserted in the first place.
type subscriptionsStub20260829190000 struct {
	ID         int64 `xorm:"autoincr not null unique pk"`
	EntityType int64 `xorm:"index not null"`
	EntityID   int64 `xorm:"bigint index not null"`
	UserID     int64 `xorm:"bigint index not null"`
}

func (subscriptionsStub20260829190000) TableName() string {
	return "subscriptions"
}

func TestAddUniqueSubscriptionIndex20260829190000(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	tables := []interface{}{subscriptionsStub20260829190000{}}
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(tables...))
	})
	require.NoError(t, x.DropTables(tables...))
	require.NoError(t, x.Sync2(tables...))

	_, err = x.Insert(
		// duplicates
		&subscriptionsStub20260829190000{ID: 1, EntityType: 3, EntityID: 10, UserID: 1},
		&subscriptionsStub20260829190000{ID: 2, EntityType: 3, EntityID: 10, UserID: 1},
		&subscriptionsStub20260829190000{ID: 3, EntityType: 3, EntityID: 10, UserID: 1},
		&subscriptionsStub20260829190000{ID: 4, EntityType: 2, EntityID: 10, UserID: 1},
		&subscriptionsStub20260829190000{ID: 5, EntityType: 2, EntityID: 10, UserID: 1},
		// distinct tuples
		&subscriptionsStub20260829190000{ID: 6, EntityType: 3, EntityID: 10, UserID: 2},
		&subscriptionsStub20260829190000{ID: 7, EntityType: 3, EntityID: 11, UserID: 1},
		&subscriptionsStub20260829190000{ID: 8, EntityType: 2, EntityID: 11, UserID: 2},
	)
	require.NoError(t, err)

	require.NoError(t, addUniqueSubscriptionIndex20260829190000(x))

	remaining := []*subscriptionsStub20260829190000{}
	require.NoError(t, x.OrderBy("id").Find(&remaining))
	require.Equal(t, []*subscriptionsStub20260829190000{
		{ID: 1, EntityType: 3, EntityID: 10, UserID: 1},
		{ID: 4, EntityType: 2, EntityID: 10, UserID: 1},
		{ID: 6, EntityType: 3, EntityID: 10, UserID: 2},
		{ID: 7, EntityType: 3, EntityID: 11, UserID: 1},
		{ID: 8, EntityType: 2, EntityID: 11, UserID: 2},
	}, remaining)

	_, err = x.Insert(&subscriptionsStub20260829190000{ID: 100, EntityType: 3, EntityID: 10, UserID: 1})
	require.Error(t, err, "the unique index should reject a second row for the same tuple")

	// rerunning must not trip over the index it just created
	require.NoError(t, addUniqueSubscriptionIndex20260829190000(x))
}
