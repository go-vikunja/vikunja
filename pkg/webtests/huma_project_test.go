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
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaProject mirrors v1's TestProject shape so v2 contract parity is
// readable side-by-side. Status-code differences from v1 are noted inline.
func TestHumaProject(t *testing.T) {
	testHandler := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects",
		idParam:  "project",
		t:        t,
	}

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Test1`)
			assert.NotContains(t, rec.Body.String(), `Test2"`)
			assert.Contains(t, rec.Body.String(), `Test3`)  // Shared directly via users_project
			assert.Contains(t, rec.Body.String(), `Test12`) // Shared via parent project
			assert.NotContains(t, rec.Body.String(), `Test5`)
			assert.NotContains(t, rec.Body.String(), `Test21`) // Archived through parent project
			assert.NotContains(t, rec.Body.String(), `Test22`) // Archived directly
		})
		t.Run("Normal with archived projects", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(url.Values{"is_archived": []string{"true"}}, nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Test1`)
			assert.Contains(t, rec.Body.String(), `Test21`) // Archived through project
			assert.Contains(t, rec.Body.String(), `Test22`) // Archived directly
		})
		t.Run("Expand permissions", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(url.Values{"expand": []string{"permissions"}}, nil)
			require.NoError(t, err)
			// User 1 owns Test1 → admin (2). With expand the field carries a real value.
			assert.Contains(t, rec.Body.String(), `"max_permission":2`)
		})
	})

	t.Run("ReadOne", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "1"})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Test1"`)
			assert.NotContains(t, rec.Body.String(), `"title":"Test2"`)
			assert.Contains(t, rec.Body.String(), `"username":"user1"`)
			assert.NotEmpty(t, rec.Result().Header.Get("ETag"))
		})
		t.Run("Expand permissions", func(t *testing.T) {
			// User 1 owns Test1 → admin (2); expand surfaces it as max_permission.
			rec, err := testHandler.testReadOneWithUser(url.Values{"expand": []string{"permissions"}}, map[string]string{"project": "1"})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"max_permission":2`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			// Projects return 404 here (CanRead → GetProjectSimpleByID → ErrProjectDoesNotExist),
			// unlike labels which return 403 from the read branch.
			_, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Project 20 exists but is owned by user13: CanRead returns false → 403.
				_, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "20"})
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "6"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test6"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "11"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test11"`)
			})
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem","description":"Ipsum"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
			assert.Contains(t, rec.Body.String(), `"description":"Ipsum"`)
		})
		t.Run("Empty title", func(t *testing.T) {
			// v2 returns 422, not v1's 400; full body shape asserted in TestHuma_ErrorShapeIsRFC9457.
			_, err := testHandler.testCreateWithUser(nil, nil, `{"title":""}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "1"}, `{"title":"TestLoremIpsum"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			// The description should not be wiped but returned as it was.
			assert.Contains(t, rec.Body.String(), `"description":"Lorem Ipsum`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "9999"}, `{"title":"TestLoremIpsum"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Empty title", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "1"}, `{"title":""}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13.
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "20"}, `{"title":"TestLoremIpsum"}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
			t.Run("Shared Via Team readonly forbidden", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "6"}, `{"title":"TestLoremIpsum"}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
			t.Run("Shared Via Team write allowed", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "7"}, `{"title":"TestLoremIpsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "1"})
			require.NoError(t, err)
			// v2 delete is 204 No Content; v1 returned 200 + a message body.
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13.
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "20"})
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
			t.Run("Shared Via Team write forbidden", func(t *testing.T) {
				// Write access is not enough to delete; needs admin.
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "7"})
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
			t.Run("Shared Via Team admin allowed", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "8"})
				require.NoError(t, err)
				assert.Equal(t, http.StatusNoContent, rec.Code)
			})
		})
	})
}
