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

package models

import (
	"sort"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gopkg.in/d4l3k/messagediff.v1"
)

// To only run a selected tests: ^\QTestTaskCollection_ReadAll\E$/^\QReadAll_Tasks_with_range\E$

func TestTaskCollection_ReadAll(t *testing.T) {
	// Dummy users
	user1 := &user.User{
		ID:                           1,
		Username:                     "user1",
		Password:                     "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6",
		Issuer:                       "local",
		EmailRemindersEnabled:        true,
		OverdueTasksRemindersEnabled: true,
		OverdueTasksRemindersTime:    "09:00",
		Created:                      testCreatedTime,
		Updated:                      testUpdatedTime,
		ExportFileID:                 1,
	}
	user2 := &user.User{
		ID:                           2,
		Username:                     "user2",
		Password:                     "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6",
		Issuer:                       "local",
		EmailRemindersEnabled:        true,
		OverdueTasksRemindersEnabled: true,
		OverdueTasksRemindersTime:    "09:00",
		DefaultProjectID:             4,
		Created:                      testCreatedTime,
		Updated:                      testUpdatedTime,
	}
	user6 := &user.User{
		ID:                           6,
		Username:                     "user6",
		Password:                     "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6",
		Issuer:                       "local",
		EmailRemindersEnabled:        true,
		OverdueTasksRemindersEnabled: true,
		OverdueTasksRemindersTime:    "09:00",
		Created:                      testCreatedTime,
		Updated:                      testUpdatedTime,
	}
	linkShareUser2 := &user.User{
		ID:       -2,
		Name:     "Link Share",
		Username: "link-share-2",
		Created:  testCreatedTime,
		Updated:  testUpdatedTime,
	}

	loc := config.GetTimeZone()

	label4 := &Label{
		ID:          4,
		Title:       "Label #4 - visible via other task",
		CreatedByID: 2,
		CreatedBy:   user2,
		Created:     testCreatedTime,
		Updated:     testUpdatedTime,
	}
	label5 := &Label{
		ID:          5,
		Title:       "Label #5",
		CreatedByID: 2,
		CreatedBy:   user2,
		Created:     testCreatedTime,
		Updated:     testUpdatedTime,
	}

	// We use individual variables for the tasks here to be able to rearrange or remove ones more easily
	task1 := &Task{
		ID:          1,
		Title:       "task #1",
		Description: "Lorem Ipsum",
		Identifier:  "test1-1",
		Index:       1,
		CreatedByID: 1,
		CreatedBy:   user1,
		ProjectID:   1,
		IsFavorite:  true,
		Labels: []*Label{
			label4,
		},
		RelatedTasks: map[RelationKind][]*Task{
			RelationKindSubtask: {
				{
					ID:          29,
					Title:       "task #29 with parent task (1)",
					Index:       14,
					CreatedByID: 1,
					ProjectID:   1,
					Created:     time.Unix(1543626724, 0).In(loc),
					Updated:     time.Unix(1543626724, 0).In(loc),
				},
			},
		},
		Attachments: []*TaskAttachment{
			{
				ID:          1,
				TaskID:      1,
				FileID:      1,
				CreatedByID: 1,
				CreatedBy:   user1,
				Created:     testCreatedTime,
				File: &files.File{
					ID:          1,
					Name:        "test",
					Size:        100,
					Created:     time.Unix(1570998791, 0).In(loc),
					CreatedByID: 1,
				},
			},
			{
				ID:          2,
				TaskID:      1,
				FileID:      9999,
				CreatedByID: 1,
				CreatedBy:   user1,
				Created:     testCreatedTime,
			},
			{
				ID:          3,
				TaskID:      1,
				FileID:      1,
				CreatedByID: -2,
				CreatedBy:   linkShareUser2,
				Created:     testCreatedTime,
				File: &files.File{
					ID:          1,
					Name:        "test",
					Size:        100,
					Created:     time.Unix(1570998791, 0).In(loc),
					CreatedByID: 1,
				},
			},
		},
		Created: time.Unix(1543626724, 0).In(loc),
		Updated: time.Unix(1543626724, 0).In(loc),
	}
	var task1WithReaction = &Task{}
	*task1WithReaction = *task1
	task1WithReaction.Reactions = ReactionMap{
		"👋": []*user.User{user1},
	}
	task2 := &Task{
		ID:          2,
		Title:       "task #2 done",
		Identifier:  "test1-2",
		Index:       2,
		Done:        true,
		CreatedByID: 1,
		CreatedBy:   user1,
		ProjectID:   1,
		Labels: []*Label{
			label4,
		},
		RelatedTasks: map[RelationKind][]*Task{},
		Reminders: []*TaskReminder{
			{
				ID:       3,
				TaskID:   2,
				Reminder: time.Unix(1543626824, 0).In(loc),
				Created:  time.Unix(1543626724, 0).In(loc),
			},
		},
		Created: time.Unix(1543626724, 0).In(loc),
		Updated: time.Unix(1543626724, 0).In(loc),
	}
	task3 := &Task{
		ID:           3,
		Title:        "task #3 high prio",
		Identifier:   "test1-3",
		Index:        3,
		CreatedByID:  1,
		CreatedBy:    user1,
		ProjectID:    1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
		Priority:     100,
	}
	task4 := &Task{
		ID:           4,
		Title:        "task #4 low prio",
		Identifier:   "test1-4",
		Index:        4,
		CreatedByID:  1,
		CreatedBy:    user1,
		ProjectID:    1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
		Priority:     1,
	}
	task5 := &Task{
		ID:           5,
		Title:        "task #5 higher due date",
		Identifier:   "test1-5",
		Index:        5,
		CreatedByID:  1,
		CreatedBy:    user1,
		ProjectID:    1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
		DueDate:      time.Unix(1543636724, 0).In(loc),
	}
	task6 := &Task{
		ID:           6,
		Title:        "task #6 lower due date",
		Description:  "This has something unique",
		Identifier:   "test1-6",
		Index:        6,
		CreatedByID:  1,
		CreatedBy:    user1,
		ProjectID:    1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
		DueDate:      time.Unix(1543616724, 0).In(loc),
	}
	task7 := &Task{
		ID:           7,
		Title:        "task #7 with start date",
		Identifier:   "test1-7",
		Index:        7,
		CreatedByID:  1,
		CreatedBy:    user1,
		ProjectID:    1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
		StartDate:    time.Unix(1544600000, 0).In(loc),
	}
	task8 := &Task{
		ID:           8,
		Title:        "task #8 with end date",
		Identifier:   "test1-8",
		Index:        8,
		CreatedByID:  1,
		CreatedBy:    user1,
		ProjectID:    1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
		EndDate:      time.Unix(1544700000, 0).In(loc),
	}
	task9 := &Task{
		ID:           9,
		Title:        "task #9 with start and end date",
		Identifier:   "test1-9",
		Index:        9,
		CreatedByID:  1,
		CreatedBy:    user1,
		ProjectID:    1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
		StartDate:    time.Unix(1544600000, 0).In(loc),
		EndDate:      time.Unix(1544700000, 0).In(loc),
	}
	task10 := &Task{
		ID:           10,
		Title:        "task #10 basic",
		Identifier:   "test1-10",
		Index:        10,
		CreatedByID:  1,
		CreatedBy:    user1,
		ProjectID:    1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task11 := &Task{
		ID:           11,
		Title:        "task #11 basic",
		Identifier:   "test1-11",
		Index:        11,
		CreatedByID:  1,
		CreatedBy:    user1,
		ProjectID:    1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task12 := &Task{
		ID:           12,
		Title:        "task #12 basic",
		Identifier:   "test1-12",
		Index:        12,
		CreatedByID:  1,
		CreatedBy:    user1,
		ProjectID:    1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task15 := &Task{
		ID:           15,
		Title:        "task #15",
		Identifier:   "test6-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ProjectID:    6,
		IsFavorite:   true,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task16 := &Task{
		ID:           16,
		Title:        "task #16",
		Identifier:   "test7-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ProjectID:    7,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task17 := &Task{
		ID:           17,
		Title:        "task #17",
		Identifier:   "test8-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ProjectID:    8,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task18 := &Task{
		ID:           18,
		Title:        "task #18",
		Identifier:   "test9-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ProjectID:    9,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task19 := &Task{
		ID:           19,
		Title:        "task #19",
		Identifier:   "test10-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ProjectID:    10,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task20 := &Task{
		ID:           20,
		Title:        "task #20",
		Identifier:   "test11-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ProjectID:    11,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task21 := &Task{
		ID:           21,
		Title:        "task #21",
		Identifier:   "#1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ProjectID:    32, // parent project is shared to user 1 via direct share
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task22 := &Task{
		ID:           22,
		Title:        "task #22",
		Identifier:   "#1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ProjectID:    33,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task23 := &Task{
		ID:           23,
		Title:        "task #23",
		Identifier:   "#1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ProjectID:    34,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task24 := &Task{
		ID:           24,
		Title:        "task #24",
		Identifier:   "test15-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ProjectID:    15, // parent project is shared to user 1 via team
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task25 := &Task{
		ID:           25,
		Title:        "task #25",
		Identifier:   "test16-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ProjectID:    16,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task26 := &Task{
		ID:           26,
		Title:        "task #26",
		Identifier:   "test17-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ProjectID:    17,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task27 := &Task{
		ID:          27,
		Title:       "task #27 with reminders and start_date",
		Identifier:  "test1-12",
		Index:       12,
		CreatedByID: 1,
		CreatedBy:   user1,
		Reminders: []*TaskReminder{
			{
				ID:       1,
				TaskID:   27,
				Reminder: time.Unix(1543626724, 0).In(loc),
				Created:  time.Unix(1543626724, 0).In(loc),
			},
			{
				ID:             2,
				TaskID:         27,
				Reminder:       time.Unix(1543626824, 0).In(loc),
				Created:        time.Unix(1543626724, 0).In(loc),
				RelativePeriod: -3600,
				RelativeTo:     "start_date",
			},
		},
		StartDate:    time.Unix(1543616724, 0).In(loc),
		ProjectID:    1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task28 := &Task{
		ID:           28,
		Title:        "task #28 with repeat after",
		Identifier:   "test1-13",
		Index:        13,
		CreatedByID:  1,
		CreatedBy:    user1,
		ProjectID:    1,
		RelatedTasks: map[RelationKind][]*Task{},
		RepeatAfter:  3600,
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task29 := &Task{
		ID:          29,
		Title:       "task #29 with parent task (1)",
		Identifier:  "test1-14",
		Index:       14,
		CreatedByID: 1,
		CreatedBy:   user1,
		ProjectID:   1,
		RelatedTasks: map[RelationKind][]*Task{
			RelationKindParenttask: {
				{
					ID:          1,
					Title:       "task #1",
					Description: "Lorem Ipsum",
					Index:       1,
					CreatedByID: 1,
					ProjectID:   1,
					IsFavorite:  true,
					Created:     time.Unix(1543626724, 0).In(loc),
					Updated:     time.Unix(1543626724, 0).In(loc),
				},
			},
		},
		Created: time.Unix(1543626724, 0).In(loc),
		Updated: time.Unix(1543626724, 0).In(loc),
	}
	task30 := &Task{
		ID:          30,
		Title:       "task #30 with assignees",
		Identifier:  "test1-15",
		Index:       15,
		CreatedByID: 1,
		CreatedBy:   user1,
		ProjectID:   1,
		Assignees: []*user.User{
			user1,
			user2,
		},
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task31 := &Task{
		ID:           31,
		Title:        "task #31 with color",
		Identifier:   "test1-16",
		Index:        16,
		HexColor:     "f0f0f0",
		CreatedByID:  1,
		CreatedBy:    user1,
		ProjectID:    1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task32 := &Task{
		ID:           32,
		Title:        "task #32",
		Identifier:   "test3-1",
		Index:        1,
		CreatedByID:  1,
		CreatedBy:    user1,
		ProjectID:    3,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task33 := &Task{
		ID:           33,
		Title:        "task #33 with percent done",
		Identifier:   "test1-17",
		Index:        17,
		CreatedByID:  1,
		CreatedBy:    user1,
		ProjectID:    1,
		PercentDone:  0.5,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task35 := &Task{
		ID:          35,
		Title:       "task #35",
		Identifier:  "test21-1",
		Index:       1,
		CreatedByID: 1,
		CreatedBy:   user1,
		ProjectID:   21,
		Assignees: []*user.User{
			user2,
		},
		Labels: []*Label{
			label4,
			label5,
		},
		RelatedTasks: map[RelationKind][]*Task{
			RelationKindRelated: {
				{
					ID:          1,
					Title:       "task #1",
					Description: "Lorem Ipsum",
					Index:       1,
					CreatedByID: 1,
					ProjectID:   1,
					IsFavorite:  true,
					Created:     time.Unix(1543626724, 0).In(loc),
					Updated:     time.Unix(1543626724, 0).In(loc),
				},
				{
					ID:          1,
					Title:       "task #1",
					Description: "Lorem Ipsum",
					Index:       1,
					CreatedByID: 1,
					ProjectID:   1,
					IsFavorite:  true,
					Created:     time.Unix(1543626724, 0).In(loc),
					Updated:     time.Unix(1543626724, 0).In(loc),
				},
			},
		},
		Created: time.Unix(1543626724, 0).In(loc),
		Updated: time.Unix(1543626724, 0).In(loc),
	}
	task39 := &Task{
		ID:           39,
		Title:        "task #39",
		Identifier:   "#0",
		CreatedByID:  1,
		CreatedBy:    user1,
		ProjectID:    25,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}

	type fields struct {
		ProjectID     int64
		ProjectViewID int64
		Projects      []*Project
		SortBy        []string // Is a string, since this is the place where a query string comes from the user
		OrderBy       []string

		FilterIncludeNulls bool
		Filter             string

		Expand []TaskCollectionExpandable

		CRUDable    web.CRUDable
		Permissions web.Permissions
	}
	type args struct {
		search string
		a      web.Auth
		page   int
	}
	type testcase struct {
		name    string
		fields  fields
		args    args
		want    []*Task
		wantErr bool
	}

	defaultArgs := args{
		search: "",
		a:      &user.User{ID: 1},
		page:   0,
	}

	taskWithPosition := func(task *Task, position float64) *Task {
		newTask := &Task{}
		*newTask = *task
		newTask.Position = position
		return newTask
	}

	tests := []testcase{
		{
			name:   "ReadAll Tasks normally",
			fields: fields{},
			args:   defaultArgs,
			want: []*Task{
				task1,
				task2,
				task3,
				task4,
				task5,
				task6,
				task7,
				task8,
				task9,
				task10,
				task11,
				task12,
				task15,
				task16,
				task17,
				task18,
				task19,
				task20,
				task21,
				task22,
				task23,
				task24,
				task25,
				task26,
				task27,
				task28,
				task29,
				task30,
				task31,
				task32,
				task33,
				task35,
				task39,
			},
			wantErr: false,
		},
		{
			name: "ReadAll Tasks with expanded reaction",
			fields: fields{
				Expand: []TaskCollectionExpandable{
					TaskCollectionExpandReactions,
				},
			},
			args: defaultArgs,
			want: []*Task{
				task1WithReaction,
				task2,
				task3,
				task4,
				task5,
				task6,
				task7,
				task8,
				task9,
				task10,
				task11,
				task12,
				task15,
				task16,
				task17,
				task18,
				task19,
				task20,
				task21,
				task22,
				task23,
				task24,
				task25,
				task26,
				task27,
				task28,
				task29,
				task30,
				task31,
				task32,
				task33,
				task35,
				task39,
			},
			wantErr: false,
		},
		{
			// For more sorting tests see task_collection_sort_test.go
			name: "sorted by done asc and id desc",
			fields: fields{
				SortBy:  []string{"done", "id"},
				OrderBy: []string{"asc", "desc"},
			},
			args: defaultArgs,
			want: []*Task{
				task35,
				task33,
				task32,
				task31,
				task30,
				task29,
				task28,
				task27,
				task26,
				task25,
				task24,
				task23,
				task22,
				task21,
				task20,
				task19,
				task18,
				task17,
				task16,
				task15,
				task12,
				task11,
				task10,
				task9,
				task8,
				task7,
				task6,
				task5,
				task4,
				task3,
				task1,
				task2,
				task39,
			},
			wantErr: false,
		},
		{
			name: "ReadAll Tasks with range",
			fields: fields{
				Filter: "start_date > '2018-12-11T03:46:40+00:00' || end_date < '2018-12-13T11:20:01+00:00'",
			},
			args: defaultArgs,
			want: []*Task{
				task7,
				task8,
				task9,
			},
			wantErr: false,
		},
		{
			name: "ReadAll Tasks with different range",
			fields: fields{
				Filter: "start_date > '2018-12-13T11:20:00+00:00' || end_date < '2018-12-16T22:40:00+00:00'",
			},
			args: defaultArgs,
			want: []*Task{
				task8,
				task9,
			},
			wantErr: false,
		},
		{
			name: "ReadAll Tasks with range with start date only",
			fields: fields{
				Filter: "start_date > '2018-12-12T07:33:20+00:00'",
			},
			args:    defaultArgs,
			want:    []*Task{},
			wantErr: false,
		},
		{
			name: "ReadAll Tasks with range with start date only between",
			fields: fields{
				Filter: "start_date > '2018-12-12T00:00:00+00:00' && start_date < '2018-12-13T00:00:00+00:00'",
			},
			args: defaultArgs,
			want: []*Task{
				task7,
				task9,
			},
			wantErr: false,
		},
		{
			name: "ReadAll Tasks with range with start date only and greater equals",
			fields: fields{
				Filter: "start_date >= '2018-12-12T07:33:20+00:00'",
			},
			args: defaultArgs,
			want: []*Task{
				task7,
				task9,
			},
			wantErr: false,
		},
		{
			name: "range and nesting",
			fields: fields{
				Filter: "(start_date > '2018-12-12T00:00:00+00:00' && start_date < '2018-12-13T00:00:00+00:00') || end_date > '2018-12-13T00:00:00+00:00'",
			},
			args: defaultArgs,
			want: []*Task{
				task7,
				task8,
				task9,
			},
			wantErr: false,
		},
		{
			name: "undone tasks only",
			fields: fields{
				Filter: "done = false",
			},
			args: defaultArgs,
			want: []*Task{
				task1,
				// Task 2 is done
				task3,
				task4,
				task5,
				task6,
				task7,
				task8,
				task9,
				task10,
				task11,
				task12,
				task15,
				task16,
				task17,
				task18,
				task19,
				task20,
				task21,

				task22,
				task23,
				task24,
				task25,
				task26,

				task27,
				task28,
				task29,
				task30,
				task31,
				task32,
				task33,
				task35,
			},
			wantErr: false,
		},
		{
			name: "done tasks only",
			fields: fields{
				Filter: "done = true",
			},
			args: defaultArgs,
			want: []*Task{
				task2,
			},
			wantErr: false,
		},
		{
			name: "done tasks only - not equals done",
			fields: fields{
				Filter: "done != false",
			},
			args: defaultArgs,
			want: []*Task{
				task2,
			},
			wantErr: false,
		},
		{
			name: "range with nulls",
			fields: fields{
				FilterIncludeNulls: true,
				Filter:             "start_date > '2018-12-11T03:46:40+00:00' || end_date < '2018-12-13T11:20:01+00:00'",
			},
			args: defaultArgs,
			want: []*Task{
				task1, // has nil dates
				task2, // has nil dates
				task3, // has nil dates
				task4, // has nil dates
				task5, // has nil dates
				task6, // has nil dates
				task7,
				task8,
				task9,
				task10, // has nil dates
				task11, // has nil dates
				task12, // has nil dates
				task15, // has nil dates
				task16, // has nil dates
				task17, // has nil dates
				task18, // has nil dates
				task19, // has nil dates
				task20, // has nil dates
				task21, // has nil dates
				task22, // has nil dates
				task23, // has nil dates
				task24, // has nil dates
				task25, // has nil dates
				task26, // has nil dates
				task27, // has nil dates
				task28, // has nil dates
				task29, // has nil dates
				task30, // has nil dates
				task31, // has nil dates
				task32, // has nil dates
				task33, // has nil dates
				task35, // has nil dates
				task39, // has nil dates
			},
			wantErr: false,
		},
		{
			name: "favorited tasks",
			args: defaultArgs,
			fields: fields{
				ProjectID: FavoritesPseudoProject.ID,
			},
			want: []*Task{
				task1,
				task15,
				// Task 34 is also a favorite, but on a project user 1 has no access to.
			},
		},
		{
			name: "filtered with like",
			fields: fields{
				Filter: "title ~ with",
			},
			args: defaultArgs,
			want: []*Task{
				task7,
				task8,
				task9,
				task27,
				task28,
				task29,
				task30,
				task31,
				task33,
			},
			wantErr: false,
		},
		{
			name: "filtered with like and '",
			fields: fields{
				Filter: "title ~ 'with'",
			},
			args: defaultArgs,
			want: []*Task{
				task7,
				task8,
				task9,
				task27,
				task28,
				task29,
				task30,
				task31,
				task33,
			},
			wantErr: false,
		},
		{
			name: "filtered reminder dates",
			fields: fields{
				Filter: "reminders > '2018-10-01T00:00:00+00:00' && reminders < '2018-12-10T00:00:00+00:00'",
			},
			args: defaultArgs,
			want: []*Task{
				task2,
				task27,
			},
			wantErr: false,
		},
		{
			name: "filter in keyword",
			fields: fields{
				Filter: "id in '1,2,34'", // user does not have permission to access task 34
			},
			args: defaultArgs,
			want: []*Task{
				task1,
				task2,
			},
			wantErr: false,
		},
		{
			name: "filter in keyword without quotes",
			fields: fields{
				Filter: "id in 1,2,34", // user does not have permission to access task 34
			},
			args: defaultArgs,
			want: []*Task{
				task1,
				task2,
			},
			wantErr: false,
		},
		{
			name: "filter in",
			fields: fields{
				Filter: "id ?= '1,2,34'", // user does not have permission to access task 34
			},
			args: defaultArgs,
			want: []*Task{
				task1,
				task2,
			},
			wantErr: false,
		},
		{
			name: "filter not in",
			fields: fields{
				Filter: "id not in '1,2,3,4'",
			},
			args: defaultArgs,
			want: []*Task{
				task5,
				task6,
				task7,
				task8,
				task9,
				task10,
				task11,
				task12,
				task15,
				task16,
				task17,
				task18,
				task19,
				task20,
				task21,
				task22,
				task23,
				task24,
				task25,
				task26,
				task27,
				task28,
				task29,
				task30,
				task31,
				task32,
				task33,
				task35,
				task39,
			},
			wantErr: false,
		},
		{
			name: "filter assignees by username",
			fields: fields{
				Filter: "assignees = 'user1'",
			},
			args: defaultArgs,
			want: []*Task{
				task30,
			},
			wantErr: false,
		},
		{
			name: "filter assignees by username with users field name",
			fields: fields{
				Filter: "users = 'user1'",
			},
			args:    defaultArgs,
			want:    nil,
			wantErr: true,
		},
		{
			name: "filter assignees by username with user_id field name",
			fields: fields{
				Filter: "user_id = 'user1'",
			},
			args:    defaultArgs,
			want:    nil,
			wantErr: true,
		},
		{
			name: "filter assignees by multiple username",
			fields: fields{
				Filter: "assignees = 'user1' || assignees = 'user2'",
			},
			args: defaultArgs,
			want: []*Task{
				task30,
				task35,
			},
			wantErr: false,
		},
		{
			name: "filter assignees by numbers",
			fields: fields{
				Filter: "assignees = 1",
			},
			args:    defaultArgs,
			want:    []*Task{},
			wantErr: false,
		},
		{
			name: "filter assignees by name with like",
			fields: fields{
				Filter: "assignees ~ 'user'",
			},
			args: defaultArgs,
			want: []*Task{
				// Same as without any filter since the filter is ignored
				task1,
				task2,
				task3,
				task4,
				task5,
				task6,
				task7,
				task8,
				task9,
				task10,
				task11,
				task12,
				task15,
				task16,
				task17,
				task18,
				task19,
				task20,
				task21,
				task22,
				task23,
				task24,
				task25,
				task26,
				task27,
				task28,
				task29,
				task30,
				task31,
				task32,
				task33,
				task35,
				task39,
			},
			wantErr: false,
		},
		{
			name: "filter assignees in by id",
			fields: fields{
				Filter: "assignees ?= '1,2'",
			},
			args:    defaultArgs,
			want:    []*Task{},
			wantErr: false,
		},
		{
			name: "filter assignees in by username",
			fields: fields{
				Filter: "assignees ?= 'user1,user2'",
			},
			args: defaultArgs,
			want: []*Task{
				task30,
				task35,
			},
			wantErr: false,
		},
		{
			name: "filter labels",
			fields: fields{
				Filter: "labels = 4",
			},
			args: defaultArgs,
			want: []*Task{
				task1,
				task2,
				task35,
			},
			wantErr: false,
		},
		{
			name: "filter labels with nulls",
			fields: fields{
				Filter:             "labels = 5",
				FilterIncludeNulls: true,
			},
			args: defaultArgs,
			want: []*Task{
				task3,
				task4,
				task5,
				task6,
				task7,
				task8,
				task9,
				task10,
				task11,
				task12,
				task15,
				task16,
				task17,
				task18,
				task19,
				task20,
				task21,
				task22,
				task23,
				task24,
				task25,
				task26,
				task27,
				task28,
				task29,
				task30,
				task31,
				task32,
				task33,
				task35,
				task39,
			},
			wantErr: false,
		},
		{
			name: "filter labels not eq",
			fields: fields{
				Filter: "labels != 5",
			},
			args: defaultArgs,
			want: []*Task{
				task1,
				task2,
				task3,
				task4,
				task5,
				task6,
				task7,
				task8,
				task9,
				task10,
				task11,
				task12,
				task15,
				task16,
				task17,
				task18,
				task19,
				task20,
				task21,
				task22,
				task23,
				task24,
				task25,
				task26,
				task27,
				task28,
				task29,
				task30,
				task31,
				task32,
				task33,
				//task35,
				// task 35 has a label 5 and 4
				task39,
			},
			wantErr: false,
		},
		{
			name: "filter labels not in",
			fields: fields{
				Filter: "labels not in 5",
			},
			args: defaultArgs,
			want: []*Task{
				task1,
				task2,
				task3,
				task4,
				task5,
				task6,
				task7,
				task8,
				task9,
				task10,
				task11,
				task12,
				task15,
				task16,
				task17,
				task18,
				task19,
				task20,
				task21,
				task22,
				task23,
				task24,
				task25,
				task26,
				task27,
				task28,
				task29,
				task30,
				task31,
				task32,
				task33,
				//task35,
				// task 35 has a label 5 and 4
				task39,
			},
			wantErr: false,
		},
		{
			name: "filter project_id",
			fields: fields{
				Filter: "project_id = 6",
			},
			args: defaultArgs,
			want: []*Task{
				task15,
			},
			wantErr: false,
		},
		{
			name: "filter project",
			fields: fields{
				Filter: "project = 6",
			},
			args: defaultArgs,
			want: []*Task{
				task15,
			},
			wantErr: false,
		},
		{
			name: "filter project forbidden",
			fields: fields{
				Filter: "project_id = 20", // user1 has no access to project 20
			},
			args:    defaultArgs,
			want:    []*Task{},
			wantErr: false,
		},
		// TODO filter parent project?
		{
			name: "filter by index",
			fields: fields{
				Filter: "index = 5",
			},
			args: defaultArgs,
			want: []*Task{
				task5,
			},
			wantErr: false,
		},
		{
			name: "order by position",
			fields: fields{
				SortBy:        []string{"position", "id"},
				OrderBy:       []string{"asc", "asc"},
				ProjectViewID: 1,
				ProjectID:     1,
			},
			args: args{
				a: &user.User{ID: 1},
			},
			want: []*Task{
				// The only tasks with a position set
				taskWithPosition(task1, 2),
				taskWithPosition(task2, 4),
				// the other ones don't have a position set
				task3,
				task4,
				task5,
				task6,
				task7,
				task8,
				task9,
				task10,
				task11,
				task12,
				//task15,
				//task16,
				//task17,
				//task18,
				//task19,
				//task20,
				//task21,
				//task22,
				//task23,
				//task24,
				//task25,
				//task26,
				task27,
				task28,
				task29,
				task30,
				task31,
				task33,
			},
		},
		{
			name: "order by due date",
			fields: fields{
				SortBy:  []string{"due_date", "id"},
				OrderBy: []string{"asc", "desc"},
			},
			args: args{
				a: &user.User{ID: 1},
			},
			want: []*Task{
				// The only tasks with a due date
				task6,
				task5,
				// The other ones don't have a due date
				task39,
				task35,
				task33,
				task32,
				task31,
				task30,
				task29,
				task28,
				task27,
				task26,
				task25,
				task24,
				task23,
				task22,
				task21,
				task20,
				task19,
				task18,
				task17,
				task16,
				task15,
				task12,
				task11,
				task10,
				task9,
				task8,
				task7,
				task4,
				task3,
				task2,
				task1,
			},
		},
		{
			name: "saved filter with sort order",
			fields: fields{
				ProjectID: -2,
				SortBy:    []string{"title", "id"},
				OrderBy:   []string{"desc", "asc"},
			},
			args: args{
				a: &user.User{ID: 1},
			},
			want: []*Task{
				task9,
				task8,
				task7,
				task6,
				task5,
			},
		},
		{
			name: "saved filter with sort order asc",
			fields: fields{
				ProjectID: -2,
				SortBy:    []string{"title", "id"},
				OrderBy:   []string{"asc", "asc"},
			},
			args: args{
				a: &user.User{ID: 1},
			},
			want: []*Task{
				task5,
				task6,
				task7,
				task8,
				task9,
			},
		},
		{
			name: "saved filter with sort by due date",
			fields: fields{
				ProjectID: -2,
				SortBy:    []string{"due_date", "id"},
				OrderBy:   []string{"asc", "asc"},
			},
			args: args{
				a: &user.User{ID: 1},
			},
			want: []*Task{
				task6,
				task5,
				task7,
				task8,
				task9,
			},
		},
		// TODO unix dates
		// TODO date magic
	}

	// Here we're explicitly testing search with and without paradeDB. Both return different results but that's
	// expected - paradeDB returns more results than other databases with a naive like-search.

	if db.ParadeDBAvailable() {
		tests = append(tests, testcase{
			name:   "search for task index",
			fields: fields{},
			args: args{
				search: "number #17",
				a:      &user.User{ID: 1},
				page:   0,
			},
			want: []*Task{
				task17, // has the text #17 in the title
				task33, // has the index 17
			},
			wantErr: false,
		})
	} else {
		tests = append(tests, testcase{
			name:   "search for task index",
			fields: fields{},
			args: args{
				search: "number #17",
				a:      &user.User{ID: 1},
				page:   0,
			},
			want: []*Task{
				task33, // has the index 17
			},
			wantErr: false,
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			lt := &TaskCollection{
				ProjectID:     tt.fields.ProjectID,
				ProjectViewID: tt.fields.ProjectViewID,
				SortBy:        tt.fields.SortBy,
				OrderBy:       tt.fields.OrderBy,

				FilterIncludeNulls: tt.fields.FilterIncludeNulls,

				Filter: tt.fields.Filter,

				Expand: tt.fields.Expand,

				CRUDable:    tt.fields.CRUDable,
				Permissions: tt.fields.Permissions,
			}
			got, _, _, err := lt.ReadAll(s, tt.args.a, tt.args.search, tt.args.page, 50)
			if (err != nil) != tt.wantErr {
				t.Errorf("Test %s, Task.ReadAll() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if diff, equal := messagediff.PrettyDiff(tt.want, got); !equal {
				var is bool
				var gotTasks []*Task
				gotTasks, is = got.([]*Task)
				if !is {
					gotTasks = []*Task{}
				}
				if len(gotTasks) == 0 && len(tt.want) == 0 {
					return
				}

				gotIDs := []int64{}
				for _, t := range got.([]*Task) {
					gotIDs = append(gotIDs, t.ID)
				}

				wantIDs := []int64{}
				for _, t := range tt.want {
					wantIDs = append(wantIDs, t.ID)
				}
				sort.Slice(wantIDs, func(i, j int) bool {
					return wantIDs[i] < wantIDs[j]
				})
				sort.Slice(gotIDs, func(i, j int) bool {
					return gotIDs[i] < gotIDs[j]
				})

				diffIDs, _ := messagediff.PrettyDiff(wantIDs, gotIDs)

				t.Errorf("Test %s, Task.ReadAll() = %v, \nwant %v, \ndiff: %v \n\n diffIDs: %v", tt.name, got, tt.want, diff, diffIDs)
			}
		})
	}
}

func TestTaskCollection_SubtaskRemainsAfterMove(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{ID: 1}

	c := &TaskCollection{
		ProjectID: 1,
		Expand:    []TaskCollectionExpandable{TaskCollectionExpandSubtasks},
	}

	res, _, _, err := c.ReadAll(s, u, "", 0, 50)
	require.NoError(t, err)
	tasks, ok := res.([]*Task)
	require.True(t, ok)

	found := false
	for _, tsk := range tasks {
		if tsk.ID == 29 {
			found = true
			break
		}
	}
	assert.True(t, found, "subtask should be returned before moving")

	subtask := &Task{ID: 29, ProjectID: 7}
	err = subtask.Update(s, u)
	require.NoError(t, err)
	require.NoError(t, s.Commit())

	s2 := db.NewSession()
	defer s2.Close()
	c = &TaskCollection{
		ProjectID: 7,
		Expand:    []TaskCollectionExpandable{TaskCollectionExpandSubtasks},
	}

	res, _, _, err = c.ReadAll(s2, u, "", 0, 50)
	require.NoError(t, err)
	tasks, ok = res.([]*Task)
	require.True(t, ok)

	found = false
	for _, tsk := range tasks {
		if tsk.ID == 29 {
			found = true
			break
		}
	}
	assert.True(t, found, "subtask should be returned after moving to another project")
}

func TestTaskSearchWithExpandSubtasks(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	project, err := GetProjectSimpleByID(s, 36)
	require.NoError(t, err)

	opts := &taskSearchOptions{
		search: "Caldav",
		expand: []TaskCollectionExpandable{TaskCollectionExpandSubtasks},
	}

	tasks, _, _, err := getRawTasksForProjects(s, []*Project{project}, &user.User{ID: 15}, opts)
	require.NoError(t, err)
	require.NotEmpty(t, tasks)
}

func TestTaskCollection_SubtaskWithMultipleParentsNoDuplicates(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{ID: 15}

	// Use existing tasks from fixtures:
	// - Task 41: Parent task in project 36 (already exists)
	// - Task 42: Another parent task in project 36 (already exists)
	// - Task 43: Subtask in project 36 (already a subtask of task 41)

	// Add a second parent relationship: task 43 -> task 42
	// This will make task 43 have multiple parents (task 41 and task 42)
	relation := &TaskRelation{
		TaskID:       43, // subtask
		OtherTaskID:  42, // second parent
		RelationKind: RelationKindParenttask,
		CreatedByID:  15,
	}
	_, err := s.Insert(relation)
	require.NoError(t, err)

	// Create inverse relation: task 42 -> task 43
	inverseRelation := &TaskRelation{
		TaskID:       42, // second parent
		OtherTaskID:  43, // subtask
		RelationKind: RelationKindSubtask,
		CreatedByID:  15,
	}
	_, err = s.Insert(inverseRelation)
	require.NoError(t, err)

	// Test Project 36 - should include tasks 41, 42, and 43, but task 43 should only appear once
	c := &TaskCollection{
		ProjectID: 36,
		Expand:    []TaskCollectionExpandable{TaskCollectionExpandSubtasks},
	}

	res, _, _, err := c.ReadAll(s, u, "", 0, 50)
	require.NoError(t, err)
	tasks, ok := res.([]*Task)
	require.True(t, ok)

	// Count how many times task 43 (the subtask) appears
	subtaskCount := 0
	for _, task := range tasks {
		if task.ID == 43 {
			subtaskCount++
		}
	}

	// The subtask should appear exactly once (as a subtask, not as a standalone task)
	assert.Equal(t, 1, subtaskCount, "Subtask should appear exactly once in Project 36")

	// Verify that both parent tasks are present
	foundParent1 := false
	foundParent2 := false
	for _, task := range tasks {
		if task.ID == 41 {
			foundParent1 = true
		}
		if task.ID == 42 {
			foundParent2 = true
		}
	}
	assert.True(t, foundParent1, "Parent task 41 should be present")
	assert.True(t, foundParent2, "Parent task 42 should be present")
}
