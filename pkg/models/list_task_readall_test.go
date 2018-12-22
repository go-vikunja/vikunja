/*
 * Copyright (c) 2018 the Vikunja Authors. All rights reserved.
 * Use of this source code is governed by a LPGLv3-style
 * license that can be found in the LICENSE file.
 */

package models

import (
	"reflect"
	"sort"
	"testing"

	"code.vikunja.io/web"
)

func sortTasksForTesting(by SortBy) (tasks []*ListTask) {
	tasks = []*ListTask{
		{
			ID:          1,
			Text:        "task #1",
			CreatedByID: 1,
			ListID:      1,
			Created:     1543626724,
			Updated:     1543626724,
		},
		{
			ID:          2,
			Text:        "task #2 done",
			Done:        true,
			CreatedByID: 1,
			ListID:      1,
			Created:     1543626724,
			Updated:     1543626724,
		},
		{
			ID:          3,
			Text:        "task #3 high prio",
			CreatedByID: 1,
			ListID:      1,
			Created:     1543626724,
			Updated:     1543626724,
			Priority:    100,
		},
		{
			ID:          4,
			Text:        "task #4 low prio",
			CreatedByID: 1,
			ListID:      1,
			Created:     1543626724,
			Updated:     1543626724,
			Priority:    1,
		},
		{
			ID:          5,
			Text:        "task #5 higher due date",
			CreatedByID: 1,
			ListID:      1,
			Created:     1543626724,
			Updated:     1543626724,
			DueDateUnix: 1543636724,
		},
		{
			ID:          6,
			Text:        "task #6 lower due date",
			CreatedByID: 1,
			ListID:      1,
			Created:     1543626724,
			Updated:     1543626724,
			DueDateUnix: 1543616724,
		},
		{
			ID:            7,
			Text:          "task #7 with start date",
			CreatedByID:   1,
			ListID:        1,
			Created:       1543626724,
			Updated:       1543626724,
			StartDateUnix: 1544600000,
		},
		{
			ID:          8,
			Text:        "task #8 with end date",
			CreatedByID: 1,
			ListID:      1,
			Created:     1543626724,
			Updated:     1543626724,
			EndDateUnix: 1544700000,
		},
		{
			ID:            9,
			Text:          "task #9 with start and end date",
			CreatedByID:   1,
			ListID:        1,
			Created:       1543626724,
			Updated:       1543626724,
			StartDateUnix: 1544600000,
			EndDateUnix:   1544700000,
		},
	}

	switch by {
	case SortTasksByPriorityDesc:
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Priority > tasks[j].Priority
		})
	case SortTasksByPriorityAsc:
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Priority < tasks[j].Priority
		})
	case SortTasksByDueDateDesc:
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].DueDateUnix > tasks[j].DueDateUnix
		})
	case SortTasksByDueDateAsc:
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].DueDateUnix < tasks[j].DueDateUnix
		})
	}

	return
}

func TestListTask_ReadAll(t *testing.T) {
	type fields struct {
		ID                int64
		Text              string
		Description       string
		Done              bool
		DueDateUnix       int64
		RemindersUnix     []int64
		CreatedByID       int64
		ListID            int64
		RepeatAfter       int64
		ParentTaskID      int64
		Priority          int64
		Sorting           string
		StartDateSortUnix int64
		EndDateSortUnix   int64
		Subtasks          []*ListTask
		Created           int64
		Updated           int64
		CreatedBy         User
		CRUDable          web.CRUDable
		Rights            web.Rights
	}
	type args struct {
		search string
		a      web.Auth
		page   int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name:   "ReadAll ListTasks normally",
			fields: fields{},
			args: args{
				search: "",
				a:      &User{ID: 1},
				page:   0,
			},
			want:    sortTasksForTesting(SortTasksByUnsorted),
			wantErr: false,
		},
		{
			name: "ReadAll ListTasks sorted by priority (desc)",
			fields: fields{
				Sorting: "priority",
			},
			args: args{
				search: "",
				a:      &User{ID: 1},
				page:   0,
			},
			want:    sortTasksForTesting(SortTasksByPriorityDesc),
			wantErr: false,
		},
		{
			name: "ReadAll ListTasks sorted by priority asc",
			fields: fields{
				Sorting: "priorityasc",
			},
			args: args{
				search: "",
				a:      &User{ID: 1},
				page:   0,
			},
			want: []*ListTask{
				{
					ID:          1,
					Text:        "task #1",
					CreatedByID: 1,
					ListID:      1,
					Created:     1543626724,
					Updated:     1543626724,
				},
				{
					ID:          2,
					Text:        "task #2 done",
					Done:        true,
					CreatedByID: 1,
					ListID:      1,
					Created:     1543626724,
					Updated:     1543626724,
				},
				{
					ID:          5,
					Text:        "task #5 higher due date",
					CreatedByID: 1,
					ListID:      1,
					Created:     1543626724,
					Updated:     1543626724,
					DueDateUnix: 1543636724,
				},
				{
					ID:          6,
					Text:        "task #6 lower due date",
					CreatedByID: 1,
					ListID:      1,
					Created:     1543626724,
					Updated:     1543626724,
					DueDateUnix: 1543616724,
				},
				{
					ID:            7,
					Text:          "task #7 with start date",
					CreatedByID:   1,
					ListID:        1,
					Created:       1543626724,
					Updated:       1543626724,
					StartDateUnix: 1544600000,
				},
				{
					ID:          8,
					Text:        "task #8 with end date",
					CreatedByID: 1,
					ListID:      1,
					Created:     1543626724,
					Updated:     1543626724,
					EndDateUnix: 1544700000,
				},
				{
					ID:            9,
					Text:          "task #9 with start and end date",
					CreatedByID:   1,
					ListID:        1,
					Created:       1543626724,
					Updated:       1543626724,
					StartDateUnix: 1544600000,
					EndDateUnix:   1544700000,
				},
				{
					ID:          4,
					Text:        "task #4 low prio",
					CreatedByID: 1,
					ListID:      1,
					Created:     1543626724,
					Updated:     1543626724,
					Priority:    1,
				},
				{
					ID:          3,
					Text:        "task #3 high prio",
					CreatedByID: 1,
					ListID:      1,
					Created:     1543626724,
					Updated:     1543626724,
					Priority:    100,
				},
			},
			wantErr: false,
		},
		{
			name: "ReadAll ListTasks sorted by priority desc",
			fields: fields{
				Sorting: "prioritydesc",
			},
			args: args{
				search: "",
				a:      &User{ID: 1},
				page:   0,
			},
			want:    sortTasksForTesting(SortTasksByPriorityDesc),
			wantErr: false,
		},
		{
			name: "ReadAll ListTasks sorted by due date default desc",
			fields: fields{
				Sorting: "dueadate",
			},
			args: args{
				search: "",
				a:      &User{ID: 1},
				page:   0,
			},
			want:    sortTasksForTesting(SortTasksByDueDateDesc),
			wantErr: false,
		},
		{
			name: "ReadAll ListTasks sorted by due date asc",
			fields: fields{
				Sorting: "duedateasc",
			},
			args: args{
				search: "",
				a:      &User{ID: 1},
				page:   0,
			},
			want:    sortTasksForTesting(SortTasksByDueDateAsc),
			wantErr: false,
		},
		{
			name: "ReadAll ListTasks sorted by due date desc",
			fields: fields{
				Sorting: "dueadatedesc",
			},
			args: args{
				search: "",
				a:      &User{ID: 1},
				page:   0,
			},
			want:    sortTasksForTesting(SortTasksByDueDateDesc),
			wantErr: false,
		},
		{
			name: "ReadAll ListTasks with range",
			fields: fields{
				StartDateSortUnix: 1544500000,
				EndDateSortUnix:   1544600000,
			},
			args: args{
				search: "",
				a:      &User{ID: 1},
				page:   0,
			},
			want: []*ListTask{
				{
					ID:            7,
					Text:          "task #7 with start date",
					CreatedByID:   1,
					ListID:        1,
					Created:       1543626724,
					Updated:       1543626724,
					StartDateUnix: 1544600000,
				},
				{
					ID:            9,
					Text:          "task #9 with start and end date",
					CreatedByID:   1,
					ListID:        1,
					Created:       1543626724,
					Updated:       1543626724,
					StartDateUnix: 1544600000,
					EndDateUnix:   1544700000,
				},
			},
			wantErr: false,
		},
		{
			name: "ReadAll ListTasks with range",
			fields: fields{
				StartDateSortUnix: 1544700000,
				EndDateSortUnix:   1545000000,
			},
			args: args{
				search: "",
				a:      &User{ID: 1},
				page:   0,
			},
			want: []*ListTask{
				{
					ID:          8,
					Text:        "task #8 with end date",
					CreatedByID: 1,
					ListID:      1,
					Created:     1543626724,
					Updated:     1543626724,
					EndDateUnix: 1544700000,
				},
				{
					ID:            9,
					Text:          "task #9 with start and end date",
					CreatedByID:   1,
					ListID:        1,
					Created:       1543626724,
					Updated:       1543626724,
					StartDateUnix: 1544600000,
					EndDateUnix:   1544700000,
				},
			},
			wantErr: false,
		},
		{
			name: "ReadAll ListTasks with range without end date",
			fields: fields{
				StartDateSortUnix: 1544700000,
			},
			args: args{
				search: "",
				a:      &User{ID: 1},
				page:   0,
			},
			want: []*ListTask{
				{
					ID:          8,
					Text:        "task #8 with end date",
					CreatedByID: 1,
					ListID:      1,
					Created:     1543626724,
					Updated:     1543626724,
					EndDateUnix: 1544700000,
				},
				{
					ID:            9,
					Text:          "task #9 with start and end date",
					CreatedByID:   1,
					ListID:        1,
					Created:       1543626724,
					Updated:       1543626724,
					StartDateUnix: 1544600000,
					EndDateUnix:   1544700000,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lt := &ListTask{
				ID:                tt.fields.ID,
				Text:              tt.fields.Text,
				Description:       tt.fields.Description,
				Done:              tt.fields.Done,
				DueDateUnix:       tt.fields.DueDateUnix,
				RemindersUnix:     tt.fields.RemindersUnix,
				CreatedByID:       tt.fields.CreatedByID,
				ListID:            tt.fields.ListID,
				RepeatAfter:       tt.fields.RepeatAfter,
				ParentTaskID:      tt.fields.ParentTaskID,
				Priority:          tt.fields.Priority,
				Sorting:           tt.fields.Sorting,
				StartDateSortUnix: tt.fields.StartDateSortUnix,
				EndDateSortUnix:   tt.fields.EndDateSortUnix,
				Subtasks:          tt.fields.Subtasks,
				Created:           tt.fields.Created,
				Updated:           tt.fields.Updated,
				CreatedBy:         tt.fields.CreatedBy,
				CRUDable:          tt.fields.CRUDable,
				Rights:            tt.fields.Rights,
			}
			got, err := lt.ReadAll(tt.args.search, tt.args.a, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListTask.ReadAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListTask.ReadAll() = %v, want %v", got, tt.want)
				/*fmt.Println("Got:")
				gotslice := got.([]*ListTask)
				for _, g := range gotslice {
					fmt.Println(g.Priority, g.Text)
					//fmt.Println(g.StartDateUnix)
					//fmt.Println(g.EndDateUnix)
				}
				fmt.Println("Want:")
				wantslice := tt.want.([]*ListTask)
				for _, w := range wantslice {
					fmt.Println(w.Priority, w.Text)
					//fmt.Println(w.StartDateUnix)
					//fmt.Println(w.EndDateUnix)
				}*/
			}
		})
	}
}
