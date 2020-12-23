// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
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
	"encoding/json"

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

	for _, content := range contents {
		if _, err := x.Table(table).Insert(content); err != nil {
			return err
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
