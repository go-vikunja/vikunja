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
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"gopkg.in/d4l3k/messagediff.v1"
	"testing"
)

func TestListUsersFromList(t *testing.T) {
	testuser1 := &user.User{
		ID:        1,
		Username:  "user1",
		Password:  "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive:  true,
		AvatarURL: "111d68d06e2d317b5a59c2c6c5bad808",
	}
	testuser2 := &user.User{
		ID:        2,
		Username:  "user2",
		Password:  "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		AvatarURL: "ab53a2911ddf9b4817ac01ddcd3d975f",
	}
	testuser3 := &user.User{
		ID:                 3,
		Username:           "user3",
		Password:           "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		AvatarURL:          "97d6d9441ff85fdc730e02a6068d267b",
		PasswordResetToken: "passwordresettesttoken",
	}
	testuser4 := &user.User{
		ID:                4,
		Username:          "user4",
		Password:          "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive:          false,
		AvatarURL:         "7e65550957227bd38fe2d7fbc6fd2f7b",
		EmailConfirmToken: "tiepiQueed8ahc7zeeFe1eveiy4Ein8osooxegiephauph2Ael",
	}
	testuser5 := &user.User{
		ID:                5,
		Username:          "user5",
		Password:          "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive:          false,
		AvatarURL:         "cfa35b8cd2ec278026357769582fa563",
		EmailConfirmToken: "tiepiQueed8ahc7zeeFe1eveiy4Ein8osooxegiephauph2Ael",
	}
	testuser6 := &user.User{
		ID:        6,
		Username:  "user6",
		Password:  "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive:  true,
		AvatarURL: "3efbe51f864c6666bc27caf4c6ff90ed",
	}
	testuser7 := &user.User{
		ID:        7,
		Username:  "user7",
		Password:  "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive:  true,
		AvatarURL: "e80a711d4de44c30054806ebbd488464",
	}
	testuser8 := &user.User{
		ID:        8,
		Username:  "user8",
		Password:  "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive:  true,
		AvatarURL: "2b9b320416cd31020bb6844c3fadefd1",
	}
	testuser9 := &user.User{
		ID:        9,
		Username:  "user9",
		Password:  "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive:  true,
		AvatarURL: "f784fdb21d26dd2c64f5135f35ec401f",
	}
	testuser10 := &user.User{
		ID:        10,
		Username:  "user10",
		Password:  "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive:  true,
		AvatarURL: "fce8ff4ff56d75ad587d1bbaa5ef0563",
	}
	testuser11 := &user.User{
		ID:        11,
		Username:  "user11",
		Password:  "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive:  true,
		AvatarURL: "ad6d67d0c4495e186010732a7d360028",
	}
	testuser12 := &user.User{
		ID:        12,
		Username:  "user12",
		Password:  "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive:  true,
		AvatarURL: "ef1debc1364806281c42eeedfdeb943b",
	}
	testuser13 := &user.User{
		ID:        13,
		Username:  "user13",
		Password:  "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive:  true,
		AvatarURL: "b9e3f76032af53c9ff2df52d51ada717",
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
