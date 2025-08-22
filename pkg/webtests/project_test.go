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

func TestProject(t *testing.T) {
	th := NewTestHelper(t)
	th.Login(t, &testuser1)

	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := th.Request(t, "GET", "/api/v1/projects", nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Test1`)
			assert.NotContains(t, rec.Body.String(), `Test2"`)
			assert.Contains(t, rec.Body.String(), `Test3`)  // Shared directly via users_project
			assert.Contains(t, rec.Body.String(), `Test12`) // Shared via parent project
			assert.NotContains(t, rec.Body.String(), `Test5`)
			assert.NotContains(t, rec.Body.String(), `Test22`) // Archived directly
		})
		t.Run("Search", func(t *testing.T) {
			rec, err := th.Request(t, "GET", "/api/v1/projects?s=Test1", nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Test1`)
			assert.NotContains(t, rec.Body.String(), `Test2`)
			assert.NotContains(t, rec.Body.String(), `Test3`)
			assert.NotContains(t, rec.Body.String(), `Test4`)
			assert.NotContains(t, rec.Body.String(), `Test5`)
		})
		t.Run("Normal with archived projects", func(t *testing.T) {
			rec, err := th.Request(t, "GET", "/api/v1/projects?is_archived=true", nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Test1`)
			assert.NotContains(t, rec.Body.String(), `Test2"`)
			assert.Contains(t, rec.Body.String(), `Test3`)  // Shared directly via users_project
			assert.Contains(t, rec.Body.String(), `Test12`) // Shared via parent project
			assert.NotContains(t, rec.Body.String(), `Test5`)
			assert.Contains(t, rec.Body.String(), `Test21`) // Archived through project
			assert.Contains(t, rec.Body.String(), `Test22`) // Archived directly
		})
	})
	t.Run("ReadOne", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := th.Request(t, "GET", "/api/v1/projects/1", nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Test1"`)
			assert.NotContains(t, rec.Body.String(), `"title":"Test2"`)
			assert.Contains(t, rec.Body.String(), `"owner":{"id":1,"name":"","username":"user1",`)
			assert.NotContains(t, rec.Body.String(), `"owner":{"id":2,"name":"","username":"user2",`)
			assert.NotContains(t, rec.Body.String(), `"tasks":`)
			assert.Equal(t, "2", rec.Header().Get("x-max-permission")) // User 1 is owner, so they should have admin permissions.
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := th.Request(t, "GET", "/api/v1/projects/9999", nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), `"code":3001`)
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := th.Request(t, "GET", "/api/v1/projects/20", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				rec, err := th.Request(t, "GET", "/api/v1/projects/6", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test6"`)
				assert.Equal(t, "0", rec.Header().Get("x-max-permission"))
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := th.Request(t, "GET", "/api/v1/projects/7", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test7"`)
				assert.Equal(t, "1", rec.Header().Get("x-max-permission"))
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := th.Request(t, "GET", "/api/v1/projects/8", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test8"`)
				assert.Equal(t, "2", rec.Header().Get("x-max-permission"))
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				rec, err := th.Request(t, "GET", "/api/v1/projects/9", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test9"`)
				assert.Equal(t, "0", rec.Header().Get("x-max-permission"))
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := th.Request(t, "GET", "/api/v1/projects/10", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test10"`)
				assert.Equal(t, "1", rec.Header().Get("x-max-permission"))
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := th.Request(t, "GET", "/api/v1/projects/11", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test11"`)
				assert.Equal(t, "2", rec.Header().Get("x-max-permission"))
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				rec, err := th.Request(t, "GET", "/api/v1/projects/12", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test12"`)
				assert.Equal(t, "0", rec.Header().Get("x-max-permission"))
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				rec, err := th.Request(t, "GET", "/api/v1/projects/13", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test13"`)
				assert.Equal(t, "1", rec.Header().Get("x-max-permission"))
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := th.Request(t, "GET", "/api/v1/projects/14", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test14"`)
				assert.Equal(t, "2", rec.Header().Get("x-max-permission"))
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				rec, err := th.Request(t, "GET", "/api/v1/projects/15", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test15"`)
				assert.Equal(t, "0", rec.Header().Get("x-max-permission"))
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				rec, err := th.Request(t, "GET", "/api/v1/projects/16", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test16"`)
				assert.Equal(t, "1", rec.Header().Get("x-max-permission"))
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := th.Request(t, "GET", "/api/v1/projects/17", nil)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test17"`)
				assert.Equal(t, "2", rec.Header().Get("x-max-permission"))
			})
		})
	})
	t.Run("Update", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			// Check the project was loaded successfully afterwards, see testReadOneWithUser
			rec, err := th.Request(t, "POST", "/api/v1/projects/1", strings.NewReader(`{"title":"TestLoremIpsum"}`))
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			// The description should not be updated but returned correctly
			assert.Contains(t, rec.Body.String(), `description":"Lorem Ipsum`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := th.Request(t, "POST", "/api/v1/projects/9999", strings.NewReader(`{"title":"TestLoremIpsum"}`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), `This project does not exist.`)
		})
		t.Run("Normal with updating the description", func(t *testing.T) {
			rec, err := th.Request(t, "POST", "/api/v1/projects/1", strings.NewReader(`{"title":"TestLoremIpsum","description":"Lorem Ipsum dolor sit amet"}`))
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			assert.Contains(t, rec.Body.String(), `"description":"Lorem Ipsum dolor sit amet`)
		})
		t.Run("Empty title", func(t *testing.T) {
			_, err := th.Request(t, "POST", "/api/v1/projects/1", strings.NewReader(`{"title":""}`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), `Struct is invalid`)
		})
		t.Run("Title too long", func(t *testing.T) {
			_, err := th.Request(t, "POST", "/api/v1/projects/1", strings.NewReader(`{"title":"Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea taki"}`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), `Struct is invalid`)
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := th.Request(t, "POST", "/api/v1/projects/20", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/projects/6", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/projects/7", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/projects/8", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/projects/9", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/projects/10", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/projects/11", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/projects/12", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `Forbidden`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/projects/13", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/projects/14", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/projects/15", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `Forbidden`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/projects/16", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := th.Request(t, "POST", "/api/v1/projects/17", strings.NewReader(`{"title":"TestLoremIpsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
		})
	})
	t.Run("Delete", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := th.Request(t, "DELETE", "/api/v1/projects/1", nil)
			require.NoError(t, err)
			assert.Equal(t, 204, rec.Code)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := th.Request(t, "DELETE", "/api/v1/projects/999", nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), `This project does not exist.`)
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := th.Request(t, "DELETE", "/api/v1/projects/20", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/projects/6", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/projects/7", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `Forbidden`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := th.Request(t, "DELETE", "/api/v1/projects/8", nil)
				require.NoError(t, err)
				assert.Equal(t, 204, rec.Code)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/projects/9", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/projects/10", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `Forbidden`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := th.Request(t, "DELETE", "/api/v1/projects/11", nil)
				require.NoError(t, err)
				assert.Equal(t, 204, rec.Code)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/projects/12", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `Forbidden`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/projects/13", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `Forbidden`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := th.Request(t, "DELETE", "/api/v1/projects/14", nil)
				require.NoError(t, err)
				assert.Equal(t, 204, rec.Code)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/projects/15", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `Forbidden`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/projects/16", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `Forbidden`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := th.Request(t, "DELETE", "/api/v1/projects/17", nil)
				require.NoError(t, err)
				assert.Equal(t, 204, rec.Code)
			})
		})
	})
	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			// Check the project was loaded successfully after update, see testReadOneWithUser
			rec, err := th.Request(t, "PUT", "/api/v1/projects", strings.NewReader(`{"title":"Lorem"}`))
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
			assert.Contains(t, rec.Body.String(), `"description":""`)
			assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
			assert.NotContains(t, rec.Body.String(), `"tasks":`)
		})
		t.Run("Normal with description", func(t *testing.T) {
			rec, err := th.Request(t, "PUT", "/api/v1/projects", strings.NewReader(`{"title":"Lorem","description":"Lorem Ipsum"}`))
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
			assert.Contains(t, rec.Body.String(), `"description":"Lorem Ipsum"`)
			assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
			assert.NotContains(t, rec.Body.String(), `"tasks":`)
		})
		t.Run("Nonexisting parent project", func(t *testing.T) {
			_, err := th.Request(t, "PUT", "/api/v1/projects", strings.NewReader(`{"title":"Lorem","parent_project_id":99999}`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), `This project does not exist.`)
		})
		t.Run("Empty title", func(t *testing.T) {
			_, err := th.Request(t, "PUT", "/api/v1/projects", strings.NewReader(`{"title":""}`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), `Struct is invalid`)
		})
		t.Run("Title too long", func(t *testing.T) {
			_, err := th.Request(t, "PUT", "/api/v1/projects", strings.NewReader(`{"title":"Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea taki"}`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), `Struct is invalid`)
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := th.Request(t, "PUT", "/api/v1/projects", strings.NewReader(`{"title":"Lorem","parent_project_id":20}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `Forbidden`)
			})
			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "PUT", "/api/v1/projects", strings.NewReader(`{"title":"Lorem","parent_project_id":32}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `Forbidden`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/projects", strings.NewReader(`{"title":"Lorem","parent_project_id":33}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
				assert.NotContains(t, rec.Body.String(), `"tasks":`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/projects", strings.NewReader(`{"title":"Lorem","parent_project_id":34}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
				assert.NotContains(t, rec.Body.String(), `"tasks":`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := th.Request(t, "PUT", "/api/v1/projects", strings.NewReader(`{"title":"Lorem","parent_project_id":9}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `Forbidden`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/projects", strings.NewReader(`{"title":"Lorem","parent_project_id":10}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
				assert.NotContains(t, rec.Body.String(), `"tasks":`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/projects", strings.NewReader(`{"title":"Lorem","parent_project_id":8}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
				assert.NotContains(t, rec.Body.String(), `"tasks":`)
			})
		})
	})
}
