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

package caldav

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func day(d int) time.Time { return time.Date(2026, time.March, d, 12, 0, 0, 0, time.UTC) }

func TestTimeRangeOverlapsTodo(t *testing.T) {
	// Ranges relative to a task living on days 10 to 12.
	var (
		before = timeRange{
			start: day(1),
			end:   day(5),
		}
		after = timeRange{
			start: day(20),
			end:   day(25),
		}
		around = timeRange{
			start: day(9),
			end:   day(13),
		}
		inside = timeRange{
			start: day(11),
			end:   time.Date(2026, time.March, 11, 18, 0, 0, 0, time.UTC),
		}
		untilDay = func(d int) timeRange { return timeRange{end: day(d)} }
		fromDay  = func(d int) timeRange {
			return timeRange{
				start: day(d),
				end:   latest,
			}
		}
		exactly10 = timeRange{
			start: day(10),
			end:   day(10).Add(time.Second),
		}
	)

	tests := []struct {
		name string
		task models.Task
		in   []timeRange
		out  []timeRange
	}{
		{
			// DTSTART + DURATION: start <= DTSTART+DURATION and (end > DTSTART or end >= DTSTART+DURATION).
			name: "start and end",
			task: models.Task{
				StartDate: day(10),
				EndDate:   day(12),
			},
			in: []timeRange{
				around,
				inside,
				untilDay(11),
				fromDay(11),
				fromDay(12),
				exactly10,
			},
			out: []timeRange{
				before,
				after,
				untilDay(10),
				fromDay(13),
			},
		},
		{
			// DTSTART + DUE: (start < DUE or start <= DTSTART) and (end > DTSTART or end >= DUE).
			name: "start and due",
			task: models.Task{
				StartDate: day(10),
				DueDate:   day(12),
			},
			in: []timeRange{
				around,
				inside,
				untilDay(11),
				fromDay(11),
				exactly10,
			},
			out: []timeRange{
				before,
				after,
				untilDay(10),
				fromDay(12),
			},
		},
		{
			// The due date takes the place of the end date, like in the rendered VTODO.
			name: "start, due and end",
			task: models.Task{
				StartDate: day(10),
				DueDate:   day(12),
				EndDate:   day(20),
			},
			in: []timeRange{
				around,
				fromDay(11),
			},
			out: []timeRange{
				fromDay(12),
				after,
			},
		},
		{
			// DTSTART only: start <= DTSTART and end > DTSTART.
			name: "start only",
			task: models.Task{StartDate: day(10)},
			in: []timeRange{
				around,
				untilDay(11),
				fromDay(10),
				exactly10,
			},
			out: []timeRange{
				before,
				after,
				inside,
				untilDay(10),
				fromDay(11),
			},
		},
		{
			// An end before the start renders no usable DURATION, so it counts as start only.
			name: "end before start",
			task: models.Task{
				StartDate: day(10),
				EndDate:   day(5),
			},
			in: []timeRange{fromDay(10)},
			out: []timeRange{
				inside,
				fromDay(11),
			},
		},
		{
			// DUE only: start < DUE and end >= DUE.
			name: "due only",
			task: models.Task{DueDate: day(12)},
			in: []timeRange{
				around,
				untilDay(12),
				fromDay(11),
			},
			out: []timeRange{
				before,
				after,
				inside,
				untilDay(11),
				fromDay(12),
			},
		},
		{
			name: "repeating, due only",
			task: models.Task{
				DueDate:     day(12),
				RepeatAfter: 7 * 24 * 3600,
			},
			in: []timeRange{{
				start: day(18),
				end:   day(20),
			}},
			out: []timeRange{before},
		},
		{
			// An end without a start renders neither DTSTART nor DURATION.
			name: "end only",
			task: models.Task{
				EndDate: day(12),
				Created: day(10),
			},
			in: []timeRange{
				fromDay(1),
				untilDay(11),
			},
			out: []timeRange{
				untilDay(10),
				before,
			},
		},
		{
			// COMPLETED + CREATED: (start <= CREATED or start <= COMPLETED) and (end >= CREATED or end >= COMPLETED).
			name: "completed and created",
			task: models.Task{
				Done:    true,
				DoneAt:  day(12),
				Created: day(10),
			},
			in: []timeRange{
				around,
				inside,
				untilDay(10),
				fromDay(12),
			},
			out: []timeRange{
				before,
				after,
				untilDay(9),
				fromDay(13),
			},
		},
		{
			// COMPLETED only: start <= COMPLETED and end >= COMPLETED.
			name: "completed only",
			task: models.Task{
				Done:   true,
				DoneAt: day(12),
			},
			in: []timeRange{
				around,
				untilDay(12),
				fromDay(12),
			},
			out: []timeRange{
				before,
				after,
				inside,
				untilDay(11),
				fromDay(13),
			},
		},
		{
			// CREATED only: end > CREATED.
			name: "created only",
			task: models.Task{Created: day(10)},
			in: []timeRange{
				around,
				after,
				untilDay(11),
				fromDay(1),
			},
			out: []timeRange{
				before,
				untilDay(10),
			},
		},
		{
			// A reopened task keeps its done date, but renders no COMPLETED.
			name: "done date of an open task",
			task: models.Task{
				DoneAt:  day(12),
				Created: day(10),
			},
			in:  []timeRange{after},
			out: []timeRange{before},
		},
		{
			// Neither of them: always.
			name: "nothing",
			task: models.Task{},
			in: []timeRange{
				before,
				after,
				around,
				untilDay(1),
				fromDay(30),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, r := range tc.in {
				assert.True(t, r.overlapsTodo(&tc.task), "%s - %s must match", r.start, r.end)
			}
			for _, r := range tc.out {
				assert.False(t, r.overlapsTodo(&tc.task), "%s - %s must not match", r.start, r.end)
			}
		})
	}
}

func parseFilterBody(t *testing.T, filter string) (calendarFilter, error) {
	t.Helper()
	req, err := parseReport(`<C:calendar-query xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav" xmlns:X="urn:example">
  <D:prop><D:getetag/></D:prop>
  <C:filter>` + filter + `</C:filter>
</C:calendar-query>`)
	if err != nil {
		return calendarFilter{}, err
	}
	query, ok := req.(calendarQueryRequest)
	require.True(t, ok)
	return query.filter, nil
}

func todoFilter(inner string) string {
	return `<C:comp-filter name="VCALENDAR"><C:comp-filter name="VTODO">` + inner + `</C:comp-filter></C:comp-filter>`
}

func TestParseFilterRejects(t *testing.T) {
	tests := []struct {
		name         string
		filter       string
		precondition string
	}{
		{
			name:         "no comp-filter",
			filter:       ``,
			precondition: "valid-filter",
		},
		{
			name:         "two comp-filters",
			filter:       `<C:comp-filter name="VCALENDAR"/><C:comp-filter name="VCALENDAR"/>`,
			precondition: "valid-filter",
		},
		{
			name:         "outermost component is no calendar",
			filter:       `<C:comp-filter name="VTODO"/>`,
			precondition: "valid-filter",
		},
		{
			name:         "comp-filter without name",
			filter:       `<C:comp-filter name="VCALENDAR"><C:comp-filter/></C:comp-filter>`,
			precondition: "valid-filter",
		},
		{
			name:         "nested calendar",
			filter:       `<C:comp-filter name="VCALENDAR"><C:comp-filter name="VCALENDAR"/></C:comp-filter>`,
			precondition: "valid-filter",
		},
		{
			name:         "event in a todo",
			filter:       todoFilter(`<C:comp-filter name="VEVENT"/>`),
			precondition: "valid-filter",
		},
		{
			name:         "component in an alarm",
			filter:       todoFilter(`<C:comp-filter name="VALARM"><C:comp-filter name="X-FOO"/></C:comp-filter>`),
			precondition: "valid-filter",
		},
		{
			name:         "time-range on the calendar",
			filter:       `<C:comp-filter name="VCALENDAR"><C:time-range start="20260101T000000Z"/></C:comp-filter>`,
			precondition: "valid-filter",
		},
		{
			name:         "time-range on an alarm",
			filter:       todoFilter(`<C:comp-filter name="VALARM"><C:time-range start="20260101T000000Z"/></C:comp-filter>`),
			precondition: "supported-filter",
		},
		{
			name:         "two time-ranges",
			filter:       todoFilter(`<C:time-range start="20260101T000000Z"/><C:time-range end="20260101T000000Z"/>`),
			precondition: "valid-filter",
		},
		{
			name:         "time-range without bounds",
			filter:       todoFilter(`<C:time-range/>`),
			precondition: "valid-filter",
		},
		{
			name:         "time-range not in UTC",
			filter:       todoFilter(`<C:time-range start="20260101T000000"/>`),
			precondition: "valid-filter",
		},
		{
			name:         "time-range as a date",
			filter:       todoFilter(`<C:time-range start="20260101"/>`),
			precondition: "valid-filter",
		},
		{
			name:         "time-range ending before it starts",
			filter:       todoFilter(`<C:time-range start="20260102T000000Z" end="20260101T000000Z"/>`),
			precondition: "valid-filter",
		},
		{
			name:         "time-range on a text property",
			filter:       todoFilter(`<C:prop-filter name="SUMMARY"><C:time-range start="20260101T000000Z"/></C:prop-filter>`),
			precondition: "valid-filter",
		},
		{
			name:         "is-not-defined with tests",
			filter:       todoFilter(`<C:is-not-defined/><C:time-range start="20260101T000000Z"/>`),
			precondition: "valid-filter",
		},
		{
			name:         "prop is-not-defined with text-match",
			filter:       todoFilter(`<C:prop-filter name="SUMMARY"><C:is-not-defined/><C:text-match>x</C:text-match></C:prop-filter>`),
			precondition: "valid-filter",
		},
		{
			name:         "text-match and time-range",
			filter:       todoFilter(`<C:prop-filter name="DUE"><C:text-match>x</C:text-match><C:time-range start="20260101T000000Z"/></C:prop-filter>`),
			precondition: "valid-filter",
		},
		{
			name:         "prop-filter without name",
			filter:       todoFilter(`<C:prop-filter><C:is-not-defined/></C:prop-filter>`),
			precondition: "valid-filter",
		},
		{
			name:         "unknown collation",
			filter:       todoFilter(`<C:prop-filter name="SUMMARY"><C:text-match collation="i;unicode-casemap">x</C:text-match></C:prop-filter>`),
			precondition: "supported-collation",
		},
		{
			name:         "bad negate-condition",
			filter:       todoFilter(`<C:prop-filter name="SUMMARY"><C:text-match negate-condition="maybe">x</C:text-match></C:prop-filter>`),
			precondition: "valid-filter",
		},
		{
			name:         "two text-matches in a param-filter",
			filter:       todoFilter(`<C:prop-filter name="X"><C:param-filter name="Y"><C:text-match>a</C:text-match><C:text-match>b</C:text-match></C:param-filter></C:prop-filter>`),
			precondition: "valid-filter",
		},
		{
			name:         "unknown CalDAV element",
			filter:       todoFilter(`<C:whatever/>`),
			precondition: "valid-filter",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseFilterBody(t, tc.filter)
			var failed preconditionError
			require.ErrorAs(t, err, &failed)
			assert.Equal(t, calDAVName(tc.precondition), failed.precondition)
		})
	}

	t.Run("a second filter element", func(t *testing.T) {
		_, err := parseReport(`<C:calendar-query xmlns:C="urn:ietf:params:xml:ns:caldav"><C:filter>` + todoFilter("") + `</C:filter><C:filter/></C:calendar-query>`)
		var failed preconditionError
		require.ErrorAs(t, err, &failed)
	})

	t.Run("too many elements", func(t *testing.T) {
		props := ""
		for range maxFilterElements {
			props += `<C:prop-filter name="SUMMARY"/>`
		}
		_, err := parseFilterBody(t, todoFilter(props))
		require.ErrorIs(t, err, errBodyTooLarge)
	})
}

func TestFilterMatches(t *testing.T) {
	task := &models.Task{
		UID:       "uid-filter-test",
		Title:     "Buy Milk",
		DueDate:   day(12),
		Created:   day(1),
		Priority:  3,
		Labels:    []*models.Label{{Title: "groceries"}},
		Reminders: []*models.TaskReminder{{Reminder: day(11)}},
	}

	tests := []struct {
		name    string
		filter  string
		matches bool
	}{
		{
			name:    "calendar only",
			filter:  `<C:comp-filter name="VCALENDAR"/>`,
			matches: true,
		},
		{
			name:    "todo",
			filter:  todoFilter(``),
			matches: true,
		},
		{
			name:    "event",
			filter:  `<C:comp-filter name="VCALENDAR"><C:comp-filter name="VEVENT"/></C:comp-filter>`,
			matches: false,
		},
		{
			name:    "no event",
			filter:  `<C:comp-filter name="VCALENDAR"><C:comp-filter name="VEVENT"><C:is-not-defined/></C:comp-filter></C:comp-filter>`,
			matches: true,
		},
		{
			name:    "no todo",
			filter:  `<C:comp-filter name="VCALENDAR"><C:comp-filter name="VTODO"><C:is-not-defined/></C:comp-filter></C:comp-filter>`,
			matches: false,
		},
		{
			name:    "component names are case-insensitive",
			filter:  `<C:comp-filter name="vcalendar"><C:comp-filter name="vtodo"/></C:comp-filter>`,
			matches: true,
		},
		{
			name:    "calendar property",
			filter:  `<C:comp-filter name="VCALENDAR"><C:prop-filter name="VERSION"><C:text-match>2.0</C:text-match></C:prop-filter></C:comp-filter>`,
			matches: true,
		},
		{
			name:    "defined property",
			filter:  todoFilter(`<C:prop-filter name="DUE"/>`),
			matches: true,
		},
		{
			name:    "undefined property",
			filter:  todoFilter(`<C:prop-filter name="DTSTART"/>`),
			matches: false,
		},
		{
			name:    "is-not-defined of an undefined property",
			filter:  todoFilter(`<C:prop-filter name="COMPLETED"><C:is-not-defined/></C:prop-filter>`),
			matches: true,
		},
		{
			name:    "is-not-defined of a defined property",
			filter:  todoFilter(`<C:prop-filter name="SUMMARY"><C:is-not-defined/></C:prop-filter>`),
			matches: false,
		},
		{
			name:    "text-match folds ASCII case by default",
			filter:  todoFilter(`<C:prop-filter name="SUMMARY"><C:text-match>buy milk</C:text-match></C:prop-filter>`),
			matches: true,
		},
		{
			name:    "text-match octet keeps case",
			filter:  todoFilter(`<C:prop-filter name="SUMMARY"><C:text-match collation="i;octet">buy milk</C:text-match></C:prop-filter>`),
			matches: false,
		},
		{
			name:    "text-match octet",
			filter:  todoFilter(`<C:prop-filter name="SUMMARY"><C:text-match collation="i;octet">Milk</C:text-match></C:prop-filter>`),
			matches: true,
		},
		{
			name:    "negated text-match",
			filter:  todoFilter(`<C:prop-filter name="STATUS"><C:text-match negate-condition="yes">COMPLETED</C:text-match></C:prop-filter>`),
			matches: true,
		},
		{
			name:    "negated text-match that matches",
			filter:  todoFilter(`<C:prop-filter name="STATUS"><C:text-match negate-condition="yes">NEEDS-ACTION</C:text-match></C:prop-filter>`),
			matches: false,
		},
		{
			name:    "negated text-match of an undefined property",
			filter:  todoFilter(`<C:prop-filter name="DESCRIPTION"><C:text-match negate-condition="yes">x</C:text-match></C:prop-filter>`),
			matches: false,
		},
		{
			name:    "text-match on categories",
			filter:  todoFilter(`<C:prop-filter name="CATEGORIES"><C:text-match>grocer</C:text-match></C:prop-filter>`),
			matches: true,
		},
		{
			name:    "time-range on a property",
			filter:  todoFilter(`<C:prop-filter name="DUE"><C:time-range start="20260312T000000Z" end="20260313T000000Z"/></C:prop-filter>`),
			matches: true,
		},
		{
			name:    "time-range on a property, outside",
			filter:  todoFilter(`<C:prop-filter name="DUE"><C:time-range start="20260313T000000Z"/></C:prop-filter>`),
			matches: false,
		},
		{
			name:    "time-range on the todo",
			filter:  todoFilter(`<C:time-range start="20260311T000000Z" end="20260313T000000Z"/>`),
			matches: true,
		},
		{
			name:    "time-range on the todo, outside",
			filter:  todoFilter(`<C:time-range end="20260311T000000Z"/>`),
			matches: false,
		},
		{
			name:    "defined parameter",
			filter:  todoFilter(`<C:comp-filter name="VALARM"><C:prop-filter name="TRIGGER"><C:param-filter name="VALUE"/></C:prop-filter></C:comp-filter>`),
			matches: true,
		},
		{
			name:    "parameter text-match",
			filter:  todoFilter(`<C:comp-filter name="VALARM"><C:prop-filter name="TRIGGER"><C:param-filter name="VALUE"><C:text-match>date-time</C:text-match></C:param-filter></C:prop-filter></C:comp-filter>`),
			matches: true,
		},
		{
			name:    "undefined parameter",
			filter:  todoFilter(`<C:prop-filter name="SUMMARY"><C:param-filter name="LANGUAGE"><C:is-not-defined/></C:param-filter></C:prop-filter>`),
			matches: true,
		},
		{
			name:    "alarm",
			filter:  todoFilter(`<C:comp-filter name="VALARM"/>`),
			matches: true,
		},
		{
			name:    "all tests must match",
			filter:  todoFilter(`<C:prop-filter name="DUE"/><C:prop-filter name="DTSTART"/>`),
			matches: false,
		},
		{
			name:    "extension elements are ignored",
			filter:  todoFilter(`<X:anything><C:whatever/></X:anything>`),
			matches: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filter, err := parseFilterBody(t, tc.filter)
			require.NoError(t, err)
			candidate := &queryCandidate{resource: &taskResource{task: task}}
			assert.Equal(t, tc.matches, filter.matches(candidate), candidate.resource.content())
		})
	}

	t.Run("properties are only rendered when asked for", func(t *testing.T) {
		filter, err := parseFilterBody(t, todoFilter(`<C:time-range start="20260311T000000Z"/>`))
		require.NoError(t, err)
		candidate := &queryCandidate{resource: &taskResource{task: task}}
		assert.True(t, filter.matches(candidate))
		assert.Nil(t, candidate.resource.rendered)
	})
}
