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

package migration

import (
	"path/filepath"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

func TestConvertLegacyRepeatToRRule(t *testing.T) {
	const (
		modeDefault         = 0
		modeMonth           = 1
		modeFromCurrentDate = 2
		modeYear            = 3
	)

	cases := []struct {
		name        string
		repeatAfter int64
		repeatMode  int
		want        string
	}{
		{"default daily", 86400, modeDefault, "FREQ=DAILY;INTERVAL=1"},
		{"default every 2 days", 172800, modeDefault, "FREQ=DAILY;INTERVAL=2"},
		{"default weekly", 604800, modeDefault, "FREQ=WEEKLY;INTERVAL=1"},
		{"default hourly", 3600, modeDefault, "FREQ=HOURLY;INTERVAL=1"},
		{"default minutely", 60, modeDefault, "FREQ=MINUTELY;INTERVAL=1"},
		{"default secondly remainder", 90, modeDefault, "FREQ=SECONDLY;INTERVAL=90"},
		{"default no interval is empty", 0, modeDefault, ""},
		{"from current date keeps the interval", 86400, modeFromCurrentDate, "FREQ=DAILY;INTERVAL=1"},
		{"from current date no interval is empty", 0, modeFromCurrentDate, ""},
		{"monthly ignores repeat_after", 86400, modeMonth, "FREQ=MONTHLY;INTERVAL=1"},
		{"monthly without interval", 0, modeMonth, "FREQ=MONTHLY;INTERVAL=1"},
		{"yearly ignores repeat_after", 86400, modeYear, "FREQ=YEARLY;INTERVAL=1"},
		{"unknown mode is empty", 86400, 99, ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, convertLegacyRepeatToRRule(c.repeatAfter, c.repeatMode))
		})
	}
}

func TestSecondsToRRule(t *testing.T) {
	cases := []struct {
		seconds int64
		want    string
	}{
		{604800, "FREQ=WEEKLY;INTERVAL=1"},
		{1209600, "FREQ=WEEKLY;INTERVAL=2"},
		{86400, "FREQ=DAILY;INTERVAL=1"},
		{3600, "FREQ=HOURLY;INTERVAL=1"},
		{60, "FREQ=MINUTELY;INTERVAL=1"},
		{30, "FREQ=SECONDLY;INTERVAL=30"},
	}
	for _, c := range cases {
		t.Run(c.want, func(t *testing.T) {
			assert.Equal(t, c.want, secondsToRRule(c.seconds))
		})
	}
}

// migTestLegacyTask mirrors the pre-migration tasks schema (legacy repeat
// columns present; repeats/repeats_from_current_date added by the migration).
type migTestLegacyTask struct {
	ID                     int64     `xorm:"bigint not null pk"`
	Title                  string    `xorm:"text not null"`
	Description            string    `xorm:"text null"`
	Done                   bool      `xorm:"null"`
	DoneAt                 time.Time `xorm:"datetime null"`
	DueDate                time.Time `xorm:"datetime null"`
	ProjectID              int64     `xorm:"bigint not null"`
	RepeatAfter            int64     `xorm:"bigint null"`
	RepeatMode             int       `xorm:"not null default 0"`
	Priority               int64     `xorm:"bigint null"`
	StartDate              time.Time `xorm:"datetime null"`
	EndDate                time.Time `xorm:"datetime null"`
	HexColor               string    `xorm:"varchar(7) null"`
	PercentDone            float64   `xorm:"double null"`
	Index                  int64     `xorm:"'index' not null default 0"`
	UID                    string    `xorm:"'uid' text null"`
	CoverImageAttachmentID int64     `xorm:"bigint null default 0"`
	Created                time.Time `xorm:"datetime not null"`
	Updated                time.Time `xorm:"datetime not null"`
	CreatedByID            int64     `xorm:"bigint not null"`
}

func (migTestLegacyTask) TableName() string { return "tasks" }

type migTestNewTask struct {
	ID                     int64  `xorm:"bigint not null pk"`
	Repeats                string `xorm:"varchar(500) null"`
	RepeatsFromCurrentDate bool   `xorm:"null"`
}

func (migTestNewTask) TableName() string { return "tasks" }

type migTestBackup struct {
	ID          int64 `xorm:"bigint not null pk"`
	RepeatAfter int64 `xorm:"bigint null"`
	RepeatMode  int   `xorm:"not null default 0"`
}

func (migTestBackup) TableName() string { return "task_repeat_legacy_backup" }

// TestRRuleMigrationSQLite exercises the migration's runtime on the SQLite path:
// it builds the legacy schema, seeds rows across every repeat_mode, runs the real
// Migrate function, and asserts the conversion, the legacy-data backup, and the
// column drop. mage test:feature only syncs current models, so this is the only
// coverage of the migration's batch/backup/drop logic.
func TestRRuleMigrationSQLite(t *testing.T) {
	prevDBType := config.DatabaseType.GetString()
	config.DatabaseType.Set("sqlite")
	t.Cleanup(func() { config.DatabaseType.Set(prevDBType) })
	log.InitLogger() // the migration logs an audit summary; without this it panics on a nil logger

	engine, err := xorm.NewEngine("sqlite3", filepath.Join(t.TempDir(), "migtest.db"))
	require.NoError(t, err)
	defer engine.Close()
	engine.SetMapper(names.GonicMapper{})

	// Build the pre-migration schema and seed legacy repeat data.
	require.NoError(t, engine.Sync2(migTestLegacyTask{}))
	now := time.Now()
	seed := []migTestLegacyTask{
		{ID: 1, Title: "daily", ProjectID: 1, Index: 1, RepeatAfter: 86400, RepeatMode: 0, Created: now, Updated: now, CreatedByID: 1},
		{ID: 2, Title: "monthly", ProjectID: 1, Index: 2, RepeatAfter: 0, RepeatMode: 1, Created: now, Updated: now, CreatedByID: 1},
		{ID: 3, Title: "weekly from current", ProjectID: 1, Index: 3, RepeatAfter: 604800, RepeatMode: 2, Created: now, Updated: now, CreatedByID: 1},
		{ID: 4, Title: "no repeat", ProjectID: 1, Index: 4, RepeatAfter: 0, RepeatMode: 0, Created: now, Updated: now, CreatedByID: 1},
	}
	for i := range seed {
		_, err = engine.Insert(&seed[i])
		require.NoError(t, err)
	}

	// Seed more than one page (batchSize = 500) of legacy rows so the migration's
	// id-paged conversion loop is actually exercised across batches.
	const bulkCount = 505
	bulk := make([]migTestLegacyTask, 0, bulkCount)
	for i := range bulkCount {
		bulk = append(bulk, migTestLegacyTask{
			ID:          int64(1001 + i),
			Title:       "bulk",
			ProjectID:   1,
			Index:       int64(1001 + i),
			RepeatAfter: 86400,
			RepeatMode:  0,
			Created:     now,
			Updated:     now,
			CreatedByID: 1,
		})
	}
	_, err = engine.Insert(&bulk)
	require.NoError(t, err)

	// Run the migration under test.
	ran := false
	for _, m := range migrations {
		if m.ID == "20251229100000" {
			require.NoError(t, m.Migrate(engine))
			ran = true
			break
		}
	}
	require.True(t, ran, "migration 20251229100000 was not found in the list")

	// Conversion: legacy modes map to the expected RRULE; mode 2 sets the flag.
	get := func(id int64) migTestNewTask {
		nt := migTestNewTask{}
		found, gerr := engine.ID(id).Get(&nt)
		require.NoError(t, gerr)
		require.True(t, found, "task %d should exist after migration", id)
		return nt
	}
	assert.Equal(t, "FREQ=DAILY;INTERVAL=1", get(1).Repeats)
	assert.Equal(t, "FREQ=MONTHLY;INTERVAL=1", get(2).Repeats)
	t3 := get(3)
	assert.Equal(t, "FREQ=WEEKLY;INTERVAL=1", t3.Repeats)
	assert.True(t, t3.RepeatsFromCurrentDate, "mode 2 should set repeats_from_current_date")
	assert.Empty(t, get(4).Repeats, "a task without legacy repeat should stay empty")
	// A row beyond the first batch (id 1505 is on page 2) is also converted.
	assert.Equal(t, "FREQ=DAILY;INTERVAL=1", get(1505).Repeats, "rows past the first batch should be converted")

	// Backup: every row that had legacy repeat data is preserved (1, 2, 3 plus the
	// bulk rows); the no-repeat task (4) is not.
	var backups []migTestBackup
	require.NoError(t, engine.OrderBy("id ASC").Find(&backups))
	require.Len(t, backups, 3+bulkCount)
	assert.Equal(t, int64(86400), backups[0].RepeatAfter)
	assert.Equal(t, 1, backups[1].RepeatMode)
	assert.Equal(t, int64(604800), backups[2].RepeatAfter)

	// Drop: the legacy columns are gone.
	_, err = engine.QueryString("SELECT repeat_after FROM tasks LIMIT 1")
	assert.Error(t, err, "repeat_after column should have been dropped")
	_, err = engine.QueryString("SELECT repeat_mode FROM tasks LIMIT 1")
	assert.Error(t, err, "repeat_mode column should have been dropped")
}
