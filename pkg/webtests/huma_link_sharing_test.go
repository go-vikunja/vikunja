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
	"strconv"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func insertTestShare(t *testing.T, id, projectID int64) {
	t.Helper()
	insertTestShareWithPermission(t, id, projectID, models.PermissionRead)
}

func insertTestShareWithPermission(t *testing.T, id, projectID int64, perm models.Permission) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	_, err := s.Insert(&models.LinkSharing{
		ID:          id,
		Hash:        "ghsa-qfwc-" + strconv.FormatInt(id, 10),
		ProjectID:   projectID,
		Permission:  perm,
		SharingType: models.SharingTypeWithoutPassword,
		SharedByID:  6,
	})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

func TestHumaLinkSharing(t *testing.T) {
	config.ServiceEnableLinkSharing.Set(true)

	onProject := func(projectID string) *webHandlerTestV2 {
		return &webHandlerTestV2{
			user:     &testuser1,
			basePath: "/api/v2/projects/" + projectID + "/shares",
			idParam:  "share",
			t:        t,
		}
	}
	base := onProject("1")
	require.NoError(t, base.ensureEnv())
	onProjectAs := func(projectID string) *webHandlerTestV2 {
		h := onProject(projectID)
		h.e = base.e
		return h
	}

	t.Run("Create", func(t *testing.T) {
		t.Run("Forbidden", func(t *testing.T) {
			for _, perm := range []string{"0", "1", "2"} {
				_, err := onProjectAs("20").testCreateWithUser(nil, nil, `{"permission":`+perm+`}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			}
		})
		t.Run("Read only access", func(t *testing.T) {
			for _, perm := range []string{"0", "1", "2"} {
				_, err := onProjectAs("9").testCreateWithUser(nil, nil, `{"permission":`+perm+`}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			}
		})
		t.Run("Write access", func(t *testing.T) {
			// GHSA-fmmf-xq98-g327: managing link shares requires project admin.
			for _, perm := range []string{"0", "1", "2"} {
				_, err := onProjectAs("10").testCreateWithUser(nil, nil, `{"permission":`+perm+`}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			}
		})
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
			rec, err := onProjectAs("1").testReadAllWithUser(nil, nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"hash":"test"`)
			assert.Contains(t, rec.Body.String(), `"hash":"testWithPassword"`)
			assert.NotContains(t, rec.Body.String(), `$2a$`)
		})
		t.Run("Search", func(t *testing.T) {
			rec, err := onProjectAs("1").testReadAllWithUser(url.Values{"q": []string{"WITHPASS"}}, nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"hash":"testWithPassword"`)
			assert.NotContains(t, rec.Body.String(), `"hash":"test"`)
		})
		t.Run("Forbidden read-only", func(t *testing.T) {
			_, err := onProjectAs("9").testReadAllWithUser(nil, nil)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden write", func(t *testing.T) {
			_, err := onProjectAs("10").testReadAllWithUser(nil, nil)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("ReadOne", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := onProjectAs("1").testReadOneWithUser(nil, map[string]string{"share": "1"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(t, rec.Body.String(), `"hash":"test"`)
			assert.Contains(t, rec.Body.String(), `"max_permission":2`)
			assert.NotEmpty(t, rec.Result().Header.Get("ETag"))
		})
		t.Run("Password is never serialized", func(t *testing.T) {
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
			_, err := onProjectAs("1").testReadOneWithUser(nil, map[string]string{"share": "2"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeProjectShareDoesNotExist)
		})
		t.Run("Forbidden non-member", func(t *testing.T) {
			h := onProjectAs("1")
			h.user = &testuser2
			_, err := h.testReadOneWithUser(nil, map[string]string{"share": "1"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden read-only member", func(t *testing.T) {
			// A by-ID read discloses the access-bearing hash (GHSA-qfwc-vx6f-3g6g).
			insertTestShare(t, 5, 9)
			_, err := onProjectAs("9").testReadOneWithUser(nil, map[string]string{"share": "5"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden write member", func(t *testing.T) {
			insertTestShare(t, 6, 10)
			_, err := onProjectAs("10").testReadOneWithUser(nil, map[string]string{"share": "6"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := onProjectAs("1").testDeleteWithUser(nil, map[string]string{"share": "9999999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeProjectShareDoesNotExist)
		})
		t.Run("Share from another project", func(t *testing.T) {
			_, err := onProjectAs("1").testDeleteWithUser(nil, map[string]string{"share": "2"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
			db.AssertExists(t, "link_shares", map[string]interface{}{"id": 2}, false)
		})
		// GHSA-fmmf-xq98-g327: the delete body carries no tier.
		t.Run("Forbidden write member, every tier", func(t *testing.T) {
			for i, perm := range []models.Permission{models.PermissionRead, models.PermissionWrite, models.PermissionAdmin} {
				id := int64(40 + i)
				insertTestShareWithPermission(t, id, 10, perm)
				_, err := onProjectAs("10").testDeleteWithUser(nil, map[string]string{"share": strconv.FormatInt(id, 10)})
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
				db.AssertExists(t, "link_shares", map[string]interface{}{"id": id}, false)
			}
		})
		t.Run("Forbidden read member", func(t *testing.T) {
			insertTestShare(t, 43, 9)
			_, err := onProjectAs("9").testDeleteWithUser(nil, map[string]string{"share": "43"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden link share principal", func(t *testing.T) {
			jwt, err := auth.NewLinkShareJWTAuthtoken(&models.LinkSharing{
				ID:          3,
				Hash:        "test3",
				ProjectID:   3,
				Permission:  models.PermissionAdmin,
				SharingType: models.SharingTypeWithoutPassword,
				SharedByID:  1,
			})
			require.NoError(t, err)
			rec := apiTokenReq(base.e, http.MethodDelete, "/api/v2/projects/3/shares/3", jwt, "")
			assert.Equal(t, http.StatusForbidden, rec.Code)
			db.AssertExists(t, "link_shares", map[string]interface{}{"id": 3}, false)
		})
		t.Run("Admin deletes every tier", func(t *testing.T) {
			for i, perm := range []models.Permission{models.PermissionRead, models.PermissionWrite, models.PermissionAdmin} {
				id := int64(50 + i)
				insertTestShareWithPermission(t, id, 11, perm)
				rec, err := onProjectAs("11").testDeleteWithUser(nil, map[string]string{"share": strconv.FormatInt(id, 10)})
				require.NoError(t, err)
				assert.Equal(t, http.StatusNoContent, rec.Code)
				db.AssertMissing(t, "link_shares", map[string]interface{}{"id": id})
			}
		})
		t.Run("Forbidden read-only", func(t *testing.T) {
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
