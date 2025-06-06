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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
)

func TestBulkTask_Update(t *testing.T) {
	type fields struct {
		IDs   []int64
		Tasks []*Task
		Task  Task
		User  *user.User
	}
	tests := []struct {
		name          string
		fields        fields
		wantErr       bool
		wantForbidden bool
	}{
		{
			name: "Test normal update",
			fields: fields{
				IDs: []int64{10, 11, 12},
				Task: Task{
					Title: "bulkupdated",
				},
				User: &user.User{ID: 1},
			},
		},
		{
			name: "Test with one task on different project",
			fields: fields{
				IDs: []int64{10, 11, 12, 13},
				Task: Task{
					Title: "bulkupdated",
				},
				User: &user.User{ID: 1},
			},
			wantForbidden: true,
		},
		{
			name: "Test without any tasks",
			fields: fields{
				IDs: []int64{},
				Task: Task{
					Title: "bulkupdated",
				},
				User: &user.User{ID: 1},
			},
			wantForbidden: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)

			s := db.NewSession()

			bt := &BulkTask{
				IDs:   tt.fields.IDs,
				Tasks: tt.fields.Tasks,
				Task:  tt.fields.Task,
			}
			allowed, _ := bt.CanUpdate(s, tt.fields.User)
			if !allowed != tt.wantForbidden {
				t.Errorf("BulkTask.Update() want forbidden, got %v, want %v", allowed, tt.wantForbidden)
			}
			if err := bt.Update(s, tt.fields.User); (err != nil) != tt.wantErr {
				t.Errorf("BulkTask.Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			s.Close()
		})
	}
}
