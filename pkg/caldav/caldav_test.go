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

import "testing"

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
				},
				events: []*Event{
					{
						Summary:     "Event #1",
						Description: "Lorem Ipsum",
						UID:         "randommduid",
						Timestamp:   1543626724,
						Start:       1543626724,
						End:         1543627824,
					},
					{
						Summary:   "Event #2",
						UID:       "randommduidd",
						Timestamp: 1543726724,
						Start:     1543726724,
						End:       1543738724,
					},
					{
						Summary:   "Event #3 with empty uid",
						UID:       "20181202T0600242aaef4a81d770c1e775e26bc5abebc87f1d3d7bffaa83",
						Timestamp: 1543726824,
						Start:     1543726824,
						End:       1543727000,
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
						Timestamp:   1543626724,
						Start:       1543626724,
						End:         1543627824,
						Alarms: []Alarm{
							{Time: 1543626524},
							{Time: 1543626224},
							{Time: 1543626024},
						},
					},
					{
						Summary:   "Event #2",
						UID:       "randommduidd",
						Timestamp: 1543726724,
						Start:     1543726724,
						End:       1543738724,
						Alarms: []Alarm{
							{Time: 1543626524},
							{Time: 1543626224},
							{Time: 1543626024},
						},
					},
					{
						Summary:   "Event #3 with empty uid",
						Timestamp: 1543726824,
						Start:     1543726824,
						End:       1543727000,
						Alarms: []Alarm{
							{Time: 1543626524},
							{Time: 1543626224},
							{Time: 1543626024},
							{Time: 1543826824},
						},
					},
					{
						Summary:   "Event #4 without any",
						Timestamp: 1543726824,
						Start:     1543726824,
						End:       1543727000,
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
TRIGGER:-PT3M
ACTION:DISPLAY
DESCRIPTION:Event #1
END:VALARM
BEGIN:VALARM
TRIGGER:-PT8M
ACTION:DISPLAY
DESCRIPTION:Event #1
END:VALARM
BEGIN:VALARM
TRIGGER:-PT11M
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
TRIGGER:-PT1670M
ACTION:DISPLAY
DESCRIPTION:Event #2
END:VALARM
BEGIN:VALARM
TRIGGER:-PT1675M
ACTION:DISPLAY
DESCRIPTION:Event #2
END:VALARM
BEGIN:VALARM
TRIGGER:-PT1678M
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
TRIGGER:-PT1671M
ACTION:DISPLAY
DESCRIPTION:Event #3 with empty uid
END:VALARM
BEGIN:VALARM
TRIGGER:-PT1676M
ACTION:DISPLAY
DESCRIPTION:Event #3 with empty uid
END:VALARM
BEGIN:VALARM
TRIGGER:-PT1680M
ACTION:DISPLAY
DESCRIPTION:Event #3 with empty uid
END:VALARM
BEGIN:VALARM
TRIGGER:PT1666M
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotCaldavevents := ParseEvents(tt.args.config, tt.args.events); gotCaldavevents != tt.wantCaldavevents {
				t.Errorf("ParseEvents() = %v, want %v", gotCaldavevents, tt.wantCaldavevents)
			}
		})
	}
}
