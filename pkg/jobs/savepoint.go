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

package jobs

import "xorm.io/xorm"

// Isolate returns an item error separately from a broken transaction.
func Isolate(s *xorm.Session, fn func() error) (itemError, transactionError error) {
	if _, err := s.Exec("SAVEPOINT job_item"); err != nil {
		return nil, err
	}
	itemError = fn()
	if itemError != nil {
		if _, err := s.Exec("ROLLBACK TO SAVEPOINT job_item"); err != nil {
			return itemError, err
		}
	}
	_, transactionError = s.Exec("RELEASE SAVEPOINT job_item")
	return
}
