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

package routes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactQueryToken(t *testing.T) {
	tests := []struct {
		name string
		uri  string
		want string
	}{
		{
			name: "no query",
			uri:  "/api/v2/projects/1",
			want: "/api/v2/projects/1",
		},
		{
			name: "unrelated query",
			uri:  "/api/v2/tasks?page=2&filter=done%3Dfalse",
			want: "/api/v2/tasks?page=2&filter=done%3Dfalse",
		},
		{
			name: "token only",
			uri:  "/api/v2/projects/1/calendar.ics?token=tk_secret",
			want: "/api/v2/projects/1/calendar.ics?token=REDACTED",
		},
		{
			name: "token among other params",
			uri:  "/api/v2/user/calendar.ics?include_done=true&token=tk_secret",
			want: "/api/v2/user/calendar.ics?include_done=true&token=REDACTED",
		},
		{
			name: "similarly named param is kept",
			uri:  "/api/v2/tasks?reset_token=abc",
			want: "/api/v2/tasks?reset_token=abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, redactQueryToken(tt.uri))
		})
	}
}
