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

import "code.vikunja.io/web"

// ListUser represents a list <-> user relation
type ListUser struct {
	// The unique, numeric id of this list <-> user relation.
	ID int64 `xorm:"int(11) autoincr not null unique pk" json:"id" param:"namespace"`
	// The user id.
	UserID int64 `xorm:"int(11) not null INDEX" json:"userID" param:"user"`
	// The list id.
	ListID int64 `xorm:"int(11) not null INDEX" json:"-" param:"list"`
	// The right this user has. 0 = Read only, 1 = Read & Write, 2 = Admin. See the docs for more details.
	Right Right `xorm:"int(11) INDEX null" json:"right" valid:"length(0|2)" maximum:"2" default:"0"`

	// A unix timestamp when this relation was created. You cannot change this value.
	Created int64 `xorm:"created not null" json:"created"`
	// A unix timestamp when this relation was last updated. You cannot change this value.
	Updated int64 `xorm:"updated not null" json:"updated"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// TableName is the table name for ListUser
func (ListUser) TableName() string {
	return "users_list"
}

// UserWithRight represents a user in combination with the right it can have on a list/namespace
type UserWithRight struct {
	User  `xorm:"extends"`
	Right Right `json:"right"`
}
