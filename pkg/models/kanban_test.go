// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBucket_ReadAll(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	testuser := &user.User{ID: 1}
	b := &Bucket{ListID: 1}
	bucketsInterface, _, _, err := b.ReadAll(testuser, "", 0, 0)
	assert.NoError(t, err)

	buckets, is := bucketsInterface.([]*Bucket)
	assert.True(t, is)

	// Assert that we have a user for each bucket
	assert.Equal(t, testuser.ID, buckets[0].CreatedBy.ID)
	assert.Equal(t, testuser.ID, buckets[1].CreatedBy.ID)
	assert.Equal(t, testuser.ID, buckets[2].CreatedBy.ID)
	assert.Equal(t, testuser.ID, buckets[3].CreatedBy.ID)

	// Assert our three test buckets + one for all tasks without a bucket
	assert.Len(t, buckets, 4)

	// Assert all tasks are in the right bucket
	assert.Len(t, buckets[0].Tasks, 10)
	assert.Len(t, buckets[1].Tasks, 2)
	assert.Len(t, buckets[2].Tasks, 3)
	assert.Len(t, buckets[3].Tasks, 3)

	// Assert we have bucket 0, 1, 2, 3 but not 4 (that belongs to a different list)
	assert.Equal(t, int64(1), buckets[1].ID)
	assert.Equal(t, int64(2), buckets[2].ID)
	assert.Equal(t, int64(3), buckets[3].ID)

	// Kinda assert all tasks are in the right buckets
	assert.Equal(t, int64(0), buckets[0].Tasks[0].BucketID)
	assert.Equal(t, int64(1), buckets[1].Tasks[0].BucketID)
	assert.Equal(t, int64(1), buckets[1].Tasks[1].BucketID)
	assert.Equal(t, int64(2), buckets[2].Tasks[0].BucketID)
	assert.Equal(t, int64(2), buckets[2].Tasks[1].BucketID)
	assert.Equal(t, int64(2), buckets[2].Tasks[2].BucketID)
	assert.Equal(t, int64(3), buckets[3].Tasks[0].BucketID)
	assert.Equal(t, int64(3), buckets[3].Tasks[1].BucketID)
	assert.Equal(t, int64(3), buckets[3].Tasks[2].BucketID)
}
