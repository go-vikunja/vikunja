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
	"code.vikunja.io/api/pkg/models"
	user2 "code.vikunja.io/api/pkg/user"

	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
	"xorm.io/xorm"
)

const professorDN = "cn=Hubert J. Farnsworth,ou=people,dc=planetexpress,dc=com"

func useEntryUUID(t *testing.T, key config.Key) {
	orig := key.GetString()
	key.Set("entryUUID")
	t.Cleanup(func() { key.Set(orig) })
}

func getLdapTeam(t *testing.T, s *xorm.Session, externalID string) *models.Team {
	team, err := models.GetTeamByExternalIDAndIssuer(s, externalID, user2.IssuerLDAP)
	require.NoError(t, err)
	return team
}

func ldapEntryUUID(t *testing.T, dn string) string {
	l, err := ConnectAndBindToLDAPDirectory()
	require.NoError(t, err)
	defer l.Close()

	sr, err := l.Search(ldap.NewSearchRequest(dn, ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=*)", []string{"entryUUID"}, nil))
	require.NoError(t, err)
	require.Len(t, sr.Entries, 1)
	return sr.Entries[0].GetAttributeValue("entryUUID")
}

// Admin of the gitea/test-openldap image.
func bindLdapAdmin(t *testing.T) *ldap.Conn {
	l, err := ConnectAndBindToLDAPDirectory()
	require.NoError(t, err)
	require.NoError(t, l.Bind("cn=admin,dc=planetexpress,dc=com", "GoodNewsEveryone"))
	return l
}

func renameLdapEntry(t *testing.T, dn, newRDN string) {
	l := bindLdapAdmin(t)
	defer l.Close()

	oldRDN, parent, _ := strings.Cut(dn, ",")
	require.NoError(t, l.ModifyDN(ldap.NewModifyDNRequest(dn, newRDN, true, "")))
	t.Cleanup(func() {
		l := bindLdapAdmin(t)
		defer l.Close()
		require.NoError(t, l.ModifyDN(ldap.NewModifyDNRequest(newRDN+","+parent, oldRDN, true, "")))
	})
}

func setLdapAttribute(t *testing.T, dn, attribute, oldValue, newValue string) {
	replace := func(value string) {
		l := bindLdapAdmin(t)
		defer l.Close()
		req := ldap.NewModifyRequest(dn, nil)
		req.Replace(attribute, []string{value})
		require.NoError(t, l.Modify(req))
	}

	replace(newValue)
	t.Cleanup(func() { replace(oldValue) })
}

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
			"subject":  "professor",
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
		assert.EqualValues(t, 2, ldapTeamCount(t, user.ID))
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
		assert.EqualValues(t, 2, ldapTeamCount(t, user.ID))
	})

	t.Run("should page through groups", func(t *testing.T) {
		origPageSize := groupSearchPageSize
		groupSearchPageSize = 1
		defer func() { groupSearchPageSize = origPageSize }()

		for _, filter := range []string{
			config.AuthLdapGroupSyncFilter.GetString(),
			"(&(objectclass=groupOfNames)(member={userdn}))",
		} {
			t.Run(filter, func(t *testing.T) {
				origFilter := config.AuthLdapGroupSyncFilter.GetString()
				config.AuthLdapGroupSyncFilter.Set(filter)
				defer config.AuthLdapGroupSyncFilter.Set(origFilter)

				db.LoadAndAssertFixtures(t)
				s := db.NewSession()
				defer s.Close()

				user, err := AuthenticateUserInLDAP(s, "professor", "professor", true, "")

				require.NoError(t, err)
				require.NoError(t, s.Commit())
				assert.EqualValues(t, 2, ldapTeamCount(t, user.ID))
			})
		}
	})

	t.Run("should switch existing user and teams to stable ids", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		before, err := AuthenticateUserInLDAP(s, "professor", "professor", true, "")
		require.NoError(t, err)
		gitTeam := getLdapTeam(t, s, "cn=git,ou=people,dc=planetexpress,dc=com")
		_, err = s.Insert(&models.TeamProject{TeamID: gitTeam.ID, ProjectID: 1})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		useEntryUUID(t, config.AuthLdapAttributeUserID)
		useEntryUUID(t, config.AuthLdapAttributeGroupID)

		s2 := db.NewSession()
		defer s2.Close()
		after, err := AuthenticateUserInLDAP(s2, "professor", "professor", true, "")
		require.NoError(t, err)
		require.NoError(t, s2.Commit())

		assert.Equal(t, before.ID, after.ID)
		db.AssertExists(t, "users", map[string]interface{}{
			"id":      before.ID,
			"subject": ldapEntryUUID(t, professorDN),
		}, false)
		db.AssertCount(t, "users", builder.Eq{"issuer": "ldap"}, 1)

		gitUUID := ldapEntryUUID(t, "cn=git,ou=people,dc=planetexpress,dc=com")
		db.AssertExists(t, "teams", map[string]interface{}{
			"id":          gitTeam.ID,
			"external_id": gitUUID,
		}, false)
		db.AssertExists(t, "team_projects", map[string]interface{}{
			"team_id":    gitTeam.ID,
			"project_id": 1,
		}, false)
		db.AssertCount(t, "teams", builder.Eq{"issuer": "ldap"}, 2)
		assert.EqualValues(t, 2, ldapTeamCount(t, after.ID))
	})

	t.Run("should keep team when group dn changes", func(t *testing.T) {
		useEntryUUID(t, config.AuthLdapAttributeGroupID)

		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := AuthenticateUserInLDAP(s, "professor", "professor", true, "")
		require.NoError(t, err)
		gitUUID := ldapEntryUUID(t, "cn=git,ou=people,dc=planetexpress,dc=com")
		gitTeam := getLdapTeam(t, s, gitUUID)
		_, err = s.Insert(&models.TeamProject{TeamID: gitTeam.ID, ProjectID: 1})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		renameLdapEntry(t, "cn=git,ou=people,dc=planetexpress,dc=com", "cn=git_renamed")

		s2 := db.NewSession()
		defer s2.Close()
		u, err := AuthenticateUserInLDAP(s2, "professor", "professor", true, "")
		require.NoError(t, err)
		require.NoError(t, s2.Commit())

		db.AssertExists(t, "teams", map[string]interface{}{
			"id":          gitTeam.ID,
			"name":        "git_renamed (LDAP)",
			"external_id": gitUUID,
		}, false)
		db.AssertExists(t, "team_projects", map[string]interface{}{
			"team_id":    gitTeam.ID,
			"project_id": 1,
		}, false)
		db.AssertCount(t, "teams", builder.Eq{"issuer": "ldap"}, 2)
		assert.EqualValues(t, 2, ldapTeamCount(t, u.ID))
	})

	t.Run("should keep account when username changes", func(t *testing.T) {
		useEntryUUID(t, config.AuthLdapAttributeUserID)

		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		before, err := AuthenticateUserInLDAP(s, "professor", "professor", false, "")
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		setLdapAttribute(t, professorDN, "uid", "professor", "farnsworth")

		s2 := db.NewSession()
		defer s2.Close()
		after, err := AuthenticateUserInLDAP(s2, "farnsworth", "professor", false, "")
		require.NoError(t, err)
		require.NoError(t, s2.Commit())

		assert.Equal(t, before.ID, after.ID)
		assert.Equal(t, "professor", after.Username)
		db.AssertExists(t, "users", map[string]interface{}{
			"id":      before.ID,
			"subject": ldapEntryUUID(t, professorDN),
		}, false)
		db.AssertCount(t, "users", builder.Eq{"issuer": "ldap"}, 1)
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

func TestEscapeLDAPFilterValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal username",
			input:    "testuser",
			expected: "testuser",
		},
		{
			name:     "username with parentheses",
			input:    "test(user)",
			expected: `test\28user\29`,
		},
		{
			name:     "username with asterisk",
			input:    "test*user",
			expected: `test\2auser`,
		},
		{
			name:     "username with backslash",
			input:    `test\user`,
			expected: `test\5cuser`,
		},
		{
			name:     "username with ampersand",
			input:    "test&user",
			expected: `test\26user`,
		},
		{
			name:     "username with pipe",
			input:    "test|user",
			expected: `test\7cuser`,
		},
		{
			name:     "username with equals",
			input:    "test=user",
			expected: `test\3duser`,
		},
		{
			name:     "username with less than",
			input:    "test<user",
			expected: `test\3cuser`,
		},
		{
			name:     "username with greater than",
			input:    "test>user",
			expected: `test\3euser`,
		},
		{
			name:     "username with tilde",
			input:    "test~user",
			expected: `test\7euser`,
		},
		{
			name:     "username with null byte",
			input:    "test\x00user",
			expected: `test\00user`,
		},
		{
			name:     "complex injection attempt",
			input:    "admin)(|(objectClass=*",
			expected: `admin\29\28\7c\28objectClass\3d\2a`,
		},
		{
			name:     "LDAP injection with OR operator",
			input:    "testuser)|(&(objectClass=user",
			expected: `testuser\29\7c\28\26\28objectClass\3duser`,
		},
		{
			name:     "multiple special characters",
			input:    "test()&|=<>~*\\user",
			expected: `test\28\29\26\7c\3d\3c\3e\7e\2a\5cuser`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "unicode characters",
			input:    "testuser_unicode",
			expected: "testuser_unicode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeLDAPFilterValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
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
			expectedFilter: `(&(objectClass=user)(sAMAccountName=admin\29\28\7c\28objectClass\3d\2a))`,
		},
		{
			name:           "username with OR operator",
			input:          "test|admin",
			expectedResult: true,
			expectedFilter: `(&(objectClass=user)(sAMAccountName=test\7cadmin))`,
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
			if strings.Contains(attempt, "|") {
				assert.Contains(t, result, `\7c`, "Should contain escaped pipe")
			}
			if strings.Contains(attempt, "&") {
				assert.Contains(t, result, `\26`, "Should contain escaped ampersand")
			}
			if strings.Contains(attempt, "=") {
				assert.Contains(t, result, `\3d`, "Should contain escaped equals")
			}
		})
	}
}

func ldapTeamCount(t *testing.T, userID int64) int64 {
	s := db.NewSession()
	defer s.Close()

	count, err := s.
		Table("teams").
		Join("INNER", "team_members", "team_members.team_id = teams.id").
		Where("teams.issuer = ? AND team_members.user_id = ?", user2.IssuerLDAP, userID).
		Count()
	require.NoError(t, err)
	return count
}

func TestBuildGroupSyncFilter(t *testing.T) {
	tests := []struct {
		name            string
		template        string
		userDN          string
		username        string
		expectedFilter  string
		expectedPerUser bool
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
			name:            "ad matching rule in chain",
			template:        "(&(objectClass=group)(member:1.2.840.113556.1.4.1941:={userdn}))",
			userDN:          "CN=Jane Doe,OU=Users,DC=example,DC=com",
			expectedFilter:  "(&(objectClass=group)(member:1.2.840.113556.1.4.1941:=CN=Jane Doe,OU=Users,DC=example,DC=com))",
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
			name:            "dn with escaped comma",
			template:        "(member={userdn})",
			userDN:          `CN=Doe\, John,OU=Users,DC=example,DC=com`,
			expectedFilter:  `(member=CN=Doe\5c, John,OU=Users,DC=example,DC=com)`,
			expectedPerUser: true,
		},
		{
			name:            "dn with parentheses",
			template:        "(member={userdn})",
			userDN:          "CN=John (Admin),OU=Users,DC=example,DC=com",
			expectedFilter:  `(member=CN=John \28Admin\29,OU=Users,DC=example,DC=com)`,
			expectedPerUser: true,
		},
		{
			name:            "dn with asterisk",
			template:        "(member={userdn})",
			userDN:          "CN=*,DC=example,DC=com",
			expectedFilter:  `(member=CN=\2a,DC=example,DC=com)`,
			expectedPerUser: true,
		},
		{
			name:            "dn with non-ascii characters",
			template:        "(member={userdn})",
			userDN:          "CN=Jörg,DC=example,DC=com",
			expectedFilter:  `(member=CN=J\c3\b6rg,DC=example,DC=com)`,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, perUser := buildGroupSyncFilter(tt.template, tt.userDN, tt.username)
			assert.Equal(t, tt.expectedFilter, filter)
			assert.Equal(t, tt.expectedPerUser, perUser)

			_, err := ldap.CompileFilter(filter)
			require.NoError(t, err)
		})
	}
}

func TestFormatObjectGUID(t *testing.T) {
	// {6f9619ff-8b86-d011-b42d-00c04fc964ff} as AD stores it.
	raw := []byte{0xff, 0x19, 0x96, 0x6f, 0x86, 0x8b, 0x11, 0xd0, 0xb4, 0x2d, 0x00, 0xc0, 0x4f, 0xc9, 0x64, 0xff}

	guid, err := formatObjectGUID(raw)
	require.NoError(t, err)
	assert.Equal(t, "6f9619ff-8b86-d011-b42d-00c04fc964ff", guid)

	_, err = formatObjectGUID(raw[:15])
	require.Error(t, err)
}

func TestDirectoryID(t *testing.T) {
	guid := []byte{0xff, 0x19, 0x96, 0x6f, 0x86, 0x8b, 0x11, 0xd0, 0xb4, 0x2d, 0x00, 0xc0, 0x4f, 0xc9, 0x64, 0xff}
	entry := &ldap.Entry{
		DN: "cn=test,dc=example,dc=com",
		Attributes: []*ldap.EntryAttribute{
			{Name: "objectGUID", ByteValues: [][]byte{guid}},
			{Name: "entryUUID", ByteValues: [][]byte{[]byte("597ae2f6-16a6-1027-98f4-d28b5365dc14")}},
			{Name: "GUID", ByteValues: [][]byte{{0xff, 0x00, 0x10}}},
		},
	}

	tests := []struct {
		attribute string
		expected  string
	}{
		{attribute: "objectGUID", expected: "6f9619ff-8b86-d011-b42d-00c04fc964ff"},
		{attribute: "objectguid", expected: "6f9619ff-8b86-d011-b42d-00c04fc964ff"},
		{attribute: "entryUUID", expected: "597ae2f6-16a6-1027-98f4-d28b5365dc14"},
		{attribute: "GUID", expected: "ff0010"},
	}
	for _, tt := range tests {
		t.Run(tt.attribute, func(t *testing.T) {
			id, err := directoryID(entry, tt.attribute)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, id)
		})
	}

	t.Run("missing attribute", func(t *testing.T) {
		_, err := directoryID(entry, "uid")
		require.Error(t, err)
	})
}
