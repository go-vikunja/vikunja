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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Postgres rejects NUL in text columns (SQLSTATE 22021), so it must be refused at
// the JSON boundary on every DB. API-CLOUD-4J.
func TestJSONBodyWithNULRejected(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	cases := []struct{ method, path string }{
		{http.MethodPut, "/api/v1/tasks/1/comments"},
		{http.MethodPost, "/api/v2/tasks/1/comments"},
	}
	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			rec := humaRequest(t, e, c.method, c.path, `{"comment":"Lorem\u0000Ipsum"}`, token, "")
			assert.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
			assert.Contains(t, rec.Body.String(), "NUL")
		})
		t.Run(c.path+" escaped backslash is plain text", func(t *testing.T) {
			rec := humaRequest(t, e, c.method, c.path, `{"comment":"Lorem\\u0000Ipsum"}`, token, "")
			assert.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
		})
	}
}
