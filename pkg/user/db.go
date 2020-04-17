// Copyright 2018-2020 Vikunja and contriubtors. All rights reserved.
//
// This file is part of Vikunja.
//
// Vikunja is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Vikunja is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Vikunja.  If not, see <https://www.gnu.org/licenses/>.

package user

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"xorm.io/xorm"
)

var x *xorm.Engine

// InitDB sets up the database connection to use in this module
func InitDB() (err error) {
	x, err = db.CreateDBEngine()
	if err != nil {
		log.Criticalf("Could not connect to db: %v", err.Error())
		return
	}

	// Cache
	if config.CacheEnabled.GetBool() && config.CacheType.GetString() == "redis" {
		db.RegisterTableStructsForCache(GetTables())
	}

	return nil
}

// GetTables returns all structs which are also a table.
func GetTables() []interface{} {
	return []interface{}{
		&User{},
		&TOTP{},
	}
}
