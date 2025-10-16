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

package openid

import (
	"testing"

	"code.vikunja.io/api/pkg/models"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOrCreateUser(t *testing.T) {
	t.Run("new user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		cl := &claims{
			Email:             "test@example.com",
			PreferredUsername: "someUserWhoDoesNotExistYet",
		}
		provider := &Provider{}
		idToken := &oidc.IDToken{Issuer: "https://some.issuer", Subject: "12345"}

		u, err := getOrCreateUser(s, cl, provider, idToken)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		db.AssertExists(t, "users", map[string]interface{}{
			"id":       u.ID,
			"email":    cl.Email,
			"username": "someUserWhoDoesNotExistYet",
		}, false)
	})
	t.Run("new user, no username provided", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		cl := &claims{
			Email:             "test@example.com",
			PreferredUsername: "",
		}
		provider := &Provider{}
		idToken := &oidc.IDToken{Issuer: "https://some.issuer", Subject: "12345"}

		u, err := getOrCreateUser(s, cl, provider, idToken)
		require.NoError(t, err)
		assert.NotEmpty(t, u.Username)
		err = s.Commit()
		require.NoError(t, err)

		db.AssertExists(t, "users", map[string]interface{}{
			"id":    u.ID,
			"email": cl.Email,
		}, false)
	})
	t.Run("new user, no email address", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		cl := &claims{
			Email: "",
		}
		provider := &Provider{}
		idToken := &oidc.IDToken{Issuer: "https://some.issuer", Subject: "12345"}

		_, err := getOrCreateUser(s, cl, provider, idToken)
		require.Error(t, err)
	})
	t.Run("existing user, different email address", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		cl := &claims{
			Email: "other-email-address@some.service.com",
		}
		provider := &Provider{}
		idToken := &oidc.IDToken{Issuer: "https://some.service.com", Subject: "12345"}

		u, err := getOrCreateUser(s, cl, provider, idToken)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		db.AssertExists(t, "users", map[string]interface{}{
			"id":    u.ID,
			"email": cl.Email,
		}, false)
	})
	t.Run("existing user, non existing team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team := "new sso team"
		oidcID := "47404"
		cl := &claims{
			Email: "other-email-address@some.service.com",
			VikunjaGroups: []map[string]interface{}{
				{"name": team, "oidcID": oidcID},
			},
		}

		provider := &Provider{}
		idToken := &oidc.IDToken{Issuer: "https://some.service.com", Subject: "12345"}

		u, err := getOrCreateUser(s, cl, provider, idToken)
		require.NoError(t, err)
		teamData := getTeamDataFromToken(cl.VikunjaGroups, nil)
		require.NoError(t, err)
		err = models.SyncExternalTeamsForUser(s, u, teamData, "https://some.issuer", "OIDC")
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		db.AssertExists(t, "users", map[string]interface{}{
			"id":    u.ID,
			"email": cl.Email,
		}, false)
		db.AssertExists(t, "teams", map[string]interface{}{
			"name":        team + " (OIDC)",
			"external_id": oidcID,
			"is_public":   false,
		}, false)
	})

	t.Run("Update IsPublic flag for existing team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team := "testteam15"
		oidcID := "15"
		cl := &claims{
			Email: "other-email-address@some.service.com",
			VikunjaGroups: []map[string]interface{}{
				{"name": team, "oidcID": oidcID, "isPublic": true},
			},
		}

		provider := &Provider{}
		idToken := &oidc.IDToken{Issuer: "https://some.service.com", Subject: "12345"}

		u, err := getOrCreateUser(s, cl, provider, idToken)
		require.NoError(t, err)
		teamData := getTeamDataFromToken(cl.VikunjaGroups, nil)
		err = models.SyncExternalTeamsForUser(s, u, teamData, "https://some.issuer", "OIDC")
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		db.AssertExists(t, "teams", map[string]interface{}{
			"name":        team + " (OIDC)",
			"external_id": oidcID,
			"is_public":   true,
		}, false)
	})

	t.Run("existing user, assign to existing team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		team := "testteam14"
		oidcID := "14"
		cl := &claims{
			Email: "other-email-address@some.service.com",
			VikunjaGroups: []map[string]interface{}{
				{"name": team, "oidcID": oidcID},
			},
		}

		u := &user.User{ID: 10}
		teamData := getTeamDataFromToken(cl.VikunjaGroups, nil)
		err := models.SyncExternalTeamsForUser(s, u, teamData, "https://some.issuer", "OIDC")
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)

		db.AssertExists(t, "team_members", map[string]interface{}{
			"user_id": u.ID,
		}, false)
	})
	t.Run("existing user, remove from existing team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		cl := &claims{
			Email:         "other-email-address@some.service.com",
			VikunjaGroups: []map[string]interface{}{},
		}

		u := &user.User{ID: 10}
		teamData := getTeamDataFromToken(cl.VikunjaGroups, nil)
		err := models.SyncExternalTeamsForUser(s, u, teamData, "https://some.issuer", "OIDC")
		require.NoError(t, err)

		db.AssertMissing(t, "team_members", map[string]interface{}{
			"team_id": 14,
			"user_id": u.ID,
		})
		db.AssertMissing(t, "team_members", map[string]interface{}{
			"team_id": 15,
			"user_id": u.ID,
		})
		// This team is not external and should not be touched
		db.AssertExists(t, "team_members", map[string]interface{}{
			"team_id": 13,
			"user_id": u.ID,
		}, false)
	})
	t.Run("ProviderFallback: Match to existing local user on username", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		cl := &claims{}
		provider := &Provider{
			UsernameFallback: true,
		}
		idToken := &oidc.IDToken{Issuer: "https://some.issuer", Subject: "user11"}

		u, err := getOrCreateUser(s, cl, provider, idToken)
		require.NoError(t, err)
		assert.Equal(t, idToken.Subject, u.Username, "subject match username")
		assert.Equal(t, user.IssuerLocal, u.Issuer, "User should be a local one")
		assert.Equal(t, 11, int(u.ID), "user id 11 expected")
	})
	t.Run("ProviderFallback: Match to existing local user on email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		cl := &claims{
			Email: "user11@example.com",
		}
		provider := &Provider{
			EmailFallback: true,
		}
		idToken := &oidc.IDToken{Issuer: "https://some.issuer", Subject: "user11"}

		u, err := getOrCreateUser(s, cl, provider, idToken)
		require.NoError(t, err)
		assert.Equal(t, cl.Email, u.Email, "email should match")
		assert.Equal(t, user.IssuerLocal, u.Issuer, "User should be a local one")
		assert.Equal(t, 11, int(u.ID), "user id 11 expected")
	})
	t.Run("ProviderFallback: Match to existing local user  on username and email", func(t *testing.T) {

		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		cl := &claims{
			Email: "user11@example.com",
		}
		provider := &Provider{
			UsernameFallback: true,
			EmailFallback:    true,
		}
		idToken := &oidc.IDToken{Issuer: "https://some.issuer", Subject: "user11"}

		u, err := getOrCreateUser(s, cl, provider, idToken)
		require.NoError(t, err)
		assert.Equal(t, cl.Email, u.Email, "email should match")
		assert.Equal(t, idToken.Subject, u.Username, "subject match username")
		assert.Equal(t, user.IssuerLocal, u.Issuer, "User should be a local one")
		assert.Equal(t, 11, int(u.ID), "user id 11 expected")
	})
	t.Run("CrossOidcFallback: Should only match local users when disabled", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create a user with a different OIDC issuer
		existingUser, err := user.CreateUser(s, &user.User{
			Username: "oidc_user",
			Email:    "oidc_user@example.com",
			Issuer:   "https://another.provider.com",
			Subject:  "external123",
			Status:   user.StatusActive,
		})
		require.NoError(t, err)

		cl := &claims{
			Email: "oidc_user@example.com",
		}
		provider := &Provider{
			EmailFallback:     true,
			CrossOidcFallback: false, // Disabled - should NOT match cross-OIDC users
		}
		idToken := &oidc.IDToken{Issuer: "https://some.issuer", Subject: "newuser123"}

		// Should create a new user instead of matching the existing OIDC user
		u, err := getOrCreateUser(s, cl, provider, idToken)
		require.NoError(t, err)
		assert.NotEqual(t, existingUser.ID, u.ID, "Should create a new user, not match existing OIDC user")
		assert.Equal(t, idToken.Issuer, u.Issuer, "New user should have the new issuer")
		assert.Equal(t, idToken.Subject, u.Subject, "New user should have the new subject")
	})
	t.Run("CrossOidcFallback: Should match users from other OIDC providers when enabled", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create a user with a different OIDC issuer
		existingUser, err := user.CreateUser(s, &user.User{
			Username: "cross_oidc_user",
			Email:    "cross_oidc@example.com",
			Issuer:   "https://provider-a.com",
			Subject:  "subjectA123",
			Status:   user.StatusActive,
		})
		require.NoError(t, err)

		cl := &claims{
			Email: "cross_oidc@example.com",
		}
		provider := &Provider{
			EmailFallback:     true,
			CrossOidcFallback: true, // Enabled - should match cross-OIDC users
		}
		idToken := &oidc.IDToken{Issuer: "https://provider-b.com", Subject: "subjectB456"}

		// Should match the existing OIDC user from a different provider
		u, err := getOrCreateUser(s, cl, provider, idToken)
		require.NoError(t, err)
		assert.Equal(t, existingUser.ID, u.ID, "Should match existing user from different OIDC provider")
		assert.Equal(t, "cross_oidc@example.com", u.Email, "Email should match")
		// Note: The existing user's issuer and subject should remain unchanged
		assert.Equal(t, "https://provider-a.com", u.Issuer, "Original issuer should be preserved")
		assert.Equal(t, "subjectA123", u.Subject, "Original subject should be preserved")
	})
	t.Run("CrossOidcFallback: Match by username across OIDC providers", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create a user with a different OIDC issuer
		existingUser, err := user.CreateUser(s, &user.User{
			Username: "shared_username",
			Email:    "user1@provider-a.com",
			Issuer:   "https://provider-a.com",
			Subject:  "shared_username",
			Status:   user.StatusActive,
		})
		require.NoError(t, err)

		cl := &claims{
			Email: "user2@provider-b.com",
		}
		provider := &Provider{
			UsernameFallback:  true,
			CrossOidcFallback: true, // Enabled - should match by username across providers
		}
		idToken := &oidc.IDToken{Issuer: "https://provider-b.com", Subject: "shared_username"}

		// Should match the existing user by username even from different provider
		u, err := getOrCreateUser(s, cl, provider, idToken)
		require.NoError(t, err)
		assert.Equal(t, existingUser.ID, u.ID, "Should match existing user by username across providers")
		assert.Equal(t, "shared_username", u.Username, "Username should match")
	})
	t.Run("CrossOidcFallback: Should still match local users when enabled", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		cl := &claims{
			Email: "user11@example.com",
		}
		provider := &Provider{
			EmailFallback:     true,
			CrossOidcFallback: true, // Enabled - should still match local users
		}
		idToken := &oidc.IDToken{Issuer: "https://some.issuer", Subject: "user11"}

		u, err := getOrCreateUser(s, cl, provider, idToken)
		require.NoError(t, err)
		assert.Equal(t, cl.Email, u.Email, "Email should match")
		assert.Equal(t, user.IssuerLocal, u.Issuer, "Should match local user")
		assert.Equal(t, 11, int(u.ID), "Should match user id 11")
	})
}

// TestMergeClaims tests the mergeClaims function with different configurations including forceUserInfo
func TestMergeClaims(t *testing.T) {
	t.Run("ForceUserInfo enabled - should use userinfo values", func(t *testing.T) {
		// Setup token claims
		tokenClaims := &claims{
			Email:             "token-email@example.com",
			Name:              "Token Name",
			PreferredUsername: "token_username",
		}

		// Setup userinfo claims
		userinfoClaims := &claims{
			Email:             "userinfo-email@example.com",
			Name:              "UserInfo Name",
			PreferredUsername: "userinfo_username",
		}

		// Test with ForceUserInfo enabled
		err := mergeClaims(tokenClaims, userinfoClaims, true)
		require.NoError(t, err)

		// Verify userinfo data was used
		assert.Equal(t, "userinfo-email@example.com", tokenClaims.Email)
		assert.Equal(t, "UserInfo Name", tokenClaims.Name)
		assert.Equal(t, "userinfo_username", tokenClaims.PreferredUsername)
	})

	t.Run("ForceUserInfo disabled - should use token values if present", func(t *testing.T) {
		// Setup token claims with all values
		tokenClaims := &claims{
			Email:             "token-email@example.com",
			Name:              "Token Name",
			PreferredUsername: "token_username",
		}

		// Setup userinfo claims
		userinfoClaims := &claims{
			Email:             "userinfo-email@example.com",
			Name:              "UserInfo Name",
			PreferredUsername: "userinfo_username",
		}

		// Test with ForceUserInfo disabled
		err := mergeClaims(tokenClaims, userinfoClaims, false)
		require.NoError(t, err)

		// Verify token data was preserved
		assert.Equal(t, "token-email@example.com", tokenClaims.Email)
		assert.Equal(t, "Token Name", tokenClaims.Name)
		assert.Equal(t, "token_username", tokenClaims.PreferredUsername)
	})

	t.Run("Missing values - should use userinfo when token is missing values", func(t *testing.T) {
		// Setup token claims with missing values
		tokenClaims := &claims{
			Email: "token-email@example.com",
			// Missing Name and PreferredUsername
		}

		// Setup userinfo claims
		userinfoClaims := &claims{
			Email:             "userinfo-email@example.com",
			Name:              "UserInfo Name",
			PreferredUsername: "userinfo_username",
		}

		// Test with ForceUserInfo disabled, but missing values in token
		err := mergeClaims(tokenClaims, userinfoClaims, false)
		require.NoError(t, err)

		// Verify token email was kept, but missing fields were filled from userinfo
		assert.Equal(t, "token-email@example.com", tokenClaims.Email)
		assert.Equal(t, "UserInfo Name", tokenClaims.Name)
		assert.Equal(t, "userinfo_username", tokenClaims.PreferredUsername)
	})

	t.Run("Use nickname when preferred_username is missing", func(t *testing.T) {
		// Setup token claims with missing preferred_username
		tokenClaims := &claims{
			Email: "token-email@example.com",
			Name:  "Token Name",
			// Missing PreferredUsername
		}

		// Setup userinfo claims with nickname but no preferred_username
		userinfoClaims := &claims{
			Email:    "userinfo-email@example.com",
			Name:     "UserInfo Name",
			Nickname: "userinfo_nickname",
			// Missing PreferredUsername to test fallback to nickname
		}

		// Test with ForceUserInfo disabled
		err := mergeClaims(tokenClaims, userinfoClaims, false)
		require.NoError(t, err)

		// Verify nickname was used for preferred_username
		assert.Equal(t, "userinfo_nickname", tokenClaims.PreferredUsername)
	})

	t.Run("Error when email is missing", func(t *testing.T) {
		// Setup token claims with missing email
		tokenClaims := &claims{
			// Missing Email
			Name:              "Token Name",
			PreferredUsername: "token_username",
		}

		// Setup userinfo claims also with missing email
		userinfoClaims := &claims{
			// Missing Email
			Name:              "UserInfo Name",
			PreferredUsername: "userinfo_username",
		}

		// Test with ForceUserInfo disabled
		err := mergeClaims(tokenClaims, userinfoClaims, false)

		// Verify error is returned for missing email
		require.Error(t, err)
		assert.IsType(t, &user.ErrNoOpenIDEmailProvided{}, err)
	})
}
