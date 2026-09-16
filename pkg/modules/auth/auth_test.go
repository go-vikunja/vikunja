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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
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
