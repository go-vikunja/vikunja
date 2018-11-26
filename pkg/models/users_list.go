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

// ListUsers returns a list with all users, filtered by an optional searchstring
func ListUsers(searchterm string) (users []User, err error) {

	if searchterm == "" {
		err = x.Find(&users)
	} else {
		err = x.
			Where("username LIKE ?", "%"+searchterm+"%").
			Find(&users)
	}

	if err != nil {
		return []User{}, err
	}

	return users, nil
}
