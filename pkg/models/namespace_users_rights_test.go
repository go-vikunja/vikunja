//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"testing"

	"code.vikunja.io/web"
)

func TestNamespaceUser_CanDoSomething(t *testing.T) {
	type fields struct {
		ID          int64
		UserID      int64
		NamespaceID int64
		Right       Right
		Created     int64
		Updated     int64
		CRUDable    web.CRUDable
		Rights      web.Rights
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
				NamespaceID: 3,
			},
			args: args{
				a: &User{ID: 3},
			},
			want: map[string]bool{"CanCreate": true, "CanDelete": true, "CanUpdate": true},
		},
		{
			name: "CanDoSomething for a nonexistant namespace",
			fields: fields{
				NamespaceID: 300,
			},
			args: args{
				a: &User{ID: 3},
			},
			want: map[string]bool{"CanCreate": false, "CanDelete": false, "CanUpdate": false},
		},
		{
			name: "CanDoSomething where the user does not have the rights",
			fields: fields{
				NamespaceID: 3,
			},
			args: args{
				a: &User{ID: 4},
			},
			want: map[string]bool{"CanCreate": false, "CanDelete": false, "CanUpdate": false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nu := &NamespaceUser{
				ID:          tt.fields.ID,
				UserID:      tt.fields.UserID,
				NamespaceID: tt.fields.NamespaceID,
				Right:       tt.fields.Right,
				Created:     tt.fields.Created,
				Updated:     tt.fields.Updated,
				CRUDable:    tt.fields.CRUDable,
				Rights:      tt.fields.Rights,
			}
			if got := nu.CanCreate(tt.args.a); got != tt.want["CanCreate"] {
				t.Errorf("NamespaceUser.CanCreate() = %v, want %v", got, tt.want["CanCreate"])
			}
			if got := nu.CanDelete(tt.args.a); got != tt.want["CanDelete"] {
				t.Errorf("NamespaceUser.CanDelete() = %v, want %v", got, tt.want["CanDelete"])
			}
			if got := nu.CanUpdate(tt.args.a); got != tt.want["CanUpdate"] {
				t.Errorf("NamespaceUser.CanUpdate() = %v, want %v", got, tt.want["CanUpdate"])
			}
		})
	}
}
