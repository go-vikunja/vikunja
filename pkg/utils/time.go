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
	"time"

	"code.vikunja.io/api/pkg/config"
)

// GetTimeWithoutNanoSeconds returns a time.Time without the nanoseconds.
func GetTimeWithoutNanoSeconds(t time.Time) time.Time {
	tz := config.GetTimeZone()

	// By default, time.Now() includes nanoseconds which we don't save. That results in getting the wrong dates,
	// so we make sure the time we use to get the reminders don't contain nanoseconds.
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location()).In(tz)
}

// GetTimeWithoutSeconds returns a time.Time with the seconds set to 0.
func GetTimeWithoutSeconds(t time.Time) time.Time {
	tz := config.GetTimeZone()

	// By default, time.Now() includes nanoseconds which we don't save. That results in getting the wrong dates,
	// so we make sure the time we use to get the reminders don't contain nanoseconds.
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location()).In(tz)
}
