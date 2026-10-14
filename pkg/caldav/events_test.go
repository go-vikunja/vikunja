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
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/models"

	ics "github.com/arran4/golang-ical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEvents(t *testing.T) {
	config := &Config{
		Name:   "Project; with, specials",
		ProdID: "RandomProdID which is not random",
	}
	updated := time.Date(2018, 12, 1, 1, 12, 4, 0, time.UTC)

	tests := []struct {
		name  string
		event *Event
		want  string
	}{
		{
			name: "timed point in time with description and url",
			event: &Event{
				UID:         "vikunja-task-1@example.com",
				URL:         "https://example.com/tasks/1",
				Summary:     "Call, the; plumber",
				Description: `<p>Lorem Ipsum</p><p>Dolor sit amet</p>`,
				Start:       time.Date(2018, 12, 1, 3, 58, 44, 0, time.UTC),
				Created:     time.Date(2018, 11, 30, 10, 0, 0, 0, time.UTC),
				Updated:     updated,
			},
			want: `BEGIN:VEVENT
UID:vikunja-task-1@example.com
DTSTAMP:20181201T011204Z
SUMMARY:Call\, the\; plumber
DTSTART:20181201T035844Z
DESCRIPTION:Lorem Ipsum\n\nDolor sit amet
URL:https://example.com/tasks/1
CREATED:20181130T100000Z
TRANSP:TRANSPARENT
LAST-MODIFIED:20181201T011204Z
END:VEVENT`,
		},
		{
			name: "timed span with categories",
			event: &Event{
				UID:        "vikunja-task-2@example.com",
				Summary:    "Workshop",
				Categories: []string{"work", "a,b"},
				Start:      time.Date(2018, 12, 12, 7, 33, 20, 0, time.UTC),
				End:        time.Date(2018, 12, 13, 11, 20, 0, 0, time.UTC),
				Updated:    updated,
			},
			want: `BEGIN:VEVENT
UID:vikunja-task-2@example.com
DTSTAMP:20181201T011204Z
SUMMARY:Workshop
DTSTART:20181212T073320Z
DTEND:20181213T112000Z
CATEGORIES:work,a\,b
TRANSP:TRANSPARENT
LAST-MODIFIED:20181201T011204Z
END:VEVENT`,
		},
		{
			name: "all-day without end lasts one day",
			event: &Event{
				UID:     "vikunja-task-3@example.com",
				Summary: "Birthday",
				Start:   time.Date(2018, 12, 31, 0, 0, 0, 0, time.UTC),
				AllDay:  true,
				Updated: updated,
			},
			want: `BEGIN:VEVENT
UID:vikunja-task-3@example.com
DTSTAMP:20181201T011204Z
SUMMARY:Birthday
DTSTART;VALUE=DATE:20181231
DTEND;VALUE=DATE:20190101
TRANSP:TRANSPARENT
LAST-MODIFIED:20181201T011204Z
END:VEVENT`,
		},
		{
			name: "all-day span keeps its exclusive end date",
			event: &Event{
				UID:     "vikunja-task-4@example.com",
				Summary: "Vacation",
				Start:   time.Date(2018, 12, 24, 0, 0, 0, 0, time.UTC),
				End:     time.Date(2018, 12, 27, 0, 0, 0, 0, time.UTC),
				AllDay:  true,
				Updated: updated,
			},
			want: `BEGIN:VEVENT
UID:vikunja-task-4@example.com
DTSTAMP:20181201T011204Z
SUMMARY:Vacation
DTSTART;VALUE=DATE:20181224
DTEND;VALUE=DATE:20181227
TRANSP:TRANSPARENT
LAST-MODIFIED:20181201T011204Z
END:VEVENT`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseEvents(config, []*Event{tt.event})
			want := `BEGIN:VCALENDAR
VERSION:2.0
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:Project\; with\, specials
PRODID:-//RandomProdID which is not random//EN
` + tt.want + `
END:VCALENDAR
`
			assert.Equal(t, strings.ReplaceAll(want, "\n", "\r\n"), got)

			parsed, err := ics.ParseCalendar(strings.NewReader(got))
			require.NoError(t, err)
			require.Len(t, parsed.Events(), 1)
			assert.Equal(t, tt.event.UID, parsed.Events()[0].Id())
		})
	}

	t.Run("empty calendar", func(t *testing.T) {
		got := ParseEvents(config, nil)
		assert.True(t, strings.HasPrefix(got, "BEGIN:VCALENDAR\r\n"))
		assert.True(t, strings.HasSuffix(got, "\r\nEND:VCALENDAR\r\n"))
		assert.NotContains(t, got, "BEGIN:VEVENT")
	})

	t.Run("newlines in text cannot inject properties", func(t *testing.T) {
		got := ParseEvents(config, []*Event{
			{
				UID:     "vikunja-task-5@example.com",
				Summary: "evil\r\nATTENDEE:mailto:x@example.com",
				Start:   time.Date(2018, 12, 1, 3, 58, 44, 0, time.UTC),
				Updated: updated,
			},
		})
		assert.Contains(t, got, "SUMMARY:evil\\nATTENDEE:mailto:x@example.com\r\n")
		assert.NotContains(t, got, "\r\nATTENDEE:")
	})
}

func TestGetEventsForTasks(t *testing.T) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	require.NoError(t, err)
	updated := time.Date(2018, 12, 1, 1, 12, 4, 0, time.UTC)

	tasks := []*models.Task{
		{
			ID:      1,
			Title:   "no dates",
			Updated: updated,
		},
		{
			ID:      2,
			Title:   "timed due date",
			DueDate: time.Date(2018, 12, 1, 3, 58, 44, 0, time.UTC),
			Labels: []*models.Label{
				{Title: "home"},
			},
			Updated: updated,
		},
		{
			ID:      3,
			Title:   "due at midnight in the user's time zone",
			DueDate: time.Date(2018, 12, 4, 23, 0, 0, 0, time.UTC),
			Updated: updated,
		},
		{
			ID:        4,
			Title:     "start and end win over the due date",
			DueDate:   time.Date(2018, 12, 20, 12, 0, 0, 0, time.UTC),
			StartDate: time.Date(2018, 12, 12, 7, 33, 20, 0, time.UTC),
			EndDate:   time.Date(2018, 12, 13, 11, 20, 0, 0, time.UTC),
			Updated:   updated,
		},
		{
			ID:        5,
			Title:     "day span from the gantt chart",
			StartDate: time.Date(2018, 12, 9, 23, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2018, 12, 12, 22, 59, 59, 0, time.UTC),
			Updated:   updated,
		},
		{
			ID:        6,
			Title:     "start date only",
			StartDate: time.Date(2018, 12, 12, 7, 33, 20, 0, time.UTC),
			Updated:   updated,
		},
		{
			ID:      7,
			Title:   "end date only",
			EndDate: time.Date(2018, 12, 13, 11, 20, 0, 0, time.UTC),
			Updated: updated,
		},
		{
			ID:        8,
			Title:     "end before start falls back to the due date",
			DueDate:   time.Date(2018, 12, 20, 12, 0, 0, 0, time.UTC),
			StartDate: time.Date(2018, 12, 13, 11, 20, 0, 0, time.UTC),
			EndDate:   time.Date(2018, 12, 12, 7, 33, 20, 0, time.UTC),
			Updated:   updated,
		},
	}

	events := GetEventsForTasks(tasks, berlin, "https://vikunja.example.com/")
	require.Len(t, events, 6)

	byUID := map[string]*Event{}
	for _, e := range events {
		byUID[e.UID] = e
	}
	get := func(t *testing.T, id string) *Event {
		t.Helper()
		e, ok := byUID["vikunja-task-"+id+"@vikunja.example.com"]
		require.True(t, ok, "no event for task %s", id)
		return e
	}

	t.Run("tasks without a usable date are skipped", func(t *testing.T) {
		_, hasNoDates := byUID["vikunja-task-1@vikunja.example.com"]
		assert.False(t, hasNoDates)
		_, hasEndOnly := byUID["vikunja-task-7@vikunja.example.com"]
		assert.False(t, hasEndOnly)
	})

	t.Run("timed due date", func(t *testing.T) {
		e := get(t, "2")
		assert.Equal(t, "timed due date", e.Summary)
		assert.Equal(t, "https://vikunja.example.com/tasks/2", e.URL)
		assert.Equal(t, []string{"home"}, e.Categories)
		assert.False(t, e.AllDay)
		assert.True(t, e.Start.Equal(tasks[1].DueDate))
		assert.True(t, e.End.IsZero())
		assert.Equal(t, updated, e.Updated)
	})

	t.Run("midnight due date becomes all-day", func(t *testing.T) {
		e := get(t, "3")
		assert.True(t, e.AllDay)
		assert.Equal(t, "20181205", e.Start.Format(dateOnlyFormat))
		assert.True(t, e.End.IsZero())
	})

	t.Run("start and end span", func(t *testing.T) {
		e := get(t, "4")
		assert.False(t, e.AllDay)
		assert.True(t, e.Start.Equal(tasks[3].StartDate))
		assert.True(t, e.End.Equal(tasks[3].EndDate))
	})

	t.Run("start of day to end of day span becomes all-day", func(t *testing.T) {
		e := get(t, "5")
		assert.True(t, e.AllDay)
		assert.Equal(t, "20181210", e.Start.Format(dateOnlyFormat))
		assert.Equal(t, "20181213", e.End.Format(dateOnlyFormat))
	})

	t.Run("start date only", func(t *testing.T) {
		e := get(t, "6")
		assert.False(t, e.AllDay)
		assert.True(t, e.Start.Equal(tasks[5].StartDate))
	})

	t.Run("invalid span falls back to the due date", func(t *testing.T) {
		e := get(t, "8")
		assert.True(t, e.Start.Equal(tasks[7].DueDate))
		assert.True(t, e.End.IsZero())
	})

	t.Run("without a public url", func(t *testing.T) {
		events := GetEventsForTasks(tasks[1:2], time.UTC, "")
		require.Len(t, events, 1)
		assert.Equal(t, "vikunja-task-2@vikunja", events[0].UID)
		assert.Empty(t, events[0].URL)
	})
}
