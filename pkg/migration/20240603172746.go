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
	"regexp"
	"strings"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

func convertChecklistInDescription(tx *xorm.Engine, table string, column string) (err error) {
	items := []map[string]interface{}{}
	err = tx.Table(table).
		Select("id, " + column).
		Find(&items)
	if err != nil {
		return
	}

	for _, item := range items {
		if !strings.Contains(item[column].(string), "<li>[") {
			continue
		}

		var re = regexp.MustCompile(`<ul>(\n)?<li>`)
		item[column] = re.ReplaceAllString(item[column].(string), `<ul data-type="taskList"><li>`)

		item[column] = strings.ReplaceAll(item[column].(string), "<li>[ ] ", `<li data-checked="false" data-type="taskItem"><label><input type="checkbox"><span></span></label>`)
		item[column] = strings.ReplaceAll(item[column].(string), "<li>[x] ", `<li data-checked="true" data-type="taskItem"><label><input type="checkbox" checked="checked"><span></span></label>`)

		_, err = tx.Where("id = ?", item["id"]).
			Table(table).
			Cols(column).
			Update(item)
		if err != nil {
			return
		}
	}

	return nil
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20240603172746",
		Description: "Convert unconverted checklists to proper html",
		Migrate: func(tx *xorm.Engine) (err error) {
			for _, table := range []string{
				"tasks",
				"labels",
				"projects",
				"saved_filters",
				"teams",
			} {
				err = convertChecklistInDescription(tx, table, "description")
				if err != nil {
					return
				}
			}

			err = convertChecklistInDescription(tx, "task_comments", "comment")
			return
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
