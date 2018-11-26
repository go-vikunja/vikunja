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

// TeamNamespace defines the relationship between a Team and a Namespace
type TeamNamespace struct {
	ID          int64     `xorm:"int(11) autoincr not null unique pk" json:"id"`
	TeamID      int64     `xorm:"int(11) not null INDEX" json:"team_id" param:"team"`
	NamespaceID int64     `xorm:"int(11) not null INDEX" json:"namespace_id" param:"namespace"`
	Right       TeamRight `xorm:"int(11) INDEX" json:"right" valid:"length(0|2)"`

	Created int64 `xorm:"created" json:"created"`
	Updated int64 `xorm:"updated" json:"updated"`

	CRUDable `xorm:"-" json:"-"`
	Rights   `xorm:"-" json:"-"`
}

// TableName makes beautiful table names
func (TeamNamespace) TableName() string {
	return "team_namespaces"
}
