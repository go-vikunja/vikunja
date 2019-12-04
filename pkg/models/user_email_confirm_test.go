// Vikunja is a todo-list application to facilitate your life.
// Copyright 2018-2019 Vikunja and contributors. All rights reserved.
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

import "testing"

func TestUserEmailConfirm(t *testing.T) {
	type args struct {
		c *EmailConfirm
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		errType func(error) bool
	}{
		{
			name: "Test Empty token",
			args: args{
				c: &EmailConfirm{
					Token: "",
				},
			},
			wantErr: true,
			errType: IsErrInvalidEmailConfirmToken,
		},
		{
			name: "Test invalid token",
			args: args{
				c: &EmailConfirm{
					Token: "invalid",
				},
			},
			wantErr: true,
			errType: IsErrInvalidEmailConfirmToken,
		},
		{
			name: "Test valid token",
			args: args{
				c: &EmailConfirm{
					Token: "tiepiQueed8ahc7zeeFe1eveiy4Ein8osooxegiephauph2Ael",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UserEmailConfirm(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("UserEmailConfirm() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
