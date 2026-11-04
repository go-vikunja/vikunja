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

	"github.com/stretchr/testify/require"
)

func TestProjectWebhooksForEvent_RequiresConfiguredProjectAccess(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	owner := &user.User{ID: 1}
	creator := &user.User{ID: 2, Username: "user2"}
	parent := &Project{Title: "webhook parent"}
	require.NoError(t, parent.Create(s, owner))
	child := &Project{Title: "webhook child", ParentProjectID: &parent.ID}
	require.NoError(t, child.Create(s, owner))
	require.NoError(t, (&ProjectUser{ProjectID: child.ID, Username: creator.Username, Permission: PermissionRead}).Create(s, owner))
	hook := &Webhook{
		TargetURL:   "https://example.com/parent",
		Events:      []string{"task.updated"},
		ProjectID:   parent.ID,
		CreatedByID: creator.ID,
	}
	_, err := s.Insert(hook)
	require.NoError(t, err)

	matching, err := projectWebhooksForEvent(s, child.ID, "task.updated")
	require.NoError(t, err)
	require.Empty(t, matching)

	require.NoError(t, (&ProjectUser{ProjectID: parent.ID, Username: creator.Username, Permission: PermissionRead}).Create(s, owner))
	matching, err = projectWebhooksForEvent(s, child.ID, "task.updated")
	require.NoError(t, err)
	require.Len(t, matching, 1)
	require.Equal(t, hook.ID, matching[0].ID)
}
