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
METHOD:PUBLISH
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
METHOD:PUBLISH
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
METHOD:PUBLISH
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
						RepeatMode:  models.TaskRepeatModeMonth,
						DueDate:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
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
			name: "with repeat mode default",
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
						RepeatMode:  models.TaskRepeatModeDefault,
						DueDate:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
						RepeatAfter: 435,
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
DTSTAMP:20181201T011204Z
SUMMARY:Todo #1
DESCRIPTION:Lorem Ipsum
DUE:20181201T011204Z
RRULE:FREQ=SECONDLY;INTERVAL=435
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
METHOD:PUBLISH
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
METHOD:PUBLISH
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
METHOD:PUBLISH
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCaldavtasks := ParseTodos(tt.args.config, tt.args.todos)
			assert.Equal(t, tt.wantCaldavtasks, gotCaldavtasks)
		})
	}
}
