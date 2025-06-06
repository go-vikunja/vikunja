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

import (
	"regexp"
	"strconv"
	"time"
)

// ParseISO8601Duration converts a ISO8601 duration into a time.Duration
func ParseISO8601Duration(str string) time.Duration {
	matches := durationRegex.FindStringSubmatch(str)

	if len(matches) == 0 {
		return 0
	}

	years := parseDurationPart(matches[2], time.Hour*24*365)
	months := parseDurationPart(matches[3], time.Hour*24*30)
	days := parseDurationPart(matches[4], time.Hour*24)
	hours := parseDurationPart(matches[5], time.Hour)
	minutes := parseDurationPart(matches[6], time.Second*60)
	seconds := parseDurationPart(matches[7], time.Second)

	duration := years + months + days + hours + minutes + seconds

	if matches[1] == "-" {
		return -duration
	}
	return duration
}

var durationRegex = regexp.MustCompile(`([-+])?P([\d\.]+Y)?([\d\.]+M)?([\d\.]+D)?T?([\d\.]+H)?([\d\.]+M)?([\d\.]+?S)?`)

func parseDurationPart(value string, unit time.Duration) time.Duration {
	if len(value) != 0 {
		if parsed, err := strconv.ParseFloat(value[:len(value)-1], 64); err == nil {
			return time.Duration(float64(unit) * parsed)
		}
	}
	return 0
}
