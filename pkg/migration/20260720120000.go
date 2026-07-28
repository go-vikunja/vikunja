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

	"code.vikunja.io/api/pkg/log"

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
		return fmt.Errorf("could not read the database schema: %w", err)
	}
	dbTableByName := make(map[string]*schemas.Table, len(dbTables))
	for _, t := range dbTables {
		dbTableByName[t.Name] = t
	}
	dbIndexes, err := dbIndexes20260720120000(tx, dbTables)
	if err != nil {
		return err
	}

	for _, bean := range schemaBeans() {
		modelTable, err := tx.TableInfo(bean)
		if err != nil {
			return fmt.Errorf("could not read the model schema of %T: %w", bean, err)
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
			if indexCoveringColsExists20260720120000(dbIndexes.byTable[dbTable.Name], index) {
				continue
			}
			// Creating it would fail with "index already exists" and abort startup —
			// the very failure mode this migration exists to prevent.
			if indexName := index.XName(modelTable.Name); dbIndexes.usedNames[strings.ToLower(indexName)] {
				log.Warningf(
					"Could not recreate the index on %s (%s): another index already uses the name %s. Drop or rename it, then run this manually: %s",
					modelTable.Name, strings.Join(index.Cols, ", "), indexName, tx.Dialect().CreateIndexSQL(modelTable.Name, index))
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

type dbIndexes20260720120000Result struct {
	byTable map[string][]*schemas.Index
	// Keyed by lowercased name. Names are case-insensitive on mysql but
	// case-sensitive on postgres (xorm quotes them there); lowercasing
	// over-approximates on postgres and mysql's per-table naming, which only
	// ever costs an extra warn-and-skip, never a missed collision.
	usedNames map[string]bool
}

func dbIndexes20260720120000(tx *xorm.Engine, dbTables []*schemas.Table) (*dbIndexes20260720120000Result, error) {
	if tx.Dialect().URI().DBType == schemas.SQLITE {
		return sqliteIndexes20260720120000(tx)
	}
	result := &dbIndexes20260720120000Result{
		byTable:   make(map[string][]*schemas.Index, len(dbTables)),
		usedNames: make(map[string]bool),
	}
	for _, dbTable := range dbTables {
		for _, index := range dbTable.Indexes {
			result.byTable[dbTable.Name] = append(result.byTable[dbTable.Name], index)
			// DBMetas strips the IDX_/UQE_ prefix it recognizes and flags those as
			// regular; XName puts it back.
			name := index.Name
			if index.IsRegular {
				name = index.XName(dbTable.Name)
			}
			result.usedNames[strings.ToLower(name)] = true
		}
	}
	return result, nil
}

// One row per index column, joined over sqlite's pragma table-valued functions.
// xorm's sqlite3 dialect instead parses the DDL in sqlite_master by looking for
// the literal uppercase "INDEX" and "ON", so it silently skips every index older
// Vikunja migrations created with lowercase `create index` SQL (#3313) — and the
// implicit indexes of UNIQUE constraints, which have no DDL at all.
func sqliteIndexes20260720120000(tx *xorm.Engine) (*dbIndexes20260720120000Result, error) {
	rows, err := tx.QueryString(`SELECT m.name AS tbl, il.name AS idx, il."unique" AS uniq, il.partial AS part, ii.name AS col, (ii.name IS NULL) AS is_expr
FROM sqlite_master m
JOIN pragma_index_list(m.name) il
JOIN pragma_index_info(il.name) ii
WHERE m.type = 'table'
ORDER BY m.name, il.name, ii.seqno`)
	if err != nil {
		return nil, fmt.Errorf("could not read the sqlite index metadata: %w", err)
	}

	type sqliteIndex struct {
		table    string
		unusable bool
		index    *schemas.Index
	}
	seen := make(map[string]*sqliteIndex, len(rows))
	ordered := make([]*sqliteIndex, 0, len(rows))
	result := &dbIndexes20260720120000Result{
		byTable:   make(map[string][]*schemas.Index),
		usedNames: make(map[string]bool, len(rows)),
	}

	for _, row := range rows {
		name := row["idx"]
		result.usedNames[strings.ToLower(name)] = true

		entry, ok := seen[name]
		if !ok {
			indexType := schemas.IndexType
			if row["uniq"] == "1" {
				indexType = schemas.UniqueType
			}
			entry = &sqliteIndex{table: row["tbl"], index: &schemas.Index{Name: name, Type: indexType}}
			seen[name] = entry
			ordered = append(ordered, entry)
		}
		// A partial index covers only some rows and an expression index no column at
		// all, so neither can stand in for a model index — but both still occupy the name.
		if row["part"] == "1" || row["is_expr"] == "1" {
			entry.unusable = true
			continue
		}
		entry.index.Cols = append(entry.index.Cols, row["col"])
	}

	for _, entry := range ordered {
		if entry.unusable {
			continue
		}
		result.byTable[entry.table] = append(result.byTable[entry.table], entry.index)
	}
	return result, nil
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
		return fmt.Errorf("could not check %s (%s) for duplicate values: %w", table, strings.Join(cols, ", "), err)
	}
	if len(rows) > 0 {
		// Some unique-indexed columns hold secrets (token_hash, oauth codes, ...) — never log the values.
		return fmt.Errorf(
			"cannot recreate the unique index on %s (%s) because %d sets of duplicate values exist — remove the duplicates manually, then restart Vikunja",
			table, strings.Join(cols, ", "), len(rows))
	}
	return nil
}
