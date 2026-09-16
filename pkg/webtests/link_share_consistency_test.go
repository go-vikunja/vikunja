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

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// collidingLinkShareJWT builds a real link-share JWT for share id 2 (hash test2),
// which has write access to project 2 (task 13) and whose id collides with user 2.
func collidingLinkShareJWT(t *testing.T) string {
	jwt, err := auth.NewLinkShareJWTAuthtoken(&models.LinkSharing{
		ID:          2,
		Hash:        "test2",
		ProjectID:   2,
		Permission:  models.PermissionWrite,
		SharingType: models.SharingTypeWithoutPassword,
		SharedByID:  1,
	})
	require.NoError(t, err)
	return jwt
}

// GHSA-vvcv-vpph-h844 consistency hardening: reactions are per-user, so a link
// share (or any non-user) must not act as one via a colliding GetID().
func TestReactionLinkShareForbidden(t *testing.T) {
	t.Run("create forbidden for link share", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		res := apiTokenReq(e, http.MethodPut, "/api/v1/tasks/13/reactions", collidingLinkShareJWT(t), `{"value":"👍"}`)
		assert.Equal(t, http.StatusForbidden, res.Code)
	})
	t.Run("delete forbidden for link share", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		res := apiTokenReq(e, http.MethodPost, "/api/v1/tasks/13/reactions/delete", collidingLinkShareJWT(t), `{"value":"👍"}`)
		assert.Equal(t, http.StatusForbidden, res.Code)
	})
	t.Run("create allowed for regular user", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		// user 3 owns project 2 and can update task 13.
		res := apiTokenReq(e, http.MethodPut, "/api/v1/tasks/13/reactions", userJWT(t, 3), `{"value":"👍"}`)
		assert.Equal(t, http.StatusCreated, res.Code)
	})
}

// GHSA-vvcv-vpph-h844 consistency hardening: the unread status is per-user, so a
// link share must not clear it via a colliding GetID().
func TestTaskReadLinkShareForbidden(t *testing.T) {
	t.Run("forbidden for link share", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		res := apiTokenReq(e, http.MethodPost, "/api/v1/tasks/13/read", collidingLinkShareJWT(t), "")
		assert.Equal(t, http.StatusForbidden, res.Code)
	})
	t.Run("allowed for regular user", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		res := apiTokenReq(e, http.MethodPost, "/api/v1/tasks/13/read", userJWT(t, 3), "")
		assert.Equal(t, http.StatusOK, res.Code)
	})
}
