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
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func vtodoWithLine(line string) string {
	return "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//test//EN\r\nBEGIN:VTODO\r\nUID:limits\r\nSUMMARY:Limits\r\n" +
		line + "\r\nEND:VTODO\r\nEND:VCALENDAR\r\n"
}

func paramList(n int, name func(i int) string) string {
	var b strings.Builder
	for i := range n {
		b.WriteString(";" + name(i) + "=v")
	}
	return b.String()
}

func plainParamName(i int) string { return "P" + strconv.Itoa(i) }

func TestCheckVCalendarLimits(t *testing.T) {
	t.Run("accepts the maximum number of parameters", func(t *testing.T) {
		content := vtodoWithLine("X-TEST" + paramList(MaxPropertyParameters, plainParamName) + ":v")
		require.NoError(t, CheckVCalendarLimits(content))

		_, _, err := ParseTaskFromVTODO(content, time.UTC)
		require.NoError(t, err)
	})

	t.Run("rejects too many parameters", func(t *testing.T) {
		content := vtodoWithLine("X-TEST" + paramList(MaxPropertyParameters+1, plainParamName) + ":v")
		require.ErrorIs(t, CheckVCalendarLimits(content), ErrTooManyPropertyParameters)

		_, _, err := ParseTaskFromVTODO(content, time.UTC)
		require.ErrorIs(t, err, ErrTooManyPropertyParameters)
	})

	t.Run("counts comma separated parameter values", func(t *testing.T) {
		values := strings.TrimSuffix(strings.Repeat("v,", MaxPropertyParameters+1), ",")
		content := vtodoWithLine("X-TEST;P=" + values + ":v")
		require.ErrorIs(t, CheckVCalendarLimits(content), ErrTooManyPropertyParameters)
	})

	t.Run("counts parameters across folded lines", func(t *testing.T) {
		content := vtodoWithLine("X-TEST" + strings.ReplaceAll(paramList(MaxPropertyParameters+1, plainParamName), ";", "\r\n ;") + ":v")
		require.ErrorIs(t, CheckVCalendarLimits(content), ErrTooManyPropertyParameters)
	})

	t.Run("rejects parameter names the parser skips ahead to", func(t *testing.T) {
		content := vtodoWithLine("X-TEST" + paramList(MaxPropertyParameters+1, func(i int) string { return ":" + plainParamName(i) }) + ":v")
		require.ErrorIs(t, CheckVCalendarLimits(content), ErrTooManyPropertyParameters)
	})

	t.Run("ignores separators in quoted parameter values and in the value", func(t *testing.T) {
		quoted := `"` + strings.Repeat("a;b,c:", MaxPropertyParameters) + `"`
		description := strings.Repeat(`text\, more\; `, MaxPropertyParameters*2)
		content := vtodoWithLine("DESCRIPTION;X-NOTE=" + quoted + ":" + description)
		require.NoError(t, CheckVCalendarLimits(content))

		task, _, err := ParseTaskFromVTODO(content, time.UTC)
		require.NoError(t, err)
		assert.Contains(t, task.Description, "text, more; ")
	})

	t.Run("rejects oversized calendar data", func(t *testing.T) {
		content := vtodoWithLine("DESCRIPTION:" + strings.Repeat("a", MaxVCalendarSize))
		require.ErrorIs(t, CheckVCalendarLimits(content), ErrVCalendarTooLarge)

		_, _, err := ParseTaskFromVTODO(content, time.UTC)
		require.ErrorIs(t, err, ErrVCalendarTooLarge)
	})
}
