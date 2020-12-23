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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
)

func TestListDuplicate(t *testing.T) {

	db.LoadAndAssertFixtures(t)
	files.InitTestFileFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{
		ID: 1,
	}

	l := &ListDuplicate{
		ListID:      1,
		NamespaceID: 1,
	}
	can, err := l.CanCreate(s, u)
	assert.NoError(t, err)
	assert.True(t, can)
	err = l.Create(s, u)
	assert.NoError(t, err)
	// To make this test 100% useful, it would need to assert a lot more stuff, but it is good enough for now.
	// Also, we're lacking utility functions to do all needed assertions.
}
