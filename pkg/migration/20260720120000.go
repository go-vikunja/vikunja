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
			if indexCoveringColsExists20260720120000(dbTable, index) {
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

func columnsExist20260720120000(dbTable *schemas.Table, cols []string) bool {
	for _, col := range cols {
		if dbTable.GetColumn(col) == nil {
			return false
		}
	}
	return true
}

func indexCoveringColsExists20260720120000(dbTable *schemas.Table, index *schemas.Index) bool {
	for _, dbIndex := range dbTable.Indexes {
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
