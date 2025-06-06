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

package trello

import (
	"bytes"
	"os"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"

	"github.com/adlio/trello"
	"github.com/d4l3k/messagediff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestBoard(t *testing.T) ([]*trello.Board, time.Time) {

	config.InitConfig()

	time1, err := time.Parse(time.RFC3339Nano, "2014-09-26T08:25:05Z")
	require.NoError(t, err)

	trelloData := []*trello.Board{
		{
			Name: "TestBoard",
			Organization: trello.Organization{
				ID:          "orgid",
				DisplayName: "TestOrg",
			},
			IDOrganization: "orgid",
			Desc:           "This is a description",
			Closed:         false,
			Lists: []*trello.List{
				{
					Name: "Test Project 1",
					Cards: []*trello.Card{
						{
							Name: "Test Card 1",
							Desc: "Card Description **bold**",
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
									ID:       "5cc71b16f0c7a57bed3c94e9",
									Name:     "Testimage.jpg",
									IsUpload: true,
									MimeType: "image/jpg",
									URL:      "https://vikunja.io/testimage.jpg",
								},
								{
									ID:       "7cc71b16f0c7a57bed3c94e9",
									Name:     "Website",
									IsUpload: false,
									MimeType: "",
									URL:      "https://vikunja.io",
								},
							},
						},
						{
							Name: "Test Card 2",
							Pos:  124,
							Checklists: []*trello.Checklist{
								{
									Name: "Checkproject 1",
									CheckItems: []trello.CheckItem{
										{
											State: "pending",
											Name:  "Pending Task",
										},
										{
											State: "complete",
											Name:  "Completed Task",
										},
									},
								},
								{
									Name: "Checkproject 2",
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
					Name: "Test Project 2",
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
								{
									ID:    "ide4",
									Name:  "Label 4",
									Color: "green_dark",
								},
								{
									ID:    "ide5",
									Name:  "Label 5",
									Color: "doesnotexist",
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
			Organization: trello.Organization{
				ID:          "orgid2",
				DisplayName: "TestOrg2",
			},
			IDOrganization: "orgid2",
			Name:           "TestBoard 2",
			Closed:         false,
			Lists: []*trello.List{
				{
					Name: "Test Project 4",
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
			Organization: trello.Organization{
				ID:          "orgid",
				DisplayName: "TestOrg",
			},
			IDOrganization: "orgid",
			Name:           "TestBoard Archived",
			Closed:         true,
			Lists: []*trello.List{
				{
					Name: "Test Project 5",
					Cards: []*trello.Card{
						{
							Name: "Test Card 63423",
							Pos:  123,
						},
					},
				},
			},
		},
		{
			Name: "Personal Board",
			Lists: []*trello.List{
				{
					Name: "Test Project 6",
					Cards: []*trello.Card{
						{
							Name: "Test Card 5659",
							Pos:  123,
						},
					},
				},
			},
		},
	}
	trelloData[0].Prefs.BackgroundImage = "https://vikunja.io/testimage.jpg" // Using an image which we are hosting, so it'll still be up

	return trelloData, time1
}

func TestConvertTrelloToVikunja(t *testing.T) {
	trelloData, time1 := getTestBoard(t)

	exampleFile, err := os.ReadFile(config.ServiceRootpath.GetString() + "/pkg/modules/migration/testimage.jpg")
	require.NoError(t, err)

	expectedHierarchyOrg := map[string][]*models.ProjectWithTasksAndBuckets{
		"orgid": {
			{
				Project: models.Project{
					ID:    1,
					Title: "orgid",
				},
			},
			{
				Project: models.Project{
					ID:                    2,
					ParentProjectID:       1,
					Title:                 "TestBoard",
					Description:           "This is a description",
					BackgroundInformation: bytes.NewBuffer(exampleFile),
				},
				Buckets: []*models.Bucket{
					{
						ID:    1,
						Title: "Test Project 1",
					},
					{
						ID:    2,
						Title: "Test Project 2",
					},
				},
				Tasks: []*models.TaskWithComments{
					{
						Task: models.Task{
							Title: "Test Card 1",
							Description: "<p>Card Description <strong>bold</strong></p>\n" +
								"<p><a href=\"https://vikunja.io\">Website</a></p>\n",
							BucketID: 1,
							DueDate:  time1,
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

<h2> Checkproject 1</h2>

<ul data-type="taskList">
<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label><div><p>Pending Task</p></div></li>
<li data-checked="true" data-type="taskItem"><label><input type="checkbox" checked="checked"><span></span></label><div><p>Completed Task</p></div></li></ul>

<h2> Checkproject 2</h2>

<ul data-type="taskList">
<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label><div><p>Pending Task</p></div></li>
<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label><div><p>Another Pending Task</p></div></li></ul>`,
							BucketID: 1,
						},
					},
					{
						Task: models.Task{
							Title:    "Test Card 3",
							BucketID: 1,
						},
					},
					{
						Task: models.Task{
							Title:    "Test Card 4",
							BucketID: 1,
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
							Title:    "Test Card 5",
							BucketID: 2,
							Labels: []*models.Label{
								{
									Title:    "Label 3",
									HexColor: trelloColorMap["blue"],
								},
								{
									Title:    "Label 4",
									HexColor: trelloColorMap["green_dark"],
								},
								{
									Title:    "Label 5",
									HexColor: trelloColorMap["transparent"],
								},
							},
						},
					},
					{
						Task: models.Task{
							Title:    "Test Card 6",
							BucketID: 2,
							DueDate:  time1,
						},
					},
					{
						Task: models.Task{
							Title:    "Test Card 7",
							BucketID: 2,
						},
					},
					{
						Task: models.Task{
							Title:    "Test Card 8",
							BucketID: 2,
						},
					},
				},
			},
			{
				Project: models.Project{
					ID:              3,
					ParentProjectID: 1,
					Title:           "TestBoard Archived",
					IsArchived:      true,
				},
				Buckets: []*models.Bucket{
					{
						ID:    3,
						Title: "Test Project 5",
					},
				},
				Tasks: []*models.TaskWithComments{
					{
						Task: models.Task{
							Title:    "Test Card 63423",
							BucketID: 3,
						},
					},
				},
			},
		},
		"orgid2": {
			{
				Project: models.Project{
					ID:    1,
					Title: "orgid2",
				},
			},
			{
				Project: models.Project{
					ID:              2,
					ParentProjectID: 1,
					Title:           "TestBoard 2",
				},
				Buckets: []*models.Bucket{
					{
						ID:    1,
						Title: "Test Project 4",
					},
				},
				Tasks: []*models.TaskWithComments{
					{
						Task: models.Task{
							Title:    "Test Card 634",
							BucketID: 1,
						},
					},
				},
			},
		},
		"Personal": {
			{
				Project: models.Project{
					ID:    1,
					Title: "Personal",
				},
			},
			{
				Project: models.Project{
					ID:              2,
					ParentProjectID: 1,
					Title:           "Personal Board",
				},
				Buckets: []*models.Bucket{
					{
						ID:    1,
						Title: "Test Project 6",
					},
				},
				Tasks: []*models.TaskWithComments{
					{
						Task: models.Task{
							Title:    "Test Card 5659",
							BucketID: 1,
						},
					},
				},
			},
		},
	}

	organizationMap := getTrelloOrganizationsWithBoards(trelloData)
	for organizationID, boards := range organizationMap {
		hierarchy, err := convertTrelloDataToVikunja(organizationID, boards, &trello.Client{}, nil)

		require.NoError(t, err)
		assert.NotNil(t, hierarchy)
		if diff, equal := messagediff.PrettyDiff(hierarchy, expectedHierarchyOrg[organizationID]); !equal {
			t.Errorf("converted trello data = %v,\nwant %v,\ndiff: %v", hierarchy, expectedHierarchyOrg[organizationID], diff)
		}
	}
}

func TestCreateOrganizationMap(t *testing.T) {
	trelloData, _ := getTestBoard(t)

	organizationMap := getTrelloOrganizationsWithBoards(trelloData)
	expectedMap := map[string][]*trello.Board{
		"orgid": {
			trelloData[0],
			trelloData[2],
		},
		"orgid2": {
			trelloData[1],
		},
		"Personal": {
			trelloData[3],
		},
	}
	if diff, equal := messagediff.PrettyDiff(organizationMap, expectedMap); !equal {
		t.Errorf("converted organization map = %v,\nwant %v,\ndiff: %v", organizationMap, expectedMap, diff)
	}
}
