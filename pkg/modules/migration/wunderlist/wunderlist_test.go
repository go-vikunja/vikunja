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

package wunderlist

import (
	"io/ioutil"
	"strconv"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"github.com/stretchr/testify/assert"
	"gopkg.in/d4l3k/messagediff.v1"
)

func TestWunderlistParsing(t *testing.T) {

	config.InitConfig()

	time1, err := time.Parse(time.RFC3339Nano, "2013-08-30T08:29:46.203Z")
	assert.NoError(t, err)
	time1 = time1.In(config.GetTimeZone())
	time2, err := time.Parse(time.RFC3339Nano, "2013-08-30T08:36:13.273Z")
	assert.NoError(t, err)
	time2 = time2.In(config.GetTimeZone())
	time3, err := time.Parse(time.RFC3339Nano, "2013-09-05T08:36:13.273Z")
	assert.NoError(t, err)
	time3 = time3.In(config.GetTimeZone())
	time4, err := time.Parse(time.RFC3339Nano, "2013-08-02T11:58:55Z")
	assert.NoError(t, err)
	time4 = time4.In(config.GetTimeZone())

	exampleFile, err := ioutil.ReadFile(config.ServiceRootpath.GetString() + "/pkg/modules/migration/wunderlist/testimage.jpg")
	assert.NoError(t, err)

	createTestTask := func(id, listID int, done bool) *task {
		completedAt, err := time.Parse(time.RFC3339Nano, "1970-01-01T00:00:00Z")
		assert.NoError(t, err)
		if done {
			completedAt = time1
		}
		completedAt = completedAt.In(config.GetTimeZone())
		return &task{
			ID:          id,
			AssigneeID:  123,
			CreatedAt:   time1,
			DueDate:     "2013-09-05",
			ListID:      listID,
			Title:       "Ipsum" + strconv.Itoa(id),
			Completed:   done,
			CompletedAt: completedAt,
		}
	}

	createTestNote := func(id, taskID int) *note {
		return &note{
			ID:        id,
			TaskID:    taskID,
			Content:   "Lorem Ipsum dolor sit amet",
			CreatedAt: time3,
			UpdatedAt: time2,
		}
	}

	fixtures := &wunderlistContents{
		folders: []*folder{
			{
				ID:        123,
				Title:     "Lorem Ipsum",
				ListIds:   []int{1, 2, 3, 4},
				CreatedAt: time1,
				UpdatedAt: time2,
			},
		},
		lists: []*list{
			{
				ID:        1,
				CreatedAt: time1,
				Title:     "Lorem1",
			},
			{
				ID:        2,
				CreatedAt: time1,
				Title:     "Lorem2",
			},
			{
				ID:        3,
				CreatedAt: time1,
				Title:     "Lorem3",
			},
			{
				ID:        4,
				CreatedAt: time1,
				Title:     "Lorem4",
			},
			{
				ID:        5,
				CreatedAt: time4,
				Title:     "List without a namespace",
			},
		},
		tasks: []*task{
			createTestTask(1, 1, false),
			createTestTask(2, 1, false),
			createTestTask(3, 2, true),
			createTestTask(4, 2, false),
			createTestTask(5, 3, false),
			createTestTask(6, 3, true),
			createTestTask(7, 3, true),
			createTestTask(8, 3, false),
			createTestTask(9, 4, true),
			createTestTask(10, 4, true),
		},
		notes: []*note{
			createTestNote(1, 1),
			createTestNote(2, 2),
			createTestNote(3, 3),
		},
		files: []*file{
			{
				ID:          1,
				URL:         "https://vikunja.io/testimage.jpg", // Using an image which we are hosting, so it'll still be up
				TaskID:      1,
				ListID:      1,
				FileName:    "file.md",
				ContentType: "text/plain",
				FileSize:    12345,
				CreatedAt:   time2,
				UpdatedAt:   time4,
			},
			{
				ID:          2,
				URL:         "https://vikunja.io/testimage.jpg",
				TaskID:      3,
				ListID:      2,
				FileName:    "file2.md",
				ContentType: "text/plain",
				FileSize:    12345,
				CreatedAt:   time3,
				UpdatedAt:   time4,
			},
		},
		reminders: []*reminder{
			{
				ID:        1,
				Date:      time4,
				TaskID:    1,
				CreatedAt: time4,
				UpdatedAt: time4,
			},
			{
				ID:        2,
				Date:      time3,
				TaskID:    4,
				CreatedAt: time3,
				UpdatedAt: time3,
			},
		},
		subtasks: []*subtask{
			{
				ID:        1,
				TaskID:    2,
				CreatedAt: time4,
				Title:     "LoremSub1",
			},
			{
				ID:        2,
				TaskID:    2,
				CreatedAt: time4,
				Title:     "LoremSub2",
			},
			{
				ID:        3,
				TaskID:    4,
				CreatedAt: time4,
				Title:     "LoremSub3",
			},
		},
	}

	expectedHierachie := []*models.NamespaceWithListsAndTasks{
		{
			Namespace: models.Namespace{
				Title:   "Lorem Ipsum",
				Created: time1,
				Updated: time2,
			},
			Lists: []*models.ListWithTasksAndBuckets{
				{
					List: models.List{
						Created: time1,
						Title:   "Lorem1",
					},
					Tasks: []*models.TaskWithComments{
						{
							Task: models.Task{
								Title:       "Ipsum1",
								DueDate:     time.Unix(1378339200, 0).In(config.GetTimeZone()),
								Created:     time1,
								Description: "Lorem Ipsum dolor sit amet",
								Attachments: []*models.TaskAttachment{
									{
										File: &files.File{
											Name:        "file.md",
											Mime:        "text/plain",
											Size:        12345,
											Created:     time2,
											FileContent: exampleFile,
										},
										Created: time2,
									},
								},
								Reminders: []time.Time{time4},
							},
						},
						{
							Task: models.Task{
								Title:       "Ipsum2",
								DueDate:     time.Unix(1378339200, 0).In(config.GetTimeZone()),
								Created:     time1,
								Description: "Lorem Ipsum dolor sit amet",
								RelatedTasks: map[models.RelationKind][]*models.Task{
									models.RelationKindSubtask: {
										{
											Title: "LoremSub1",
										},
										{
											Title: "LoremSub2",
										},
									},
								},
							},
						},
					},
				},
				{
					List: models.List{
						Created: time1,
						Title:   "Lorem2",
					},
					Tasks: []*models.TaskWithComments{
						{
							Task: models.Task{
								Title:       "Ipsum3",
								Done:        true,
								DoneAt:      time1,
								DueDate:     time.Unix(1378339200, 0).In(config.GetTimeZone()),
								Created:     time1,
								Description: "Lorem Ipsum dolor sit amet",
								Attachments: []*models.TaskAttachment{
									{
										File: &files.File{
											Name:        "file2.md",
											Mime:        "text/plain",
											Size:        12345,
											Created:     time3,
											FileContent: exampleFile,
										},
										Created: time3,
									},
								},
							},
						},
						{
							Task: models.Task{
								Title:     "Ipsum4",
								DueDate:   time.Unix(1378339200, 0).In(config.GetTimeZone()),
								Created:   time1,
								Reminders: []time.Time{time3},
								RelatedTasks: map[models.RelationKind][]*models.Task{
									models.RelationKindSubtask: {
										{
											Title: "LoremSub3",
										},
									},
								},
							},
						},
					},
				},
				{
					List: models.List{
						Created: time1,
						Title:   "Lorem3",
					},
					Tasks: []*models.TaskWithComments{
						{
							Task: models.Task{
								Title:   "Ipsum5",
								DueDate: time.Unix(1378339200, 0).In(config.GetTimeZone()),
								Created: time1,
							},
						},
						{
							Task: models.Task{
								Title:   "Ipsum6",
								DueDate: time.Unix(1378339200, 0).In(config.GetTimeZone()),
								Created: time1,
								Done:    true,
								DoneAt:  time1,
							},
						},
						{
							Task: models.Task{
								Title:   "Ipsum7",
								DueDate: time.Unix(1378339200, 0).In(config.GetTimeZone()),
								Created: time1,
								Done:    true,
								DoneAt:  time1,
							},
						},
						{
							Task: models.Task{
								Title:   "Ipsum8",
								DueDate: time.Unix(1378339200, 0).In(config.GetTimeZone()),
								Created: time1,
							},
						},
					},
				},
				{
					List: models.List{
						Created: time1,
						Title:   "Lorem4",
					},
					Tasks: []*models.TaskWithComments{
						{
							Task: models.Task{
								Title:   "Ipsum9",
								DueDate: time.Unix(1378339200, 0).In(config.GetTimeZone()),
								Created: time1,
								Done:    true,
								DoneAt:  time1,
							},
						},
						{
							Task: models.Task{
								Title:   "Ipsum10",
								DueDate: time.Unix(1378339200, 0).In(config.GetTimeZone()),
								Created: time1,
								Done:    true,
								DoneAt:  time1,
							},
						},
					},
				},
			},
		},
		{
			Namespace: models.Namespace{
				Title: "Migrated from wunderlist",
			},
			Lists: []*models.ListWithTasksAndBuckets{
				{
					List: models.List{
						Created: time4,
						Title:   "List without a namespace",
					},
				},
			},
		},
	}

	hierachie, err := convertWunderlistToVikunja(fixtures)
	assert.NoError(t, err)
	assert.NotNil(t, hierachie)
	if diff, equal := messagediff.PrettyDiff(hierachie, expectedHierachie); !equal {
		t.Errorf("converted wunderlist data = %v, want %v, diff: %v", hierachie, expectedHierachie, diff)
	}
}
