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
	"context"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedConfig makes the config what a fresh install has, and restores it afterwards.
func seedConfig(t *testing.T) {
	t.Helper()
	config.ServiceCreateDefaultAdmin.Set(true)
	config.AuthLocalEnabled.Set(true)
	config.ServiceDefaultAdminPassword.Set("admin123")
	config.ServiceDefaultAdminEmail.Set("admin@vika.local")
	t.Cleanup(func() {
		config.ServiceCreateDefaultAdmin.Set(true)
		config.AuthLocalEnabled.Set(true)
		config.ServiceDefaultAdminPassword.Set("admin123")
		config.ServiceDefaultAdminEmail.Set("admin@vika.local")
	})
}

func ensure(t *testing.T) (bool, error) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	created, err := EnsureDefaultAdmin(s)
	if err != nil {
		_ = s.Rollback()
		return false, err
	}
	require.NoError(t, s.Commit())
	return created, nil
}

// freshInstall is a database without a single user, like the first start.
func freshInstall(t *testing.T) {
	t.Helper()
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	_, err := s.Where("1 = 1").Delete(&user.User{})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}
func countAdmins(t *testing.T) int64 {
	t.Helper()
	s := db.NewReadSession()
	defer s.Close()
	n, err := s.Where("is_admin = ?", true).Count(&user.User{})
	require.NoError(t, err)
	return n
}

func TestEnsureDefaultAdmin(t *testing.T) {
	t.Run("creates the admin on a brand-new installation", func(t *testing.T) {
		freshInstall(t)
		seedConfig(t)

		created, err := ensure(t)
		require.NoError(t, err)
		require.True(t, created)

		s := db.NewSession()
		defer s.Close()
		admin := &user.User{}
		has, err := s.Where("username = ?", DefaultAdminUsername).Get(admin)
		require.NoError(t, err)
		require.True(t, has)

		assert.True(t, admin.IsAdmin)
		assert.True(t, admin.MustChangePassword)
		assert.Equal(t, user.StatusActive, admin.Status)
		assert.Equal(t, "admin@vika.local", admin.Email)
		assert.Equal(t, user.IssuerLocal, admin.Issuer)
		assert.NotEqual(t, "admin123", admin.Password, "stored as a hash")
		require.NoError(t, user.CheckUserPassword(admin, "admin123"))
		db.AssertExists(t, "projects", map[string]interface{}{"owner_id": admin.ID, "title": "Inbox"}, false)
	})
	t.Run("a second call changes nothing", func(t *testing.T) {
		freshInstall(t)
		seedConfig(t)

		first, err := ensure(t)
		require.NoError(t, err)
		require.True(t, first)
		second, err := ensure(t)
		require.NoError(t, err)

		assert.False(t, second)
		assert.Equal(t, int64(1), countAdmins(t))
	})
	t.Run("an existing instance is never changed, with or without an admin", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		seedConfig(t)

		// The fixtures hold users and no admin at all: exactly the instance that must not get a
		// published password by being upgraded.
		created, err := ensure(t)
		require.NoError(t, err)
		assert.False(t, created)
		assert.Zero(t, countAdmins(t))
		db.AssertMissing(t, "users", map[string]interface{}{"username": DefaultAdminUsername})

		// One admin of any login method changes nothing either.
		s := db.NewSession()
		_, err = s.ID(14).Cols("is_admin").Update(&user.User{IsAdmin: true}) // an OpenID user
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()
		created, err = ensure(t)
		require.NoError(t, err)
		assert.False(t, created)
		db.AssertMissing(t, "users", map[string]interface{}{"username": DefaultAdminUsername})
	})
	t.Run("a single existing user, bot or disabled account is enough to count as existing", func(t *testing.T) {
		freshInstall(t)
		seedConfig(t)
		s := db.NewSession()
		_, err := s.Insert(&user.User{Username: "someone", Email: "someone@example.com", Issuer: user.IssuerLocal, Status: user.StatusDisabled})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		created, err := ensure(t)

		require.NoError(t, err)
		assert.False(t, created)
		assert.Zero(t, countAdmins(t))
	})	t.Run("it is switched off by the config key and by disabled local login", func(t *testing.T) {
		freshInstall(t)
		seedConfig(t)

		config.ServiceCreateDefaultAdmin.Set(false)
		created, err := ensure(t)
		require.NoError(t, err)
		assert.False(t, created)

		config.ServiceCreateDefaultAdmin.Set(true)
		config.AuthLocalEnabled.Set(false)
		created, err = ensure(t)
		require.NoError(t, err)
		assert.False(t, created, "a known password is pointless when local login is off")
		assert.Zero(t, countAdmins(t))
	})
	t.Run("an unusable password or an empty email is refused, not stored", func(t *testing.T) {
		freshInstall(t)
		seedConfig(t)

		config.ServiceDefaultAdminPassword.Set("short")
		created, err := ensure(t)
		require.NoError(t, err)
		assert.False(t, created)

		config.ServiceDefaultAdminPassword.Set("admin123")
		config.ServiceDefaultAdminEmail.Set("  ")
		created, err = ensure(t)
		require.NoError(t, err)
		assert.False(t, created)
		assert.Zero(t, countAdmins(t))
	})
	t.Run("the account can sign in only to change its password", func(t *testing.T) {
		freshInstall(t)
		seedConfig(t)
		created, err := ensure(t)
		require.NoError(t, err)
		require.True(t, created)

		s := db.NewSession()
		defer s.Close()
		admin, err := user.CheckUserCredentials(context.Background(), s, &user.Login{Username: DefaultAdminUsername, Password: "admin123"})
		require.NoError(t, err)

		assert.True(t, admin.MustChangePassword)
	})
}
