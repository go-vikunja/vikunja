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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"gopkg.in/d4l3k/messagediff.v1"
)

func TestParseTaskFromVTODO(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		name      string
		args      args
		wantVTask *models.Task
		wantErr   bool
	}{
		{
			name: "normal",
			args: args{content: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randomuid
DTSTAMP:20181201T011204
SUMMARY:Todo #1
DESCRIPTION:Lorem Ipsum
LAST-MODIFIED:00010101T000000
END:VTODO
END:VCALENDAR`,
			},
			wantVTask: &models.Task{
				Title:       "Todo #1",
				UID:         "randomuid",
				Description: "Lorem Ipsum",
				Updated:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
			},
		},
		{
			name: "With priority",
			args: args{content: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randomuid
DTSTAMP:20181201T011204
SUMMARY:Todo #1
DESCRIPTION:Lorem Ipsum
PRIORITY:9
LAST-MODIFIED:00010101T000000
END:VTODO
END:VCALENDAR`,
			},
			wantVTask: &models.Task{
				Title:       "Todo #1",
				UID:         "randomuid",
				Description: "Lorem Ipsum",
				Priority:    1,
				Updated:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
			},
		},
		{
			name: "With categories",
			args: args{content: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randomuid
DTSTAMP:20181201T011204
SUMMARY:Todo #1
DESCRIPTION:Lorem Ipsum
CATEGORIES:cat1,cat2
LAST-MODIFIED:00010101T000000
END:VTODO
END:VCALENDAR`,
			},
			wantVTask: &models.Task{
				Title:       "Todo #1",
				UID:         "randomuid",
				Description: "Lorem Ipsum",
				Labels: []*models.Label{
					{
						Title: "cat1",
					},
					{
						Title: "cat2",
					},
				},
				Updated: time.Unix(1543626724, 0).In(config.GetTimeZone()),
			},
		},
		{
			name: "With alarm (time trigger)",
			args: args{content: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randomuid
DTSTAMP:20181201T011204
SUMMARY:Todo #1
DESCRIPTION:Lorem Ipsum
BEGIN:VALARM
TRIGGER;VALUE=DATE-TIME:20181201T011210Z
ACTION:DISPLAY
END:VALARM
END:VTODO
END:VCALENDAR`,
			},
			wantVTask: &models.Task{
				Title:       "Todo #1",
				UID:         "randomuid",
				Description: "Lorem Ipsum",
				Reminders: []*models.TaskReminder{
					{
						Reminder: time.Date(2018, 12, 1, 1, 12, 10, 0, config.GetTimeZone()),
					},
				},
				Updated: time.Unix(1543626724, 0).In(config.GetTimeZone()),
			},
		},
		{
			name: "With alarm (relative trigger)",
			args: args{content: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randomuid
DTSTAMP:20181201T011204
SUMMARY:Todo #1
DESCRIPTION:Lorem Ipsum
DTSTART:20230228T170000Z
DUE:20230304T150000Z
BEGIN:VALARM
TRIGGER:PT0S
ACTION:DISPLAY
END:VALARM
BEGIN:VALARM
TRIGGER;VALUE=DURATION:-PT60M
ACTION:DISPLAY
END:VALARM
BEGIN:VALARM
TRIGGER:-PT61M
ACTION:DISPLAY
END:VALARM
BEGIN:VALARM
TRIGGER;RELATED=START:-P1D
ACTION:DISPLAY
END:VALARM
BEGIN:VALARM
TRIGGER;RELATED=END:-PT30M
ACTION:DISPLAY
END:VALARM
END:VTODO
END:VCALENDAR`,
			},
			wantVTask: &models.Task{
				Title:       "Todo #1",
				UID:         "randomuid",
				Description: "Lorem Ipsum",
				StartDate:   time.Date(2023, 2, 28, 17, 0, 0, 0, config.GetTimeZone()),
				DueDate:     time.Date(2023, 3, 4, 15, 0, 0, 0, config.GetTimeZone()),
				Reminders: []*models.TaskReminder{
					{
						RelativeTo:     models.ReminderRelationStartDate,
						RelativePeriod: 0,
					},
					{
						RelativeTo:     models.ReminderRelationStartDate,
						RelativePeriod: -3600,
					},
					{
						RelativeTo:     models.ReminderRelationStartDate,
						RelativePeriod: -3660,
					},
					{
						RelativeTo:     models.ReminderRelationStartDate,
						RelativePeriod: -86400,
					},
					{
						RelativeTo:     models.ReminderRelationDueDate,
						RelativePeriod: -1800,
					},
				},
				Updated: time.Unix(1543626724, 0).In(config.GetTimeZone()),
			},
		},
		{
			name: "With parent",
			args: args{content: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randomuid
DTSTAMP:20181201T011204
SUMMARY:SubTask #1
DESCRIPTION:Lorem Ipsum
LAST-MODIFIED:00010101T000000
RELATED-TO;RELTYPE=PARENT:randomuid_parent
END:VTODO
END:VCALENDAR`,
			},
			wantVTask: &models.Task{
				Title:       "SubTask #1",
				UID:         "randomuid",
				Description: "Lorem Ipsum",
				Updated:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
				RelatedTasks: map[models.RelationKind][]*models.Task{
					models.RelationKindParenttask: {
						{
							UID: "randomuid_parent",
						},
					},
				},
			},
		},
		{
			name: "With subtask",
			args: args{content: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randomuid
DTSTAMP:20181201T011204
SUMMARY:Parent
DESCRIPTION:Lorem Ipsum
LAST-MODIFIED:00010101T000000
RELATED-TO;RELTYPE=CHILD:randomuid_child
END:VTODO
END:VCALENDAR`,
			},
			wantVTask: &models.Task{
				Title:       "Parent",
				UID:         "randomuid",
				Description: "Lorem Ipsum",
				Updated:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
				RelatedTasks: map[models.RelationKind][]*models.Task{
					models.RelationKindSubtask: {
						{
							UID: "randomuid_child",
						},
					},
				},
			},
		},
		{
			name: "example task from tasks.org app",
			args: args{content: `BEGIN:VCALENDAR
VERSION:2.0
PRODID:+//IDN tasks.org//android-130102//EN
BEGIN:VTODO
DTSTAMP:20230402T074158Z
UID:4290517349243274514
CREATED:20230402T060451Z
LAST-MODIFIED:20230402T074154Z
SUMMARY:Test with tasks.org
PRIORITY:9
CATEGORIES:Vikunja
X-APPLE-SORT-ORDER:697384109
DUE;TZID=Europe/Berlin:20230402T170001
DTSTART;TZID=Europe/Berlin:20230401T090000
BEGIN:VALARM
TRIGGER;RELATED=END:PT0S
ACTION:DISPLAY
DESCRIPTION:Default Tasks.org description
END:VALARM
BEGIN:VALARM
TRIGGER;VALUE=DATE-TIME:20230402T100000Z
ACTION:DISPLAY
DESCRIPTION:Default Tasks.org description
END:VALARM
END:VTODO
BEGIN:VTIMEZONE
TZID:Europe/Berlin
LAST-MODIFIED:20220816T024022Z
BEGIN:DAYLIGHT
TZNAME:CEST
TZOFFSETFROM:+0100
TZOFFSETTO:+0200
DTSTART:19810329T020000
RRULE:FREQ=YEARLY;BYMONTH=3;BYDAY=-1SU
END:DAYLIGHT
BEGIN:STANDARD
TZNAME:CET
TZOFFSETFROM:+0200
TZOFFSETTO:+0100
DTSTART:19961027T030000
RRULE:FREQ=YEARLY;BYMONTH=10;BYDAY=-1SU
END:STANDARD
END:VTIMEZONE
END:VCALENDAR`,
			},
			wantVTask: &models.Task{
				Updated:  time.Date(2023, 4, 2, 7, 41, 58, 0, config.GetTimeZone()),
				UID:      "4290517349243274514",
				Title:    "Test with tasks.org",
				Priority: 1,
				Labels: []*models.Label{
					{
						Title: "Vikunja",
					},
				},
				DueDate:   time.Date(2023, 4, 2, 15, 0, 1, 0, config.GetTimeZone()),
				StartDate: time.Date(2023, 4, 1, 7, 0, 0, 0, config.GetTimeZone()),
				Reminders: []*models.TaskReminder{
					{
						RelativeTo:     models.ReminderRelationDueDate,
						RelativePeriod: 0,
					},
					{
						Reminder: time.Date(2023, 4, 2, 10, 0, 0, 0, config.GetTimeZone()),
					},
				},
			},
		},
		{
			name: "with apple hex color",
			args: args{content: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
X-APPLE-CALENDAR-COLOR:#affffeFF
END:VTODO
END:VCALENDAR`,
			},
			wantVTask: &models.Task{
				HexColor: "affffe",
			},
		},
		{
			name: "with apple css color",
			args: args{content: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
X-APPLE-CALENDAR-COLOR:mediumslateblue
END:VTODO
END:VCALENDAR`,
			},
			wantVTask: &models.Task{
				HexColor: "7b68ee",
			},
		},
		{
			name: "with outlook hex color",
			args: args{content: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
X-OUTLOOK-COLOR:#affffeFF
END:VTODO
END:VCALENDAR`,
			},
			wantVTask: &models.Task{
				HexColor: "affffe",
			},
		},
		{
			name: "with outlook css color",
			args: args{content: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
X-OUTLOOK-COLOR:mediumslateblue
END:VTODO
END:VCALENDAR`,
			},
			wantVTask: &models.Task{
				HexColor: "7b68ee",
			},
		},
		{
			name: "with funambol hex color",
			args: args{content: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
X-FUNAMBOL-COLOR:#affffeFF
END:VTODO
END:VCALENDAR`,
			},
			wantVTask: &models.Task{
				HexColor: "affffe",
			},
		},
		{
			name: "with funambol css color",
			args: args{content: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
X-FUNAMBOL-COLOR:mediumslateblue
END:VTODO
END:VCALENDAR`,
			},
			wantVTask: &models.Task{
				HexColor: "7b68ee",
			},
		},
		{
			name: "with hex color",
			args: args{content: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
COLOR:#affffeFF
END:VTODO
END:VCALENDAR`,
			},
			wantVTask: &models.Task{
				HexColor: "affffe",
			},
		},
		{
			name: "with css color",
			args: args{content: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
COLOR:mediumslateblue
END:VTODO
END:VCALENDAR`,
			},
			wantVTask: &models.Task{
				HexColor: "7b68ee",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTaskFromVTODO(tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTaskFromVTODO() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff, equal := messagediff.PrettyDiff(got, tt.wantVTask); !equal {
				t.Errorf("ParseTaskFromVTODO()\n gotVTask = %v\n want %v\n diff = %s", got, tt.wantVTask, diff)
			}
		})
	}
}

func TestGetCaldavTodosForTasks(t *testing.T) {
	type args struct {
		list  *models.ProjectWithTasksAndBuckets
		tasks []*models.TaskWithComments
	}
	tests := []struct {
		name       string
		args       args
		wantCaldav string
	}{
		{
			name: "Format single Task as CalDAV",
			args: args{
				list: &models.ProjectWithTasksAndBuckets{
					Project: models.Project{
						Title: "List title",
					},
				},
				tasks: []*models.TaskWithComments{
					{
						Task: models.Task{
							Title:       "Task 1",
							UID:         "randomuid",
							Description: "Description",
							Priority:    3,
							Created:     time.Unix(1543626721, 0).In(config.GetTimeZone()),
							DueDate:     time.Unix(1543626722, 0).In(config.GetTimeZone()),
							StartDate:   time.Unix(1543626723, 0).In(config.GetTimeZone()),
							EndDate:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
							Updated:     time.Unix(1543626725, 0).In(config.GetTimeZone()),
							DoneAt:      time.Unix(1543626726, 0).In(config.GetTimeZone()),
							RepeatAfter: 86400,
							Labels: []*models.Label{
								{
									ID:    1,
									Title: "label1",
								},
								{
									ID:    2,
									Title: "label2",
								},
							},
							Reminders: []*models.TaskReminder{
								{
									Reminder: time.Unix(1543626730, 0).In(config.GetTimeZone()),
								},
								{
									Reminder:       time.Unix(1543626731, 0).In(config.GetTimeZone()),
									RelativePeriod: -3600,
									RelativeTo:     models.ReminderRelationDueDate,
								},
							},
						},
					},
				},
			},
			wantCaldav: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:List title
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:randomuid
DTSTAMP:20181201T011205Z
SUMMARY:Task 1
DTSTART:20181201T011203Z
DTEND:20181201T011204Z
DESCRIPTION:Description
COMPLETED:20181201T011206Z
STATUS:COMPLETED
DUE:20181201T011202Z
CREATED:20181201T011201Z
PRIORITY:3
RRULE:FREQ=DAILY;INTERVAL=1
CATEGORIES:label1,label2
LAST-MODIFIED:20181201T011205Z
BEGIN:VALARM
TRIGGER;VALUE=DATE-TIME:20181201T011210Z
ACTION:DISPLAY
DESCRIPTION:Task 1
END:VALARM
BEGIN:VALARM
TRIGGER;RELATED=END:-PT1H0M0S
ACTION:DISPLAY
DESCRIPTION:Task 1
END:VALARM
END:VTODO
END:VCALENDAR`,
		},
		{
			name: "Format Task with Related Tasks as CalDAV",
			args: args{
				list: &models.ProjectWithTasksAndBuckets{
					Project: models.Project{
						Title: "List title",
					},
				},
				tasks: []*models.TaskWithComments{
					{
						Task: models.Task{
							Title:       "Parent Task",
							UID:         "randomuid_parent",
							Description: "A parent task",
							Priority:    3,
							Created:     time.Unix(1543626721, 0).In(config.GetTimeZone()),
							Updated:     time.Unix(1543626725, 0).In(config.GetTimeZone()),
							RelatedTasks: map[models.RelationKind][]*models.Task{
								models.RelationKindSubtask: {
									{
										Title:       "Subtask 1",
										UID:         "randomuid_child_1",
										Description: "The first child task",
										Created:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
										Updated:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
									},
									{
										Title:       "Subtask 2",
										UID:         "randomuid_child_2",
										Description: "The second child task",
										Created:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
										Updated:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
									},
								},
							},
						},
					},
					{
						Task: models.Task{
							Title:       "Subtask 1",
							UID:         "randomuid_child_1",
							Description: "The first child task",
							Created:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
							Updated:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
							RelatedTasks: map[models.RelationKind][]*models.Task{
								models.RelationKindParenttask: {
									{
										Title:       "Parent task",
										UID:         "randomuid_parent",
										Description: "A parent task",
										Priority:    3,
										Created:     time.Unix(1543626721, 0).In(config.GetTimeZone()),
										Updated:     time.Unix(1543626725, 0).In(config.GetTimeZone()),
									},
								},
							},
						},
					},
					{
						Task: models.Task{
							Title:       "Subtask 2",
							UID:         "randomuid_child_2",
							Description: "The second child task",
							Created:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
							Updated:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
							RelatedTasks: map[models.RelationKind][]*models.Task{
								models.RelationKindParenttask: {
									{
										Title:       "Parent task",
										UID:         "randomuid_parent",
										Description: "A parent task",
										Priority:    3,
										Created:     time.Unix(1543626721, 0).In(config.GetTimeZone()),
										Updated:     time.Unix(1543626725, 0).In(config.GetTimeZone()),
									},
								},
							},
						},
					},
				},
			},
			wantCaldav: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:List title
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:randomuid_parent
DTSTAMP:20181201T011205Z
SUMMARY:Parent Task
DESCRIPTION:A parent task
CREATED:20181201T011201Z
PRIORITY:3
LAST-MODIFIED:20181201T011205Z
RELATED-TO;RELTYPE=CHILD:randomuid_child_1
RELATED-TO;RELTYPE=CHILD:randomuid_child_2
END:VTODO
BEGIN:VTODO
UID:randomuid_child_1
DTSTAMP:20181201T011204Z
SUMMARY:Subtask 1
DESCRIPTION:The first child task
CREATED:20181201T011204Z
LAST-MODIFIED:20181201T011204Z
RELATED-TO;RELTYPE=PARENT:randomuid_parent
END:VTODO
BEGIN:VTODO
UID:randomuid_child_2
DTSTAMP:20181201T011204Z
SUMMARY:Subtask 2
DESCRIPTION:The second child task
CREATED:20181201T011204Z
LAST-MODIFIED:20181201T011204Z
RELATED-TO;RELTYPE=PARENT:randomuid_parent
END:VTODO
END:VCALENDAR`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCaldavTodosForTasks(tt.args.list, tt.args.tasks)
			if diff, equal := messagediff.PrettyDiff(got, tt.wantCaldav); !equal {
				t.Errorf("GetCaldavTodosForTasks() gotVTask = %v, want %v, diff = %s", got, tt.wantCaldav, diff)
			}
		})
	}
}
