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

	"xorm.io/xorm"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBucket_Delete(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		b := &Bucket{
			ID:            2, // The second bucket only has 3 tasks
			ProjectID:     1,
			ProjectViewID: 4,
		}
		err := b.Delete(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		// Assert all tasks have been moved to bucket 1 as that one is the first
		tasks := []*TaskBucket{}
		err = s.Where("bucket_id = ?", 1).Find(&tasks)
		require.NoError(t, err)
		assert.Len(t, tasks, 14)
		db.AssertMissing(t, "buckets", map[string]interface{}{
			"id":              2,
			"project_view_id": 4,
		})
	})
	t.Run("last bucket in project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		b := &Bucket{
			ID:            34,
			ProjectID:     18,
			ProjectViewID: 72,
		}
		err := b.Delete(s, u)
		require.Error(t, err)
		assert.True(t, IsErrCannotRemoveLastBucket(err))
		err = s.Commit()
		require.NoError(t, err)

		db.AssertExists(t, "buckets", map[string]interface{}{
			"id":              34,
			"project_view_id": 72,
		}, false)
	})
	t.Run("done bucket should be reset", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		b := &Bucket{
			ID:            3,
			ProjectID:     1,
			ProjectViewID: 4,
		}
		err := b.Delete(s, u)
		require.NoError(t, err)

		db.AssertExists(t, "project_views", map[string]interface{}{
			"id":             b.ProjectViewID,
			"done_bucket_id": 0,
		}, false)
	})
}

func TestBucket_Update(t *testing.T) {

	testAndAssertBucketUpdate := func(t *testing.T, b *Bucket, s *xorm.Session) {
		err := b.Update(s, &user.User{ID: 1})
		require.NoError(t, err)

		err = s.Commit()
		require.NoError(t, err)

		db.AssertExists(t, "buckets", map[string]interface{}{
			"id":    1,
			"title": b.Title,
			"limit": b.Limit,
		}, false)
	}

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		b := &Bucket{
			ID:            1,
			Title:         "New Name",
			Limit:         2,
			ProjectViewID: 4,
		}

		testAndAssertBucketUpdate(t, b, s)
	})
	t.Run("reset limit", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		b := &Bucket{
			ID:            1,
			Title:         "testbucket1",
			Limit:         0,
			ProjectViewID: 4,
		}

		testAndAssertBucketUpdate(t, b, s)
	})
}
