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

package caldav

import "testing"

func Test_parseVTODOPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority int64
		want     int64
	}{
		{
			name:     "unset",
			priority: 0,
			want:     0,
		},
		{
			name:     "DO NOW",
			priority: 1,
			want:     5,
		},
		{
			name:     "urgent",
			priority: 2,
			want:     4,
		},
		{
			name:     "high 1",
			priority: 3,
			want:     3,
		},
		{
			name:     "high 2",
			priority: 4,
			want:     3,
		},
		{
			name:     "medium",
			priority: 5,
			want:     2,
		},
		{
			name:     "low 1",
			priority: 6,
			want:     1,
		},
		{
			name:     "low 2",
			priority: 7,
			want:     1,
		},
		{
			name:     "low 3",
			priority: 8,
			want:     1,
		},
		{
			name:     "low 4",
			priority: 9,
			want:     1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotVikunjaPriority := parseVTODOPriority(tt.priority); gotVikunjaPriority != tt.want {
				t.Errorf("parseVTODOPriority() = %v, want %v", gotVikunjaPriority, tt.want)
			}
		})
	}
}

func Test_mapPriorityToCaldav(t *testing.T) {
	tests := []struct {
		name               string
		priority           int64
		wantCaldavPriority int
	}{
		{
			name:               "unset",
			priority:           0,
			wantCaldavPriority: 0,
		},
		{
			name:               "low",
			priority:           1,
			wantCaldavPriority: 9,
		},
		{
			name:               "medium",
			priority:           2,
			wantCaldavPriority: 5,
		},
		{
			name:               "high",
			priority:           3,
			wantCaldavPriority: 3,
		},
		{
			name:               "urgent",
			priority:           4,
			wantCaldavPriority: 2,
		},
		{
			name:               "DO NOW",
			priority:           5,
			wantCaldavPriority: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotCaldavPriority := mapPriorityToCaldav(tt.priority); gotCaldavPriority != tt.wantCaldavPriority {
				t.Errorf("mapPriorityToCaldav() = %v, want %v", gotCaldavPriority, tt.wantCaldavPriority)
			}
		})
	}
}
