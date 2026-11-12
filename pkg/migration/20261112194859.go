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
	"src.techknowlogick.com/xormigrate"
	"xorm.io/builder"
	"xorm.io/xorm"
)

type jobs20261112194859 struct {
	Queue string `xorm:"varchar(100) not null"`
}

func (jobs20261112194859) TableName() string { return "jobs" }

func moveWebhookDelivery20261112194859(tx *xorm.Engine) error {
	_, err := tx.Where(builder.Eq{
		"queue": "webhooks",
		"kind":  "webhook.delivery.webhook.delivery.listener",
	}).And(builder.In("state", "available", "running", "failed")).
		Cols("queue").Update(&jobs20261112194859{Queue: "webhook-delivery"})
	return err
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261112194859",
		Description: "Move pending outbound webhook deliveries to their own queue",
		Migrate:     moveWebhookDelivery20261112194859,
		Rollback:    func(*xorm.Engine) error { return nil },
	})
}
