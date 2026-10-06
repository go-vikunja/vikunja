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

package ldap

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	user2 "code.vikunja.io/api/pkg/user"

	"github.com/go-ldap/ldap/v3"
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

		user, err := AuthenticateUserInLDAP(s, "professor", "professor", false, "")

		require.NoError(t, err)
		assert.Equal(t, "professor", user.Username)
		require.NoError(t, s.Commit())
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

		_, err := AuthenticateUserInLDAP(s, "professor", "wrongpassword", false, "")

		require.Error(t, err)
		assert.True(t, user2.IsErrWrongUsernameOrPassword(err))
	})

	t.Run("should not create account for wrong user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := AuthenticateUserInLDAP(s, "gnome", "professor", false, "")

		require.Error(t, err)
		assert.True(t, user2.IsErrWrongUsernameOrPassword(err))
	})

	t.Run("should sync groups", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		user, err := AuthenticateUserInLDAP(s, "professor", "professor", true, "")

		require.NoError(t, err)
		assert.Equal(t, "professor", user.Username)
		require.NoError(t, s.Commit())
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
		assertLdapTeamCount(t, user.ID, 2)
	})

	t.Run("should sync groups using service account rebind", func(t *testing.T) {
		// Verifies that re-binding as the service account before the group
		// search works correctly — the fix for directories where regular users
		// cannot enumerate group membership.
		origFlag := config.AuthLdapGroupSyncUseServiceAccount.GetBool()
		config.AuthLdapGroupSyncUseServiceAccount.Set(true)
		defer config.AuthLdapGroupSyncUseServiceAccount.Set(origFlag)

		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		user, err := AuthenticateUserInLDAP(s, "professor", "professor", true, "")

		require.NoError(t, err)
		assert.Equal(t, "professor", user.Username)
		require.NoError(t, s.Commit())
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

	t.Run("should sync groups using user binding", func(t *testing.T) {
		// Verifies the flag=false path where the connection stays bound as the
		// authenticated user during the group search. Works on directories that
		// grant regular users read access to group objects.
		origFlag := config.AuthLdapGroupSyncUseServiceAccount.GetBool()
		config.AuthLdapGroupSyncUseServiceAccount.Set(false)
		defer config.AuthLdapGroupSyncUseServiceAccount.Set(origFlag)

		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		user, err := AuthenticateUserInLDAP(s, "professor", "professor", true, "")

		require.NoError(t, err)
		assert.Equal(t, "professor", user.Username)
		require.NoError(t, s.Commit())
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

	t.Run("should sync groups with per-user filter", func(t *testing.T) {
		origFilter := config.AuthLdapGroupSyncFilter.GetString()
		config.AuthLdapGroupSyncFilter.Set("(&(objectclass=groupOfNames)(member={userdn}))")
		defer config.AuthLdapGroupSyncFilter.Set(origFilter)

		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		user, err := AuthenticateUserInLDAP(s, "professor", "professor", true, "")

		require.NoError(t, err)
		require.NoError(t, s.Commit())
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
		assertLdapTeamCount(t, user.ID, 2)
	})

	t.Run("should sync groups with per-user filter for non-ascii dn", func(t *testing.T) {
		origFilter := config.AuthLdapGroupSyncFilter.GetString()
		config.AuthLdapGroupSyncFilter.Set("(&(objectclass=groupOfNames)(member={userdn}))")
		defer config.AuthLdapGroupSyncFilter.Set(origFilter)

		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// cn=Bender Bending Rodríguez,...
		user, err := AuthenticateUserInLDAP(s, "bender", "bender", true, "")

		require.NoError(t, err)
		require.NoError(t, s.Commit())
		db.AssertExists(t, "teams", map[string]interface{}{
			"name":        "ship_crew (LDAP)",
			"issuer":      "ldap",
			"external_id": "cn=ship_crew,ou=people,dc=planetexpress,dc=com",
		}, false)
		db.AssertExists(t, "teams", map[string]interface{}{
			"name":        "git (LDAP)",
			"issuer":      "ldap",
			"external_id": "cn=git,ou=people,dc=planetexpress,dc=com",
		}, false)
		assertLdapTeamCount(t, user.ID, 2)
	})

	t.Run("should sync avatar when enabled", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		user, err := AuthenticateUserInLDAP(s, "professor", "professor", false, "jpegPhoto")

		require.NoError(t, err)
		assert.Equal(t, "professor", user.Username)
		require.NoError(t, s.Commit())
		db.AssertExists(t, "users", map[string]interface{}{
			"username":        "professor",
			"issuer":          "ldap",
			"avatar_provider": "ldap",
		}, false)
	})

	t.Run("should bind anonymously", func(t *testing.T) {
		// Backup original config
		origBindDN := config.AuthLdapBindDN.GetString()
		origBindPW := config.AuthLdapBindPassword.GetString()
		defer func() {
			config.AuthLdapBindDN.Set(origBindDN)
			config.AuthLdapBindPassword.Set(origBindPW)
		}()

		// Set empty bind credentials
		config.AuthLdapBindDN.Set("")
		config.AuthLdapBindPassword.Set("")

		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Attempt to authenticate
		// Note: This test might fail if the test LDAP server doesn't support anonymous bind,
		// but it verifies the code path executes
		user, err := AuthenticateUserInLDAP(s, "professor", "professor", false, "")

		// We mainly want to ensure we don't panic or error out due to missing config
		if err != nil {
			// If it fails, it should be an LDAP error, not a "configuration missing" error
			require.NotContains(t, err.Error(), "configured")
		} else {
			assert.Equal(t, "professor", user.Username)
		}
	})
}

func TestSanitizedUserQuery(t *testing.T) {
	// Set up a test filter for this test
	originalFilter := config.AuthLdapUserFilter.GetString()
	config.AuthLdapUserFilter.Set("(&(objectClass=user)(sAMAccountName=%[1]s))")
	defer func() {
		if originalFilter != "" {
			config.AuthLdapUserFilter.Set(originalFilter)
		}
	}()

	tests := []struct {
		name           string
		input          string
		expectedResult bool
		expectedFilter string
	}{
		{
			name:           "normal username",
			input:          "testuser",
			expectedResult: true,
			expectedFilter: "(&(objectClass=user)(sAMAccountName=testuser))",
		},
		{
			name:           "username with injection attempt",
			input:          "admin)(|(objectClass=*",
			expectedResult: true,
			expectedFilter: `(&(objectClass=user)(sAMAccountName=admin\29\28|\28objectClass=\2a))`,
		},
		{
			name:           "username with OR operator",
			input:          "test|admin",
			expectedResult: true,
			expectedFilter: `(&(objectClass=user)(sAMAccountName=test|admin))`,
		},
		{
			name:           "empty username",
			input:          "",
			expectedResult: false,
			expectedFilter: "",
		},
		{
			name:           "username with null byte",
			input:          "test\x00user",
			expectedResult: false,
			expectedFilter: "",
		},
		{
			name:           "username with other control characters",
			input:          "test\x01user",
			expectedResult: false,
			expectedFilter: "",
		},
		{
			name:           "username with allowed whitespace",
			input:          "test user",
			expectedResult: true,
			expectedFilter: "(&(objectClass=user)(sAMAccountName=test user))",
		},
		{
			name:           "username with tab (allowed)",
			input:          "test\tuser",
			expectedResult: true,
			expectedFilter: "(&(objectClass=user)(sAMAccountName=test\tuser))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := sanitizedUserQuery(tt.input)
			assert.Equal(t, tt.expectedResult, ok)
			if ok {
				assert.Equal(t, tt.expectedFilter, result)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestSanitizedUserQueryPreventsInjection(t *testing.T) {
	// Set up a test filter
	config.AuthLdapUserFilter.Set("(&(objectClass=user)(uid=%[1]s))")
	defer config.AuthLdapUserFilter.Set("")

	// Test various injection attempts
	injectionAttempts := []string{
		"admin)(uid=*",                    // Try to match any uid
		"*)(|(uid=admin",                  // OR injection
		"admin))(&(objectClass=*",         // Try to match any object class
		"admin))(|(|(uid=admin)(uid=root", // Complex OR injection
		"admin&admin",                     // AND injection
		"admin=admin",                     // Equals injection
		"admin<admin",                     // Less than injection
		"admin>admin",                     // Greater than injection
		"admin~admin",                     // Approximate match injection
	}

	for i, attempt := range injectionAttempts {
		t.Run(fmt.Sprintf("injection_attempt_%d", i+1), func(t *testing.T) {
			result, ok := sanitizedUserQuery(attempt)
			assert.True(t, ok, "Query should be sanitized, not rejected")

			// Verify that all special characters are properly escaped
			assert.NotContains(t, result, ")(uid=*", "Should not contain unescaped injection")
			assert.NotContains(t, result, "|(", "Should not contain unescaped OR operator")
			assert.NotContains(t, result, "))(", "Should not contain unescaped parentheses")
			assert.NotContains(t, result, "=*", "Should not contain unescaped equals with wildcard")

			// Verify escaping is present where expected
			if strings.Contains(attempt, "(") {
				assert.Contains(t, result, `\28`, "Should contain escaped opening parenthesis")
			}
			if strings.Contains(attempt, ")") {
				assert.Contains(t, result, `\29`, "Should contain escaped closing parenthesis")
			}

			packet, err := ldap.CompileFilter(result)
			require.NoError(t, err)
			require.Len(t, packet.Children, 2, "Should not add filter clauses")
			uid := packet.Children[1]
			assert.EqualValues(t, ldap.FilterEqualityMatch, uid.Tag)
			assert.Equal(t, attempt, uid.Children[1].Value, "Should match the whole input as the uid value")
		})
	}
}

func assertLdapTeamCount(t *testing.T, userID int64, expected int64) {
	s := db.NewSession()
	defer s.Close()

	count, err := s.
		Table("teams").
		Join("INNER", "team_members", "team_members.team_id = teams.id").
		Where("teams.issuer = ? AND team_members.user_id = ?", user2.IssuerLDAP, userID).
		Count()
	require.NoError(t, err)
	assert.Equal(t, expected, count)
}

func TestBuildGroupSyncFilter(t *testing.T) {
	tests := []struct {
		name            string
		template        string
		userDN          string
		username        string
		expectedFilter  string
		expectedPerUser bool
		expectedErr     bool
	}{
		{
			name:           "no placeholder",
			template:       "(&(objectclass=*)(|(objectclass=group)(objectclass=groupOfNames)))",
			userDN:         "cn=professor,ou=people,dc=planetexpress,dc=com",
			username:       "professor",
			expectedFilter: "(&(objectclass=*)(|(objectclass=group)(objectclass=groupOfNames)))",
		},
		{
			name:            "userdn",
			template:        "(&(objectclass=groupOfNames)(member={userdn}))",
			userDN:          "cn=professor,ou=people,dc=planetexpress,dc=com",
			expectedFilter:  "(&(objectclass=groupOfNames)(member=cn=professor,ou=people,dc=planetexpress,dc=com))",
			expectedPerUser: true,
		},
		{
			name:            "username",
			template:        "(&(objectclass=posixGroup)(memberUid={username}))",
			username:        "professor",
			expectedFilter:  "(&(objectclass=posixGroup)(memberUid=professor))",
			expectedPerUser: true,
		},
		{
			name:            "both placeholders, repeated",
			template:        "(|(member={userdn})(memberUid={username})(uniqueMember={userdn}))",
			userDN:          "cn=a,dc=b",
			username:        "a",
			expectedFilter:  "(|(member=cn=a,dc=b)(memberUid=a)(uniqueMember=cn=a,dc=b))",
			expectedPerUser: true,
		},
		{
			name:            "injection attempt in username",
			template:        "(&(objectclass=posixGroup)(memberUid={username}))",
			username:        "x)(|(objectclass=*",
			expectedFilter:  `(&(objectclass=posixGroup)(memberUid=x\29\28|\28objectclass=\2a))`,
			expectedPerUser: true,
		},
		{
			name:            "placeholder in value is not substituted again",
			template:        "(|(member={userdn})(memberUid={username}))",
			userDN:          "cn={username},dc=example",
			username:        "professor",
			expectedFilter:  "(|(member=cn={username},dc=example)(memberUid=professor))",
			expectedPerUser: true,
		},
		{
			name:        "empty username",
			template:    "(&(objectclass=posixGroup)(memberUid={username}*))",
			userDN:      "cn=professor,ou=people,dc=planetexpress,dc=com",
			expectedErr: true,
		},
		{
			name:        "empty userdn",
			template:    "(&(objectclass=groupOfNames)(member={userdn}))",
			username:    "professor",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, perUser, err := buildGroupSyncFilter(tt.template, tt.userDN, tt.username)
			if tt.expectedErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedFilter, filter)
			assert.Equal(t, tt.expectedPerUser, perUser)

			_, err = ldap.CompileFilter(filter)
			require.NoError(t, err)
		})
	}
}
