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

package wekan

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertWekanToVikunja(t *testing.T) {
	startDate := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	dueDate := time.Date(2024, 2, 15, 10, 0, 0, 0, time.UTC)
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	commentDate := time.Date(2024, 1, 20, 14, 30, 0, 0, time.UTC)

	board := &wekanBoard{
		ID:    "board1",
		Title: "My Board",
		Labels: []wekanLabel{
			{ID: "lbl1", Name: "Bug", Color: "red"},
			{ID: "lbl2", Name: "Feature", Color: "green"},
			{ID: "lbl3", Name: "", Color: "blue"},
		},
		Lists: []wekanList{
			{ID: "list1", Title: "To Do", Sort: 1},
			{ID: "list2", Title: "In Progress", Sort: 2},
		},
		Cards: []wekanCard{
			{
				ID:          "card1",
				Title:       "Fix login bug",
				Description: "The login page crashes",
				ListID:      "list1",
				LabelIDs:    []string{"lbl1"},
				Sort:        1,
				StartAt:     &startDate,
				DueAt:       &dueDate,
				CreatedAt:   &createdAt,
			},
			{
				ID:       "card2",
				Title:    "Add dashboard",
				ListID:   "list2",
				LabelIDs: []string{"lbl2", "lbl3"},
				Sort:     1,
			},
			{
				ID:       "card3",
				Title:    "Archived task",
				ListID:   "list1",
				Sort:     2,
				Archived: true,
			},
		},
		Checklists: []wekanChecklist{
			{ID: "cl1", CardID: "card1", Title: "Steps to reproduce", Sort: 1},
		},
		ChecklistItems: []wekanChecklistItem{
			{ID: "cli1", ChecklistID: "cl1", CardID: "card1", Title: "Open browser", Sort: 1, IsFinished: true},
			{ID: "cli2", ChecklistID: "cl1", CardID: "card1", Title: "Click login", Sort: 2, IsFinished: false},
		},
		Comments: []wekanComment{
			{ID: "com1", Text: "This is urgent", CreatedAt: &commentDate, CardID: "card1"},
		},
	}

	result := convertWekanToVikunja(board)

	// Should have 1 project (the board itself)
	require.Len(t, result, 1)
	project := result[0]

	assert.Equal(t, "My Board", project.Title)

	// Should have 2 buckets (one per list)
	require.Len(t, project.Buckets, 2)
	assert.Equal(t, "To Do", project.Buckets[0].Title)
	assert.Equal(t, "In Progress", project.Buckets[1].Title)

	// Should have 3 tasks
	require.Len(t, project.Tasks, 3)

	// Task 1: Fix login bug
	task1 := project.Tasks[0]
	assert.Equal(t, "Fix login bug", task1.Title)
	assert.Contains(t, task1.Description, "The login page crashes")
	assert.Equal(t, startDate, task1.StartDate)
	assert.Equal(t, dueDate, task1.DueDate)
	// Should be in bucket 1 (To Do)
	assert.Equal(t, int64(1), task1.BucketID)

	// Labels on task 1
	require.Len(t, task1.Labels, 1)
	assert.Equal(t, "Bug", task1.Labels[0].Title)
	assert.Equal(t, "eb4646", task1.Labels[0].HexColor) // red mapped

	// Checklist should be appended to description as HTML task list
	assert.Contains(t, task1.Description, "Steps to reproduce")
	assert.Contains(t, task1.Description, "Open browser")
	assert.Contains(t, task1.Description, "Click login")
	assert.Contains(t, task1.Description, `checked="checked"`) // first item is done

	// Comment on task 1 (markdown converted to HTML)
	require.Len(t, task1.Comments, 1)
	assert.Contains(t, task1.Comments[0].Comment, "This is urgent")
	assert.Equal(t, commentDate, task1.Comments[0].Created)

	// Task 2: Add dashboard
	task2 := project.Tasks[1]
	assert.Equal(t, "Add dashboard", task2.Title)
	assert.Equal(t, int64(2), task2.BucketID) // In Progress
	require.Len(t, task2.Labels, 2)
	assert.Equal(t, "Feature", task2.Labels[0].Title)
	assert.Equal(t, "3cb500", task2.Labels[0].HexColor) // green mapped
	// Label with empty name but color "blue" should still be created with the color name as title
	assert.Equal(t, "blue", task2.Labels[1].Title)

	// Task 3: Archived → done
	task3 := project.Tasks[2]
	assert.Equal(t, "Archived task", task3.Title)
	assert.True(t, task3.Done)
	assert.Equal(t, int64(1), task3.BucketID) // To Do
}

func TestMigrateValidJSON(t *testing.T) {
	validJSON := `{
		"_id": "board1",
		"title": "Test Board",
		"labels": [{"_id": "l1", "name": "Bug", "color": "red"}],
		"lists": [{"_id": "list1", "title": "To Do", "sort": 1}],
		"cards": [{"_id": "c1", "title": "Task 1", "listId": "list1", "sort": 1, "labelIds": ["l1"]}],
		"checklists": [],
		"checklistItems": [],
		"comments": []
	}`

	m := &Migrator{}
	assert.Equal(t, "wekan", m.Name())

	// Test that parsing works without error
	r := bytes.NewReader([]byte(validJSON))
	board, err := parseWekanJSON(r)
	require.NoError(t, err)
	assert.Equal(t, "Test Board", board.Title)
	require.Len(t, board.Cards, 1)
	assert.Equal(t, "Task 1", board.Cards[0].Title)
}

func TestParseWekanJSONInvalid(t *testing.T) {
	_, err := parseWekanJSON(bytes.NewReader([]byte("not json")))
	require.Error(t, err)
}

func TestParseWekanJSONEmpty(t *testing.T) {
	_, err := parseWekanJSON(bytes.NewReader([]byte("")))
	require.Error(t, err)
}

func TestConvertWekanEmptyBoard(t *testing.T) {
	board := &wekanBoard{
		Title: "Empty Board",
	}

	result := convertWekanToVikunja(board)
	require.Len(t, result, 1)
	assert.Equal(t, "Empty Board", result[0].Title)
	assert.Empty(t, result[0].Tasks)
	assert.Empty(t, result[0].Buckets)
}

func TestConvertWekanCardWithoutList(t *testing.T) {
	// Card references a list that doesn't exist
	board := &wekanBoard{
		Title: "Board",
		Lists: []wekanList{{ID: "list1", Title: "To Do", Sort: 1}},
		Cards: []wekanCard{
			{ID: "c1", Title: "Orphan card", ListID: "nonexistent", Sort: 1},
			{ID: "c2", Title: "Normal card", ListID: "list1", Sort: 2},
		},
	}

	result := convertWekanToVikunja(board)
	require.Len(t, result, 1)
	// Orphan card should still be created, just with bucket ID 0
	require.Len(t, result[0].Tasks, 2)
	assert.Equal(t, int64(0), result[0].Tasks[0].BucketID)
	assert.Equal(t, int64(1), result[0].Tasks[1].BucketID)
}

func TestConvertWekanLabelColorMapping(t *testing.T) {
	board := &wekanBoard{
		Title: "Board",
		Labels: []wekanLabel{
			{ID: "l1", Name: "Urgent", Color: "red"},
			{ID: "l2", Name: "Nice", Color: "green"},
			{ID: "l3", Name: "Unknown color", Color: "chartreuse"},
		},
		Lists: []wekanList{{ID: "list1", Title: "List", Sort: 1}},
		Cards: []wekanCard{
			{ID: "c1", Title: "Task", ListID: "list1", LabelIDs: []string{"l1", "l2", "l3"}, Sort: 1},
		},
	}

	result := convertWekanToVikunja(board)
	labels := result[0].Tasks[0].Labels
	require.Len(t, labels, 3)
	assert.Equal(t, "eb4646", labels[0].HexColor) // red
	assert.Equal(t, "3cb500", labels[1].HexColor) // green
	assert.Empty(t, labels[2].HexColor)           // unknown color has no hex
}

func TestConvertWekanMultipleChecklists(t *testing.T) {
	board := &wekanBoard{
		Title: "Board",
		Lists: []wekanList{{ID: "list1", Title: "List", Sort: 1}},
		Cards: []wekanCard{{ID: "c1", Title: "Task", ListID: "list1", Sort: 1}},
		Checklists: []wekanChecklist{
			{ID: "cl1", CardID: "c1", Title: "Checklist A", Sort: 1},
			{ID: "cl2", CardID: "c1", Title: "Checklist B", Sort: 2},
		},
		ChecklistItems: []wekanChecklistItem{
			{ID: "i1", ChecklistID: "cl1", CardID: "c1", Title: "Item A1", Sort: 1, IsFinished: false},
			{ID: "i2", ChecklistID: "cl2", CardID: "c1", Title: "Item B1", Sort: 1, IsFinished: true},
		},
	}

	result := convertWekanToVikunja(board)
	desc := result[0].Tasks[0].Description
	assert.Contains(t, desc, "Checklist A")
	assert.Contains(t, desc, "Checklist B")
	assert.Contains(t, desc, "Item A1")
	assert.Contains(t, desc, "Item B1")
}

func TestParseWekanUnsupportedFieldsIgnored(t *testing.T) {
	// WeKan exports include fields we don't import (swimlanes, activities, rules, etc.).
	// Verify they are silently ignored and parsing succeeds.
	jsonWithExtras := `{
		"_id": "board1",
		"title": "Board With Extras",
		"labels": [],
		"lists": [{"_id": "list1", "title": "List", "sort": 1}],
		"cards": [{"_id": "c1", "title": "Task", "listId": "list1", "sort": 1}],
		"checklists": [],
		"checklistItems": [],
		"comments": [],
		"swimlanes": [{"_id": "sw1", "title": "Default"}],
		"activities": [{"_id": "act1", "activityType": "addComment"}],
		"rules": [{"_id": "rule1", "title": "Auto move"}],
		"triggers": [{"_id": "trig1", "activityType": "cardMove"}],
		"actions": [{"_id": "action1", "actionType": "moveCard"}],
		"customFields": [{"_id": "cf1", "name": "Priority", "type": "text"}]
	}`

	board, err := parseWekanJSON(bytes.NewReader([]byte(jsonWithExtras)))
	require.NoError(t, err)
	assert.Equal(t, "Board With Extras", board.Title)
	require.Len(t, board.Cards, 1)
	assert.Equal(t, "Task", board.Cards[0].Title)

	// Conversion should also work fine
	result := convertWekanToVikunja(board)
	require.Len(t, result, 1)
	assert.Equal(t, "Board With Extras", result[0].Title)
	require.Len(t, result[0].Tasks, 1)
}

func TestConvertWekanFromFixtureFile(t *testing.T) {
	file, err := os.Open("testdata_wekan_export.json")
	require.NoError(t, err)
	defer file.Close()

	board, err := parseWekanJSON(file)
	require.NoError(t, err)
	assert.Equal(t, "Sample Project Board", board.Title)

	result := convertWekanToVikunja(board)
	require.Len(t, result, 1)
	project := result[0]

	assert.Equal(t, "Sample Project Board", project.Title)
	require.Len(t, project.Buckets, 3)
	assert.Equal(t, "Backlog", project.Buckets[0].Title)
	assert.Equal(t, "In Progress", project.Buckets[1].Title)
	assert.Equal(t, "Done", project.Buckets[2].Title)

	require.Len(t, project.Tasks, 3)

	// Card 1 - has labels, checklist, comment, dates
	task1 := project.Tasks[0]
	assert.Equal(t, "Fix authentication flow", task1.Title)
	assert.Contains(t, task1.Description, "Users are getting logged out")
	assert.Contains(t, task1.Description, "Steps")
	assert.Contains(t, task1.Description, "Reproduce the issue")
	require.Len(t, task1.Labels, 2)
	require.Len(t, task1.Comments, 1)
	assert.False(t, task1.Done)

	// Card 3 - archived
	task3 := project.Tasks[2]
	assert.Equal(t, "Update README", task3.Title)
	assert.True(t, task3.Done)
}
