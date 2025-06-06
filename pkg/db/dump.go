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
		for colName, value := range content {
			// Date fields might get restored as 0001-01-01 from null dates. This can have unintended side-effects like
			// users being scheduled for deletion after a restore.
			// To avoid this, we set these dates to nil so that they'll end up as null in the db.
			col := metaForCurrentTable.GetColumn(colName)
			strVal, is := value.(string)
			if is && col.SQLType.IsTime() && (strVal == "" || strings.HasPrefix(strVal, "0001-")) {
				content[colName] = nil
			}
		}

		if _, err := x.Table(table).Insert(content); err != nil {
			return err
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
