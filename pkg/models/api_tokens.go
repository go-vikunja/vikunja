// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2023 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/web"
	"time"
)

type APIPermissions map[string][]string

type APIToken struct {
	// The unique, numeric id of this api key.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"token"`

	// A human-readable name for this token
	Title string `xorm:"not null" json:"title"`
	// The actual api key. Only visible after creation.
	Key string `xorm:"not null varchar(50)" json:"key,omitempty"`
	// The permissions this key has. Possible values are available via the /routes endpoint and consist of the keys of the list from that endpoint. For example, if the token should be able to read all tasks as well as update existing tasks, you should add `{"tasks":["read_all","update"]}`.
	Permissions APIPermissions `xorm:"json not null" json:"permissions"`
	// The date when this key expires.
	ExpiresAt time.Time `xorm:"not null" json:"expires_at"`

	// A timestamp when this api key was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this api key was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	OwnerID int64 `xorm:"bigint not null" json:"-"`

	web.Rights   `xorm:"-" json:"-"`
	web.CRUDable `xorm:"-" json:"-"`
}

func (*APIToken) TableName() string {
	return "api_tokens"
}
