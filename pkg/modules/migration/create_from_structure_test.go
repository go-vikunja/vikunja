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
	"fmt"
	"testing"
	"time"

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
					ParentProjectID: models.Ptr(int64(1)),
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
	t.Run("done tasks stay done when placed in an imported bucket", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		doneAt := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
		structure := []*models.ProjectWithTasksAndBuckets{
			{
				Project: models.Project{Title: "Done bucket import"},
				Buckets: []*models.Bucket{
					{ID: 1, Title: "Archive"},
				},
				Tasks: []*models.TaskWithComments{
					{Task: models.Task{Title: "Archived task", Done: true, BucketID: 1}},
					{Task: models.Task{Title: "Open task", Done: false, BucketID: 1}},
					{Task: models.Task{Title: "Task done earlier", Done: true, DoneAt: doneAt, BucketID: 1}},
				},
			},
		}
		require.NoError(t, InsertFromStructure(structure, u))

		db.AssertExists(t, "tasks", map[string]interface{}{
			"title": "Archived task",
			"done":  true,
		}, false)
		db.AssertExists(t, "tasks", map[string]interface{}{
			"title": "Open task",
			"done":  false,
		}, false)
		db.AssertExists(t, "task_buckets", map[string]interface{}{
			"task_id":   structure[0].Tasks[0].ID,
			"bucket_id": structure[0].Buckets[0].ID,
		}, false)

		s := db.NewSession()
		defer s.Close()
		task := &models.Task{}
		found, err := s.ID(structure[0].Tasks[2].ID).Get(task)
		require.NoError(t, err)
		require.True(t, found)
		assert.WithinDuration(t, doneAt, task.DoneAt, time.Second, "an explicit done_at is kept")
	})
	t.Run("reuses existing labels across imports", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		makeStructure := func() []*models.ProjectWithTasksAndBuckets {
			return []*models.ProjectWithTasksAndBuckets{
				{
					Project: models.Project{Title: "Import project"},
					Tasks: []*models.TaskWithComments{
						{
							Task: models.Task{
								Title: "Task with label",
								Labels: []*models.Label{
									{Title: "Mealie", HexColor: "abcdef"},
								},
							},
						},
					},
				},
			}
		}

		require.NoError(t, InsertFromStructure(makeStructure(), u))
		require.NoError(t, InsertFromStructure(makeStructure(), u))

		s := db.NewSession()
		defer s.Close()
		count, err := s.Where("created_by_id = ? AND title = ?", u.ID, "Mealie").Count(&models.Label{})
		require.NoError(t, err)
		assert.Equal(t, int64(1), count, "second import must reuse the existing 'Mealie' label")
	})
	t.Run("does not merge into another user's label", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		// Fixture label #3 'Label #3 - other user' is created_by_id: 2.
		// Importing the same title for user 1 must create a new, user-owned label.
		structure := []*models.ProjectWithTasksAndBuckets{
			{
				Project: models.Project{Title: "Import project"},
				Tasks: []*models.TaskWithComments{
					{
						Task: models.Task{
							Title:  "Task",
							Labels: []*models.Label{{Title: "Label #3 - other user"}},
						},
					},
				},
			},
		}
		require.NoError(t, InsertFromStructure(structure, u))

		db.AssertExists(t, "labels", map[string]interface{}{
			"title":         "Label #3 - other user",
			"created_by_id": u.ID,
		}, false)
	})
	t.Run("seeds positions when the export has none", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		const taskCount = 100

		tasks := make([]*models.TaskWithComments, 0, taskCount)
		for i := range taskCount {
			tasks = append(tasks, &models.TaskWithComments{
				Task: models.Task{Title: fmt.Sprintf("Task %d", i)},
			})
		}
		require.NoError(t, InsertFromStructure([]*models.ProjectWithTasksAndBuckets{{
			Project: models.Project{Title: "Import project"},
			Tasks:   tasks,
		}}, u))

		s := db.NewSession()
		defer s.Close()

		project := &models.Project{}
		exists, err := s.Where("title = ?", "Import project").Get(project)
		require.NoError(t, err)
		require.True(t, exists)

		positions := []*models.TaskPosition{}
		require.NoError(t, s.
			Join("INNER", "tasks", "tasks.id = task_positions.task_id").
			Where("tasks.project_id = ?", project.ID).
			OrderBy("task_positions.project_view_id, tasks.id").
			Find(&positions))
		require.Len(t, positions, taskCount*4, "every task needs a position in all four default views")

		// Halving the lowest position for every task collapses positions towards zero and
		// reverses the import order, and recalculates the whole view on the way (#3297).
		for i, p := range positions {
			assert.GreaterOrEqualf(t, p.Position, models.MinPositionSpacing, "position %d of view %d collapsed", i, p.ProjectViewID)
			if i > 0 && positions[i-1].ProjectViewID == p.ProjectViewID {
				assert.Greaterf(t, p.Position, positions[i-1].Position, "tasks must keep their import order in view %d", p.ProjectViewID)
			}
		}
	})
	t.Run("archives children of an archived project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		// Old exports only flag the parent as archived.
		require.NoError(t, InsertFromStructure([]*models.ProjectWithTasksAndBuckets{
			{
				Project: models.Project{
					ID:         1,
					Title:      "Archived parent",
					IsArchived: true,
				},
			},
			{
				Project: models.Project{
					ID:              2,
					Title:           "Unflagged child",
					ParentProjectID: models.Ptr(int64(1)),
				},
			},
			{
				Project: models.Project{
					ID:              3,
					Title:           "Unflagged grandchild",
					ParentProjectID: models.Ptr(int64(2)),
				},
			},
			{
				Project: models.Project{
					ID:    4,
					Title: "Unarchived sibling root",
				},
			},
			{
				Project: models.Project{
					ID:              5,
					Title:           "Unarchived sibling's child",
					ParentProjectID: models.Ptr(int64(4)),
				},
			},
		}, u))

		s := db.NewSession()
		defer s.Close()

		for _, title := range []string{"Archived parent", "Unflagged child", "Unflagged grandchild"} {
			project := &models.Project{}
			exists, err := s.Where("title = ?", title).Get(project)
			require.NoError(t, err)
			require.True(t, exists)
			assert.True(t, project.IsArchived)
		}

		for _, title := range []string{"Unarchived sibling root", "Unarchived sibling's child"} {
			project := &models.Project{}
			exists, err := s.Where("title = ?", title).Get(project)
			require.NoError(t, err)
			require.True(t, exists)
			assert.False(t, project.IsArchived)
		}
	})
	t.Run("keeps positions the export provides", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		require.NoError(t, InsertFromStructure([]*models.ProjectWithTasksAndBuckets{{
			Project: models.Project{Title: "Import project"},
			Tasks: []*models.TaskWithComments{
				{Task: models.Task{Title: "Second", Position: 200}},
				{Task: models.Task{Title: "First", Position: 100}},
			},
		}}, u))

		s := db.NewSession()
		defer s.Close()

		for title, position := range map[string]float64{"First": 100, "Second": 200} {
			task := &models.Task{}
			exists, err := s.Where("title = ?", title).Get(task)
			require.NoError(t, err)
			require.True(t, exists)

			count, err := s.Where("task_id = ? AND position = ?", task.ID, position).Count(&models.TaskPosition{})
			require.NoError(t, err)
			assert.Equal(t, int64(4), count, "task %q must keep position %v in all views", title, position)
		}
	})
	t.Run("assignees from a foreign instance", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		foreignID := int64(999)
		require.NoError(t, InsertFromStructure([]*models.ProjectWithTasksAndBuckets{{
			Project: models.Project{Title: "Import project"},
			Tasks: []*models.TaskWithComments{
				{Task: models.Task{Title: "email match", Assignees: []*user.User{{ID: foreignID, Username: "someone-else", Email: "USER1@example.com"}}}},
				{Task: models.Task{Title: "username match", Assignees: []*user.User{{ID: foreignID, Username: "user1"}}}},
				{Task: models.Task{Title: "no match", Assignees: []*user.User{{ID: 2, Username: "other", Email: "other@example.com"}}}},
				{Task: models.Task{
					Title: "related",
					RelatedTasks: map[models.RelationKind][]*models.Task{
						models.RelationKindSubtask: {{Title: "related match", Assignees: []*user.User{{ID: foreignID, Username: "user1"}}}},
					},
				}},
			},
		}}, u))

		s := db.NewSession()
		defer s.Close()

		for title, wantAssignee := range map[string]bool{"email match": true, "username match": true, "related match": true, "no match": false} {
			task := &models.Task{}
			exists, err := s.Where("title = ?", title).Get(task)
			require.NoError(t, err)
			require.True(t, exists, title)

			assignees := []*models.TaskAssginee{}
			require.NoError(t, s.Where("task_id = ?", task.ID).Find(&assignees))
			if wantAssignee {
				require.Len(t, assignees, 1, title)
				assert.Equal(t, u.ID, assignees[0].UserID, title)
			} else {
				assert.Empty(t, assignees, title)
			}
		}
	})
}
