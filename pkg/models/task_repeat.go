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
	"slices"
	"strings"
	"time"

	"github.com/teambition/rrule-go"
)

// TaskRepeat represents a structured recurrence pattern for API serialization.
// Fields mirror rrule-go's ROption 1:1, excluding Dtstart (which maps to the task's
// due date) and Byeaster (non-standard extension). ReadOnly is response-only: it marks a
// projection that drops parts of the stored rule, and is ignored on input.
type TaskRepeat struct {
	// The recurrence frequency: "yearly", "monthly", "weekly", "daily", "hourly", "minutely", or "secondly".
	Freq string `json:"freq" doc:"How often the task repeats: yearly, monthly, weekly, daily, hourly, minutely or secondly. Any case is accepted; responses are lowercase. The sub-daily frequencies (hourly, minutely, secondly) cannot be combined with any by_* field."`
	// The interval between occurrences. Defaults to 1.
	Interval int `json:"interval" doc:"Repeat every this many periods of freq, for example 2 with weekly for every other week. 1 by default; 0 is read as 1 and responses return 1; negative values are refused. One step may span at most 100 years: at most 100 for yearly, 1200 monthly, 5218 weekly, 36525 daily, 876600 hourly, 52596000 minutely and 3155760000 secondly."`
	// Days of the week: "mo","tu","we","th","fr","sa","su". Supports nth weekday with prefix, e.g. "2tu" (second Tuesday), "-1fr" (last Friday).
	ByDay []string `json:"by_day,omitempty" doc:"Weekdays as two-letter codes mo, tu, we, th, fr, sa, su. A signed position may prefix a code for monthly and yearly rules only: 2tu is the second Tuesday, -1fr the last Friday. Positions range from -5 to 5 within a month (monthly, or yearly with by_month) and from -53 to 53 within a year (yearly). Any case is accepted; responses are lowercase. No value may appear twice."`
	// Days of the month (1-31, or negative from end).
	ByMonthDay []int `json:"by_month_day,omitempty" doc:"Days of the month, 1 to 31, or -31 to -1 counting from the end (-1 is the last day). No value may appear twice."`
	// Months of the year (1-12).
	ByMonth []int `json:"by_month,omitempty" doc:"Months of the year, 1 (January) to 12. No value may appear twice."`
	// Days of the year (1-366, or negative from end).
	ByYearDay []int `json:"by_year_day,omitempty" doc:"Days of the year, 1 to 366, or -366 to -1 counting from the end. No value may appear twice."`
	// ISO week numbers (1-53, or negative from end).
	ByWeekNo []int `json:"by_week_no,omitempty" doc:"ISO week numbers, 1 to 53, or -53 to -1 counting from the end of the year. No value may appear twice."`
	// Positions within the recurrence set (-366 to 366).
	BySetPos []int `json:"by_set_pos,omitempty" doc:"Picks occurrences by their position within each period: 1 is the first, -1 the last, up to 366 or -366. Combine it with another by_* field, for example by_day mo,tu,we,th,fr with by_set_pos -1 for the last weekday of the month. At most 10 values, none twice, and none beyond the number of occurrences a period can hold."`
	// Hours of the day (0-23).
	ByHour []int `json:"by_hour,omitempty" doc:"Hours of the day, 0 to 23. No value may appear twice. One period of freq (counted as 366 days for yearly, 31 for monthly, 7 for weekly and 1 otherwise) times the combinations of by_hour, by_minute and by_second may hold at most 10000 occurrences."`
	// Minutes of the hour (0-59).
	ByMinute []int `json:"by_minute,omitempty" doc:"Minutes of the hour, 0 to 59. No value may appear twice. One period of freq (counted as 366 days for yearly, 31 for monthly, 7 for weekly and 1 otherwise) times the combinations of by_hour, by_minute and by_second may hold at most 10000 occurrences."`
	// Seconds of the minute (0-59).
	BySecond []int `json:"by_second,omitempty" doc:"Seconds of the minute, 0 to 59. No value may appear twice. One period of freq (counted as 366 days for yearly, 31 for monthly, 7 for weekly and 1 otherwise) times the combinations of by_hour, by_minute and by_second may hold at most 10000 occurrences."`
	// Maximum number of occurrences. 0 means unlimited.
	Count int `json:"count,omitempty" doc:"How many occurrences remain, including the current one. 0 or omitted means unlimited; at most 10000; negative values are refused. Each completion lowers it by one; completing the task when it is 1 ends the recurrence: the repeat is removed and the task stays done."`
	// End date for the recurrence in RFC 3339 format. Nil means no end date.
	Until *string `json:"until,omitempty" doc:"The last moment an occurrence may fall on, as an RFC 3339 date-time (inclusive). Responses return it in UTC. Omit it for no end. Once it has passed, completing the task leaves it done and keeps the repeat."`
	// Start day of the week: "mo" (default), "tu", "we", etc.
	Wkst string `json:"wkst,omitempty" doc:"The day a week starts on, which matters for weekly rules with an interval above 1 and by_day, and for by_week_no: mo (the default, omitted in responses), tu, we, th, fr, sa or su. Any case is accepted."`
	// True when the stored rule has parts this structured view cannot hold (for example BYEASTER or
	// DTSTART from a CalDAV client). Such a rule can be removed but not edited through the API.
	ReadOnly bool `json:"read_only,omitempty" readOnly:"true" doc:"Response only; ignored on write. True when the stored recurrence rule has parts this structured form cannot represent, such as BYEASTER or a DTSTART set by a CalDAV client. Such a rule can be removed (send repeat: null) but not edited through the API."`
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

// NormalizeRRule strips an optional (case-insensitive) RRULE: prefix, validates
// the value with the rrule parser, and returns it in canonical form. ok is false
// when the value is not a parseable RRULE or is rejected by rrulePolicyRejects,
// so callers can skip it instead of storing or emitting calendar data that would
// later fail validation.
func NormalizeRRule(raw string) (normalized string, ok bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	const prefix = "RRULE:"
	if len(raw) >= len(prefix) && strings.EqualFold(raw[:len(prefix)], prefix) {
		raw = strings.TrimSpace(raw[len(prefix):])
	}
	opt, err := rrule.StrToROption(raw)
	if err != nil {
		return "", false
	}
	if checkRRuleWriteLimits(opt) != nil || rrulePolicyRejects(opt) {
		return "", false
	}
	normalized = opt.RRuleString()
	if len(normalized) > maxRRuleLength {
		return "", false
	}
	return normalized, true
}

// maxRRuleLength is the size of the tasks.repeats column (varchar(500), in the model and in
// migration 20261022220000). SQLite ignores it, but PostgreSQL and MySQL refuse a longer value
// with a server error, so every write checks the rule text against it first.
const maxRRuleLength = 500

// rruleIntDomains are the value domains of the integer BY* parts (RFC 5545 section 3.3.10).
// signed means -hi..-lo is valid too.
var rruleIntDomains = []struct {
	part   string
	values func(*rrule.ROption) []int
	lo, hi int
	signed bool
}{
	// RFC 5545 allows 60 (a leap second); rrule-go's validateBounds does not, so neither do we:
	// a CalDAV rule with BYSECOND=60 is then dropped like any other invalid rule.
	{"BYSECOND", func(o *rrule.ROption) []int { return o.Bysecond }, 0, 59, false},
	{"BYMINUTE", func(o *rrule.ROption) []int { return o.Byminute }, 0, 59, false},
	{"BYHOUR", func(o *rrule.ROption) []int { return o.Byhour }, 0, 23, false},
	{"BYMONTHDAY", func(o *rrule.ROption) []int { return o.Bymonthday }, 1, 31, true},
	{"BYYEARDAY", func(o *rrule.ROption) []int { return o.Byyearday }, 1, 366, true},
	{"BYWEEKNO", func(o *rrule.ROption) []int { return o.Byweekno }, 1, 53, true},
	{"BYMONTH", func(o *rrule.ROption) []int { return o.Bymonth }, 1, 12, false},
	{"BYSETPOS", func(o *rrule.ROption) []int { return o.Bysetpos }, 1, 366, true},
	// BYEASTER is rrule-go's extension, not RFC 5545: a day offset from Easter Sunday. rrule-go
	// indexes a mask of the year's days with it and panics outside -80..257, the offsets that
	// stay inside every year (Easter falls between day 81 and day 116).
	{"BYEASTER", func(o *rrule.ROption) []int { return o.Byeaster }, -80, 257, false},
}

// maxOccurrencesPerPeriod is the occurrences-per-period budget (see checkRRuleBounds). 10,000
// keeps rrule-go's per-period slice well under a MB and still admits every rule the editor, the
// importers and the legacy migration produce, which have at most a handful of times per day: a
// yearly rule may hold up to 27 times of day, a monthly one 322, a weekly one 1,428 and a daily one
// 10,000. Monthly and yearly anchors are not clamped, so a dense in-budget rule on an
// ancient due date still costs a fraction of a second per completion; under the earlier 100,000
// it cost seconds.
const maxOccurrencesPerPeriod = 10_000

// maxBySetPos is the most BYSETPOS values a rule may hold.
// rrule-go evaluates every position in every period, so a rule that can never match (an
// impossible date) with ~120 in-bound positions cost ~0.85 s per completion, scanning every day
// to the year 9999; with 10 the worst measured case is ~170 ms, against ~100 ms for the same
// daily scan with a single position.
const maxBySetPos = 10

// maxRRuleCount is the largest COUNT accepted: a counted rule is
// evaluated occurrence by occurrence from its start, so a huge COUNT on a fine-grained rule with
// an old due date burned minutes of CPU per completion. 10,000 is about 27 years of daily repeats.
const maxRRuleCount = 10_000

// The recurrence time horizon: nothing in the app looks further
// than 100 to 200 years. A rule's INTERVAL may span at most 100 years (maxIntervalByHorizon), on
// every write and at completion; completion also leaves alone a task whose due date lies more
// than maxAnchorAgeYears in the past.
//
// The INTERVAL limit also keeps every computation far from the overflows larger ones hit: a WEEKLY
// interval near 2^61 made the anchor clamp and rrule-go loop forever, a huge MONTHLY one panicked
// rrule-go's month counter, and a DAILY one near 2^63 wrapped its day arithmetic backwards. Every
// realistic legacy repeat is within it: upstream has capped repeat_after at 10 years since 2026.
const (
	horizonYears      = 100
	maxAnchorAgeYears = 200
)

// maxIntervalByHorizon is the largest INTERVAL per frequency: a span of 100 years. int64, because
// the SECONDLY limit does not fit a 32-bit int (upstream ships arm and 386 builds).
var maxIntervalByHorizon = map[rrule.Frequency]int64{
	rrule.YEARLY:   100,
	rrule.MONTHLY:  1_200,
	rrule.WEEKLY:   5_218,
	rrule.DAILY:    36_525,
	rrule.HOURLY:   876_600,
	rrule.MINUTELY: 52_596_000,
	rrule.SECONDLY: 3_155_760_000,
}

// intervalBeyondHorizon reports whether the rule's INTERVAL spans more than 100 years.
func intervalBeyondHorizon(opt *rrule.ROption) bool {
	limit, ok := maxIntervalByHorizon[opt.Freq]
	return ok && int64(opt.Interval) > limit
}

// periodDays is how many days rrule-go fills per generated period of each frequency.
var periodDays = map[rrule.Frequency]int{
	rrule.YEARLY:   366,
	rrule.MONTHLY:  31,
	rrule.WEEKLY:   7,
	rrule.DAILY:    1,
	rrule.HOURLY:   1,
	rrule.MINUTELY: 1,
	rrule.SECONDLY: 1,
}

// checkRRuleBounds is the safety bound on a parsed rule, applied before rrule-go evaluates it.
// It is separate from the creation policy (rrulePolicyRejects): it applies on every path,
// including trusted copies, and it rejects nothing a well-formed rule needs. It refuses:
//
//   - a BY* list that holds a value twice (new). rrule-go checks neither list sizes nor
//     duplicates, so ~3 KB of rule text with 500 zeros in each time list made it build ~125M
//     timestamps (~3 GB). With distinct values every list is at most its domain's size.
//   - a value outside its domain. For BYSECOND, BYMINUTE, BYHOUR, BYMONTHDAY, BYYEARDAY,
//     BYWEEKNO, BYMONTH and BYSETPOS this duplicates rrule-go's validateBounds, so a rule is
//     refused before rrule-go allocates for it. BYEASTER (rrule-go's extension, new) is bounded
//     to the offsets rrule-go can evaluate without panicking.
//   - a BYDAY ordinal rrule-go cannot evaluate (new): beyond the 5th within a month (MONTHLY, or
//     YEARLY with BYMONTH) or the 53rd within a year. rrule-go indexes its weekday mask with the
//     ordinal unchecked and panics, or, for an ordinal that never matches such as 7MO, iterates
//     to the year 9999 (~300 MB). Ordinals on other frequencies are left alone: RFC 5545 allows
//     them only for MONTHLY and YEARLY, but rrule-go ignores them there, so stored ones keep
//     working.
//   - more than maxOccurrencesPerPeriod occurrences per period (new): rrule-go materialises
//     every day of a period times the time set (BYHOUR x BYMINUTE x BYSECOND) before filtering,
//     so a yearly rule with a full time set (465 characters) allocated ~769 MB per evaluation.
//     The estimate is periodDays(FREQ) x max(1, |BYHOUR|) x max(1, |BYMINUTE|) x
//     max(1, |BYSECOND|), an upper bound since it ignores the day filters.
//   - a BYSETPOS beyond that estimate (new): it can never select anything, so rrule-go scanned
//     every period to the year 9999 (FREQ=DAILY;BYSETPOS=2..130 took ~5 s).
//   - more than maxBySetPos BYSETPOS values, a COUNT over maxRRuleCount and an INTERVAL spanning
//     more than the 100-year horizon (new).
//
// The write path also applies checkRRuleWriteLimits.
func checkRRuleBounds(opt *rrule.ROption) error {
	if opt.Count > maxRRuleCount {
		return ErrInvalidData{Message: fmt.Sprintf("Invalid RRULE: COUNT %d is too large; at most %d are allowed", opt.Count, maxRRuleCount)}
	}
	if intervalBeyondHorizon(opt) {
		return ErrInvalidData{Message: fmt.Sprintf("Invalid RRULE: INTERVAL %d spans more than %d years; at most %d for this frequency", opt.Interval, horizonYears, maxIntervalByHorizon[opt.Freq])}
	}

	for _, d := range rruleIntDomains {
		values := d.values(opt)
		seen := make(map[int]bool, len(values))
		for _, v := range values {
			inDomain := (v >= d.lo && v <= d.hi) || (d.signed && v <= -d.lo && v >= -d.hi)
			if !inDomain {
				return ErrInvalidData{Message: fmt.Sprintf("Invalid RRULE: %s value %d is out of range", d.part, v)}
			}
			if seen[v] {
				return ErrInvalidData{Message: fmt.Sprintf("Invalid RRULE: %s lists %d more than once", d.part, v)}
			}
			seen[v] = true
		}
	}

	maxOrdinal := 53
	if opt.Freq == rrule.MONTHLY || (opt.Freq == rrule.YEARLY && len(opt.Bymonth) > 0) {
		maxOrdinal = 5
	}
	type nthDay struct{ n, day int }
	seenDays := make(map[nthDay]bool, len(opt.Byweekday))
	for _, wd := range opt.Byweekday {
		if (opt.Freq == rrule.MONTHLY || opt.Freq == rrule.YEARLY) && (wd.N() > maxOrdinal || wd.N() < -maxOrdinal) {
			return ErrInvalidData{Message: fmt.Sprintf("Invalid RRULE: BYDAY position %d is out of range (at most %d)", wd.N(), maxOrdinal)}
		}
		if wd.N() > 53 || wd.N() < -53 {
			return ErrInvalidData{Message: fmt.Sprintf("Invalid RRULE: BYDAY position %d is out of range", wd.N())}
		}
		key := nthDay{wd.N(), wd.Day()}
		if seenDays[key] {
			return ErrInvalidData{Message: "Invalid RRULE: BYDAY lists " + strings.ToUpper(weekdayToString(wd)) + " more than once"}
		}
		seenDays[key] = true
	}

	occurrences := periodDays[opt.Freq]
	for _, times := range [][]int{opt.Byhour, opt.Byminute, opt.Bysecond} {
		occurrences *= max(1, len(times))
	}
	if occurrences > maxOccurrencesPerPeriod {
		return ErrInvalidData{Message: fmt.Sprintf("Invalid RRULE: the rule has up to %d occurrences per period; at most %d are allowed", occurrences, maxOccurrencesPerPeriod)}
	}
	// Duplicates were refused above (and dropped first for a stored rule), so this counts distinct values.
	if len(opt.Bysetpos) > maxBySetPos {
		return ErrInvalidData{Message: fmt.Sprintf("Invalid RRULE: BYSETPOS has %d values; at most %d are allowed", len(opt.Bysetpos), maxBySetPos)}
	}
	for _, pos := range opt.Bysetpos {
		if pos > occurrences || -pos > occurrences {
			return ErrInvalidData{Message: fmt.Sprintf("Invalid RRULE: BYSETPOS %d can never match; a period has at most %d occurrences", pos, occurrences)}
		}
	}
	return nil
}

// checkRRuleWriteLimits is checkRRuleBounds plus the limits only the write path applies: no
// negative INTERVAL or COUNT, and no BYDAY position below MONTHLY. A stored rule breaking these is
// still evaluated at completion.
func checkRRuleWriteLimits(opt *rrule.ROption) error {
	// rrule-go refuses a negative INTERVAL only when building the rule (which normalisation skips)
	// and reads a negative COUNT as unlimited; neither is a rule a client means to store.
	if opt.Interval < 0 {
		return ErrInvalidData{Message: fmt.Sprintf("Invalid RRULE: INTERVAL %d is negative", opt.Interval)}
	}
	if opt.Count < 0 {
		return ErrInvalidData{Message: fmt.Sprintf("Invalid RRULE: COUNT %d is negative", opt.Count)}
	}
	if err := checkRRuleBounds(opt); err != nil {
		return err
	}
	// RFC 5545 allows a BYDAY position (2MO, -1FR) only for MONTHLY and YEARLY rules. rrule-go
	// ignores it on other frequencies, so a stored one keeps working, but a write must not
	// store a position that means nothing.
	if opt.Freq != rrule.MONTHLY && opt.Freq != rrule.YEARLY {
		for _, wd := range opt.Byweekday {
			if wd.N() != 0 {
				return ErrInvalidData{Message: "Invalid RRULE: BYDAY " + strings.ToUpper(weekdayToString(wd)) +
					" has a position; positions are only allowed for MONTHLY and YEARLY rules (use a plain weekday such as " +
					strings.ToUpper(weekdayIntToString[wd.Day()]) + ")"}
			}
		}
	}
	return nil
}

// boundedForEvaluation returns a copy of opt that rrule-go can evaluate without unbounded work,
// for a stored rule that predates checkRRuleBounds: duplicate BY* values are dropped, which
// leaves the rule's occurrences unchanged. Anything else checkRRuleBounds refuses (a value
// outside its domain, a BYDAY ordinal rrule-go cannot evaluate, too many occurrences per
// period) is an error.
func boundedForEvaluation(opt *rrule.ROption) (rrule.ROption, error) {
	bounded := *opt
	dedupe := func(values []int) []int {
		if len(values) == 0 {
			return values
		}
		out := make([]int, 0, len(values))
		seen := make(map[int]bool, len(values))
		for _, v := range values {
			if !seen[v] {
				seen[v] = true
				out = append(out, v)
			}
		}
		return out
	}
	bounded.Bysecond = dedupe(opt.Bysecond)
	bounded.Byminute = dedupe(opt.Byminute)
	bounded.Byhour = dedupe(opt.Byhour)
	bounded.Bymonthday = dedupe(opt.Bymonthday)
	bounded.Byyearday = dedupe(opt.Byyearday)
	bounded.Byweekno = dedupe(opt.Byweekno)
	bounded.Bymonth = dedupe(opt.Bymonth)
	bounded.Bysetpos = dedupe(opt.Bysetpos)
	bounded.Byeaster = dedupe(opt.Byeaster)
	if len(opt.Byweekday) > 0 {
		bounded.Byweekday = make([]rrule.Weekday, 0, len(opt.Byweekday))
		seen := make(map[rrule.Weekday]bool, len(opt.Byweekday))
		for _, wd := range opt.Byweekday {
			if !seen[wd] {
				seen[wd] = true
				bounded.Byweekday = append(bounded.Byweekday, wd)
			}
		}
	}
	if err := checkRRuleBounds(&bounded); err != nil {
		return rrule.ROption{}, err
	}
	return bounded, nil
}

// rrulePolicyRejects reports whether a rule is barred by our recurrence policy.
// Sub-daily frequencies combined with any By* filter defeat the rrule library's
// day-skip optimization, so scanning for the next occurrence can iterate
// second-by-second over a large window (DoS). Nothing legitimate produces this:
// the legacy migration and the UI only author plain sub-daily interval rules with
// no By-parts.
func rrulePolicyRejects(opt *rrule.ROption) bool {
	return opt.Freq >= rrule.HOURLY && rruleHasByPart(opt)
}

func rruleHasByPart(opt *rrule.ROption) bool {
	return len(opt.Bysetpos) > 0 || len(opt.Bymonth) > 0 || len(opt.Bymonthday) > 0 ||
		len(opt.Byyearday) > 0 || len(opt.Byweekno) > 0 || len(opt.Byweekday) > 0 ||
		len(opt.Byhour) > 0 || len(opt.Byminute) > 0 || len(opt.Bysecond) > 0 ||
		len(opt.Byeaster) > 0
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

	r.ReadOnly = projectionLoses(opt, r)

	return r
}

// projectionLoses reports whether serialising the projection r of the stored rule opt yields a
// rule that is not semantically equal to it. It compares parsed options, never evaluated ones:
// rrule.NewRRule injects a default DTSTART, which would hide a stored one.
func projectionLoses(opt *rrule.ROption, r *TaskRepeat) bool {
	projected, err := r.toRRule()
	if err != nil {
		return true
	}
	projectedOpt, err := rrule.StrToROption(projected)
	if err != nil {
		return true
	}
	return normaliseROption(opt) != normaliseROption(projectedOpt)
}

// normaliseROption renders parsed rule options canonically, so two rules that differ only in
// semantics-preserving ways compare equal: INTERVAL 0 reads as 1, list order is ignored and nil
// equals empty, BYDAY compares (n, day) pairs, UNTIL compares in UTC, and the default WKST (MO)
// equals an unset one. DTSTART and BYEASTER are included, so a rule holding either differs from
// any projection.
func normaliseROption(opt *rrule.ROption) string {
	interval := opt.Interval
	if interval == 0 {
		interval = 1
	}
	days := make([]string, len(opt.Byweekday))
	for i, wd := range opt.Byweekday {
		days[i] = fmt.Sprintf("%d%s", wd.N(), weekdayIntToString[wd.Day()])
	}
	slices.Sort(days)
	return fmt.Sprintf(
		"freq=%d;interval=%d;wkst=%d;count=%d;until=%s;dtstart=%s;byday=%q;bysetpos=%v;bymonth=%v;"+
			"bymonthday=%v;byyearday=%v;byweekno=%v;byhour=%v;byminute=%v;bysecond=%v;byeaster=%v",
		opt.Freq, interval, opt.Wkst.Day(), opt.Count, canonicalTime(opt.Until), canonicalTime(opt.Dtstart),
		days, sortedInts(opt.Bysetpos), sortedInts(opt.Bymonth), sortedInts(opt.Bymonthday),
		sortedInts(opt.Byyearday), sortedInts(opt.Byweekno), sortedInts(opt.Byhour),
		sortedInts(opt.Byminute), sortedInts(opt.Bysecond), sortedInts(opt.Byeaster),
	)
}

// repeatsEqual reports whether two structured repeats describe the same rule, ignoring ReadOnly.
// It decides whether an incoming repeat is an unchanged echo of the stored one. The only
// normalisations are semantics-preserving: INTERVAL 0 reads as 1, lists are compared sorted (BYDAY
// case-insensitively), UNTIL is compared in UTC, WKST is case-insensitive with "" equal to "mo", and
// FREQ is case-insensitive.
func repeatsEqual(a, b *TaskRepeat) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return repeatFingerprint(a) == repeatFingerprint(b)
}

func repeatFingerprint(r *TaskRepeat) string {
	interval := r.Interval
	if interval == 0 {
		interval = 1
	}
	days := make([]string, len(r.ByDay))
	for i, d := range r.ByDay {
		days[i] = strings.ToLower(d)
	}
	slices.Sort(days)
	until := ""
	if r.Until != nil {
		if t, err := time.Parse(time.RFC3339, *r.Until); err == nil {
			until = canonicalTime(t)
		} else {
			until = "unparsed:" + *r.Until
		}
	}
	wkst := strings.ToLower(r.Wkst)
	if wkst == "" {
		wkst = "mo"
	}
	return fmt.Sprintf(
		"freq=%s;interval=%d;wkst=%s;count=%d;until=%s;byday=%q;bysetpos=%v;bymonth=%v;"+
			"bymonthday=%v;byyearday=%v;byweekno=%v;byhour=%v;byminute=%v;bysecond=%v",
		strings.ToLower(r.Freq), interval, wkst, r.Count, until, days,
		sortedInts(r.BySetPos), sortedInts(r.ByMonth), sortedInts(r.ByMonthDay), sortedInts(r.ByYearDay),
		sortedInts(r.ByWeekNo), sortedInts(r.ByHour), sortedInts(r.ByMinute), sortedInts(r.BySecond),
	)
}

// sortedInts returns a sorted copy, so callers' slices (often aliased from parsed options) are
// never reordered. nil and empty both yield an empty slice.
func sortedInts(v []int) []int {
	out := slices.Clone(v)
	slices.Sort(out)
	if out == nil {
		out = []int{}
	}
	return out
}

// canonicalTime renders t in UTC; the zero time renders as "".
func canonicalTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
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
