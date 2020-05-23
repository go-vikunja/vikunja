// Vikunja is a todo-list application to facilitate your life.
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

package migration

import (
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"testing"
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
						},
					},
				},
			},
		}
		err := InsertFromStructure(testStructure, u)
		assert.NoError(t, err)
	})
}
