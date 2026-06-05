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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaLinkSharing ports the v1 link-sharing management coverage to
// /api/v2: create (with the full project-permission matrix), list, read and
// delete. There is no update operation. It re-proves the permission matrix
// independently because the v1 routes and their tests will be removed.
//
// Managing shares requires write/admin access to the parent project:
//   - creating an admin share needs project admin; a read/write share needs
//     write access,
//   - listing shares needs project admin,
//   - deleting a share needs write access.
//
// testuser1 owns projects 1/2/3 (admin) and is a member of projects 9 (read),
// 10 (write) and 11 (admin); project 20 is not shared with them at all.
func TestHumaLinkSharing(t *testing.T) {
	// ServiceEnableLinkSharing defaults to true, but the routes only register
	// when it is on — make the precondition explicit for this suite.
	config.ServiceEnableLinkSharing.Set(true)

	onProject := func(projectID string) *webHandlerTestV2 {
		return &webHandlerTestV2{
			user:     &testuser1,
			basePath: "/api/v2/projects/" + projectID + "/shares",
			idParam:  "share",
			t:        t,
		}
	}
	// One shared Echo instance (and its single fixture load) across the suite.
	base := onProject("1")
	require.NoError(t, base.ensureEnv())
	onProjectAs := func(projectID string) *webHandlerTestV2 {
		h := onProject(projectID)
		h.e = base.e
		return h
	}

	t.Run("Create", func(t *testing.T) {
		// Forbidden: project 20 is not shared with testuser1 at all.
		t.Run("Forbidden", func(t *testing.T) {
			for _, perm := range []string{"0", "1", "2"} {
				_, err := onProjectAs("20").testCreateWithUser(nil, nil, `{"permission":`+perm+`}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			}
		})
		// Read-only access (project 9): every share kind is forbidden.
		t.Run("Read only access", func(t *testing.T) {
			for _, perm := range []string{"0", "1", "2"} {
				_, err := onProjectAs("9").testCreateWithUser(nil, nil, `{"permission":`+perm+`}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			}
		})
		// Write access (project 10): read & write shares allowed, admin forbidden.
		t.Run("Write access", func(t *testing.T) {
			t.Run("read only", func(t *testing.T) {
				rec, err := onProjectAs("10").testCreateWithUser(nil, nil, `{"permission":0}`)
				require.NoError(t, err)
				assert.Equal(t, http.StatusCreated, rec.Code)
				assert.Contains(t, rec.Body.String(), `"hash":`)
			})
			t.Run("write", func(t *testing.T) {
				rec, err := onProjectAs("10").testCreateWithUser(nil, nil, `{"permission":1}`)
				require.NoError(t, err)
				assert.Equal(t, http.StatusCreated, rec.Code)
				assert.Contains(t, rec.Body.String(), `"hash":`)
			})
			t.Run("admin", func(t *testing.T) {
				_, err := onProjectAs("10").testCreateWithUser(nil, nil, `{"permission":2}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
		})
		// Admin access (project 11): every share kind allowed.
		t.Run("Admin access", func(t *testing.T) {
			for _, perm := range []string{"0", "1", "2"} {
				rec, err := onProjectAs("11").testCreateWithUser(nil, nil, `{"permission":`+perm+`}`)
				require.NoError(t, err)
				assert.Equal(t, http.StatusCreated, rec.Code)
				assert.Contains(t, rec.Body.String(), `"hash":`)
			}
		})
		t.Run("Password is write-only", func(t *testing.T) {
			rec, err := onProjectAs("11").testCreateWithUser(nil, nil, `{"permission":0,"password":"hunter2"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			// The plaintext password must never be echoed back, and the share
			// type must flip to with-password (2).
			assert.NotContains(t, rec.Body.String(), `hunter2`)
			assert.Contains(t, rec.Body.String(), `"sharing_type":2`)
		})
		t.Run("Nonexisting project", func(t *testing.T) {
			_, err := onProjectAs("9999999").testCreateWithUser(nil, nil, `{"permission":0}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeProjectDoesNotExist)
		})
	})

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			// Project 1 is owned by testuser1 (admin) and has shares 1 and 4.
			rec, err := onProjectAs("1").testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"hash":"test"`)
			assert.Contains(t, rec.Body.String(), `"hash":"testWithPassword"`)
			// Passwords must never leak through the list.
			assert.NotContains(t, rec.Body.String(), `$2a$`)
		})
		t.Run("Search", func(t *testing.T) {
			rec, err := onProjectAs("1").testReadAllWithUser(url.Values{"q": []string{"WITHPASS"}}, nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"hash":"testWithPassword"`)
			assert.NotContains(t, rec.Body.String(), `"hash":"test"`)
		})
		t.Run("Forbidden read-only", func(t *testing.T) {
			// project 9: testuser1 only has read access, not admin.
			_, err := onProjectAs("9").testReadAllWithUser(nil, nil)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden write", func(t *testing.T) {
			// project 10: testuser1 has write access but not admin.
			_, err := onProjectAs("10").testReadAllWithUser(nil, nil)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("ReadOne", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			// share 1 belongs to project 1, owned by testuser1. CanRead resolves
			// the parent project from the path's {project}, so the by-id read
			// succeeds and surfaces the caller's max_permission.
			rec, err := onProjectAs("1").testReadOneWithUser(nil, map[string]string{"share": "1"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), `"hash":"test"`)
			assert.Contains(t, rec.Body.String(), `"max_permission":2`)
			assert.NotEmpty(t, rec.Result().Header.Get("ETag"))
		})
		t.Run("Password is never serialized", func(t *testing.T) {
			// share 4 is a password-protected share on project 1; the bcrypt hash
			// must never appear in the response (password is write-only).
			rec, err := onProjectAs("1").testReadOneWithUser(nil, map[string]string{"share": "4"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), `"sharing_type":2`)
			assert.NotContains(t, rec.Body.String(), `$2a$`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := onProjectAs("1").testReadOneWithUser(nil, map[string]string{"share": "9999999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeProjectShareDoesNotExist)
		})
		t.Run("Share from another project (no IDOR)", func(t *testing.T) {
			// share 2 belongs to project 2. Reading it under project 1 — which
			// testuser1 can read — must 404: ReadOne scopes by id AND project_id,
			// so the share from the other project is never leaked even though the
			// caller has access to the project in the path.
			_, err := onProjectAs("1").testReadOneWithUser(nil, map[string]string{"share": "2"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeProjectShareDoesNotExist)
		})
		t.Run("Forbidden non-member", func(t *testing.T) {
			// user2 is not a member of project 1, so reading its share 1 is denied.
			h := onProjectAs("1")
			h.user = &testuser2
			_, err := h.testReadOneWithUser(nil, map[string]string{"share": "1"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Nonexisting is idempotent", func(t *testing.T) {
			// Deletion is gated on project write access, not on the share
			// existing: deleting a missing share by an authorized user is a
			// no-op that still returns 204 (same as v1).
			rec, err := onProjectAs("1").testDeleteWithUser(nil, map[string]string{"share": "9999999"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
		})
		t.Run("Forbidden read-only", func(t *testing.T) {
			// share 1 is on project 1; user 2 is not even a member.
			h := onProjectAs("1")
			h.user = &testuser2
			_, err := h.testDeleteWithUser(nil, map[string]string{"share": "1"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Normal", func(t *testing.T) {
			// share 1 is on project 1, owned by testuser1. Run last: it removes a
			// fixture row used by the ReadAll cases above.
			rec, err := onProjectAs("1").testDeleteWithUser(nil, map[string]string{"share": "1"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
		})
	})
}
