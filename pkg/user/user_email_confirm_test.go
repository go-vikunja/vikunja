// Copyright2018-2020 Vikunja and contriubtors. All rights reserved.
//
// This file is part of Vikunja.
//
// Vikunja is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Vikunja is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Vikunja.  If not, see <https://www.gnu.org/licenses/>.

package user

import (
	"code.vikunja.io/api/pkg/db"
	"testing"
)

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
			db.LoadAndAssertFixtures(t)
			if err := ConfirmEmail(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("ConfirmEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
