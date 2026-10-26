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

package models

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teambition/rrule-go"
)

// The frontend preview reads the same file and must agree with these chains.
const repeatPreviewMatrixPath = "testdata/repeat_preview_matrix.json"

type repeatPreviewStep struct {
	Due     time.Time `json:"due"`
	Count   *int      `json:"count"`
	Cleared bool      `json:"cleared"`
	Done    bool      `json:"done"`
}

type repeatPreviewRow struct {
	Name            string              `json:"name"`
	UserZone        string              `json:"user_zone"`
	ServerZone      string              `json:"server_zone"`
	Repeat          TaskRepeat          `json:"repeat"`
	Due             time.Time           `json:"due"`
	Completions     int                 `json:"completions"`
	PreviewEligible bool                `json:"preview_eligible"`
	Summary         string              `json:"summary"`
	Expect          []repeatPreviewStep `json:"expect"`
}

type repeatPreviewMatrix struct {
	Version int                `json:"version"`
	Rows    []repeatPreviewRow `json:"rows"`
}

func loadRepeatPreviewMatrix(t *testing.T) repeatPreviewMatrix {
	t.Helper()
	raw, err := os.ReadFile(repeatPreviewMatrixPath)
	require.NoError(t, err)
	var m repeatPreviewMatrix
	// time.Time only accepts RFC 3339 with an offset, so every due is an unambiguous instant.
	require.NoError(t, json.Unmarshal(raw, &m))
	require.Equal(t, 1, m.Version)
	return m
}

func remainingCount(t *testing.T, repeats string) *int {
	t.Helper()
	if repeats == "" {
		return nil
	}
	opt, err := rrule.StrToROption(repeats)
	require.NoError(t, err)
	if opt.Count == 0 {
		return nil
	}
	return &opt.Count
}

func TestRepeatPreviewMatrix(t *testing.T) {
	m := loadRepeatPreviewMatrix(t)

	// The chains search from the due date, never from the clock, only while every due is far ahead.
	horizon := time.Now().AddDate(50, 0, 0)
	eligible := 0
	for _, row := range m.Rows {
		if row.PreviewEligible {
			eligible++
		}
		assert.True(t, row.Due.After(horizon), "%s: due %v is under 50 years ahead", row.Name, row.Due)
		for i, step := range row.Expect {
			assert.True(t, step.Due.After(horizon), "%s: step %d due %v is under 50 years ahead", row.Name, i+1, step.Due)
		}
		assert.NotEmpty(t, row.Summary, row.Name)
		if n := len(row.Expect); n != row.Completions {
			assert.True(t, n > 0 && n < row.Completions && row.Expect[n-1].Done, "%s: %d expected steps for %d completions, and the last does not end the chain", row.Name, n, row.Completions)
		}
	}
	assert.GreaterOrEqual(t, eligible, 27, "preview-eligible rows")

	origZone := config.ServiceTimeZone.GetString()
	t.Cleanup(func() { config.ServiceTimeZone.Set(origZone) })

	for _, row := range m.Rows {
		t.Run(row.Name, func(t *testing.T) {
			userLoc := mustZone(t, row.UserZone)
			config.ServiceTimeZone.Set(row.ServerZone)
			serverLoc := config.GetTimeZone()
			require.Equal(t, row.ServerZone, serverLoc.String())

			rule, err := row.Repeat.toRRule()
			require.NoError(t, err)
			require.NotEmpty(t, rule)

			// As loaded from the database: the instant read in the server's zone.
			old := &Task{Repeats: rule, DueDate: row.Due.In(serverLoc)}
			var got []repeatPreviewStep
			for i := 0; i < row.Completions; i++ {
				withFixedNow(t, old.DueDate.Add(time.Second))
				// The write path hands completion a copy of the stored task with done set.
				nw := &Task{
					Done:            true,
					Repeats:         old.Repeats,
					DueDate:         old.DueDate,
					StartDate:       old.StartDate,
					EndDate:         old.EndDate,
					IntendedDueTime: old.IntendedDueTime,
				}
				updateDone(old, nw, userLoc)

				step := repeatPreviewStep{Due: nw.DueDate, Count: remainingCount(t, nw.Repeats), Cleared: nw.Repeats == "", Done: nw.Done}
				got = append(got, step)
				assert.Equal(t, serverLoc.String(), step.Due.Location().String(), "completion %d: due is stored in the server zone", i+1)
				if step.Cleared {
					assert.Nil(t, nw.Repeat, "completion %d: a cleared rule leaves no structured repeat", i+1)
				}
				if step.Done {
					// A task left done is not completed again.
					break
				}
				old = nw
			}
			require.Len(t, row.Expect, len(got), "one expected step per completion run (the chain stops once the task stays done)")
			for i, want := range row.Expect {
				g := got[i]
				assert.True(t, g.Due.Equal(want.Due), "completion %d: due %s, want %s", i+1, g.Due.Format(time.RFC3339), want.Due.Format(time.RFC3339))
				assert.Equal(t, want.Count, g.Count, "completion %d: remaining count", i+1)
				assert.Equal(t, want.Cleared, g.Cleared, "completion %d: rule cleared", i+1)
				assert.Equal(t, want.Done, g.Done, "completion %d: done", i+1)
			}
		})
	}
}
