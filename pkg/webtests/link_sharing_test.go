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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkSharing(t *testing.T) {
	th := NewTestHelper(t)
	th.Login(t, &testuser1)

	t.Run("New Link Share", func(t *testing.T) {
		t.Run("Forbidden", func(t *testing.T) {
			t.Run("read only", func(t *testing.T) {
				_, err := th.Request(t, "PUT", "/api/v1/projects/20/shares", strings.NewReader(`{"permission":0}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("write", func(t *testing.T) {
				_, err := th.Request(t, "PUT", "/api/v1/projects/20/shares", strings.NewReader(`{"permission":1}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("admin", func(t *testing.T) {
				_, err := th.Request(t, "PUT", "/api/v1/projects/20/shares", strings.NewReader(`{"permission":2}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
		})
		t.Run("Read only access", func(t *testing.T) {
			t.Run("read only", func(t *testing.T) {
				_, err := th.Request(t, "PUT", "/api/v1/projects/9/shares", strings.NewReader(`{"permission":0}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("write", func(t *testing.T) {
				_, err := th.Request(t, "PUT", "/api/v1/projects/9/shares", strings.NewReader(`{"permission":1}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("admin", func(t *testing.T) {
				_, err := th.Request(t, "PUT", "/api/v1/projects/9/shares", strings.NewReader(`{"permission":2}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
		})
		t.Run("Write access", func(t *testing.T) {
			t.Run("read only", func(t *testing.T) {
				req, err := th.Request(t, "PUT", "/api/v1/projects/10/shares", strings.NewReader(`{"permission":0}`))
				require.NoError(t, err)
				assert.Contains(t, req.Body.String(), `"hash":`)
			})
			t.Run("write", func(t *testing.T) {
				req, err := th.Request(t, "PUT", "/api/v1/projects/10/shares", strings.NewReader(`{"permission":1}`))
				require.NoError(t, err)
				assert.Contains(t, req.Body.String(), `"hash":`)
			})
			t.Run("admin", func(t *testing.T) {
				_, err := th.Request(t, "PUT", "/api/v1/projects/10/shares", strings.NewReader(`{"permission":2}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
		})
		t.Run("Admin access", func(t *testing.T) {
			t.Run("read only", func(t *testing.T) {
				req, err := th.Request(t, "PUT", "/api/v1/projects/11/shares", strings.NewReader(`{"permission":0}`))
				require.NoError(t, err)
				assert.Contains(t, req.Body.String(), `"hash":`)
			})
			t.Run("write", func(t *testing.T) {
				req, err := th.Request(t, "PUT", "/api/v1/projects/11/shares", strings.NewReader(`{"permission":1}`))
				require.NoError(t, err)
				assert.Contains(t, req.Body.String(), `"hash":`)
			})
			t.Run("admin", func(t *testing.T) {
				req, err := th.Request(t, "PUT", "/api/v1/projects/11/shares", strings.NewReader(`{"permission":2}`))
				require.NoError(t, err)
				assert.Contains(t, req.Body.String(), `"hash":`)
			})
		})
	})
}
