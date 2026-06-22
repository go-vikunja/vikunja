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

package migration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertLegacyRepeatToRRule(t *testing.T) {
	const (
		modeDefault         = 0
		modeMonth           = 1
		modeFromCurrentDate = 2
		modeYear            = 3
	)

	cases := []struct {
		name        string
		repeatAfter int64
		repeatMode  int
		want        string
	}{
		{"default daily", 86400, modeDefault, "FREQ=DAILY;INTERVAL=1"},
		{"default every 2 days", 172800, modeDefault, "FREQ=DAILY;INTERVAL=2"},
		{"default weekly", 604800, modeDefault, "FREQ=WEEKLY;INTERVAL=1"},
		{"default hourly", 3600, modeDefault, "FREQ=HOURLY;INTERVAL=1"},
		{"default minutely", 60, modeDefault, "FREQ=MINUTELY;INTERVAL=1"},
		{"default secondly remainder", 90, modeDefault, "FREQ=SECONDLY;INTERVAL=90"},
		{"default no interval is empty", 0, modeDefault, ""},
		{"from current date keeps the interval", 86400, modeFromCurrentDate, "FREQ=DAILY;INTERVAL=1"},
		{"from current date no interval is empty", 0, modeFromCurrentDate, ""},
		{"monthly ignores repeat_after", 86400, modeMonth, "FREQ=MONTHLY;INTERVAL=1"},
		{"monthly without interval", 0, modeMonth, "FREQ=MONTHLY;INTERVAL=1"},
		{"yearly ignores repeat_after", 86400, modeYear, "FREQ=YEARLY;INTERVAL=1"},
		{"unknown mode is empty", 86400, 99, ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, convertLegacyRepeatToRRule(c.repeatAfter, c.repeatMode))
		})
	}
}

func TestSecondsToRRule(t *testing.T) {
	cases := []struct {
		seconds int64
		want    string
	}{
		{604800, "FREQ=WEEKLY;INTERVAL=1"},
		{1209600, "FREQ=WEEKLY;INTERVAL=2"},
		{86400, "FREQ=DAILY;INTERVAL=1"},
		{3600, "FREQ=HOURLY;INTERVAL=1"},
		{60, "FREQ=MINUTELY;INTERVAL=1"},
		{30, "FREQ=SECONDLY;INTERVAL=30"},
	}
	for _, c := range cases {
		t.Run(c.want, func(t *testing.T) {
			assert.Equal(t, c.want, secondsToRRule(c.seconds))
		})
	}
}
