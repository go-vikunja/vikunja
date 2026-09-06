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
	"bytes"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/entitlement"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// user16 carries the restricted fixture rows: flags off, max_projects 1 (owns
// project 37), max_storage_bytes 1024 (uses 0).
var (
	restrictedUser   = &user.User{ID: 16}
	unrestrictedUser = &user.User{ID: 1}
)

func TestEntitlement_TeamCreation(t *testing.T) {
	t.Run("restricted user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		can, err := (&Team{}).CanCreate(s, restrictedUser)
		assert.False(t, can)
		assert.True(t, entitlement.IsErrFeatureDisabledForUser(err))
	})
	t.Run("unrestricted user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		can, err := (&Team{}).CanCreate(s, unrestrictedUser)
		require.NoError(t, err)
		assert.True(t, can)
	})
	t.Run("bot of restricted user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		can, err := (&Team{}).CanCreate(s, &user.User{ID: 23, BotOwnerID: 16})
		assert.False(t, can)
		assert.True(t, entitlement.IsErrFeatureDisabledForUser(err))
	})
}

func TestEntitlement_TimeTracking(t *testing.T) {
	license.SetForTests([]license.Feature{license.FeatureTimeTracking})
	defer license.ResetForTests()

	t.Run("restricted user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		can, err := (&TimeEntry{ProjectID: 37}).CanCreate(s, restrictedUser)
		assert.False(t, can)
		assert.True(t, entitlement.IsErrFeatureDisabledForUser(err))
	})
	t.Run("unrestricted user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		can, err := (&TimeEntry{ProjectID: 1}).CanCreate(s, unrestrictedUser)
		require.NoError(t, err)
		assert.True(t, can)
	})
	t.Run("unlicensed instance", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		license.ResetForTests()
		defer license.SetForTests([]license.Feature{license.FeatureTimeTracking})

		_, err := (&TimeEntry{ProjectID: 1}).CanCreate(s, unrestrictedUser)
		assert.True(t, entitlement.IsErrFeatureNotLicensed(err))
	})
}

func TestEntitlement_ProjectLimit(t *testing.T) {
	t.Run("at limit", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := (&Project{Title: "one too many"}).Create(s, restrictedUser)
		require.True(t, entitlement.IsErrLimitReached(err))
		assert.Equal(t, entitlement.ErrLimitReached{Feature: entitlement.FeatureMaxProjects, Limit: 1, Current: 1}, err)
	})
	t.Run("child projects count", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		parent := int64(37)
		err := (&Project{Title: "child", ParentProjectID: &parent}).Create(s, restrictedUser)
		assert.True(t, entitlement.IsErrLimitReached(err))
	})
	t.Run("duplicate counts", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		pd := &ProjectDuplicate{ProjectID: 37}
		can, err := pd.CanCreate(s, restrictedUser)
		require.NoError(t, err)
		require.True(t, can)
		assert.True(t, entitlement.IsErrLimitReached(pd.Create(s, restrictedUser)))
	})
	t.Run("under limit", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		require.NoError(t, entitlement.Replace(s, 16, map[entitlement.Feature]int64{entitlement.FeatureMaxProjects: 2}))
		p := &Project{Title: "second"}
		require.NoError(t, p.Create(s, restrictedUser))
		assert.True(t, entitlement.IsErrLimitReached((&Project{Title: "third"}).Create(s, restrictedUser)))
	})
	t.Run("no row is unlimited", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		require.NoError(t, (&Project{Title: "free"}).Create(s, unrestrictedUser))
	})
	t.Run("bot is charged to its owner", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// user 21 owns project 44 and bot 23
		require.NoError(t, entitlement.Replace(s, 21, map[entitlement.Feature]int64{entitlement.FeatureMaxProjects: 1}))
		bot, err := user.GetUserByID(s, 23)
		require.NoError(t, err)
		err = (&Project{Title: "by bot"}).Create(s, bot)
		assert.True(t, entitlement.IsErrLimitReached(err))
	})
}

func newTaskIn(t *testing.T, s *xorm.Session, projectID int64, a *user.User) *Task {
	t.Helper()
	task := &Task{Title: "storage", ProjectID: projectID}
	require.NoError(t, task.Create(s, a))
	return task
}

func upload(s *xorm.Session, taskID int64, size int, a *user.User) error {
	ta := &TaskAttachment{TaskID: taskID}
	content := bytes.Repeat([]byte("x"), size)
	return ta.NewAttachment(s, bytes.NewReader(content), "blob", uint64(len(content)), a)
}

func TestEntitlement_StorageLimit(t *testing.T) {
	t.Run("owner under and over the limit", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task := newTaskIn(t, s, 37, restrictedUser)
		require.NoError(t, upload(s, task.ID, 1000, restrictedUser))

		err := upload(s, task.ID, 25, restrictedUser)
		require.True(t, entitlement.IsErrLimitReached(err))
		assert.Equal(t, entitlement.ErrLimitReached{Feature: entitlement.FeatureMaxStorageBytes, Limit: 1024, Current: 1000}, err)

		require.NoError(t, upload(s, task.ID, 24, restrictedUser))
	})
	t.Run("unrestricted member of a restricted owner's project is blocked", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := s.Insert(&ProjectUser{ProjectID: 37, UserID: 1, Permission: PermissionWrite})
		require.NoError(t, err)
		task := newTaskIn(t, s, 37, unrestrictedUser)
		assert.True(t, entitlement.IsErrLimitReached(upload(s, task.ID, 2000, unrestrictedUser)))
	})
	t.Run("restricted member of an unrestricted owner's project may upload", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := s.Insert(&ProjectUser{ProjectID: 1, UserID: 16, Permission: PermissionWrite})
		require.NoError(t, err)
		require.NoError(t, upload(s, 1, 5000, restrictedUser))
	})
	t.Run("background upload counts", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		require.NoError(t, CheckStorageLimit(s, 16, 1024))
		assert.True(t, entitlement.IsErrLimitReached(CheckStorageLimit(s, 16, 1025)))
	})
}

func TestEntitlementUsage(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// user1: 5 projects; task 1 (project 1) has two attachments of file 1 (100 bytes each).
	got, err := EntitlementUsage(s, 1)
	require.NoError(t, err)
	assert.Equal(t, map[entitlement.Feature]int64{
		entitlement.FeatureMaxProjects:     5,
		entitlement.FeatureMaxStorageBytes: 200,
	}, got)

	// user6 owns project 35 whose background is file 1.
	got, err = EntitlementUsage(s, 6)
	require.NoError(t, err)
	assert.Equal(t, int64(100), got[entitlement.FeatureMaxStorageBytes])
}
