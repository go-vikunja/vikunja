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
	"strings"

	"mvdan.cc/xurls/v2"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

func convertLinksToHTMLElements(input string) (output string) {

	links := xurls.Strict().FindAllString(input, -1)

	if len(links) == 0 {
		return input
	}

	unique := make(map[string]bool)
	for _, link := range links {
		unique[link] = true
	}

	for link := range unique {

		if strings.Contains(input, `href="`+link) {
			continue
		}

		input = strings.ReplaceAll(input, link, `<a href="`+link+`">`+link+`</a>`)
	}

	return input
}

func convertDescriptionToLinks(tx *xorm.Engine, table string, column string) (err error) {
	items := []map[string]interface{}{}
	err = tx.Table(table).
		Select("id, " + column).
		Find(&items)
	if err != nil {
		return
	}

	for _, task := range items {
		if task[column] == "" || task[column] == "<p></p>" {
			continue
		}

		task[column] = convertLinksToHTMLElements(task[column].(string))
		_, err = tx.Where("id = ?", task["id"]).
			Table(table).
			Cols(column).
			Update(task)
		if err != nil {
			return
		}
	}

	return
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20240114224713",
		Description: "Convert all non-html links to html a href",
		Migrate: func(tx *xorm.Engine) (err error) {

			for _, table := range []string{
				"tasks",
				"labels",
				"projects",
				"saved_filters",
				"teams",
			} {
				err = convertDescriptionToLinks(tx, table, "description")
				if err != nil {
					return
				}
			}

			err = convertDescriptionToLinks(tx, "task_comments", "comment")
			return
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
