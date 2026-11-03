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

package doctor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckOpenIDProviderRedirects(t *testing.T) {
	t.Run("rejects redirect to another origin", func(t *testing.T) {
		otherHit := false
		other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			otherHit = true
			_, _ = w.Write([]byte(`{"issuer":"other"}`))
		}))
		defer other.Close()

		provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, other.URL+"/.well-known/openid-configuration", http.StatusFound)
		}))
		defer provider.Close()

		result, issuer := checkOpenIDProvider("test", map[string]interface{}{"authurl": provider.URL})

		assert.False(t, result.Passed)
		assert.Contains(t, result.Error, "redirect")
		assert.Empty(t, issuer)
		assert.False(t, otherHit)
	})

	t.Run("follows redirect within the same origin", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/real/.well-known/openid-configuration", http.StatusFound)
		})
		mux.HandleFunc("/real/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"issuer":"same"}`))
		})
		provider := httptest.NewServer(mux)
		defer provider.Close()

		result, issuer := checkOpenIDProvider("test", map[string]interface{}{"authurl": provider.URL})

		assert.True(t, result.Passed, result.Error)
		assert.Equal(t, "same", issuer)
	})
}
