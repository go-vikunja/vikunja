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
	"net/http"
	"net/http/httptest"
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
	exampleFile, err := os.ReadFile("../testimage.jpg")
	require.NoError(t, err)

	// Serve the attachment from a local test server so the test does not depend
	// on an external host being reachable. The SSRF-safe client used by
	// migration.DownloadFile rejects non-routable IPs by default, so allow them
	// for the duration of this test.
	prevAllowNonRoutable := config.OutgoingRequestsAllowNonRoutableIPs.GetBool()
	config.OutgoingRequestsAllowNonRoutableIPs.Set("true")
	t.Cleanup(func() {
		config.OutgoingRequestsAllowNonRoutableIPs.Set(prevAllowNonRoutable)
	})
	attachmentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(exampleFile)
	}))
	t.Cleanup(attachmentServer.Close)

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
					FileURL:     attachmentServer.URL,
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
				ParentProjectID: models.Ptr(int64(1)),
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
				ParentProjectID: models.Ptr(int64(1)),
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
									Name: "file.md",
									Mime: "text/plain",
									// Size from content, not API metadata (GHSA-qh78-rvg3-cv54 defense-in-depth).
									Size:        uint64(len(exampleFile)),
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
				ParentProjectID: models.Ptr(int64(1)),
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

func TestConvertTodoistToVikunjaWithBrokenAttachment(t *testing.T) {
	// Todoist returns opaque identifiers instead of urls for attachments it does not host itself.
	// Those must not fail the whole migration, see https://github.com/go-vikunja/vikunja/issues/3435
	testSync := &sync{
		Projects: []*project{
			{
				ID:   "396936926",
				Name: "Project1",
			},
		},
		Items: []*item{
			{
				ID:        "400000001",
				ProjectID: "396936926",
				Content:   "Task1",
			},
		},
		Notes: []*note{
			{
				ID:      "101478",
				ItemID:  "400000001",
				Content: "Lorem Ipsum dolor sit amet",
				FileAttachment: &fileAttachment{
					FileName:    "mail attachment",
					FileType:    "text/plain",
					FileURL:     "[[outlook=id3=aWQ9MDAwMDAwMDBDRjVENTQ1RjUzOTJERDQ1OER, Skattemeldingen kommer ]]",
					UploadState: "completed",
				},
			},
		},
	}

	hierachie, err := convertTodoistToVikunja(testSync, make(map[string]*doneItem))
	require.NoError(t, err)
	require.Len(t, hierachie, 2)
	require.Len(t, hierachie[1].Tasks, 1)
	assert.Empty(t, hierachie[1].Tasks[0].Attachments)
	assert.Equal(t, "Lorem Ipsum dolor sit amet", hierachie[1].Tasks[0].Description)
}

func TestIsDownloadableURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{name: "https", url: "https://todoist.com/file.md", want: true},
		{name: "http", url: "http://todoist.com/file.md", want: true},
		{name: "todoist mail attachment id", url: "[[outlook=id3=aWQ9MDAwMDAwMDBDRjVENTQ1RjUzOTJERDQ1OER]]", want: false},
		{name: "no scheme", url: "todoist.com/file.md", want: false},
		{name: "no host", url: "https://", want: false},
		{name: "other scheme", url: "file:///etc/passwd", want: false},
		{name: "empty", url: "", want: false},
		{name: "invalid", url: "https://%zz", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isDownloadableURL(tt.url))
		})
	}
}

func TestParseTodoistRepeat(t *testing.T) {
	tests := []struct {
		name     string
		due      *dueDate
		want     int64
		wantMode models.TaskRepeatMode
	}{
		{name: "nil due", due: nil, want: 0, wantMode: models.TaskRepeatModeDefault},
		{name: "not recurring", due: &dueDate{String: "every day", IsRecurring: false}, want: 0, wantMode: models.TaskRepeatModeDefault},

		{name: "every day", due: &dueDate{String: "every day", IsRecurring: true}, want: secondsPerDay, wantMode: models.TaskRepeatModeDefault},
		{name: "daily", due: &dueDate{String: "daily", IsRecurring: true}, want: secondsPerDay, wantMode: models.TaskRepeatModeDefault},
		{name: "every other day", due: &dueDate{String: "every other day", IsRecurring: true}, want: 2 * secondsPerDay, wantMode: models.TaskRepeatModeDefault},
		{name: "every 3 days", due: &dueDate{String: "every 3 days", IsRecurring: true}, want: 3 * secondsPerDay, wantMode: models.TaskRepeatModeDefault},
		{name: "every 1 day", due: &dueDate{String: "every 1 day", IsRecurring: true}, want: secondsPerDay, wantMode: models.TaskRepeatModeDefault},

		{name: "every week", due: &dueDate{String: "every week", IsRecurring: true}, want: secondsPerWeek, wantMode: models.TaskRepeatModeDefault},
		{name: "weekly", due: &dueDate{String: "weekly", IsRecurring: true}, want: secondsPerWeek, wantMode: models.TaskRepeatModeDefault},
		{name: "every other week", due: &dueDate{String: "every other week", IsRecurring: true}, want: 2 * secondsPerWeek, wantMode: models.TaskRepeatModeDefault},
		{name: "every 2 weeks", due: &dueDate{String: "every 2 weeks", IsRecurring: true}, want: 2 * secondsPerWeek, wantMode: models.TaskRepeatModeDefault},

		{name: "every month", due: &dueDate{String: "every month", IsRecurring: true}, want: secondsPerMonth, wantMode: models.TaskRepeatModeDefault},
		{name: "monthly", due: &dueDate{String: "monthly", IsRecurring: true}, want: secondsPerMonth, wantMode: models.TaskRepeatModeDefault},
		{name: "every 3 months", due: &dueDate{String: "every 3 months", IsRecurring: true}, want: 3 * secondsPerMonth, wantMode: models.TaskRepeatModeDefault},
		{name: "every 6 months", due: &dueDate{String: "every 6 months", IsRecurring: true}, want: 6 * secondsPerMonth, wantMode: models.TaskRepeatModeDefault},
		{name: "every 1 years", due: &dueDate{String: "every 1 years", IsRecurring: true}, want: secondsPerYear, wantMode: models.TaskRepeatModeDefault},

		{name: "every year", due: &dueDate{String: "every year", IsRecurring: true}, want: secondsPerYear, wantMode: models.TaskRepeatModeDefault},
		{name: "yearly", due: &dueDate{String: "yearly", IsRecurring: true}, want: secondsPerYear, wantMode: models.TaskRepeatModeDefault},
		{name: "annually", due: &dueDate{String: "annually", IsRecurring: true}, want: secondsPerYear, wantMode: models.TaskRepeatModeDefault},

		{name: "case insensitive", due: &dueDate{String: "Every Day", IsRecurring: true}, want: secondsPerDay, wantMode: models.TaskRepeatModeDefault},
		{name: "time of day stripped", due: &dueDate{String: "every day at 9am", IsRecurring: true}, want: secondsPerDay, wantMode: models.TaskRepeatModeDefault},

		// Dutch interval forms.
		{name: "elke dag", due: &dueDate{String: "elke dag", IsRecurring: true}, want: secondsPerDay, wantMode: models.TaskRepeatModeDefault},
		{name: "elke week", due: &dueDate{String: "elke week", IsRecurring: true}, want: secondsPerWeek, wantMode: models.TaskRepeatModeDefault},
		{name: "elke 4 weken", due: &dueDate{String: "elke 4 weken", IsRecurring: true}, want: 4 * secondsPerWeek, wantMode: models.TaskRepeatModeDefault},
		{name: "elke maand", due: &dueDate{String: "elke maand", IsRecurring: true}, want: secondsPerMonth, wantMode: models.TaskRepeatModeDefault},
		{name: "elke 3 maanden", due: &dueDate{String: "elke 3 maanden", IsRecurring: true}, want: 3 * secondsPerMonth, wantMode: models.TaskRepeatModeDefault},
		{name: "elke 6 maanden", due: &dueDate{String: "elke 6 maanden", IsRecurring: true}, want: 6 * secondsPerMonth, wantMode: models.TaskRepeatModeDefault},
		{name: "elke 24 maanden", due: &dueDate{String: "elke 24 maanden", IsRecurring: true}, want: 24 * secondsPerMonth, wantMode: models.TaskRepeatModeDefault},
		{name: "elke jaar", due: &dueDate{String: "elke jaar", IsRecurring: true}, want: secondsPerYear, wantMode: models.TaskRepeatModeDefault},
		{name: "elke 2 jaren", due: &dueDate{String: "elke 2 jaren", IsRecurring: true}, want: 2 * secondsPerYear, wantMode: models.TaskRepeatModeDefault},
		{name: "elke! 4 maand typo", due: &dueDate{String: "elke! 4 maand", IsRecurring: true}, want: 4 * secondsPerMonth, wantMode: models.TaskRepeatModeDefault},

		// A full date (day + month name) recurs yearly.
		{name: "yearly 1 April", due: &dueDate{String: "yearly 1 April", IsRecurring: true}, want: secondsPerYear, wantMode: models.TaskRepeatModeDefault},
		{name: "yearly 1st May", due: &dueDate{String: "yearly 1st May", IsRecurring: true}, want: secondsPerYear, wantMode: models.TaskRepeatModeDefault},
		{name: "yearly 21st September", due: &dueDate{String: "yearly 21st September", IsRecurring: true}, want: secondsPerYear, wantMode: models.TaskRepeatModeDefault},
		{name: "every 1 june", due: &dueDate{String: "every 1 june", IsRecurring: true}, want: secondsPerYear, wantMode: models.TaskRepeatModeDefault},
		{name: "elke 1 juni", due: &dueDate{String: "elke 1 juni", IsRecurring: true}, want: secondsPerYear, wantMode: models.TaskRepeatModeDefault},
		{name: "elke 1 maart", due: &dueDate{String: "elke 1 maart", IsRecurring: true}, want: secondsPerYear, wantMode: models.TaskRepeatModeDefault},
		{name: "elke 1 oktober", due: &dueDate{String: "elke 1 oktober", IsRecurring: true}, want: secondsPerYear, wantMode: models.TaskRepeatModeDefault},
		{name: "elke 10 jan", due: &dueDate{String: "elke 10 jan", IsRecurring: true}, want: secondsPerYear, wantMode: models.TaskRepeatModeDefault},
		{name: "elke 15 januari", due: &dueDate{String: "elke 15 januari", IsRecurring: true}, want: secondsPerYear, wantMode: models.TaskRepeatModeDefault},
		{name: "elke 16e jul", due: &dueDate{String: "elke 16e jul", IsRecurring: true}, want: secondsPerYear, wantMode: models.TaskRepeatModeDefault},

		// A numeric day/month date recurs yearly.
		{name: "every 15/03", due: &dueDate{String: "every 15/03", IsRecurring: true}, want: secondsPerYear, wantMode: models.TaskRepeatModeDefault},
		{name: "every 1/11", due: &dueDate{String: "every 1/11", IsRecurring: true}, want: secondsPerYear, wantMode: models.TaskRepeatModeDefault},

		// A day of the month without a month name recurs monthly, anchored via RepeatModeMonth.
		{name: "day of month", due: &dueDate{String: "every 27th", IsRecurring: true}, want: secondsPerMonth, wantMode: models.TaskRepeatModeMonth},
		{name: "every 15th", due: &dueDate{String: "every 15th", IsRecurring: true}, want: secondsPerMonth, wantMode: models.TaskRepeatModeMonth},
		{name: "every 10th", due: &dueDate{String: "every 10th", IsRecurring: true}, want: secondsPerMonth, wantMode: models.TaskRepeatModeMonth},
		{name: "every 1st day", due: &dueDate{String: "every 1st day", IsRecurring: true}, want: secondsPerMonth, wantMode: models.TaskRepeatModeMonth},
		{name: "elke laatste dag", due: &dueDate{String: "elke laatste dag", IsRecurring: true}, want: secondsPerMonth, wantMode: models.TaskRepeatModeMonth},

		// A weekday recurs weekly - the due date already anchors the weekday.
		{name: "specific weekday", due: &dueDate{String: "every monday", IsRecurring: true}, want: secondsPerWeek, wantMode: models.TaskRepeatModeDefault},
		{name: "dutch weekday", due: &dueDate{String: "elke maandag", IsRecurring: true}, want: secondsPerWeek, wantMode: models.TaskRepeatModeDefault},
		{name: "dutch weekday sunday", due: &dueDate{String: "elke zondag", IsRecurring: true}, want: secondsPerWeek, wantMode: models.TaskRepeatModeDefault},

		// Recurrences we can't represent (other languages, unparseable text) stay non-repeating.
		{name: "non-english", due: &dueDate{String: "cada día", IsRecurring: true}, want: 0, wantMode: models.TaskRepeatModeDefault},
		{name: "gibberish", due: &dueDate{String: "whenever", IsRecurring: true}, want: 0, wantMode: models.TaskRepeatModeDefault},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repeatAfter, repeatMode := parseTodoistRepeat(tt.due)
			assert.Equal(t, tt.want, repeatAfter)
			assert.Equal(t, tt.wantMode, repeatMode)
		})
	}
}

// The recurring due.string values of a real Todoist export (recurring_strings.json in the
// issue, not committed). Every one of them must migrate to a repeating task - exact
// interval and mode expectations are covered by the table test above.
func TestParseTodoistRepeatRealExportStrings(t *testing.T) {
	for _, s := range []string{
		"elke 1 juni",
		"elke 1 maart",
		"elke 1 oktober",
		"elke 10 jan",
		"elke 15 januari",
		"elke 16e jul",
		"elke 2 jaren",
		"elke 24 maanden",
		"elke 3 maanden",
		"elke 4 maanden",
		"elke 4 weken",
		"elke 6 maanden",
		"elke jaar",
		"elke laatste dag",
		"elke maandag",
		"elke week",
		"elke! 4 maand",
		"every 1 years",
		"every 1/11",
		"every 10th",
		"every 15/03",
		"every 15th",
		"every 1st day",
		"every 3 months",
		"every 6 months",
		"every month",
		"yearly",
		"yearly 1 April",
		"yearly 1 July",
		"yearly 1st May",
		"yearly 21st September",
	} {
		t.Run(s, func(t *testing.T) {
			repeatAfter, _ := parseTodoistRepeat(&dueDate{String: s, IsRecurring: true})
			assert.NotZero(t, repeatAfter, "recurrence from the real export must not migrate as non-repeating")
		})
	}
}
