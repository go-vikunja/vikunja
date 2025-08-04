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
	"regexp"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
)

// DateFormat is the caldav date format
const DateFormat = `20060102T150405`

// Todo holds a single VTODO
type Todo struct {
	// Required
	Timestamp time.Time
	UID       string

	// Optional
	Summary     string
	Description string
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
	RepeatAfter int64
	RepeatMode  models.TaskRepeatMode
	Alarms      []Alarm

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

func getCaldavColor(color string) (caldavcolor string) {
	if color == "" {
		return ""
	}

	if !strings.HasPrefix(color, "#") {
		color = "#" + color
	}

	color += "FF"

	return `
X-APPLE-CALENDAR-COLOR:` + color + `
X-OUTLOOK-COLOR:` + color + `
X-FUNAMBOL-COLOR:` + color + `
COLOR:` + color
}

func formatDuration(duration time.Duration) string {
	seconds := duration.Seconds() - duration.Minutes()*60
	minutes := duration.Minutes() - duration.Hours()*60

	return strconv.FormatFloat(duration.Hours(), 'f', 0, 64) + `H` +
		strconv.FormatFloat(minutes, 'f', 0, 64) + `M` +
		strconv.FormatFloat(seconds, 'f', 0, 64) + `S`
}

func getRruleFromInterval(interval int64) (freq string, newInterval int64) {
	const (
		minute = 60
		hour   = minute * 60
		day    = hour * 24
		week   = day * 7
	)

	switch {
	case interval%week == 0:
		return "WEEKLY", interval / week
	case interval%day == 0:
		return "DAILY", interval / day
	case interval%hour == 0:
		return "HOURLY", interval / hour
	case interval%minute == 0:
		return "MINUTELY", interval / minute
	default:
		return "SECONDLY", interval
	}
}

// ParseTodos returns a caldav vcalendar string with todos
func ParseTodos(config *Config, todos []*Todo) (caldavtodos string) {
	caldavtodos = `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:` + config.Name + `
PRODID:-//` + config.ProdID + `//EN` + getCaldavColor(config.Color)

	for _, t := range todos {
		if t.UID == "" {
			t.UID = makeCalDavTimeFromTimeStamp(t.Timestamp) + utils.Sha256(t.Summary)
		}

		caldavtodos += `
BEGIN:VTODO
UID:` + t.UID + `
DTSTAMP:` + makeCalDavTimeFromTimeStamp(t.Timestamp) + `
SUMMARY:` + t.Summary + getCaldavColor(t.Color)

		if t.Start.Unix() > 0 {
			caldavtodos += `
DTSTART:` + makeCalDavTimeFromTimeStamp(t.Start)
			if t.Duration != 0 && t.DueDate.Unix() == 0 {
				caldavtodos += `
DURATION:PT` + formatDuration(t.Duration)
			}
		}
		if t.End.Unix() > 0 {
			caldavtodos += `
DTEND:` + makeCalDavTimeFromTimeStamp(t.End)
		}
		if t.Description != "" {
			re := regexp.MustCompile(`\r?\n`)
			formattedDescription := re.ReplaceAllString(t.Description, "\\n")
			caldavtodos += `
DESCRIPTION:` + formattedDescription
		}
		if t.Completed.Unix() > 0 {
			caldavtodos += `
COMPLETED:` + makeCalDavTimeFromTimeStamp(t.Completed) + `
STATUS:COMPLETED`
		}
		if t.Organizer != nil {
			caldavtodos += `
ORGANIZER;CN=:` + t.Organizer.Username
		}

		if t.DueDate.Unix() > 0 {
			caldavtodos += `
DUE:` + makeCalDavTimeFromTimeStamp(t.DueDate)
		}

		if t.Created.Unix() > 0 {
			caldavtodos += `
CREATED:` + makeCalDavTimeFromTimeStamp(t.Created)
		}

		if t.Priority != 0 {
			caldavtodos += `
PRIORITY:` + strconv.Itoa(mapPriorityToCaldav(t.Priority))
		}

		if t.RepeatAfter > 0 || t.RepeatMode == models.TaskRepeatModeMonth {
			if t.RepeatMode == models.TaskRepeatModeMonth {
				caldavtodos += `
RRULE:FREQ=MONTHLY;BYMONTHDAY=` + t.DueDate.Format("02") // Day of the month
			} else {
				freq, interval := getRruleFromInterval(t.RepeatAfter)
				caldavtodos += `
RRULE:FREQ=` + freq + `;INTERVAL=` + strconv.FormatInt(interval, 10)
			}
		}

		if len(t.Categories) > 0 {
			caldavtodos += `
CATEGORIES:` + strings.Join(t.Categories, ",")
		}

		caldavtodos += `
LAST-MODIFIED:` + makeCalDavTimeFromTimeStamp(t.Updated)
		caldavtodos += ParseAlarms(t.Alarms, t.Summary)
		caldavtodos += ParseRelations(t.Relations)
		caldavtodos += `
END:VTODO`
	}

	caldavtodos += `
END:VCALENDAR` // Need a line break

	return
}

func ParseAlarms(alarms []Alarm, taskDescription string) (caldavalarms string) {
	for _, a := range alarms {
		if a.Description == "" {
			a.Description = taskDescription
		}

		caldavalarms += `
BEGIN:VALARM`
		switch a.RelativeTo {
		case models.ReminderRelationStartDate:
			caldavalarms += `
TRIGGER;RELATED=START:` + makeCalDavDuration(a.Duration)
		case models.ReminderRelationEndDate, models.ReminderRelationDueDate:
			caldavalarms += `
TRIGGER;RELATED=END:` + makeCalDavDuration(a.Duration)
		default:
			caldavalarms += `
TRIGGER;VALUE=DATE-TIME:` + makeCalDavTimeFromTimeStamp(a.Time)
		}
		caldavalarms += `
ACTION:DISPLAY
DESCRIPTION:` + a.Description + `
END:VALARM`
	}
	return caldavalarms
}

func ParseRelations(relations []Relation) (caldavrelatedtos string) {

	for _, r := range relations {
		switch r.Type {
		case models.RelationKindParenttask:
			caldavrelatedtos += `
RELATED-TO;RELTYPE=PARENT:`
		case models.RelationKindSubtask:
			caldavrelatedtos += `
RELATED-TO;RELTYPE=CHILD:`
		case models.RelationKindUnknown:
			continue
		case models.RelationKindRelated:
			continue
		case models.RelationKindDuplicateOf:
			continue
		case models.RelationKindDuplicates:
			continue
		case models.RelationKindBlocking:
			continue
		case models.RelationKindBlocked:
			continue
		case models.RelationKindPreceeds:
			continue
		case models.RelationKindFollows:
			continue
		case models.RelationKindCopiedFrom:
			continue
		case models.RelationKindCopiedTo:
			continue
		default:
			caldavrelatedtos += `
RELATED-TO:`
		}

		caldavrelatedtos += r.UID
	}

	return caldavrelatedtos
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
