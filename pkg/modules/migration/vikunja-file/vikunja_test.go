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
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVikunjaFileMigrator_Migrate(t *testing.T) {
	t.Run("migrate successfully", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		m := &FileMigrator{}
		u := &user.User{ID: 1}

		f, err := os.Open("export.zip")
		if err != nil {
			t.Fatalf("Could not open file: %s", err)
		}
		defer f.Close()
		s, err := f.Stat()
		if err != nil {
			t.Fatalf("Could not stat file: %s", err)
		}

		err = m.Migrate(u, f, s.Size())
		require.NoError(t, err)
		db.AssertExists(t, "projects", map[string]interface{}{
			"title":    "test project",
			"owner_id": u.ID,
		}, false)
		db.AssertExists(t, "projects", map[string]interface{}{
			"title":    "Inbox",
			"owner_id": u.ID,
		}, false)
		db.AssertExists(t, "tasks", map[string]interface{}{
			"title":         "some other task",
			"created_by_id": u.ID,
		}, false)
		db.AssertExists(t, "task_comments", map[string]interface{}{
			"comment":   "This is a comment",
			"author_id": u.ID,
		}, false)
		db.AssertExists(t, "files", map[string]interface{}{
			"name":          "grant-whitty-546453-unsplash.jpg",
			"created_by_id": u.ID,
		}, false)
		db.AssertExists(t, "labels", map[string]interface{}{
			"title":         "test",
			"created_by_id": u.ID,
		}, false)
		db.AssertExists(t, "buckets", map[string]interface{}{
			"title":         "Test Bucket",
			"created_by_id": u.ID,
		}, false)
	})
	t.Run("rejects attachment with forged size smaller than content", func(t *testing.T) {
		// Regression: GHSA-qh78-rvg3-cv54. Zip entry carries 2 MB of
		// content but data.json claims size: 0, bypassing the 1 MB cap.
		db.LoadAndAssertFixtures(t)

		oldMax := config.FilesMaxSize.GetString()
		config.FilesMaxSize.Set("1MB")
		defer func() {
			config.FilesMaxSize.Set(oldMax)
			_ = config.SetMaxFileSizeMBytesFromString(oldMax)
		}()
		require.NoError(t, config.SetMaxFileSizeMBytesFromString("1MB"))

		payload := make([]byte, 2*1024*1024)

		dataJSON := `[{
			"id": 1,
			"title": "evil project",
			"tasks": [{
				"id": 1,
				"title": "evil task",
				"attachments": [{
					"id": 1,
					"file": {"id": 1, "name": "evil.bin", "size": 0}
				}]
			}],
			"views": []
		}]`

		var zipBuf bytes.Buffer
		zw := zip.NewWriter(&zipBuf)

		vf, err := zw.Create("VERSION")
		require.NoError(t, err)
		_, err = vf.Write([]byte("dev"))
		require.NoError(t, err)

		df, err := zw.Create("data.json")
		require.NoError(t, err)
		_, err = df.Write([]byte(dataJSON))
		require.NoError(t, err)

		ff, err := zw.Create("files/1")
		require.NoError(t, err)
		_, err = ff.Write(payload)
		require.NoError(t, err)

		require.NoError(t, zw.Close())

		m := &FileMigrator{}
		u := &user.User{ID: 1}

		reader := bytes.NewReader(zipBuf.Bytes())
		err = m.Migrate(u, reader, int64(reader.Len()))

		// create_from_structure.go skips the oversized attachment via
		// `continue`, so Migrate succeeds for the rest of the project.
		require.NoError(t, err)

		// Forged 0-size row was the pre-fix outcome; assert neither
		// size ends up persisted.
		db.AssertMissing(t, "files", map[string]interface{}{
			"name": "evil.bin",
			"size": 2 * 1024 * 1024,
		})
		db.AssertMissing(t, "files", map[string]interface{}{
			"name": "evil.bin",
			"size": 0,
		})
	})
	t.Run("dispatches events queued while importing saved filters", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		var zipBuf bytes.Buffer
		zw := zip.NewWriter(&zipBuf)
		for name, content := range map[string]string{
			"VERSION":      "dev",
			"data.json":    "[]",
			"filters.json": `[{"title":"imported filter","filters":{"filter":"done = false"}}]`,
		} {
			w, err := zw.Create(name)
			require.NoError(t, err)
			_, err = w.Write([]byte(content))
			require.NoError(t, err)
		}
		require.NoError(t, zw.Close())

		m := &FileMigrator{}
		u := &user.User{ID: 1}
		reader := bytes.NewReader(zipBuf.Bytes())

		events.ClearDispatchedEvents()
		require.NoError(t, m.Migrate(u, reader, int64(reader.Len())))

		s := db.NewSession()
		defer s.Close()
		sf := &models.SavedFilter{}
		has, err := s.Where("title = ? AND owner_id = ?", "imported filter", u.ID).Get(sf)
		require.NoError(t, err)
		require.True(t, has)
		filterViewIDs := []int64{}
		require.NoError(t, s.Table("project_views").Where("project_id = ?", -sf.ID-1).Cols("id").Find(&filterViewIDs))
		require.NotEmpty(t, filterViewIDs)

		dispatched := events.GetDispatchedEvents((&models.TaskPositionsRecalculatedEvent{}).Name())
		require.NotEmpty(t, dispatched)
		for _, e := range dispatched {
			recalculated, ok := e.(*models.TaskPositionsRecalculatedEvent)
			require.True(t, ok, "dispatched %T", e)
			require.NotEmpty(t, recalculated.NewTaskPositions)
			assert.Contains(t, filterViewIDs, recalculated.NewTaskPositions[0].ProjectViewID)
		}
	})
	t.Run("should not accept an old import", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		m := &FileMigrator{}
		u := &user.User{ID: 1}

		f, err := os.Open("export_pre_0.21.0.zip")
		if err != nil {
			t.Fatalf("Could not open file: %s", err)
		}
		defer f.Close()
		s, err := f.Stat()
		if err != nil {
			t.Fatalf("Could not stat file: %s", err)
		}

		err = m.Migrate(u, f, s.Size())
		require.Error(t, err)
		var unsupported *migration.ErrImportFromUnsupportedVersion
		require.ErrorAs(t, err, &unsupported)
	})
	t.Run("should reject a zip without a data file", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		var zipBuf bytes.Buffer
		zw := zip.NewWriter(&zipBuf)
		w, err := zw.Create("VERSION")
		require.NoError(t, err)
		_, err = w.Write([]byte("dev"))
		require.NoError(t, err)
		require.NoError(t, zw.Close())

		m := &FileMigrator{}
		u := &user.User{ID: 1}
		reader := bytes.NewReader(zipBuf.Bytes())

		err = m.Migrate(u, reader, int64(reader.Len()))
		var noDataFile *migration.ErrNoDataFileInZip
		require.ErrorAs(t, err, &noDataFile)
		assert.Equal(t, http.StatusBadRequest, noDataFile.HTTPError().HTTPCode)
		assert.Equal(t, migration.ErrCodeNoDataFileInZip, noDataFile.HTTPError().Code)
	})
}

// setStoredUserZone writes a user's stored zone in the database, as the settings page would. The
// fixtures stay untouched: every test loads them afresh.
func setStoredUserZone(t *testing.T, userID int64, zone string) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	_, err := s.ID(userID).Cols("timezone").Update(&user.User{Timezone: zone})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

// importedTask returns the most recently created task with this title, which is the one an
// import in the current test just created.
func importedTask(t *testing.T, title string) *models.Task {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	task := &models.Task{}
	has, err := s.Where("title = ? AND created_by_id = ?", title, 1).OrderBy("id DESC").Get(task)
	require.NoError(t, err)
	require.True(t, has, "task %q must have been imported", title)
	return task
}

// zipWith builds an export zip from the given entries.
func zipWith(t *testing.T, entries map[string][]byte) *bytes.Reader {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range entries {
		w, err := zw.Create(name)
		require.NoError(t, err)
		_, err = w.Write(content)
		require.NoError(t, err)
	}
	require.NoError(t, zw.Close())
	return bytes.NewReader(buf.Bytes())
}

// bundledExportEntries reads every entry of the bundled export.zip, a real pre-RRULE export.
func bundledExportEntries(t *testing.T) map[string][]byte {
	t.Helper()
	zr, err := zip.OpenReader("export.zip")
	require.NoError(t, err)
	defer zr.Close()
	entries := map[string][]byte{}
	for _, f := range zr.File {
		rc, err := f.Open()
		require.NoError(t, err)
		content, err := io.ReadAll(rc)
		require.NoError(t, err)
		require.NoError(t, rc.Close())
		entries[f.Name] = content
	}
	return entries
}

// Import prefers repeats_rrule, falls back to repeat, and converts the legacy
// repeat_after/repeat_mode keys of pre-RRULE export files.
func TestVikunjaFileMigrator_MigrateRecurrence(t *testing.T) {
	migrate := func(t *testing.T, r *bytes.Reader) {
		t.Helper()
		require.NoError(t, (&FileMigrator{}).Migrate(&user.User{ID: 1}, r, int64(r.Len())))
	}

	t.Run("legacy keys in the bundled export", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		withServiceTimeZone(t, "UTC")

		// The bundled export carries repeat_after: 0 and repeat_mode: 0, which mean "no repeat".
		migrate(t, zipWith(t, bundledExportEntries(t)))

		for _, title := range []string{"test project", "some other task"} {
			task := importedTask(t, title)
			assert.Empty(t, task.Repeats, "%s: a legacy task without a repeat imports without a rule", title)
			assert.False(t, task.RepeatsFromCurrentDate, title)
		}
	})

	t.Run("legacy keys with a repeat are converted", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		withServiceTimeZone(t, "UTC")

		// The bundled export with its two tasks given legacy repeats, so the real file shape
		// is what gets converted.
		entries := bundledExportEntries(t)
		var projects []map[string]any
		require.NoError(t, json.Unmarshal(entries["data.json"], &projects))
		patched := 0
		for _, p := range projects {
			tasks, _ := p["tasks"].([]any)
			for _, raw := range tasks {
				task := raw.(map[string]any)
				switch task["title"] {
				case "test project":
					task["repeat_after"] = 86400
					task["repeat_mode"] = 2
					patched++
				case "some other task":
					task["repeat_after"] = 0
					task["repeat_mode"] = 1
					task["due_date"] = "2026-01-31T12:00:00Z"
					patched++
				}
			}
		}
		require.Equal(t, 2, patched)
		data, err := json.Marshal(projects)
		require.NoError(t, err)
		entries["data.json"] = data

		migrate(t, zipWith(t, entries))

		daily := importedTask(t, "test project")
		assert.Equal(t, "FREQ=DAILY;INTERVAL=1", daily.Repeats)
		assert.True(t, daily.RepeatsFromCurrentDate)

		monthEnd := importedTask(t, "some other task")
		assert.Equal(t, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=28,29,30,31;BYSETPOS=-1", monthEnd.Repeats)
		assert.False(t, monthEnd.RepeatsFromCurrentDate)
	})

	// The month-end day is read in the importing user's stored zone, re-fetched by id:
	// the event-carried user (&user.User{ID: 1}) has an empty Timezone. A month-end task due
	// Jan 31 20:00 New York is Feb 1 01:00 UTC.
	for _, tc := range []struct {
		zone, want string
	}{
		{"America/New_York", "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=28,29,30,31;BYSETPOS=-1"},
		{"", "FREQ=MONTHLY;INTERVAL=1"},             // no stored zone: the server zone (UTC)
		{"Mars/Olympus", "FREQ=MONTHLY;INTERVAL=1"}, // an invalid stored zone: the server zone
	} {
		t.Run("legacy month-end in the importing user's zone "+tc.zone, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			withServiceTimeZone(t, "UTC")
			setStoredUserZone(t, 1, tc.zone)

			entries := bundledExportEntries(t)
			var projects []map[string]any
			require.NoError(t, json.Unmarshal(entries["data.json"], &projects))
			patched := 0
			for _, p := range projects {
				tasks, _ := p["tasks"].([]any)
				for _, raw := range tasks {
					task := raw.(map[string]any)
					if task["title"] == "some other task" {
						task["repeat_after"] = 0
						task["repeat_mode"] = 1
						task["due_date"] = "2026-02-01T01:00:00Z"
						patched++
					}
				}
			}
			require.Equal(t, 1, patched)
			data, err := json.Marshal(projects)
			require.NoError(t, err)
			entries["data.json"] = data

			migrate(t, zipWith(t, entries))

			assert.Equal(t, tc.want, importedTask(t, "some other task").Repeats)
		})
	}

	t.Run("repeats_rrule round-trips verbatim", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		// Shaped like an export of this version: each task carries the exact rule and its
		// structured projection, which for the Easter rule is lossy.
		data := `[{
			"id": 1,
			"title": "rrule round trip",
			"views": [],
			"tasks": [
				{"id": 11, "title": "rt easter", "repeats_rrule": "FREQ=YEARLY;BYEASTER=0",
				 "repeat": {"freq": "yearly", "interval": 1, "read_only": true}},
				{"id": 12, "title": "rt month end", "repeats_rrule": "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=28,29,30,31;BYSETPOS=-1",
				 "repeat": {"freq": "monthly", "interval": 1, "by_month_day": [28, 29, 30, 31], "by_set_pos": [-1]}},
				{"id": 13, "title": "rt count", "repeats_rrule": "FREQ=DAILY;COUNT=2;INTERVAL=1", "repeats_from_current_date": true,
				 "repeat": {"freq": "daily", "interval": 1, "count": 2}},
				{"id": 14, "title": "rt rrule beats legacy", "repeats_rrule": "FREQ=WEEKLY;INTERVAL=3",
				 "repeat_after": 86400, "repeat_mode": 0}
			]
		}]`
		migrate(t, zipWith(t, map[string][]byte{"VERSION": []byte("dev"), "data.json": []byte(data)}))

		for title, want := range map[string]string{
			"rt easter":             "FREQ=YEARLY;BYEASTER=0",
			"rt month end":          "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=28,29,30,31;BYSETPOS=-1",
			"rt count":              "FREQ=DAILY;COUNT=2;INTERVAL=1",
			"rt rrule beats legacy": "FREQ=WEEKLY;INTERVAL=3",
		} {
			assert.Equal(t, want, importedTask(t, title).Repeats, title)
		}
		assert.True(t, importedTask(t, "rt count").RepeatsFromCurrentDate)
	})

	// An import file is user-supplied: a rule that fails the creation policy, the safety
	// bound or the column length is dropped and the task imported without it; the import goes on.
	t.Run("a rule failing the policy, the safety bound or the length is dropped, not fatal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		yearDays := make([]string, 200)
		for i := range yearDays {
			yearDays[i] = strconv.Itoa(i + 1)
		}
		tooLong := "FREQ=YEARLY;BYYEARDAY=" + strings.Join(yearDays, ",")
		require.Greater(t, len(tooLong), 500)

		data := `[{
			"id": 1,
			"title": "rrule drop",
			"views": [],
			"tasks": [
				{"id": 41, "title": "dr hang rrule", "repeats_rrule": "FREQ=HOURLY;INTERVAL=2;BYHOUR=1", "repeats_from_current_date": true},
				{"id": 42, "title": "dr hang repeat", "repeat": {"freq": "hourly", "interval": 2, "by_hour": [1]}},
				{"id": 43, "title": "dr bound", "repeats_rrule": "FREQ=DAILY;BYHOUR=9,9"},
				{"id": 44, "title": "dr too long", "repeats_rrule": "` + tooLong + `"},
				{"id": 45, "title": "dr unparseable", "repeats_rrule": "not a rule"},
				{"id": 46, "title": "dr kept easter", "repeats_rrule": "FREQ=YEARLY;BYEASTER=0", "repeats_from_current_date": true},
				{"id": 47, "title": "dr kept projection", "repeat": {"freq": "weekly", "interval": 2}}
			]
		}]`
		migrate(t, zipWith(t, map[string][]byte{"VERSION": []byte("dev"), "data.json": []byte(data)}))

		for _, title := range []string{"dr hang rrule", "dr hang repeat", "dr bound", "dr too long", "dr unparseable"} {
			task := importedTask(t, title)
			assert.Empty(t, task.Repeats, "%s: imported without its rule", title)
			assert.False(t, task.RepeatsFromCurrentDate, "%s: no rule, no from-current-date flag", title)
		}
		kept := importedTask(t, "dr kept easter")
		assert.Equal(t, "FREQ=YEARLY;BYEASTER=0", kept.Repeats, "a valid rule still round-trips verbatim")
		assert.True(t, kept.RepeatsFromCurrentDate)
		assert.Equal(t, "FREQ=WEEKLY;INTERVAL=2", importedTask(t, "dr kept projection").Repeats)
	})

	t.Run("repeat is the fallback without repeats_rrule", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		data := `[{
			"id": 1,
			"title": "rrule fallback",
			"views": [],
			"tasks": [
				{"id": 21, "title": "fb projection", "repeat": {"freq": "weekly", "interval": 2}},
				{"id": 22, "title": "fb projection beats legacy", "repeat": {"freq": "daily", "interval": 3},
				 "repeat_after": 3600, "repeat_mode": 0},
				{"id": 23, "title": "fb nothing"}
			]
		}]`
		migrate(t, zipWith(t, map[string][]byte{"VERSION": []byte("dev"), "data.json": []byte(data)}))

		assert.Equal(t, "FREQ=WEEKLY;INTERVAL=2", importedTask(t, "fb projection").Repeats)
		assert.Equal(t, "FREQ=DAILY;INTERVAL=3", importedTask(t, "fb projection beats legacy").Repeats)
		assert.Empty(t, importedTask(t, "fb nothing").Repeats)
	})

	t.Run("legacy keys in a sub-project are converted", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		withServiceTimeZone(t, "UTC")

		// Exports list every project at the top level; a sub-project names its parent.
		data := `[
			{"id": 1, "title": "legacy parent", "views": [],
			 "tasks": [{"id": 31, "title": "lg parent task", "repeat_after": 604800, "repeat_mode": 0}]},
			{"id": 2, "title": "legacy child", "parent_project_id": 1, "views": [],
			 "tasks": [{"id": 32, "title": "lg child task", "repeat_after": 7200, "repeat_mode": 2}]}
		]`
		migrate(t, zipWith(t, map[string][]byte{"VERSION": []byte("dev"), "data.json": []byte(data)}))

		assert.Equal(t, "FREQ=WEEKLY;INTERVAL=1", importedTask(t, "lg parent task").Repeats)
		child := importedTask(t, "lg child task")
		assert.Equal(t, "FREQ=HOURLY;INTERVAL=2", child.Repeats)
		assert.True(t, child.RepeatsFromCurrentDate)
	})
}

// captureLog sends the standard log to a file for the rest of the test and returns a reader for it.
func captureLog(t *testing.T) func() string {
	t.Helper()
	dir := t.TempDir()
	log.ConfigureStandardLogger(true, "file", dir, "INFO", "text")
	t.Cleanup(log.InitLogger)
	return func() string {
		b, err := os.ReadFile(filepath.Join(dir, "standard.log"))
		require.NoError(t, err)
		return string(b)
	}
}

// storedIntendedDueTime reads the imported task's intended_due_time column.
func storedIntendedDueTime(t *testing.T, title string) *int {
	t.Helper()
	return importedTask(t, title).IntendedDueTime
}

// Import restores the export's intended_due_time when it is a time of day
// (0-86399) on a task with a due date, and drops it otherwise with one log line, never failing.
// The import half of the round trip: the daily task exported while gap-shifted (Sun 2027-03-14 03:30
// EDT, 9000) keeps 9000; the models export test completes it.
func TestVikunjaFileMigrator_MigrateIntendedDueTime(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	setStoredUserZone(t, 1, "America/New_York")
	readLog := captureLog(t)

	shifted := time.Date(2027, 3, 14, 7, 30, 0, 0, time.UTC)
	intended := func(v int) *int { return &v }
	task := func(id int64, title string, due time.Time, v *int) *models.TaskWithComments {
		return &models.TaskWithComments{
			Task:            models.Task{ID: id, Title: title, DueDate: due},
			RepeatsRRule:    "FREQ=DAILY;INTERVAL=1",
			IntendedDueTime: v,
		}
	}
	projects := []*models.ProjectWithTasksAndBuckets{{
		Project: models.Project{ID: 1, Title: "intended due time"},
		Tasks: []*models.TaskWithComments{
			task(51, "it shifted", shifted, intended(9000)),
			task(52, "it midnight", shifted, intended(0)),
			task(53, "it last second", shifted, intended(86399)),
			task(54, "it a day", shifted, intended(86400)),
			task(55, "it too large", shifted, intended(90000)),
			task(56, "it negative", shifted, intended(-1)),
			task(57, "it no due date", time.Time{}, intended(9000)),
			task(58, "it none", shifted, nil),
		},
	}}
	data, err := json.Marshal(projects)
	require.NoError(t, err)

	r := zipWith(t, map[string][]byte{"VERSION": []byte("dev"), "data.json": data})
	require.NoError(t, (&FileMigrator{}).Migrate(&user.User{ID: 1}, r, int64(r.Len())))

	for title, want := range map[string]int{"it shifted": 9000, "it midnight": 0, "it last second": 86399} {
		got := storedIntendedDueTime(t, title)
		require.NotNil(t, got, title)
		assert.Equal(t, want, *got, title)
	}
	for _, title := range []string{"it a day", "it too large", "it negative", "it no due date", "it none"} {
		assert.Nil(t, storedIntendedDueTime(t, title), title)
	}
	assert.Equal(t, "FREQ=DAILY;INTERVAL=1", importedTask(t, "it shifted").Repeats)

	logged := readLog()
	for _, id := range []string{"54", "55", "56", "57"} {
		assert.Equal(t, 1, strings.Count(logged, "Importing task "+id+" of project 1 without its intended due time"), "one drop line for task %s", id)
	}
	assert.Equal(t, 4, strings.Count(logged, "without its intended due time"), "one line per dropped value, none for kept or absent ones")
}
