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
			name: "Test with one task on different list",
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

			bt := &BulkTask{
				IDs:   tt.fields.IDs,
				Tasks: tt.fields.Tasks,
				Task:  tt.fields.Task,
			}
			allowed, _ := bt.CanUpdate(tt.fields.User)
			if !allowed != tt.wantForbidden {
				t.Errorf("BulkTask.Update() want forbidden, got %v, want %v", allowed, tt.wantForbidden)
			}
			if err := bt.Update(); (err != nil) != tt.wantErr {
				t.Errorf("BulkTask.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
