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
	"regexp"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"

	"github.com/kentonium3/rrule-go"
)

// First-occurrence selection: the one place that turns a rule, an initial_due_time, a
// zone and now into a frozen due date, or into no date and a reason. The preview and the v2 writes
// call this and nothing else.

// EngineNow is the one clock of a recurrence operation (preview, write), read through the engine
// clock timeNow and truncated to the second, the precision dates are stored at. An operation
// samples it once and passes the value down; selectInitialDue itself never reads a clock.
func EngineNow() time.Time {
	return timeNow().Truncate(time.Second)
}

// PinEngineClockForTests makes the engine clock (EngineNow, completion's timeNow) return now
// until the returned restore function runs; packages outside models (pkg/webtests) use it to pin
// "now" over HTTP. The clock is a package variable: tests that pin it must not run in parallel
// with each other or with anything that reads it.
func PinEngineClockForTests(now time.Time) (restore func()) {
	orig := timeNow
	timeNow = func() time.Time { return now }
	return func() { timeNow = orig }
}

// initialDueTimePattern is the HH:MM the HTTP schemas accept (00:00 to 23:59).
var initialDueTimePattern = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

// ParseInitialDueTime parses an initial_due_time, exactly HH:MM from 00:00 to 23:59. The HTTP
// schemas refuse anything else with 422 before this runs; it is defence in depth.
func ParseInitialDueTime(s string) (hour, minute int, err error) {
	if !initialDueTimePattern.MatchString(s) {
		return 0, 0, ErrInvalidData{Message: "initial_due_time must be HH:MM, from 00:00 to 23:59."}
	}
	hour, _ = strconv.Atoi(s[:2])
	minute, _ = strconv.Atoi(s[3:])
	return hour, minute, nil
}

// InitialDueConflictKind is which rule an initial_due_time breaks.
type InitialDueConflictKind string

const (
	// InitialDueConflictDueDate: the (merged) due date is set as well.
	InitialDueConflictDueDate InitialDueConflictKind = "due_date"
	// InitialDueConflictNoRule: there is no repeat rule to search.
	InitialDueConflictNoRule InitialDueConflictKind = "no_rule"
	// InitialDueConflictDone: the (merged) task is done.
	InitialDueConflictDone InitialDueConflictKind = "done"
)

// ErrInitialDueConflict is an initial_due_time that cannot be used with the rest of the input. It
// is not an HTTP error because the status depends on the caller: a write maps it with WriteError
// (400, code 2002), the preview with PreviewError (422). The message is the same on both.
type ErrInitialDueConflict struct {
	Kind    InitialDueConflictKind
	Field   string // the JSON field that conflicts with initial_due_time
	Message string
}

func (err ErrInitialDueConflict) Error() string { return err.Message }

// WriteError is the conflict as a v2 write returns it: ErrInvalidData, 400 with code 2002.
func (err ErrInitialDueConflict) WriteError() error { return ErrInvalidData{Message: err.Message} }

// PreviewError is the conflict as the preview returns it: an invalid-field error, which the v2
// error mapping turns into 422.
func (err ErrInitialDueConflict) PreviewError() error {
	return InvalidFieldErrorWithMessage([]string{"initial_due_time", err.Field}, err.Message)
}

// InitialDueConflict is the one conflict rule for an initial_due_time, shared by the preview and
// the writes so they cannot drift. The inputs are the merged state: dueDateSet is a
// non-zero due date, hasRule a non-empty repeat rule, done the effective done flag. An empty
// initialDueTime never conflicts. The format is ParseInitialDueTime's concern.
func InitialDueConflict(dueDateSet bool, initialDueTime string, hasRule, done bool) error {
	switch {
	case initialDueTime == "":
		return nil
	case dueDateSet:
		return ErrInitialDueConflict{Kind: InitialDueConflictDueDate, Field: "due_date",
			Message: "initial_due_time and due_date cannot both be set."}
	case !hasRule:
		return ErrInitialDueConflict{Kind: InitialDueConflictNoRule, Field: "repeats_rrule",
			Message: "initial_due_time needs a repeat rule."}
	case done:
		return ErrInitialDueConflict{Kind: InitialDueConflictDone, Field: "done",
			Message: "initial_due_time cannot be used on a done task."}
	}
	return nil
}

// initialSearchRule is the evaluation copy of a stored rule for the initial search: the rule
// without COUNT, so completion's COUNT guards (COUNT=1 is exhausted, the COUNT-1 write-back
// length) do not apply. The stored rule keeps its COUNT, which then counts the selected date as
// instance 1. A rule that does not parse is returned unchanged for the guards to refuse.
func initialSearchRule(rule string) string {
	opt, err := parseRuleText(rule)
	if err != nil || opt.Count == 0 {
		return rule
	}
	opt.Count = 0
	return opt.RRuleString()
}

// selectInitialDue is a new task's first due date: the first instance of rule strictly after
// now, searched from the seed "today in zone at hour:minute" and within UNTIL and the engine's
// bounds. COUNT is ignored during the search. The seed lives on the floating timeline and is never
// an instant, so a seed in a spring-forward gap keeps its nominal phase; a gap candidate is
// skipped, a fall-back one takes the earlier instant (localise). The result is in the server zone
// at second precision.
//
// No date comes with a reason from the occurrences query's vocabulary: UNTIL before now is until_passed; no
// instance between now and UNTIL, or none in the search window, is no_occurrence; a rule a guard
// refuses has that guard's reason (invalid_rule, interval_beyond_horizon, outside_bounds,
// policy_rejected, ...). Pure: no session, no clock, no process zone; now comes in truncated.
func selectInitialDue(rule string, hour, minute int, zone *time.Location, now time.Time) (due time.Time, reason noNextReason) {
	in := nextOccurrenceInput{Rule: initialSearchRule(rule), Now: now, Zone: zone}
	opt, bounded, reason := checkNextOccurrenceGuards(in)
	if reason != "" {
		return time.Time{}, reason
	}
	return searchInitialDue(in, opt, bounded, rule, hour, minute)
}

// searchInitialDue is selectInitialDue's evaluation stage, after the guards: the search and its
// panic recovery. The bounds keep every known panic out of rrule-go; as defence in depth a panic
// here is logged and selects nothing (evaluation_failed), as evaluateNextOccurrence does for
// completion. rule is the caller's rule text, for log lines only; the search reads in.Rule.
func searchInitialDue(in nextOccurrenceInput, opt *rrule.ROption, bounded rrule.ROption, rule string, hour, minute int) (due time.Time, reason noNextReason) {
	now := in.Now
	defer func() {
		if p := recover(); p != nil {
			log.Errorf("Not selecting a first due date: evaluating RRULE '%s' panicked: %v", rule, p)
			due, reason = time.Time{}, reasonEvaluationFailed
		}
	}()

	r, err := newInitialSearch(in, opt, bounded)
	if err != nil {
		log.Errorf("Failed to parse RRULE '%s': %v", rule, err)
		return time.Time{}, reasonInvalidRule
	}
	today := now.In(in.Zone)
	seed := time.Date(today.Year(), today.Month(), today.Day(), hour, minute, 0, 0, time.UTC)
	o, ok := r.next(seed, time.Time{}, now, false)
	if !ok {
		return time.Time{}, reasonNoOccurrence
	}
	if (nextOccurrenceResult{Due: datedValue{At: o.at}}).outsideStorableRange() {
		return time.Time{}, reasonOutsideStorableRange
	}
	return o.at.Truncate(time.Second).In(config.GetTimeZone()), ""
}
