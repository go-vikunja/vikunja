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

// NamespaceUser represents a namespace <-> user relation
type NamespaceUser struct {
	ID          int64     `xorm:"int(11) autoincr not null unique pk" json:"id" param:"namespace"`
	UserID      int64     `xorm:"int(11) not null INDEX" json:"user_id" param:"user"`
	NamespaceID int64     `xorm:"int(11) not null INDEX" json:"namespace_id" param:"namespace"`
	Right       UserRight `xorm:"int(11) INDEX" json:"right" valid:"length(0|2)"`

	Created int64 `xorm:"created" json:"created"`
	Updated int64 `xorm:"updated" json:"updated"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// TableName is the table name for NamespaceUser
func (NamespaceUser) TableName() string {
	return "users_namespace"
}
