// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	_ "github.com/go-sql-driver/mysql" // Because.
	"github.com/go-xorm/xorm"

	_ "github.com/mattn/go-sqlite3" // Because.
)

var (
	x *xorm.Engine
)

// GetTables returns all structs which are also a table.
func GetTables() []interface{} {
	return []interface{}{
		&user.User{},
		&List{},
		&Task{},
		&Team{},
		&TeamMember{},
		&TeamList{},
		&TeamNamespace{},
		&Namespace{},
		&ListUser{},
		&NamespaceUser{},
		&TaskAssginee{},
		&Label{},
		&LabelTask{},
		&TaskReminder{},
		&LinkSharing{},
		&TaskRelation{},
		&TaskAttachment{},
	}
}

// SetEngine sets the xorm.Engine
func SetEngine() (err error) {
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

func getLimitFromPageIndex(page int, perPage int) (limit, start int) {

	// Get everything when page index is -1
	if page < 0 {
		return 0, 0
	}

	limit = config.ServiceMaxItemsPerPage.GetInt()
	if perPage > 0 {
		limit = perPage
	}

	start = limit * (page - 1)
	return
}

// GetTotalCount returns the total amount of something
func GetTotalCount(counting interface{}) (count int64, err error) {
	return x.Count(counting)
}
