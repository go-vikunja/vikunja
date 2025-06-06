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

package utils

import "testing"

func TestSha256(t *testing.T) {
	type args struct {
		cleartext string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test sha256 generation",
			args: args{cleartext: "vikunjarandomstringwhichisnotrandom"},
			want: "46fc0f603dd986cf7ed3e631917d43da89a8df2bdf291",
		},
		{
			name: "Test sha256 generation",
			args: args{cleartext: "vikunjastring"},
			want: "f54d310f4d9a0bc13479dad5c5701e8d581744666b69f",
		},
		{
			name: "Test sha256 generation",
			args: args{cleartext: "somethingsomething"},
			want: "00aef67d6df7fdee0419aa3713820e7084cbcb8b8f7c4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sha256(tt.args.cleartext); got != tt.want {
				t.Errorf("Sha256() = %v, want %v", got, tt.want)
			}
		})
	}
}
