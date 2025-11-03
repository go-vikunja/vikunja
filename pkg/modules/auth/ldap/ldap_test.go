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

	t.Run("should sync avatar when enabled", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		user, err := AuthenticateUserInLDAP(s, "professor", "professor", false, "jpegPhoto")

		require.NoError(t, err)
		assert.Equal(t, "professor", user.Username)
		db.AssertExists(t, "users", map[string]interface{}{
			"username":        "professor",
			"issuer":          "ldap",
			"avatar_provider": "ldap",
		}, false)
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
