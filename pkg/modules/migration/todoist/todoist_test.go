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

package todoist

import (
	"os"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/d4l3k/messagediff.v1"
)

func TestConvertTodoistToVikunja(t *testing.T) {

	config.InitConfig()

	time1, err := time.Parse(time.RFC3339Nano, "2014-09-26T08:25:05Z")
	require.NoError(t, err)
	time1 = time1.In(config.GetTimeZone())
	time3, err := time.Parse(time.RFC3339Nano, "2014-10-21T08:25:05Z")
	require.NoError(t, err)
	time3 = time3.In(config.GetTimeZone())
	dueTime, err := time.Parse(time.RFC3339Nano, "2020-05-31T23:59:00Z")
	require.NoError(t, err)
	dueTime = dueTime.In(config.GetTimeZone())
	dueTimeWithTime, err := time.Parse(time.RFC3339Nano, "2021-01-31T19:00:00Z")
	require.NoError(t, err)
	dueTimeWithTime = dueTimeWithTime.In(config.GetTimeZone())
	nilTime, err := time.Parse(time.RFC3339Nano, "0001-01-01T00:00:00Z")
	require.NoError(t, err)
	exampleFile, err := os.ReadFile(config.ServiceRootpath.GetString() + "/pkg/modules/migration/testimage.jpg")
	require.NoError(t, err)

	makeTestItem := func(id, projectId string, hasDueDate, hasLabels, done bool) *item {
		item := &item{
			ID:            id,
			UserID:        "1855589",
			ProjectID:     projectId,
			Content:       "Task" + id,
			Priority:      1,
			ChildOrder:    1,
			DateAdded:     time1,
			DateCompleted: nilTime,
		}

		if done {
			item.Checked = true
			item.DateCompleted = time3
		}

		if hasLabels {
			item.Labels = []string{
				"Label1",
				"Label2",
				"Label3",
				"Label4",
			}
		}

		if hasDueDate {
			item.Due = &dueDate{
				Date:        "2020-05-31",
				Timezone:    nil,
				IsRecurring: false,
			}
		}

		return item
	}

	testSync := &sync{
		Projects: []*project{
			{
				ID:         "396936926",
				Name:       "Project1",
				Color:      "berry_red",
				ChildOrder: 1,
				Collapsed:  false,
				Shared:     false,
				IsDeleted:  false,
				IsArchived: false,
				IsFavorite: false,
			},
			{
				ID:         "396936927",
				Name:       "Project2",
				Color:      "mint_green",
				ChildOrder: 1,
				Collapsed:  false,
				Shared:     false,
				IsDeleted:  false,
				IsArchived: false,
				IsFavorite: false,
			},
			{
				ID:         "396936928",
				Name:       "Project3 - Archived",
				Color:      "mint_green",
				ChildOrder: 1,
				Collapsed:  false,
				Shared:     false,
				IsDeleted:  false,
				IsArchived: true,
				IsFavorite: false,
			},
		},
		Items: []*item{
			makeTestItem("400000000", "396936926", false, false, false),
			makeTestItem("400000001", "396936926", false, false, false),
			makeTestItem("400000002", "396936926", false, false, false),
			makeTestItem("400000003", "396936926", true, true, true),
			makeTestItem("400000004", "396936926", false, true, false),
			makeTestItem("400000005", "396936926", true, false, true),
			makeTestItem("400000006", "396936926", true, false, true),
			{
				ID:         "400000110",
				UserID:     "1855589",
				ProjectID:  "396936926",
				Content:    "Task with parent",
				Priority:   2,
				ParentID:   "400000006",
				ChildOrder: 1,
				Checked:    false,
				DateAdded:  time1,
			},
			{
				ID:            "400000106",
				UserID:        "1855589",
				ProjectID:     "396936926",
				Content:       "Task400000106",
				Priority:      1,
				ParentID:      "",
				ChildOrder:    1,
				DateAdded:     time1,
				Checked:       true,
				DateCompleted: time3,
				Due: &dueDate{
					Date:        "2021-01-31T19:00:00Z",
					Timezone:    nil,
					IsRecurring: false,
				},
				Labels: []string{
					"Label1",
					"Label2",
					"Label3",
					"Label4",
				},
			},
			makeTestItem("400000107", "396936926", false, false, true),
			makeTestItem("400000108", "396936926", false, false, true),
			{
				ID:            "400000109",
				UserID:        "1855589",
				ProjectID:     "396936926",
				Content:       "Task400000109",
				Priority:      1,
				ChildOrder:    1,
				Checked:       true,
				DateAdded:     time1,
				DateCompleted: time3,
				SectionID:     "1234",
			},

			makeTestItem("400000007", "396936927", true, false, false),
			makeTestItem("400000008", "396936927", true, false, false),
			makeTestItem("400000009", "396936927", false, false, false),
			makeTestItem("400000010", "396936927", false, false, true),
			makeTestItem("400000101", "396936927", false, false, false),
			makeTestItem("400000102", "396936927", true, true, false),
			makeTestItem("400000103", "396936927", false, true, false),
			makeTestItem("400000104", "396936927", false, true, false),
			makeTestItem("400000105", "396936927", true, true, false),

			makeTestItem("400000111", "396936928", false, false, true),
		},
		Labels: []*label{
			{
				ID:    "80000",
				Name:  "Label1",
				Color: "berry_red",
			},
			{
				ID:    "80001",
				Name:  "Label2",
				Color: "red",
			},
			{
				ID:    "80002",
				Name:  "Label3",
				Color: "orange",
			},
			{
				ID:    "80003",
				Name:  "Label4",
				Color: "yellow",
			},
		},
		Notes: []*note{
			{
				ID:      "101476",
				ItemID:  "400000000",
				Content: "Lorem Ipsum dolor sit amet",
				Posted:  time1,
			},
			{
				ID:      "101477",
				ItemID:  "400000001",
				Content: "Lorem Ipsum dolor sit amet",
				Posted:  time1,
			},
			{
				ID:      "101478",
				ItemID:  "400000003",
				Content: "Lorem Ipsum dolor sit amet",
				Posted:  time1,
			},
			{
				ID:      "101479",
				ItemID:  "400000010",
				Content: "Lorem Ipsum dolor sit amet",
				Posted:  time1,
			},
			{
				ID:      "101480",
				ItemID:  "400000101",
				Content: "Lorem Ipsum dolor sit amet",
				FileAttachment: &fileAttachment{
					FileName:    "file.md",
					FileType:    "text/plain",
					FileSize:    12345,
					FileURL:     "https://vikunja.io/testimage.jpg", // Using an image which we are hosting, so it'll still be up
					UploadState: "completed",
				},
				Posted: time1,
			},
		},
		ProjectNotes: []*projectNote{
			{
				ID:        "102000",
				Content:   "Lorem Ipsum dolor sit amet",
				ProjectID: "396936926",
				Posted:    time3,
			},
			{
				ID:        "102001",
				Content:   "Lorem Ipsum dolor sit amet 2",
				ProjectID: "396936926",
				Posted:    time3,
			},
			{
				ID:        "102002",
				Content:   "Lorem Ipsum dolor sit amet 3",
				ProjectID: "396936926",
				Posted:    time3,
			},
			{
				ID:        "102003",
				Content:   "Lorem Ipsum dolor sit amet 4",
				ProjectID: "396936927",
				Posted:    time3,
			},
			{
				ID:        "102004",
				Content:   "Lorem Ipsum dolor sit amet 5",
				ProjectID: "396936927",
				Posted:    time3,
			},
		},
		Reminders: []*reminder{
			{
				ID:     "103000",
				ItemID: "400000000",
				Due: &dueDate{
					Date:        "2020-06-15",
					IsRecurring: false,
				},
				MmOffset: 180,
			},
			{
				ID:     "103001",
				ItemID: "400000000",
				Due: &dueDate{
					Date:        "2020-06-16T07:00:00",
					IsRecurring: false,
				},
			},
			{
				ID:     "103002",
				ItemID: "400000002",
				Due: &dueDate{
					Date:        "2020-07-15T07:00:00Z",
					IsRecurring: true,
				},
			},
			{
				ID:     "103003",
				ItemID: "400000003",
				Due: &dueDate{
					Date:        "2020-06-15T07:00:00",
					IsRecurring: false,
				},
			},
			{
				ID:     "103004",
				ItemID: "400000005",
				Due: &dueDate{
					Date:        "2020-06-15T07:00:00",
					IsRecurring: false,
				},
			},
			{
				ID:     "103006",
				ItemID: "400000009",
				Due: &dueDate{
					Date:        "2020-06-15T07:00:00",
					IsRecurring: false,
				},
			},
		},
		Sections: []*section{
			{
				ID:        "1234",
				Name:      "Some Bucket",
				ProjectID: "396936926",
			},
		},
	}

	vikunjaLabels := []*models.Label{
		{
			Title:    "Label1",
			HexColor: todoistColors["berry_red"],
		},
		{
			Title:    "Label2",
			HexColor: todoistColors["red"],
		},
		{
			Title:    "Label3",
			HexColor: todoistColors["orange"],
		},
		{
			Title:    "Label4",
			HexColor: todoistColors["yellow"],
		},
	}

	expectedHierachie := []*models.ProjectWithTasksAndBuckets{
		{
			Project: models.Project{
				ID:    1,
				Title: "Migrated from todoist",
			},
		},
		{
			Project: models.Project{
				ID:              2,
				ParentProjectID: 1,
				Title:           "Project1",
				Description:     "Lorem Ipsum dolor sit amet\nLorem Ipsum dolor sit amet 2\nLorem Ipsum dolor sit amet 3",
				HexColor:        todoistColors["berry_red"],
			},
			Buckets: []*models.Bucket{
				{
					ID:    1,
					Title: "Some Bucket",
				},
			},
			Tasks: []*models.TaskWithComments{
				{
					Task: models.Task{
						Title:       "Task400000000",
						Description: "Lorem Ipsum dolor sit amet",
						Done:        false,
						Created:     time1,
						Reminders: []*models.TaskReminder{
							{Reminder: time.Date(2020, time.June, 15, 23, 59, 0, 0, time.UTC).In(config.GetTimeZone())},
							{Reminder: time.Date(2020, time.June, 16, 7, 0, 0, 0, time.UTC).In(config.GetTimeZone())},
						},
					},
				},
				{
					Task: models.Task{
						Title:       "Task400000001",
						Description: "Lorem Ipsum dolor sit amet",
						Done:        false,
						Created:     time1,
					},
				},
				{
					Task: models.Task{
						Title:   "Task400000002",
						Done:    false,
						Created: time1,
						Reminders: []*models.TaskReminder{
							{Reminder: time.Date(2020, time.July, 15, 7, 0, 0, 0, time.UTC).In(config.GetTimeZone())},
						},
					},
				},
				{
					Task: models.Task{
						Title:       "Task400000003",
						Description: "Lorem Ipsum dolor sit amet",
						Done:        true,
						DueDate:     dueTime,
						Created:     time1,
						DoneAt:      time3,
						Labels:      vikunjaLabels,
						Reminders: []*models.TaskReminder{
							{Reminder: time.Date(2020, time.June, 15, 7, 0, 0, 0, time.UTC).In(config.GetTimeZone())},
						},
					},
				},
				{
					Task: models.Task{
						Title:   "Task400000004",
						Done:    false,
						Created: time1,
						Labels:  vikunjaLabels,
					},
				},
				{
					Task: models.Task{
						Title:   "Task400000005",
						Done:    true,
						DueDate: dueTime,
						Created: time1,
						DoneAt:  time3,
						Reminders: []*models.TaskReminder{
							{Reminder: time.Date(2020, time.June, 15, 7, 0, 0, 0, time.UTC).In(config.GetTimeZone())},
						},
					},
				},
				{
					Task: models.Task{
						Title:   "Task400000006",
						Done:    true,
						DueDate: dueTime,
						Created: time1,
						DoneAt:  time3,
						RelatedTasks: map[models.RelationKind][]*models.Task{
							models.RelationKindSubtask: {
								{
									Title:    "Task with parent",
									Done:     false,
									Priority: 2,
									Created:  time1,
									DoneAt:   nilTime,
								},
							},
						},
					},
				},
				{
					Task: models.Task{
						Title:   "Task400000106",
						Done:    true,
						DueDate: dueTimeWithTime,
						Created: time1,
						DoneAt:  time3,
						Labels:  vikunjaLabels,
					},
				},
				{
					Task: models.Task{
						Title:   "Task400000107",
						Done:    true,
						Created: time1,
						DoneAt:  time3,
					},
				},
				{
					Task: models.Task{
						Title:   "Task400000108",
						Done:    true,
						Created: time1,
						DoneAt:  time3,
					},
				},
				{
					Task: models.Task{
						Title:    "Task400000109",
						Done:     true,
						Created:  time1,
						DoneAt:   time3,
						BucketID: 1,
					},
				},
			},
		},
		{
			Project: models.Project{
				ID:              3,
				ParentProjectID: 1,
				Title:           "Project2",
				Description:     "Lorem Ipsum dolor sit amet 4\nLorem Ipsum dolor sit amet 5",
				HexColor:        todoistColors["mint_green"],
			},
			Tasks: []*models.TaskWithComments{
				{
					Task: models.Task{
						Title:   "Task400000007",
						Done:    false,
						DueDate: dueTime,
						Created: time1,
					},
				},
				{
					Task: models.Task{
						Title:   "Task400000008",
						Done:    false,
						DueDate: dueTime,
						Created: time1,
					},
				},
				{
					Task: models.Task{
						Title:   "Task400000009",
						Done:    false,
						Created: time1,
						Reminders: []*models.TaskReminder{
							{Reminder: time.Date(2020, time.June, 15, 7, 0, 0, 0, time.UTC).In(config.GetTimeZone())},
						},
					},
				},
				{
					Task: models.Task{
						Title:       "Task400000010",
						Description: "Lorem Ipsum dolor sit amet",
						Done:        true,
						Created:     time1,
						DoneAt:      time3,
					},
				},
				{
					Task: models.Task{
						Title:       "Task400000101",
						Description: "Lorem Ipsum dolor sit amet",
						Done:        false,
						Created:     time1,
						Attachments: []*models.TaskAttachment{
							{
								File: &files.File{
									Name:        "file.md",
									Mime:        "text/plain",
									Size:        12345,
									Created:     time1,
									FileContent: exampleFile,
								},
								Created: time1,
							},
						},
					},
				},
				{
					Task: models.Task{
						Title:   "Task400000102",
						Done:    false,
						DueDate: dueTime,
						Created: time1,
						Labels:  vikunjaLabels,
					},
				},
				{
					Task: models.Task{
						Title:   "Task400000103",
						Done:    false,
						Created: time1,
						Labels:  vikunjaLabels,
					},
				},
				{
					Task: models.Task{
						Title:   "Task400000104",
						Done:    false,
						Created: time1,
						Labels:  vikunjaLabels,
					},
				},
				{
					Task: models.Task{
						Title:   "Task400000105",
						Done:    false,
						DueDate: dueTime,
						Created: time1,
						Labels:  vikunjaLabels,
					},
				},
			},
		},
		{
			Project: models.Project{
				ID:              4,
				ParentProjectID: 1,
				Title:           "Project3 - Archived",
				HexColor:        todoistColors["mint_green"],
				IsArchived:      true,
			},
			Tasks: []*models.TaskWithComments{
				{
					Task: models.Task{
						Title:   "Task400000111",
						Done:    true,
						Created: time1,
						DoneAt:  time3,
					},
				},
			},
		},
	}

	doneItems := make(map[string]*doneItem)
	hierachie, err := convertTodoistToVikunja(testSync, doneItems)
	require.NoError(t, err)
	assert.NotNil(t, hierachie)
	if diff, equal := messagediff.PrettyDiff(hierachie, expectedHierachie); !equal {
		t.Errorf("converted todoist data = %v, want %v, diff: %v", hierachie, expectedHierachie, diff)
	}
}
