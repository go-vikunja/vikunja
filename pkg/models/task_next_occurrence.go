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
	"time"

	"code.vikunja.io/api/pkg/log"

	"github.com/teambition/rrule-go"
)

// noNextReason says why nextOccurrence moved nothing. The strings are stable: other code and
// tests assert on them.
type noNextReason string

const (
	reasonNoZone                noNextReason = "no_zone"
	reasonNoRule                noNextReason = "no_rule"
	reasonInvalidRule           noNextReason = "invalid_rule"
	reasonIntervalBeyondHorizon noNextReason = "interval_beyond_horizon"
	reasonAnchorBeyondHorizon   noNextReason = "anchor_beyond_horizon"
	reasonOutsideBounds         noNextReason = "outside_bounds"
	reasonPolicyRejected        noNextReason = "policy_rejected"
	reasonRuleTooLong           noNextReason = "rule_too_long"
	reasonCountExhausted        noNextReason = "count_exhausted"
	reasonUntilPassed           noNextReason = "until_passed"
	reasonEvaluationFailed      noNextReason = "evaluation_failed"
	reasonNoOccurrence          noNextReason = "no_occurrence"
	reasonOutsideStorableRange  noNextReason = "outside_storable_range"
)

// datedValue is one date that steps, with its remembered intended clock time. Intended is
// meaningful on Due only; it is ignored on every other date and nil on every other output.
type datedValue struct {
	At       time.Time      // the stored instant; zero = the date is not set
	Intended *time.Duration // the due date's intended local time of day (0 ≤ d < 24h), or nil
	// Deprecated: ignored on input and always "" on output; retained for now, the tests assert it
	// stays empty.
	IntendedZone string
}

// nextOccurrenceInput is everything nextOccurrence reads.
type nextOccurrenceInput struct {
	Rule           string     // stored RRULE text: no "RRULE:" prefix, no DTSTART
	Due            datedValue // the anchor when set
	Start, End     datedValue
	Reminders      []datedValue   // absolute reminders only; relative ones are recomputed by completion
	Now            time.Time      // the completion instant
	FromCompletion bool           // repeats_from_current_date
	Zone           *time.Location // the resolved zone; never defaulted here
	TaskID         int64          // log lines only
}

// nextOccurrenceResult is a task's dates at its next occurrence, or the reason there is none.
type nextOccurrenceResult struct {
	Due, Start, End datedValue   // moved dates in Zone; Due carries the Intended to store (nil = store NULL), the others nil
	Reminders       []datedValue // same order as the input
	Reason          noNextReason // "" when the dates moved; otherwise nothing moved
}

// nextOccurrence is the next occurrence of a repeating task's dates: the one function completion
// calls to move them. It first applies completion's guards in their order; the first that applies
// decides the reason, and nothing moves. It is pure: it reads neither the clock, the
// configuration, the database nor the process zone, and it never changes its input. It never
// decrements COUNT either; completion writes COUNT-1 back.
//
// The rule is evaluated in Zone: calendar-unit rules on its local calendar and wall clock,
// sub-daily ones by elapsed time. Moved dates are returned in Zone, the due date with the intended
// clock time to store; completion converts them to the server zone.
func nextOccurrence(in nextOccurrenceInput) nextOccurrenceResult {
	opt, bounded, reason := checkNextOccurrenceGuards(in)
	if reason != "" {
		return unmovedResult(in, reason).withoutCompanionIntent()
	}
	return evaluateNextOccurrence(in, opt, bounded).withoutCompanionIntent()
}

// withoutCompanionIntent is the result as every path returns it, the ones that echo an input date
// included: Start, End and each reminder carry no Intended, and no date carries an IntendedZone.
// The due date's Intended is left as it is. The reminders slice is always the result's own copy
// (unmovedResult), never the caller's.
func (res nextOccurrenceResult) withoutCompanionIntent() nextOccurrenceResult {
	res.Due.IntendedZone = ""
	res.Start, res.End = datedValue{At: res.Start.At}, datedValue{At: res.End.At}
	for i := range res.Reminders {
		res.Reminders[i] = datedValue{At: res.Reminders[i].At}
	}
	return res
}

// checkNextOccurrenceGuards applies completion's guards, in the order completion has always
// applied them, and returns the parsed rule and the bounded copy rrule-go evaluates.
func checkNextOccurrenceGuards(in nextOccurrenceInput) (opt *rrule.ROption, bounded rrule.ROption, reason noNextReason) {
	if in.Zone == nil {
		return nil, bounded, reasonNoZone
	}
	if in.Rule == "" {
		return nil, bounded, reasonNoRule
	}

	opt, err := rrule.StrToROption(in.Rule)
	if err != nil {
		log.Errorf("Failed to parse RRULE '%s' for task %d: %v", in.Rule, in.TaskID, err)
		return nil, bounded, reasonInvalidRule
	}
	// The recurrence horizon: a rule whose INTERVAL spans more than 100 years, or a task whose due
	// date lies more than 200 years back, is not advanced. Each is logged with its own reason,
	// apart from COUNT or UNTIL running out.
	if intervalBeyondHorizon(opt) {
		log.Errorf("Not advancing task %d: its RRULE's INTERVAL spans more than %d years, beyond the recurrence horizon", in.TaskID, horizonYears)
		return nil, bounded, reasonIntervalBeyondHorizon
	}
	if dueDateBeyondHorizon(in.Due.At, in.Now) {
		log.Errorf("Not advancing task %d: its due date is more than %d years in the past, beyond the recurrence horizon", in.TaskID, maxAnchorAgeYears)
		return nil, bounded, reasonAnchorBeyondHorizon
	}

	// A stored rule may predate the write-time safety bound. It is evaluated without its
	// duplicate values (same occurrences, bounded work); one rrule-go cannot evaluate at all
	// leaves the task done and its dates unchanged, like a rule that no longer parses.
	bounded, err = boundedForEvaluation(opt)
	if err != nil {
		log.Errorf("Not advancing task %d: its RRULE is outside the supported bounds: %v", in.TaskID, err)
		return nil, bounded, reasonOutsideBounds
	}
	// Nor is a stored rule the creation policy rejects (sub-daily with a BY* part, which only rows
	// from before the policy and their verbatim duplicates can hold): some never match, and rrule-go
	// would iterate without returning. The task stays done, as for a rule outside the bounds.
	if rrulePolicyRejects(&bounded) {
		log.Errorf("Not advancing task %d: its RRULE combines a sub-daily frequency with a BY* part, which the recurrence policy rejects", in.TaskID)
		return nil, bounded, reasonPolicyRejected
	}
	// Completing a COUNT rule writes it back serialised with COUNT-1. rrule-go's serialisation
	// can be longer than the stored text (a date-only UNTIL gains a time), and a row from before
	// the length check could then no longer fit the repeats column: PostgreSQL and MySQL would
	// fail the update. Such a task is treated like the refusals above.
	if opt.Count > 1 {
		decremented := *opt
		decremented.Count--
		if serialised := decremented.RRuleString(); len(serialised) > maxRRuleLength {
			log.Errorf("Not advancing task %d: its RRULE would be %d characters long once written back; at most %d fit", in.TaskID, len(serialised), maxRRuleLength)
			return nil, bounded, reasonRuleTooLong
		}
	}
	// The stored COUNT is the number of occurrences remaining, including the one just completed
	// (COUNT 0 means no limit). When only one remains, the recurrence is exhausted.
	if opt.Count == 1 {
		return nil, bounded, reasonCountExhausted
	}
	// An UNTIL already behind now leaves no occurrence. It is checked up front because a date
	// without a due date steps from itself and could otherwise land on an occurrence that is
	// already in the past.
	if !opt.Until.IsZero() && opt.Until.Before(in.Now) {
		return nil, bounded, reasonUntilPassed
	}
	return opt, bounded, ""
}

// evaluateNextOccurrence is the evaluation stage of nextOccurrence, after its guards: it steps the
// dates with the rule (bounded, the options rrule-go evaluates) in in.Zone and range-checks the
// result.
//
// The bounds (checkRRuleBounds, the policy) keep every known panic out of rrule-go. As defence in
// depth, a panic during evaluation is recovered here: it is logged with the task id and nothing
// moves.
func evaluateNextOccurrence(in nextOccurrenceInput, opt *rrule.ROption, bounded rrule.ROption) (res nextOccurrenceResult) {
	defer func() {
		if p := recover(); p != nil {
			log.Errorf("Not advancing task %d: evaluating its RRULE panicked: %v", in.TaskID, p)
			res = unmovedResult(in, reasonEvaluationFailed)
		}
	}()

	r, err := newRecurrence(in, opt, bounded)
	if err != nil {
		log.Errorf("Failed to parse RRULE '%s' for task %d: %v", in.Rule, in.TaskID, err)
		return unmovedResult(in, reasonInvalidRule)
	}

	var moved bool
	switch {
	case !in.Due.At.IsZero():
		res, moved = r.withDueDate(in)
	case in.FromCompletion:
		res, moved = r.fromCompletionWithoutDueDate(in)
	default:
		res, moved = r.withoutDueDate(in)
	}
	if !moved {
		// A rule without COUNT or UNTIL never runs out (rrule-go reads a negative COUNT as none), so
		// finding no occurrence means rrule-go's search window ended first, or the rule can never
		// match. The window is DTSTART + ~292 years (time.Duration's range), capped at the year
		// 9999 (rrule-go's MAXYEAR). Logged apart from exhaustion; the task stays done either way.
		if opt.Count <= 0 && opt.Until.IsZero() {
			logNoOccurrenceInWindow(in.TaskID)
		}
		return unmovedResult(in, reasonNoOccurrence)
	}

	// A date outside 1000-01-01..9999-12-31 cannot be stored (MySQL's DATETIME range, the
	// strictest supported database): nothing moves. Completion checks relative reminders too.
	if res.outsideStorableRange() {
		log.Errorf("Not advancing task %d: a date of its next occurrence is outside the years 1000 to 9999", in.TaskID)
		return unmovedResult(in, reasonOutsideStorableRange)
	}
	return res
}

// unmovedResult returns the input's dates unchanged with the reason nothing moved, so a caller
// that applied them anyway would change nothing.
func unmovedResult(in nextOccurrenceInput, reason noNextReason) nextOccurrenceResult {
	reminders := make([]datedValue, len(in.Reminders))
	copy(reminders, in.Reminders)
	return nextOccurrenceResult{Due: in.Due, Start: in.Start, End: in.End, Reminders: reminders, Reason: reason}
}

// outsideStorableRange reports whether a set date of the result lies outside the years 1000 to
// 9999 in UTC, the range of MySQL's DATETIME.
func (res nextOccurrenceResult) outsideStorableRange() bool {
	dates := []datedValue{res.Due, res.Start, res.End}
	dates = append(dates, res.Reminders...)
	for _, d := range dates {
		if year := d.At.UTC().Year(); !d.At.IsZero() && (year < 1000 || year > 9999) {
			return true
		}
	}
	return false
}

// The floating timeline. A floating value is a wall-clock reading in a zone, written as a UTC
// time, so it has no DST: calendar-unit rules are evaluated on it, and localise is the one place a
// wall clock becomes an instant again. Floating values are never stored or returned.

// float returns t's wall clock in zone as a floating value: its local fields, rebuilt in UTC.
func float(t time.Time, zone *time.Location) time.Time {
	l := t.In(zone)
	return time.Date(l.Year(), l.Month(), l.Day(), l.Hour(), l.Minute(), l.Second(), l.Nanosecond(), time.UTC)
}

// localise returns the instant whose wall clock in zone is the floating value f, read in zone:
//
//   - a wall clock that occurs twice (an overlap) takes the earlier instant;
//   - one that does not occur (a gap) moves forward by the length of the gap, the pre-transition
//     offset's instant;
//   - any other has exactly one instant.
//
// The offsets are probed 48 hours either side of the wall clock, which is safe because no
// transition moves a wall clock by more than 24 hours. Go's own normalisation in time.Date (which
// moves a New York gap backwards, an overlap in Berlin to the later instant and Santiago's
// midnight gap to the previous day) is never relied on.
func localise(f time.Time, zone *time.Location) time.Time {
	w := time.Date(f.Year(), f.Month(), f.Day(), f.Hour(), f.Minute(), f.Second(), f.Nanosecond(), time.UTC)
	_, before := w.Add(-48 * time.Hour).In(zone).Zone()
	_, after := w.Add(48 * time.Hour).In(zone).Zone()
	c1 := w.Add(-time.Duration(before) * time.Second)
	c2 := w.Add(-time.Duration(after) * time.Second)
	valid1, valid2 := float(c1, zone).Equal(w), float(c2, zone).Equal(w)
	switch {
	case valid1 && valid2 && c2.Before(c1):
		return c2.In(zone)
	case valid1:
		return c1.In(zone)
	case valid2:
		return c2.In(zone)
	default:
		// A gap: the pre-transition offset's instant reads as f moved forward by the gap.
		return c1.In(zone)
	}
}

// timeOfDay is a floating value's local time of day, 0 <= d < 24h.
func timeOfDay(f time.Time) time.Duration {
	h, m, s := f.Clock()
	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second + time.Duration(f.Nanosecond())
}

// relativeReminderTime is the instant of a reminder periodSeconds relative to base. The period's
// whole days are calendar days in zone: they keep base's local clock time across a clock change, a
// wall clock in a gap moving forward by the gap and one in an overlap taking the earlier instant
// (localise). The remainder, and any period under a day, is elapsed time. A mixed period applies
// its days first, then the remainder, so the remainder cannot change which calendar day the days
// land on. A period under a day never reads zone, so an instant in an overlap keeps its identity;
// a nil zone with whole days is a caller bug and gets the old, elapsed value.
//
// The zone is the saver's or the completer's. There is no backfill: a stored reminder takes this
// rule at its task's next save or completion. CalDAV still exports a day as -PT24H, so a CalDAV
// client may fire a whole-day reminder an hour apart from Vikunja on a clock-change weekend: due
// Sun 2027-03-14 20:00 EDT, one day before, is Sat 20:00 EST here and Sat 19:00 EST there. Pure:
// it reads no configuration, clock or process zone.
func relativeReminderTime(base time.Time, periodSeconds int64, zone *time.Location) time.Time {
	days := periodSeconds / (24 * 60 * 60) // truncated toward zero
	rest := periodSeconds % (24 * 60 * 60) // the sign of periodSeconds
	if days == 0 || zone == nil {
		return base.Add(time.Duration(periodSeconds) * time.Second)
	}
	return localise(float(base, zone).AddDate(0, 0, int(days)), zone).Add(time.Duration(rest) * time.Second)
}
