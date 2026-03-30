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
	"fmt"
	"strconv"
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContentLengthOnListEndpoints verifies that API list endpoints set the
// Content-Length response header. This is a server-side mitigation for a known
// macOS curl bug where piping curl output to another program (e.g. curl | jq)
// can produce empty stdin when the response uses chunked transfer encoding
// without Content-Length. By ensuring Content-Length is always present, curl
// can reliably determine when the response body is complete before passing
// data through the pipe.
func TestContentLengthOnListEndpoints(t *testing.T) {
	testHandler := webHandlerTest{
		user: &testuser1,
		strFunc: func() handler.CObject {
			return &models.Project{}
		},
		t: t,
	}

	t.Run("ReadAll response has Content-Length", func(t *testing.T) {
		rec, err := testHandler.testReadAllWithUser(nil, nil)
		require.NoError(t, err)

		cl := rec.Header().Get("Content-Length")
		assert.NotEmpty(t, cl, "Content-Length header must be set on list responses to prevent macOS curl piping issues")

		clInt, err := strconv.Atoi(cl)
		require.NoError(t, err)
		assert.Equal(t, rec.Body.Len(), clInt,
			fmt.Sprintf("Content-Length (%d) must match actual body size (%d)", clInt, rec.Body.Len()))
	})
}
