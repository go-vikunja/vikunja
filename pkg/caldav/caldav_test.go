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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
)

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
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
X-APPLE-CALENDAR-COLOR:#ffffffFF
X-OUTLOOK-COLOR:#ffffffFF
X-FUNAMBOL-COLOR:#ffffffFF
COLOR:#ffffffFF
BEGIN:VTODO
UID:randommduid
DTSTAMP:20181201T011204Z
SUMMARY:Todo #1
X-APPLE-CALENDAR-COLOR:#affffeFF
X-OUTLOOK-COLOR:#affffeFF
X-FUNAMBOL-COLOR:#affffeFF
COLOR:#affffeFF
DESCRIPTION:Lorem Ipsum\nDolor sit amet
LAST-MODIFIED:00010101T000000Z
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
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randommduid
DTSTAMP:20181201T011204Z
SUMMARY:Todo #1
DESCRIPTION:Lorem Ipsum
COMPLETED:20181201T013024Z
STATUS:COMPLETED
LAST-MODIFIED:00010101T000000Z
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
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randommduid
DTSTAMP:20181201T011204Z
SUMMARY:Todo #1
DESCRIPTION:Lorem Ipsum
PRIORITY:9
LAST-MODIFIED:00010101T000000Z
END:VTODO
END:VCALENDAR`,
		},
		{
			name: "with repeating monthly",
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
						Repeats:     "FREQ=MONTHLY;BYMONTHDAY=01",
						DueDate:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
					},
				},
			},
			wantCaldavtasks: `BEGIN:VCALENDAR
VERSION:2.0
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randommduid
DTSTAMP:20181201T011204Z
SUMMARY:Todo #1
DESCRIPTION:Lorem Ipsum
DUE:20181201T011204Z
RRULE:FREQ=MONTHLY;BYMONTHDAY=01
LAST-MODIFIED:00010101T000000Z
END:VTODO
END:VCALENDAR`,
		},
		{
			name: "with repeat interval",
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
						DueDate:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
						Repeats:     "FREQ=DAILY;INTERVAL=1",
					},
				},
			},
			wantCaldavtasks: `BEGIN:VCALENDAR
VERSION:2.0
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randommduid
DTSTAMP:20181201T011204Z
SUMMARY:Todo #1
DESCRIPTION:Lorem Ipsum
DUE:20181201T011204Z
RRULE:FREQ=DAILY;INTERVAL=1
LAST-MODIFIED:00010101T000000Z
END:VTODO
END:VCALENDAR`,
		},
		{
			name: "with categories",
			args: args{
				config: &Config{
					Name:   "test",
					ProdID: "RandomProdID which is not random",
					Color:  "ffffff",
				},
				todos: []*Todo{
					{
						Summary:    "Todo #1",
						UID:        "randommduid",
						Timestamp:  time.Unix(1543626724, 0).In(config.GetTimeZone()),
						Color:      "affffe",
						Categories: []string{"label1", "label2"},
					},
				},
			},
			wantCaldavtasks: `BEGIN:VCALENDAR
VERSION:2.0
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
X-APPLE-CALENDAR-COLOR:#ffffffFF
X-OUTLOOK-COLOR:#ffffffFF
X-FUNAMBOL-COLOR:#ffffffFF
COLOR:#ffffffFF
BEGIN:VTODO
UID:randommduid
DTSTAMP:20181201T011204Z
SUMMARY:Todo #1
X-APPLE-CALENDAR-COLOR:#affffeFF
X-OUTLOOK-COLOR:#affffeFF
X-FUNAMBOL-COLOR:#affffeFF
COLOR:#affffeFF
CATEGORIES:label1,label2
LAST-MODIFIED:00010101T000000Z
END:VTODO
END:VCALENDAR`,
		},
		{
			name: "with alarm",
			args: args{
				config: &Config{
					Name:   "test",
					ProdID: "RandomProdID which is not random",
				},
				todos: []*Todo{
					{
						Summary:   "Todo #1",
						UID:       "randommduid",
						Timestamp: time.Unix(1543626724, 0).In(config.GetTimeZone()),
						Alarms: []Alarm{
							{
								Time: time.Unix(1543626724, 0).In(config.GetTimeZone()),
							},
							{
								Time:        time.Unix(1543626724, 0).In(config.GetTimeZone()),
								Description: "alarm description",
							},
							{
								Duration:   -2 * time.Hour,
								RelativeTo: "due_date",
							},
							{
								Duration:   1 * time.Hour,
								RelativeTo: "start_date",
							},
							{
								Duration:   time.Duration(0),
								RelativeTo: "end_date",
							},
						},
					},
				},
			},
			wantCaldavtasks: `BEGIN:VCALENDAR
VERSION:2.0
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randommduid
DTSTAMP:20181201T011204Z
SUMMARY:Todo #1
LAST-MODIFIED:00010101T000000Z
BEGIN:VALARM
TRIGGER;VALUE=DATE-TIME:20181201T011204Z
ACTION:DISPLAY
DESCRIPTION:Todo #1
END:VALARM
BEGIN:VALARM
TRIGGER;VALUE=DATE-TIME:20181201T011204Z
ACTION:DISPLAY
DESCRIPTION:alarm description
END:VALARM
BEGIN:VALARM
TRIGGER;RELATED=END:-PT2H0M0S
ACTION:DISPLAY
DESCRIPTION:Todo #1
END:VALARM
BEGIN:VALARM
TRIGGER;RELATED=START:PT1H0M0S
ACTION:DISPLAY
DESCRIPTION:Todo #1
END:VALARM
BEGIN:VALARM
TRIGGER;RELATED=END:PT0S
ACTION:DISPLAY
DESCRIPTION:Todo #1
END:VALARM
END:VTODO
END:VCALENDAR`,
		},
		{
			name: "with related-to",
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
						Relations: []Relation{
							{
								Type: models.RelationKindParenttask,
								UID:  "parentuid",
							},
							{
								Type: models.RelationKindSubtask,
								UID:  "subtaskuid",
							},
						},
						Timestamp: time.Unix(1543626724, 0).In(config.GetTimeZone()),
					},
				},
			},
			wantCaldavtasks: `BEGIN:VCALENDAR
VERSION:2.0
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randommduid
DTSTAMP:20181201T011204Z
SUMMARY:Todo #1
DESCRIPTION:Lorem Ipsum
LAST-MODIFIED:00010101T000000Z
RELATED-TO;RELTYPE=PARENT:parentuid
RELATED-TO;RELTYPE=CHILD:subtaskuid
END:VTODO
END:VCALENDAR`,
		},
		{
			name: "GHSA-2g7h-7rqr-9p4r: CRLF injection in Summary is escaped",
			args: args{
				config: &Config{
					Name:   "test",
					ProdID: "RandomProdID which is not random",
				},
				todos: []*Todo{
					{
						Summary:   "Meeting\r\nATTACH:https://evil.com/malware.exe\r\nX-INJECTED:pwned",
						UID:       "randommduid",
						Timestamp: time.Unix(1543626724, 0).In(config.GetTimeZone()),
					},
				},
			},
			wantCaldavtasks: `BEGIN:VCALENDAR
VERSION:2.0
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randommduid
DTSTAMP:20181201T011204Z
SUMMARY:Meeting\nATTACH:https://evil.com/malware.exe\nX-INJECTED:pwned
LAST-MODIFIED:00010101T000000Z
END:VTODO
END:VCALENDAR`,
		},
		{
			name: "GHSA-2g7h-7rqr-9p4r: semicolons and commas in Summary are escaped",
			args: args{
				config: &Config{
					Name:   "te;st,ed",
					ProdID: "RandomProdID which is not random",
				},
				todos: []*Todo{
					{
						Summary:    "a;b,c\\d",
						UID:        "randommduid",
						Timestamp:  time.Unix(1543626724, 0).In(config.GetTimeZone()),
						Categories: []string{"lab;el1", "lab,el2"},
						Organizer:  &user.User{Username: "al;ic,e"},
					},
				},
			},
			wantCaldavtasks: `BEGIN:VCALENDAR
VERSION:2.0
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:te\;st\,ed
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randommduid
DTSTAMP:20181201T011204Z
SUMMARY:a\;b\,c\\d
ORGANIZER;CN=:al\;ic\,e
CATEGORIES:lab\;el1,lab\,el2
LAST-MODIFIED:00010101T000000Z
END:VTODO
END:VCALENDAR`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCaldavtasks := ParseTodos(tt.args.config, tt.args.todos)
			assert.Equal(t, tt.wantCaldavtasks, gotCaldavtasks)
			assert.NotContains(t, gotCaldavtasks, "METHOD:")
		})
	}
}

func TestGetCaldavColor(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"plain hex", "ff8800", "\nX-APPLE-CALENDAR-COLOR:#ff8800FF\nX-OUTLOOK-COLOR:#ff8800FF\nX-FUNAMBOL-COLOR:#ff8800FF\nCOLOR:#ff8800FF"},
		{"leading hash", "#ff8800", "\nX-APPLE-CALENDAR-COLOR:#ff8800FF\nX-OUTLOOK-COLOR:#ff8800FF\nX-FUNAMBOL-COLOR:#ff8800FF\nCOLOR:#ff8800FF"},
		{"mixed case", "AaBbCc", "\nX-APPLE-CALENDAR-COLOR:#AaBbCcFF\nX-OUTLOOK-COLOR:#AaBbCcFF\nX-FUNAMBOL-COLOR:#AaBbCcFF\nCOLOR:#AaBbCcFF"},
		{"CRLF injection stripped", "a\r\nB", "\nX-APPLE-CALENDAR-COLOR:#aBFF\nX-OUTLOOK-COLOR:#aBFF\nX-FUNAMBOL-COLOR:#aBFF\nCOLOR:#aBFF"},
		{"property injection stripped", "ff\r\nATTACH:https://evil.com", "\nX-APPLE-CALENDAR-COLOR:#ffAACecFF\nX-OUTLOOK-COLOR:#ffAACecFF\nX-FUNAMBOL-COLOR:#ffAACecFF\nCOLOR:#ffAACecFF"},
		{"non-hex chars dropped entirely", "zz!@#", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCaldavColor(tt.in)
			assert.Equal(t, tt.want, got)
			// No output may ever contain CR or LF inside a property value.
			assert.NotContains(t, got, "\r")
		})
	}
}

func TestEscapeICalText(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"plain ASCII", "Hello World", "Hello World"},
		{"backslash", `a\b`, `a\\b`},
		{"semicolon", "a;b", `a\;b`},
		{"comma", "a,b", `a\,b`},
		{"newline LF", "a\nb", `a\nb`},
		{"carriage return LF pair", "a\r\nb", `a\nb`},
		{"lone carriage return", "a\rb", `a\nb`},
		{"multiple specials", `a\;b,c` + "\n" + "d", `a\\\;b\,c\nd`},
		{"backslash-n in source stays separate", `a\nb`, `a\\nb`},
		{"advisory PoC (CRLF + ATTACH)", "Meeting\r\nATTACH:https://evil.com/malware.exe", `Meeting\nATTACH:https://evil.com/malware.exe`},
		{"colon is NOT escaped (not a TEXT special)", "a:b", "a:b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeICalText(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}
