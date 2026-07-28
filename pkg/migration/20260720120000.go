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
	"fmt"
	"strings"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260720120000",
		Description: "Recreate indexes dropped by partial-struct sync migrations",
		Migrate:     recreateMissingIndexes20260720120000,
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}

// Migrations which synced a partial struct made xorm drop every index the
// struct didn't declare (#3244). Recreates model-declared indexes that are
// missing, matching by column set because converted DBs (pgloader) carry
// equivalent indexes under different names.
func recreateMissingIndexes20260720120000(tx *xorm.Engine) error {
	dbTables, err := tx.DBMetas()
	if err != nil {
		return err
	}
	dbTableByName := make(map[string]*schemas.Table, len(dbTables))
	for _, t := range dbTables {
		dbTableByName[t.Name] = t
	}
	dbIndexesByTable, err := dbIndexes20260720120000(tx, dbTables)
	if err != nil {
		return err
	}

	for _, bean := range schemaBeans() {
		modelTable, err := tx.TableInfo(bean)
		if err != nil {
			return err
		}
		dbTable, exists := dbTableByName[modelTable.Name]
		if !exists {
			continue
		}
		for _, index := range modelTable.Indexes {
			// Columns from migrations that run after this one don't exist yet.
			if !columnsExist20260720120000(dbTable, index.Cols) {
				continue
			}
			if indexCoveringColsExists20260720120000(dbIndexesByTable[dbTable.Name], index) {
				continue
			}
			if index.Type == schemas.UniqueType {
				if err := ensureNoDuplicates20260720120000(tx, modelTable.Name, index.Cols); err != nil {
					return err
				}
			}
			if _, err := tx.Exec(tx.Dialect().CreateIndexSQL(modelTable.Name, index)); err != nil {
				return fmt.Errorf("could not recreate index on %s (%s): %w", modelTable.Name, strings.Join(index.Cols, ", "), err)
			}
		}
	}
	return nil
}

func dbIndexes20260720120000(tx *xorm.Engine, dbTables []*schemas.Table) (map[string][]*schemas.Index, error) {
	if tx.Dialect().URI().DBType == schemas.SQLITE {
		return sqliteIndexes20260720120000(tx)
	}
	byTable := make(map[string][]*schemas.Index, len(dbTables))
	for _, dbTable := range dbTables {
		for _, index := range dbTable.Indexes {
			byTable[dbTable.Name] = append(byTable[dbTable.Name], index)
		}
	}
	return byTable, nil
}

// xorm's sqlite3 dialect parses index definitions by looking for the literal
// uppercase "INDEX" and "ON", so it silently skips every index older Vikunja
// migrations created with lowercase `create index` SQL (#3313) — and skips the
// implicit indexes of UNIQUE column constraints, which have no SQL at all.
// Read sqlite_master instead of trusting DBMetas.
func sqliteIndexes20260720120000(tx *xorm.Engine) (map[string][]*schemas.Index, error) {
	rows, err := tx.QueryString("SELECT name, tbl_name, sql FROM sqlite_master WHERE type = 'index'")
	if err != nil {
		return nil, err
	}

	byTable := make(map[string][]*schemas.Index, len(rows))
	for _, row := range rows {
		createSQL := strings.ToUpper(strings.TrimSpace(row["sql"]))
		// A partial index only covers some rows, so it can't stand in for a full one.
		if strings.Contains(createSQL, " WHERE ") {
			continue
		}

		cols, err := sqliteIndexCols20260720120000(tx, row["name"])
		if err != nil {
			return nil, err
		}
		if len(cols) == 0 {
			continue
		}

		index := &schemas.Index{Name: row["name"], Type: schemas.IndexType, Cols: cols}
		// No SQL means sqlite generated the index for a UNIQUE or PK constraint.
		words := strings.Fields(createSQL)
		if len(words) < 2 || words[1] == "UNIQUE" {
			index.Type = schemas.UniqueType
		}
		byTable[row["tbl_name"]] = append(byTable[row["tbl_name"]], index)
	}
	return byTable, nil
}

func sqliteIndexCols20260720120000(tx *xorm.Engine, indexName string) ([]string, error) {
	rows, err := tx.QueryString("PRAGMA index_info(" + tx.Quote(indexName) + ")")
	if err != nil {
		return nil, err
	}
	cols := make([]string, 0, len(rows))
	for _, row := range rows {
		// Expression indexes report a NULL column name; they can't cover a model index.
		if row["name"] == "" {
			return nil, nil
		}
		cols = append(cols, row["name"])
	}
	return cols, nil
}

func columnsExist20260720120000(dbTable *schemas.Table, cols []string) bool {
	for _, col := range cols {
		if dbTable.GetColumn(col) == nil {
			return false
		}
	}
	return true
}

// Not schemas.Index.Equal: it's case-sensitive, doesn't trim, and requires exact
// type match, but a unique db index should also satisfy a non-unique model index.
func indexCoveringColsExists20260720120000(dbIndexes []*schemas.Index, index *schemas.Index) bool {
	for _, dbIndex := range dbIndexes {
		if index.Type == schemas.UniqueType && dbIndex.Type != schemas.UniqueType {
			continue
		}
		if len(dbIndex.Cols) != len(index.Cols) {
			continue
		}
		// Match columns as an unordered set: pgloader-converted DBs may list the
		// same composite columns in a different order.
		match := true
		for _, col := range index.Cols {
			if !colInList20260720120000(dbIndex.Cols, col) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func colInList20260720120000(cols []string, want string) bool {
	for _, col := range cols {
		if strings.EqualFold(strings.TrimSpace(col), want) {
			return true
		}
	}
	return false
}

func ensureNoDuplicates20260720120000(tx *xorm.Engine, table string, cols []string) error {
	quoted := make([]string, 0, len(cols))
	notNull := make([]string, 0, len(cols))
	for _, col := range cols {
		quoted = append(quoted, tx.Quote(col))
		// Unique indexes allow multiple NULLs, GROUP BY does not.
		notNull = append(notNull, tx.Quote(col)+" IS NOT NULL")
	}
	query := "SELECT " + strings.Join(quoted, ", ") + " FROM " + tx.Quote(table) +
		" WHERE " + strings.Join(notNull, " AND ") +
		" GROUP BY " + strings.Join(quoted, ", ") +
		" HAVING COUNT(*) > 1"
	rows, err := tx.QueryString(query)
	if err != nil {
		return err
	}
	if len(rows) > 0 {
		// Some unique-indexed columns hold secrets (token_hash, oauth codes, ...) — never log the values.
		return fmt.Errorf(
			"cannot recreate the unique index on %s (%s) because %d sets of duplicate values exist — remove the duplicates manually, then restart Vikunja",
			table, strings.Join(cols, ", "), len(rows))
	}
	return nil
}
