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

package feeds

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationsAtomFeed(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	t.Run("returns valid atom XML for authenticated user with no notifications", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/feeds/notifications.atom", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set("userBasicAuth", &user.User{ID: 1, Name: "User 1", Username: "user1", Language: "en"})

		err := NotificationsAtomFeed(c)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, strings.HasPrefix(rec.Header().Get(echo.HeaderContentType), "application/atom+xml"),
			"unexpected content type: %s", rec.Header().Get(echo.HeaderContentType))

		// Must be parseable as XML.
		var doc struct {
			XMLName xml.Name `xml:"feed"`
			Title   string   `xml:"title"`
		}
		require.NoError(t, xml.Unmarshal(rec.Body.Bytes(), &doc))
		assert.Contains(t, doc.Title, "User 1", "feed title should include the user's name")
	})

	t.Run("returns 500 when context has no authenticated user", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/feeds/notifications.atom", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := NotificationsAtomFeed(c)
		require.Error(t, err)
	})
}
