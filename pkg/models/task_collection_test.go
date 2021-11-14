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

package models

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"
	"gopkg.in/d4l3k/messagediff.v1"
)

func TestTaskCollection_ReadAll(t *testing.T) {
	// Dummy users
	user1 := &user.User{
		ID:                           1,
		Username:                     "user1",
		Password:                     "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Issuer:                       "local",
		EmailRemindersEnabled:        true,
		OverdueTasksRemindersEnabled: true,
		Created:                      testCreatedTime,
		Updated:                      testUpdatedTime,
	}
	user2 := &user.User{
		ID:                           2,
		Username:                     "user2",
		Password:                     "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Issuer:                       "local",
		EmailRemindersEnabled:        true,
		OverdueTasksRemindersEnabled: true,
		Created:                      testCreatedTime,
		Updated:                      testUpdatedTime,
	}
	user6 := &user.User{
		ID:                           6,
		Username:                     "user6",
		Password:                     "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Issuer:                       "local",
		EmailRemindersEnabled:        true,
		OverdueTasksRemindersEnabled: true,
		Created:                      testCreatedTime,
		Updated:                      testUpdatedTime,
	}
	linkShareUser2 := &user.User{
		ID:      -2,
		Name:    "Link Share",
		Created: testCreatedTime,
		Updated: testUpdatedTime,
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

	// We use individual variables for the tasks here to be able to rearrange or remove ones more easily
	task1 := &Task{
		ID:          1,
		Title:       "task #1",
		Description: "Lorem Ipsum",
		Identifier:  "test1-1",
		Index:       1,
		CreatedByID: 1,
		CreatedBy:   user1,
		ListID:      1,
		BucketID:    1,
		IsFavorite:  true,
		Position:    2,
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
					ListID:      1,
					BucketID:    1,
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
	task2 := &Task{
		ID:          2,
		Title:       "task #2 done",
		Identifier:  "test1-2",
		Index:       2,
		Done:        true,
		CreatedByID: 1,
		CreatedBy:   user1,
		ListID:      1,
		BucketID:    1,
		Position:    4,
		Labels: []*Label{
			label4,
		},
		RelatedTasks: map[RelationKind][]*Task{},
		Reminders: []time.Time{
			time.Unix(1543626824, 0).In(loc),
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
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
		Priority:     100,
		BucketID:     2,
	}
	task4 := &Task{
		ID:           4,
		Title:        "task #4 low prio",
		Identifier:   "test1-4",
		Index:        4,
		CreatedByID:  1,
		CreatedBy:    user1,
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
		Priority:     1,
		BucketID:     2,
	}
	task5 := &Task{
		ID:           5,
		Title:        "task #5 higher due date",
		Identifier:   "test1-5",
		Index:        5,
		CreatedByID:  1,
		CreatedBy:    user1,
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
		DueDate:      time.Unix(1543636724, 0).In(loc),
		BucketID:     2,
	}
	task6 := &Task{
		ID:           6,
		Title:        "task #6 lower due date",
		Identifier:   "test1-6",
		Index:        6,
		CreatedByID:  1,
		CreatedBy:    user1,
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
		DueDate:      time.Unix(1543616724, 0).In(loc),
		BucketID:     3,
	}
	task7 := &Task{
		ID:           7,
		Title:        "task #7 with start date",
		Identifier:   "test1-7",
		Index:        7,
		CreatedByID:  1,
		CreatedBy:    user1,
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
		StartDate:    time.Unix(1544600000, 0).In(loc),
		BucketID:     3,
	}
	task8 := &Task{
		ID:           8,
		Title:        "task #8 with end date",
		Identifier:   "test1-8",
		Index:        8,
		CreatedByID:  1,
		CreatedBy:    user1,
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
		EndDate:      time.Unix(1544700000, 0).In(loc),
		BucketID:     3,
	}
	task9 := &Task{
		ID:           9,
		Title:        "task #9 with start and end date",
		Identifier:   "test1-9",
		Index:        9,
		CreatedByID:  1,
		CreatedBy:    user1,
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		BucketID:     1,
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
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		BucketID:     1,
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
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		BucketID:     1,
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
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		BucketID:     1,
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
		ListID:       6,
		IsFavorite:   true,
		RelatedTasks: map[RelationKind][]*Task{},
		BucketID:     6,
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
		ListID:       7,
		RelatedTasks: map[RelationKind][]*Task{},
		BucketID:     7,
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
		ListID:       8,
		RelatedTasks: map[RelationKind][]*Task{},
		BucketID:     8,
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
		ListID:       9,
		RelatedTasks: map[RelationKind][]*Task{},
		BucketID:     9,
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
		ListID:       10,
		RelatedTasks: map[RelationKind][]*Task{},
		BucketID:     10,
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
		ListID:       11,
		RelatedTasks: map[RelationKind][]*Task{},
		BucketID:     11,
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task21 := &Task{
		ID:           21,
		Title:        "task #21",
		Identifier:   "test12-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ListID:       12,
		RelatedTasks: map[RelationKind][]*Task{},
		BucketID:     12,
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task22 := &Task{
		ID:           22,
		Title:        "task #22",
		Identifier:   "test13-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ListID:       13,
		RelatedTasks: map[RelationKind][]*Task{},
		BucketID:     13,
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task23 := &Task{
		ID:           23,
		Title:        "task #23",
		Identifier:   "test14-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ListID:       14,
		RelatedTasks: map[RelationKind][]*Task{},
		BucketID:     14,
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
		ListID:       15,
		RelatedTasks: map[RelationKind][]*Task{},
		BucketID:     15,
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
		ListID:       16,
		RelatedTasks: map[RelationKind][]*Task{},
		BucketID:     16,
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
		ListID:       17,
		RelatedTasks: map[RelationKind][]*Task{},
		BucketID:     17,
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}
	task27 := &Task{
		ID:          27,
		Title:       "task #27 with reminders",
		Identifier:  "test1-12",
		Index:       12,
		CreatedByID: 1,
		CreatedBy:   user1,
		Reminders: []time.Time{
			time.Unix(1543626724, 0).In(loc),
			time.Unix(1543626824, 0).In(loc),
		},
		ListID:       1,
		BucketID:     1,
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
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		RepeatAfter:  3600,
		BucketID:     1,
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
		ListID:      1,
		RelatedTasks: map[RelationKind][]*Task{
			RelationKindParenttask: {
				{
					ID:          1,
					Title:       "task #1",
					Description: "Lorem Ipsum",
					Index:       1,
					CreatedByID: 1,
					ListID:      1,
					IsFavorite:  true,
					Created:     time.Unix(1543626724, 0).In(loc),
					Updated:     time.Unix(1543626724, 0).In(loc),
					BucketID:    1,
					Position:    2,
				},
			},
		},
		BucketID: 1,
		Created:  time.Unix(1543626724, 0).In(loc),
		Updated:  time.Unix(1543626724, 0).In(loc),
	}
	task30 := &Task{
		ID:          30,
		Title:       "task #30 with assignees",
		Identifier:  "test1-15",
		Index:       15,
		CreatedByID: 1,
		CreatedBy:   user1,
		ListID:      1,
		Assignees: []*user.User{
			user1,
			user2,
		},
		RelatedTasks: map[RelationKind][]*Task{},
		BucketID:     1,
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
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		BucketID:     1,
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
		ListID:       3,
		RelatedTasks: map[RelationKind][]*Task{},
		BucketID:     21,
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
		ListID:       1,
		PercentDone:  0.5,
		RelatedTasks: map[RelationKind][]*Task{},
		BucketID:     1,
		Created:      time.Unix(1543626724, 0).In(loc),
		Updated:      time.Unix(1543626724, 0).In(loc),
	}

	type fields struct {
		ListID  int64
		Lists   []*List
		SortBy  []string // Is a string, since this is the place where a query string comes from the user
		OrderBy []string

		FilterBy           []string
		FilterValue        []string
		FilterComparator   []string
		FilterIncludeNulls bool

		CRUDable web.CRUDable
		Rights   web.Rights
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
		want    interface{}
		wantErr bool
	}

	defaultArgs := args{
		search: "",
		a:      &user.User{ID: 1},
		page:   0,
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
			},
			wantErr: false,
		},
		{
			// For more sorting tests see task_collection_sort_test.go
			name: "ReadAll Tasks sorted by done asc and id desc",
			fields: fields{
				SortBy:  []string{"done", "id"},
				OrderBy: []string{"asc", "desc"},
			},
			args: defaultArgs,
			want: []*Task{
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
			},
			wantErr: false,
		},
		{
			name: "ReadAll Tasks with range",
			fields: fields{
				FilterBy:         []string{"start_date", "end_date"},
				FilterValue:      []string{"2018-12-11T03:46:40+00:00", "2018-12-13T11:20:01+00:00"},
				FilterComparator: []string{"greater", "less"},
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
				FilterBy:         []string{"start_date", "end_date"},
				FilterValue:      []string{"2018-12-13T11:20:00+00:00", "2018-12-16T22:40:00+00:00"},
				FilterComparator: []string{"greater", "less"},
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
				FilterBy:         []string{"start_date"},
				FilterValue:      []string{"2018-12-12T07:33:20+00:00"},
				FilterComparator: []string{"greater"},
			},
			args:    defaultArgs,
			want:    []*Task{},
			wantErr: false,
		},
		{
			name: "ReadAll Tasks with range with start date only and greater equals",
			fields: fields{
				FilterBy:         []string{"start_date"},
				FilterValue:      []string{"2018-12-12T07:33:20+00:00"},
				FilterComparator: []string{"greater_equals"},
			},
			args: defaultArgs,
			want: []*Task{
				task7,
				task9,
			},
			wantErr: false,
		},
		{
			name: "undone tasks only",
			fields: fields{
				FilterBy:         []string{"done"},
				FilterValue:      []string{"false"},
				FilterComparator: []string{"equals"},
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
			},
			wantErr: false,
		},
		{
			name: "done tasks only",
			fields: fields{
				FilterBy:         []string{"done"},
				FilterValue:      []string{"true"},
				FilterComparator: []string{"equals"},
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
				FilterBy:         []string{"done"},
				FilterValue:      []string{"false"},
				FilterComparator: []string{"not_equals"},
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
				FilterBy:           []string{"start_date", "end_date"},
				FilterValue:        []string{"2018-12-11T03:46:40+00:00", "2018-12-13T11:20:01+00:00"},
				FilterComparator:   []string{"greater", "less"},
				FilterIncludeNulls: true,
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
			},
			wantErr: false,
		},
		{
			name: "favorited tasks",
			args: defaultArgs,
			fields: fields{
				ListID: FavoritesPseudoList.ID,
			},
			want: []*Task{
				task1,
				task15,
				// Task 34 is also a favorite, but on a list user 1 has no access to.
			},
		},
		{
			name: "filtered with like",
			fields: fields{
				FilterBy:         []string{"title"},
				FilterValue:      []string{"with"},
				FilterComparator: []string{"like"},
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
			name: "filtered reminders",
			fields: fields{
				FilterBy:         []string{"reminders", "reminders"},
				FilterValue:      []string{"2018-10-01T00:00:00+00:00", "2018-12-10T00:00:00+00:00"},
				FilterComparator: []string{"greater", "less"},
			},
			args: defaultArgs,
			want: []*Task{
				task2,
				task27,
			},
			wantErr: false,
		},
		{
			name: "filter in",
			fields: fields{
				FilterBy:         []string{"id"},
				FilterValue:      []string{"1,2,34"}, // Task 34 is forbidden for user 1
				FilterComparator: []string{"in"},
			},
			args: defaultArgs,
			want: []*Task{
				task1,
				task2,
			},
			wantErr: false,
		},
		{
			name: "filter assignees",
			fields: fields{
				FilterBy:         []string{"assignees"},
				FilterValue:      []string{"1"},
				FilterComparator: []string{"equals"},
			},
			args: defaultArgs,
			want: []*Task{
				task30,
			},
			wantErr: false,
		},
		{
			name: "filter assignees in",
			fields: fields{
				FilterBy:         []string{"assignees"},
				FilterValue:      []string{"1,2"},
				FilterComparator: []string{"in"},
			},
			args: defaultArgs,
			want: []*Task{
				task30,
			},
			wantErr: false,
		},
		{
			name: "filter labels",
			fields: fields{
				FilterBy:         []string{"labels"},
				FilterValue:      []string{"4"},
				FilterComparator: []string{"equals"},
			},
			args: defaultArgs,
			want: []*Task{
				task1,
				task2,
			},
			wantErr: false,
		},
		{
			name: "filter list",
			fields: fields{
				FilterBy:         []string{"list_id"},
				FilterValue:      []string{"6"},
				FilterComparator: []string{"equals"},
			},
			args: defaultArgs,
			want: []*Task{
				task15,
			},
			wantErr: false,
		},
		{
			name: "filter namespace",
			fields: fields{
				FilterBy:         []string{"namespace"},
				FilterValue:      []string{"7"},
				FilterComparator: []string{"equals"},
			},
			args: defaultArgs,
			want: []*Task{
				task21,
			},
			wantErr: false,
		},
		{
			name: "filter namespace in",
			fields: fields{
				FilterBy:         []string{"namespace"},
				FilterValue:      []string{"7,8"},
				FilterComparator: []string{"in"},
			},
			args: defaultArgs,
			want: []*Task{
				task21,
				task22,
			},
			wantErr: false,
		},
		{
			name: "filter by index",
			fields: fields{
				FilterBy:         []string{"index"},
				FilterValue:      []string{"5"},
				FilterComparator: []string{"equals"},
			},
			args: defaultArgs,
			want: []*Task{
				task5,
			},
			wantErr: false,
		},
		{
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
		},
		{
			name: "order by position",
			fields: fields{
				SortBy:  []string{"position", "id"},
				OrderBy: []string{"asc", "asc"},
			},
			args: args{
				a: &user.User{ID: 1},
			},
			want: []*Task{
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
				// The only tasks with a position set
				task1,
				task2,
			},
		},
		{
			name: "saved filter with sort order",
			fields: fields{
				ListID:  -2,
				SortBy:  []string{"title", "id"},
				OrderBy: []string{"desc", "asc"},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			lt := &TaskCollection{
				ListID:  tt.fields.ListID,
				SortBy:  tt.fields.SortBy,
				OrderBy: tt.fields.OrderBy,

				FilterBy:           tt.fields.FilterBy,
				FilterValue:        tt.fields.FilterValue,
				FilterComparator:   tt.fields.FilterComparator,
				FilterIncludeNulls: tt.fields.FilterIncludeNulls,

				CRUDable: tt.fields.CRUDable,
				Rights:   tt.fields.Rights,
			}
			got, _, _, err := lt.ReadAll(s, tt.args.a, tt.args.search, tt.args.page, 50)
			if (err != nil) != tt.wantErr {
				t.Errorf("Test %s, Task.ReadAll() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if diff, equal := messagediff.PrettyDiff(got, tt.want); !equal {
				if len(got.([]*Task)) == 0 && len(tt.want.([]*Task)) == 0 {
					return
				}

				t.Errorf("Test %s, Task.ReadAll() = %v, want %v, \ndiff: %v", tt.name, got, tt.want, diff)
			}
		})
	}
}
