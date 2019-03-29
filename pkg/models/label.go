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
	// The unique, numeric id of this label.
	ID int64 `xorm:"int(11) autoincr not null unique pk" json:"id" param:"label"`
	// The title of the lable. You'll see this one on tasks associated with it.
	Title string `xorm:"varchar(250) not null" json:"title" valid:"runelength(3|250)" minLength:"3" maxLength:"250"`
	// The label description.
	Description string `xorm:"varchar(250) null" json:"description" valid:"runelength(0|250)" maxLength:"250"`
	// The color this label has
	HexColor string `xorm:"varchar(6) null" json:"hex_color" valid:"runelength(0|6)" maxLength:"6"`

	CreatedByID int64 `xorm:"int(11) not null" json:"-"`
	// The user who created this label
	CreatedBy *User `xorm:"-" json:"created_by"`

	// A unix timestamp when this label was created. You cannot change this value.
	Created int64 `xorm:"created not null" json:"created"`
	// A unix timestamp when this label was last updated. You cannot change this value.
	Updated int64 `xorm:"updated not null" json:"updated"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// TableName makes a pretty table name
func (Label) TableName() string {
	return "labels"
}
