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

package trello

import (
	"bytes"
	"io/ioutil"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"github.com/adlio/trello"
	"github.com/d4l3k/messagediff"
	"github.com/stretchr/testify/assert"
)

func TestConvertTrelloToVikunja(t *testing.T) {

	config.InitConfig()

	time1, err := time.Parse(time.RFC3339Nano, "2014-09-26T08:25:05Z")
	assert.NoError(t, err)
	exampleFile, err := ioutil.ReadFile(config.ServiceRootpath.GetString() + "/pkg/modules/migration/wunderlist/testimage.jpg")
	assert.NoError(t, err)

	trelloData := []*trello.Board{
		{
			Name:   "TestBoard",
			Desc:   "This is a description",
			Closed: false,
			Lists: []*trello.List{
				{
					Name: "Test List 1",
					Cards: []*trello.Card{
						{
							Name: "Test Card 1",
							Desc: "Card Description",
							Pos:  123,
							Due:  &time1,
							Labels: []*trello.Label{
								{
									ID:    "ide1",
									Name:  "Label 1",
									Color: "green",
								},
								{
									ID:    "ide2",
									Name:  "Label 2",
									Color: "orange",
								},
							},
							Attachments: []*trello.Attachment{
								{
									Name:     "Testimage.jpg",
									IsUpload: true,
									MimeType: "image/jpg",
									URL:      "https://vikunja.io/testimage.jpg",
								},
							},
						},
						{
							Name: "Test Card 2",
							Pos:  124,
							Checklists: []*trello.Checklist{
								{
									Name: "Checklist 1",
									CheckItems: []trello.CheckItem{
										{
											State: "pending",
											Name:  "Pending Task",
										},
										{
											State: "completed",
											Name:  "Completed Task",
										},
									},
								},
								{
									Name: "Checklist 2",
									CheckItems: []trello.CheckItem{
										{
											State: "pending",
											Name:  "Pending Task",
										},
										{
											State: "pending",
											Name:  "Another Pending Task",
										},
									},
								},
							},
						},
						{
							Name: "Test Card 3",
							Pos:  126,
						},
						{
							Name: "Test Card 4",
							Pos:  127,
							Labels: []*trello.Label{
								{
									ID:    "ide2",
									Name:  "Label 2",
									Color: "orange",
								},
							},
						},
					},
				},
				{
					Name: "Test List 2",
					Cards: []*trello.Card{
						{
							Name: "Test Card 5",
							Pos:  111,
							Labels: []*trello.Label{
								{
									ID:    "ide3",
									Name:  "Label 3",
									Color: "blue",
								},
							},
						},
						{
							Name: "Test Card 6",
							Due:  &time1,
							Pos:  222,
						},
						{
							Name: "Test Card 7",
							Pos:  333,
						},
						{
							Name: "Test Card 8",
							Pos:  444,
						},
					},
				},
			},
		},
		{
			Name:   "TestBoard 2",
			Closed: false,
			Lists: []*trello.List{
				{
					Name: "Test List 4",
					Cards: []*trello.Card{
						{
							Name: "Test Card 634",
							Pos:  123,
						},
					},
				},
			},
		},
		{
			Name:   "TestBoard Archived",
			Closed: true,
			Lists: []*trello.List{
				{
					Name: "Test List 5",
					Cards: []*trello.Card{
						{
							Name: "Test Card 63423",
							Pos:  123,
						},
					},
				},
			},
		},
	}
	trelloData[0].Prefs.BackgroundImage = "https://vikunja.io/testimage.jpg" // Using an image which we are hosting, so it'll still be up

	expectedHierachie := []*models.NamespaceWithListsAndTasks{
		{
			Namespace: models.Namespace{
				Title: "Imported from Trello",
			},
			Lists: []*models.ListWithTasksAndBuckets{
				{
					List: models.List{
						Title:                 "TestBoard",
						Description:           "This is a description",
						BackgroundInformation: bytes.NewBuffer(exampleFile),
					},
					Buckets: []*models.Bucket{
						{
							ID:    1,
							Title: "Test List 1",
						},
						{
							ID:    2,
							Title: "Test List 2",
						},
					},
					Tasks: []*models.TaskWithComments{
						{
							Task: models.Task{
								Title:          "Test Card 1",
								Description:    "Card Description",
								BucketID:       1,
								KanbanPosition: 123,
								DueDate:        time1,
								Labels: []*models.Label{
									{
										Title:    "Label 1",
										HexColor: trelloColorMap["green"],
									},
									{
										Title:    "Label 2",
										HexColor: trelloColorMap["orange"],
									},
								},
								Attachments: []*models.TaskAttachment{
									{
										File: &files.File{
											Name:        "Testimage.jpg",
											Mime:        "image/jpg",
											Size:        uint64(len(exampleFile)),
											FileContent: exampleFile,
										},
									},
								},
							},
						},
						{
							Task: models.Task{
								Title: "Test Card 2",
								Description: `

## Checklist 1

* [ ] Pending Task
* [x] Completed Task

## Checklist 2

* [ ] Pending Task
* [ ] Another Pending Task`,
								BucketID:       1,
								KanbanPosition: 124,
							},
						},
						{
							Task: models.Task{
								Title:          "Test Card 3",
								BucketID:       1,
								KanbanPosition: 126,
							},
						},
						{
							Task: models.Task{
								Title:          "Test Card 4",
								BucketID:       1,
								KanbanPosition: 127,
								Labels: []*models.Label{
									{
										Title:    "Label 2",
										HexColor: trelloColorMap["orange"],
									},
								},
							},
						},
						{
							Task: models.Task{
								Title:          "Test Card 5",
								BucketID:       2,
								KanbanPosition: 111,
								Labels: []*models.Label{
									{
										Title:    "Label 3",
										HexColor: trelloColorMap["blue"],
									},
								},
							},
						},
						{
							Task: models.Task{
								Title:          "Test Card 6",
								BucketID:       2,
								KanbanPosition: 222,
								DueDate:        time1,
							},
						},
						{
							Task: models.Task{
								Title:          "Test Card 7",
								BucketID:       2,
								KanbanPosition: 333,
							},
						},
						{
							Task: models.Task{
								Title:          "Test Card 8",
								BucketID:       2,
								KanbanPosition: 444,
							},
						},
					},
				},
				{
					List: models.List{
						Title: "TestBoard 2",
					},
					Buckets: []*models.Bucket{
						{
							ID:    3,
							Title: "Test List 4",
						},
					},
					Tasks: []*models.TaskWithComments{
						{
							Task: models.Task{
								Title:          "Test Card 634",
								BucketID:       3,
								KanbanPosition: 123,
							},
						},
					},
				},
				{
					List: models.List{
						Title:      "TestBoard Archived",
						IsArchived: true,
					},
					Buckets: []*models.Bucket{
						{
							ID:    4,
							Title: "Test List 5",
						},
					},
					Tasks: []*models.TaskWithComments{
						{
							Task: models.Task{
								Title:          "Test Card 63423",
								BucketID:       4,
								KanbanPosition: 123,
							},
						},
					},
				},
			},
		},
	}

	hierachie, err := convertTrelloDataToVikunja(trelloData, "")
	assert.NoError(t, err)
	assert.NotNil(t, hierachie)
	if diff, equal := messagediff.PrettyDiff(hierachie, expectedHierachie); !equal {
		t.Errorf("converted trello data = %v, want %v, diff: %v", hierachie, expectedHierachie, diff)
	}
}
