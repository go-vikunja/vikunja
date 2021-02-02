// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package migration

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
)

func TestInsertFromStructure(t *testing.T) {
	u := &user.User{
		ID: 1,
	}
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		testStructure := []*models.NamespaceWithLists{
			{
				Namespace: models.Namespace{
					Title:       "Test1",
					Description: "Lorem Ipsum",
				},
				Lists: []*models.List{
					{
						Title:       "Testlist1",
						Description: "Something",
						Buckets: []*models.Bucket{
							{
								ID:    1234,
								Title: "Test Bucket",
							},
						},
						Tasks: []*models.Task{
							{
								Title:       "Task1",
								Description: "Lorem",
							},
							{
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
							{
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
							{
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
							{
								Title: "Task with same label",
								Labels: []*models.Label{
									{
										Title:    "Label1",
										HexColor: "ff00ff",
									},
								},
							},
							{
								Title:    "Task in a bucket",
								BucketID: 1234,
							},
							{
								Title:    "Task in a nonexisting bucket",
								BucketID: 1111,
							},
						},
					},
				},
			},
		}
		err := InsertFromStructure(testStructure, u)
		assert.NoError(t, err)
		db.AssertExists(t, "namespaces", map[string]interface{}{
			"title":       testStructure[0].Namespace.Title,
			"description": testStructure[0].Namespace.Description,
		}, false)
		db.AssertExists(t, "list", map[string]interface{}{
			"title":       testStructure[0].Lists[0].Title,
			"description": testStructure[0].Lists[0].Description,
		}, false)
		db.AssertExists(t, "tasks", map[string]interface{}{
			"title":     testStructure[0].Lists[0].Tasks[5].Title,
			"bucket_id": testStructure[0].Lists[0].Buckets[0].ID,
		}, false)
		db.AssertMissing(t, "tasks", map[string]interface{}{
			"title":     testStructure[0].Lists[0].Tasks[6].Title,
			"bucket_id": 1111, // No task with that bucket should exist
		})
		assert.NotEqual(t, 0, testStructure[0].Lists[0].Tasks[0].BucketID) // Should get the default bucket
		assert.NotEqual(t, 0, testStructure[0].Lists[0].Tasks[6].BucketID) // Should get the default bucket
	})
}
