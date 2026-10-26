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
	"math"
	"time"

	"github.com/teambition/rrule-go"
)

// The wall-clock engine behind nextOccurrence.

// lookAround is the bound the normalisation rests on: no transition moves a wall clock by more
// than 24 hours, so an occurrence whose floating value is more than 48 hours from another's
// normalises on the same side of it.
const lookAround = 48 * time.Hour

// ruleFamily is how one step of a rule is computed.
type ruleFamily int

const (
	// familyElapsed: SECONDLY, MINUTELY and HOURLY rules, and their companion dates, step by
	// elapsed time, independent of the zone.
	familyElapsed ruleFamily = iota
	// familyWallclock: DAILY to YEARLY rules, with or without BY parts, COUNT or UNTIL, are
	// evaluated on the zone's local calendar and wall clock.
	familyWallclock
)

// noDuePolicy is which dates of a task without a due date move, and how far (unchanged from the
// earlier fixed-duration engine). It is decided apart from the family.
type noDuePolicy int

const (
	// policyIndependent: each date jumps past now on its own by whole steps. Exactly the set the
	// earlier fixed-duration engine stepped by a fixed duration: plain SECONDLY to WEEKLY rules with
	// no BY part, COUNT or UNTIL.
	policyIndependent noDuePolicy = iota
	// policyOneStep: each date takes exactly one rule step, and the end keeps its gap to the start.
	policyOneStep
)

// familyOf returns the rule's arithmetic family.
func familyOf(opt *rrule.ROption) ruleFamily {
	switch opt.Freq {
	case rrule.SECONDLY, rrule.MINUTELY, rrule.HOURLY:
		return familyElapsed
	case rrule.DAILY, rrule.WEEKLY, rrule.MONTHLY, rrule.YEARLY:
		return familyWallclock
	default:
		return familyWallclock
	}
}

// policyOf returns the rule's no-due-date policy: independent exactly for the rules the earlier
// fixed-duration engine stepped by a fixed duration (plain, SECONDLY to WEEKLY, a step that fits
// time.Duration).
func policyOf(opt *rrule.ROption) noDuePolicy {
	if _, ok := fixedDurationForRRule(opt); ok || wallclockPeriodDays(opt) > 0 {
		return policyIndependent
	}
	return policyOneStep
}

// plainRule reports whether the rule has no BY part (BYEASTER included), no COUNT and no UNTIL.
func plainRule(opt *rrule.ROption) bool {
	return opt.Count <= 0 && opt.Until.IsZero() &&
		len(opt.Bymonthday) == 0 && len(opt.Bymonth) == 0 && len(opt.Byyearday) == 0 &&
		len(opt.Byweekno) == 0 && len(opt.Bysetpos) == 0 && len(opt.Byhour) == 0 &&
		len(opt.Byminute) == 0 && len(opt.Bysecond) == 0 && len(opt.Byweekday) == 0 &&
		len(opt.Byeaster) == 0
}

// wallclockPeriodDays is the period of a plain DAILY or WEEKLY rule in local calendar days,
// INTERVAL x 1 or 7, or 0 for any other rule. A period whose length in seconds does not fit
// time.Duration was never a fixed duration in the earlier fixed-duration engine, so it is not one
// here either (such an INTERVAL is far beyond the recurrence horizon anyway).
func wallclockPeriodDays(opt *rrule.ROption) int {
	if !plainRule(opt) || (opt.Freq != rrule.DAILY && opt.Freq != rrule.WEEKLY) {
		return 0
	}
	interval, days := opt.Interval, 1
	if interval <= 0 {
		interval = 1
	}
	if opt.Freq == rrule.WEEKLY {
		days = 7
	}
	if int64(interval) > math.MaxInt64/int64(time.Duration(days)*24*time.Hour) {
		return 0
	}
	return interval * days
}

// recurrence is one completion's view of a rule: the parsed rule, its attributes, and the single
// "now" and zone every date of the completion is computed from.
type recurrence struct {
	opt     *rrule.ROption // as stored
	bounded rrule.ROption  // what rrule-go evaluates (boundedForEvaluation)
	family  ruleFamily
	policy  noDuePolicy
	// namesTime: the rule has a BYHOUR, BYMINUTE or BYSECOND part.
	namesTime bool
	// step is a plain sub-daily rule's elapsed step; untilStep and until are set for a plain
	// sub-daily rule with UNTIL (plainUntilStep). Both are zero for any other rule.
	step, untilStep time.Duration
	until           time.Time
	// periodDays is a plain DAILY or WEEKLY rule's period in local days, else 0.
	periodDays int
	// stepsDueDate: this recurrence steps a due date, so an occurrence whose wall clock does not
	// exist in the zone is skipped (RFC 5545 §3.3.10). Set once in newRecurrence.
	stepsDueDate bool
	now          time.Time // in zone
	zone         *time.Location
}

// occurrence is a moved (or reference) date: its instant, and the floating value it stands for,
// its wall clock. For an instant a gap shifted forward (a date that is not the due date) that is
// the unshifted wall clock.
type occurrence struct {
	at time.Time
	f  time.Time
}

func newRecurrence(in nextOccurrenceInput, opt *rrule.ROption, bounded rrule.ROption) (*recurrence, error) {
	r := &recurrence{
		opt:       opt,
		bounded:   bounded,
		family:    familyOf(opt),
		policy:    policyOf(opt),
		namesTime: len(opt.Byhour) > 0 || len(opt.Byminute) > 0 || len(opt.Bysecond) > 0,
		now:       in.Now.In(in.Zone),
		zone:      in.Zone,
		// Exactly the tasks withDueDate moves (evaluateNextOccurrence's switch); the other two
		// paths step dates that are not recurrence instances and keep the forward reading.
		stepsDueDate: !in.Due.At.IsZero(),
	}
	if r.family == familyElapsed {
		r.step, _ = fixedDurationForRRule(opt)
		r.untilStep, r.until = plainUntilStep(opt)
	} else {
		r.periodDays = wallclockPeriodDays(opt)
	}
	// rrule-go must be able to build the rule; an explicit DTSTART keeps the clock out of it.
	check := bounded
	check.Dtstart = r.float(r.now)
	if _, err := rrule.NewRRule(check); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *recurrence) float(t time.Time) time.Time    { return float(t, r.zone) }
func (r *recurrence) localise(f time.Time) time.Time { return localise(f, r.zone) }

// instant is f's instant in the zone (localise), and whether f exists there: false in a
// spring-forward gap, where the instant reads as f moved forward by the gap.
func (r *recurrence) instant(f time.Time) (time.Time, bool) {
	at := r.localise(f)
	return at, r.float(at).Equal(f)
}

// skips reports whether a candidate that does not exist in the zone is dropped: only while the
// due date is stepped.
func (r *recurrence) skips(exists bool) bool { return r.stepsDueDate && !exists }

// --- stepping one date ---

// next returns the rule's first occurrence strictly after from, for the rule anchored at a date:
// anchorF is the date's floating value (wallclock family), anchorAt its instant (elapsed family).
// clamp allows the anchor clamp, which only the due date takes. false means no occurrence is left.
// When the due date is stepped, a wallclock occurrence whose wall clock does not exist in the zone
// is skipped and never ranked (RFC 5545 §3.3.10); elapsed rules never meet one.
func (r *recurrence) next(anchorF, anchorAt, from time.Time, clamp bool) (occurrence, bool) {
	if r.family == familyElapsed {
		at := r.nextElapsed(anchorAt, from)
		return occurrence{at: at, f: r.float(at)}, !at.IsZero()
	}
	if r.periodDays > 0 {
		return r.nextWholePeriod(anchorF, from), true
	}
	if r.opt.Count > 0 {
		// COUNT counts from DTSTART, so a COUNT rule is never clamped.
		return r.nextCounted(anchorF, from)
	}
	dtstart := anchorF
	if clamp {
		// BY-part DAILY/WEEKLY anchors skip whole periods to within a year of now. On the floating
		// timeline the day count is exact, so the phase is kept.
		dtstart = clampRRuleAnchor(r.opt, anchorF, r.float(r.now))
	}
	return r.nextIterated(dtstart, from)
}

// nextElapsed steps a sub-daily rule on the instant timeline, unchanged from the earlier
// fixed-duration engine: plain rules by whole elapsed steps, plain UNTIL rules the same with UNTIL applied to the result at
// second precision (rrule-go truncates DTSTART to the second), and COUNT rules with rrule-go on
// UTC, where no transition can make a step anything but elapsed. Zero means none is left.
func (r *recurrence) nextElapsed(anchor, from time.Time) time.Time {
	switch {
	case r.step > 0:
		return addRepeatIntervalToTime(from, anchor, r.step)
	case r.untilStep > 0:
		next := addRepeatIntervalToTime(from, anchor, r.untilStep)
		if next.Truncate(time.Second).After(r.until) {
			return time.Time{}
		}
		return next
	default:
		o := r.bounded
		o.Dtstart = anchor.UTC()
		rule, err := rrule.NewRRule(o)
		if err != nil {
			return time.Time{}
		}
		return rule.After(from, false)
	}
}

// nextWholePeriod steps a plain DAILY or WEEKLY rule: whole periods of local calendar days from
// the anchor, past from, in O(1). The first candidate is at least one period after the anchor, so
// a date that is already after from still takes one step, as the earlier fixed-duration engine's
// step did. A due-date candidate in a gap is skipped; a period's drift against the transition
// date's weekday window ends any run of skips within a few years (research R-3).
func (r *recurrence) nextWholePeriod(anchorF, from time.Time) occurrence {
	n := r.periodDays
	j := 1
	if k := daysBetween(anchorF, r.float(from)) / n; k > 2 {
		j = k - 1
	}
	for ; ; j++ {
		f := anchorF.AddDate(0, 0, j*n)
		at, exists := r.instant(f)
		if r.skips(exists) {
			continue
		}
		if at.After(from) {
			return occurrence{at: at, f: f}
		}
	}
}

// floatingRule is the rule rrule-go iterates on the floating timeline from dtstart, without native
// COUNT (the caller ranks) and with the conservative UNTIL float(UNTIL) + 48 h, so no occurrence
// whose normalised instant is within the real UNTIL is cut off by rrule-go. It also returns the
// real UNTIL at second precision, zero when the rule has none.
func (r *recurrence) floatingRule(dtstart time.Time) (*rrule.RRule, time.Time, error) {
	o := r.bounded
	o.Dtstart = dtstart
	if o.Count > 0 {
		o.Count = 0
	}
	var until time.Time
	if !o.Until.IsZero() {
		until = o.Until.Truncate(time.Second)
		o.Until = r.float(until).Add(lookAround)
	}
	rule, err := rrule.NewRRule(o)
	return rule, until, err
}

// withinUntil reports whether a normalised instant is at or before the rule's real UNTIL, compared
// at second precision.
func withinUntil(at, until time.Time) bool {
	return until.IsZero() || !at.Truncate(time.Second).After(until)
}

// nextIterated chooses the next occurrence of a rule without COUNT in one rrule.Iterator pass:
// floating occurrences more than 48 h before from cannot normalise after it and are passed over;
// every later one is normalised, and the earliest normalised instant strictly after from (and
// within UNTIL) is the best. A due-date occurrence in a gap is skipped before it is compared. The
// pass stops at the first floating occurrence more than 48 h after the best instant's wall clock, a
// proven lower bound on every later one. The stop is relative to the best candidate, not to from.
func (r *recurrence) nextIterated(dtstart, from time.Time) (occurrence, bool) {
	rule, until, err := r.floatingRule(dtstart)
	if err != nil {
		return occurrence{}, false
	}
	next := rule.Iterator()
	skip := r.float(from).Add(-lookAround)
	var best occurrence
	var stop time.Time
	found := false
	for f, ok := next(); ok; f, ok = next() {
		if found && f.After(stop) {
			break
		}
		if !f.After(skip) {
			continue
		}
		at, exists := r.instant(f)
		if r.skips(exists) || !withinUntil(at, until) || !at.After(from) {
			continue
		}
		switch {
		case !found || at.Before(best.at):
			best, found = occurrence{at: at, f: f}, true
			stop = r.float(at).Add(lookAround)
		case at.Equal(best.at):
			best = r.representative(best, occurrence{at: at, f: f})
		}
	}
	return best, found
}

// nextCounted chooses the next occurrence of a wallclock rule with COUNT. COUNT counts the
// distinct normalised instants from DTSTART in time order: every floating occurrence from DTSTART
// is normalised (the look-back skip never affects the ranks), collisions count once, and since
// normalisation can reorder occurrences an instant's rank is final only once the iterator has
// passed its wall clock by 48 h. Ranks are assigned in time order, so the first final instant
// after from within COUNT is the answer; the rule is exhausted once COUNT instants are final. A
// due-date occurrence in a gap is skipped before it is buffered, so it never gets a rank and does
// not count.
func (r *recurrence) nextCounted(dtstart, from time.Time) (occurrence, bool) {
	rule, until, err := r.floatingRule(dtstart)
	if err != nil {
		return occurrence{}, false
	}
	next := rule.Iterator()
	var pending []occurrence // not yet final, distinct, in instant order
	var lastFinal time.Time
	rank := 0
	for {
		f, ok := next()
		for len(pending) > 0 && (!ok || f.After(r.float(pending[0].at).Add(lookAround))) {
			o := pending[0]
			pending = pending[1:]
			if rank++; rank > r.opt.Count {
				return occurrence{}, false
			}
			if o.at.After(from) {
				return o, true
			}
			lastFinal = o.at
		}
		if !ok {
			return occurrence{}, false
		}
		// The skip comes after the finality block above, which reads the incoming floating f only.
		at, exists := r.instant(f)
		if r.skips(exists) {
			continue
		}
		// An instant at or before a final one cannot occur (by lookAround); it would be a duplicate.
		if !withinUntil(at, until) || (rank > 0 && !at.After(lastFinal)) {
			continue
		}
		pending = r.insertPending(pending, occurrence{at: at, f: f})
	}
}

// insertPending adds an occurrence to the instant-ordered buffer, merging a collision into the
// instant already there. Occurrences arrive almost in order, so the position is found from the end.
func (r *recurrence) insertPending(pending []occurrence, o occurrence) []occurrence {
	i := len(pending)
	for i > 0 && pending[i-1].at.After(o.at) {
		i--
	}
	if i > 0 && pending[i-1].at.Equal(o.at) {
		pending[i-1] = r.representative(pending[i-1], o)
		return pending
	}
	pending = append(pending, occurrence{})
	copy(pending[i+1:], pending[i:])
	pending[i] = o
	return pending
}

// representative is the floating value an instant keeps when two floating occurrences normalise
// to it: the one that is not shifted (its wall clock is the instant's local reading), else the
// earliest, which is the one already kept since rrule-go yields in order.
func (r *recurrence) representative(kept, other occurrence) occurrence {
	if !r.float(kept.at).Equal(kept.f) && r.float(other.at).Equal(other.f) {
		return other
	}
	return kept
}

// --- the three ways a task moves ---

// withDueDate moves the due date to the next occurrence, and every other date with it. In normal
// mode the rule is anchored at the due date's wall clock and the search starts at now, or at the
// due date when that is later; from completion it is anchored at now, whose clock time the
// re-based due date takes. In both modes a generated due date whose wall clock does not exist is
// skipped (recurrence.stepsDueDate).
func (r *recurrence) withDueDate(in nextOccurrenceInput) (nextOccurrenceResult, bool) {
	ref := occurrence{at: in.Due.At, f: r.float(in.Due.At)}
	var next occurrence
	var ok bool
	if in.FromCompletion {
		next, ok = r.next(r.float(r.now), r.now, r.now, false)
	} else {
		from := r.now
		if in.Due.At.After(r.now) {
			from = in.Due.At
		}
		next, ok = r.next(ref.f, in.Due.At, from, true)
	}
	if !ok {
		return nextOccurrenceResult{}, false
	}

	res := unmovedResult(in, "")
	res.Due = datedValue{At: next.at.In(r.zone)}
	if next.at.Equal(in.Due.At) {
		// The next occurrence is the stored due date itself: no companion moves.
		return res, true
	}
	res.Start = r.moveCompanion(in.Start, ref, next)
	res.End = r.moveCompanion(in.End, ref, next)
	for i, rem := range in.Reminders {
		res.Reminders[i] = r.moveCompanion(rem, ref, next)
	}
	return res, true
}

// fromCompletionWithoutDueDate moves a task without a due date that repeats from the current date,
// as upstream's setTaskDatesFromCurrentDateRepeat did: the earliest absolute reminder lands on the
// first occurrence after now and the others move with it; the start lands there too and the end
// moves with the start, or lands there itself when there is no start. The re-based dates take
// now's clock time. They are not recurrence instances, so one in a gap moves forward by the gap:
// the same rule and now can give a different instant here than for a due date in withDueDate,
// which skips the gap occurrence instead.
func (r *recurrence) fromCompletionWithoutDueDate(in nextOccurrenceInput) (nextOccurrenceResult, bool) {
	next, ok := r.next(r.float(r.now), r.now, r.now, false)
	if !ok {
		return nextOccurrenceResult{}, false
	}
	rebased := datedValue{At: next.at.In(r.zone)}
	res := unmovedResult(in, "")
	if first, ok := earliestSet(in.Reminders); ok {
		// Exactly one reminder is the reference: the first in input order among the earliest. Every
		// other one, an instant tie included, is a companion, so a reminder that collided with the
		// reference keeps its own local clock time and separates again.
		ref := occurrence{at: in.Reminders[first].At, f: r.float(in.Reminders[first].At)}
		for i, rem := range in.Reminders {
			if i == first {
				res.Reminders[i] = rebased
				continue
			}
			res.Reminders[i] = r.moveCompanion(rem, ref, next)
		}
	}
	switch {
	case !in.Start.At.IsZero():
		res.Start = rebased
		res.End = r.moveCompanion(in.End, occurrence{at: in.Start.At, f: r.float(in.Start.At)}, next)
	case !in.End.At.IsZero():
		res.End = rebased
	}
	return res, true
}

// earliestSet returns the index of the earliest set date, the first in input order among equal
// instants, so the choice is deterministic.
func earliestSet(dates []datedValue) (int, bool) {
	first := -1
	for i, d := range dates {
		if !d.At.IsZero() && (first < 0 || d.At.Before(dates[first].At)) {
			first = i
		}
	}
	return first, first >= 0
}

// withoutDueDate moves a task without a due date in normal mode: under the independent policy
// each date jumps past now on its own; under the one-step policy each takes exactly one rule step
// from itself, and the end moves with the start when both are set. Nothing moves unless every
// date has a next occurrence.
func (r *recurrence) withoutDueDate(in nextOccurrenceInput) (nextOccurrenceResult, bool) {
	res := unmovedResult(in, "")
	for i, rem := range in.Reminders {
		if rem.At.IsZero() {
			continue
		}
		moved, _, ok := r.stepOnItsOwn(rem)
		if !ok {
			return nextOccurrenceResult{}, false
		}
		res.Reminders[i] = moved
	}
	var start occurrence
	if !in.Start.At.IsZero() {
		var ok bool
		if res.Start, start, ok = r.stepOnItsOwn(in.Start); !ok {
			return nextOccurrenceResult{}, false
		}
	}
	if !in.End.At.IsZero() {
		if r.policy == policyOneStep && !in.Start.At.IsZero() {
			res.End = r.moveCompanion(in.End, occurrence{at: in.Start.At, f: r.float(in.Start.At)}, start)
			return res, true
		}
		var ok bool
		if res.End, _, ok = r.stepOnItsOwn(in.End); !ok {
			return nextOccurrenceResult{}, false
		}
	}
	return res, true
}

// stepOnItsOwn moves one date of a task without a due date by the rule anchored at itself: past now
// under the independent policy, one step under the one-step policy. These dates are never the due
// date, so one in a gap moves forward by the gap.
func (r *recurrence) stepOnItsOwn(d datedValue) (datedValue, occurrence, bool) {
	from := d.At
	if r.policy == policyIndependent {
		from = r.now
	}
	o, ok := r.next(r.float(d.At), d.At, from, false)
	if !ok {
		return datedValue{}, occurrence{}, false
	}
	return datedValue{At: o.at.In(r.zone)}, o, true
}

// --- companions ---

// moveCompanion moves a companion date c with a reference that moved from refOld to refNew, by
// the rule's family:
//
//   - elapsed: by the reference's elapsed delta, on instants;
//   - wallclock, the rule names no time: by the reference's whole local-calendar-day delta, keeping
//     c's own local clock time;
//   - wallclock, the rule names a time: keeping c's calendar offset to the reference, never
//     landing on the other side of the new reference in calendar order; a companion normalisation
//     would put there is clamped to the reference's instant, which is always forward of its wall
//     clock.
//
// A companion is not a recurrence instance: one that lands in a gap moves forward by the gap, and
// one that is gap-shifted or clamped keeps the shifted time on later steps. An unset date stays
// unset.
func (r *recurrence) moveCompanion(c datedValue, refOld, refNew occurrence) datedValue {
	if c.At.IsZero() {
		return c
	}
	if r.family == familyElapsed {
		return datedValue{At: shiftTime(c.At, refOld.at, refNew.at).In(r.zone)}
	}
	cf := r.float(c.At)
	if !r.namesTime {
		return datedValue{At: r.localise(cf.AddDate(0, 0, daysBetween(refOld.f, refNew.f))).In(r.zone)}
	}
	at := r.localise(shiftTime(cf, refOld.f, refNew.f))
	if (cf.Before(refOld.f) && at.After(refNew.at)) || (cf.After(refOld.f) && at.Before(refNew.at)) {
		at = refNew.at
	}
	return datedValue{At: at.In(r.zone)}
}
