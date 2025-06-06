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
	"bytes"
	templatehtml "html/template"
	"strings"

	"github.com/yuin/goldmark"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

func convertMarkdownToHTML(input string) (output string, err error) {
	md := []byte(templatehtml.HTMLEscapeString(input))
	var buf bytes.Buffer
	err = goldmark.Convert(md, &buf)
	if err != nil {
		return
	}
	//#nosec - the html is escaped few lines before
	return buf.String(), nil
}

func convertDescription(tx *xorm.Engine, table string, column string) (err error) {
	items := []map[string]interface{}{}
	err = tx.Table(table).
		Select("id, " + column).
		Find(&items)
	if err != nil {
		return
	}

	for _, task := range items {
		if task[column] == "" || strings.HasPrefix(task[column].(string), "<") {
			continue
		}

		task[column], err = convertMarkdownToHTML(task[column].(string))
		if err != nil {
			return
		}
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
		ID:          "20231022144641",
		Description: "Convert all descriptions to HTML",
		Migrate: func(tx *xorm.Engine) (err error) {

			for _, table := range []string{
				"tasks",
				"labels",
				"projects",
				"saved_filters",
				"teams",
			} {
				err = convertDescription(tx, table, "description")
				if err != nil {
					return
				}
			}

			err = convertDescription(tx, "task_comments", "comment")
			return
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
