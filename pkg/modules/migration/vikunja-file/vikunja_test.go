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
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/require"
	"xorm.io/builder"
	"xorm.io/xorm"
)

func TestVikunjaFileMigrator_Migrate(t *testing.T) {
	t.Run("migrate successfully", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		m := &FileMigrator{}
		u := &user.User{ID: 1}

		before := captureCounts(t, u.ID)

		f, err := os.Open(config.ServiceRootpath.GetString() + "/pkg/modules/migration/vikunja-file/export.zip")
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

		after := captureCounts(t, u.ID)
		stats := loadExportStats(t)
		delta := after.subtract(before)

		require.EqualValues(t, stats.projects, delta.projects, "unexpected project count delta")
		require.EqualValues(t, stats.tasks, delta.tasks, "unexpected task count delta")
		require.EqualValues(t, stats.buckets, delta.buckets, "unexpected bucket count delta")
		require.EqualValues(t, stats.tasks, delta.taskBuckets, "unexpected task bucket count delta")
		require.EqualValues(t, stats.labels, delta.labels, "unexpected label count delta")
		require.EqualValues(t, stats.attachments, delta.files, "unexpected file count delta")
		require.EqualValues(t, stats.comments, delta.comments, "unexpected comment count delta")
		require.EqualValues(t, stats.relations, delta.relations, "unexpected task relation count delta")
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
	t.Run("should not accept an old import", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		m := &FileMigrator{}
		u := &user.User{ID: 1}

		f, err := os.Open(config.ServiceRootpath.GetString() + "/pkg/modules/migration/vikunja-file/export_pre_0.21.0.zip")
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
		require.ErrorContainsf(t, err, "export was created with an older version", "Invalid error message")
	})
}

type exportStats struct {
	projects    int
	tasks       int
	buckets     int
	labels      int
	attachments int
	comments    int
	relations   int
}

func loadExportStats(t *testing.T) exportStats {
	t.Helper()

	archivePath := filepath.Join(config.ServiceRootpath.GetString(), "pkg/modules/migration/vikunja-file/export.zip")
	r, err := zip.OpenReader(archivePath)
	require.NoError(t, err)
	defer r.Close()

	var dataFile *zip.File
	for _, f := range r.File {
		if f.Name == "data.json" {
			dataFile = f
			break
		}
	}
	require.NotNil(t, dataFile, "export archive is missing data.json")

	rc, err := dataFile.Open()
	require.NoError(t, err)
	defer rc.Close()

	projects, err := decodeExportProjects(rc)
	require.NoError(t, err)

	stats := exportStats{projects: len(projects)}
	for _, p := range projects {
		stats.tasks += len(p.Tasks)
		stats.buckets += len(p.Buckets)
		for _, task := range p.Tasks {
			stats.labels += len(task.Labels)
			stats.attachments += len(task.Attachments)
			stats.comments += len(task.Comments)
			for _, rel := range task.RelatedTasks {
				stats.relations += len(rel)
			}
		}
	}

	return stats
}

type exportedProject struct {
	Tasks   []exportedTask    `json:"tasks"`
	Buckets []json.RawMessage `json:"buckets"`
}

type exportedTask struct {
	Labels       []json.RawMessage            `json:"labels"`
	Attachments  []json.RawMessage            `json:"attachments"`
	Comments     []json.RawMessage            `json:"comments"`
	RelatedTasks map[string][]json.RawMessage `json:"related_tasks"`
}

func decodeExportProjects(r io.Reader) ([]exportedProject, error) {
	decoder := json.NewDecoder(r)
	var projects []exportedProject
	if err := decoder.Decode(&projects); err != nil {
		return nil, err
	}

	// Ensure slices and maps are non-nil for easier counting later.
	for i := range projects {
		if projects[i].Tasks == nil {
			projects[i].Tasks = []exportedTask{}
		}
		if projects[i].Buckets == nil {
			projects[i].Buckets = []json.RawMessage{}
		}
		for j := range projects[i].Tasks {
			task := &projects[i].Tasks[j]
			if task.Labels == nil {
				task.Labels = []json.RawMessage{}
			}
			if task.Attachments == nil {
				task.Attachments = []json.RawMessage{}
			}
			if task.Comments == nil {
				task.Comments = []json.RawMessage{}
			}
			if task.RelatedTasks == nil {
				task.RelatedTasks = map[string][]json.RawMessage{}
			}
		}
	}

	return projects, nil
}

type dbCounts struct {
	projects    int64
	tasks       int64
	buckets     int64
	taskBuckets int64
	labels      int64
	files       int64
	comments    int64
	relations   int64
}

func (after dbCounts) subtract(before dbCounts) dbCounts {
	return dbCounts{
		projects:    after.projects - before.projects,
		tasks:       after.tasks - before.tasks,
		buckets:     after.buckets - before.buckets,
		taskBuckets: after.taskBuckets - before.taskBuckets,
		labels:      after.labels - before.labels,
		files:       after.files - before.files,
		comments:    after.comments - before.comments,
		relations:   after.relations - before.relations,
	}
}

func captureCounts(t *testing.T, userID int64) dbCounts {
	t.Helper()
	engine := db.GetEngine()

	counts := dbCounts{}
	counts.projects = countWhere(t, engine, "projects", builder.Eq{"owner_id": userID})
	counts.tasks = countWhere(t, engine, "tasks", builder.Eq{"created_by_id": userID})
	counts.labels = countWhere(t, engine, "labels", builder.Eq{"created_by_id": userID})
	counts.buckets = countWhere(t, engine, "buckets", builder.Eq{"created_by_id": userID})
	counts.files = countWhere(t, engine, "files", builder.Eq{"created_by_id": userID})
	counts.comments = countWhere(t, engine, "task_comments", builder.Eq{"author_id": userID})

	taskIDs := fetchTaskIDs(t, engine, userID)
	if len(taskIDs) > 0 {
		counts.taskBuckets = countWhere(t, engine, "task_buckets", builder.In("task_id", toInterfaces(taskIDs)...))
		taskCond := builder.In("task_id", toInterfaces(taskIDs)...)
		otherCond := builder.In("other_task_id", toInterfaces(taskIDs)...)
		counts.relations = countWhere(t, engine, "task_relations", builder.Or(taskCond, otherCond))
	}

	return counts
}

func countWhere(t *testing.T, engine *xorm.Engine, table string, cond builder.Cond) int64 {
	t.Helper()
	cnt, err := engine.Table(table).Where(cond).Count()
	require.NoErrorf(t, err, "failed to count table %s", table)
	return cnt
}

func fetchTaskIDs(t *testing.T, engine *xorm.Engine, userID int64) []int64 {
	t.Helper()
	ids := []int64{}
	err := engine.Table("tasks").Where("created_by_id = ?", userID).Cols("id").Find(&ids)
	require.NoError(t, err)
	return ids
}

func toInterfaces(ints []int64) []interface{} {
	res := make([]interface{}, len(ints))
	for i, v := range ints {
		res[i] = v
	}
	return res
}
