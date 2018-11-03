package caldav

import (
	"code.vikunja.io/api/pkg/utils"
	"strconv"
	"time"
)

// Event holds a single caldav event
type Event struct {
	Summary     string
	Description string
	UID         string
	Alarms      []Alarm

	TimestampUnix int64
	StartUnix     int64
	EndUnix       int64
}

// Alarm holds infos about an alarm from a caldav event
type Alarm struct {
	TimeUnix    int64
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
			e.UID = makeCalDavTimeFromUnixTime(e.TimestampUnix) + utils.Sha256(e.Summary)
		}

		caldavevents += `
BEGIN:VEVENT
UID:` + e.UID + `
SUMMARY:` + e.Summary + `
DESCRIPTION:` + e.Description + `
DTSTAMP:` + makeCalDavTimeFromUnixTime(e.TimestampUnix) + `
DTSTART:` + makeCalDavTimeFromUnixTime(e.StartUnix) + `
DTEND:` + makeCalDavTimeFromUnixTime(e.EndUnix)

		for _, a := range e.Alarms {
			if a.Description == "" {
				a.Description = e.Summary
			}

			caldavevents += `
BEGIN:VALARM
TRIGGER:` + calcAlarmDateFromReminder(e.StartUnix, a.TimeUnix) + `
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

func makeCalDavTimeFromUnixTime(unixtime int64) (caldavtime string) {
	tm := time.Unix(unixtime, 0)
	return tm.Format("20060102T150405")
}

func calcAlarmDateFromReminder(eventStartUnix, reminderUnix int64) (alarmTime string) {
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
