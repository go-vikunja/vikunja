// Vikunja is a todo-list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package timeutil

import (
	"code.vikunja.io/api/pkg/config"
	"encoding/json"
	"time"
)

// TimeStamp is a type which holds a unix timestamp, but becomes a RFC3339 date when parsed to json.
// This allows us to save the time as a unix timestamp into the database while returning it as an iso
// date to the api client.
type TimeStamp int64

// ToTime returns a time.Time from a TimeStamp
func (ts *TimeStamp) ToTime() time.Time {
	return time.Unix(int64(*ts), 0)
}

// FromTime converts a time.Time to a TimeStamp
func FromTime(t time.Time) TimeStamp {
	return TimeStamp(t.Unix())
}

// MarshalJSON converts a TimeStamp to a json string
func (ts *TimeStamp) MarshalJSON() ([]byte, error) {
	if int64(*ts) == 0 {
		return []byte("null"), nil
	}

	loc, err := time.LoadLocation(config.ServiceTimeZone.GetString())
	if err != nil {
		return nil, err
	}

	s := `"` + ts.ToTime().In(loc).Format(time.RFC3339) + `"`
	return []byte(s), nil
}

// UnmarshalJSON converts an iso date string from json to a TimeStamp
func (ts *TimeStamp) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	if s == "" {
		*ts = FromTime(time.Unix(0, 0))
		return nil
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	loc, err := time.LoadLocation(config.ServiceTimeZone.GetString())
	if err != nil {
		return err
	}

	*ts = TimeStamp(t.In(loc).Unix())
	return nil
}
