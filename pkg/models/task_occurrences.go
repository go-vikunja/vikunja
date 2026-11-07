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
	"errors"
	"fmt"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

const (
	// DefaultOccurrencesCount is how many occurrences a query returns when it asks for no count.
	DefaultOccurrencesCount = 10
	// MaxOccurrencesCount is the most occurrences one query may ask for. N=100 on the worst rules
	// of TestNextOccurrenceTiming stays within its 2 s ceiling (TestTaskOccurrencesTiming).
	MaxOccurrencesCount = 100
)

// reasonNoDueDate is the API-level reason a task without a due date has no due dates to list:
// the engine's no-due-date branch moves reminders, not due dates.
const reasonNoDueDate = "no_due_date"

// reasonTaskDone is the API-level reason a task that is already done has no due dates to list:
// completion advances a task only when an update marks a not-done task done, so marking a done
// task done again moves nothing. A rule saved on a done task leaves it done.
const reasonTaskDone = "task_done"

// OccurrenceAnswer is the next due dates a recurring task would take, at most the requested
// count, and why it stopped early.
type OccurrenceAnswer struct {
	// Occurrences are the due dates successive completions would store, in order, each in the
	// server zone and to the second, as a stored due date is. Never nil.
	Occurrences []time.Time
	// Reason is "" when all requested occurrences were found; otherwise one of nextOccurrence's
	// reason codes, no_due_date or task_done.
	Reason string
}

// PreviewInput is an unsaved rule to evaluate, already validated by the caller with the
// same validator a task write uses (ValidateRepeat or ValidateRuleText). Exactly one of DueDate and
// InitialDueTime is set.
type PreviewInput struct {
	Rule string // stored-form RRULE text, as Task.Repeats holds it
	// DueDate is the anchor instant, as a newly saved task stores it; nil means absent.
	DueDate *time.Time
	// InitialDueTime is an HH:MM a newly saved task's first due date is selected from
	// (selectInitialDue), as a v2 write with initial_due_time selects it.
	InitialDueTime string
	FromCompletion bool
	TimeZone       string // IANA name; "" means the project's zone, else the caller's, then the server's
	// ProjectID is a project the caller has already been authorized to read; its zone
	// (resolveTaskZone) is used when TimeZone is empty. 0 means none.
	ProjectID int64
	Count     int
}

// Due-date sources a preview or a v2 write reports (due_date_source): the due date was sent, or
// the server selected it.
const (
	DueDateSourceExplicit  = "explicit"
	DueDateSourceGenerated = "generated"
)

// PreviewAnswer is a preview's occurrences and the context they were computed in. Occurrences are
// strictly after DueDate; Reason may sit beside a valid DueDate when the list ended before the
// count (count_exhausted, for example).
type PreviewAnswer struct {
	OccurrenceAnswer
	// DueDate is the anchor used (explicit) or selected (generated), in the server zone to the
	// second; nil when there is none.
	DueDate *time.Time
	// DueDateSource is DueDateSourceExplicit or DueDateSourceGenerated; "" without a DueDate.
	DueDateSource string
	// TimeZone is the IANA name of the zone the preview was evaluated in.
	TimeZone string
	// EvaluatedAt is the operation's now, in UTC, to the second.
	EvaluatedAt time.Time
}

// TaskOccurrences is the next n due dates of task, as n successive completions would store them:
// the same engine (nextOccurrence), input (completionInput), zone (resolveTaskZone) and COUNT
// decrement (decrementedRule) completion uses. For a task that repeats from the current date the
// list is a conditional on-time-completion forecast: the first completion is simulated at
// the later of now and the due date (forecastFirstCompletion), each later one on the previous
// due date. It is not a promise about actual completion, which schedules from its real time.
// Other tasks are completed at now. task must be the stored task with its reminders; it is never
// written to, nor are its reminders.
//
// Known limit (no mechanism): relative reminders are not stepped, so any recomputed
// relative reminder outside the storable range (years 1000-9999) makes completion refuse while
// the query lists the date.
func TaskOccurrences(s *xorm.Session, a web.Auth, task *Task, n int, now time.Time) (*OccurrenceAnswer, error) {
	if err := validateOccurrencesCount(n); err != nil {
		return nil, err
	}
	if reason := occurrencesRefusal(task); reason != "" {
		return &OccurrenceAnswer{Occurrences: []time.Time{}, Reason: reason}, nil
	}
	zone := resolveTaskZone(s, a, task.ProjectID)
	return stepOccurrences(completionInput(task, absoluteReminders(task.Reminders), forecastFirstCompletion(task, now), zone), n), nil
}

// PreviewOccurrences is TaskOccurrences for an unsaved rule, evaluated as a newly saved task with
// that rule and due date would be: no reminders, start or end. With InitialDueTime, the due date
// is the one a write would select (ParseInitialDueTime, then selectInitialDue at now), and no date
// comes back with selectInitialDue's reason. now is the operation's one sample (EngineNow).
//
// The zone is in.TimeZone when set (an unusable name is ErrInvalidTimezone), else the zone of
// in.ProjectID as a write to it resolves it (resolveTaskZone), else the caller's stored zone, else
// the server's. A rule that repeats from the current date gives the same conditional
// on-time-completion forecast as TaskOccurrences: not a promise about actual completion.
func PreviewOccurrences(s *xorm.Session, a web.Auth, in PreviewInput, now time.Time) (*PreviewAnswer, error) {
	if err := validateOccurrencesCount(in.Count); err != nil {
		return nil, err
	}
	// Defence in depth: the route refuses both with the same rule and message.
	if err := InitialDueConflict(in.DueDate != nil, in.InitialDueTime, in.Rule != "", false); err != nil {
		var conflict ErrInitialDueConflict
		if errors.As(err, &conflict) {
			return nil, conflict.PreviewError()
		}
		return nil, err
	}
	zone, err := previewContextZone(s, a, in)
	if err != nil {
		return nil, err
	}
	now = now.Truncate(time.Second)
	answer := &PreviewAnswer{TimeZone: zone.String(), EvaluatedAt: now.UTC()}

	task := &Task{Repeats: in.Rule, RepeatsFromCurrentDate: in.FromCompletion}
	source := DueDateSourceExplicit
	if in.InitialDueTime != "" {
		hour, minute, err := ParseInitialDueTime(in.InitialDueTime)
		if err != nil {
			return nil, err
		}
		due, reason := selectInitialDue(in.Rule, hour, minute, zone, now)
		if reason != "" {
			answer.OccurrenceAnswer = OccurrenceAnswer{Occurrences: []time.Time{}, Reason: string(reason)}
			return answer, nil
		}
		task.DueDate, source = due, DueDateSourceGenerated
	} else if in.DueDate != nil {
		// A saved due date is whole seconds (DATETIME).
		task.DueDate = in.DueDate.Truncate(time.Second)
	}
	if !task.DueDate.IsZero() {
		due := task.DueDate.In(config.GetTimeZone())
		answer.DueDate, answer.DueDateSource = &due, source
	}

	if reason := occurrencesRefusal(task); reason != "" {
		answer.OccurrenceAnswer = OccurrenceAnswer{Occurrences: []time.Time{}, Reason: reason}
		return answer, nil
	}
	answer.OccurrenceAnswer = *stepOccurrences(completionInput(task, nil, forecastFirstCompletion(task, now), zone), in.Count)
	return answer, nil
}

// previewContextZone is the zone a preview is evaluated in: the given name, else the authorized
// project's zone as a write resolves it, else previewZone's caller-then-server order.
func previewContextZone(s *xorm.Session, a web.Auth, in PreviewInput) (*time.Location, error) {
	if in.TimeZone == "" && in.ProjectID > 0 {
		return resolveTaskZone(s, a, in.ProjectID), nil
	}
	return previewZone(s, a, in.TimeZone)
}

// forecastFirstCompletion is the moment the first completion of a listed task is simulated at,
// from the operation's now (sampled once by the caller; truncated here to the second). A task that
// repeats from the current date is completed on time: at the
// later of now and its due date, so with a future due date no listed date precedes it, and with an
// overdue one it is completed at now. Any other task is completed at now; its rule does
// not read the completion time. Dates are never sorted, clamped or forced later: the engine alone
// decides each step, so COUNT, UNTIL and the gap skip keep their meaning.
func forecastFirstCompletion(task *Task, now time.Time) time.Time {
	now = now.Truncate(time.Second)
	if due := task.DueDate.Truncate(time.Second); task.RepeatsFromCurrentDate && due.After(now) {
		return due
	}
	return now
}

// validateOccurrencesCount refuses a count outside 1 to MaxOccurrencesCount, so the functions are
// safe on their own; the API rejects such a count before calling them.
func validateOccurrencesCount(n int) error {
	if n < 1 || n > MaxOccurrencesCount {
		return ErrInvalidData{Message: fmt.Sprintf("the occurrence count must be between 1 and %d", MaxOccurrencesCount)}
	}
	return nil
}

// occurrencesRefusal is why task has no due dates to list before the engine runs, or "".
func occurrencesRefusal(task *Task) string {
	if task.Repeats == "" {
		return string(reasonNoRule)
	}
	if task.DueDate.IsZero() {
		return reasonNoDueDate
	}
	if task.Done {
		return reasonTaskDone
	}
	return ""
}

// previewZone is the zone a preview is evaluated in.
func previewZone(s *xorm.Session, a web.Auth, name string) (*time.Location, error) {
	if name != "" {
		loc, ok := loadZone(name)
		if !ok {
			_, loadErr := time.LoadLocation(name)
			if loadErr == nil {
				// "Local" loads, but the process zone never decides a task's dates (loadZone).
				loadErr = errors.New("the process zone is not a time zone name")
			}
			return nil, ErrInvalidTimezone{Name: name, LoadError: loadErr}
		}
		return loc, nil
	}
	if loc, ok := doerZone(s, a, 0); ok {
		return loc, nil
	}
	return config.GetTimeZone(), nil
}

// stepOccurrences is the one occurrence loop: nextOccurrence up to n times, carrying between
// steps exactly what completion persists (nextOccurrenceStep). It stops at the first reason.
func stepOccurrences(in nextOccurrenceInput, n int) *OccurrenceAnswer {
	answer := &OccurrenceAnswer{Occurrences: make([]time.Time, 0, n)}
	server := config.GetTimeZone()
	for range n {
		res := nextOccurrence(in)
		if res.Reason != "" {
			answer.Reason = string(res.Reason)
			break
		}
		in = nextOccurrenceStep(in, res)
		// Converted to the server zone, as applyNextOccurrence stores it; the instant is unchanged.
		answer.Occurrences = append(answer.Occurrences, in.Due.At.In(server))
	}
	return answer
}

// nextOccurrenceStep is the input of the completion after the one that produced res, from in:
// what setTaskDatesRRule and the save persist, read back. Every date is truncated to the second
// (the DATETIME columns); the rule takes the COUNT decrement completion writes after a successful
// step; Now becomes the due date for a task that repeats from the current date (completed on its
// due date), and is otherwise unchanged.
func nextOccurrenceStep(in nextOccurrenceInput, res nextOccurrenceResult) nextOccurrenceInput {
	due := res.Due.At.Truncate(time.Second)
	in.Due = datedValue{At: due}
	in.Start = datedValue{At: res.Start.At.Truncate(time.Second)}
	in.End = datedValue{At: res.End.At.Truncate(time.Second)}
	reminders := make([]datedValue, len(res.Reminders))
	for i, rem := range res.Reminders {
		reminders[i] = datedValue{At: rem.At.Truncate(time.Second)}
	}
	in.Reminders = reminders
	if next, ok := decrementedRule(in.Rule); ok {
		in.Rule = next
	}
	if in.FromCompletion {
		in.Now = due
	}
	return in
}
