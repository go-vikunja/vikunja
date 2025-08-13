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

func TestTeam_CanDoSomething(t *testing.T) {
	type fields struct {
		ID          int64
		Name        string
		Description string
		CreatedByID int64
		CreatedBy   *user.User
		Members     []*TeamUser
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
				ID: 1,
			},
			args: args{
				a: &user.User{ID: 1},
			},
			want: map[string]bool{"CanCreate": true, "IsAdmin": true, "CanRead": true, "CanDelete": true, "CanUpdate": true},
		},
		{
			name: "CanDoSomething where the user does not have the permissions",
			fields: fields{
				ID: 1,
			},
			args: args{
				a: &user.User{ID: 4},
			},
			want: map[string]bool{"CanCreate": true, "IsAdmin": false, "CanRead": false, "CanDelete": false, "CanUpdate": false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			tm := &Team{
				ID:          tt.fields.ID,
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				CreatedByID: tt.fields.CreatedByID,
				CreatedBy:   tt.fields.CreatedBy,
				Members:     tt.fields.Members,
				Created:     tt.fields.Created,
				Updated:     tt.fields.Updated,
				CRUDable:    tt.fields.CRUDable,
				Permissions: tt.fields.Permissions,
			}

			if got, _ := tm.CanCreate(s, tt.args.a); got != tt.want["CanCreate"] { // CanCreate is currently always true
				t.Errorf("Team.CanCreate() = %v, want %v", got, tt.want["CanCreate"])
			}
			if got, _ := tm.CanDelete(s, tt.args.a); got != tt.want["CanDelete"] {
				t.Errorf("Team.CanDelete() = %v, want %v", got, tt.want["CanDelete"])
			}
			if got, _ := tm.CanUpdate(s, tt.args.a); got != tt.want["CanUpdate"] {
				t.Errorf("Team.CanUpdate() = %v, want %v", got, tt.want["CanUpdate"])
			}
			if got, _, _ := tm.CanRead(s, tt.args.a); got != tt.want["CanRead"] {
				t.Errorf("Team.CanRead() = %v, want %v", got, tt.want["CanRead"])
			}
			if got, _ := tm.IsAdmin(s, tt.args.a); got != tt.want["IsAdmin"] {
				t.Errorf("Team.IsAdmin() = %v, want %v", got, tt.want["IsAdmin"])
			}
		})
	}
}
