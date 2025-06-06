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
