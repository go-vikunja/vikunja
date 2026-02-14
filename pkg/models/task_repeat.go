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

package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/teambition/rrule-go"
)

// TaskRepeat represents a structured recurrence pattern for API serialization.
// Fields mirror rrule-go's ROption 1:1, excluding Dtstart (which maps to the task's
// due date) and Byeaster (non-standard extension).
type TaskRepeat struct {
	// The recurrence frequency: "yearly", "monthly", "weekly", "daily", "hourly", "minutely", or "secondly".
	Freq string `json:"freq"`
	// The interval between occurrences. Defaults to 1.
	Interval int `json:"interval"`
	// Days of the week: "mo","tu","we","th","fr","sa","su". Supports nth weekday with prefix, e.g. "2tu" (second Tuesday), "-1fr" (last Friday).
	ByDay []string `json:"by_day,omitempty"`
	// Days of the month (1-31, or negative from end).
	ByMonthDay []int `json:"by_month_day,omitempty"`
	// Months of the year (1-12).
	ByMonth []int `json:"by_month,omitempty"`
	// Days of the year (1-366, or negative from end).
	ByYearDay []int `json:"by_year_day,omitempty"`
	// ISO week numbers (1-53, or negative from end).
	ByWeekNo []int `json:"by_week_no,omitempty"`
	// Positions within the recurrence set (-366 to 366).
	BySetPos []int `json:"by_set_pos,omitempty"`
	// Hours of the day (0-23).
	ByHour []int `json:"by_hour,omitempty"`
	// Minutes of the hour (0-59).
	ByMinute []int `json:"by_minute,omitempty"`
	// Seconds of the minute (0-59).
	BySecond []int `json:"by_second,omitempty"`
	// Maximum number of occurrences. 0 means unlimited.
	Count int `json:"count,omitempty"`
	// End date for the recurrence in RFC 3339 format. Nil means no end date.
	Until *string `json:"until,omitempty"`
	// Start day of the week: "mo" (default), "tu", "we", etc.
	Wkst string `json:"wkst,omitempty"`
}

var freqToString = map[rrule.Frequency]string{
	rrule.YEARLY:   "yearly",
	rrule.MONTHLY:  "monthly",
	rrule.WEEKLY:   "weekly",
	rrule.DAILY:    "daily",
	rrule.HOURLY:   "hourly",
	rrule.MINUTELY: "minutely",
	rrule.SECONDLY: "secondly",
}

var stringToFreq = map[string]rrule.Frequency{
	"yearly":   rrule.YEARLY,
	"monthly":  rrule.MONTHLY,
	"weekly":   rrule.WEEKLY,
	"daily":    rrule.DAILY,
	"hourly":   rrule.HOURLY,
	"minutely": rrule.MINUTELY,
	"secondly": rrule.SECONDLY,
}

var weekdayIntToString = map[int]string{
	0: "mo", 1: "tu", 2: "we", 3: "th", 4: "fr", 5: "sa", 6: "su",
}

var stringToWeekdayMap = map[string]rrule.Weekday{
	"mo": rrule.MO, "tu": rrule.TU, "we": rrule.WE, "th": rrule.TH,
	"fr": rrule.FR, "sa": rrule.SA, "su": rrule.SU,
}

// taskRepeatFromRRule converts an RRULE string to a TaskRepeat struct.
// Returns nil if the input is empty.
func taskRepeatFromRRule(rruleStr string) *TaskRepeat {
	if rruleStr == "" {
		return nil
	}

	opt, err := rrule.StrToROption(rruleStr)
	if err != nil {
		return nil
	}

	r := &TaskRepeat{
		ByMonthDay: opt.Bymonthday,
		ByMonth:    opt.Bymonth,
		ByYearDay:  opt.Byyearday,
		ByWeekNo:   opt.Byweekno,
		BySetPos:   opt.Bysetpos,
		ByHour:     opt.Byhour,
		ByMinute:   opt.Byminute,
		BySecond:   opt.Bysecond,
		Count:      opt.Count,
	}

	if freqStr, ok := freqToString[opt.Freq]; ok {
		r.Freq = freqStr
	}

	r.Interval = opt.Interval
	if r.Interval == 0 {
		r.Interval = 1
	}

	if len(opt.Byweekday) > 0 {
		r.ByDay = make([]string, len(opt.Byweekday))
		for i, wd := range opt.Byweekday {
			r.ByDay[i] = weekdayToString(wd)
		}
	}

	if !opt.Until.IsZero() {
		s := opt.Until.UTC().Format(time.RFC3339)
		r.Until = &s
	}

	// Only include wkst if non-default (default is MO = weekday 0)
	if opt.Wkst.Day() != 0 {
		r.Wkst = weekdayIntToString[opt.Wkst.Day()]
	}

	return r
}

func weekdayToString(wd rrule.Weekday) string {
	base := weekdayIntToString[wd.Day()]
	if wd.N() == 0 {
		return base
	}
	return fmt.Sprintf("%d%s", wd.N(), base)
}

// toRRule converts a TaskRepeat struct to an RRULE string.
func (r *TaskRepeat) toRRule() (string, error) {
	if r == nil || r.Freq == "" {
		return "", nil
	}

	freq, ok := stringToFreq[strings.ToLower(r.Freq)]
	if !ok {
		return "", ErrInvalidData{Message: "Invalid frequency: " + r.Freq}
	}

	opt := rrule.ROption{
		Freq:       freq,
		Interval:   r.Interval,
		Bymonthday: r.ByMonthDay,
		Bymonth:    r.ByMonth,
		Byyearday:  r.ByYearDay,
		Byweekno:   r.ByWeekNo,
		Bysetpos:   r.BySetPos,
		Byhour:     r.ByHour,
		Byminute:   r.ByMinute,
		Bysecond:   r.BySecond,
		Count:      r.Count,
	}

	if len(r.ByDay) > 0 {
		opt.Byweekday = make([]rrule.Weekday, len(r.ByDay))
		for i, d := range r.ByDay {
			wd, err := parseWeekdayString(d)
			if err != nil {
				return "", ErrInvalidData{Message: "Invalid weekday in by_day: " + err.Error()}
			}
			opt.Byweekday[i] = wd
		}
	}

	if r.Until != nil {
		t, err := time.Parse(time.RFC3339, *r.Until)
		if err != nil {
			return "", ErrInvalidData{Message: "Invalid until date (expected RFC 3339): " + err.Error()}
		}
		opt.Until = t
	}

	if r.Wkst != "" {
		wd, err := parseWeekdayString(r.Wkst)
		if err != nil {
			return "", ErrInvalidData{Message: "Invalid wkst: " + err.Error()}
		}
		opt.Wkst = wd
	}

	return opt.RRuleString(), nil
}

func parseWeekdayString(s string) (rrule.Weekday, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if len(s) < 2 {
		return rrule.Weekday{}, fmt.Errorf("weekday string too short: %q", s)
	}

	// Check for nth prefix (e.g., "2tu", "-1fr")
	dayStr := s[len(s)-2:]
	wd, ok := stringToWeekdayMap[dayStr]
	if !ok {
		return rrule.Weekday{}, fmt.Errorf("unknown weekday: %q", dayStr)
	}

	if len(s) > 2 {
		nStr := s[:len(s)-2]
		var n int
		_, err := fmt.Sscanf(nStr, "%d", &n)
		if err != nil {
			return rrule.Weekday{}, fmt.Errorf("invalid nth weekday prefix %q: %w", nStr, err)
		}
		return wd.Nth(n), nil
	}

	return wd, nil
}
