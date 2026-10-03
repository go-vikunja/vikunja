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
	"net/url"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/models"
)

// dateOnlyFormat is the RFC 5545 DATE value format used for all-day events.
const dateOnlyFormat = `20060102`

// Event holds a single VEVENT.
type Event struct {
	UID         string
	URL         string
	Summary     string
	Description string
	Categories  []string

	// Start is required. End is optional; without it a timed event is a point
	// in time and an all-day event lasts one day.
	Start  time.Time
	End    time.Time
	AllDay bool

	Created time.Time
	Updated time.Time
}

// ParseEvents returns a vcalendar string with one VEVENT per event.
func ParseEvents(config *Config, events []*Event) string {
	var b strings.Builder
	b.WriteString(calendarHeader(config))

	for _, e := range events {
		b.WriteString(`
BEGIN:VEVENT
UID:` + escapeICalText(e.UID) + `
DTSTAMP:` + makeCalDavTimeFromTimeStamp(e.Updated) + `
SUMMARY:` + escapeICalText(e.Summary))

		if e.AllDay {
			end := e.End
			if !end.After(e.Start) {
				end = e.Start.AddDate(0, 0, 1)
			}
			b.WriteString(`
DTSTART;VALUE=DATE:` + e.Start.Format(dateOnlyFormat) + `
DTEND;VALUE=DATE:` + end.Format(dateOnlyFormat))
		} else {
			b.WriteString(`
DTSTART:` + makeCalDavTimeFromTimeStamp(e.Start))
			if e.End.After(e.Start) {
				b.WriteString(`
DTEND:` + makeCalDavTimeFromTimeStamp(e.End))
			}
		}

		if e.Description != "" {
			description := descriptionToPlainText(e.UID, e.Description)
			if description != "" {
				b.WriteString(`
DESCRIPTION:` + escapeICalText(description))
			}
		}

		if e.URL != "" {
			b.WriteString(`
URL:` + e.URL)
		}

		if len(e.Categories) > 0 {
			escapedCategories := make([]string, len(e.Categories))
			for i, c := range e.Categories {
				escapedCategories[i] = escapeICalText(c)
			}
			b.WriteString(`
CATEGORIES:` + strings.Join(escapedCategories, ","))
		}

		if e.Created.Unix() > 0 {
			b.WriteString(`
CREATED:` + makeCalDavTimeFromTimeStamp(e.Created))
		}

		// Tasks are not appointments, so they must not mark the subscriber as busy.
		b.WriteString(`
TRANSP:TRANSPARENT
LAST-MODIFIED:` + makeCalDavTimeFromTimeStamp(e.Updated) + `
END:VEVENT`)
	}

	b.WriteString(`
END:VCALENDAR
`)

	// RFC 5545 §3.1 requires CRLF. Escaped TEXT values carry `\n` as two
	// characters, so only the line breaks written above are affected.
	return strings.ReplaceAll(b.String(), "\n", "\r\n")
}

// GetEventsForTasks maps tasks to calendar events. A task becomes an event
// spanning its start and end date when it has both, otherwise a single event
// on its due date, or else on its start date. Tasks without any of these dates
// are skipped. Dates at the start or end of a day in loc become all-day events,
// since the frontend's date-only pickers store them that way.
func GetEventsForTasks(tasks []*models.Task, loc *time.Location, publicURL string) []*Event {
	host := "vikunja"
	if u, err := url.Parse(publicURL); err == nil && u.Hostname() != "" {
		host = u.Hostname()
	}
	publicURL = strings.TrimSuffix(publicURL, "/")

	events := make([]*Event, 0, len(tasks))
	for _, t := range tasks {
		start, end, allDay, ok := eventTimesForTask(t, loc)
		if !ok {
			continue
		}

		var categories []string
		for _, label := range t.Labels {
			categories = append(categories, label.Title)
		}

		taskID := strconv.FormatInt(t.ID, 10)
		e := &Event{
			// Not the task's CalDAV UID: a client subscribed to both would
			// merge the VTODO and this VEVENT into one object.
			UID:         "vikunja-task-" + taskID + "@" + host,
			Summary:     t.Title,
			Description: t.Description,
			Categories:  categories,
			Start:       start,
			End:         end,
			AllDay:      allDay,
			Created:     t.Created,
			Updated:     t.Updated,
		}
		if publicURL != "" {
			e.URL = publicURL + "/tasks/" + taskID
		}
		events = append(events, e)
	}

	return events
}

func eventTimesForTask(t *models.Task, loc *time.Location) (start, end time.Time, allDay, ok bool) {
	switch {
	case hasDate(t.StartDate) && hasDate(t.EndDate) && t.EndDate.After(t.StartDate):
		start = t.StartDate.In(loc)
		end = t.EndDate.In(loc)
		if isStartOfDay(start) && (isStartOfDay(end) || isEndOfDay(end)) {
			startDay := truncateToDay(start)
			endDay := truncateToDay(end)
			if isEndOfDay(end) {
				endDay = endDay.AddDate(0, 0, 1)
			}
			return startDay, endDay, true, true
		}
		return start, end, false, true
	case hasDate(t.DueDate):
		start = t.DueDate.In(loc)
	case hasDate(t.StartDate):
		start = t.StartDate.In(loc)
	default:
		return time.Time{}, time.Time{}, false, false
	}

	if isStartOfDay(start) || isEndOfDay(start) {
		return truncateToDay(start), time.Time{}, true, true
	}
	return start, time.Time{}, false, true
}

// Same unset-date check as ParseTodos.
func hasDate(t time.Time) bool {
	return t.Unix() > 0
}

func isStartOfDay(t time.Time) bool {
	return t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0
}

func isEndOfDay(t time.Time) bool {
	return t.Hour() == 23 && t.Minute() == 59 && t.Second() == 59
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
