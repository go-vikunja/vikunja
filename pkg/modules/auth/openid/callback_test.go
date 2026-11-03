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
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/modules/avatar"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/user"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mockIDPProviderKey = "provider1"

type mockIDP struct {
	*httptest.Server
	key    *rsa.PrivateKey
	mu     sync.Mutex
	claims map[string]any
}

func newMockIDP(t *testing.T) *mockIDP {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	idp := &mockIDP{
		key:    key,
		claims: map[string]any{"sub": "user1"},
	}
	b64 := base64.RawURLEncoding.EncodeToString

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"issuer":                 idp.URL,
			"authorization_endpoint": idp.URL + "/auth",
			"token_endpoint":         idp.URL + "/token",
			"userinfo_endpoint":      idp.URL + "/userinfo",
			"jwks_uri":               idp.URL + "/jwks",
		})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
			"kty": "RSA",
			"alg": "RS256",
			"use": "sig",
			"kid": "test",
			"n":   b64(key.N.Bytes()),
			"e":   b64(big.NewInt(int64(key.E)).Bytes()),
		}}})
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, _ *http.Request) {
		payload := idp.currentClaims()
		payload["iss"] = idp.URL
		payload["aud"] = "client1"
		payload["iat"] = time.Now().Unix()
		payload["exp"] = time.Now().Add(time.Hour).Unix()
		header := b64([]byte(`{"alg":"RS256","kid":"test","typ":"JWT"}`))
		body, _ := json.Marshal(payload)
		signingInput := header + "." + b64(body)
		digest := sha256.Sum256([]byte(signingInput))
		signature, _ := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "access",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"id_token":     signingInput + "." + b64(signature),
		})
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(idp.currentClaims())
	})

	idp.Server = httptest.NewServer(mux)
	t.Cleanup(idp.Close)
	return idp
}

func (idp *mockIDP) setClaims(claims map[string]any) {
	idp.mu.Lock()
	defer idp.mu.Unlock()
	idp.claims = claims
}

func (idp *mockIDP) currentClaims() map[string]any {
	idp.mu.Lock()
	defer idp.mu.Unlock()
	out := make(map[string]any, len(idp.claims))
	for k, v := range idp.claims {
		out[k] = v
	}
	return out
}

func (idp *mockIDP) configure(t *testing.T) {
	t.Helper()
	config.AuthOpenIDEnabled.Set(true)
	config.AuthOpenIDProviders.Set(map[string]interface{}{
		mockIDPProviderKey: map[string]interface{}{
			"name":         "Provider One",
			"authurl":      idp.URL,
			"clientid":     "client1",
			"clientsecret": "secret1",
		},
	})
	_ = keyvalue.Del("openid_providers")
	_ = keyvalue.Del("openid_provider_" + mockIDPProviderKey)
	t.Cleanup(CleanupSavedOpenIDProviders)
}

func TestAuthenticateCallback(t *testing.T) {
	t.Run("OIDC-created user gets email, name, avatar and teams synced", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		pictureURL, hits := servePicture(t)
		idp := newMockIDP(t)
		idp.configure(t)

		claims := map[string]any{
			"sub":                "oidc-created",
			"email":              "oidc-created@example.com",
			"name":               "OIDC Created",
			"preferred_username": "oidcCreatedUser",
			"picture":            pictureURL,
		}
		idp.setClaims(claims)
		created, _, err := AuthenticateCallback(t.Context(), &Callback{Code: "code"}, mockIDPProviderKey)
		require.NoError(t, err)
		avatar.WaitForProviderSyncsForTests()
		assert.Equal(t, idp.URL, created.Issuer)

		claims["email"] = "changed@example.com"
		claims["name"] = "Changed Name"
		claims["vikunja_groups"] = []map[string]any{
			{
				"name":   "idp team",
				"oidcID": "idp-team-1",
			},
		}
		idp.setClaims(claims)
		u, _, err := AuthenticateCallback(t.Context(), &Callback{Code: "code"}, mockIDPProviderKey)
		require.NoError(t, err)
		avatar.WaitForProviderSyncsForTests()

		assert.Equal(t, created.ID, u.ID)
		db.AssertExists(t, "users", map[string]interface{}{
			"id":    u.ID,
			"email": "changed@example.com",
			"name":  "Changed Name",
		}, false)
		assert.Equal(t, int32(2), hits.Load(), "picture should be fetched on every login")
		assertTeamSynced(t, u.ID, "idp-team-1")
	})
	t.Run("disabled OIDC-created user is refused", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		idp := newMockIDP(t)
		idp.configure(t)

		s := db.NewSession()
		_, err := s.Insert(&user.User{
			Username: "disabledOidcUser",
			Email:    "disabled-oidc@example.com",
			Issuer:   idp.URL,
			Subject:  "disabled-oidc",
			Status:   user.StatusDisabled,
		})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		idp.setClaims(map[string]any{
			"sub":                "disabled-oidc",
			"email":              "disabled-oidc@example.com",
			"name":               "Disabled",
			"preferred_username": "disabledOidcUser",
		})
		_, _, err = AuthenticateCallback(t.Context(), &Callback{Code: "code"}, mockIDPProviderKey)
		require.Error(t, err)
		assert.True(t, user.IsErrAccountDisabled(err), "expected ErrAccountDisabled, got %v", err)
	})
}

func TestAuthenticateCallbackLinkedIdentity(t *testing.T) {
	t.Run("logs in the local user without syncing profile or avatar, but syncs teams", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		pictureURL, hits := servePicture(t)
		idp := newMockIDP(t)
		idp.configure(t)
		linkIdentity(t, 1, idp.URL, "linked-sub")

		idp.setClaims(map[string]any{
			"sub":                "linked-sub",
			"email":              "idp-email@example.com",
			"email_verified":     true,
			"name":               "IdP Name",
			"preferred_username": "idpUsername",
			"picture":            pictureURL,
			"vikunja_groups": []map[string]any{
				{
					"name":   "idp team",
					"oidcID": "idp-team-1",
				},
			},
		})
		u, _, err := AuthenticateCallback(t.Context(), &Callback{Code: "code"}, mockIDPProviderKey)
		require.NoError(t, err)
		avatar.WaitForProviderSyncsForTests()

		assert.Equal(t, int64(1), u.ID)
		db.AssertExists(t, "users", map[string]interface{}{
			"id":     1,
			"email":  "user1@example.com",
			"issuer": user.IssuerLocal,
		}, false)
		assert.Empty(t, u.Name)
		db.AssertMissing(t, "users", map[string]interface{}{
			"issuer": idp.URL,
		})
		assert.Equal(t, int32(0), hits.Load(), "picture must not be fetched for a linked user")
		assertTeamSynced(t, 1, "idp-team-1")
	})
	t.Run("keeps an openid avatar when the IdP sends no picture", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		idp := newMockIDP(t)
		idp.configure(t)
		linkIdentity(t, 1, idp.URL, "avatar-sub")

		s := db.NewSession()
		defer s.Close()
		_, err := s.ID(1).Cols("avatar_provider").Update(&user.User{AvatarProvider: "openid"})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		idp.setClaims(map[string]any{
			"sub":                "avatar-sub",
			"email":              "idp-email@example.com",
			"preferred_username": "idpUsername",
		})
		u, _, err := AuthenticateCallback(t.Context(), &Callback{Code: "code"}, mockIDPProviderKey)
		require.NoError(t, err)

		assert.Equal(t, int64(1), u.ID)
		db.AssertExists(t, "users", map[string]interface{}{
			"id":              1,
			"avatar_provider": "openid",
		}, false)
	})
	t.Run("refuses a linked user who cannot authenticate", func(t *testing.T) {
		cases := map[string]struct {
			userID  int64
			isError func(error) bool
		}{
			"disabled": {
				userID:  17,
				isError: user.IsErrAccountDisabled,
			},
			"locked": {
				userID:  18,
				isError: user.IsErrAccountLocked,
			},
			"email not confirmed": {
				userID:  4,
				isError: user.IsErrEmailNotConfirmed,
			},
		}
		for name, tc := range cases {
			t.Run(name, func(t *testing.T) {
				db.LoadAndAssertFixtures(t)
				idp := newMockIDP(t)
				idp.configure(t)
				linkIdentity(t, tc.userID, idp.URL, "linked-sub")

				idp.setClaims(map[string]any{
					"sub":                "linked-sub",
					"email":              "idp-email@example.com",
					"name":               "IdP Name",
					"preferred_username": "idpUsername",
				})
				_, _, err := AuthenticateCallback(t.Context(), &Callback{Code: "code"}, mockIDPProviderKey)
				require.Error(t, err)
				assert.True(t, tc.isError(err), "unexpected error %v", err)
				db.AssertMissing(t, "users", map[string]interface{}{
					"issuer": idp.URL,
				})
			})
		}
	})
	t.Run("enforces TOTP for a linked user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		idp := newMockIDP(t)
		idp.configure(t)
		// Own TOTP user: user 10's fixture TOTP would mark its passcode used for other tests.
		const secret = "KRSXG5CTMVRXEZLU"
		s := db.NewSession()
		_, err := s.Insert(&user.TOTP{
			UserID:  2,
			Secret:  secret,
			Enabled: true,
		})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()
		t.Cleanup(func() {
			_ = keyvalue.Del((&user.User{ID: 2}).GetFailedTOTPAttemptsKey())
		})
		linkIdentity(t, 2, idp.URL, "linked-sub")
		idp.setClaims(map[string]any{
			"sub":                "linked-sub",
			"email":              "idp-email@example.com",
			"name":               "IdP Name",
			"preferred_username": "idpUsername",
		})

		_, _, err = AuthenticateCallback(t.Context(), &Callback{Code: "code"}, mockIDPProviderKey)
		require.Error(t, err)
		assert.True(t, user.IsErrInvalidTOTPPasscode(err), "unexpected error %v", err)

		passcode, err := totp.GenerateCode(secret, time.Now())
		require.NoError(t, err)
		u, _, err := AuthenticateCallback(t.Context(), &Callback{
			Code:         "code",
			TOTPPasscode: passcode,
		}, mockIDPProviderKey)
		require.NoError(t, err)
		assert.Equal(t, int64(2), u.ID)
	})
}

func linkIdentity(t *testing.T, userID int64, issuer, subject string) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	_, err := s.Insert(&user.Identity{
		UserID:  userID,
		Issuer:  issuer,
		Subject: subject,
	})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

func assertTeamSynced(t *testing.T, userID int64, externalID string) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	teamIDs := []int64{}
	require.NoError(t, s.Table("teams").Where("external_id = ?", externalID).Cols("id").Find(&teamIDs))
	require.Len(t, teamIDs, 1, "team %s should have been created", externalID)
	db.AssertExists(t, "team_members", map[string]interface{}{
		"team_id": teamIDs[0],
		"user_id": userID,
	}, false)
}
