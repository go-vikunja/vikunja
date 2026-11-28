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

package db

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"code.vikunja.io/api/pkg/config"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
	"xorm.io/xorm/contexts"
	"xorm.io/xorm/schemas"
)

var (
	fixtures    *testfixtures.Loader
	loaderCache = map[string]*testfixtures.Loader{}
	//go:embed fixtures
	fixturesFS embed.FS
)

// InitFixtures initialize test fixtures for a test database
func InitFixtures(tablenames ...string) (err error) {
	loadMu.Lock()
	defer loadMu.Unlock()

	// Building a loader introspects the database, so loaders are reused per table set.
	cacheKey := strings.Join(tablenames, ",")
	if l, ok := loaderCache[cacheKey]; ok {
		fixtures = l
		return nil
	}

	dir := "fixtures"

	// Without table names, load all fixtures
	if len(tablenames) == 0 {
		entries, err := fixturesFS.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			tablenames = append(tablenames, strings.TrimSuffix(entry.Name(), ".yml"))
		}
	}

	files := schemaQualifiedFS{FS: fixturesFS}
	if GetDialect() == "postgres" {
		files.prefix = config.DatabaseSchema.GetString() + "."
	}
	paths := make([]string, len(tablenames))
	for i, name := range tablenames {
		paths[i] = filepath.Join(dir, files.prefix+name+".yml")
	}

	loaderOptions := []func(loader *testfixtures.Loader) error{
		testfixtures.FS(files),
		testfixtures.Database(x.DB().DB),
		testfixtures.Dialect(GetDialect()),
		testfixtures.DangerousSkipTestDatabaseCheck(),
		testfixtures.Location(config.GetTimeZone()),
		testfixtures.Files(paths...),
	}

	// testfixtures would reset every table's counter after each load, one ALTER TABLE per
	// table on MySQL. Postgres gets its sequences reset in LoadFixtures; MySQL moves its
	// counter past explicitly inserted ids on its own.
	if GetDialect() != builder.SQLITE {
		loaderOptions = append(loaderOptions, testfixtures.SkipResetSequences())
	}

	l, err := testfixtures.New(loaderOptions...)
	if err != nil {
		return err
	}
	loaderCache[cacheKey] = l
	fixtures = l
	return nil
}

// testfixtures only reloads tables whose checksum changed, but on Postgres it keys
// the checksums by schema.table. Fixtures named after the bare table never match
// and every table is reloaded on every load, so they are served schema-qualified.
type schemaQualifiedFS struct {
	fs.FS
	prefix string
}

func (f schemaQualifiedFS) Open(name string) (fs.File, error) {
	dir, file := path.Split(name)
	return f.FS.Open(dir + strings.TrimPrefix(file, f.prefix))
}

// Bumped at statement time, so rolled-back writes count too.
var writeGeneration atomic.Uint64

type writeCountingHook struct{}

func (writeCountingHook) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
	if isWriteStatement(c.SQL) || advancesSequence(c.SQL) {
		writeGeneration.Add(1)
	}
	return c.Ctx, nil
}

func (writeCountingHook) AfterProcess(*contexts.ContextHook) error {
	return nil
}

// Sequences move without a write verb.
func advancesSequence(sqlStr string) bool {
	lower := strings.ToLower(sqlStr)
	return strings.Contains(lower, "setval") || strings.Contains(lower, "nextval")
}

// What the last successful load left behind: the loader and the write
// generation it was taken at. A nil loader means the state is unknown.
var (
	loadMu           sync.Mutex
	loadedFixtures   *testfixtures.Loader
	loadedWriteCount uint64
)

// LoadFixtures load fixtures for a test database.
// It is a no-op when the same fixtures were already loaded and nothing has written since.
func LoadFixtures() error {
	loadMu.Lock()
	defer loadMu.Unlock()

	gen := writeGeneration.Load()
	if fixtures != nil && loadedFixtures == fixtures && loadedWriteCount == gen {
		return nil
	}

	loadedFixtures = nil
	// Deferred, not at the end: a load that fails partway still truncated tables.
	defer invalidateAllSessionCaches()

	err := fixtures.Load()
	if err != nil {
		return err
	}

	// Copied from https://github.com/go-gitea/gitea/blob/master/models/test_fixtures.go#L39
	// Now if we're running postgres we need to tell it to update the sequences
	if x.Dialect().URI().DBType == schemas.POSTGRES {
		// database/sql handle: xorm's would count our own sequence reset as a write.
		stmts, err := sequenceResetStatements()
		if err != nil {
			fmt.Printf("Failed to generate sequence update: %v\n", err)
			return err
		}
		for _, value := range stmts {
			if _, err = x.DB().Exec(value); err != nil {
				fmt.Printf("Failed to update sequence: %s Error: %v\n", value, err)
				return err
			}
		}
	}

	loadedFixtures, loadedWriteCount = fixtures, gen
	return nil
}

// LoadAndAssertFixtures loads all fixtures defined before and asserts they are correctly loaded
func LoadAndAssertFixtures(t *testing.T) {
	err := LoadFixtures()
	require.NoError(t, err)
}

func sequenceResetStatements() (stmts []string, err error) {
	rows, err := x.DB().DB.Query(`SELECT 'SELECT SETVAL(' ||
		quote_literal(quote_ident(PGT.schemaname) || '.' || quote_ident(S.relname)) ||
		', COALESCE(MAX(' ||quote_ident(C.attname)|| '), 1) ) FROM ' ||
		quote_ident(PGT.schemaname)|| '.'||quote_ident(T.relname)|| ';'
	 FROM pg_class AS S,
	      pg_depend AS D,
	      pg_class AS T,
	      pg_attribute AS C,
	      pg_tables AS PGT
	 WHERE S.relkind = 'S'
	     AND S.oid = D.objid
	     AND D.refobjid = T.oid
	     AND D.refobjid = C.attrelid
	     AND D.refobjsubid = C.attnum
	     AND T.relname = PGT.tablename
	 ORDER BY S.relname;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var stmt string
		if err = rows.Scan(&stmt); err != nil {
			return nil, err
		}
		stmts = append(stmts, stmt)
	}
	return stmts, rows.Err()
}
