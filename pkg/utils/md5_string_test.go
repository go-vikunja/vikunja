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

func TestMd5String(t *testing.T) {
	type args struct {
		cleartext string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test md5 generation",
			args: args{cleartext: "vikunjarandomstringwhichisnotrandom"},
			want: "58b27d8a1e45a9427dcfb8dea261c5ae",
		},
		{
			name: "Test md5 generation",
			args: args{cleartext: "vikunjastring"},
			want: "3e22b01e055d3d113a946742c2f67b90",
		},
		{
			name: "Test md5 generation",
			args: args{cleartext: "somethingsomething"},
			want: "2264cdc4cf48a80cc00d23730b6c03ea",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Md5String(tt.args.cleartext); got != tt.want {
				t.Errorf("Md5String() = %v, want %v", got, tt.want)
			}
		})
	}
}
