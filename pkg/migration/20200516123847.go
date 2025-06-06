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

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type list20200516123847 struct {
	ID          int64  `xorm:"int(11) autoincr not null unique pk" json:"id" param:"list"`
	Title       string `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(3|250)" minLength:"3" maxLength:"250"`
	Description string `xorm:"longtext null" json:"description"`
	Identifier  string `xorm:"varchar(10) null" json:"identifier" valid:"runelength(0|10)" minLength:"0" maxLength:"10"`
	HexColor    string `xorm:"varchar(6) null" json:"hex_color" valid:"runelength(0|6)" maxLength:"6"`
	OwnerID     int64  `xorm:"int(11) INDEX not null" json:"-"`
	NamespaceID int64  `xorm:"int(11) INDEX not null" json:"-" param:"namespace"`
	IsArchived  bool   `xorm:"not null default false" json:"is_archived" query:"is_archived"`
	Created     int64  `xorm:"created not null" json:"created"`
	Updated     int64  `xorm:"updated not null" json:"updated"`
}

func (l *list20200516123847) TableName() string {
	return "list"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20200516123847",
		Description: "Generate a list identifier for each list",
		Migrate: func(tx *xorm.Engine) error {
			lists := []*list20200516123847{}
			err := tx.Find(&lists)
			if err != nil {
				return err
			}

			// Copied and adopted from pkg/models/list.go:374
			generateListIdentifier := func(l *list20200516123847, sess *xorm.Engine) (err error) {
				// The general idea here is to take the title and slice it into pieces, until we found a unique piece.

				var exists = true
				titleSlug := []rune(strings.ReplaceAll(strings.ToUpper(l.Title), " ", ""))

				// We can save at most 10 characters in the db, so we need to ensure it has at most 10 characters
				if len(titleSlug) > 10 {
					titleSlug = titleSlug[0:9]
				}

				var i = 0

				for exists {

					// Prevent endless looping
					if i == len(titleSlug) {
						break
					}

					// Take a random part of the title slug, starting at the beginning
					l.Identifier = string(titleSlug[i:])
					exists, err = sess.
						Where("identifier = ?", l.Identifier).
						And("id != ?", l.ID).
						Exist(&list20200516123847{})
					if err != nil {
						return
					}
					i++
				}
				return nil
			}

			for _, l := range lists {
				if l.Identifier != "" {
					continue
				}

				err := generateListIdentifier(l, tx)
				if err != nil {
					return err
				}
				_, err = tx.Where("id = ?", l.ID).Update(l)
				if err != nil {
					return err
				}
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
