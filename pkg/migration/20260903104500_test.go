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

package migration

import (
	"testing"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/require"
)

type projectsBefore20260903104500 struct {
	ID              int64  `xorm:"bigint autoincr not null unique pk"`
	Title           string `xorm:"varchar(250) not null"`
	ParentProjectID *int64 `xorm:"bigint INDEX null"`
}

func (projectsBefore20260903104500) TableName() string {
	return "projects"
}

func TestRootProjectsParentToNull20260903104500(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	table := projectsBefore20260903104500{}
	t.Cleanup(func() {
		require.NoError(t, x.DropTables(table))
	})
	require.NoError(t, x.DropTables(table))
	require.NoError(t, x.Sync2(table)) //nolint:forbidigo // test-local table

	root := int64(0)
	child := int64(1)
	_, err = x.Insert([]*projectsBefore20260903104500{
		{ID: 1, Title: "root stored as 0", ParentProjectID: &root},
		{ID: 2, Title: "root stored as null"},
		{ID: 3, Title: "child", ParentProjectID: &child},
	})
	require.NoError(t, err)

	require.NoError(t, rootProjectsParentToNull20260903104500(x))

	projects := []*projectsBefore20260903104500{}
	require.NoError(t, x.OrderBy("id").Find(&projects))
	require.Len(t, projects, 3)
	require.Nil(t, projects[0].ParentProjectID)
	require.Nil(t, projects[1].ParentProjectID)
	require.NotNil(t, projects[2].ParentProjectID)
	require.Equal(t, child, *projects[2].ParentProjectID)
}
