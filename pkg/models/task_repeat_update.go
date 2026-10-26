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

import "github.com/teambition/rrule-go"

// repeatDecision says what the write path must store for a task's recurrence.
type repeatDecision struct {
	Repeats string // the RRULE text to store ("" = no recurrence)
	Keep    bool   // true when the stored text is kept byte-identical
}

// resolveIncomingRepeat decides the stored rule for a write. It owns the recurrence write
// semantics, so create, update, bulk, duplicate and CalDAV cannot diverge. It is pure: no DB access.
//
// stored: the rule currently stored ("" on create).
// incoming: the API's structured repeat (nil when absent or null).
// incomingRaw: an RRULE string set directly by a trusted caller (CalDAV, duplication, import), else "".
//
// The rules, in order:
//  0. incomingRaw equal to stored keeps it, without re-validation: a stored rule a later policy
//     rejects must not fail an unrelated update. So does an incomingRaw that is the stored rule
//     with a stale, higher COUNT (isStaleRawCount: CalDAV clients send the RRULE as last synced).
//  1. Any other incomingRaw is validated and stored verbatim. A non-nil incoming must agree with it.
//  2. A nil incoming clears the rule. At model level nil always means "clear"; v2 PATCH turns an
//     absent repeat into an equal echo before it gets here.
//  3. An incoming equal to the stored rule's projection keeps the stored text byte-identical.
//     So does a stale COUNT: see isStaleCount.
//  4. A stored rule whose projection is read-only refuses any other change.
//  5. Otherwise incoming is serialised, validated and stored.
//
// Masked updates whose field mask excludes recurrence must not call this function: the stored rule
// is left untouched and never re-validated.
func resolveIncomingRepeat(stored string, incoming *TaskRepeat, incomingRaw string) (repeatDecision, error) {
	if incomingRaw != "" {
		if incomingRaw == stored || isStaleRawCount(stored, incomingRaw) {
			return repeatDecision{Repeats: stored, Keep: true}, nil
		}
		if err := validateRRule(incomingRaw); err != nil {
			return repeatDecision{}, err
		}
		if incoming != nil && !repeatsEqual(incoming, taskRepeatFromRRule(incomingRaw)) {
			return repeatDecision{}, ErrInvalidData{Message: "The structured repeat and the RRULE text disagree; set only one of them."}
		}
		return repeatDecision{Repeats: incomingRaw}, nil
	}

	if incoming == nil {
		return repeatDecision{}, nil
	}

	if stored != "" {
		if projection := taskRepeatFromRRule(stored); projection != nil {
			if repeatsEqual(incoming, projection) || isStaleCount(incoming, projection) {
				return repeatDecision{Repeats: stored, Keep: true}, nil
			}
			if projection.ReadOnly {
				return repeatDecision{}, ErrInvalidData{Message: "This task's recurrence rule has parts the structured repeat cannot represent, " +
					"so it cannot be edited here. It can be removed (send repeat: null) or replaced over CalDAV."}
			}
		}
	}

	text, err := incoming.toRRule()
	if err != nil {
		return repeatDecision{}, err
	}
	if err := validateRRule(text); err != nil {
		return repeatDecision{}, err
	}
	return repeatDecision{Repeats: text}, nil
}

// isStaleCount reports whether incoming is the stored rule (its projection) echoed with a
// COUNT read before the server decremented it.
//
// The stored COUNT is the number of occurrences left, and the server lowers it on every
// completion (setTaskDatesRRule), while clients round-trip the whole structured repeat on
// every save. A client that read the task before a completion (a second tab, a cached copy,
// an editor left open) sends the old, higher COUNT back with an unrelated change, which would
// make the rule run past its limit. The rule we apply:
//
//   - every part except COUNT equal to the stored rule, and a COUNT greater than the stored
//     remaining COUNT: a stale echo, so the stored rule (and its COUNT) is kept;
//   - a lower COUNT, a COUNT removed (0) or added to an unlimited rule, or any other part
//     changed: a deliberate edit, saved as sent.
//
// A completion can only ever lower COUNT, so a stale echo always carries a higher one; a
// user raising the COUNT of an otherwise unchanged rule is the one edit this rule cannot tell
// from a stale echo, so it keeps the stored COUNT (to raise it, clear the rule and set it
// again). Lowering, removing or changing the rule always goes through. It applies whether or
// not the update also completes the task: a completion decrements from the stored rule anyway.
//
// A raw RRULE from CalDAV follows the same rule (isStaleRawCount).
// Not covered: an exhausted rule, which completion clears, so there is no stored COUNT to compare
// against; a stale client re-setting it is accepted last-write-wins.
func isStaleCount(incoming, projection *TaskRepeat) bool {
	if projection.Count <= 0 || incoming.Count <= projection.Count {
		return false
	}
	echo := *incoming
	echo.Count = projection.Count
	return repeatsEqual(&echo, projection)
}

// isStaleRawCount is isStaleCount for raw RRULE text (CalDAV PUT): incoming equals the stored rule
// in every part except COUNT, and its COUNT is higher than the stored remaining one. Both are
// compared as parsed options, so parts the structured repeat cannot hold (DTSTART, BYEASTER)
// count too; any other difference is a deliberate change and replaces the rule.
func isStaleRawCount(stored, incoming string) bool {
	if stored == "" {
		return false
	}
	storedOpt, err := rrule.StrToROption(stored)
	if err != nil || storedOpt.Count <= 0 {
		return false
	}
	incomingOpt, err := rrule.StrToROption(incoming)
	if err != nil || incomingOpt.Count <= storedOpt.Count {
		return false
	}
	echo := *incomingOpt
	echo.Count = storedOpt.Count
	return normaliseROption(&echo) == normaliseROption(storedOpt)
}

// resolveCreateRepeat decides the stored rule for a task being created: a rule copied from
// another task's stored rule (copiesStoredRule) skips the creation policy, everything else goes
// through resolveIncomingRepeat.
func resolveCreateRepeat(t *Task) (repeatDecision, error) {
	if t.copiesStoredRule {
		return resolveCopiedRepeat(t.Repeats)
	}
	return resolveIncomingRepeat("", t.Repeat, t.Repeats)
}

// resolveCopiedRepeat decides the stored rule for a task copied from another task's stored rule
// (duplication). The text is kept verbatim after a syntax and safety-bound check
// only, never the creation policy: a rule the policy rejects
// today, stored before the policy existed or by a CalDAV client, must not make the copy fail or
// silently lose its rule. The copy's exposure equals the original row's.
func resolveCopiedRepeat(raw string) (repeatDecision, error) {
	if _, err := validateRRuleSyntax(raw); err != nil {
		return repeatDecision{}, err
	}
	return repeatDecision{Repeats: raw}, nil
}

// DropUnimportableRepeat checks an imported task's rule the way creation will, and removes it
// when creation would refuse it: unparseable, too long, outside the safety bound, or rejected
// by the creation policy. The rule and the from-current-date flag are cleared and the reason
// returned, so the importer can log it and import the task without a rule; nil means the rule
// is kept. An import file is user-supplied, so a rule the policy rejects is never copied
// verbatim from it (the policy also guards against rules that never return, e.g.
// FREQ=HOURLY;INTERVAL=2;BYHOUR=1).
func DropUnimportableRepeat(t *Task) error {
	if _, err := resolveIncomingRepeat("", t.Repeat, t.Repeats); err != nil {
		t.Repeats = ""
		t.Repeat = nil
		t.RepeatsFromCurrentDate = false
		return err
	}
	return nil
}
