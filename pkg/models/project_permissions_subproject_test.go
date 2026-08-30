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
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// The parent's owner must stay admin on subprojects a write member created below it,
// and every way we report that permission must agree. See #3574.
func TestProjectPermissions_OwnerOfParentOnMemberCreatedSubproject(t *testing.T) {
	owner := &user.User{ID: 1, Username: "user1"}
	member := &user.User{ID: 2, Username: "user2"}

	assertOwnerIsAdminOnSubproject := func(t *testing.T, shareParent func(s *xorm.Session, parentID int64)) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		parent := &Project{Title: "parent"}
		require.NoError(t, parent.Create(s, owner))
		shareParent(s, parent.ID)

		sub := &Project{Title: "sub", ParentProjectID: &parent.ID}
		require.NoError(t, sub.Create(s, member))
		subsub := &Project{Title: "subsub", ParentProjectID: &sub.ID}
		require.NoError(t, subsub.Create(s, member))
		require.NoError(t, s.Commit())

		s = db.NewSession()
		defer s.Close()

		for _, project := range []*Project{sub, subsub} {
			canRead, maxPermission, err := (&Project{ID: project.ID}).CanRead(s, owner)
			require.NoError(t, err)
			assert.True(t, canRead)
			assert.Equal(t, int(PermissionAdmin), maxPermission, "%s: CanRead", project.Title)

			isAdmin, err := (&Project{ID: project.ID}).IsAdmin(s, owner)
			require.NoError(t, err)
			assert.True(t, isAdmin, "%s: IsAdmin", project.Title)
		}

		all, _, _, err := (&Project{Expand: ProjectExpandableRights}).ReadAll(s, owner, "", 1, 500)
		require.NoError(t, err)
		seen := 0
		for _, project := range all.([]*Project) {
			if project.ID == sub.ID || project.ID == subsub.ID {
				seen++
				require.NotNil(t, project.MaxPermission, "%s: ReadAll", project.Title)
				assert.Equal(t, PermissionAdmin, *project.MaxPermission, "%s: ReadAll", project.Title)
			}
		}
		assert.Equal(t, 2, seen, "both subprojects should show up in the owner's project list")

		unrelated := &user.User{ID: 3, Username: "user3"}
		for _, project := range []*Project{sub, subsub} {
			canRead, _, err := (&Project{ID: project.ID}).CanRead(s, unrelated)
			require.NoError(t, err)
			assert.False(t, canRead, "%s: CanRead for unrelated user", project.Title)

			isAdmin, err := (&Project{ID: project.ID}).IsAdmin(s, unrelated)
			require.NoError(t, err)
			assert.False(t, isAdmin, "%s: IsAdmin for unrelated user", project.Title)
		}

		allUnrelated, _, _, err := (&Project{Expand: ProjectExpandableRights}).ReadAll(s, unrelated, "", 1, 500)
		require.NoError(t, err)
		for _, project := range allUnrelated.([]*Project) {
			assert.NotEqual(t, sub.ID, project.ID, "sub should not show up in an unrelated user's project list")
			assert.NotEqual(t, subsub.ID, project.ID, "subsub should not show up in an unrelated user's project list")
		}

		memberIsAdmin, err := (&Project{ID: parent.ID}).IsAdmin(s, member)
		require.NoError(t, err)
		assert.False(t, memberIsAdmin, "%s: IsAdmin for write member", parent.Title)
	}

	t.Run("parent shared with the member directly", func(t *testing.T) {
		assertOwnerIsAdminOnSubproject(t, func(s *xorm.Session, parentID int64) {
			share := &ProjectUser{ProjectID: parentID, Username: member.Username, Permission: PermissionWrite}
			require.NoError(t, share.Create(s, owner))
		})
	})

	t.Run("parent shared with a team the member is in", func(t *testing.T) {
		assertOwnerIsAdminOnSubproject(t, func(s *xorm.Session, parentID int64) {
			team := &Team{Name: "sharing team"}
			require.NoError(t, team.Create(s, owner))
			require.NoError(t, (&TeamMember{TeamID: team.ID, Username: member.Username}).Create(s, owner))

			share := &TeamProject{TeamID: team.ID, ProjectID: parentID, Permission: PermissionWrite}
			require.NoError(t, share.Create(s, owner))
		})
	})
}
