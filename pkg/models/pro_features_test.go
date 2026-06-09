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
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsProFeatureEnabledForUser(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("unlicensed instance disables the feature regardless of overrides", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		license.ResetForTests()

		granted := &user.User{ID: 1, ProFeatureOverrides: map[string]bool{"time_tracking": true}}
		enabled, err := IsProFeatureEnabledForUser(s, granted, license.FeatureTimeTracking)
		require.NoError(t, err)
		assert.False(t, enabled)
	})

	t.Run("non-toggleable feature is on for everyone when licensed", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		defer license.ResetForTests()

		enabled, err := IsProFeatureEnabledForUser(s, u, license.FeatureAdminPanel)
		require.NoError(t, err)
		assert.True(t, enabled)
	})

	t.Run("code default applies without instance default or override", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		license.SetForTests([]license.Feature{license.FeatureTimeTracking})
		defer license.ResetForTests()

		enabled, err := IsProFeatureEnabledForUser(s, u, license.FeatureTimeTracking)
		require.NoError(t, err)
		assert.True(t, enabled, "time_tracking's code default is enabled")
	})

	t.Run("instance default overrides the code default", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		license.SetForTests([]license.Feature{license.FeatureTimeTracking})
		defer license.ResetForTests()

		require.NoError(t, SetProFeatureInstanceDefault(s, license.FeatureTimeTracking, false))

		enabled, err := IsProFeatureEnabledForUser(s, u, license.FeatureTimeTracking)
		require.NoError(t, err)
		assert.False(t, enabled)
	})

	t.Run("user override wins over the instance default", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		license.SetForTests([]license.Feature{license.FeatureTimeTracking})
		defer license.ResetForTests()

		require.NoError(t, SetProFeatureInstanceDefault(s, license.FeatureTimeTracking, false))
		granted := &user.User{ID: 1, ProFeatureOverrides: map[string]bool{"time_tracking": true}}

		enabled, err := IsProFeatureEnabledForUser(s, granted, license.FeatureTimeTracking)
		require.NoError(t, err)
		assert.True(t, enabled)
	})

	t.Run("user override can revoke a default-enabled feature", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		license.SetForTests([]license.Feature{license.FeatureTimeTracking})
		defer license.ResetForTests()

		revoked := &user.User{ID: 1, ProFeatureOverrides: map[string]bool{"time_tracking": false}}

		enabled, err := IsProFeatureEnabledForUser(s, revoked, license.FeatureTimeTracking)
		require.NoError(t, err)
		assert.False(t, enabled)
	})
}

func TestIsProFeatureEnabledForAuth(t *testing.T) {
	t.Run("reads the user's override from the db", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		license.SetForTests([]license.Feature{license.FeatureTimeTracking})
		defer license.ResetForTests()

		u, err := user.GetUserByID(s, 1)
		require.NoError(t, err)
		require.NoError(t, SetUserProFeatureOverride(s, u, license.FeatureTimeTracking, boolPtr(false)))

		// A claim-derived auth user without the override field must still be revoked.
		enabled, err := IsProFeatureEnabledForAuth(s, &user.User{ID: 1}, license.FeatureTimeTracking)
		require.NoError(t, err)
		assert.False(t, enabled)
	})

	t.Run("link shares follow the instance default", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()
		license.SetForTests([]license.Feature{license.FeatureTimeTracking})
		defer license.ResetForTests()

		enabled, err := IsProFeatureEnabledForAuth(s, &LinkSharing{ID: 1}, license.FeatureTimeTracking)
		require.NoError(t, err)
		assert.True(t, enabled, "code default applies to link shares")

		require.NoError(t, SetProFeatureInstanceDefault(s, license.FeatureTimeTracking, false))
		enabled, err = IsProFeatureEnabledForAuth(s, &LinkSharing{ID: 1}, license.FeatureTimeTracking)
		require.NoError(t, err)
		assert.False(t, enabled)
	})
}

func TestSetUserProFeatureOverride(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	license.SetForTests([]license.Feature{license.FeatureTimeTracking})
	defer license.ResetForTests()

	u, err := user.GetUserByID(s, 1)
	require.NoError(t, err)

	require.NoError(t, SetUserProFeatureOverride(s, u, license.FeatureTimeTracking, boolPtr(false)))
	fresh, err := user.GetUserByID(s, 1)
	require.NoError(t, err)
	assert.Equal(t, map[string]bool{"time_tracking": false}, fresh.ProFeatureOverrides)

	require.NoError(t, SetUserProFeatureOverride(s, fresh, license.FeatureTimeTracking, nil))
	fresh, err = user.GetUserByID(s, 1)
	require.NoError(t, err)
	assert.Empty(t, fresh.ProFeatureOverrides)
}

func TestSetProFeatureInstanceDefault(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	require.NoError(t, SetProFeatureInstanceDefault(s, license.FeatureTimeTracking, false))
	defaults, err := GetProFeatureInstanceDefaults(s)
	require.NoError(t, err)
	assert.Equal(t, map[string]bool{"time_tracking": false}, defaults)

	// Upsert: setting again must update, not duplicate.
	require.NoError(t, SetProFeatureInstanceDefault(s, license.FeatureTimeTracking, true))
	defaults, err = GetProFeatureInstanceDefaults(s)
	require.NoError(t, err)
	assert.Equal(t, map[string]bool{"time_tracking": true}, defaults)

	require.NoError(t, ResetProFeatureInstanceDefault(s, license.FeatureTimeTracking))
	defaults, err = GetProFeatureInstanceDefaults(s)
	require.NoError(t, err)
	assert.Empty(t, defaults)
}

func TestEffectiveProFeaturesForUser(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	license.SetForTests([]license.Feature{license.FeatureAdminPanel, license.FeatureTimeTracking})
	defer license.ResetForTests()

	revoked := &user.User{ID: 1, ProFeatureOverrides: map[string]bool{"time_tracking": false}}
	features, err := EffectiveProFeaturesForUser(s, revoked)
	require.NoError(t, err)
	assert.Equal(t, []license.Feature{license.FeatureAdminPanel}, features,
		"a revoked toggleable feature must drop out while instance-wide features stay")

	granted := &user.User{ID: 1}
	features, err = EffectiveProFeaturesForUser(s, granted)
	require.NoError(t, err)
	assert.Equal(t, []license.Feature{license.FeatureAdminPanel, license.FeatureTimeTracking}, features)
}

func boolPtr(b bool) *bool { return &b }
