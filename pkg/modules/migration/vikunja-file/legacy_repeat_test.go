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

package vikunjafile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teambition/rrule-go"
)

// legacyRepeatCase is one row of the fixture shared with migration 20261022220000's frozen
// converter. Both converters are tested against it, so they agree today and can diverge
// deliberately later.
type legacyRepeatCase struct {
	Name                string  `json:"name"`
	RepeatAfter         int64   `json:"repeat_after"`
	RepeatMode          int     `json:"repeat_mode"`
	DueDate             *string `json:"due_date"`
	Timezone            string  `json:"timezone"`
	WantRule            string  `json:"want_rule"`
	WantFromCurrentDate bool    `json:"want_from_current_date"`
}

func loadLegacyRepeatCases(t *testing.T) []legacyRepeatCase {
	t.Helper()
	_, here, _, ok := runtime.Caller(0)
	require.True(t, ok)
	path := filepath.Join(filepath.Dir(here), "..", "..", "..", "migration", "testdata", "legacy_repeat_cases.json")
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	var cases []legacyRepeatCase
	require.NoError(t, json.Unmarshal(raw, &cases))
	require.NotEmpty(t, cases)
	return cases
}

func withServiceTimeZone(t *testing.T, zone string) {
	t.Helper()
	if zone == "" {
		zone = "UTC"
	}
	prev := config.ServiceTimeZone.GetString()
	config.ServiceTimeZone.Set(zone)
	t.Cleanup(func() { config.ServiceTimeZone.Set(prev) })
}

// The converter reads the due day in the zone it is given (the importing user's), never in
// the server zone, which is set to a different zone here on purpose.
func TestConvertLegacyRepeat(t *testing.T) {
	withServiceTimeZone(t, "Asia/Tokyo")
	for _, c := range loadLegacyRepeatCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			zone := c.Timezone
			if zone == "" {
				zone = "UTC"
			}
			loc, err := time.LoadLocation(zone)
			require.NoError(t, err)

			var due time.Time
			if c.DueDate != nil {
				due, err = time.Parse(time.RFC3339, *c.DueDate)
				require.NoError(t, err)
			}

			rule, fromCurrentDate := convertLegacyRepeat(c.RepeatAfter, c.RepeatMode, due, loc)
			assert.Equal(t, c.WantRule, rule)
			assert.Equal(t, c.WantFromCurrentDate, fromCurrentDate)
			if rule != "" {
				_, err := rrule.StrToRRule(rule)
				require.NoError(t, err, "emitted rule must parse")
			}
		})
	}
}

func TestApplyImportedRecurrence(t *testing.T) {
	withServiceTimeZone(t, "UTC")

	// Task ids repeat across the projects on purpose: tasks are matched to their own entry in
	// the file, never by id alone.
	data := []byte(`[
		{"id": 1, "tasks": [
			{"id": 7, "repeats_rrule": "FREQ=YEARLY;BYEASTER=0", "repeat": {"freq": "yearly", "interval": 1}, "repeat_after": 3600},
			{"id": 8, "repeat": {"freq": "daily", "interval": 2}, "repeat_after": 3600},
			{"id": 9, "repeat_after": 86400, "repeat_mode": 2},
			{"id": 10, "repeat_after": 0, "repeat_mode": 0}
		],
		 "child_projects": [{"id": 2, "tasks": [{"id": 7, "repeat_mode": 1, "due_date": "2026-01-30T12:00:00Z"}]}]},
		{"id": 3, "tasks": [{"id": 7, "repeat_after": 1209600}]}
	]`)
	var projects []*models.ProjectWithTasksAndBuckets
	require.NoError(t, json.Unmarshal(data, &projects))
	require.NoError(t, applyImportedRecurrence(projects, data, time.UTC))

	tasks := projects[0].Tasks
	assert.Equal(t, "FREQ=YEARLY;BYEASTER=0", tasks[0].Repeats, "repeats_rrule wins")
	assert.Nil(t, tasks[0].Repeat, "the projection is dropped so the text is stored verbatim")

	assert.Empty(t, tasks[1].Repeats, "a structured repeat is left for the create path")
	require.NotNil(t, tasks[1].Repeat)
	assert.Equal(t, 2, tasks[1].Repeat.Interval)

	assert.Equal(t, "FREQ=DAILY;INTERVAL=1", tasks[2].Repeats)
	assert.True(t, tasks[2].RepeatsFromCurrentDate)

	assert.Empty(t, tasks[3].Repeats)
	assert.False(t, tasks[3].RepeatsFromCurrentDate)

	child := projects[0].ChildProjects[0].Tasks[0]
	assert.Equal(t, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=28,29,30;BYSETPOS=-1", child.Repeats)

	assert.Equal(t, "FREQ=WEEKLY;INTERVAL=2", projects[1].Tasks[0].Repeats)
}

func TestApplyImportedRecurrenceRejectsMalformedLegacyKeys(t *testing.T) {
	data := []byte(`[{"id": 1, "tasks": [{"id": 1, "repeat_after": "daily"}]}]`)
	var projects []*models.ProjectWithTasksAndBuckets
	require.NoError(t, json.Unmarshal(data, &projects))
	require.Error(t, applyImportedRecurrence(projects, data, time.UTC))
}
