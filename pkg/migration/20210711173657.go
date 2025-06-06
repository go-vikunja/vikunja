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
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type user20210711173657 struct {
	ID                 int64  `xorm:"bigint autoincr not null unique pk" json:"id"`
	PasswordResetToken string `xorm:"varchar(450) null" json:"-"`
	EmailConfirmToken  string `xorm:"varchar(450) null" json:"-"`
}

func (u user20210711173657) TableName() string {
	return "users"
}

type userTokens20210711173657 struct {
	ID      int64     `xorm:"bigint autoincr not null unique pk"`
	UserID  int64     `xorm:"not null"`
	Token   string    `xorm:"varchar(450) not null index"`
	Kind    int       `xorm:"not null"`
	Created time.Time `xorm:"created not null"`
}

func (userTokens20210711173657) TableName() string {
	return "user_tokens"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20210711173657",
		Description: "Add user tokens table",
		Migrate: func(tx *xorm.Engine) error {
			_ = tx.DropTables(&userTokens20210711173657{}) // Allow running this migration multiple times

			err := tx.Sync2(userTokens20210711173657{})
			if err != nil {
				return err
			}

			users := []*user20210711173657{}
			err = tx.Where(`password_reset_token != '' OR email_confirm_token != ''`).Find(&users)
			if err != nil {
				return err
			}

			const tokenPasswordReset = 1
			const tokenEmailConfirm = 2

			for _, user := range users {
				if user.PasswordResetToken != "" {
					_, err = tx.Insert(&userTokens20210711173657{
						UserID: user.ID,
						Token:  user.PasswordResetToken,
						Kind:   tokenPasswordReset,
					})
					if err != nil {
						return err
					}
				}

				if user.EmailConfirmToken != "" {
					_, err = tx.Insert(&userTokens20210711173657{
						UserID: user.ID,
						Token:  user.EmailConfirmToken,
						Kind:   tokenEmailConfirm,
					})
					if err != nil {
						return err
					}
				}
			}

			err = dropTableColum(tx, "users", "password_reset_token")
			if err != nil {
				return err
			}
			return dropTableColum(tx, "users", "email_confirm_token")
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
