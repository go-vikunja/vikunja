// Copyright 2018-2020 Vikunja and contriubtors. All rights reserved.
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

package models

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"gopkg.in/d4l3k/messagediff.v1"
)

func TestListUsersFromList(t *testing.T) {
	testuser1 := &user.User{
		ID:       1,
		Username: "user1",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive: true,
		Issuer:   "local",
		Created:  testCreatedTime,
		Updated:  testUpdatedTime,
	}
	testuser2 := &user.User{
		ID:       2,
		Username: "user2",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Issuer:   "local",
		Created:  testCreatedTime,
		Updated:  testUpdatedTime,
	}
	testuser3 := &user.User{
		ID:                 3,
		Username:           "user3",
		Password:           "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		PasswordResetToken: "passwordresettesttoken",
		Issuer:             "local",
		Created:            testCreatedTime,
		Updated:            testUpdatedTime,
	}
	testuser4 := &user.User{
		ID:                4,
		Username:          "user4",
		Password:          "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive:          false,
		EmailConfirmToken: "tiepiQueed8ahc7zeeFe1eveiy4Ein8osooxegiephauph2Ael",
		Issuer:            "local",
		Created:           testCreatedTime,
		Updated:           testUpdatedTime,
	}
	testuser5 := &user.User{
		ID:                5,
		Username:          "user5",
		Password:          "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive:          false,
		EmailConfirmToken: "tiepiQueed8ahc7zeeFe1eveiy4Ein8osooxegiephauph2Ael",
		Issuer:            "local",
		Created:           testCreatedTime,
		Updated:           testUpdatedTime,
	}
	testuser6 := &user.User{
		ID:       6,
		Username: "user6",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive: true,
		Issuer:   "local",
		Created:  testCreatedTime,
		Updated:  testUpdatedTime,
	}
	testuser7 := &user.User{
		ID:       7,
		Username: "user7",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive: true,
		Issuer:   "local",
		Created:  testCreatedTime,
		Updated:  testUpdatedTime,
	}
	testuser8 := &user.User{
		ID:       8,
		Username: "user8",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive: true,
		Issuer:   "local",
		Created:  testCreatedTime,
		Updated:  testUpdatedTime,
	}
	testuser9 := &user.User{
		ID:       9,
		Username: "user9",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive: true,
		Issuer:   "local",
		Created:  testCreatedTime,
		Updated:  testUpdatedTime,
	}
	testuser10 := &user.User{
		ID:       10,
		Username: "user10",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive: true,
		Issuer:   "local",
		Created:  testCreatedTime,
		Updated:  testUpdatedTime,
	}
	testuser11 := &user.User{
		ID:       11,
		Username: "user11",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive: true,
		Issuer:   "local",
		Created:  testCreatedTime,
		Updated:  testUpdatedTime,
	}
	testuser12 := &user.User{
		ID:       12,
		Username: "user12",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive: true,
		Issuer:   "local",
		Created:  testCreatedTime,
		Updated:  testUpdatedTime,
	}
	testuser13 := &user.User{
		ID:       13,
		Username: "user13",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive: true,
		Issuer:   "local",
		Created:  testCreatedTime,
		Updated:  testUpdatedTime,
	}

	type args struct {
		l      *List
		search string
	}
	tests := []struct {
		name      string
		args      args
		wantUsers []*user.User
		wantErr   bool
	}{
		{
			name:      "Check owner only",
			args:      args{l: &List{ID: 18, OwnerID: 7}},
			wantUsers: []*user.User{testuser7},
		},
		{
			// This list has another different user shared for each possible method
			name: "Check with owner and other users",
			args: args{l: &List{ID: 19, OwnerID: 7}},
			wantUsers: []*user.User{
				testuser1, // Shared Via Team readonly
				testuser2, // Shared Via Team write
				testuser3, // Shared Via Team admin

				testuser4, // Shared Via User readonly
				testuser5, // Shared Via User write
				testuser6, // Shared Via User admin

				testuser7, // Owner

				testuser8,  // Shared Via NamespaceTeam readonly
				testuser9,  // Shared Via NamespaceTeam write
				testuser10, // Shared Via NamespaceTeam admin

				testuser11, // Shared Via NamespaceUser readonly
				testuser12, // Shared Via NamespaceUser write
				testuser13, // Shared Via NamespaceUser admin
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)

			gotUsers, err := ListUsersFromList(tt.args.l, tt.args.search)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListUsersFromList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff, equal := messagediff.PrettyDiff(tt.wantUsers, gotUsers); !equal {
				t.Errorf("Test %s, LabelTask.ReadAll() = %v, want %v, \ndiff: %v", tt.name, gotUsers, tt.wantUsers, diff)
			}
		})
	}
}
