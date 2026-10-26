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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teambition/rrule-go"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
	"xorm.io/xorm/schemas"
)

// tasksBefore20261022220000 is upstream's tasks table as it stands just before this
// migration (pkg/models/tasks.go at upstream/main a5dadc514), legacy repeat columns
// included. It is synced with plain Sync2, so its indexes are exactly the ones a
// fresh upstream install has. The deleted behaviour tag is left off DeletedAt so the
// test's own reads see soft-deleted rows; it does not change the DDL.
type tasksBefore20261022220000 struct {
	ID                     int64     `xorm:"bigint autoincr not null unique pk"`
	Title                  string    `xorm:"TEXT not null"`
	Description            string    `xorm:"longtext null"`
	ProjectID              int64     `xorm:"bigint INDEX not null unique(tasks_project_index) index(project_done_due_date)"`
	Done                   bool      `xorm:"INDEX null index(project_done_due_date)"`
	DoneAt                 time.Time `xorm:"INDEX null 'done_at'"`
	DueDate                time.Time `xorm:"DATETIME INDEX null index(project_done_due_date) 'due_date'"`
	RepeatAfter            int64     `xorm:"bigint INDEX null"`
	RepeatMode             int       `xorm:"not null default 0"`
	Priority               int64     `xorm:"bigint null"`
	StartDate              time.Time `xorm:"DATETIME INDEX null 'start_date'"`
	EndDate                time.Time `xorm:"DATETIME INDEX null 'end_date'"`
	HexColor               string    `xorm:"varchar(6) null"`
	PercentDone            float64   `xorm:"DOUBLE null"`
	Index                  int64     `xorm:"bigint not null default 0 unique(tasks_project_index)"`
	UID                    string    `xorm:"varchar(250) null"`
	CoverImageAttachmentID int64     `xorm:"bigint default 0"`
	Created                time.Time `xorm:"created not null"`
	Updated                time.Time `xorm:"updated not null"`
	DeletedAt              time.Time `xorm:"datetime null INDEX 'deleted_at'"`
	CreatedByID            int64     `xorm:"bigint not null"`
}

func (tasksBefore20261022220000) TableName() string { return "tasks" }

// tasksCore20261022220000 is the part of a task row that must survive the migration
// unchanged. Times are compared as unix seconds.
type tasksCore20261022220000 struct {
	ID          int64     `xorm:"pk 'id'"`
	Title       string    `xorm:"'title'"`
	Description string    `xorm:"'description'"`
	ProjectID   int64     `xorm:"'project_id'"`
	Index       int64     `xorm:"'index'"`
	Done        bool      `xorm:"'done'"`
	DueDate     time.Time `xorm:"'due_date'"`
	Priority    int64     `xorm:"'priority'"`
	DeletedAt   time.Time `xorm:"'deleted_at'"`
	CreatedByID int64     `xorm:"'created_by_id'"`
}

func (tasksCore20261022220000) TableName() string { return "tasks" }

type tasksAfter20261022220000 struct {
	ID                     int64  `xorm:"pk 'id'"`
	Repeats                string `xorm:"'repeats'"`
	RepeatsFromCurrentDate bool   `xorm:"'repeats_from_current_date'"`
}

func (tasksAfter20261022220000) TableName() string { return "tasks" }

// tasksRRuleColumns20261022220000 adds the new columns the way the migration does,
// for tests that build an already (partly) converted schema by hand.
type tasksRRuleColumns20261022220000 struct {
	ID                     int64  `xorm:"bigint autoincr not null unique pk"`
	Repeats                string `xorm:"varchar(500) null"`
	RepeatsFromCurrentDate bool   `xorm:"null default false"`
}

func (tasksRRuleColumns20261022220000) TableName() string { return "tasks" }

// projectsOwner20261022220000 and usersZone20261022220000 are the only columns of projects
// and users the conversion joins on; users.timezone is tagged as in pkg/user/user.go.
type projectsOwner20261022220000 struct {
	ID      int64 `xorm:"bigint autoincr not null unique pk"`
	OwnerID int64 `xorm:"bigint not null"`
}

func (projectsOwner20261022220000) TableName() string { return "projects" }

type usersZone20261022220000 struct {
	ID       int64  `xorm:"bigint autoincr not null unique pk"`
	Timezone string `xorm:"varchar(255) null"`
}

func (usersZone20261022220000) TableName() string { return "users" }

// legacyRepeatCase20261022220000 is one element of testdata/legacy_repeat_cases.json.
type legacyRepeatCase20261022220000 struct {
	Name                string  `json:"name"`
	RepeatAfter         int64   `json:"repeat_after"`
	RepeatMode          int     `json:"repeat_mode"`
	DueDate             *string `json:"due_date"`
	Timezone            string  `json:"timezone"`
	WantRule            string  `json:"want_rule"`
	WantFromCurrentDate bool    `json:"want_from_current_date"`
}

func (c legacyRepeatCase20261022220000) dueDate(t *testing.T) time.Time {
	t.Helper()
	if c.DueDate == nil {
		return time.Time{}
	}
	d, err := time.Parse(time.RFC3339, *c.DueDate)
	require.NoError(t, err, c.Name)
	return d
}

func loadLegacyRepeatCases20261022220000(t *testing.T) []legacyRepeatCase20261022220000 {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "legacy_repeat_cases.json"))
	require.NoError(t, err)
	var cases []legacyRepeatCase20261022220000
	require.NoError(t, json.Unmarshal(raw, &cases))
	require.NotEmpty(t, cases)
	return cases
}

func withServiceTimeZone20261022220000(t *testing.T, zone string) {
	t.Helper()
	prev := config.ServiceTimeZone.GetString()
	config.ServiceTimeZone.Set(zone)
	t.Cleanup(func() { config.ServiceTimeZone.Set(prev) })
}

// captureLog20261022220000 sends the standard log to a file for the rest of the
// test and returns a reader for it.
func captureLog20261022220000(t *testing.T) func() string {
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

// The converter reads the due day in the zone it is given (the owner's), never in
// the server zone, which is set to a different zone here on purpose.
func TestConvertLegacyRepeat20261022220000(t *testing.T) {
	withServiceTimeZone20261022220000(t, "Asia/Tokyo")
	for _, c := range loadLegacyRepeatCases20261022220000(t) {
		t.Run(c.Name, func(t *testing.T) {
			zone := c.Timezone
			if zone == "" {
				zone = "UTC"
			}
			loc, err := time.LoadLocation(zone)
			require.NoError(t, err)

			rule, fromCurrentDate := convertLegacyRepeat20261022220000(c.RepeatAfter, c.RepeatMode, c.dueDate(t), loc)
			assert.Equal(t, c.WantRule, rule)
			assert.Equal(t, c.WantFromCurrentDate, fromCurrentDate)
			if rule != "" {
				_, err := rrule.StrToRRule(rule)
				require.NoError(t, err, "emitted rule must parse")
			}
		})
	}
}

func TestConvertLegacyRepeatMonthEndOccurrences20261022220000(t *testing.T) {
	withServiceTimeZone20261022220000(t, "UTC")

	for _, tc := range []struct {
		year int
		feb  int
	}{{2025, 28}, {2024, 29}} {
		t.Run(strconv.Itoa(tc.year), func(t *testing.T) {
			dtstart := time.Date(tc.year, time.January, 31, 9, 0, 0, 0, time.UTC)
			rule, _ := convertLegacyRepeat20261022220000(0, 1, dtstart, time.UTC)

			opt, err := rrule.StrToROption(rule)
			require.NoError(t, err)
			opt.Dtstart = dtstart
			r, err := rrule.NewRRule(*opt)
			require.NoError(t, err)

			want := []time.Time{
				time.Date(tc.year, time.February, tc.feb, 9, 0, 0, 0, time.UTC),
				time.Date(tc.year, time.March, 31, 9, 0, 0, 0, time.UTC),
				time.Date(tc.year, time.April, 30, 9, 0, 0, 0, time.UTC),
			}
			next := dtstart
			for _, w := range want {
				next = r.After(next, false)
				assert.Equal(t, w, next)
			}
		})
	}
}

// The owner-zone rules match pkg/models' task-zone resolver: an empty name and "Local"
// are unset, a name time.LoadLocation rejects is invalid, and a missing project or owner
// row falls back too. Every fallback is the server zone.
func TestOwnerZone20261022220000(t *testing.T) {
	withServiceTimeZone20261022220000(t, "Asia/Tokyo")
	tokyo, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)
	z := ownerZones20261022220000{loaded: map[string]*time.Location{}}

	for _, tc := range []struct {
		name         string
		row          legacyTaskOwnerZone20261022220000
		wantZone     string
		wantFallback string
	}{
		{"valid zone", legacyTaskOwnerZone20261022220000{ProjectID: 1, OwnerID: 1, Timezone: "America/New_York"}, "America/New_York", ""},
		{"valid zone again, from the cache", legacyTaskOwnerZone20261022220000{ProjectID: 2, OwnerID: 2, Timezone: "America/New_York"}, "America/New_York", ""},
		{"UTC is a zone", legacyTaskOwnerZone20261022220000{ProjectID: 1, OwnerID: 1, Timezone: "UTC"}, "UTC", ""},
		{"empty or NULL zone", legacyTaskOwnerZone20261022220000{ProjectID: 1, OwnerID: 1}, tokyo.String(), "empty zone"},
		{"Local is the process zone, never a user's", legacyTaskOwnerZone20261022220000{ProjectID: 1, OwnerID: 1, Timezone: "Local"}, tokyo.String(), "invalid zone"},
		{"invalid zone", legacyTaskOwnerZone20261022220000{ProjectID: 1, OwnerID: 1, Timezone: "Mars/Olympus"}, tokyo.String(), "invalid zone"},
		{"invalid zone again, from the cache", legacyTaskOwnerZone20261022220000{ProjectID: 1, OwnerID: 1, Timezone: "Mars/Olympus"}, tokyo.String(), "invalid zone"},
		{"no owner row", legacyTaskOwnerZone20261022220000{ProjectID: 1, Timezone: ""}, tokyo.String(), "no owner"},
		{"no project row", legacyTaskOwnerZone20261022220000{}, tokyo.String(), "no project"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			loc, fallback := z.zone(&tc.row)
			require.NotNil(t, loc)
			assert.Equal(t, tc.wantZone, loc.String())
			assert.Equal(t, tc.wantFallback, fallback)
		})
	}
	assert.Len(t, z.loaded, 4, "one cache entry per distinct name, misses included")
}

// upstreamHeadMigration20261022220000 is the newest migration on upstream/main when
// this migration was dated (a5dadc514, fetched 2026-10-22). A newer upstream
// migration sorting below ours would mean this one must be re-dated.
const upstreamHeadMigration20261022220000 = "20261020024804"

// The RRULE migration sorts after every upstream migration. Later
// migrations of this change (e.g. 20261022220100) may sort after it.
func TestMigrationSortsAfterUpstream20261022220000(t *testing.T) {
	found := false
	for _, m := range migrations {
		assert.NotEqual(t, "20260711120000", m.ID, "the PR's out-of-order migration must be retired")
		if m.ID == "20261022220000" {
			found = true
			require.NotNil(t, m.Migrate)
			continue
		}
		if m.ID < "20261022220000" {
			assert.LessOrEqual(t, m.ID, upstreamHeadMigration20261022220000,
				"migration %s sorts between upstream's head and the RRULE migration; re-date 20261022220000", m.ID)
		}
	}
	require.True(t, found, "migration 20261022220000 is not registered")
}

func TestRebuildTableDDL20261022220000(t *testing.T) {
	drop := map[string]bool{"repeat_after": true, "repeat_mode": true}

	t.Run("xorm DDL", func(t *testing.T) {
		in := "CREATE TABLE `tasks` (`id` INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, `title` TEXT NOT NULL, " +
			"`repeat_after` INTEGER NULL, `repeat_mode` INTEGER DEFAULT 0 NOT NULL, `index` INTEGER DEFAULT 0 NOT NULL)"
		got, err := rebuildTableDDL20261022220000(in, drop)
		require.NoError(t, err)
		assert.Equal(t, "CREATE TABLE \"tasks_rrule_tmp\" (`id` INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, `title` TEXT NOT NULL, "+
			"`index` INTEGER DEFAULT 0 NOT NULL)", got)
	})

	t.Run("quotes, parentheses, comments and table options", func(t *testing.T) {
		in := "CREATE TABLE \"tasks\"(\n" +
			"  [id] INTEGER PRIMARY KEY,\n" +
			"  \"odd,name\" TEXT DEFAULT 'a, b '' c' CHECK (length(\"odd,name\") < (10 + 2)),\n" +
			"  repeat_after INTEGER, -- legacy, gone\n" +
			"  created INTEGER DEFAULT (strftime('%s', 'now')),\n" +
			"  `repeat_mode` INTEGER /* legacy, gone */,\n" +
			"  UNIQUE (created, [id])\n" +
			") WITHOUT ROWID"
		got, err := rebuildTableDDL20261022220000(in, drop)
		require.NoError(t, err)
		assert.Equal(t, "CREATE TABLE \"tasks_rrule_tmp\" ([id] INTEGER PRIMARY KEY, "+
			"\"odd,name\" TEXT DEFAULT 'a, b '' c' CHECK (length(\"odd,name\") < (10 + 2)), "+
			"created INTEGER DEFAULT (strftime('%s', 'now')), "+
			"UNIQUE (created, [id])) WITHOUT ROWID", got)
	})

	t.Run("a table constraint on a dropped column fails loudly", func(t *testing.T) {
		in := "CREATE TABLE `tasks` (`id` INTEGER, `repeat_after` INTEGER, UNIQUE (`id`, `repeat_after`))"
		_, err := rebuildTableDDL20261022220000(in, drop)
		require.ErrorContains(t, err, "repeat_after")
	})

	t.Run("unbalanced DDL fails loudly", func(t *testing.T) {
		_, err := rebuildTableDDL20261022220000("CREATE TABLE `tasks` (`id` INTEGER", drop)
		require.Error(t, err)
	})
}

// rruleTestDB20261022220000 is a legacy-shaped tasks table seeded from the shared
// fixture, plus enough rows to need more than one page.
type rruleTestDB20261022220000 struct {
	x            *xorm.Engine
	wantRule     map[int64]string
	wantFromCurr map[int64]bool
	selected     int // rows the page predicate picks up
	converted    int
	unconvert    int
	rows         int
	nullRowID    int64
	deletedRowID int64
	fallback     int // converted rows whose owner zone fell back to the server zone
}

const rruleBulkRows20261022220000 = 505

func dropRRuleTables20261022220000(t *testing.T, x *xorm.Engine) {
	t.Helper()
	require.NoError(t, x.DropTables("tasks", "task_repeat_legacy_backup", "tasks_rrule_tmp", "projects", "users"))
}

func testEngine20261022220000(t *testing.T) *xorm.Engine {
	t.Helper()
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	dropRRuleTables20261022220000(t, x)
	t.Cleanup(func() { dropRRuleTables20261022220000(t, x) })
	return x
}

// createLegacyTasks20261022220000 builds upstream's tasks table: the fresh-install
// indexes from the struct tags, plus UQE_tasks_project_index created exactly as
// upstream's migration 20260411013328 creates it on upgraded installs.
func createLegacyTasks20261022220000(t *testing.T, x *xorm.Engine) {
	t.Helper()
	require.NoError(t, x.Sync2(tasksBefore20261022220000{}))
	idx := `"index"`
	query := "CREATE UNIQUE INDEX IF NOT EXISTS UQE_tasks_project_index ON tasks (project_id, " + idx + ")"
	if x.Dialect().URI().DBType == schemas.MYSQL {
		idx = "`index`"
		query = "CREATE UNIQUE INDEX UQE_tasks_project_index ON tasks (project_id, " + idx + ")"
	}
	_, err := x.Exec(query)
	require.NoError(t, err)
	createOwnerTables20261022220000(t, x)
}

// ownerProject20261022220000 is the project every test row lives in unless it says
// otherwise; its owner has a valid UTC zone, so its rows never fall back.
const ownerProject20261022220000 = 1

// createOwnerTables20261022220000 creates the projects and users columns the conversion
// joins on, with project 1 owned by a UTC user. Every path that reaches the conversion
// needs it: without the tables the join is an SQL error.
func createOwnerTables20261022220000(t *testing.T, x *xorm.Engine) {
	t.Helper()
	require.NoError(t, x.Sync2(projectsOwner20261022220000{}, usersZone20261022220000{}))
	_, err := x.Insert(&usersZone20261022220000{ID: 1, Timezone: "UTC"},
		&projectsOwner20261022220000{ID: ownerProject20261022220000, OwnerID: 1})
	require.NoError(t, err)
}

// addOwner20261022220000 adds a user with the zone and a project it owns, and returns the
// project. A nil zone stores NULL.
func addOwner20261022220000(t *testing.T, x *xorm.Engine, id int64, zone *string) int64 {
	t.Helper()
	_, err := x.Insert(&usersZone20261022220000{ID: id}, &projectsOwner20261022220000{ID: id, OwnerID: id})
	require.NoError(t, err)
	_, err = x.Exec("UPDATE users SET timezone = ? WHERE id = ?", zone, id)
	require.NoError(t, err)
	return id
}

func seedLegacyTasks20261022220000(t *testing.T) *rruleTestDB20261022220000 {
	t.Helper()
	x := testEngine20261022220000(t)
	withServiceTimeZone20261022220000(t, "UTC")
	createLegacyTasks20261022220000(t, x)

	s := &rruleTestDB20261022220000{x: x, wantRule: map[int64]string{}, wantFromCurr: map[int64]bool{}}
	add := func(task tasksBefore20261022220000, rule string, fromCurrentDate bool) {
		if task.ProjectID == 0 {
			task.ProjectID = ownerProject20261022220000
		}
		task.Index = task.ID
		task.CreatedByID = 1
		_, err := x.Insert(&task)
		require.NoError(t, err)
		s.wantRule[task.ID] = rule
		s.wantFromCurr[task.ID] = fromCurrentDate
		s.rows++
		if task.RepeatAfter > 0 || task.RepeatMode > 0 {
			s.selected++
			if rule == "" {
				s.unconvert++
			} else {
				s.converted++
			}
		}
	}

	// Every fixture case runs through the database, in a project whose owner has the
	// case's zone. The server zone stays UTC.
	ownerOf := map[string]int64{"": ownerProject20261022220000, "UTC": ownerProject20261022220000}
	nextOwner := int64(100)
	id := int64(1)
	for _, c := range loadLegacyRepeatCases20261022220000(t) {
		project, ok := ownerOf[c.Timezone]
		if !ok {
			zone := c.Timezone
			project = addOwner20261022220000(t, x, nextOwner, &zone)
			ownerOf[c.Timezone] = project
			nextOwner++
		}
		add(tasksBefore20261022220000{
			ID: id, Title: c.Name, Description: "case " + c.Name, ProjectID: project,
			RepeatAfter: c.RepeatAfter, RepeatMode: c.RepeatMode, DueDate: c.dueDate(t),
		}, c.WantRule, c.WantFromCurrentDate)
		id++
	}

	// The same New York month-end due date where the owner zone is unusable is read
	// in the server zone (UTC, Feb 1), and the migration goes on.
	empty, invalid := "", "Mars/Olympus"
	missingOwner := int64(200)
	_, err := x.Insert(&projectsOwner20261022220000{ID: missingOwner, OwnerID: 9999})
	require.NoError(t, err)
	for _, project := range []int64{
		addOwner20261022220000(t, x, 201, &empty),
		addOwner20261022220000(t, x, 202, &invalid),
		addOwner20261022220000(t, x, 203, nil),
		missingOwner,
		999, // no project row at all
	} {
		add(tasksBefore20261022220000{
			ID: id, Title: "fallback " + strconv.FormatInt(project, 10), ProjectID: project,
			RepeatMode: 1, DueDate: time.Date(2026, 2, 1, 1, 0, 0, 0, time.UTC),
		}, "FREQ=MONTHLY;INTERVAL=1", false)
		s.fallback++
		id++
	}

	// A plain task: repeats is NULL once the column is added and must end up ''.
	s.nullRowID = id
	add(tasksBefore20261022220000{ID: id, Title: "plain", Priority: 3}, "", false)
	id++

	// A soft-deleted repeating task is converted too.
	s.deletedRowID = id
	add(tasksBefore20261022220000{ID: id, Title: "deleted weekly", RepeatAfter: 604800}, "FREQ=WEEKLY;INTERVAL=1", false)
	_, err = x.Exec("UPDATE tasks SET deleted_at = ? WHERE id = ?", time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), id)
	require.NoError(t, err)

	for i := range rruleBulkRows20261022220000 {
		add(tasksBefore20261022220000{ID: int64(1001 + i), Title: "bulk", RepeatAfter: 86400}, "FREQ=DAILY;INTERVAL=1", false)
	}
	return s
}

// tasksIndexes20261022220000 returns every index on tasks, keyed by name, with a
// value that captures its definition in the dialect's own catalog.
func tasksIndexes20261022220000(t *testing.T, x *xorm.Engine) map[string]string {
	t.Helper()
	out := map[string]string{}
	switch x.Dialect().URI().DBType {
	case schemas.SQLITE:
		rows, err := x.QueryString("SELECT name, COALESCE(sql, '<implicit>') AS def FROM sqlite_master WHERE type = 'index' AND tbl_name = 'tasks'")
		require.NoError(t, err)
		for _, r := range rows {
			out[r["name"]] = r["def"]
		}
	case schemas.MYSQL:
		rows, err := x.QueryString("SELECT index_name AS n, column_name AS c, non_unique AS u FROM information_schema.statistics " +
			"WHERE table_schema = DATABASE() AND table_name = 'tasks' ORDER BY index_name, seq_in_index")
		require.NoError(t, err)
		for _, r := range rows {
			if prev, ok := out[r["n"]]; ok {
				out[r["n"]] = prev + "," + r["c"]
				continue
			}
			out[r["n"]] = "non_unique=" + r["u"] + " cols=" + r["c"]
		}
	case schemas.POSTGRES:
		rows, err := x.QueryString("SELECT indexname, indexdef FROM pg_indexes WHERE schemaname = current_schema() AND tablename = 'tasks'")
		require.NoError(t, err)
		for _, r := range rows {
			out[r["indexname"]] = r["indexdef"]
		}
	default:
		t.Fatalf("unsupported dialect %s", x.Dialect().URI().DBType)
	}
	return out
}

// legacyIndexName20261022220000 is the only tasks index the migration may drop: the
// one upstream's RepeatAfter `INDEX` tag creates. The expectation is named here on
// purpose, independently of the migration's own rule for which indexes to skip.
const legacyIndexName20261022220000 = "IDX_tasks_repeat_after"

func withoutLegacyIndex20261022220000(in map[string]string) map[string]string {
	out := map[string]string{}
	for name, def := range in {
		if strings.EqualFold(name, legacyIndexName20261022220000) {
			continue
		}
		out[name] = def
	}
	return out
}

// requirePreservedIndexes20261022220000 checks the after-migration index set against
// the before set minus the legacy index, by name, and pins the indexes that must survive.
func requirePreservedIndexes20261022220000(t *testing.T, before, after map[string]string) {
	t.Helper()
	assert.Equal(t, withoutLegacyIndex20261022220000(before), after,
		"every index except %s must survive verbatim", legacyIndexName20261022220000)
	requireIndexNames20261022220000(t, after,
		"UQE_tasks_project_index", "UQE_tasks_tasks_project_index", "IDX_tasks_project_done_due_date", "IDX_tasks_due_date")
	for name := range after {
		assert.NotEqualf(t, strings.ToLower(legacyIndexName20261022220000), strings.ToLower(name), "the legacy index must be gone")
	}
}

func requireIndexNames20261022220000(t *testing.T, indexes map[string]string, names ...string) {
	t.Helper()
	lower := map[string]bool{}
	for n := range indexes {
		lower[strings.ToLower(n)] = true
	}
	for _, n := range names {
		require.Truef(t, lower[strings.ToLower(n)], "index %s missing; have %v", n, indexes)
	}
}

func tasksColumns20261022220000(t *testing.T, x *xorm.Engine) []string {
	t.Helper()
	tables, err := x.DBMetas()
	require.NoError(t, err)
	for _, table := range tables {
		if table.Name == "tasks" {
			cols := append([]string(nil), table.ColumnsSeq()...)
			sort.Strings(cols)
			return cols
		}
	}
	t.Fatal("tasks table not found")
	return nil
}

func tableExists20261022220000(t *testing.T, x *xorm.Engine, name string) bool {
	t.Helper()
	exists, err := x.IsTableExist(name)
	require.NoError(t, err)
	return exists
}

func coreRows20261022220000(t *testing.T, x *xorm.Engine) []string {
	t.Helper()
	var rows []tasksCore20261022220000
	require.NoError(t, x.OrderBy("id").Find(&rows))
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, fmt.Sprintf("%d|%s|%s|%d|%d|%t|%d|%d|%d|%d", r.ID, r.Title, r.Description, r.ProjectID, r.Index,
			r.Done, r.DueDate.Unix(), r.Priority, r.DeletedAt.Unix(), r.CreatedByID))
	}
	return out
}

func sqliteSequence20261022220000(t *testing.T, x *xorm.Engine) string {
	t.Helper()
	rows, err := x.QueryString("SELECT seq FROM sqlite_sequence WHERE name = 'tasks'")
	require.NoError(t, err)
	if len(rows) == 0 {
		return ""
	}
	return rows[0]["seq"]
}

// rruleSnapshot20261022220000 is what the migration must leave untouched.
type rruleSnapshot20261022220000 struct {
	indexes map[string]string
	columns []string
	rows    []string
	seq     string
}

func snapshot20261022220000(t *testing.T, x *xorm.Engine) rruleSnapshot20261022220000 {
	t.Helper()
	snap := rruleSnapshot20261022220000{
		indexes: tasksIndexes20261022220000(t, x),
		columns: tasksColumns20261022220000(t, x),
		rows:    coreRows20261022220000(t, x),
	}
	if x.Dialect().URI().DBType == schemas.SQLITE {
		snap.seq = sqliteSequence20261022220000(t, x)
	}
	return snap
}

func expectedColumnsAfter20261022220000(before []string) []string {
	out := []string{"repeats", "repeats_from_current_date"}
	for _, c := range before {
		if c != "repeat_after" && c != "repeat_mode" {
			out = append(out, c)
		}
	}
	sort.Strings(out)
	return out
}

func backupCounts20261022220000(t *testing.T, x *xorm.Engine) (rows, distinct int) {
	t.Helper()
	res, err := x.QueryString("SELECT COUNT(*) AS n, COUNT(DISTINCT id) AS d FROM task_repeat_legacy_backup")
	require.NoError(t, err)
	rows, err = strconv.Atoi(res[0]["n"])
	require.NoError(t, err)
	distinct, err = strconv.Atoi(res[0]["d"])
	require.NoError(t, err)
	return rows, distinct
}

// assertConverted20261022220000 checks the end state every successful run must reach.
func (s *rruleTestDB20261022220000) assertConverted(t *testing.T, before rruleSnapshot20261022220000) {
	t.Helper()
	x := s.x

	var after []tasksAfter20261022220000
	require.NoError(t, x.OrderBy("id").Find(&after))
	require.Len(t, after, s.rows, "no row may be lost")
	for _, task := range after {
		assert.Equalf(t, s.wantRule[task.ID], task.Repeats, "rule of task %d", task.ID)
		assert.Equalf(t, s.wantFromCurr[task.ID], task.RepeatsFromCurrentDate, "repeats_from_current_date of task %d", task.ID)
	}

	nulls, err := x.Where("repeats IS NULL").Count(&tasksAfter20261022220000{})
	require.NoError(t, err)
	assert.Zero(t, nulls, "every NULL repeats must be normalised to ''")

	var backups []taskRepeatBackup20261022220000
	require.NoError(t, x.OrderBy("id").Find(&backups))
	require.Len(t, backups, s.selected, "one backup row per legacy-repeat task")
	rows, distinct := backupCounts20261022220000(t, x)
	assert.Equal(t, rows, distinct, "backup must not hold duplicate rows")

	assert.Equal(t, expectedColumnsAfter20261022220000(before.columns), tasksColumns20261022220000(t, x))
	assert.Equal(t, before.rows, coreRows20261022220000(t, x), "surviving columns must be unchanged")
	requirePreservedIndexes20261022220000(t, before.indexes, tasksIndexes20261022220000(t, x))
	if x.Dialect().URI().DBType == schemas.SQLITE {
		assert.Equal(t, before.seq, sqliteSequence20261022220000(t, x), "AUTOINCREMENT high-water mark must survive the rebuild")
	}
}

func (s *rruleTestDB20261022220000) assertBackupMatchesSeed(t *testing.T, seed map[int64][2]int64) {
	t.Helper()
	var backups []taskRepeatBackup20261022220000
	require.NoError(t, s.x.OrderBy("id").Find(&backups))
	for _, b := range backups {
		want, ok := seed[b.ID]
		require.Truef(t, ok, "backup row %d was never a legacy-repeat task", b.ID)
		assert.Equalf(t, want[0], b.RepeatAfter, "backup repeat_after of %d", b.ID)
		assert.Equalf(t, want[1], int64(b.RepeatMode), "backup repeat_mode of %d", b.ID)
	}
}

func legacySeedValues20261022220000(t *testing.T, x *xorm.Engine) map[int64][2]int64 {
	t.Helper()
	var rows []tasksBefore20261022220000
	require.NoError(t, x.Where("repeat_after > 0 OR repeat_mode > 0").Find(&rows))
	out := map[int64][2]int64{}
	for _, r := range rows {
		out[r.ID] = [2]int64{r.RepeatAfter, int64(r.RepeatMode)}
	}
	return out
}

// prepareSequence20261022220000 leaves the AUTOINCREMENT high-water mark above
// max(id), so the test can see whether the rebuild keeps it.
func prepareSequence20261022220000(t *testing.T, x *xorm.Engine) {
	t.Helper()
	if x.Dialect().URI().DBType != schemas.SQLITE {
		return
	}
	_, err := x.Insert(&tasksBefore20261022220000{ID: 900000, Title: "gone", ProjectID: 2, Index: 1, CreatedByID: 1})
	require.NoError(t, err)
	_, err = x.Exec("DELETE FROM tasks WHERE id = 900000")
	require.NoError(t, err)
	require.Equal(t, "900000", sqliteSequence20261022220000(t, x))
}

func TestRRuleMigration20261022220000(t *testing.T) {
	s := seedLegacyTasks20261022220000(t)
	prepareSequence20261022220000(t, s.x)
	before := snapshot20261022220000(t, s.x)
	requireIndexNames20261022220000(t, before.indexes,
		"UQE_tasks_project_index", "UQE_tasks_tasks_project_index", "IDX_tasks_project_done_due_date", "IDX_tasks_due_date", "IDX_tasks_repeat_after")
	seed := legacySeedValues20261022220000(t, s.x)
	readLog := captureLog20261022220000(t)

	require.NoError(t, migrateRRule20261022220000(s.x))

	s.assertConverted(t, before)
	s.assertBackupMatchesSeed(t, seed)
	assert.Empty(t, s.wantRule[s.nullRowID])
	assert.NotEmpty(t, s.wantRule[s.deletedRowID])

	// External post-upgrade verification parses the summary line; keep it byte-identical.
	logged := readLog()
	want := fmt.Sprintf("RRULE migration: %d legacy-repeat task(s) processed — %d converted, %d already had an RRULE, %d unconvertible (originals backed up to task_repeat_legacy_backup); %d row(s) used the server zone",
		s.selected, s.converted, 0, s.unconvert, s.fallback)
	require.Equal(t, 5, s.fallback)
	assert.Contains(t, logged, "msg=\""+want+"\"\n", "the whole summary line")
	assert.Equal(t, []string{strconv.Itoa(s.converted)}, regexp.MustCompile(` ([0-9]+) converted`).FindAllStringSubmatch(logged, -1)[0][1:])
	assert.Equal(t, []string{strconv.Itoa(s.unconvert)}, regexp.MustCompile(` ([0-9]+) unconvertible`).FindAllStringSubmatch(logged, -1)[0][1:])
}

func TestRRuleMigrationRetriesAfterFailure20261022220000(t *testing.T) {
	t.Cleanup(func() { testFailAfter20261022220000 = "" })

	sqliteOnly := map[string]bool{"after-copy": true, "after-drop": true, "during-index-recreate": true}
	serverOnly := map[string]bool{"between-drops": true}

	for _, point := range []string{"after-backup-page", "after-copy", "after-drop", "during-index-recreate", "between-drops"} {
		t.Run(point, func(t *testing.T) {
			x, err := db.CreateTestEngine() // also loads the DB config the skip depends on
			require.NoError(t, err)
			dbType := x.Dialect().URI().DBType
			isSQLite := dbType == schemas.SQLITE
			if sqliteOnly[point] && !isSQLite || serverOnly[point] && isSQLite {
				t.Skipf("%s does not apply to %s", point, dbType)
			}
			s := seedLegacyTasks20261022220000(t)
			prepareSequence20261022220000(t, s.x)
			before := snapshot20261022220000(t, s.x)
			seed := legacySeedValues20261022220000(t, s.x)

			testFailAfter20261022220000 = point
			err = migrateRRule20261022220000(s.x)
			testFailAfter20261022220000 = ""
			require.ErrorContains(t, err, point)

			if sqliteOnly[point] {
				// The rebuild is one transaction: the original table is intact.
				mid := snapshot20261022220000(t, s.x)
				assert.Equal(t, before.rows, mid.rows, "an interrupted rebuild must not lose rows")
				assert.Equal(t, before.indexes, mid.indexes, "an interrupted rebuild must not lose indexes")
				assert.Contains(t, mid.columns, "repeat_after")
				assert.Contains(t, mid.columns, "repeat_mode")
				assert.False(t, tableExists20261022220000(t, s.x, "tasks_rrule_tmp"), "the temp table is rolled back")
			}

			require.NoError(t, migrateRRule20261022220000(s.x))
			s.assertConverted(t, before)
			s.assertBackupMatchesSeed(t, seed)
		})
	}
}

func TestRRuleMigrationAlreadyConverted20261022220000(t *testing.T) {
	s := seedLegacyTasks20261022220000(t)
	before := snapshot20261022220000(t, s.x)
	require.NoError(t, migrateRRule20261022220000(s.x))
	s.assertConverted(t, before)

	checksum := func() []string {
		var rows []tasksAfter20261022220000
		require.NoError(t, s.x.OrderBy("id").Find(&rows))
		out := coreRows20261022220000(t, s.x)
		for _, r := range rows {
			out = append(out, fmt.Sprintf("%d|%s|%t", r.ID, r.Repeats, r.RepeatsFromCurrentDate))
		}
		return out
	}
	converted := checksum()
	indexes := tasksIndexes20261022220000(t, s.x)
	readLog := captureLog20261022220000(t)

	require.NoError(t, migrateRRule20261022220000(s.x))
	assert.Contains(t, readLog(), "RRULE migration: already converted, nothing to do")
	assert.Equal(t, converted, checksum(), "a re-run must change nothing")
	assert.Equal(t, indexes, tasksIndexes20261022220000(t, s.x))

	// A NULL written after conversion is still normalised on a re-run.
	_, err := s.x.Exec("UPDATE tasks SET repeats = NULL WHERE id = ?", s.nullRowID)
	require.NoError(t, err)
	require.NoError(t, migrateRRule20261022220000(s.x))
	nulls, err := s.x.Where("repeats IS NULL").Count(&tasksAfter20261022220000{})
	require.NoError(t, err)
	assert.Zero(t, nulls, "the already-converted state must still normalise NULL repeats") // the checksum reads NULL as ''
	assert.Equal(t, converted, checksum())
}

func TestRRuleMigrationFinishesPartialStates20261022220000(t *testing.T) {
	for _, tc := range []struct {
		name      string
		dropFirst string
		remaining string
	}{
		{"repeat_mode left", "repeat_after", "repeat_mode"},
		{"repeat_after left", "repeat_mode", "repeat_after"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			x := testEngine20261022220000(t)
			createLegacyTasks20261022220000(t, x)
			// One row already converted, one with legacy data but no rule: if the
			// conversion query ran it would fill the second, so it must stay empty.
			for _, task := range []tasksBefore20261022220000{
				{ID: 1, Title: "converted", ProjectID: 1, Index: 1, CreatedByID: 1, RepeatAfter: 86400},
				{ID: 2, Title: "legacy only", ProjectID: 1, Index: 2, CreatedByID: 1, RepeatMode: 1},
			} {
				_, err := x.Insert(&task)
				require.NoError(t, err)
			}
			require.NoError(t, partialSync(x, tasksRRuleColumns20261022220000{}))
			_, err := x.Exec("UPDATE tasks SET repeats = 'FREQ=DAILY;INTERVAL=1' WHERE id = 1")
			require.NoError(t, err)

			if x.Dialect().URI().DBType == schemas.SQLITE && tc.dropFirst == "repeat_after" {
				_, err = x.Exec("DROP INDEX IDX_tasks_repeat_after") // SQLite refuses to drop an indexed column
				require.NoError(t, err)
			}
			_, err = x.Exec("ALTER TABLE tasks DROP COLUMN " + tc.dropFirst)
			require.NoError(t, err)

			before := snapshot20261022220000(t, x)
			require.Contains(t, before.columns, tc.remaining)

			require.NoError(t, migrateRRule20261022220000(x))

			after := snapshot20261022220000(t, x)
			assert.NotContains(t, after.columns, tc.remaining)
			assert.Equal(t, before.rows, after.rows)
			requirePreservedIndexes20261022220000(t, before.indexes, after.indexes)
			assert.False(t, tableExists20261022220000(t, x, "task_repeat_legacy_backup"), "no conversion may run in a finish state")

			var rows []tasksAfter20261022220000
			require.NoError(t, x.OrderBy("id").Find(&rows))
			require.Len(t, rows, 2)
			assert.Equal(t, "FREQ=DAILY;INTERVAL=1", rows[0].Repeats)
			assert.Empty(t, rows[1].Repeats, "no conversion query may run in a finish state")
			nulls, err := x.Where("repeats IS NULL").Count(&tasksAfter20261022220000{})
			require.NoError(t, err)
			assert.Zero(t, nulls)
		})
	}
}

type tasksUnexpected20261022220000 struct {
	ID    int64  `xorm:"bigint autoincr not null unique pk"`
	Title string `xorm:"TEXT not null"`
}

func (tasksUnexpected20261022220000) TableName() string { return "tasks" }

func TestRRuleMigrationRejectsUnexpectedSchema20261022220000(t *testing.T) {
	x := testEngine20261022220000(t)
	require.NoError(t, x.Sync2(tasksUnexpected20261022220000{}))

	err := migrateRRule20261022220000(x)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "repeat_after")
	assert.Contains(t, err.Error(), "repeats")
}

func TestRRuleMigrationSQLiteForeignKeysOn20261022220000(t *testing.T) {
	shared, err := db.CreateTestEngine() // also loads the DB config the migration's helpers read
	require.NoError(t, err)
	if shared.Dialect().URI().DBType != schemas.SQLITE {
		t.Skip("SQLite only")
	}
	x, err := xorm.NewEngine("sqlite3", filepath.Join(t.TempDir(), "fk.db")+"?_foreign_keys=1")
	require.NoError(t, err)
	t.Cleanup(func() { _ = x.Close() })
	x.SetMapper(names.GonicMapper{})

	require.NoError(t, x.Sync2(tasksBefore20261022220000{}))
	createOwnerTables20261022220000(t, x)
	_, err = x.Insert(&tasksBefore20261022220000{ID: 1, Title: "t", ProjectID: 1, Index: 1, CreatedByID: 1, RepeatAfter: 60})
	require.NoError(t, err)
	before := snapshot20261022220000(t, x)

	err = migrateRRule20261022220000(x)
	require.ErrorContains(t, err, "foreign_keys")

	after := snapshot20261022220000(t, x)
	assert.Equal(t, before.rows, after.rows)
	assert.Equal(t, before.indexes, after.indexes)
	assert.Contains(t, after.columns, "repeat_after")
}

// The full run with the repeats column already present must keep a rule a row
// already has, and count it in the summary.
func TestRRuleMigrationKeepsExistingRules20261022220000(t *testing.T) {
	x := testEngine20261022220000(t)
	createLegacyTasks20261022220000(t, x)
	for _, task := range []tasksBefore20261022220000{
		{ID: 1, Title: "has a rule", ProjectID: 1, Index: 1, CreatedByID: 1, RepeatAfter: 86400},
		{ID: 2, Title: "legacy weekly", ProjectID: 1, Index: 2, CreatedByID: 1, RepeatAfter: 604800},
	} {
		_, err := x.Insert(&task)
		require.NoError(t, err)
	}
	require.NoError(t, partialSync(x, tasksRRuleColumns20261022220000{}))
	_, err := x.Exec("UPDATE tasks SET repeats = 'FREQ=YEARLY;INTERVAL=1' WHERE id = 1")
	require.NoError(t, err)
	readLog := captureLog20261022220000(t)

	require.NoError(t, migrateRRule20261022220000(x))

	var rows []tasksAfter20261022220000
	require.NoError(t, x.OrderBy("id").Find(&rows))
	require.Len(t, rows, 2)
	assert.Equal(t, "FREQ=YEARLY;INTERVAL=1", rows[0].Repeats, "an existing rule must not be overwritten")
	assert.Equal(t, "FREQ=WEEKLY;INTERVAL=1", rows[1].Repeats)
	assert.Contains(t, readLog(), "msg=\"RRULE migration: 2 legacy-repeat task(s) processed — 1 converted, 1 already had an RRULE, 0 unconvertible (originals backed up to task_repeat_legacy_backup); 0 row(s) used the server zone\"\n")
	rowsBackedUp, _ := backupCounts20261022220000(t, x)
	assert.Equal(t, 2, rowsBackedUp)
}

// An instance whose tasks were all deleted still has a sequence row; the rebuilt
// table receives no rows, so the high-water mark must be carried over explicitly.
func TestRRuleMigrationKeepsSequenceOfEmptyTable20261022220000(t *testing.T) {
	x := testEngine20261022220000(t)
	if x.Dialect().URI().DBType != schemas.SQLITE {
		t.Skip("SQLite only")
	}
	createLegacyTasks20261022220000(t, x)
	prepareSequence20261022220000(t, x) // inserts and deletes the only row

	require.NoError(t, migrateRRule20261022220000(x))

	assert.Equal(t, "900000", sqliteSequence20261022220000(t, x))
}
