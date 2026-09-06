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

package entitlement

import (
	"os"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	log.InitLogger()
	config.InitDefaultConfig()
	keyvalue.InitStorage() // license.SetForTests persists state through keyvalue

	x, err := db.CreateTestEngine()
	if err != nil {
		log.Fatal(err)
	}
	if err := x.Sync2(append(user.GetTables(), GetTables()...)...); err != nil {
		log.Fatal(err)
	}
	if err := db.InitTestFixtures("users", "user_entitlements"); err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())
}

var (
	restricted   = &user.User{ID: 16}
	unrestricted = &user.User{ID: 1}
	botOf16      = &user.User{ID: 99, BotOwnerID: 16}
)

type shareAuth struct{}

func (shareAuth) GetID() int64 { return -1 }

func TestCheck(t *testing.T) {
	t.Run("license off yields not licensed regardless of rows", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		license.ResetForTests()

		err := Check(s, unrestricted, FeatureTimeTracking)
		assert.True(t, IsErrFeatureNotLicensed(err))
		has, err := Has(s, unrestricted, FeatureTimeTracking)
		require.NoError(t, err)
		assert.False(t, has)
	})
	t.Run("license on and no row", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		license.SetForTests([]license.Feature{license.FeatureTimeTracking})
		defer license.ResetForTests()

		require.NoError(t, Check(s, unrestricted, FeatureTimeTracking))
	})
	t.Run("license on and row off", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		license.SetForTests([]license.Feature{license.FeatureTimeTracking})
		defer license.ResetForTests()

		err := Check(s, restricted, FeatureTimeTracking)
		assert.True(t, IsErrFeatureDisabledForUser(err))
	})
	t.Run("bot resolves to owner", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := Check(s, botOf16, FeatureTeamCreation)
		assert.True(t, IsErrFeatureDisabledForUser(err))
	})
	t.Run("feature unknown to license needs no license", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		license.ResetForTests()

		require.NoError(t, Check(s, unrestricted, FeatureTeamCreation))
		assert.True(t, IsErrFeatureDisabledForUser(Check(s, restricted, FeatureTeamCreation)))
	})
	t.Run("link share gets the license-only answer", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		license.SetForTests([]license.Feature{license.FeatureTimeTracking})
		defer license.ResetForTests()

		require.NoError(t, Check(s, shareAuth{}, FeatureTimeTracking))
	})
	t.Run("instance-only features ignore rows", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		_, err := s.Insert(&UserEntitlement{UserID: 16, Feature: FeatureAdminPanel, Value: 0})
		require.NoError(t, err)
		require.NoError(t, Check(s, restricted, FeatureAdminPanel))
	})
}

func TestLimit(t *testing.T) {
	t.Run("no row is unlimited", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, limited, err := Limit(s, 1, FeatureMaxProjects)
		require.NoError(t, err)
		assert.False(t, limited)
	})
	t.Run("row limits", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		limit, limited, err := Limit(s, 16, FeatureMaxProjects)
		require.NoError(t, err)
		assert.True(t, limited)
		assert.Equal(t, int64(1), limit)
	})
}

func TestForUser(t *testing.T) {
	t.Run("without rows", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		license.SetForTests([]license.Feature{license.FeatureTimeTracking})
		defer license.ResetForTests()

		got, err := ForUser(s, 1)
		require.NoError(t, err)
		assert.Equal(t, map[Feature]int64{
			FeatureAdminPanel:   0,
			FeatureAuditLogs:    0,
			FeatureTimeTracking: 1,
			FeatureTeamCreation: 1,
		}, got)
	})
	t.Run("with rows", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		license.SetForTests([]license.Feature{license.FeatureTimeTracking, license.FeatureAdminPanel})
		defer license.ResetForTests()

		got, err := ForUser(s, 16)
		require.NoError(t, err)
		assert.Equal(t, map[Feature]int64{
			FeatureAdminPanel:      1,
			FeatureAuditLogs:       0,
			FeatureTimeTracking:    0,
			FeatureTeamCreation:    0,
			FeatureMaxProjects:     1,
			FeatureMaxStorageBytes: 1024,
		}, got)
	})
}

func TestReplace(t *testing.T) {
	t.Run("replaces the full set", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := Replace(s, 16, map[Feature]int64{FeatureTimeTracking: 1, FeatureMaxProjects: 5})
		require.NoError(t, err)

		rows, err := Rows(s, 16)
		require.NoError(t, err)
		assert.Equal(t, map[Feature]int64{FeatureTimeTracking: 1, FeatureMaxProjects: 5}, rows)
	})
	t.Run("empty set deletes everything", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		require.NoError(t, Replace(s, 16, map[Feature]int64{}))
		rows, err := Rows(s, 16)
		require.NoError(t, err)
		assert.Empty(t, rows)
	})
	t.Run("rejects unknown and instance-only features", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		assert.True(t, IsErrUnknownFeature(Replace(s, 16, map[Feature]int64{"nope": 1})))
		assert.True(t, IsErrUnknownFeature(Replace(s, 16, map[Feature]int64{FeatureAdminPanel: 1})))
	})
}
