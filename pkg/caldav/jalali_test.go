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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTodosJalaliOmitsRRULE(t *testing.T) {
	cfg := &Config{Name: "test", ProdID: "RandomProdID which is not random"}
	ts := time.Unix(1543626724, 0).In(config.GetTimeZone())
	due := time.Unix(1543626724, 0).In(config.GetTimeZone())

	tests := []struct {
		name string
		todo *Todo
	}{
		{
			name: "jalali month",
			todo: &Todo{
				Summary: "Todo #1", UID: "jalali-month", Timestamp: ts,
				RepeatMode: models.TaskRepeatModeJalaliMonth, DueDate: due,
			},
		},
		{
			name: "jalali year",
			todo: &Todo{
				Summary: "Todo #1", UID: "jalali-year", Timestamp: ts,
				RepeatMode: models.TaskRepeatModeJalaliYear, DueDate: due,
			},
		},
		{
			name: "jalali month with repeat_after still omits rrule",
			todo: &Todo{
				Summary: "Todo #1", UID: "jalali-month-interval", Timestamp: ts,
				RepeatMode: models.TaskRepeatModeJalaliMonth, DueDate: due, RepeatAfter: 86400,
			},
		},
		{
			name: "jalali year with repeat_after still omits rrule",
			todo: &Todo{
				Summary: "Todo #1", UID: "jalali-year-interval", Timestamp: ts,
				RepeatMode: models.TaskRepeatModeJalaliYear, DueDate: due, RepeatAfter: 86400,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := ParseTodos(cfg, []*Todo{tt.todo})
			assert.NotContains(t, out, "RRULE")
			assert.Contains(t, out, "BEGIN:VTODO")
			assert.Contains(t, out, "STATUS:NEEDS-ACTION")
			assert.Contains(t, out, "DUE:")
		})
	}

	t.Run("gregorian monthly still emits rrule", func(t *testing.T) {
		out := ParseTodos(cfg, []*Todo{{
			Summary: "Todo #1", UID: "greg-month", Timestamp: ts,
			RepeatMode: models.TaskRepeatModeMonth, DueDate: due,
		}})
		assert.Contains(t, out, "RRULE:FREQ=MONTHLY;BYMONTHDAY=")
	})
}

func TestParseJalaliVTODORoundtrip(t *testing.T) {
	// Inbound parsing ignores RRULE, so a Jalali VTODO without one parses like any other task.
	cfg := &Config{Name: "test", ProdID: "RandomProdID which is not random"}
	ts := time.Unix(1543626724, 0).In(config.GetTimeZone())
	out := ParseTodos(cfg, []*Todo{{
		Summary: "Jalali task", UID: "jalali-uid", Timestamp: ts,
		RepeatMode: models.TaskRepeatModeJalaliMonth,
		DueDate:    time.Unix(1543626724, 0).In(config.GetTimeZone()),
	}})

	vTask, _, err := ParseTaskFromVTODO(out)
	require.NoError(t, err)
	assert.Equal(t, "Jalali task", vTask.Title)
	assert.Equal(t, "jalali-uid", vTask.UID)
	assert.False(t, vTask.DueDate.IsZero())
}

// P11: Jalali repeats stay RFC5545 Gregorian/Zulu on the wire.
func TestParseTodosJalaliGregorianZuluWire(t *testing.T) {
	cfg := &Config{Name: "test", ProdID: "RandomProdID which is not random"}
	ts := time.Unix(1543626724, 0).In(config.GetTimeZone())
	tehran, err := time.LoadLocation("Asia/Tehran")
	require.NoError(t, err)
	// 1405/06/16 00:00 Asia/Tehran wall-clock.
	dueTehran := time.Date(2026, 9, 7, 0, 0, 0, 0, tehran)
	startTehran := time.Date(2026, 9, 6, 0, 0, 0, 0, tehran)
	require.True(t, dueTehran.In(time.UTC).Equal(time.Date(2026, 9, 6, 20, 30, 0, 0, time.UTC)))
	require.True(t, startTehran.In(time.UTC).Equal(time.Date(2026, 9, 5, 20, 30, 0, 0, time.UTC)))

	modes := []struct {
		name string
		mode models.TaskRepeatMode
	}{
		{"jalali month mode 3", models.TaskRepeatModeJalaliMonth},
		{"jalali year mode 4", models.TaskRepeatModeJalaliYear},
	}
	for _, m := range modes {
		t.Run(m.name, func(t *testing.T) {
			out := ParseTodos(cfg, []*Todo{
				{
					Summary: "Jalali wire", UID: "jalali-wire", Timestamp: ts,
					RepeatMode: m.mode, DueDate: dueTehran, Start: startTehran,
				},
			})
			assert.Contains(t, out, "DUE:20260906T203000Z")
			assert.Contains(t, out, "DTSTART:20260905T203000Z")
			assert.NotContains(t, out, "RRULE")
			assert.NotContains(t, out, "۱۴۰")
		})
	}

	t.Run("absolute valarm trigger stays zulu", func(t *testing.T) {
		out := ParseTodos(cfg, []*Todo{
			{
				Summary: "Jalali alarm", UID: "jalali-alarm", Timestamp: ts,
				RepeatMode: models.TaskRepeatModeJalaliMonth, DueDate: dueTehran,
				Alarms: []Alarm{{Time: dueTehran}},
			},
		})
		assert.Contains(t, out, "TRIGGER;VALUE=DATE-TIME:20260906T203000Z")
		assert.NotContains(t, out, "RRULE")
	})
}

// P11: inbound VTODO preserves Gregorian instants; RRULE never parsed.
func TestParseJalaliVTODOInboundPreservesGregorian(t *testing.T) {
	cfg := &Config{Name: "test", ProdID: "RandomProdID which is not random"}
	ts := time.Unix(1543626724, 0).In(config.GetTimeZone())
	tehran, err := time.LoadLocation("Asia/Tehran")
	require.NoError(t, err)
	dueTehran := time.Date(2026, 9, 7, 0, 0, 0, 0, tehran)
	startTehran := time.Date(2026, 9, 6, 0, 0, 0, 0, tehran)
	alarmTehran := dueTehran

	out := ParseTodos(cfg, []*Todo{
		{
			Summary: "Jalali inbound", UID: "jalali-inbound", Timestamp: ts,
			RepeatMode: models.TaskRepeatModeJalaliMonth, DueDate: dueTehran, Start: startTehran,
			Alarms: []Alarm{{Time: alarmTehran}},
		},
	})
	vTask, _, err := ParseTaskFromVTODO(out)
	require.NoError(t, err)
	assert.True(t, vTask.DueDate.Equal(dueTehran))
	assert.True(t, vTask.StartDate.Equal(startTehran))
	require.Len(t, vTask.Reminders, 1)
	assert.True(t, vTask.Reminders[0].Reminder.Equal(alarmTehran))

	t.Run("tzid due parses to same instant", func(t *testing.T) {
		content := "BEGIN:VCALENDAR\nVERSION:2.0\nPRODID:-//x//EN\nBEGIN:VTODO\nUID:tzid-uid\nDTSTAMP:20181201T011204Z\nSUMMARY:TZID\nDUE;TZID=Asia/Tehran:20260907T000000\nEND:VTODO\nEND:VCALENDAR"
		got, _, err := ParseTaskFromVTODO(content)
		require.NoError(t, err)
		assert.True(t, got.DueDate.Equal(dueTehran))
	})

	t.Run("inbound rrule ignored", func(t *testing.T) {
		content := "BEGIN:VCALENDAR\nVERSION:2.0\nPRODID:-//x//EN\nBEGIN:VTODO\nUID:rrule-uid\nDTSTAMP:20181201T011204Z\nSUMMARY:RRULE\nDUE:20260906T203000Z\nRRULE:FREQ=MONTHLY;BYMONTHDAY=07\nEND:VTODO\nEND:VCALENDAR"
		got, _, err := ParseTaskFromVTODO(content)
		require.NoError(t, err)
		assert.True(t, got.DueDate.Equal(dueTehran))
	})
}
