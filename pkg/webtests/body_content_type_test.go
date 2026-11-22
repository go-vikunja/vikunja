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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXMLBodiesRejected(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	t.Run("project create", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPut, "/api/v1/projects",
			`<Project><Title>xml-project</Title><BackgroundFileID>1</BackgroundFileID></Project>`, token, "application/xml")
		assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"code":2004`)
		db.AssertMissing(t, "projects", map[string]interface{}{"title": "xml-project"})
	})

	t.Run("bot create", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPut, "/api/v1/user/bots",
			`<BotUser><Username>bot-xml</Username><IsAdmin>true</IsAdmin></BotUser>`, token, "application/xml")
		assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"code":2004`)
		db.AssertMissing(t, "users", map[string]interface{}{"username": "bot-xml"})
	})

	t.Run("json still works", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodPut, "/api/v1/projects", `{"title":"json-project"}`, token, "application/json")
		assert.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	})

	t.Run("bodyless request still works", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v1/projects/1", "", token, "application/xml")
		assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	})
}
