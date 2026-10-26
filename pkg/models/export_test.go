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
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func TestGetUserDataExportStatus(t *testing.T) {
	t.Run("no export", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		status, err := GetUserDataExportStatus(s, &user.User{ID: 15})
		require.NoError(t, err)
		assert.Nil(t, status)
	})

	t.Run("with export", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		status, err := GetUserDataExportStatus(s, &user.User{ID: 1, ExportFileID: 1})
		require.NoError(t, err)
		require.NotNil(t, status)
		assert.Equal(t, int64(1), status.ID)
	})

	t.Run("export points at a missing file", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		// A dangling ExportFileID must read as "no export" rather than erroring,
		// matching the download path which 404s the same case.
		s := db.NewSession()
		defer s.Close()
		status, err := GetUserDataExportStatus(s, &user.User{ID: 15, ExportFileID: 9999})
		require.NoError(t, err)
		assert.Nil(t, status)
	})
}

func TestCleanupOldUserExports(t *testing.T) {
	t.Run("keeps fresh exports", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		fresh := &files.File{Name: "fresh", Size: 100, Created: time.Now(), CreatedByID: 2}
		_, err := s.Insert(fresh)
		require.NoError(t, err)

		// File 1 from the fixtures is from 2019 and therefore expired.
		_, err = s.ID(1).Cols("export_file_id").Update(&user.User{ExportFileID: 1})
		require.NoError(t, err)
		_, err = s.ID(2).Cols("export_file_id").Update(&user.User{ExportFileID: fresh.ID})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		cleanupOldUserExports()

		db.AssertMissing(t, "users", map[string]interface{}{"id": 1, "export_file_id": 1})
		db.AssertMissing(t, "files", map[string]interface{}{"id": 1})
		db.AssertExists(t, "users", map[string]interface{}{"id": 2, "export_file_id": fresh.ID}, false)
		db.AssertExists(t, "files", map[string]interface{}{"id": fresh.ID}, false)
	})
}

func handleUserDataExport(t *testing.T, u *user.User) error {
	payload, err := json.Marshal(&UserDataExportRequestedEvent{User: u})
	require.NoError(t, err)
	return (&HandleUserDataExport{}).Handle(message.NewMessage("test", payload))
}

func assertBlobMissing(t *testing.T, id int64) {
	_, err := files.FileStat(&files.File{ID: id})
	assert.True(t, os.IsNotExist(err), "blob %d should be gone, stat err: %v", id, err)
}

// GHSA-7gmh-4fw6-8xc5
func TestHandleUserDataExport(t *testing.T) {
	t.Run("removes the previous export", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)
		notifications.Fake()
		t.Cleanup(notifications.Unfake)

		s := db.NewSession()
		_, err := s.ID(4).Cols("export_file_id").Update(&user.User{ExportFileID: 1})
		require.NoError(t, err)
		u, err := user.GetUserByID(s, 4)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		_ = s.Close()

		require.NoError(t, handleUserDataExport(t, u))

		db.AssertMissing(t, "users", map[string]interface{}{"id": 4, "export_file_id": 1})
		db.AssertMissing(t, "files", map[string]interface{}{"id": 1})
		assertBlobMissing(t, 1)
	})
	t.Run("removes the new blob when the export fails", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		// Empty storage, so any blob found below is the failed export's.
		files.InitTestFileHandler()
		t.Cleanup(func() { files.InitTestFileFixtures(t) })

		// Notifying a user that doesn't exist fails after the blob was stored.
		err := handleUserDataExport(t, &user.User{ID: 9999, Username: "ghost"})
		require.Error(t, err)

		// Ids aren't reused on every db, so a probe bounds the id the export got.
		s := db.NewSession()
		defer s.Close()
		probe, err := files.CreateWithSession(s, bytes.NewReader([]byte("probe")), "probe", 5, &user.User{ID: 1})
		require.NoError(t, err)
		for id := int64(1); id < probe.ID; id++ {
			assertBlobMissing(t, id)
		}
	})
}

// exportedTasks runs user userID's project and task export and returns each exported task's JSON
// object from data.json, by id.
func exportedTasks(t *testing.T, s *xorm.Session, userID int64) map[string]map[string]json.RawMessage {
	t.Helper()
	u, err := user.GetUserByID(s, userID)
	require.NoError(t, err)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	_, err = exportProjectsAndTasks(s, u, zw)
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)
	var data []byte
	for _, f := range zr.File {
		if f.Name != "data.json" {
			continue
		}
		rc, err := f.Open()
		require.NoError(t, err)
		data, err = io.ReadAll(rc)
		require.NoError(t, err)
		require.NoError(t, rc.Close())
	}
	require.NotEmpty(t, data, "the export must contain data.json")

	var projects []struct {
		Tasks []map[string]json.RawMessage `json:"tasks"`
	}
	require.NoError(t, json.Unmarshal(data, &projects))
	tasks := map[string]map[string]json.RawMessage{}
	for _, p := range projects {
		for _, task := range p.Tasks {
			tasks[string(task["id"])] = task
		}
	}
	return tasks
}

// The export carries each task's exact rule in the export-only repeats_rrule field,
// next to the structured repeat the API also returns.
func TestExportProjectsAndTasksCarriesRepeatsRRule(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	files.InitTestFileFixtures(t)
	s := db.NewSession()
	defer s.Close()

	_, err := s.ID(1).Cols("repeats").Update(&Task{Repeats: "FREQ=YEARLY;BYEASTER=0"})
	require.NoError(t, err)

	tasks := exportedTasks(t, s, 1)

	for id, want := range map[string]string{
		"28": "FREQ=HOURLY;INTERVAL=1",
		"1":  "FREQ=YEARLY;BYEASTER=0",
	} {
		task, ok := tasks[id]
		require.True(t, ok, "task %s must be exported", id)
		var got string
		require.NoError(t, json.Unmarshal(task["repeats_rrule"], &got), "task %s: repeats_rrule", id)
		assert.Equal(t, want, got, "task %s", id)
		assert.NotEqual(t, "null", string(task["repeat"]), "task %s keeps the structured repeat", id)
		assert.NotEmpty(t, task["repeat"], "task %s keeps the structured repeat", id)
	}

	// A task without a rule carries no repeats_rrule key at all.
	plain, ok := tasks["2"]
	require.True(t, ok)
	_, has := plain["repeats_rrule"]
	assert.False(t, has, "a task without a rule must not carry repeats_rrule")
}
