package models

import (
	"testing"
)

func TestBulkTask_Update(t *testing.T) {
	type fields struct {
		IDs      []int64
		Tasks    []*ListTask
		ListTask ListTask
		User     *User
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
				ListTask: ListTask{
					Text: "bulkupdated",
				},
				User: &User{ID: 1},
			},
		},
		{
			name: "Test with one task on different list",
			fields: fields{
				IDs: []int64{10, 11, 12, 13},
				ListTask: ListTask{
					Text: "bulkupdated",
				},
				User: &User{ID: 1},
			},
			wantForbidden: true,
		},
		{
			name: "Test without any tasks",
			fields: fields{
				IDs: []int64{},
				ListTask: ListTask{
					Text: "bulkupdated",
				},
				User: &User{ID: 1},
			},
			wantForbidden: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bt := &BulkTask{
				IDs:      tt.fields.IDs,
				Tasks:    tt.fields.Tasks,
				ListTask: tt.fields.ListTask,
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
