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

package caldavtests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPinDELETE(t *testing.T) {
	t.Run("collection", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavDELETE(t, e, "/dav/projects/36")

		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code, "Body:\n%s", rec.Body.String())
		assert.Equal(t, allowCollectionReadWrite, rec.Header().Get("Allow"))
		assert.Empty(t, rec.Body.String())
		assert.Equal(t, http.StatusOK, caldavGET(t, e, pinTaskHref).Code)
	})

	t.Run("pseudo collection", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavDELETE(t, e, "/dav/projects/-1")

		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code, "Body:\n%s", rec.Body.String())
		assert.Equal(t, allowCollectionReadWrite, rec.Header().Get("Allow"))
	})

	t.Run("home set", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavDELETE(t, e, "/dav/projects")

		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code, "Body:\n%s", rec.Body.String())
		assert.Equal(t, allowHomeSet, rec.Header().Get("Allow"))
	})

	t.Run("stale If-Match", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavRequest(t, e, http.MethodDelete, pinTaskHref, "", map[string]string{"If-Match": `"stale"`})

		assert.Equal(t, http.StatusPreconditionFailed, rec.Code, "Body:\n%s", rec.Body.String())
		assert.Equal(t, http.StatusOK, caldavGET(t, e, pinTaskHref).Code)
	})

	for name, ifMatch := range map[string]func(etag string) string{
		"matching If-Match": func(etag string) string { return etag },
		"If-Match star":     func(string) string { return "*" },
	} {
		t.Run(name, func(t *testing.T) {
			e := setupTestEnv(t)
			etag := caldavGET(t, e, pinTaskHref).Header().Get("ETag")
			require.NotEmpty(t, etag)

			rec := caldavRequest(t, e, http.MethodDelete, pinTaskHref, "", map[string]string{"If-Match": ifMatch(etag)})

			assert.Equal(t, http.StatusNoContent, rec.Code, "Body:\n%s", rec.Body.String())
			assert.Equal(t, http.StatusNotFound, caldavGET(t, e, pinTaskHref).Code)
		})
	}
}
