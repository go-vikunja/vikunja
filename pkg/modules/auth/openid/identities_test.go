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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

const identityTestPassword = "12345678"

func loadIdentityTestUser(t *testing.T, id int64) *user.User {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	u, err := user.GetUserByID(s, id)
	require.NoError(t, err)

	reset := func() {
		_ = keyvalue.Del(u.GetFailedPasswordAttemptsKey())
	}
	reset()
	t.Cleanup(reset)
	return u
}

func enableTestTOTP(t *testing.T, userID int64) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	_, err := s.Insert(&user.TOTP{
		UserID:  userID,
		Secret:  "JBSWY3DPEHPK3PXP",
		Enabled: true,
	})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

func linkRequest(password string) *LinkIdentityRequest {
	return &LinkIdentityRequest{
		Provider: mockIDPProviderKey,
		Code:     "code",
		Password: password,
	}
}

func runLink(t *testing.T, u *user.User, req *LinkIdentityRequest) (*LinkedIdentity, error) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	defer events.CleanupPending(s)
	identity, err := LinkIdentity(t.Context(), s, u, req)
	if err == nil {
		require.NoError(t, s.Commit())
		events.DispatchPending(t.Context(), s)
	}
	return identity, err
}

func runUnlink(t *testing.T, u *user.User, identityID int64, password, passcode string) error {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	err := UnlinkIdentity(t.Context(), s, u, identityID, password, passcode)
	if err == nil {
		require.NoError(t, s.Commit())
	}
	return err
}

func TestLinkIdentity(t *testing.T) {
	t.Run("links the identity", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		idp := newMockIDP(t)
		idp.configure(t)
		idp.setClaims(map[string]any{"sub": "new-sub"})
		u := loadIdentityTestUser(t, 2)
		notifications.Fake()
		t.Cleanup(notifications.Unfake)

		identity, err := runLink(t, u, linkRequest(identityTestPassword))
		require.NoError(t, err)

		notifications.AssertSent(t, &user.IdentityLinkedNotification{})
		assert.Equal(t, idp.URL, identity.Issuer)
		assert.Equal(t, mockIDPProviderKey, identity.ProviderKey)
		assert.Equal(t, "Provider One", identity.ProviderName)
		db.AssertExists(t, "user_identities", map[string]interface{}{
			"id":      identity.ID,
			"user_id": 2,
			"issuer":  idp.URL,
			"subject": "new-sub",
		}, false)
	})
	t.Run("wrong password is refused before the code is exchanged and counts toward the lockout", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		idp := newMockIDP(t)
		idp.configure(t)
		idp.setClaims(map[string]any{"sub": "new-sub"})
		u := loadIdentityTestUser(t, 2)

		_, err := runLink(t, u, linkRequest("wrong"))
		require.Error(t, err)

		assert.True(t, user.IsErrWrongUsernameOrPassword(err), "unexpected error %v", err)
		assert.Equal(t, int32(0), idp.tokenRequests.Load(), "the single-use code must survive a typo")
		attempts, _, err := keyvalue.GetInt64(u.GetFailedPasswordAttemptsKey())
		require.NoError(t, err)
		assert.Equal(t, int64(1), attempts)
		db.AssertMissing(t, "user_identities", map[string]interface{}{
			"user_id": 2,
		})
	})
	t.Run("missing passcode is refused for a TOTP user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		idp := newMockIDP(t)
		idp.configure(t)
		idp.setClaims(map[string]any{"sub": "totp-sub"})
		u := loadIdentityTestUser(t, 3)
		enableTestTOTP(t, u.ID)

		_, err := runLink(t, u, linkRequest(identityTestPassword))
		require.Error(t, err)

		assert.True(t, user.IsErrInvalidTOTPPasscode(err), "unexpected error %v", err)
		assert.Equal(t, int32(0), idp.tokenRequests.Load())
		db.AssertMissing(t, "user_identities", map[string]interface{}{
			"user_id": u.ID,
		})
	})
	t.Run("refuses an identity linked to another user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		idp := newMockIDP(t)
		idp.configure(t)
		linkIdentity(t, 1, idp.URL, "taken-sub")
		idp.setClaims(map[string]any{"sub": "taken-sub"})
		u := loadIdentityTestUser(t, 2)

		_, err := runLink(t, u, linkRequest(identityTestPassword))
		require.Error(t, err)

		assert.True(t, models.IsErrUserIdentityLinkedToOtherUser(err), "unexpected error %v", err)
		db.AssertMissing(t, "user_identities", map[string]interface{}{
			"user_id": 2,
		})
	})
	t.Run("refuses the identity of an OIDC-created user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		idp := newMockIDP(t)
		idp.configure(t)
		s := db.NewSession()
		_, err := s.Insert(&user.User{
			Username: "oidcCreated",
			Email:    "oidc-created@example.com",
			Issuer:   idp.URL,
			Subject:  "oidc-created",
			Status:   user.StatusActive,
		})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()
		idp.setClaims(map[string]any{"sub": "oidc-created"})
		u := loadIdentityTestUser(t, 2)

		_, err = runLink(t, u, linkRequest(identityTestPassword))
		require.Error(t, err)

		assert.True(t, models.IsErrUserIdentityLinkedToOtherUser(err), "unexpected error %v", err)
		db.AssertMissing(t, "user_identities", map[string]interface{}{
			"user_id": 2,
		})
	})
	t.Run("refuses a second identity for the same issuer", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		idp := newMockIDP(t)
		idp.configure(t)
		linkIdentity(t, 2, idp.URL, "first-sub")
		idp.setClaims(map[string]any{"sub": "second-sub"})
		u := loadIdentityTestUser(t, 2)

		_, err := runLink(t, u, linkRequest(identityTestPassword))
		require.Error(t, err)

		assert.True(t, models.IsErrUserIdentityProviderAlreadyLinked(err), "unexpected error %v", err)
		db.AssertMissing(t, "user_identities", map[string]interface{}{
			"subject": "second-sub",
		})
	})
	t.Run("relinking the own identity returns it unchanged", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		idp := newMockIDP(t)
		idp.configure(t)
		idp.setClaims(map[string]any{"sub": "own-sub"})
		u := loadIdentityTestUser(t, 2)

		first, err := runLink(t, u, linkRequest(identityTestPassword))
		require.NoError(t, err)

		notifications.Fake()
		t.Cleanup(notifications.Unfake)
		second, err := runLink(t, u, linkRequest(identityTestPassword))
		require.NoError(t, err)
		notifications.AssertNotSent(t, &user.IdentityLinkedNotification{})

		assert.Equal(t, first.ID, second.ID)
		assert.Equal(t, mockIDPProviderKey, second.ProviderKey)
		db.AssertCount(t, "user_identities", builder.Eq{"user_id": u.ID}, 1)
	})
}

func TestPreviewIdentity(t *testing.T) {
	t.Run("refuses an identity linked to another user before asking for a password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		idp := newMockIDP(t)
		idp.configure(t)
		linkIdentity(t, 1, idp.URL, "taken-sub")
		idp.setClaims(map[string]any{"sub": "taken-sub"})
		u := loadIdentityTestUser(t, 2)

		preview, err := PreviewIdentity(t.Context(), u, linkRequest(""))
		require.Error(t, err)
		assert.Nil(t, preview)
		assert.True(t, models.IsErrUserIdentityLinkedToOtherUser(err), "unexpected error %v", err)
	})
	t.Run("refuses the identity of an OIDC-created user before asking for a password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		idp := newMockIDP(t)
		idp.configure(t)
		s := db.NewSession()
		_, err := s.Insert(&user.User{Username: "previewOIDC", Email: "preview-oidc@example.com", Issuer: idp.URL, Subject: "oidc-sub", Status: user.StatusActive})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()
		idp.setClaims(map[string]any{"sub": "oidc-sub"})

		_, err = PreviewIdentity(t.Context(), loadIdentityTestUser(t, 2), linkRequest(""))
		require.Error(t, err)
		assert.True(t, models.IsErrUserIdentityLinkedToOtherUser(err), "unexpected error %v", err)
	})
	t.Run("refuses a second account at the same provider before asking for a password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		idp := newMockIDP(t)
		idp.configure(t)
		linkIdentity(t, 2, idp.URL, "first-sub")
		idp.setClaims(map[string]any{"sub": "second-sub"})

		_, err := PreviewIdentity(t.Context(), loadIdentityTestUser(t, 2), linkRequest(""))
		require.Error(t, err)
		assert.True(t, models.IsErrUserIdentityProviderAlreadyLinked(err), "unexpected error %v", err)
	})
	t.Run("rechecks ownership when reusing a preview", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		idp := newMockIDP(t)
		idp.configure(t)
		idp.setClaims(map[string]any{"sub": "preview-race-sub"})
		u := loadIdentityTestUser(t, 2)
		req := linkRequest("")
		t.Cleanup(func() { _ = keyvalue.Del(identityPreviewKey(u.ID, req)) })
		_, err := PreviewIdentity(t.Context(), u, req)
		require.NoError(t, err)
		linkIdentity(t, 1, idp.URL, "preview-race-sub")

		_, err = PreviewIdentity(t.Context(), u, req)
		require.Error(t, err)
		assert.True(t, models.IsErrUserIdentityLinkedToOtherUser(err), "unexpected error %v", err)
	})
	t.Run("shows the verified account and preserves it across reload and password retries", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		idp := newMockIDP(t)
		idp.configure(t)
		idp.setClaims(map[string]any{
			"sub":   "preview-sub",
			"email": "alice@example.com",
		})
		u := loadIdentityTestUser(t, 2)
		req := linkRequest("wrong")
		for range 2 {
			preview, err := PreviewIdentity(t.Context(), u, req)
			require.NoError(t, err)
			assert.Equal(t, "alice@example.com", preview.Account)
		}
		_, err := runLink(t, u, req)
		require.Error(t, err)
		assert.True(t, user.IsErrWrongUsernameOrPassword(err))
		db.AssertMissing(t, "user_identities", map[string]interface{}{"user_id": u.ID})
		req.Password = identityTestPassword
		identity, err := runLink(t, u, req)
		require.NoError(t, err)
		assert.Equal(t, int32(1), idp.tokenRequests.Load())
		db.AssertExists(t, "user_identities", map[string]interface{}{
			"id":      identity.ID,
			"user_id": u.ID,
			"subject": "preview-sub",
		}, false)
		require.NoError(t, runUnlink(t, u, identity.ID, identityTestPassword, ""))
		idp.Close()
		_, err = runLink(t, u, req)
		require.Error(t, err, "the consumed preview must not relink after unlinking")
		db.AssertMissing(t, "user_identities", map[string]interface{}{"user_id": u.ID})
	})

	t.Run("binds the prepared identity to the current user and callback", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		idp := newMockIDP(t)
		idp.configure(t)
		idp.setClaims(map[string]any{"sub": "private-sub"})
		u := loadIdentityTestUser(t, 2)
		req := linkRequest(identityTestPassword)
		preview, err := PreviewIdentity(t.Context(), u, req)
		require.NoError(t, err)
		assert.Equal(t, "private-sub", preview.Account)
		idp.Close()
		other := loadIdentityTestUser(t, 3)
		_, err = runLink(t, other, req)
		require.Error(t, err)
		changed := *req
		changed.RedirectURL = "https://other.example.com/callback"
		_, err = runLink(t, u, &changed)
		require.Error(t, err)
		db.AssertMissing(t, "user_identities", map[string]interface{}{"user_id": other.ID})
		_, err = runLink(t, u, req)
		require.NoError(t, err)
		db.AssertExists(t, "user_identities", map[string]interface{}{
			"user_id": u.ID,
			"subject": "private-sub",
		}, false)
	})

	t.Run("refuses a prepared identity after its provider is removed", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		idp := newMockIDP(t)
		idp.configure(t)
		u := loadIdentityTestUser(t, 2)
		req := linkRequest(identityTestPassword)
		t.Cleanup(func() { _ = keyvalue.Del(identityPreviewKey(u.ID, req)) })
		_, err := PreviewIdentity(t.Context(), u, req)
		require.NoError(t, err)
		config.AuthOpenIDProviders.Set(map[string]interface{}{})
		_, err = runLink(t, u, req)
		require.Error(t, err)
		var providerErr *models.ErrOpenIDBadRequest
		require.ErrorAs(t, err, &providerErr)
		db.AssertMissing(t, "user_identities", map[string]interface{}{"user_id": u.ID})
	})
}

func TestUnlinkIdentity(t *testing.T) {
	t.Run("missing passcode is refused for a TOTP user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		u := loadIdentityTestUser(t, 6)
		enableTestTOTP(t, u.ID)

		err := runUnlink(t, u, 3, identityTestPassword, "")
		require.Error(t, err)

		assert.True(t, user.IsErrInvalidTOTPPasscode(err), "unexpected error %v", err)
		db.AssertExists(t, "user_identities", map[string]interface{}{
			"id": 3,
		}, false)
	})
}

func TestListIdentities(t *testing.T) {
	t.Run("resolves the provider from the issuer", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		idp := newMockIDP(t)
		idp.configure(t)
		linkIdentity(t, 2, idp.URL, "listed-sub")

		s := db.NewSession()
		defer s.Close()
		identities, err := ListIdentities(s, &user.User{ID: 2})
		require.NoError(t, err)

		require.Len(t, identities, 1)
		assert.Equal(t, idp.URL, identities[0].Issuer)
		assert.Equal(t, mockIDPProviderKey, identities[0].ProviderKey)
		assert.Equal(t, "Provider One", identities[0].ProviderName)
	})
	t.Run("leaves the provider empty when it is no longer configured", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		idp := newMockIDP(t)
		idp.configure(t)

		s := db.NewSession()
		defer s.Close()
		identities, err := ListIdentities(s, &user.User{ID: 1})
		require.NoError(t, err)

		require.Len(t, identities, 2)
		assert.Equal(t, int64(1), identities[0].ID)
		assert.Equal(t, "https://some.service.com", identities[0].Issuer)
		assert.Empty(t, identities[0].ProviderKey)
		assert.Empty(t, identities[0].ProviderName)
		assert.Equal(t, "https://other.service.com", identities[1].Issuer)
	})
}
