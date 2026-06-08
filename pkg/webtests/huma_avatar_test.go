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

// TestAvatar covers the v2 binary-response endpoint GET /api/v2/avatar/{username}.
// It is the reference for serving raw bytes with a runtime-chosen Content-Type.
// Unlike v1's CRUD resources there is no model — the input is a path username and
// an optional size query param. The endpoint is authenticated (global security),
// so an anonymous request must be rejected with 401.
func TestAvatar(t *testing.T) {
	t.Run("Authenticated known user returns bytes and a content type", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/avatar/user1", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.NotEmpty(t, rec.Body.Bytes(), "avatar bytes must be returned")
		assert.NotEmpty(t, rec.Header().Get("Content-Type"), "a content type must be set")
		// user1 has no avatar_provider configured, so the empty provider serves the default SVG.
		assert.Equal(t, "image/svg+xml", rec.Header().Get("Content-Type"))
	})

	t.Run("Size query param is accepted and clamped", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		// A size far above config.ServiceMaxAvatarSize must be clamped, not rejected.
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/avatar/user1?size=99999", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "oversized size must clamp, not error; body: %s", rec.Body.String())
		assert.NotEmpty(t, rec.Body.Bytes())

		// A normal size is honored.
		rec = humaRequest(t, e, http.MethodGet, "/api/v2/avatar/user1?size=64", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.NotEmpty(t, rec.Body.Bytes())
	})

	t.Run("Anonymous request is rejected with 401", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		// Empty token => no Authorization header => anonymous. Proves the endpoint
		// inherits the global security and is NOT public.
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/avatar/user1", "", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "anonymous must get 401; body: %s", rec.Body.String())
	})

	t.Run("Unknown user falls back to the default avatar like v1", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		// v1 GetAvatar does not 404 for an unknown user — it serves the empty
		// provider's default SVG with a 200. v2 must match that behaviour.
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/avatar/this-user-does-not-exist", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.NotEmpty(t, rec.Body.Bytes())
		assert.Equal(t, "image/svg+xml", rec.Header().Get("Content-Type"))
	})

	t.Run("Authenticated as a link share returns bytes and a content type", func(t *testing.T) {
		// Ports v1's TestLinkShareAvatar: a request authenticated as a link-share
		// user (not a regular JWT) must succeed. The avatar endpoint inherits the
		// global security; a link-share JWT is a valid credential and resolves to a
		// link-share web.Auth in authFromCtx, so the request must not 401. The
		// requested username here is the synthetic link-share-1 handle (matching v1):
		// it is not a real user row, so GetAvatarForUsername serves the empty
		// provider's default SVG with a 200 — proving the link-share auth path works.
		e, err := setupTestEnv()
		require.NoError(t, err)

		share := &models.LinkSharing{
			ID:          1,
			Hash:        "test",
			ProjectID:   1,
			Permission:  models.PermissionRead,
			SharingType: models.SharingTypeWithoutPassword,
			SharedByID:  1,
		}
		token, err := auth.NewLinkShareJWTAuthtoken(share)
		require.NoError(t, err)

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/avatar/link-share-1", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "link-share auth must succeed; body: %s", rec.Body.String())
		assert.NotEmpty(t, rec.Body.Bytes(), "avatar bytes must be returned for a link-share caller")
		assert.NotEmpty(t, rec.Header().Get("Content-Type"), "a content type must be set")
	})

	t.Run("Bot user avatar is rendered by the botmarble provider", func(t *testing.T) {
		// A bot user (bot_owner_id set) routes through the botmarble provider rather
		// than its configured provider — a distinct rendering path v1's single test
		// never exercised. It must still return 200 + non-empty bytes. botmarble
		// renders a marble-style SVG, so the content type is image/svg+xml.
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/avatar/bot-owner-a-assistant", "", token, "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.NotEmpty(t, rec.Body.Bytes(), "bot avatar bytes must be returned")
		assert.Equal(t, "image/svg+xml", rec.Header().Get("Content-Type"), "botmarble renders an SVG")
		// The marble mask id is unique to the (bot)marble renderer; the empty
		// fallback SVG never contains it. Asserting on it proves IsBot() routed
		// to botmarble rather than silently falling through to the default avatar.
		assert.Contains(t, rec.Body.String(), "mask__marble",
			"bot avatar must be rendered by the (bot)marble provider, not the default placeholder")
	})

	t.Run("Non-numeric size is rejected", func(t *testing.T) {
		// v1 parses ?size with strconv.ParseInt and returns ErrInvalidModel
		// ("Invalid size parameter") -> 400 for a non-numeric value. v2 declares
		// Size as int64, so Huma's own request validation rejects the malformed
		// query value before the handler runs and returns 422 Unprocessable Entity.
		// Either client-error status proves the malformed input is refused rather
		// than silently coerced; assert v2's actual behaviour (422).
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := humaTokenFor(t, &testuser1)

		rec := humaRequest(t, e, http.MethodGet, "/api/v2/avatar/user1?size=notanumber", "", token, "")
		require.Equal(t, http.StatusUnprocessableEntity, rec.Code,
			"non-numeric size must be rejected by Huma's int64 validation; body: %s", rec.Body.String())
	})
}
