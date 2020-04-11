// Vikunja is a to-do list application to facilitate your life.
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

package models

import (
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/timeutil"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"
	"gopkg.in/d4l3k/messagediff.v1"
	"testing"
)

func TestTaskCollection_ReadAll(t *testing.T) {
	// Dummy users
	user1 := &user.User{
		ID:       1,
		Username: "user1",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive: true,
	}
	user2 := &user.User{
		ID:       2,
		Username: "user2",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
	}
	user6 := &user.User{
		ID:       6,
		Username: "user6",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive: true,
	}

	// We use individual variables for the tasks here to be able to rearrange or remove ones more easily
	task1 := &Task{
		ID:          1,
		Text:        "task #1",
		Description: "Lorem Ipsum",
		Identifier:  "test1-1",
		Index:       1,
		CreatedByID: 1,
		CreatedBy:   user1,
		ListID:      1,
		Labels: []*Label{
			{
				ID:          4,
				Title:       "Label #4 - visible via other task",
				CreatedByID: 2,
				CreatedBy:   user2,
				Updated:     0,
				Created:     0,
			},
		},
		RelatedTasks: map[RelationKind][]*Task{
			RelationKindSubtask: {
				{
					ID:          29,
					Text:        "task #29 with parent task (1)",
					Index:       14,
					CreatedByID: 1,
					ListID:      1,
					Created:     1543626724,
					Updated:     1543626724,
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
				File: &files.File{
					ID:          1,
					Name:        "test",
					Size:        100,
					CreatedUnix: 1570998791,
					CreatedByID: 1,
				},
			},
			{
				ID:          2,
				TaskID:      1,
				FileID:      9999,
				CreatedByID: 1,
				CreatedBy:   user1,
			},
		},
		Created: 1543626724,
		Updated: 1543626724,
	}
	task2 := &Task{
		ID:          2,
		Text:        "task #2 done",
		Identifier:  "test1-2",
		Index:       2,
		Done:        true,
		CreatedByID: 1,
		CreatedBy:   user1,
		ListID:      1,
		Labels: []*Label{
			{
				ID:          4,
				Title:       "Label #4 - visible via other task",
				CreatedByID: 2,
				CreatedBy:   user2,
				Updated:     0,
				Created:     0,
			},
		},
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
	}
	task3 := &Task{
		ID:           3,
		Text:         "task #3 high prio",
		Identifier:   "test1-3",
		Index:        3,
		CreatedByID:  1,
		CreatedBy:    user1,
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
		Priority:     100,
	}
	task4 := &Task{
		ID:           4,
		Text:         "task #4 low prio",
		Identifier:   "test1-4",
		Index:        4,
		CreatedByID:  1,
		CreatedBy:    user1,
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
		Priority:     1,
	}
	task5 := &Task{
		ID:           5,
		Text:         "task #5 higher due date",
		Identifier:   "test1-5",
		Index:        5,
		CreatedByID:  1,
		CreatedBy:    user1,
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
		DueDate:      1543636724,
	}
	task6 := &Task{
		ID:           6,
		Text:         "task #6 lower due date",
		Identifier:   "test1-6",
		Index:        6,
		CreatedByID:  1,
		CreatedBy:    user1,
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
		DueDate:      1543616724,
	}
	task7 := &Task{
		ID:           7,
		Text:         "task #7 with start date",
		Identifier:   "test1-7",
		Index:        7,
		CreatedByID:  1,
		CreatedBy:    user1,
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
		StartDate:    1544600000,
	}
	task8 := &Task{
		ID:           8,
		Text:         "task #8 with end date",
		Identifier:   "test1-8",
		Index:        8,
		CreatedByID:  1,
		CreatedBy:    user1,
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
		EndDate:      1544700000,
	}
	task9 := &Task{
		ID:           9,
		Text:         "task #9 with start and end date",
		Identifier:   "test1-9",
		Index:        9,
		CreatedByID:  1,
		CreatedBy:    user1,
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
		StartDate:    1544600000,
		EndDate:      1544700000,
	}
	task10 := &Task{
		ID:           10,
		Text:         "task #10 basic",
		Identifier:   "test1-10",
		Index:        10,
		CreatedByID:  1,
		CreatedBy:    user1,
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
	}
	task11 := &Task{
		ID:           11,
		Text:         "task #11 basic",
		Identifier:   "test1-11",
		Index:        11,
		CreatedByID:  1,
		CreatedBy:    user1,
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
	}
	task12 := &Task{
		ID:           12,
		Text:         "task #12 basic",
		Identifier:   "test1-12",
		Index:        12,
		CreatedByID:  1,
		CreatedBy:    user1,
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
	}
	task15 := &Task{
		ID:           15,
		Text:         "task #15",
		Identifier:   "test6-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ListID:       6,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
	}
	task16 := &Task{
		ID:           16,
		Text:         "task #16",
		Identifier:   "test7-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ListID:       7,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
	}
	task17 := &Task{
		ID:           17,
		Text:         "task #17",
		Identifier:   "test8-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ListID:       8,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
	}
	task18 := &Task{
		ID:           18,
		Text:         "task #18",
		Identifier:   "test9-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ListID:       9,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
	}
	task19 := &Task{
		ID:           19,
		Text:         "task #19",
		Identifier:   "test10-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ListID:       10,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
	}
	task20 := &Task{
		ID:           20,
		Text:         "task #20",
		Identifier:   "test11-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ListID:       11,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
	}
	task21 := &Task{
		ID:           21,
		Text:         "task #21",
		Identifier:   "test12-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ListID:       12,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
	}
	task22 := &Task{
		ID:           22,
		Text:         "task #22",
		Identifier:   "test13-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ListID:       13,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
	}
	task23 := &Task{
		ID:           23,
		Text:         "task #23",
		Identifier:   "test14-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ListID:       14,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
	}
	task24 := &Task{
		ID:           24,
		Text:         "task #24",
		Identifier:   "test15-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ListID:       15,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
	}
	task25 := &Task{
		ID:           25,
		Text:         "task #25",
		Identifier:   "test16-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ListID:       16,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
	}
	task26 := &Task{
		ID:           26,
		Text:         "task #26",
		Identifier:   "test17-1",
		Index:        1,
		CreatedByID:  6,
		CreatedBy:    user6,
		ListID:       17,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
	}
	task27 := &Task{
		ID:           27,
		Text:         "task #27 with reminders",
		Identifier:   "test1-12",
		Index:        12,
		CreatedByID:  1,
		CreatedBy:    user1,
		Reminders:    []timeutil.TimeStamp{1543626724, 1543626824},
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
	}
	task28 := &Task{
		ID:           28,
		Text:         "task #28 with repeat after",
		Identifier:   "test1-13",
		Index:        13,
		CreatedByID:  1,
		CreatedBy:    user1,
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		RepeatAfter:  3600,
		Created:      1543626724,
		Updated:      1543626724,
	}
	task29 := &Task{
		ID:          29,
		Text:        "task #29 with parent task (1)",
		Identifier:  "test1-14",
		Index:       14,
		CreatedByID: 1,
		CreatedBy:   user1,
		ListID:      1,
		RelatedTasks: map[RelationKind][]*Task{
			RelationKindParenttask: {
				{
					ID:          1,
					Text:        "task #1",
					Description: "Lorem Ipsum",
					Index:       1,
					CreatedByID: 1,
					ListID:      1,
					Created:     1543626724,
					Updated:     1543626724,
				},
			},
		},
		Created: 1543626724,
		Updated: 1543626724,
	}
	task30 := &Task{
		ID:          30,
		Text:        "task #30 with assignees",
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
		Created:      1543626724,
		Updated:      1543626724,
	}
	task31 := &Task{
		ID:           31,
		Text:         "task #31 with color",
		Identifier:   "test1-16",
		Index:        16,
		HexColor:     "f0f0f0",
		CreatedByID:  1,
		CreatedBy:    user1,
		ListID:       1,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
	}
	task32 := &Task{
		ID:           32,
		Text:         "task #32",
		Identifier:   "test3-1",
		Index:        1,
		CreatedByID:  1,
		CreatedBy:    user1,
		ListID:       3,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
	}
	task33 := &Task{
		ID:           33,
		Text:         "task #33 with percent done",
		Identifier:   "test1-17",
		Index:        17,
		CreatedByID:  1,
		CreatedBy:    user1,
		ListID:       1,
		PercentDone:  0.5,
		RelatedTasks: map[RelationKind][]*Task{},
		Created:      1543626724,
		Updated:      1543626724,
	}

	type fields struct {
		ListID  int64
		Lists   []*List
		SortBy  []string // Is a string, since this is the place where a query string comes from the user
		OrderBy []string

		FilterBy         []string
		FilterValue      []string
		FilterComparator []string

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
				FilterValue:      []string{"1544500000", "1544700001"},
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
				FilterValue:      []string{"1544700000", "1545000000"},
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
				FilterValue:      []string{"1544600000"},
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
				FilterValue:      []string{"1544600000"},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)

			lt := &TaskCollection{
				ListID:  tt.fields.ListID,
				SortBy:  tt.fields.SortBy,
				OrderBy: tt.fields.OrderBy,

				FilterBy:         tt.fields.FilterBy,
				FilterValue:      tt.fields.FilterValue,
				FilterComparator: tt.fields.FilterComparator,

				CRUDable: tt.fields.CRUDable,
				Rights:   tt.fields.Rights,
			}
			got, _, _, err := lt.ReadAll(tt.args.a, tt.args.search, tt.args.page, 50)
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
