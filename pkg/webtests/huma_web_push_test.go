// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package webtests

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/routes"
	"code.vikunja.io/api/pkg/user"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

func setupWebPushHTTPTest(t *testing.T) (*echo.Echo, string, string) {
	t.Helper()
	_, err := setupTestEnv()
	require.NoError(t, err)

	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	require.NoError(t, err)
	config.WebPushEnabled.Set(true)
	config.WebPushPublicKey.Set(publicKey)
	config.WebPushPrivateKey.Set(privateKey)
	t.Cleanup(func() {
		config.WebPushEnabled.Set(false)
		config.WebPushPublicKey.Set("")
		config.WebPushPrivateKey.Set("")
	})

	createSession := func(id string, userID int64) {
		s := db.NewSession()
		defer s.Close()
		_, err := s.Insert(&models.Session{
			ID:         id,
			UserID:     userID,
			TokenHash:  fmt.Sprintf("web-push-test-%s", id),
			LastActive: time.Now(),
		})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
	}
	const session1 = "550e8400-e29b-41d4-a716-446655449001"
	const session2 = "550e8400-e29b-41d4-a716-446655449002"
	createSession(session1, testuser1.ID)
	createSession(session2, testuser2.ID)

	tokenFor := func(u *user.User, sessionID string) string {
		token, err := auth.NewUserJWTAuthtoken(u, sessionID)
		require.NoError(t, err)
		return token
	}

	e := routes.NewEcho()
	routes.RegisterRoutes(e)
	return e, tokenFor(&testuser1, session1), tokenFor(&testuser2, session2)
}

func TestHumaWebPushSubscriptions(t *testing.T) {
	e, user1Token, user2Token := setupWebPushHTTPTest(t)
	deviceID := "9a34527d-1357-4c64-8171-6d6e25f01d62"
	endpoint := "https://push.example.test/subscription/one"
	body := fmt.Sprintf(`{"endpoint":%q,"expiration_time":null,"keys":{"p256dh":%q,"auth":%q}}`,
		endpoint,
		config.WebPushPublicKey.GetString(),
		base64.RawURLEncoding.EncodeToString([]byte("0123456789abcdef")),
	)
	path := "/api/v2/user/settings/web-push/subscriptions/" + deviceID

	rec := humaRequest(t, e, http.MethodPut, path, body, user1Token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	assert.NotContains(t, rec.Body.String(), endpoint)
	assert.NotContains(t, rec.Body.String(), `"p256dh"`)
	assert.NotContains(t, rec.Body.String(), `"auth"`)

	rec = humaRequest(t, e, http.MethodPut, path, body, user1Token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	db.AssertCount(t, "web_push_subscriptions", builder.Eq{"user_id": testuser1.ID, "device_id": deviceID}, 1)

	rec = humaRequest(t, e, http.MethodPut, "/api/v2/user/settings/web-push/subscriptions/ec35ec57-c51d-4d5f-a1d3-219982df56c0", body, user2Token, "")
	assert.Equal(t, http.StatusConflict, rec.Code, "body: %s", rec.Body.String())

	rec = humaRequest(t, e, http.MethodDelete, path, "", user1Token, "")
	require.Equal(t, http.StatusNoContent, rec.Code, "body: %s", rec.Body.String())
	db.AssertMissing(t, "web_push_subscriptions", map[string]any{"user_id": testuser1.ID, "device_id": deviceID})
}

func TestHumaWebPushRequiresActiveSession(t *testing.T) {
	e, _, _ := setupWebPushHTTPTest(t)
	token, err := auth.NewUserJWTAuthtoken(&testuser1, "550e8400-e29b-41d4-a716-446655449099")
	require.NoError(t, err)
	body := fmt.Sprintf(`{"endpoint":"https://push.example.test/missing-session","keys":{"p256dh":%q,"auth":%q}}`,
		config.WebPushPublicKey.GetString(),
		base64.RawURLEncoding.EncodeToString([]byte("0123456789abcdef")),
	)
	rec := humaRequest(t, e, http.MethodPut, "/api/v2/user/settings/web-push/subscriptions/9a34527d-1357-4c64-8171-6d6e25f01d62", body, token, "")
	assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
}

func TestHumaWebPushDisabled(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	rec := humaRequest(t, e, http.MethodPut, "/api/v2/user/settings/web-push/subscriptions/9a34527d-1357-4c64-8171-6d6e25f01d62", `{}`, humaTokenFor(t, &testuser1), "")
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHumaInfoExposesOnlyPublicWebPushKey(t *testing.T) {
	e, _, _ := setupWebPushHTTPTest(t)
	for _, path := range []string{"/api/v1/info", "/api/v2/info"} {
		t.Run(path, func(t *testing.T) {
			rec := humaRequest(t, e, http.MethodGet, path, "", "", "")
			require.Equal(t, http.StatusOK, rec.Code)
			var body map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, true, body["web_push_enabled"])
			assert.Equal(t, config.WebPushPublicKey.GetString(), body["web_push_public_key"])
			assert.NotContains(t, rec.Body.String(), config.WebPushPrivateKey.GetString())
		})
	}
}
