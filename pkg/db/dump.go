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
	"encoding/json"
	"fmt"
	"strings"

	"code.vikunja.io/api/pkg/log"

	"xorm.io/xorm/schemas"
)

// Dump dumps all database tables
func Dump() (data map[string][]byte, err error) {
	tables, err := x.DBMetas()
	if err != nil {
		return
	}

	data = make(map[string][]byte, len(tables))
	for _, table := range tables {
		entries := []map[string]interface{}{}
		err := x.Table(table.Name).Find(&entries)
		if err != nil {
			return nil, err
		}
		data[table.Name], err = json.Marshal(entries)
		if err != nil {
			return nil, err
		}
	}

	return
}

// Restore restores a table with all its entries
func Restore(table string, contents []map[string]interface{}) (err error) {
	if _, err := x.IsTableExist(table); err != nil {
		return err
	}

	meta, err := x.DBMetas()
	if err != nil {
		return err
	}

	var metaForCurrentTable *schemas.Table
	for _, m := range meta {
		if m.Name == table {
			metaForCurrentTable = m
			break
		}
	}

	if metaForCurrentTable == nil {
		log.Fatalf("Could not find table definition for table %s", table)
	}

	for _, content := range contents {
		// For all tables, use XORM Insert() with proper type conversion
		if false {
			// Build column names and placeholders
			cols := make([]string, 0, len(content))
			placeholders := make([]string, 0, len(content))
			values := make([]interface{}, 0, len(content))
			i := 1

			for colName, value := range content {
				col := metaForCurrentTable.GetColumn(colName)
				if col == nil {
					log.Debugf("Skipping unknown column '%s' in table '%s'", colName, table)
					continue
				}

				// Skip nil JSON columns - they'll use database default (NULL)
				if value == nil && col.SQLType.Name == "JSON" {
					log.Debugf("Skipping nil JSON column '%s' in table '%s'", colName, table)
					continue
				}

				cols = append(cols, colName)

				// Handle JSON columns specially
				if rawMsg, ok := value.(json.RawMessage); ok {
					// Pass as []byte without cast - pq driver should handle json.RawMessage
					placeholders = append(placeholders, fmt.Sprintf("$%d", i))
					values = append(values, []byte(rawMsg))
				} else if byteVal, ok := value.([]byte); ok && col.SQLType.Name == "JSON" {
					placeholders = append(placeholders, fmt.Sprintf("$%d", i))
					values = append(values, byteVal)
				} else {
					// Handle date/time fields
					strVal, is := value.(string)
					if is && col.SQLType.IsTime() {
						if strVal == "" || strings.HasPrefix(strVal, "0001-") {
							value = nil
						}
						// If it's a non-null time string, keep it as string - PostgreSQL will parse it
					}

					// Handle other types - convert float64 to appropriate type
					if floatVal, ok := value.(float64); ok {
						// Check if column is boolean type
						if strings.ToUpper(col.SQLType.Name) == "BOOL" || strings.ToUpper(col.SQLType.Name) == "BOOLEAN" {
							value = floatVal != 0
						} else if col.SQLType.IsNumeric() && !strings.Contains(strings.ToUpper(col.SQLType.Name), "FLOAT") && !strings.Contains(strings.ToUpper(col.SQLType.Name), "DOUBLE") {
							// Check if column is integer type
							value = int64(floatVal)
						}
					}

					placeholders = append(placeholders, fmt.Sprintf("$%d", i))
					values = append(values, value)
				}
				i++
			}

			if len(cols) > 0 {
				sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
					table,
					strings.Join(cols, ", "),
					strings.Join(placeholders, ", "))

				// Debug: log final parameter types before execution
				if table == "project_views" {
					log.Infof("DEBUG: Final values array for %s:", table)
					for idx, v := range values {
						if byteVal, ok := v.([]byte); ok {
							log.Infof("  [%d] (%s) []byte: %s", idx, cols[idx], string(byteVal))
						} else {
							log.Infof("  [%d] (%s) %T: %v", idx, cols[idx], v, v)
						}
					}
					log.Infof("DEBUG: SQL: %s", sql)
				}

				// Use prepared statement for better parameter handling
				stmt, err := x.DB().DB.Prepare(sql)
				if err != nil {
					log.Errorf("Failed to prepare statement for %s. SQL: %s, Error: %v", table, sql, err)
					return err
				}
				defer stmt.Close()

				_, err = stmt.Exec(values...)
				if err != nil {
					log.Errorf("Failed to insert into %s. SQL: %s, Error: %v", table, sql, err)
					return err
				}
			}
		} else {
			// Original logic for non-JSON tables
			for colName, value := range content {
				col := metaForCurrentTable.GetColumn(colName)
				if col == nil {
					log.Debugf("Skipping unknown column '%s' in table '%s'", colName, table)
					delete(content, colName)
					continue
				}

				// Convert json.RawMessage to []byte for PostgreSQL JSON columns
				// PostgreSQL driver needs JSON as bytes, not as interface{}
				if rawMsg, ok := value.(json.RawMessage); ok {
					content[colName] = []byte(rawMsg)
					continue
				}

				// Handle empty string for JSON columns - convert to NULL
				strVal, is := value.(string)
				if is && col.SQLType.Name == "JSON" && strVal == "" {
					content[colName] = nil
					continue
				}

				if is && col.SQLType.IsTime() && (strVal == "" || strings.HasPrefix(strVal, "0001-")) {
					content[colName] = nil
				}
			}

			if _, err := x.Table(table).Insert(content); err != nil {
				return err
			}
		}
	}

	if Type() == schemas.POSTGRES {
		idSequence := table + "_id_seq"
		_, err = x.Query("SELECT setval('" + idSequence + "', COALESCE(MAX(id), 1) )")
		if err != nil {
			log.Warningf("Could not reset id sequence for %s: %s", idSequence, err)
			err = nil
		}
	}

	return
}

// RestoreAndTruncate removes all content from the table before restoring it from the contents map
func RestoreAndTruncate(table string, contents []map[string]interface{}) (err error) {
	if _, err := x.IsTableExist(table); err != nil {
		return err
	}

	if x.Dialect().URI().DBType == schemas.SQLITE {
		if _, err := x.Query("DELETE FROM " + table); err != nil {
			return err
		}
	} else {
		if _, err := x.Query("TRUNCATE TABLE ?", table); err != nil {
			return err
		}
	}

	return Restore(table, contents)
}
