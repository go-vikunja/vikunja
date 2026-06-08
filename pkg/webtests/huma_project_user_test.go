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
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProjectUser ports the model-level matrix in pkg/models/project_users_test.go
// to the v2 HTTP surface. Project<->user shares live under
// /projects/{project}/users/{user}; {user} is the username (a string), and
// there is no read-one. basePath carries the literal {project}, idParam picks
// {user}.
//
// The whole test shares one Echo instance, so fixtures load once and mutations
// persist across subtests — each mutating case therefore targets a distinct
// (project, user) pair so order cannot make them interfere.
//
// Permission gradient — ProjectUser.Can* delegate to Project.IsAdmin (create/
// update/delete), and ReadAll checks Project.CanRead. The shares in
// pkg/db/fixtures/users_projects.yml give testuser1 every rung against projects
// owned by user6:
//   - read share  (project 9):  CAN list, CANNOT create/update/delete
//   - write share (project 10): CAN list, CANNOT create/update/delete
//   - admin share (project 11): CAN do everything
//
// Project 3 (owned by user3) is shared read-only to testuser1 and user2 — used
// for the list/cardinality and read-share-cannot-write assertions. Project 1 is
// owned by testuser1 (admin via ownership): the create/update/delete happy path.
func TestProjectUser(t *testing.T) {
	owned := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/1/users",
		idParam:  "user",
		t:        t,
	}
	require.NoError(t, owned.ensureEnv())
	// Share owned's Echo across harnesses: setupTestEnv() regenerates the JWT
	// secret, so independent harnesses would invalidate each other's tokens.
	readProject := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/3/users",
		idParam:  "user",
		t:        t,
		e:        owned.e,
	}
	forbidden := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/2/users",
		idParam:  "user",
		t:        t,
		e:        owned.e,
	}
	readShared := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/9/users",
		idParam:  "user",
		t:        t,
		e:        owned.e,
	}
	writeShared := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/10/users",
		idParam:  "user",
		t:        t,
		e:        owned.e,
	}
	adminShared := webHandlerTestV2{
		user:     &testuser1,
		basePath: "/api/v2/projects/11/users",
		idParam:  "user",
		t:        t,
		e:        owned.e,
	}

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal - exact shared set for project 3", func(t *testing.T) {
			// project 3 is shared to user1 and user2 (both read). The list must
			// surface exactly those two users with their permission and nothing else.
			rec, err := readProject.testReadAllWithUser(nil, nil)
			require.NoError(t, err)

			var env struct {
				Items []struct {
					ID         int64  `json:"id"`
					Username   string `json:"username"`
					Email      string `json:"email"`
					Permission int    `json:"permission"`
				} `json:"items"`
				Total int64 `json:"total"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
			assert.Len(t, env.Items, 2)
			assert.Equal(t, int64(2), env.Total)
			usernames := make([]string, 0, len(env.Items))
			for _, u := range env.Items {
				usernames = append(usernames, u.Username)
				assert.Empty(t, u.Email, "user emails must be obfuscated in the share list")
				assert.Equal(t, 0, u.Permission, "both shares are read-only (0)")
			}
			assert.ElementsMatch(t, []string{"user1", "user2"}, usernames)
		})
		t.Run("Search", func(t *testing.T) {
			rec, err := readProject.testReadAllWithUser(map[string][]string{"q": {"USER2"}}, nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"username":"user2"`)
			assert.NotContains(t, rec.Body.String(), `"username":"user1"`)
		})
		t.Run("Read share can list", func(t *testing.T) {
			// CanRead delegates to Project.CanRead; a read share is enough to list.
			_, err := readShared.testReadAllWithUser(nil, nil)
			require.NoError(t, err)
		})
		t.Run("Forbidden - no access to the project", func(t *testing.T) {
			_, err := forbidden.testReadAllWithUser(nil, nil)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := owned.testCreateWithUser(nil, nil, `{"username":"user2","permission":0}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"username":"user2"`)
		})
		t.Run("Admin share can create", func(t *testing.T) {
			// project 11 admin share clears Project.IsAdmin → CanCreate passes.
			rec, err := adminShared.testCreateWithUser(nil, nil, `{"username":"user2","permission":1}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"username":"user2"`)
		})
		t.Run("Read share cannot create", func(t *testing.T) {
			_, err := readShared.testCreateWithUser(nil, nil, `{"username":"user3","permission":0}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Write share cannot create", func(t *testing.T) {
			_, err := writeShared.testCreateWithUser(nil, nil, `{"username":"user3","permission":0}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden - no access to the project", func(t *testing.T) {
			_, err := forbidden.testCreateWithUser(nil, nil, `{"username":"user3","permission":0}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Duplicate share", func(t *testing.T) {
			// Adding user3 to project 1 twice surfaces ErrUserAlreadyHasAccess (409).
			rec, err := owned.testCreateWithUser(nil, nil, `{"username":"user3","permission":0}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			_, err = owned.testCreateWithUser(nil, nil, `{"username":"user3","permission":0}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusConflict, getHTTPErrorCode(err))
		})
		t.Run("Share with the project owner", func(t *testing.T) {
			// testuser1 owns project 1; adding the owner returns ErrUserAlreadyHasAccess (409).
			_, err := owned.testCreateWithUser(nil, nil, `{"username":"user1","permission":0}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusConflict, getHTTPErrorCode(err))
		})
		t.Run("Nonexisting project", func(t *testing.T) {
			missing := webHandlerTestV2{
				user:     &testuser1,
				basePath: "/api/v2/projects/2000/users",
				idParam:  "user",
				t:        t,
				e:        owned.e,
			}
			_, err := missing.testCreateWithUser(nil, nil, `{"username":"user2","permission":0}`)
			require.Error(t, err)
			// CanCreate → Project.IsAdmin surfaces ErrProjectDoesNotExist (404), not a bare forbidden.
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Nonexisting user", func(t *testing.T) {
			_, err := owned.testCreateWithUser(nil, nil, `{"username":"user500","permission":0}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
		t.Run("Invalid permission", func(t *testing.T) {
			// permission=500 is above the schema maximum (2) → Huma rejects with 422
			// before the model's isValid runs (v1 returned 400 from the model).
			_, err := owned.testCreateWithUser(nil, nil, `{"username":"user4","permission":500}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			// Update needs an existing share, so create user5 first.
			_, err := owned.testCreateWithUser(nil, nil, `{"username":"user5","permission":0}`)
			require.NoError(t, err)
			rec, err := owned.testUpdateWithUser(nil, map[string]string{"user": "user5"}, `{"permission":2}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"permission":2`)
		})
		t.Run("Admin share can update", func(t *testing.T) {
			_, err := adminShared.testCreateWithUser(nil, nil, `{"username":"user3","permission":0}`)
			require.NoError(t, err)
			rec, err := adminShared.testUpdateWithUser(nil, map[string]string{"user": "user3"}, `{"permission":1}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"permission":1`)
		})
		t.Run("Read share cannot update", func(t *testing.T) {
			// project 3 is shared read-only to testuser1; user2 already has a share
			// there. Updating needs admin (Can* fails before the user is touched).
			_, err := readProject.testUpdateWithUser(nil, map[string]string{"user": "user2"}, `{"permission":2}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Write share cannot update", func(t *testing.T) {
			_, err := writeShared.testUpdateWithUser(nil, map[string]string{"user": "user2"}, `{"permission":2}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden - no access to the project", func(t *testing.T) {
			_, err := forbidden.testUpdateWithUser(nil, map[string]string{"user": "user2"}, `{"permission":2}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Invalid permission", func(t *testing.T) {
			_, err := owned.testUpdateWithUser(nil, map[string]string{"user": "user5"}, `{"permission":500}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			// Delete needs an existing share, so create user6 first.
			_, err := owned.testCreateWithUser(nil, nil, `{"username":"user6","permission":0}`)
			require.NoError(t, err)
			rec, err := owned.testDeleteWithUser(nil, map[string]string{"user": "user6"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, rec.Body.String())
		})
		t.Run("Admin share can delete", func(t *testing.T) {
			_, err := adminShared.testCreateWithUser(nil, nil, `{"username":"user4","permission":0}`)
			require.NoError(t, err)
			rec, err := adminShared.testDeleteWithUser(nil, map[string]string{"user": "user4"})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, rec.Code)
		})
		t.Run("Read share cannot delete", func(t *testing.T) {
			// project 3 shares user2 read-only; testuser1 (read share) lacks admin.
			_, err := readProject.testDeleteWithUser(nil, map[string]string{"user": "user2"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Write share cannot delete", func(t *testing.T) {
			_, err := writeShared.testDeleteWithUser(nil, map[string]string{"user": "user2"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Forbidden - no access to the project", func(t *testing.T) {
			_, err := forbidden.testDeleteWithUser(nil, map[string]string{"user": "user2"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("User without a share on the project", func(t *testing.T) {
			// user7 has no share on project 1; deleting their (nonexistent) share
			// returns ErrUserDoesNotHaveAccessToProject (403).
			_, err := owned.testDeleteWithUser(nil, map[string]string{"user": "user7"})
			require.Error(t, err)
			assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
		})
		t.Run("Nonexisting user", func(t *testing.T) {
			_, err := owned.testDeleteWithUser(nil, map[string]string{"user": "user1000"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
		})
	})
}
