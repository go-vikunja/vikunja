// Copyright2018-2020 Vikunja and contriubtors. All rights reserved.
//
// This file is part of Vikunja.
//
// Vikunja is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Vikunja is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with Vikunja.  If not, see <https://www.gnu.org/licenses/>.

package user

import (
	"strconv"
	"strings"

	"xorm.io/xorm"

	"code.vikunja.io/api/pkg/log"
)

// ListUsers returns a list with all users, filtered by an optional searchstring
func ListUsers(s *xorm.Session, searchterm string) (users []*User, err error) {

	vals := strings.Split(searchterm, ",")
	ids := []int64{}
	for _, val := range vals {
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			log.Debugf("User search string part '%s' is not a number: %s", val, err)
			continue
		}
		ids = append(ids, v)
	}

	if len(ids) > 0 {
		err = s.
			In("id", ids).
			Find(&users)
		return
	}

	if searchterm == "" {
		err = s.Find(&users)
		return
	}

	err = s.
		Where("username LIKE ?", "%"+searchterm+"%").
		Find(&users)
	return
}
