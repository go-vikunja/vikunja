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

package models

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
)

func TestAdminBypass_Project(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// A user who owns no projects and shares none.
	stranger := &user.User{ID: 9999, IsAdmin: true}
	p := &Project{ID: 1}

	can, _, err := p.CanRead(s, stranger)
	assert.NoError(t, err)
	assert.True(t, can, "admin must be able to read any project")

	can, err = p.CanWrite(s, stranger)
	assert.NoError(t, err)
	assert.True(t, can)

	can, err = p.CanUpdate(s, stranger)
	assert.NoError(t, err)
	assert.True(t, can)

	can, err = p.CanDelete(s, stranger)
	assert.NoError(t, err)
	assert.True(t, can)
}
