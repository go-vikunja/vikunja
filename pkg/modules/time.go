// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package modules

import "time"

type Time time.Time

func (t *Time) MarshalJSON() ([]byte, error) {
	if time.Time(*t).IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + time.Time(*t).Format(time.RFC3339) + `"`), nil
}

func (t *Time) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*t = Time{}
		return nil
	}
	parsedTime, err := time.Parse(`"`+time.RFC3339+`"`, string(data))
	if err != nil {
		return err
	}
	*t = Time(parsedTime)
	return nil
}

func (t *Time) Time() time.Time {
	return time.Time(*t)
}

func (t *Time) IsZero() bool {
	return t.Time().IsZero()
}

func (t *Time) Add(d time.Duration) *Time {
	newTime := Time(t.Time().Add(d))
	return &newTime
}

func (t *Time) After(u *Time) bool {
	return t.Time().After(u.Time())
}

func (t *Time) Before(u *Time) bool {
	return t.Time().Before(u.Time())
}

func (t *Time) Sub(u *Time) time.Duration {
	return t.Time().Sub(u.Time())
}

func (t *Time) Unix() int64 {
	return t.Time().Unix()
}

func (t *Time) In(loc *time.Location) *Time {
	newTime := Time(t.Time().In(loc))
	return &newTime
}

func (t *Time) Format(layout string) string {
	return t.Time().Format(layout)
}

func (t *Time) Month() time.Month {
	return t.Time().Month()
}

func TimeFromTime(time time.Time) *Time {
	t := Time(time)
	return &t
}
