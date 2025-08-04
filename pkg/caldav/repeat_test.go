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

func Test_getRruleFromInterval(t *testing.T) {
	tests := []struct {
		name         string
		interval     int64
		wantFreq     string
		wantInterval int64
	}{
		{"seconds", 435, "SECONDLY", 435},
		{"minutes", 120, "MINUTELY", 2},
		{"hours", 7200, "HOURLY", 2},
		{"daily", 86400, "DAILY", 1},
		{"weekly", 1209600, "WEEKLY", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFreq, gotInterval := getRruleFromInterval(tt.interval)
			if gotFreq != tt.wantFreq || gotInterval != tt.wantInterval {
				t.Errorf("getRruleFromInterval() = %s,%d; want %s,%d", gotFreq, gotInterval, tt.wantFreq, tt.wantInterval)
			}
		})
	}
}
