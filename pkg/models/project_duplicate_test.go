// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
)

func TestProjectDuplicate(t *testing.T) {

	db.LoadAndAssertFixtures(t)
	files.InitTestFileFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{
		ID: 1,
	}

	l := &ProjectDuplicate{
		ProjectID: 1,
	}
	can, err := l.CanCreate(s, u)
	assert.NoError(t, err)
	assert.True(t, can)
	err = l.Create(s, u)
	assert.NoError(t, err)

	// assert the new project has the same number of buckets as the old one
	numberOfOriginalBuckets, err := s.Where("project_id = ?", l.ProjectID).Count(&Bucket{})
	assert.NoError(t, err)
	numberOfDuplicatedBuckets, err := s.Where("project_id = ?", l.Project.ID).Count(&Bucket{})
	assert.NoError(t, err)
	assert.Equal(t, numberOfOriginalBuckets, numberOfDuplicatedBuckets, "duplicated project does not have the same amount of buckets as the original one")

	// To make this test 100% useful, it would need to assert a lot more stuff, but it is good enough for now.
	// Also, we're lacking utility functions to do all needed assertions.
}
