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
