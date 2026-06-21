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

package webtests

import (
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	userDeletionRequestPath = "/api/v2/user/deletion/request"
	userDeletionConfirmPath = "/api/v2/user/deletion/confirm"
	userDeletionCancelPath  = "/api/v2/user/deletion/cancel"
	// testUserPassword is the plaintext password for every local fixture user.
	testUserPassword = "12345678"
)

// deletionTokenFor reads the cleartext account-deletion token RequestDeletion
// stored for the user. RequestDeletion only mails the token, so the test pulls
// it straight from user_tokens (kind 3 = TokenAccountDeletion).
func deletionTokenFor(t *testing.T, userID int64) string {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	tok := struct {
		Token string `xorm:"token"`
	}{}
	has, err := s.Table("user_tokens").
		Where("user_id = ? AND kind = ?", userID, 3).
		Get(&tok)
	require.NoError(t, err)
	require.True(t, has, "RequestDeletion must have stored a deletion token for user %d", userID)
	return tok.Token
}

func deletionScheduledFor(t *testing.T, userID int64) bool {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	u, err := user.GetUserByID(s, userID)
	require.NoError(t, err)
	return !u.DeletionScheduledAt.IsZero()
}

// TestHumaUserDeletion ports v1's account-deletion flow (request → confirm →
// cancel) to v2. v1 returned 200/204 with a confirmation message body; v2
// normalises all three to an empty 204 (the action returns no resource), so
// every success here asserts 204 + empty body.
func TestHumaUserDeletion(t *testing.T) {
	t.Run("Request - wrong password rejected", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodPost, userDeletionRequestPath, `{"password":"wrong"}`, token, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
		assert.False(t, deletionScheduledFor(t, testuser1.ID), "a rejected request must not schedule deletion")
	})

	t.Run("Confirm - invalid token rejected", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodPost, userDeletionConfirmPath, `{"token":"not-a-real-token"}`, token, "")
		assert.Equal(t, http.StatusBadRequest, rec.Code, "body: %s", rec.Body.String())
		assert.False(t, deletionScheduledFor(t, testuser1.ID))
	})

	t.Run("Confirm - missing token is a validation error", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodPost, userDeletionConfirmPath, `{"token":""}`, token, "")
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("Request then confirm schedules deletion", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		req := humaRequest(t, e, http.MethodPost, userDeletionRequestPath, `{"password":"`+testUserPassword+`"}`, token, "")
		require.Equal(t, http.StatusNoContent, req.Code, "body: %s", req.Body.String())
		assert.Empty(t, req.Body.String(), "v2 normalises the request action to an empty 204")
		assert.False(t, deletionScheduledFor(t, testuser1.ID), "request alone must not schedule; confirmation does")

		confirm := humaRequest(t, e, http.MethodPost, userDeletionConfirmPath,
			`{"token":"`+deletionTokenFor(t, testuser1.ID)+`"}`, token, "")
		require.Equal(t, http.StatusNoContent, confirm.Code, "body: %s", confirm.Body.String())
		assert.Empty(t, confirm.Body.String())
		assert.True(t, deletionScheduledFor(t, testuser1.ID), "confirm must schedule the deletion")
	})

	t.Run("Cancel - wrong password rejected", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		// Schedule first so there is something to cancel.
		req := humaRequest(t, e, http.MethodPost, userDeletionRequestPath, `{"password":"`+testUserPassword+`"}`, token, "")
		require.Equal(t, http.StatusNoContent, req.Code, "body: %s", req.Body.String())
		confirm := humaRequest(t, e, http.MethodPost, userDeletionConfirmPath,
			`{"token":"`+deletionTokenFor(t, testuser1.ID)+`"}`, token, "")
		require.Equal(t, http.StatusNoContent, confirm.Code, "body: %s", confirm.Body.String())
		require.True(t, deletionScheduledFor(t, testuser1.ID))

		cancel := humaRequest(t, e, http.MethodPost, userDeletionCancelPath, `{"password":"wrong"}`, token, "")
		assert.Equal(t, http.StatusForbidden, cancel.Code, "body: %s", cancel.Body.String())
		assert.True(t, deletionScheduledFor(t, testuser1.ID), "a rejected cancel must leave the deletion scheduled")
	})

	t.Run("Cancel - correct password clears the schedule", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		req := humaRequest(t, e, http.MethodPost, userDeletionRequestPath, `{"password":"`+testUserPassword+`"}`, token, "")
		require.Equal(t, http.StatusNoContent, req.Code, "body: %s", req.Body.String())
		confirm := humaRequest(t, e, http.MethodPost, userDeletionConfirmPath,
			`{"token":"`+deletionTokenFor(t, testuser1.ID)+`"}`, token, "")
		require.Equal(t, http.StatusNoContent, confirm.Code, "body: %s", confirm.Body.String())
		require.True(t, deletionScheduledFor(t, testuser1.ID))

		cancel := humaRequest(t, e, http.MethodPost, userDeletionCancelPath, `{"password":"`+testUserPassword+`"}`, token, "")
		require.Equal(t, http.StatusNoContent, cancel.Code, "body: %s", cancel.Body.String())
		assert.Empty(t, cancel.Body.String())
		assert.False(t, deletionScheduledFor(t, testuser1.ID), "cancel must clear the scheduled deletion")
	})

	t.Run("Unauthenticated", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		for _, path := range []string{userDeletionRequestPath, userDeletionConfirmPath, userDeletionCancelPath} {
			rec := humaRequest(t, e, http.MethodPost, path, `{}`, "", "")
			assert.Equal(t, http.StatusUnauthorized, rec.Code, "%s body: %s", path, rec.Body.String())
		}
	})
}

// TestHumaUserDeletion_LinkShareForbidden asserts a link share — which has no
// account — is refused (403) on every deletion action.
func TestHumaUserDeletion_LinkShareForbidden(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token, err := auth.NewLinkShareJWTAuthtoken(&models.LinkSharing{
		ID:          1,
		Hash:        "test",
		ProjectID:   1,
		Permission:  models.PermissionRead,
		SharingType: models.SharingTypeWithoutPassword,
		SharedByID:  1,
	})
	require.NoError(t, err)

	for _, tc := range []struct {
		name string
		path string
		body string
	}{
		{"request", userDeletionRequestPath, `{"password":"` + testUserPassword + `"}`},
		{"confirm", userDeletionConfirmPath, `{"token":"x"}`},
		{"cancel", userDeletionCancelPath, `{"password":"` + testUserPassword + `"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := humaRequest(t, e, http.MethodPost, tc.path, tc.body, token, "")
			assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
		})
	}
}
