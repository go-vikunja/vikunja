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

func TestKanbanViewBucketFiltering(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	view, err := GetProjectViewByID(s, 4)
	require.NoError(t, err)

	project, err := GetProjectSimpleByID(s, view.ProjectID)
	require.NoError(t, err)

	buckets, err := GetTasksInBucketsForView(s, view, []*Project{project}, &taskSearchOptions{}, &user.User{ID: 1})
	require.NoError(t, err)

	taskBuckets := map[int64][]int64{}
	for _, b := range buckets {
		for _, tsk := range b.Tasks {
			taskBuckets[tsk.ID] = append(taskBuckets[tsk.ID], b.ID)
		}
	}

	for tid, bs := range taskBuckets {
		assert.Lenf(t, bs, 1, "task %d appears in multiple buckets: %v", tid, bs)
	}

	for _, id := range []int64{40, 41, 42, 43, 44, 45, 46} {
		assert.NotContains(t, taskBuckets, id)
	}
}
