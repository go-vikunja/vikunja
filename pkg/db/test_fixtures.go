// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package db

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm/schemas"
)

var fixtures *testfixtures.Loader

// InitFixtures initialize test fixtures for a test database
func InitFixtures(tablenames ...string) (err error) {

	var testfiles func(loader *testfixtures.Loader) error
	dir := filepath.Join(config.ServiceRootpath.GetString(), "pkg", "db", "fixtures")

	// If fixture table names are specified, load them
	// Otherwise, load all fixtures
	if len(tablenames) > 0 {
		for i, name := range tablenames {
			tablenames[i] = filepath.Join(dir, name+".yml")
		}
		testfiles = testfixtures.Files(tablenames...)
	} else {
		testfiles = testfixtures.Directory(dir)
	}

	loaderOptions := []func(loader *testfixtures.Loader) error{
		testfixtures.Database(x.DB().DB),
		testfixtures.Dialect(config.DatabaseType.GetString()),
		testfixtures.DangerousSkipTestDatabaseCheck(),
		testfixtures.Location(config.GetTimeZone()),
		testfiles,
	}

	if config.DatabaseType.GetString() == "postgres" {
		loaderOptions = append(loaderOptions,
			testfixtures.SkipResetSequences(),
			testfixtures.UseAlterConstraint(),
			testfixtures.SkipTableChecksumComputation(),
		)
	}

	fixtures, err = testfixtures.New(loaderOptions...)
	return err
}

func InitFixturesWithT(t *testing.T, tablenames ...string) (err error) {
	startTime := time.Now()
	
	// Track individual timings
	var setupTime, optionsTime, fixturesCreationTime time.Duration
	
	// Setup phase
	setupStart := time.Now()
	var testfiles func(loader *testfixtures.Loader) error
	dir := filepath.Join(config.ServiceRootpath.GetString(), "pkg", "db", "fixtures")

	// If fixture table names are specified, load them
	// Otherwise, load all fixtures
	if len(tablenames) > 0 {
		for i, name := range tablenames {
			tablenames[i] = filepath.Join(dir, name+".yml")
		}
		testfiles = testfixtures.Files(tablenames...)
	} else {
		testfiles = testfixtures.Directory(dir)
	}
	setupTime = time.Since(setupStart)
	
	// Options configuration phase
	optionsStart := time.Now()
	loaderOptions := []func(loader *testfixtures.Loader) error{
		testfixtures.Database(x.DB().DB),
		testfixtures.Dialect(config.DatabaseType.GetString()),
		testfixtures.DangerousSkipTestDatabaseCheck(),
		testfixtures.Location(config.GetTimeZone()),
		testfiles,
	}

	if config.DatabaseType.GetString() == "postgres" {
		loaderOptions = append(loaderOptions,
			testfixtures.SkipResetSequences(),
			testfixtures.UseAlterConstraint(),
			testfixtures.SkipTableChecksumComputation(),
		)
	}
	optionsTime = time.Since(optionsStart)
	
	// Fixtures creation phase
	fixturesStart := time.Now()
	fixtures, err = testfixtures.New(loaderOptions...)
	fixturesCreationTime = time.Since(fixturesStart)
	
	// Log all timings in one statement
	totalTime := time.Since(startTime)
	t.Logf("Fixtures setup: total=%v (setup=%v, options=%v, creation=%v)", 
		totalTime, setupTime, optionsTime, fixturesCreationTime)
	
	return err
}

// LoadFixtures load fixtures for a test database
func LoadFixtures() error {
	err := fixtures.Load()
	if err != nil {
		return err
	}

	// Copied from https://github.com/go-gitea/gitea/blob/master/models/test_fixtures.go#L39
	// Now if we're running postgres we need to tell it to update the sequences
	if x.Dialect().URI().DBType == schemas.POSTGRES {
		results, err := x.QueryString(`SELECT 'SELECT SETVAL(' ||
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
			fmt.Printf("Failed to generate sequence update: %v\n", err)
			return err
		}
		for _, r := range results {
			for _, value := range r {
				_, err = x.Exec(value)
				if err != nil {
					fmt.Printf("Failed to update sequence: %s Error: %v\n", value, err)
					return err
				}
			}
		}
	}
	return nil
}

// LoadAndAssertFixtures loads all fixtures defined before and asserts they are correctly loaded
func LoadAndAssertFixtures(t *testing.T) {
	err := LoadFixtures()
	require.NoError(t, err)
}
