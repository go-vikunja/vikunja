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
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"code.vikunja.io/api/pkg/web"
)

func TestProjectUser_CanDoSomething(t *testing.T) {
	type fields struct {
		ID          int64
		UserID      int64
		ProjectID   int64
		Permission  Permission
		Created     time.Time
		Updated     time.Time
		CRUDable    web.CRUDable
		Permissions web.Permissions
	}
	type args struct {
		a web.Auth
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]bool
	}{
		{
			name: "CanDoSomething Normally",
			fields: fields{
				ProjectID: 3,
			},
			args: args{
				a: &user.User{ID: 3},
			},
			want: map[string]bool{"CanCreate": true, "CanDelete": true, "CanUpdate": true},
		},
		{
			name: "CanDoSomething for a nonexistant project",
			fields: fields{
				ProjectID: 300,
			},
			args: args{
				a: &user.User{ID: 3},
			},
			want: map[string]bool{"CanCreate": false, "CanDelete": false, "CanUpdate": false},
		},
		{
			name: "CanDoSomething where the user does not have the permissions",
			fields: fields{
				ProjectID: 3,
			},
			args: args{
				a: &user.User{ID: 4},
			},
			want: map[string]bool{"CanCreate": false, "CanDelete": false, "CanUpdate": false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()

			lu := &ProjectUser{
				ID:          tt.fields.ID,
				UserID:      tt.fields.UserID,
				ProjectID:   tt.fields.ProjectID,
				Permission:  tt.fields.Permission,
				Created:     tt.fields.Created,
				Updated:     tt.fields.Updated,
				CRUDable:    tt.fields.CRUDable,
				Permissions: tt.fields.Permissions,
			}
			if got, _ := lu.CanCreate(s, tt.args.a); got != tt.want["CanCreate"] {
				t.Errorf("ProjectUser.CanCreate() = %v, want %v", got, tt.want["CanCreate"])
			}
			if got, _ := lu.CanDelete(s, tt.args.a); got != tt.want["CanDelete"] {
				t.Errorf("ProjectUser.CanDelete() = %v, want %v", got, tt.want["CanDelete"])
			}
			if got, _ := lu.CanUpdate(s, tt.args.a); got != tt.want["CanUpdate"] {
				t.Errorf("ProjectUser.CanUpdate() = %v, want %v", got, tt.want["CanUpdate"])
			}
			_ = s.Close()
		})
	}
}
