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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskBucket_Update(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("full bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		tb := &TaskBucket{
			TaskID:        1,
			BucketID:      2, // Bucket 2 already has 3 tasks and a limit of 3
			ProjectViewID: 4,
			ProjectID:     1, // In actual web requests set via the url
		}

		err := tb.Update(s, u)
		require.Error(t, err)
		assert.True(t, IsErrBucketLimitExceeded(err))
	})
	t.Run("full bucket but not changing the bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		tb := &TaskBucket{
			TaskID:        4,
			BucketID:      2, // Bucket 2 already has 3 tasks and a limit of 3
			ProjectViewID: 4,
			ProjectID:     1, // In actual web requests set via the url
		}
		err := tb.Update(s, u)
		require.NoError(t, err)
	})
	t.Run("bucket on other project view", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		tb := &TaskBucket{
			TaskID:        1,
			BucketID:      4, // Bucket 4 belongs to project 2
			ProjectViewID: 4,
			ProjectID:     1, // In actual web requests set via the url
		}
		err := tb.Update(s, u)
		require.Error(t, err)
		assert.True(t, IsErrBucketDoesNotBelongToProject(err))
	})
	t.Run("moving a task to the done bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		tb := &TaskBucket{
			TaskID:        1,
			BucketID:      3, // Bucket 3 is the done bucket
			ProjectViewID: 4,
			ProjectID:     1, // In actual web requests set via the url
		}
		err := tb.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		assert.True(t, tb.Task.Done)

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":   1,
			"done": true,
		}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   1,
			"bucket_id": 3,
		}, false)
	})
	t.Run("move done task out of done bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		tb := &TaskBucket{
			TaskID:        2,
			BucketID:      1, // Bucket 1 is the default bucket
			ProjectViewID: 4,
			ProjectID:     1, // In actual web requests set via the url
		}
		err := tb.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		assert.False(t, tb.Task.Done)

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":   tb.TaskID,
			"done": false,
		}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   tb.TaskID,
			"bucket_id": 1,
		}, false)
		db.AssertMissing(t, "task_buckets", map[string]interface{}{
			"task_id":   tb.TaskID,
			"bucket_id": 3,
		})
	})
	t.Run("moving a repeating task to the done bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		tb := &TaskBucket{
			TaskID:        28,
			BucketID:      3, // Bucket 3 is the done bucket
			ProjectViewID: 4,
			ProjectID:     1, // In actual web requests set via the url
		}
		err := tb.Update(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		assert.False(t, tb.Task.Done)
		assert.Equal(t, int64(1), tb.BucketID) // This should be the actual bucket

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":   1,
			"done": false,
		}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   1,
			"bucket_id": 1,
		}, false)
	})
}
