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
	"embed"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"code.vikunja.io/api/pkg/config"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/require"
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

	var testfiles func(loader *testfixtures.Loader) error
	dir := "fixtures"

	// If fixture table names are specified, load them
	// Otherwise, load all fixtures
	if len(tablenames) > 0 {
		paths := make([]string, len(tablenames))
		for i, name := range tablenames {
			paths[i] = filepath.Join(dir, name+".yml")
		}
		testfiles = testfixtures.Files(paths...)
	} else {
		testfiles = testfixtures.Directory(dir)
	}

	loaderOptions := []func(loader *testfixtures.Loader) error{
		testfixtures.FS(fixturesFS),
		testfixtures.Database(x.DB().DB),
		testfixtures.Dialect(GetDialect()),
		testfixtures.DangerousSkipTestDatabaseCheck(),
		testfixtures.Location(config.GetTimeZone()),
		testfiles,
	}

	if GetDialect() == "postgres" {
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
		// Raw handle: the hook would count our own sequence reset as a write.
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
	rows, err := x.DB().Query(`SELECT 'SELECT SETVAL(' ||
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
