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

package shared

import (
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/log"
)

// dependentTestingTables lists tables that reference a reset table by ID and
// must be truncated alongside it. Without foreign key cascades, stale rows
// would persist and pollute subsequent tests that reuse the same
// auto-increment IDs.
var dependentTestingTables = map[string][]string{
	"users": {"notifications"},
}

// ReplaceTableContents resets a single table to the provided rows for the e2e
// testing endpoint and returns the table's resulting contents. When truncate is
// true the table (and any dependent tables) is emptied first; otherwise the rows
// are restored on top of existing data. Callers must already have verified the
// testing token.
func ReplaceTableContents(table string, content []map[string]interface{}, truncate bool) ([]map[string]interface{}, error) {
	// Wait for all async event handlers from the previous test to complete
	// before modifying the database. Without this, handlers hold SQLite
	// connections and starve this request's truncate/insert operations.
	events.WaitForPendingHandlers()

	var err error
	if truncate {
		for _, dep := range dependentTestingTables[table] {
			if err = db.RestoreAndTruncate(dep, nil); err != nil {
				return nil, err
			}
		}
		err = db.RestoreAndTruncate(table, content)
	} else {
		err = db.Restore(table, content)
	}
	if err != nil {
		return nil, err
	}

	// License state is cached at startup; re-apply so tests take effect without a restart.
	if table == "license_status" {
		if err := license.ReloadFromCache(); err != nil {
			return nil, err
		}
	}

	s := db.NewSession()
	defer s.Close()
	data := []map[string]interface{}{}
	if err := s.Table(table).Find(&data); err != nil {
		return nil, err
	}
	return data, nil
}

// TruncateAllTestingTables empties every Vikunja table for the e2e testing
// endpoint. Callers must already have verified the testing token.
func TruncateAllTestingTables() error {
	events.WaitForPendingHandlers()

	if err := db.TruncateAllTables(); err != nil {
		return err
	}

	// Reload after truncate; otherwise features enabled by a prior test outlive
	// the now-empty license_status table. A reload failure here is non-fatal —
	// the truncate already succeeded — so it is logged and swallowed.
	if err := license.ReloadFromCache(); err != nil {
		log.Errorf("Error reloading license after truncate: %v", err)
	}
	return nil
}
