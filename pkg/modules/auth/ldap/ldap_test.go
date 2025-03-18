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

package ldap

import (
	"os"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	user2 "code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLdapLogin(t *testing.T) {
	if os.Getenv("VIKUNJA_TESTS_USE_CONFIG") != "1" || !config.AuthLdapEnabled.GetBool() {
		t.Skip("Skipping LDAP tests because ldap is not configured")
	}

	// We assume this ldap test server is used: https://gitea.com/gitea/test-openldap

	t.Run("should create account", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		user, err := AuthenticateUserInLDAP(s, "professor", "professor", false)

		require.NoError(t, err)
		assert.Equal(t, "professor", user.Username)
		db.AssertExists(t, "users", map[string]interface{}{
			"username": "professor",
			"issuer":   "ldap",
		}, false)
		db.AssertMissing(t, "teams", map[string]interface{}{
			"issuer": "ldap",
		})
	})

	t.Run("should not create account for wrong password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := AuthenticateUserInLDAP(s, "professor", "wrongpassword", false)

		require.Error(t, err)
		assert.True(t, user2.IsErrWrongUsernameOrPassword(err))
	})

	t.Run("should not create account for wrong user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := AuthenticateUserInLDAP(s, "gnome", "professor", false)

		require.Error(t, err)
		assert.True(t, user2.IsErrWrongUsernameOrPassword(err))
	})

	t.Run("should sync groups", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		user, err := AuthenticateUserInLDAP(s, "professor", "professor", true)

		require.NoError(t, err)
		assert.Equal(t, "professor", user.Username)
		db.AssertExists(t, "users", map[string]interface{}{
			"username": "professor",
			"issuer":   "ldap",
		}, false)
		db.AssertExists(t, "teams", map[string]interface{}{
			"name":        "admin_staff (LDAP)",
			"issuer":      "ldap",
			"external_id": "cn=admin_staff,ou=people,dc=planetexpress,dc=com",
		}, false)
		db.AssertExists(t, "teams", map[string]interface{}{
			"name":        "git (LDAP)",
			"issuer":      "ldap",
			"external_id": "cn=git,ou=people,dc=planetexpress,dc=com",
		}, false)
	})
}
