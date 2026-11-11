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

package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/config"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestChangePasswordRedirect(t *testing.T) {
	tests := []struct {
		name      string
		publicURL string
		location  string
	}{
		{
			name:      "with public url",
			publicURL: "https://vikunja.example.com/sub/",
			location:  "https://vikunja.example.com/sub/user/settings/password-update",
		},
		{
			name:      "without public url",
			publicURL: "",
			location:  "/user/settings/password-update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := config.ServicePublicURL.GetString()
			config.ServicePublicURL.Set(tt.publicURL)
			t.Cleanup(func() { config.ServicePublicURL.Set(old) })

			e := echo.New()
			e.GET("/.well-known/change-password", ChangePasswordRedirect)

			req := httptest.NewRequest(http.MethodGet, "/.well-known/change-password", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusFound, rec.Code)
			assert.Equal(t, tt.location, rec.Header().Get("Location"))
		})
	}
}
