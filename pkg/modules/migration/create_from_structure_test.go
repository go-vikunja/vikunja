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
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsertFromStructure(t *testing.T) {
	u := &user.User{
		ID: 1,
	}
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		testStructure := []*models.ProjectWithTasksAndBuckets{
			{
				Project: models.Project{
					ID:          1,
					Title:       "Test1",
					Description: "Lorem Ipsum",
				},
				Tasks: []*models.TaskWithComments{
					{
						Task: models.Task{
							Title: "Task on parent",
						},
					},
				},
			},
			{
				Project: models.Project{
					Title:           "Testproject1",
					Description:     "Something",
					ParentProjectID: 1,
				},
				Buckets: []*models.Bucket{
					{
						ID:    1234,
						Title: "Test Bucket",
					},
				},
				Tasks: []*models.TaskWithComments{
					{
						Task: models.Task{
							Title:       "Task1",
							Description: "Lorem",
						},
					},
					{
						Task: models.Task{
							Title: "Task with related tasks",
							RelatedTasks: map[models.RelationKind][]*models.Task{
								models.RelationKindSubtask: {
									{
										Title:       "Related to task with related task",
										Description: "As subtask",
									},
								},
							},
						},
					},
					{
						Task: models.Task{
							Title: "Task with attachments",
							Attachments: []*models.TaskAttachment{
								{
									File: &files.File{
										Name:        "testfile",
										Size:        4,
										FileContent: []byte{1, 2, 3, 4},
									},
								},
							},
						},
					},
					{
						Task: models.Task{
							Title: "Task with labels",
							Labels: []*models.Label{
								{
									Title:    "Label1",
									HexColor: "ff00ff",
								},
								{
									Title:    "Label2",
									HexColor: "ff00ff",
								},
							},
						},
					},
					{
						Task: models.Task{
							Title: "Task with same label",
							Labels: []*models.Label{
								{
									Title:    "Label1",
									HexColor: "ff00ff",
								},
							},
						},
					},
					{
						Task: models.Task{
							Title:    "Task in a bucket",
							BucketID: 1234,
						},
					},
					{
						Task: models.Task{
							Title:    "Task in a nonexisting bucket",
							BucketID: 1111,
						},
					},
				},
			},
		}
		err := InsertFromStructure(testStructure, u)
		require.NoError(t, err)
		db.AssertExists(t, "projects", map[string]interface{}{
			"title":       testStructure[1].Title,
			"description": testStructure[1].Description,
		}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   testStructure[1].Tasks[5].ID,
			"bucket_id": testStructure[1].Buckets[0].ID,
		}, false)
		db.AssertMissing(t, "task_buckets", map[string]interface{}{
			"bucket_id": 1111, // No task with that bucket should exist
		})
		db.AssertExists(t, "tasks", map[string]interface{}{
			"title": testStructure[0].Tasks[0].Title,
		}, false)
		assert.NotEqual(t, 0, testStructure[1].Tasks[0].BucketID) // Should get the default bucket
		assert.NotEqual(t, 0, testStructure[1].Tasks[6].BucketID) // Should get the default bucket
	})
}
