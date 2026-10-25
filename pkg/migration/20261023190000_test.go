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
	"sort"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

const intendedDueTimeColumn20261023190000 = "intended_due_time"

// tasksBefore20261023190000 is the tasks table just before this migration: the model of
// pkg/models/tasks.go after the RRULE migration 20261022220000, without intended_due_time. It
// is synced with plain Sync2, so its indexes are the ones the model's tags create.
type tasksBefore20261023190000 struct {
	ID                     int64     `xorm:"bigint autoincr not null unique pk"`
	Title                  string    `xorm:"TEXT not null"`
	Description            string    `xorm:"longtext null"`
	ProjectID              int64     `xorm:"bigint INDEX not null unique(tasks_project_index) index(project_done_due_date)"`
	Done                   bool      `xorm:"INDEX null index(project_done_due_date)"`
	DoneAt                 time.Time `xorm:"INDEX null 'done_at'"`
	DueDate                time.Time `xorm:"DATETIME INDEX null index(project_done_due_date) 'due_date'"`
	Repeats                string    `xorm:"varchar(500) null"`
	RepeatsFromCurrentDate bool      `xorm:"null default false"`
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

func (tasksBefore20261023190000) TableName() string { return "tasks" }

// taskRemindersBefore20261023190000 is the task_reminders table (pkg/models/task_reminder.go),
// which this migration must leave alone.
type taskRemindersBefore20261023190000 struct {
	ID             int64     `xorm:"bigint autoincr not null unique pk"`
	TaskID         int64     `xorm:"bigint not null INDEX"`
	Reminder       time.Time `xorm:"DATETIME not null INDEX 'reminder'"`
	Created        time.Time `xorm:"created not null"`
	RelativePeriod int64     `xorm:"bigint null"`
	RelativeTo     string    `xorm:"varchar(50) null"`
}

func (taskRemindersBefore20261023190000) TableName() string { return "task_reminders" }

// tasksIntended20261023190000 reads the new column back with the row's id.
type tasksIntended20261023190000 struct {
	ID              int64 `xorm:"pk 'id'"`
	IntendedDueTime *int  `xorm:"'intended_due_time'"`
}

func (tasksIntended20261023190000) TableName() string { return "tasks" }

func dropTables20261023190000(t *testing.T, x *xorm.Engine) {
	t.Helper()
	require.NoError(t, x.DropTables("tasks", "task_reminders"))
}

func testEngine20261023190000(t *testing.T) *xorm.Engine {
	t.Helper()
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	dropTables20261023190000(t, x)
	t.Cleanup(func() { dropTables20261023190000(t, x) })
	return x
}

// createTasksBefore20261023190000 builds the pre-migration tables, with UQE_tasks_project_index
// as upstream's migration 20260411013328 creates it on upgraded installs.
func createTasksBefore20261023190000(t *testing.T, x *xorm.Engine) {
	t.Helper()
	require.NoError(t, x.Sync2(tasksBefore20261023190000{}, taskRemindersBefore20261023190000{}))
	query := `CREATE UNIQUE INDEX IF NOT EXISTS UQE_tasks_project_index ON tasks (project_id, "index")`
	if x.Dialect().URI().DBType == schemas.MYSQL {
		query = "CREATE UNIQUE INDEX UQE_tasks_project_index ON tasks (project_id, `index`)"
	}
	_, err := x.Exec(query)
	require.NoError(t, err)
}

func migration20261023190000(t *testing.T) *xormigrate.Migration {
	t.Helper()
	for _, m := range migrations {
		if m.ID == "20261023190000" {
			return m
		}
	}
	t.Fatal("migration 20261023190000 is not registered")
	return nil
}

func tableColumns20261023190000(t *testing.T, x *xorm.Engine, name string) map[string]*schemas.Column {
	t.Helper()
	tables, err := x.DBMetas()
	require.NoError(t, err)
	for _, table := range tables {
		if table.Name == name {
			out := map[string]*schemas.Column{}
			for _, c := range table.Columns() {
				out[c.Name] = c
			}
			return out
		}
	}
	t.Fatalf("table %s not found", name)
	return nil
}

func columnNames20261023190000(cols map[string]*schemas.Column) []string {
	names := make([]string, 0, len(cols))
	for n := range cols {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func TestIntendedDueTimeMigration20261023190000(t *testing.T) {
	x := testEngine20261023190000(t)
	createTasksBefore20261023190000(t, x)

	for i, done := range []bool{false, true} {
		_, err := x.Insert(&tasksBefore20261023190000{
			Title: "existing", ProjectID: 1, Index: int64(i + 1), Done: done, CreatedByID: 1,
			DueDate: time.Date(2027, 3, 14, 7, 30, 0, 0, time.UTC), Repeats: "FREQ=DAILY;INTERVAL=1",
		})
		require.NoError(t, err)
	}
	_, err := x.Insert(&taskRemindersBefore20261023190000{TaskID: 1, Reminder: time.Date(2027, 3, 14, 7, 0, 0, 0, time.UTC)})
	require.NoError(t, err)

	tasksBefore := tableColumns20261023190000(t, x, "tasks")
	require.NotContains(t, tasksBefore, intendedDueTimeColumn20261023190000)
	remindersBefore := columnNames20261023190000(tableColumns20261023190000(t, x, "task_reminders"))
	indexesBefore := tasksIndexes20261022220000(t, x)
	requireIndexNames20261022220000(t, indexesBefore,
		"UQE_tasks_project_index", "UQE_tasks_tasks_project_index", "IDX_tasks_project_done_due_date", "IDX_tasks_due_date")

	m := migration20261023190000(t)
	require.NoError(t, m.Migrate(x))

	tasksAfter := tableColumns20261023190000(t, x, "tasks")
	require.Contains(t, tasksAfter, intendedDueTimeColumn20261023190000)
	assert.True(t, tasksAfter[intendedDueTimeColumn20261023190000].Nullable, "intended_due_time is nullable")
	assert.ElementsMatch(t,
		append(columnNames20261023190000(tasksBefore), intendedDueTimeColumn20261023190000),
		columnNames20261023190000(tasksAfter), "exactly one column is added")
	assert.Equal(t, indexesBefore, tasksIndexes20261022220000(t, x), "every tasks index survives verbatim")
	assert.Equal(t, remindersBefore, columnNames20261023190000(tableColumns20261023190000(t, x, "task_reminders")),
		"task_reminders gains no column")

	rows := []tasksIntended20261023190000{}
	require.NoError(t, x.Find(&rows))
	require.Len(t, rows, 2)
	for _, r := range rows {
		assert.Nil(t, r.IntendedDueTime, "task %d: an existing row is NULL", r.ID)
	}

	t.Run("a second run is a no-op", func(t *testing.T) {
		require.NoError(t, m.Migrate(x))
		assert.Equal(t, columnNames20261023190000(tasksAfter), columnNames20261023190000(tableColumns20261023190000(t, x, "tasks")))
		assert.Equal(t, indexesBefore, tasksIndexes20261022220000(t, x))
	})
}

// A fresh install (the model synced) has the same column with the same type as an upgraded one.
func TestIntendedDueTimeFreshInstallMatches20261023190000(t *testing.T) {
	x := testEngine20261023190000(t)
	createTasksBefore20261023190000(t, x)
	require.NoError(t, migration20261023190000(t).Migrate(x))
	migrated := tableColumns20261023190000(t, x, "tasks")[intendedDueTimeColumn20261023190000]
	require.NotNil(t, migrated)
	dropTables20261023190000(t, x)

	existing := map[string]bool{}
	metas, err := x.DBMetas()
	require.NoError(t, err)
	for _, m := range metas {
		existing[m.Name] = true
	}
	tables := append(models.GetTables(), &user.User{})
	require.NoError(t, x.Sync2(tables...))
	t.Cleanup(func() {
		metas, err := x.DBMetas()
		require.NoError(t, err)
		var created []any
		for _, m := range metas {
			if !existing[m.Name] {
				created = append(created, m.Name)
			}
		}
		require.NoError(t, x.DropTables(created...))
	})

	fresh := tableColumns20261023190000(t, x, "tasks")[intendedDueTimeColumn20261023190000]
	require.NotNil(t, fresh, "the model declares intended_due_time")
	assert.Equal(t, migrated.SQLType.Name, fresh.SQLType.Name)
	assert.Equal(t, migrated.Nullable, fresh.Nullable)
}

// The migration is registered and sorts after the earlier recurrence migrations and after every
// other registered migration, upstream's included. Upstream migrations may sort between the
// recurrence migrations (20261023094626 already does), so no gap is asserted.
func TestMigrationOrdering20261023190000(t *testing.T) {
	const id = "20261023190000"
	found := false
	for _, m := range migrations {
		if m.ID == id {
			found = true
			require.NotNil(t, m.Migrate)
			continue
		}
		assert.Lessf(t, m.ID, id, "migration %s sorts after %s; re-date %s", m.ID, id, id)
	}
	require.True(t, found, "migration %s is not registered", id)
	for _, earlier := range []string{"20261022220000", "20261022220100"} {
		assert.Less(t, earlier, id)
	}
}
