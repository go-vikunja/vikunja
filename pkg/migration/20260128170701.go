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

type UserWebhookSetting20260128170701 struct {
	ID               int64     `xorm:"autoincr not null unique pk"`
	UserID           int64     `xorm:"not null unique(user_notification_type)"`
	NotificationType string    `xorm:"varchar(100) not null unique(user_notification_type)"`
	Enabled          bool      `xorm:"default 1"`
	TargetURL        string    `xorm:"text not null"`
	Created          time.Time `xorm:"created not null"`
	Updated          time.Time `xorm:"updated not null"`
}

func (UserWebhookSetting20260128170701) TableName() string {
	return "user_webhook_settings"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260128170701",
		Description: "Add user_webhook_settings table for per-notification-type webhook configuration",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync(UserWebhookSetting20260128170701{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(UserWebhookSetting20260128170701{})
		},
	})
}
