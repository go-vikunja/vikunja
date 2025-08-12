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
	"net/url"
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProject(t *testing.T) {
	testHandler := webHandlerTest{
		user: &testuser1,
		strFunc: func() handler.CObject {
			return &models.Project{}
		},
		t: t,
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
			assert.NotContains(t, rec.Body.String(), `Test22`) // Archived directly
		})
		t.Run("Search", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(url.Values{"s": []string{"Test1"}}, nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Test1`)
			assert.NotContains(t, rec.Body.String(), `Test2`)
			assert.NotContains(t, rec.Body.String(), `Test3`)
			assert.NotContains(t, rec.Body.String(), `Test4`)
			assert.NotContains(t, rec.Body.String(), `Test5`)
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
		})
	})
	t.Run("ReadOne", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "1"})
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Test1"`)
			assert.NotContains(t, rec.Body.String(), `"title":"Test2"`)
			assert.Contains(t, rec.Body.String(), `"owner":{"id":1,"name":"","username":"user1",`)
			assert.NotContains(t, rec.Body.String(), `"owner":{"id":2,"name":"","username":"user2",`)
			assert.NotContains(t, rec.Body.String(), `"tasks":`)
			assert.Equal(t, "2", rec.Result().Header.Get("x-max-permission")) // User 1 is owner, so they should have admin permissions.
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "9999"})
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeProjectDoesNotExist)
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "20"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `You don't have the permission to see this`)
				assert.Empty(t, rec.Result().Header.Get("x-max-permissions"))
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "6"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test6"`)
				assert.Equal(t, "0", rec.Result().Header.Get("x-max-permission"))
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "7"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test7"`)
				assert.Equal(t, "1", rec.Result().Header.Get("x-max-permission"))
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "8"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test8"`)
				assert.Equal(t, "2", rec.Result().Header.Get("x-max-permission"))
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "9"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test9"`)
				assert.Equal(t, "0", rec.Result().Header.Get("x-max-permission"))
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "10"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test10"`)
				assert.Equal(t, "1", rec.Result().Header.Get("x-max-permission"))
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "11"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test11"`)
				assert.Equal(t, "2", rec.Result().Header.Get("x-max-permission"))
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "12"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test12"`)
				assert.Equal(t, "0", rec.Result().Header.Get("x-max-permission"))
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "13"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test13"`)
				assert.Equal(t, "1", rec.Result().Header.Get("x-max-permission"))
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "14"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test14"`)
				assert.Equal(t, "2", rec.Result().Header.Get("x-max-permission"))
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "15"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test15"`)
				assert.Equal(t, "0", rec.Result().Header.Get("x-max-permission"))
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "16"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test16"`)
				assert.Equal(t, "1", rec.Result().Header.Get("x-max-permission"))
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"project": "17"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test17"`)
				assert.Equal(t, "2", rec.Result().Header.Get("x-max-permission"))
			})
		})
	})
	t.Run("Update", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			// Check the project was loaded successfully afterwards, see testReadOneWithUser
			rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "1"}, `{"title":"TestLoremIpsum"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			// The description should not be updated but returned correctly
			assert.Contains(t, rec.Body.String(), `description":"Lorem Ipsum`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "9999"}, `{"title":"TestLoremIpsum"}`)
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeProjectDoesNotExist)
		})
		t.Run("Normal with updating the description", func(t *testing.T) {
			rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "1"}, `{"title":"TestLoremIpsum","description":"Lorem Ipsum dolor sit amet"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			assert.Contains(t, rec.Body.String(), `"description":"Lorem Ipsum dolor sit amet`)
		})
		t.Run("Empty title", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "1"}, `{"title":""}`)
			require.Error(t, err)
			assert.Contains(t, err.(*echo.HTTPError).Message.(models.ValidationHTTPError).InvalidFields, "title: non zero value required")
		})
		t.Run("Title too long", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "1"}, `{"title":"Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea taki"}`)
			require.Error(t, err)
			assert.Contains(t, err.(*echo.HTTPError).Message.(models.ValidationHTTPError).InvalidFields[0], "does not validate as runelength(1|250)")
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "20"}, `{"title":"TestLoremIpsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "6"}, `{"title":"TestLoremIpsum"}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
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
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
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
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
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
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
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
			assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "999"})
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeProjectDoesNotExist)
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "20"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "6"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "7"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "8"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "9"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "10"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "11"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "12"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "13"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "14"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "15"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "16"})
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"project": "17"})
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})
		})
	})
	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			// Check the project was loaded successfully after update, see testReadOneWithUser
			rec, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
			assert.Contains(t, rec.Body.String(), `"description":""`)
			assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
			assert.NotContains(t, rec.Body.String(), `"tasks":`)
		})
		t.Run("Normal with description", func(t *testing.T) {
			rec, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem","description":"Lorem Ipsum"}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
			assert.Contains(t, rec.Body.String(), `"description":"Lorem Ipsum"`)
			assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
			assert.NotContains(t, rec.Body.String(), `"tasks":`)
		})
		t.Run("Nonexisting parent project", func(t *testing.T) {
			_, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem","parent_project_id":99999}`)
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeProjectDoesNotExist)
		})
		t.Run("Empty title", func(t *testing.T) {
			_, err := testHandler.testCreateWithUser(nil, nil, `{"title":""}`)
			require.Error(t, err)
			assert.Contains(t, err.(*echo.HTTPError).Message.(models.ValidationHTTPError).InvalidFields, "title: non zero value required")
		})
		t.Run("Title too long", func(t *testing.T) {
			_, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea taki"}`)
			require.Error(t, err)
			assert.Contains(t, err.(*echo.HTTPError).Message.(models.ValidationHTTPError).InvalidFields[0], "does not validate as runelength(1|250)")
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem","parent_project_id":20}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem","parent_project_id":32}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem","parent_project_id":33}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
				assert.NotContains(t, rec.Body.String(), `"tasks":`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem","parent_project_id":34}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
				assert.NotContains(t, rec.Body.String(), `"tasks":`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem","parent_project_id":9}`)
				require.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, nil, `{"title":"Lorem","parent_project_id":10}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
				assert.NotContains(t, rec.Body.String(), `"tasks":`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"namespace": "12"}, `{"title":"Lorem","parent_project_id":11}`)
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
				assert.NotContains(t, rec.Body.String(), `"tasks":`)
			})
		})
	})
}
