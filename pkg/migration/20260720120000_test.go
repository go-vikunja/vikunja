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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

type usersPartial20260720120000 struct {
	BotOwnerID int64 `xorm:"bigint null index"`
}

func (usersPartial20260720120000) TableName() string {
	return "users"
}

func usersIndexOnUsername20260720120000(t *testing.T, x *xorm.Engine) *schemas.Index {
	t.Helper()
	indexes := usersIndexesOnUsername20260720120000(t, x)
	if len(indexes) == 0 {
		return nil
	}
	return indexes[0]
}

func TestRecreateMissingIndexes20260720120000(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	require.NoError(t, x.Sync2(user.GetTables()...))
	require.NotNil(t, usersIndexOnUsername20260720120000(t, x))

	// A partial-struct Sync makes xorm drop every index the struct doesn't declare.
	require.NoError(t, x.Sync(usersPartial20260720120000{}))
	require.Nil(t, usersIndexOnUsername20260720120000(t, x))

	_, err = x.Insert(&user.User{Username: "dup20260720120000"})
	require.NoError(t, err)
	_, err = x.Insert(&user.User{Username: "dup20260720120000"})
	require.NoError(t, err)

	err = recreateMissingIndexes20260720120000(x)
	require.ErrorContains(t, err, "users")
	require.ErrorContains(t, err, "username")

	_, err = x.Exec("DELETE FROM users WHERE username = ?", "dup20260720120000")
	require.NoError(t, err)

	require.NoError(t, recreateMissingIndexes20260720120000(x))
	index := usersIndexOnUsername20260720120000(t, x)
	require.NotNil(t, index)
	require.Equal(t, schemas.UniqueType, index.Type)

	// Idempotent on a healthy schema.
	require.NoError(t, recreateMissingIndexes20260720120000(x))
}

func usersIndexesOnUsername20260720120000(t *testing.T, x *xorm.Engine) []*schemas.Index {
	t.Helper()
	tables, err := x.DBMetas()
	require.NoError(t, err)
	for _, table := range tables {
		if table.Name != "users" {
			continue
		}
		indexes := make([]*schemas.Index, 0)
		for _, index := range table.Indexes {
			if len(index.Cols) == 1 && index.Cols[0] == "username" {
				indexes = append(indexes, index)
			}
		}
		return indexes
	}
	t.Fatal("users table not found")
	return nil
}

// An equivalent index under a different name (as pgloader produces) must not be
// duplicated by recreating the model's UQE_users_username index.
func TestRecreateMissingIndexesKeepsDifferentlyNamedIndex20260720120000(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	require.NoError(t, x.Sync2(user.GetTables()...))

	require.NoError(t, x.Sync(usersPartial20260720120000{}))
	require.Nil(t, usersIndexOnUsername20260720120000(t, x))

	_, err = x.Exec("CREATE UNIQUE INDEX some_nonmodel_name ON users (username)")
	require.NoError(t, err)
	t.Cleanup(func() {
		// x is the process-global test engine; leaving this index around leaks into later tests.
		// Can't use x.Dialect().DropIndexSQL here: its sqlite3 impl ignores Index.IsRegular and
		// always guesses an "IDX_"/"UQE_" prefix, mangling this deliberately non-convention name.
		dropSQL := "DROP INDEX some_nonmodel_name"
		if x.Dialect().URI().DBType == schemas.MYSQL {
			dropSQL += " ON users"
		}
		_, err := x.Exec(dropSQL)
		require.NoError(t, err)
		// The partial sync above dropped the model indexes on users.
		require.NoError(t, x.Sync2(user.GetTables()...))
	})

	require.NoError(t, recreateMissingIndexes20260720120000(x))

	indexes := usersIndexesOnUsername20260720120000(t, x)
	require.Len(t, indexes, 1)
	require.Equal(t, "some_nonmodel_name", indexes[0].Name)
}

// Older migrations create indexes with lowercase SQL, which xorm's sqlite3 dialect
// silently skips when reading the schema, so the migration tried to create them a
// second time and aborted the whole upgrade (#3313).
func TestRecreateMissingIndexesFindsLowercaseSQLiteIndexes20260720120000(t *testing.T) {
	x := sqliteTestEngine20260720120000(t)
	require.NoError(t, x.Sync2(user.GetTables()...))

	require.NoError(t, x.Sync(usersPartial20260720120000{}))
	require.Nil(t, usersIndexOnUsername20260720120000(t, x))

	_, err := x.Exec("create unique index UQE_users_username\n    on users (username)")
	require.NoError(t, err)
	t.Cleanup(func() {
		// x is the process-global test engine; a lowercase index leaks into later tests.
		_, err := x.Exec("DROP INDEX UQE_users_username")
		require.NoError(t, err)
		// The partial sync above dropped the model indexes on users.
		require.NoError(t, x.Sync2(user.GetTables()...))
	})

	require.NoError(t, recreateMissingIndexes20260720120000(x))

	// DBMetas is blind to the lowercase index by construction, so read the schema directly.
	rows, err := x.QueryString("SELECT name, sql FROM sqlite_master WHERE type = 'index' AND tbl_name = 'users' AND sql LIKE '%username%'")
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "UQE_users_username", rows[0]["name"])

	dbSchema, err := sqliteSchema20260720120000(x)
	require.NoError(t, err)
	require.NotNil(t, indexByName20260720120000(dbSchema.indexesByTable["users"], "UQE_users_username"))
}

// xorm's sqlite3 dialect splits an index's DDL on "(", so an expression index
// makes DBMetas fail with "Unknown col lower(username" — which used to abort this
// migration on its very first line.
func TestRecreateMissingIndexesWithSQLiteExpressionIndex20260720120000(t *testing.T) {
	x := sqliteTestEngine20260720120000(t)
	require.NoError(t, x.Sync2(user.GetTables()...))

	require.NoError(t, x.Sync(usersPartial20260720120000{}))

	_, err := x.Exec("CREATE INDEX IDX_users_expr20260720120000 ON users (lower(username))")
	require.NoError(t, err)
	t.Cleanup(func() {
		// x is the process-global test engine; leaving this index around breaks every
		// later test which reads the schema through DBMetas.
		_, err := x.Exec("DROP INDEX IDX_users_expr20260720120000")
		require.NoError(t, err)
		// The partial sync above dropped the model indexes on users.
		require.NoError(t, x.Sync2(user.GetTables()...))
	})

	require.NoError(t, recreateMissingIndexes20260720120000(x))

	// DBMetas cannot read this schema at all, so read it directly.
	dbSchema, err := sqliteSchema20260720120000(x)
	require.NoError(t, err)
	index := indexByName20260720120000(dbSchema.indexesByTable["users"], "UQE_users_username")
	require.NotNil(t, index)
	require.Equal(t, schemas.UniqueType, index.Type)
}

// The name is taken by an index which cannot stand in for the model's, so the
// migration must warn and move on instead of aborting the upgrade.
func TestRecreateMissingIndexesSkipsWhenIndexNameIsTaken20260720120000(t *testing.T) {
	x := sqliteTestEngine20260720120000(t)
	require.NoError(t, x.Sync2(user.GetTables()...))

	require.NoError(t, x.Sync(usersPartial20260720120000{}))
	require.Nil(t, usersIndexOnUsername20260720120000(t, x))

	_, err := x.Exec("CREATE UNIQUE INDEX UQE_users_username ON users (username) WHERE username IS NOT NULL")
	require.NoError(t, err)
	t.Cleanup(func() {
		// x is the process-global test engine; the index leaks into later tests.
		_, err := x.Exec("DROP INDEX UQE_users_username")
		require.NoError(t, err)
		// The partial sync above dropped the model indexes on users.
		require.NoError(t, x.Sync2(user.GetTables()...))
	})

	require.NoError(t, recreateMissingIndexes20260720120000(x))

	rows, err := x.QueryString("SELECT name, sql FROM sqlite_master WHERE type = 'index' AND tbl_name = 'users' AND sql LIKE '%username%'")
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "UQE_users_username", rows[0]["name"])
	require.Contains(t, rows[0]["sql"], "WHERE username IS NOT NULL")

	dbSchema, err := sqliteSchema20260720120000(x)
	require.NoError(t, err)
	require.Nil(t, indexByName20260720120000(dbSchema.indexesByTable["users"], "UQE_users_username"),
		"a partial unique index does not enforce uniqueness over all rows, so it cannot stand in for the model index")
}

// sqliteSchema20260720120000 only runs on sqlite; usedNames is built from
// Index.IsRegular/XName on every dialect, so this must run on all of them.
func TestRecreateMissingIndexesSkipsWhenNameCollidesOnAnyDialect20260720120000(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	require.NoError(t, x.Sync2(user.GetTables()...))

	require.NoError(t, x.Sync(usersPartial20260720120000{}))
	require.Nil(t, usersIndexOnUsername20260720120000(t, x))

	// Unique on a different column under the model index's name: non-covering
	// (columns differ), so only the name-collision guard stops the CREATE.
	_, err = x.Exec("CREATE UNIQUE INDEX UQE_users_username ON users (email)")
	require.NoError(t, err)
	t.Cleanup(func() {
		// x is the process-global test engine; the index leaks into later tests.
		dropSQL := "DROP INDEX UQE_users_username"
		if x.Dialect().URI().DBType == schemas.MYSQL {
			dropSQL += " ON users"
		}
		_, err := x.Exec(dropSQL)
		require.NoError(t, err)
		// The partial sync above dropped the model indexes on users.
		require.NoError(t, x.Sync2(user.GetTables()...))
	})

	require.NoError(t, recreateMissingIndexes20260720120000(x))
}

func TestSQLiteIndexesReportsNonUniqueIndex20260720120000(t *testing.T) {
	x := sqliteTestEngine20260720120000(t)
	table := "scratch_nonunique20260720120000"
	createScratchTable20260720120000(t, x, table, "id INTEGER PRIMARY KEY, code TEXT")

	_, err := x.Exec("create index IDX_scratch_code20260720120000 on " + table + " (code)")
	require.NoError(t, err)

	dbSchema, err := sqliteSchema20260720120000(x)
	require.NoError(t, err)

	index := indexByName20260720120000(dbSchema.indexesByTable[table], "IDX_scratch_code20260720120000")
	require.NotNil(t, index)
	require.Equal(t, schemas.IndexType, index.Type)
	require.Equal(t, []string{"code"}, index.Cols)
	require.True(t, dbSchema.usedNames["idx_scratch_code20260720120000"])
}

// A UNIQUE column constraint creates sqlite_autoindex_*, whose sqlite_master.sql
// is NULL — there is no DDL to parse, only the pragmas know about it.
func TestSQLiteIndexesReportsImplicitUniqueIndex20260720120000(t *testing.T) {
	x := sqliteTestEngine20260720120000(t)
	table := "scratch_autoindex20260720120000"
	createScratchTable20260720120000(t, x, table, "id INTEGER PRIMARY KEY, code TEXT UNIQUE")

	dbSchema, err := sqliteSchema20260720120000(x)
	require.NoError(t, err)

	index := indexByName20260720120000(dbSchema.indexesByTable[table], "sqlite_autoindex_"+table+"_1")
	require.NotNil(t, index)
	require.Equal(t, schemas.UniqueType, index.Type)
	require.Equal(t, []string{"code"}, index.Cols)
}

func TestSQLiteIndexesReportsAllColumnsInOrder20260720120000(t *testing.T) {
	x := sqliteTestEngine20260720120000(t)
	table := "scratch_multicol20260720120000"
	createScratchTable20260720120000(t, x, table, "id INTEGER PRIMARY KEY, a TEXT, b TEXT")

	// (b, a), not the table's column order, so the assertion pins the index order.
	_, err := x.Exec("create index IDX_scratch_multicol20260720120000 on " + table + " (b, a)")
	require.NoError(t, err)

	dbSchema, err := sqliteSchema20260720120000(x)
	require.NoError(t, err)

	index := indexByName20260720120000(dbSchema.indexesByTable[table], "IDX_scratch_multicol20260720120000")
	require.NotNil(t, index)
	require.Equal(t, []string{"b", "a"}, index.Cols)
}

func TestSQLiteIndexesExcludesPartialIndexButKeepsItsName20260720120000(t *testing.T) {
	x := sqliteTestEngine20260720120000(t)
	table := "scratch_partial20260720120000"
	createScratchTable20260720120000(t, x, table, "id INTEGER PRIMARY KEY, code TEXT")

	_, err := x.Exec("CREATE UNIQUE INDEX UQE_scratch_partial20260720120000 ON " + table + " (code) WHERE code IS NOT NULL")
	require.NoError(t, err)

	dbSchema, err := sqliteSchema20260720120000(x)
	require.NoError(t, err)

	require.Nil(t, indexByName20260720120000(dbSchema.indexesByTable[table], "UQE_scratch_partial20260720120000"),
		"a partial index covers only some rows, so it cannot stand in for a model index")
	require.True(t, dbSchema.usedNames["uqe_scratch_partial20260720120000"],
		"it still occupies the name")
}

func TestSQLiteIndexesExcludesExpressionIndexButKeepsItsName20260720120000(t *testing.T) {
	x := sqliteTestEngine20260720120000(t)
	table := "scratch_expr20260720120000"
	createScratchTable20260720120000(t, x, table, "id INTEGER PRIMARY KEY, code TEXT")

	_, err := x.Exec("CREATE INDEX IDX_scratch_expr20260720120000 ON " + table + " (lower(code))")
	require.NoError(t, err)

	dbSchema, err := sqliteSchema20260720120000(x)
	require.NoError(t, err)

	require.Nil(t, indexByName20260720120000(dbSchema.indexesByTable[table], "IDX_scratch_expr20260720120000"),
		"an expression index covers no column, so it cannot stand in for a model index")
	require.True(t, dbSchema.usedNames["idx_scratch_expr20260720120000"],
		"it still occupies the name")
}

// schemas.Table.GetColumn, which this replaced, folds case; a model index whose
// declared column case differs from the database's must not count as missing.
func TestSQLiteSchemaMatchesColumnsCaseInsensitively20260720120000(t *testing.T) {
	x := sqliteTestEngine20260720120000(t)
	table := "scratch_columns20260720120000"
	createScratchTable20260720120000(t, x, table, "id INTEGER PRIMARY KEY, MixedCase TEXT")

	dbSchema, err := sqliteSchema20260720120000(x)
	require.NoError(t, err)

	require.True(t, columnsExist20260720120000(dbSchema.columnsByTable[table], []string{"mixedcase", "ID"}))
	require.False(t, columnsExist20260720120000(dbSchema.columnsByTable[table], []string{"mixedcase", "added_by_a_later_migration"}))
}

func sqliteTestEngine20260720120000(t *testing.T) *xorm.Engine {
	t.Helper()
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	if x.Dialect().URI().DBType != schemas.SQLITE {
		t.Skip("sqlite-specific index metadata")
	}
	return x
}

func createScratchTable20260720120000(t *testing.T, x *xorm.Engine, name, columns string) {
	t.Helper()
	_, err := x.Exec("CREATE TABLE " + name + " (" + columns + ")")
	require.NoError(t, err)
	t.Cleanup(func() {
		// x is the process-global test engine; the table and its indexes leak into later tests.
		_, err := x.Exec("DROP TABLE " + name)
		require.NoError(t, err)
	})
}

func indexByName20260720120000(indexes []*schemas.Index, name string) *schemas.Index {
	for _, index := range indexes {
		if index.Name == name {
			return index
		}
	}
	return nil
}

func TestPartialSyncKeepsIndexes20260720120000(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	require.NoError(t, x.Sync2(user.GetTables()...))
	require.NoError(t, recreateMissingIndexes20260720120000(x))

	require.NoError(t, partialSync(x, usersPartial20260720120000{}))

	index := usersIndexOnUsername20260720120000(t, x)
	require.NotNil(t, index)
	require.Equal(t, schemas.UniqueType, index.Type)
}
