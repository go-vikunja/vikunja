// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package caldav

import (
	"code.vikunja.io/api/pkg/timeutil"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"fmt"
	"strconv"
	"time"
)

// DateFormat ist the caldav date format
const DateFormat = `20060102T150405`

// Event holds a single caldav event
type Event struct {
	Summary     string
	Description string
	UID         string
	Alarms      []Alarm

	Timestamp timeutil.TimeStamp
	Start     timeutil.TimeStamp
	End       timeutil.TimeStamp
}

// Todo holds a single VTODO
type Todo struct {
	// Required
	Timestamp timeutil.TimeStamp
	UID       string

	// Optional
	Summary      string
	Description  string
	Completed    timeutil.TimeStamp
	Organizer    *user.User
	Priority     int64 // 0-9, 1 is highest
	RelatedToUID string

	Start    timeutil.TimeStamp
	End      timeutil.TimeStamp
	DueDate  timeutil.TimeStamp
	Duration time.Duration

	Created timeutil.TimeStamp
	Updated timeutil.TimeStamp // last-mod
}

// Alarm holds infos about an alarm from a caldav event
type Alarm struct {
	Time        timeutil.TimeStamp
	Description string
}

// Config is the caldav calendar config
type Config struct {
	Name   string
	ProdID string
}

// ParseEvents parses an array of caldav events and gives them back as string
func ParseEvents(config *Config, events []*Event) (caldavevents string) {
	caldavevents += `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:` + config.Name + `
PRODID:-//` + config.ProdID + `//EN`

	for _, e := range events {

		if e.UID == "" {
			e.UID = makeCalDavTimeFromTimeStamp(e.Timestamp) + utils.Sha256(e.Summary)
		}

		caldavevents += `
BEGIN:VEVENT
UID:` + e.UID + `
SUMMARY:` + e.Summary + `
DESCRIPTION:` + e.Description + `
DTSTAMP:` + makeCalDavTimeFromTimeStamp(e.Timestamp) + `
DTSTART:` + makeCalDavTimeFromTimeStamp(e.Start) + `
DTEND:` + makeCalDavTimeFromTimeStamp(e.End)

		for _, a := range e.Alarms {
			if a.Description == "" {
				a.Description = e.Summary
			}

			caldavevents += `
BEGIN:VALARM
TRIGGER:` + calcAlarmDateFromReminder(e.Start, a.Time) + `
ACTION:DISPLAY
DESCRIPTION:` + a.Description + `
END:VALARM`
		}
		caldavevents += `
END:VEVENT`
	}

	caldavevents += `
END:VCALENDAR` // Need a line break

	return
}

// ParseTodos returns a caldav vcalendar string with todos
func ParseTodos(config *Config, todos []*Todo) (caldavtodos string) {
	caldavtodos = `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:` + config.Name + `
PRODID:-//` + config.ProdID + `//EN`

	for _, t := range todos {
		if t.UID == "" {
			t.UID = makeCalDavTimeFromTimeStamp(t.Timestamp) + utils.Sha256(t.Summary)
		}

		caldavtodos += `
BEGIN:VTODO
UID:` + t.UID + `
DTSTAMP:` + makeCalDavTimeFromTimeStamp(t.Timestamp) + `
SUMMARY:` + t.Summary

		if t.Start != 0 {
			caldavtodos += `
DTSTART: ` + makeCalDavTimeFromTimeStamp(t.Start)
		}
		if t.End != 0 {
			caldavtodos += `
DTEND: ` + makeCalDavTimeFromTimeStamp(t.End)
		}
		if t.Description != "" {
			caldavtodos += `
DESCRIPTION:` + t.Description
		}
		if t.Completed != 0 {
			caldavtodos += `
COMPLETED: ` + makeCalDavTimeFromTimeStamp(t.Completed)
		}
		if t.Organizer != nil {
			caldavtodos += `
ORGANIZER;CN=:` + t.Organizer.Username
		}

		if t.RelatedToUID != "" {
			caldavtodos += `
RELATED-TO:` + t.RelatedToUID
		}

		if t.DueDate != 0 {
			caldavtodos += `
DUE:` + makeCalDavTimeFromTimeStamp(t.DueDate)
		}

		if t.Created != 0 {
			caldavtodos += `
CREATED:` + makeCalDavTimeFromTimeStamp(t.Created)
		}

		if t.Duration != 0 {
			caldavtodos += `
DURATION:PT` + fmt.Sprintf("%.6f", t.Duration.Hours()) + `H` + fmt.Sprintf("%.6f", t.Duration.Minutes()) + `M` + fmt.Sprintf("%.6f", t.Duration.Seconds()) + `S`
		}

		if t.Priority != 0 {
			caldavtodos += `
PRIORITY:` + strconv.Itoa(int(t.Priority))
		}

		caldavtodos += `
LAST-MODIFIED:` + makeCalDavTimeFromTimeStamp(t.Updated)

		caldavtodos += `
END:VTODO`
	}

	caldavtodos += `
END:VCALENDAR` // Need a line break

	return
}

func makeCalDavTimeFromTimeStamp(ts timeutil.TimeStamp) (caldavtime string) {
	tz, _ := time.LoadLocation("UTC")
	return ts.ToTime().In(tz).Format(DateFormat)
}

func calcAlarmDateFromReminder(eventStartUnix, reminderUnix timeutil.TimeStamp) (alarmTime string) {
	if eventStartUnix > reminderUnix {
		alarmTime += `-`
	}
	alarmTime += `PT`
	diff := eventStartUnix - reminderUnix
	if diff < 0 { // Make it positive
		diff = diff * -1
	}
	alarmTime += strconv.Itoa(int(diff/60)) + "M"
	return
}
