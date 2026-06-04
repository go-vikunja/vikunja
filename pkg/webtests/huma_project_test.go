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
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runelength250Title is the >250-rune title used by both Create and Update to
// trip the title length limit. v1 asserted the govalidator runelength(1|250)
// message; v2 enforces the same bound at the schema layer (maxLength:250),
// rejecting with 422 before the handler runs.
const runelength250Title = `Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea taki`

// TestHumaProject is a full 1:1 port of v1's TestProject (pkg/webtests/project_test.go).
// It proves the v2 routes independently enforce the complete permission/sharing
// matrix v1 covered (owner; team / user / parent-project shares × read/write/admin;
// member-but-not-admin; non-member), plus v2's own HTTP-layer contract (status
// codes, absent ETag, expand=permissions). Status-code differences from v1 are
// noted inline. All cases drive testuser1, whose share fixtures (projects 6–17,
// 20, 32–34, 9–11) exercise every share-kind×level just like the v1 test.
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
		t.Run("Search", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(url.Values{"q": []string{"Test1"}}, nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Test1`)

			// v2 wraps the list in a Paginated envelope; the items live under
			// "items". Unmarshal that to assert exact cardinality the way v1
			// asserted on the bare slice.
			var paginated struct {
				Items []models.Project `json:"items"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &paginated))

			if db.ParadeDBAvailable() {
				// ParadeDB fuzzy(1, prefix=true) on "Test1" matches Test2-Test9
				// (edit distance 1), Test10+ (prefix), etc. The recursive CTE
				// also pulls in child projects of matched parents.
				// +1 for the reparent-escalation fixture child (project 43).
				require.Len(t, paginated.Items, 27)
			} else {
				// ILIKE '%Test1%' matches Test1, Test10, Test11, Test19, + favorites.
				// The recursive CTE also pulls in project 43 as a child of the
				// matched project 10 (reparent-escalation fixture).
				require.Len(t, paginated.Items, 6)
				assert.NotContains(t, rec.Body.String(), `Test2"`)
				assert.NotContains(t, rec.Body.String(), `Test3`)
				assert.NotContains(t, rec.Body.String(), `Test4`)
				assert.NotContains(t, rec.Body.String(), `Test5`)
			}
		})
		t.Run("Normal with archived projects", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(url.Values{"is_archived": []string{"true"}}, nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Test1`)
			assert.NotContains(t, rec.Body.String(), `Test2"`)
			assert.Contains(t, rec.Body.String(), `Test3`)  // Shared directly via users_project
			assert.Contains(t, rec.Body.String(), `Test12`) // Shared via parent project
			assert.NotContains(t, rec.Body.String(), `Test5`)
			assert.Contains(t, rec.Body.String(), `Test21`) // Archived through project
			assert.Contains(t, rec.Body.String(), `Test22`) // Archived directly

			// Verify is_archived is propagated to child projects of archived parents.
			var paginated struct {
				Items []models.Project `json:"items"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &paginated))
			found21 := false
			for _, p := range paginated.Items {
				if p.ID == 21 {
					found21 = true
					assert.True(t, p.IsArchived, "Project 21 should have is_archived=true because its parent is archived")
				}
			}
			assert.True(t, found21, "Project 21 should be present when listing archived projects")
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
			// Owner is echoed as the full user object, like v1.
			assert.Contains(t, rec.Body.String(), `"owner":{"id":1,"name":"","username":"user1",`)
			assert.NotContains(t, rec.Body.String(), `"owner":{"id":2,"name":"","username":"user2",`)
			// Tasks are never embedded on a plain project read.
			assert.NotContains(t, rec.Body.String(), `"tasks":`)
			// max_permission is always present on a read and reflects the caller's
			// permission. User1 owns Test1 → admin (2).
			assert.Contains(t, rec.Body.String(), `"max_permission":2`)
			// The project read is served fresh on every call; no ETag is sent
			// because the response carries derived state that changes without
			// bumping project.Updated.
			assert.Empty(t, rec.Result().Header.Get("ETag"))
		})
		t.Run("Nonexisting", func(t *testing.T) {
			// Projects return 404 here (CanRead → GetProjectSimpleByID → ErrProjectDoesNotExist),
			// unlike labels which return 403 from the read branch.
			_, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "9999"})
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeProjectDoesNotExist)
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Project 20 exists but is owned by user13: CanRead returns false → 403.
				_, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "20"})
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})

			// readOneWithMaxPermission reads a shared project and asserts the
			// granted level via the always-present max_permission field, the v2
			// equivalent of v1's x-max-permission header assertion.
			readOneWithMaxPermission := func(t *testing.T, projectID, title string, want models.Permission) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": projectID})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"`+title+`"`)

				var p models.Project
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &p))
				assert.Equal(t, want, p.MaxPermission)
			}

			t.Run("Shared Via Team readonly", func(t *testing.T) {
				readOneWithMaxPermission(t, "6", "Test6", models.PermissionRead)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				readOneWithMaxPermission(t, "7", "Test7", models.PermissionWrite)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				readOneWithMaxPermission(t, "8", "Test8", models.PermissionAdmin)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				readOneWithMaxPermission(t, "9", "Test9", models.PermissionRead)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				readOneWithMaxPermission(t, "10", "Test10", models.PermissionWrite)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				readOneWithMaxPermission(t, "11", "Test11", models.PermissionAdmin)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				readOneWithMaxPermission(t, "12", "Test12", models.PermissionRead)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				readOneWithMaxPermission(t, "13", "Test13", models.PermissionWrite)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				readOneWithMaxPermission(t, "14", "Test14", models.PermissionAdmin)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				readOneWithMaxPermission(t, "15", "Test15", models.PermissionRead)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				readOneWithMaxPermission(t, "16", "Test16", models.PermissionWrite)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				readOneWithMaxPermission(t, "17", "Test17", models.PermissionAdmin)
			})
		})
	})

	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
			assert.Contains(t, rec.Body.String(), `"description":""`)
			// The creating user becomes the owner.
			assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
			// Tasks are not embedded in the create response.
			assert.NotContains(t, rec.Body.String(), `"tasks":`)
			// Create doesn't compute the caller's permission; null over a
			// misleading 0 (read) for the owner. Computed on a subsequent read.
			assert.Contains(t, rec.Body.String(), `"max_permission":null`)
		})
		t.Run("Normal with description", func(t *testing.T) {
			rec, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem","description":"Ipsum"}`)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, rec.Code)
			assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
			assert.Contains(t, rec.Body.String(), `"description":"Ipsum"`)
			assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
			assert.NotContains(t, rec.Body.String(), `"tasks":`)
		})
		t.Run("Nonexisting parent project", func(t *testing.T) {
			_, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem","parent_project_id":99999}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeProjectDoesNotExist)
		})
		t.Run("Empty title", func(t *testing.T) {
			// v2 returns 422, not v1's 400; full body shape asserted in TestHuma_ErrorShapeIsRFC9457.
			_, err := testHandler.testCreateWithUser(nil, nil, `{"title":""}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
		t.Run("Title too long", func(t *testing.T) {
			// v1 hit govalidator runelength(1|250); v2 enforces maxLength:250 at
			// the schema layer → 422 before the handler.
			_, err := testHandler.testCreateWithUser(nil, nil, `{"title":"`+runelength250Title+`"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
		t.Run("Permissions check", func(t *testing.T) {
			// Parent permission matrix: creating a child under a parent requires
			// write access to that parent.
			t.Run("Forbidden", func(t *testing.T) {
				// Parent 20 is owned by user13; user1 has no access.
				_, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem","parent_project_id":20}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				// Read-only on parent 32 is not enough to create a child.
				_, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem","parent_project_id":32}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem","parent_project_id":33}`)
				require.NoError(t, err)
				assert.Equal(t, http.StatusCreated, rec.Code)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
				assert.NotContains(t, rec.Body.String(), `"tasks":`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem","parent_project_id":34}`)
				require.NoError(t, err)
				assert.Equal(t, http.StatusCreated, rec.Code)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
				assert.NotContains(t, rec.Body.String(), `"tasks":`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				// Read-only on parent 9 is not enough to create a child.
				_, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem","parent_project_id":9}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem","parent_project_id":10}`)
				require.NoError(t, err)
				assert.Equal(t, http.StatusCreated, rec.Code)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
				assert.NotContains(t, rec.Body.String(), `"tasks":`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem","parent_project_id":11}`)
				require.NoError(t, err)
				assert.Equal(t, http.StatusCreated, rec.Code)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
				assert.NotContains(t, rec.Body.String(), `"tasks":`)
			})
		})
	})

	t.Run("Update", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "1"}, `{"title":"TestLoremIpsum"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			// The description should not be wiped but returned as it was.
			assert.Contains(t, rec.Body.String(), `"description":"Lorem Ipsum`)
			// Update doesn't recompute the permission; null, like create.
			assert.Contains(t, rec.Body.String(), `"max_permission":null`)
		})
		t.Run("Normal with updating the description", func(t *testing.T) {
			rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "1"}, `{"title":"TestLoremIpsum","description":"Lorem Ipsum dolor sit amet"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			assert.Contains(t, rec.Body.String(), `"description":"Lorem Ipsum dolor sit amet`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "9999"}, `{"title":"TestLoremIpsum"}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusNotFound, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeProjectDoesNotExist)
		})
		t.Run("Empty title", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "1"}, `{"title":""}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, getHTTPErrorCode(err))
		})
		t.Run("Title too long", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "1"}, `{"title":"`+runelength250Title+`"}`)
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

			t.Run("Shared Via Team readonly", func(t *testing.T) {
				// Read access is not enough to update.
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "6"}, `{"title":"TestLoremIpsum"}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "7"}, `{"title":"TestLoremIpsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "8"}, `{"title":"TestLoremIpsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "9"}, `{"title":"TestLoremIpsum"}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "10"}, `{"title":"TestLoremIpsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "11"}, `{"title":"TestLoremIpsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "12"}, `{"title":"TestLoremIpsum"}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "13"}, `{"title":"TestLoremIpsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "14"}, `{"title":"TestLoremIpsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "15"}, `{"title":"TestLoremIpsum"}`)
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "16"}, `{"title":"TestLoremIpsum"}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "17"}, `{"title":"TestLoremIpsum"}`)
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
			assertHandlerErrorCode(t, err, models.ErrCodeProjectDoesNotExist)
		})
		t.Run("Permissions check", func(t *testing.T) {
			// Delete needs admin everywhere: read and write must be refused, admin allowed.
			deleteForbidden := func(t *testing.T, projectID string) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": projectID})
				require.Error(t, err)
				assert.Equal(t, http.StatusForbidden, getHTTPErrorCode(err))
			}
			deleteAllowed := func(t *testing.T, projectID string) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": projectID})
				require.NoError(t, err)
				assert.Equal(t, http.StatusNoContent, rec.Code)
				assert.Empty(t, rec.Body.String())
			}

			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13.
				deleteForbidden(t, "20")
			})

			t.Run("Shared Via Team readonly", func(t *testing.T) {
				deleteForbidden(t, "6")
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				// Write access is not enough to delete; needs admin.
				deleteForbidden(t, "7")
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				deleteAllowed(t, "8")
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				deleteForbidden(t, "9")
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				deleteForbidden(t, "10")
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				deleteAllowed(t, "11")
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				deleteForbidden(t, "12")
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				deleteForbidden(t, "13")
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				deleteAllowed(t, "14")
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				deleteForbidden(t, "15")
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				deleteForbidden(t, "16")
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				deleteAllowed(t, "17")
			})
		})
	})
}

// TestHumaProject_PATCHMergePatch confirms AutoPatch round-trips: it GETs the
// read body (which carries the read-only max_permission) and re-PUTs it, so the
// update body sharing the read shape must accept that echo without 422.
func TestHumaProject_PATCHMergePatch(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)

	rec := humaRequest(t, e, http.MethodPost, "/api/v2/projects",
		`{"title":"before","description":"keep me"}`, token, "")
	require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())
	var created struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))

	// PATCH only title; AutoPatch must leave description alone.
	rec = humaRequest(t, e, http.MethodPatch, fmt.Sprintf("/api/v2/projects/%d", created.ID),
		`{"title":"after"}`, token, "application/merge-patch+json")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

	rec = humaRequest(t, e, http.MethodGet, fmt.Sprintf("/api/v2/projects/%d", created.ID), "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
	var after struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &after))
	assert.Equal(t, "after", after.Title)
	assert.Equal(t, "keep me", after.Description, "description must survive the PATCH")
}
