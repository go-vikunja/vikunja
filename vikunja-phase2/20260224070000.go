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

type autoTaskTemplates20260224070000 struct {
	ID              int64      `xorm:"bigint autoincr not null unique pk"`
	OwnerID         int64      `xorm:"bigint not null INDEX"`
	ProjectID       int64      `xorm:"bigint null"`
	Title           string     `xorm:"varchar(250) not null"`
	Description     string     `xorm:"longtext null"`
	Priority        int64      `xorm:"bigint null"`
	HexColor        string     `xorm:"varchar(7) null"`
	LabelIDs        string     `xorm:"json null"`
	AssigneeIDs     string     `xorm:"json null"`
	IntervalValue   int        `xorm:"int not null default 1"`
	IntervalUnit    string     `xorm:"varchar(10) not null default 'days'"`
	StartDate       time.Time  `xorm:"datetime not null"`
	EndDate         *time.Time `xorm:"datetime null"`
	Active          bool       `xorm:"bool not null default true"`
	LastCreatedAt   *time.Time `xorm:"datetime null"`
	LastCompletedAt *time.Time `xorm:"datetime null"`
	NextDueAt       *time.Time `xorm:"datetime null"`
	Created         time.Time  `xorm:"created"`
	Updated         time.Time  `xorm:"updated"`
}

func (autoTaskTemplates20260224070000) TableName() string {
	return "auto_task_templates"
}

type autoTaskTemplateAttachments20260224070000 struct {
	ID          int64     `xorm:"bigint autoincr not null unique pk"`
	TemplateID  int64     `xorm:"bigint not null INDEX"`
	FileID      int64     `xorm:"bigint not null"`
	FileName    string    `xorm:"varchar(250) not null"`
	CreatedByID int64     `xorm:"bigint not null"`
	Created     time.Time `xorm:"created"`
}

func (autoTaskTemplateAttachments20260224070000) TableName() string {
	return "auto_task_template_attachments"
}

type autoTaskLog20260224070000 struct {
	ID            int64     `xorm:"bigint autoincr not null unique pk"`
	TemplateID    int64     `xorm:"bigint not null INDEX"`
	TaskID        int64     `xorm:"bigint not null"`
	TriggerType   string    `xorm:"varchar(20) not null"`
	TriggeredByID int64     `xorm:"bigint null"`
	Created       time.Time `xorm:"created"`
}

func (autoTaskLog20260224070000) TableName() string {
	return "auto_task_log"
}

// Add auto_template_id to the tasks table
type tasks20260224070000 struct {
	AutoTemplateID int64 `xorm:"bigint null INDEX"`
}

func (tasks20260224070000) TableName() string {
	return "tasks"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260224070000",
		Description: "Add auto-task templates, log, attachments tables and task.auto_template_id",
		Migrate: func(tx *xorm.Engine) error {
			if err := tx.Sync(autoTaskTemplates20260224070000{}); err != nil {
				return err
			}
			if err := tx.Sync(autoTaskTemplateAttachments20260224070000{}); err != nil {
				return err
			}
			if err := tx.Sync(autoTaskLog20260224070000{}); err != nil {
				return err
			}
			return tx.Sync(tasks20260224070000{})
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
