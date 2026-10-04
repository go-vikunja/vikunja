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

type Risk20261003100000 struct {
	ID             int64     `xorm:"bigint autoincr not null unique pk"`
	ProjectID      int64     `xorm:"bigint not null index"`
	Title          string    `xorm:"varchar(250) not null"`
	Description    string    `xorm:"text null"`
	Category       string    `xorm:"varchar(100) null"`
	Probability    int       `xorm:"int not null default 3"`
	Impact         int       `xorm:"int not null default 3"`
	OwnerID        int64     `xorm:"bigint null index"`
	Mitigation     string    `xorm:"text null"`
	Contingency    string    `xorm:"text null"`
	Status         string    `xorm:"varchar(20) not null default 'open' index"`
	IdentifiedDate time.Time `xorm:"DATETIME null"`
	DueDate        time.Time `xorm:"DATETIME null index"`
	ClosedAt       time.Time `xorm:"DATETIME null"`
	ClosedByID     int64     `xorm:"bigint null"`
	Resolution     string    `xorm:"text null"`
	CreatedByID    int64     `xorm:"bigint not null"`
	Created        time.Time `xorm:"created not null"`
	Updated        time.Time `xorm:"updated not null"`
}

func (Risk20261003100000) TableName() string { return "risks" }

type RiskStatusHistory20261003100000 struct {
	ID          int64     `xorm:"bigint autoincr not null unique pk"`
	RiskID      int64     `xorm:"bigint not null index"`
	FromStatus  string    `xorm:"varchar(20) null"`
	ToStatus    string    `xorm:"varchar(20) not null"`
	Note        string    `xorm:"text null"`
	ChangedByID int64     `xorm:"bigint not null"`
	Created     time.Time `xorm:"created not null"`
}

func (RiskStatusHistory20261003100000) TableName() string { return "risk_status_history" }

func addRisks20261003100000(tx *xorm.Engine) error {
	// brand-new tables, nothing to drop
	return tx.Sync(Risk20261003100000{}, RiskStatusHistory20261003100000{}) //nolint:forbidigo // brand-new tables
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261003100000",
		Description: "Add the risks and risk_status_history tables",
		Migrate:     addRisks20261003100000,
		Rollback: func(_ *xorm.Engine) error {
			return nil
		},
	})
}
