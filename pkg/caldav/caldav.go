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
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/richtext"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
)

// DateFormat is the caldav date format
const DateFormat = `20060102T150405`

// escapeICalText escapes a string for use inside an RFC 5545 §3.3.11 TEXT
// property value. Backslash MUST be replaced first so the backslashes this
// function inserts are not themselves re-escaped.
func escapeICalText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, `;`, `\;`)
	s = strings.ReplaceAll(s, `,`, `\,`)
	return s
}

// Todo holds a single VTODO
type Todo struct {
	// Required
	Timestamp time.Time
	UID       string

	// Optional
	Summary     string
	Description string
	Done        bool
	Completed   time.Time
	Organizer   *user.User
	Priority    int64 // 0-9, 1 is highest
	Relations   []Relation
	Color       string
	Categories  []string
	Start       time.Time
	End         time.Time
	DueDate     time.Time
	Duration    time.Duration
	// Repeats is an RFC 5545 RRULE string defining the recurrence pattern.
	Repeats string
	Alarms  []Alarm

	Created time.Time
	Updated time.Time // last-mod
}

// Alarm holds infos about an alarm from a caldav event
type Alarm struct {
	Time        time.Time
	Duration    time.Duration
	RelativeTo  models.ReminderRelation
	Description string
}

type Relation struct {
	Type models.RelationKind
	UID  string
}

// Config is the caldav calendar config
type Config struct {
	Name   string
	ProdID string
	Color  string
}

// getCaldavColor sanitizes color at the output boundary (non-hex chars
// dropped) so upstream HexColor validation gaps cannot lead to iCal
// property injection via CR/LF in a crafted color string.
func getCaldavColor(color string) (caldavcolor string) {
	color = strings.TrimPrefix(color, "#")

	var b strings.Builder
	b.Grow(len(color))
	for _, r := range color {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			b.WriteRune(r)
		}
	}
	color = b.String()
	if color == "" {
		return ""
	}

	color = "#" + color + "FF"

	return `
X-APPLE-CALENDAR-COLOR:` + color + `
X-OUTLOOK-COLOR:` + color + `
X-FUNAMBOL-COLOR:` + color + `
COLOR:` + color
}

func formatDuration(duration time.Duration) string {
	seconds := int64(duration.Seconds())
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	seconds %= 60

	return strconv.FormatInt(hours, 10) + `H` +
		strconv.FormatInt(minutes, 10) + `M` +
		strconv.FormatInt(seconds, 10) + `S`
}

// ParseTodos returns a caldav vcalendar string with todos.
func ParseTodos(config *Config, todos []*Todo) string {
	var b strings.Builder
	b.WriteString(`BEGIN:VCALENDAR
VERSION:2.0
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:` + escapeICalText(config.Name) + `
PRODID:-//` + config.ProdID + `//EN` + getCaldavColor(config.Color))

	for _, t := range todos {
		if t.UID == "" {
			t.UID = makeCalDavTimeFromTimeStamp(t.Timestamp) + utils.Sha256(t.Summary)
		}

		b.WriteString(`
BEGIN:VTODO
UID:` + escapeICalText(t.UID) + `
DTSTAMP:` + makeCalDavTimeFromTimeStamp(t.Timestamp) + `
SUMMARY:` + escapeICalText(t.Summary) + getCaldavColor(t.Color))

		if t.Start.Unix() > 0 {
			b.WriteString(`
DTSTART:` + makeCalDavTimeFromTimeStamp(t.Start))
			if t.Duration != 0 && t.DueDate.Unix() == 0 {
				b.WriteString(`
DURATION:PT` + formatDuration(t.Duration))
			}
		}
		if t.End.Unix() > 0 {
			b.WriteString(`
DTEND:` + makeCalDavTimeFromTimeStamp(t.End))
		}
		if t.Description != "" {
			// CalDAV clients show plain text, so emit markdown. On the near-impossible
			// conversion error, log it and keep the stored value (GetContent can't
			// return an error) rather than drop the description.
			description, err := richtext.HTMLToMarkdown(t.Description)
			if err != nil {
				log.Errorf("[CALDAV] Failed to convert description to markdown for task %q: %v", t.UID, err)
				description = t.Description
			}
			if description != "" {
				b.WriteString(`
DESCRIPTION:` + escapeICalText(description))
			}
		}
		// Keyed on Done, not Completed: a reopened repeating task keeps its done_at, and reporting
		// it as completed makes clients echo it back as done, rescheduling it on every sync.
		if t.Done {
			if t.Completed.Unix() > 0 {
				b.WriteString(`
COMPLETED:` + makeCalDavTimeFromTimeStamp(t.Completed))
			}
			b.WriteString(`
STATUS:COMPLETED`)
		} else {
			b.WriteString(`
STATUS:NEEDS-ACTION`)
		}
		if t.Organizer != nil {
			b.WriteString(`
ORGANIZER;CN=:` + escapeICalText(t.Organizer.Username))
		}

		if t.DueDate.Unix() > 0 {
			b.WriteString(`
DUE:` + makeCalDavTimeFromTimeStamp(t.DueDate))
		}

		if t.Created.Unix() > 0 {
			b.WriteString(`
CREATED:` + makeCalDavTimeFromTimeStamp(t.Created))
		}

		if t.Priority != 0 {
			b.WriteString(`
PRIORITY:` + strconv.Itoa(mapPriorityToCaldav(t.Priority)))
		}

		if t.Repeats != "" {
			// Normalize before emitting so we never write a malformed RRULE line
			// into the calendar output (e.g. from imported or legacy-migrated data).
			if normalized, ok := models.NormalizeRRule(t.Repeats); ok {
				b.WriteString(`
RRULE:` + normalized)
			} else {
				log.Debugf("[CalDAV] Ignoring unparseable RRULE %q on exported task %q", t.Repeats, t.UID)
			}
		}

		if len(t.Categories) > 0 {
			escapedCategories := make([]string, len(t.Categories))
			for i, c := range t.Categories {
				escapedCategories[i] = escapeICalText(c)
			}
			b.WriteString(`
CATEGORIES:` + strings.Join(escapedCategories, ","))
		}

		b.WriteString(`
LAST-MODIFIED:` + makeCalDavTimeFromTimeStamp(t.Updated))
		writeAlarms(&b, t.Alarms, t.Summary)
		writeRelations(&b, t.Relations)
		b.WriteString(`
END:VTODO`)
	}

	b.WriteString(`
END:VCALENDAR`) // Need a line break

	return b.String()
}

func writeAlarms(b *strings.Builder, alarms []Alarm, taskDescription string) {
	escapedTaskDescription := escapeICalText(taskDescription)
	for _, a := range alarms {
		description := escapedTaskDescription
		if a.Description != "" {
			description = escapeICalText(a.Description)
		}

		b.WriteString(`
BEGIN:VALARM`)
		switch a.RelativeTo {
		case models.ReminderRelationStartDate:
			b.WriteString(`
TRIGGER;RELATED=START:` + makeCalDavDuration(a.Duration))
		case models.ReminderRelationEndDate, models.ReminderRelationDueDate:
			b.WriteString(`
TRIGGER;RELATED=END:` + makeCalDavDuration(a.Duration))
		default:
			b.WriteString(`
TRIGGER;VALUE=DATE-TIME:` + makeCalDavTimeFromTimeStamp(a.Time))
		}
		b.WriteString(`
ACTION:DISPLAY
DESCRIPTION:`)
		b.WriteString(description)
		b.WriteString(`
END:VALARM`)
	}
}

func writeRelations(b *strings.Builder, relations []Relation) {
	for _, r := range relations {
		switch r.Type {
		case models.RelationKindParenttask:
			b.WriteString(`
RELATED-TO;RELTYPE=PARENT:`)
		case models.RelationKindSubtask:
			b.WriteString(`
RELATED-TO;RELTYPE=CHILD:`)
		case models.RelationKindUnknown,
			models.RelationKindRelated,
			models.RelationKindDuplicateOf,
			models.RelationKindDuplicates,
			models.RelationKindBlocking,
			models.RelationKindBlocked,
			models.RelationKindPreceeds,
			models.RelationKindFollows,
			models.RelationKindCopiedFrom,
			models.RelationKindCopiedTo:
			continue
		default:
			b.WriteString(`
RELATED-TO:`)
		}

		b.WriteString(escapeICalText(r.UID))
	}
}

func makeCalDavTimeFromTimeStamp(ts time.Time) (caldavtime string) {
	return ts.In(time.UTC).Format(DateFormat) + "Z"
}

func makeCalDavDuration(duration time.Duration) (caldavtime string) {
	if duration < 0 {
		duration = duration.Abs()
		caldavtime = "-"
	}
	caldavtime += "PT" + strings.ToUpper(duration.Truncate(time.Millisecond).String())
	return
}
