// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package caldav

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestParseEvents(t *testing.T) {
	type args struct {
		config *Config
		events []*Event
	}
	tests := []struct {
		name             string
		args             args
		wantCaldavevents string
	}{
		{
			name: "Test caldavparsing without reminders",
			args: args{
				config: &Config{
					Name:   "test",
					ProdID: "RandomProdID which is not random",
					Color:  "ffffff",
				},
				events: []*Event{
					{
						Summary:     "Event #1",
						Description: "Lorem Ipsum",
						UID:         "randommduid",
						Timestamp:   time.Unix(1543626724, 0).In(config.GetTimeZone()),
						Start:       time.Unix(1543626724, 0).In(config.GetTimeZone()),
						End:         time.Unix(1543627824, 0).In(config.GetTimeZone()),
						Color:       "affffe",
					},
					{
						Summary:   "Event #2",
						UID:       "randommduidd",
						Timestamp: time.Unix(1543726724, 0).In(config.GetTimeZone()),
						Start:     time.Unix(1543726724, 0).In(config.GetTimeZone()),
						End:       time.Unix(1543738724, 0).In(config.GetTimeZone()),
					},
					{
						Summary:   "Event #3 with empty uid",
						UID:       "20181202T0600242aaef4a81d770c1e775e26bc5abebc87f1d3d7bffaa83",
						Timestamp: time.Unix(1543726824, 0).In(config.GetTimeZone()),
						Start:     time.Unix(1543726824, 0).In(config.GetTimeZone()),
						End:       time.Unix(1543727000, 0).In(config.GetTimeZone()),
					},
				},
			},
			wantCaldavevents: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
X-APPLE-CALENDAR-COLOR:#ffffffFF
X-OUTLOOK-COLOR:#ffffffFF
X-FUNAMBOL-COLOR:#ffffffFF
BEGIN:VEVENT
UID:randommduid
SUMMARY:Event #1
X-APPLE-CALENDAR-COLOR:#affffeFF
X-OUTLOOK-COLOR:#affffeFF
X-FUNAMBOL-COLOR:#affffeFF
DESCRIPTION:Lorem Ipsum
DTSTAMP:20181201T011204
DTSTART:20181201T011204
DTEND:20181201T013024
END:VEVENT
BEGIN:VEVENT
UID:randommduidd
SUMMARY:Event #2
DESCRIPTION:
DTSTAMP:20181202T045844
DTSTART:20181202T045844
DTEND:20181202T081844
END:VEVENT
BEGIN:VEVENT
UID:20181202T0600242aaef4a81d770c1e775e26bc5abebc87f1d3d7bffaa83
SUMMARY:Event #3 with empty uid
DESCRIPTION:
DTSTAMP:20181202T050024
DTSTART:20181202T050024
DTEND:20181202T050320
END:VEVENT
END:VCALENDAR`,
		},
		{
			name: "Test caldavparsing with reminders",
			args: args{
				config: &Config{
					Name:   "test2",
					ProdID: "RandomProdID which is not random",
				},
				events: []*Event{
					{
						Summary:     "Event #1",
						Description: "Lorem Ipsum",
						UID:         "randommduid",
						Timestamp:   time.Unix(1543626724, 0).In(config.GetTimeZone()),
						Start:       time.Unix(1543626724, 0).In(config.GetTimeZone()),
						End:         time.Unix(1543627824, 0).In(config.GetTimeZone()),
						Alarms: []Alarm{
							{Time: time.Unix(1543626524, 0).In(config.GetTimeZone())},
							{Time: time.Unix(1543626224, 0).In(config.GetTimeZone())},
							{Time: time.Unix(1543626024, 0)},
						},
					},
					{
						Summary:   "Event #2",
						UID:       "randommduidd",
						Timestamp: time.Unix(1543726724, 0).In(config.GetTimeZone()),
						Start:     time.Unix(1543726724, 0).In(config.GetTimeZone()),
						End:       time.Unix(1543738724, 0).In(config.GetTimeZone()),
						Alarms: []Alarm{
							{Time: time.Unix(1543626524, 0).In(config.GetTimeZone())},
							{Time: time.Unix(1543626224, 0).In(config.GetTimeZone())},
							{Time: time.Unix(1543626024, 0).In(config.GetTimeZone())},
						},
					},
					{
						Summary:   "Event #3 with empty uid",
						Timestamp: time.Unix(1543726824, 0).In(config.GetTimeZone()),
						Start:     time.Unix(1543726824, 0).In(config.GetTimeZone()),
						End:       time.Unix(1543727000, 0).In(config.GetTimeZone()),
						Alarms: []Alarm{
							{Time: time.Unix(1543626524, 0).In(config.GetTimeZone())},
							{Time: time.Unix(1543626224, 0).In(config.GetTimeZone())},
							{Time: time.Unix(1543626024, 0).In(config.GetTimeZone())},
							{Time: time.Unix(1543826824, 0).In(config.GetTimeZone())},
						},
					},
					{
						Summary:   "Event #4 without any",
						Timestamp: time.Unix(1543726824, 0),
						Start:     time.Unix(1543726824, 0),
						End:       time.Unix(1543727000, 0),
					},
				},
			},
			wantCaldavevents: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test2
PRODID:-//RandomProdID which is not random//EN
BEGIN:VEVENT
UID:randommduid
SUMMARY:Event #1
DESCRIPTION:Lorem Ipsum
DTSTAMP:20181201T011204
DTSTART:20181201T011204
DTEND:20181201T013024
BEGIN:VALARM
TRIGGER:-PT3M20S
ACTION:DISPLAY
DESCRIPTION:Event #1
END:VALARM
BEGIN:VALARM
TRIGGER:-PT8M20S
ACTION:DISPLAY
DESCRIPTION:Event #1
END:VALARM
BEGIN:VALARM
TRIGGER:-PT11M40S
ACTION:DISPLAY
DESCRIPTION:Event #1
END:VALARM
END:VEVENT
BEGIN:VEVENT
UID:randommduidd
SUMMARY:Event #2
DESCRIPTION:
DTSTAMP:20181202T045844
DTSTART:20181202T045844
DTEND:20181202T081844
BEGIN:VALARM
TRIGGER:-PT27H50M0S
ACTION:DISPLAY
DESCRIPTION:Event #2
END:VALARM
BEGIN:VALARM
TRIGGER:-PT27H55M0S
ACTION:DISPLAY
DESCRIPTION:Event #2
END:VALARM
BEGIN:VALARM
TRIGGER:-PT27H58M20S
ACTION:DISPLAY
DESCRIPTION:Event #2
END:VALARM
END:VEVENT
BEGIN:VEVENT
UID:20181202T0500242aaef4a81d770c1e775e26bc5abebc87f1d3d7bffaa83
SUMMARY:Event #3 with empty uid
DESCRIPTION:
DTSTAMP:20181202T050024
DTSTART:20181202T050024
DTEND:20181202T050320
BEGIN:VALARM
TRIGGER:-PT27H51M40S
ACTION:DISPLAY
DESCRIPTION:Event #3 with empty uid
END:VALARM
BEGIN:VALARM
TRIGGER:-PT27H56M40S
ACTION:DISPLAY
DESCRIPTION:Event #3 with empty uid
END:VALARM
BEGIN:VALARM
TRIGGER:-PT28H0M0S
ACTION:DISPLAY
DESCRIPTION:Event #3 with empty uid
END:VALARM
BEGIN:VALARM
TRIGGER:PT27H46M40S
ACTION:DISPLAY
DESCRIPTION:Event #3 with empty uid
END:VALARM
END:VEVENT
BEGIN:VEVENT
UID:20181202T050024ae7548ce9556df85038abe90dc674d4741a61ce74d1cf
SUMMARY:Event #4 without any
DESCRIPTION:
DTSTAMP:20181202T050024
DTSTART:20181202T050024
DTEND:20181202T050320
END:VEVENT
END:VCALENDAR`,
		},
		{
			name: "Test caldavparsing with multiline description",
			args: args{
				config: &Config{
					Name:   "test",
					ProdID: "RandomProdID which is not random",
				},
				events: []*Event{
					{
						Summary: "Event #1",
						Description: `Lorem Ipsum
Dolor sit amet`,
						UID:       "randommduid",
						Timestamp: time.Unix(1543626724, 0).In(config.GetTimeZone()),
						Start:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
						End:       time.Unix(1543627824, 0).In(config.GetTimeZone()),
					},
				},
			},
			wantCaldavevents: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VEVENT
UID:randommduid
SUMMARY:Event #1
DESCRIPTION:Lorem Ipsum\nDolor sit amet
DTSTAMP:20181201T011204
DTSTART:20181201T011204
DTEND:20181201T013024
END:VEVENT
END:VCALENDAR`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCaldavevents := ParseEvents(tt.args.config, tt.args.events)
			assert.Equal(t, gotCaldavevents, tt.wantCaldavevents)
		})
	}
}

func TestParseTodos(t *testing.T) {
	type args struct {
		config *Config
		todos  []*Todo
	}
	tests := []struct {
		name            string
		args            args
		wantCaldavtasks string
	}{
		{
			name: "Test caldavparsing with multiline description",
			args: args{
				config: &Config{
					Name:   "test",
					ProdID: "RandomProdID which is not random",
					Color:  "ffffff",
				},
				todos: []*Todo{
					{
						Summary: "Todo #1",
						Description: `Lorem Ipsum
Dolor sit amet`,
						UID:       "randommduid",
						Timestamp: time.Unix(1543626724, 0).In(config.GetTimeZone()),
						Color:     "affffe",
					},
				},
			},
			wantCaldavtasks: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
X-APPLE-CALENDAR-COLOR:#ffffffFF
X-OUTLOOK-COLOR:#ffffffFF
X-FUNAMBOL-COLOR:#ffffffFF
BEGIN:VTODO
UID:randommduid
DTSTAMP:20181201T011204
SUMMARY:Todo #1
X-APPLE-CALENDAR-COLOR:#affffeFF
X-OUTLOOK-COLOR:#affffeFF
X-FUNAMBOL-COLOR:#affffeFF
DESCRIPTION:Lorem Ipsum\nDolor sit amet
LAST-MODIFIED:00010101T000000
END:VTODO
END:VCALENDAR`,
		},
		{
			name: "Test caldavparsing with completed task",
			args: args{
				config: &Config{
					Name:   "test",
					ProdID: "RandomProdID which is not random",
				},
				todos: []*Todo{
					{
						Summary:     "Todo #1",
						Description: "Lorem Ipsum",
						UID:         "randommduid",
						Timestamp:   time.Unix(1543626724, 0).In(config.GetTimeZone()),
						Completed:   time.Unix(1543627824, 0).In(config.GetTimeZone()),
					},
				},
			},
			wantCaldavtasks: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randommduid
DTSTAMP:20181201T011204
SUMMARY:Todo #1
DESCRIPTION:Lorem Ipsum
COMPLETED:20181201T013024
STATUS:COMPLETED
LAST-MODIFIED:00010101T000000
END:VTODO
END:VCALENDAR`,
		},
		{
			name: "with priority",
			args: args{
				config: &Config{
					Name:   "test",
					ProdID: "RandomProdID which is not random",
				},
				todos: []*Todo{
					{
						Summary:     "Todo #1",
						Description: "Lorem Ipsum",
						UID:         "randommduid",
						Priority:    1,
						Timestamp:   time.Unix(1543626724, 0).In(config.GetTimeZone()),
					},
				},
			},
			wantCaldavtasks: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randommduid
DTSTAMP:20181201T011204
SUMMARY:Todo #1
DESCRIPTION:Lorem Ipsum
PRIORITY:9
LAST-MODIFIED:00010101T000000
END:VTODO
END:VCALENDAR`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCaldavtasks := ParseTodos(tt.args.config, tt.args.todos)
			assert.Equal(t, gotCaldavtasks, tt.wantCaldavtasks)
		})
	}
}
