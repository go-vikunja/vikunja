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

package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAuthFromContext_NoEchoContext(t *testing.T) {
	_, err := GetAuthFromContext(context.Background())
	assert.Error(t, err, "should fail when echo.Context isn't stashed on ctx")
}

func TestGetRefreshTokenCookiePaths(t *testing.T) {
	original := config.ServicePublicURL.GetString()
	t.Cleanup(func() { config.ServicePublicURL.Set(original) })

	tests := []struct {
		name      string
		publicURL string
		basePath  string
	}{
		{"empty", "", ""},
		{"root", "https://h/", ""},
		{"subpath", "https://h/vikunja", "/vikunja"},
		{"subpath with trailing slash", "https://h/vikunja/", "/vikunja"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.ServicePublicURL.Set(tt.publicURL)
			assert.Equal(t, []string{tt.basePath + RefreshTokenPathV1, tt.basePath + RefreshTokenPathV2}, getRefreshTokenCookiePaths())
		})
	}
}

func TestIsUnusableRefreshToken(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"session expired", &models.ErrSessionExpired{}, true},
		{"user status error", &user.ErrAccountDisabled{UserID: 1}, true},
		{"transient error", errors.New("db"), false},
		// A concurrent refresh rotated the token away; the cookie it set must survive.
		{"refresh token already used", &models.ErrRefreshTokenAlreadyUsed{}, false},
		{"invalid refresh token", &models.ErrInvalidRefreshToken{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsUnusableRefreshToken(tt.err))
		})
	}
}

func TestIssuedJWTsCarryIssuedAt(t *testing.T) {
	for _, key := range []config.Key{
		config.ServiceSecret,
		config.ServiceJWTTTL,
		config.ServiceJWTTTLShort,
	} {
		original := key.GetString()
		t.Cleanup(func() { key.Set(original) })
	}
	config.ServiceSecret.Set("test-secret")
	config.ServiceJWTTTL.Set(3600)
	config.ServiceJWTTTLShort.Set(600)

	parseClaims := func(t *testing.T, token string) jwt.MapClaims {
		parsed, err := jwt.Parse(token, func(_ *jwt.Token) (any, error) {
			return []byte("test-secret"), nil
		})
		require.NoError(t, err)
		return parsed.Claims.(jwt.MapClaims)
	}

	tests := []struct {
		name  string
		issue func() (string, error)
		ttl   int64
	}{
		{
			name: "user",
			issue: func() (string, error) {
				return NewUserJWTAuthtoken(&user.User{
					ID:       1,
					Username: "user1",
				}, "session-id")
			},
			ttl: 600,
		},
		{
			name: "link share",
			issue: func() (string, error) {
				return NewLinkShareJWTAuthtoken(&models.LinkSharing{
					ID:        1,
					Hash:      "hash",
					ProjectID: 1,
				})
			},
			ttl: 3600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().Unix()
			token, err := tt.issue()
			require.NoError(t, err)
			after := time.Now().Unix()

			claims := parseClaims(t, token)
			iat, ok := claims["iat"].(float64)
			require.True(t, ok, "iat claim missing")
			exp, ok := claims["exp"].(float64)
			require.True(t, ok, "exp claim missing")

			assert.GreaterOrEqual(t, int64(iat), before)
			assert.LessOrEqual(t, int64(iat), after)
			assert.Equal(t, tt.ttl, int64(exp-iat))
		})
	}
}
