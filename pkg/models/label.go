//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/web"
)

// Label represents a label
type Label struct {
	ID          int64  `xorm:"int(11) autoincr not null unique pk" json:"id" param:"label"`
	Title       string `xorm:"varchar(250) not null" json:"title" valid:"runelength(3|250)"`
	Description string `xorm:"varchar(250)" json:"description" valid:"runelength(0|250)"`
	HexColor    string `xorm:"varchar(6)" json:"hex_color" valid:"runelength(0|6)"`

	CreatedByID int64 `xorm:"int(11) not null" json:"-"`
	CreatedBy   *User `xorm:"-" json:"created_by"`

	Created int64 `xorm:"created" json:"created"`
	Updated int64 `xorm:"updated" json:"updated"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// TableName makes a pretty table name
func (Label) TableName() string {
	return "labels"
}

// LabelTask represents a relation between a label and a task
type LabelTask struct {
	ID      int64 `xorm:"int(11) autoincr not null unique pk" json:"id"`
	TaskID  int64 `xorm:"int(11) INDEX not null" json:"-" param:"listtask"`
	LabelID int64 `xorm:"int(11) INDEX not null" json:"label_id" param:"label"`
	Created int64 `xorm:"created" json:"created"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// TableName makes a pretty table name
func (LabelTask) TableName() string {
	return "label_task"
}
