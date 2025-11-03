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

package files

import (
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"xorm.io/xorm"
)

var x *xorm.Engine

// SetEngine sets the xorm.Engine
func SetEngine() (err error) {
	x, err = db.CreateDBEngine()
	if err != nil {
		log.Criticalf("Could not connect to db: %v", err.Error())
		return
	}

	return nil
}

// GetTables returns all structs which are also a table.
func GetTables() []interface{} {
	return []interface{}{
		&File{},
	}
}
