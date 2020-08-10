// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package integrations

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/web/handler"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestList(t *testing.T) {
	testHandler := webHandlerTest{
		user: &testuser1,
		strFunc: func() handler.CObject {
			return &models.List{}
		},
		t: t,
	}
	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(nil, nil)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Test1`)
			assert.NotContains(t, rec.Body.String(), `Test2"`)
			assert.Contains(t, rec.Body.String(), `Test3`) // Shared directly via users_list
			assert.Contains(t, rec.Body.String(), `Test4`) // Shared via namespace
			assert.NotContains(t, rec.Body.String(), `Test5`)
			assert.NotContains(t, rec.Body.String(), `Test21`) // Archived through namespace
			assert.NotContains(t, rec.Body.String(), `Test22`) // Archived directly
		})
		t.Run("Search", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(url.Values{"s": []string{"Test1"}}, nil)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Test1`)
			assert.NotContains(t, rec.Body.String(), `Test2`)
			assert.NotContains(t, rec.Body.String(), `Test3`)
			assert.NotContains(t, rec.Body.String(), `Test4`)
			assert.NotContains(t, rec.Body.String(), `Test5`)
		})
		t.Run("Normal with archived lists", func(t *testing.T) {
			rec, err := testHandler.testReadAllWithUser(url.Values{"is_archived": []string{"true"}}, nil)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Test1`)
			assert.NotContains(t, rec.Body.String(), `Test2"`)
			assert.Contains(t, rec.Body.String(), `Test3`) // Shared directly via users_list
			assert.Contains(t, rec.Body.String(), `Test4`) // Shared via namespace
			assert.NotContains(t, rec.Body.String(), `Test5`)
			assert.Contains(t, rec.Body.String(), `Test21`) // Archived through namespace
			assert.Contains(t, rec.Body.String(), `Test22`) // Archived directly
		})
	})
	t.Run("ReadOne", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"list": "1"})
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Test1"`)
			assert.NotContains(t, rec.Body.String(), `"title":"Test2"`)
			assert.Contains(t, rec.Body.String(), `"owner":{"id":1,"username":"user1",`)
			assert.NotContains(t, rec.Body.String(), `"owner":{"id":2,"username":"user2",`)
			assert.NotContains(t, rec.Body.String(), `"tasks":`)
			assert.Equal(t, "2", rec.Result().Header.Get("x-max-right")) // User 1 is owner so they should have admin rights.
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testReadOneWithUser(nil, map[string]string{"list": "9999"})
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeListDoesNotExist)
		})
		t.Run("Rights check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"list": "20"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `You don't have the right to see this`)
				assert.Empty(t, rec.Result().Header.Get("x-max-rights"))
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"list": "6"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test6"`)
				assert.Equal(t, "0", rec.Result().Header.Get("x-max-right"))
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"list": "7"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test7"`)
				assert.Equal(t, "1", rec.Result().Header.Get("x-max-right"))
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"list": "8"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test8"`)
				assert.Equal(t, "2", rec.Result().Header.Get("x-max-right"))
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"list": "9"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test9"`)
				assert.Equal(t, "0", rec.Result().Header.Get("x-max-right"))
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"list": "10"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test10"`)
				assert.Equal(t, "1", rec.Result().Header.Get("x-max-right"))
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"list": "11"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test11"`)
				assert.Equal(t, "2", rec.Result().Header.Get("x-max-right"))
			})

			t.Run("Shared Via NamespaceTeam readonly", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"list": "12"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test12"`)
				assert.Equal(t, "0", rec.Result().Header.Get("x-max-right"))
			})
			t.Run("Shared Via NamespaceTeam write", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"list": "13"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test13"`)
				assert.Equal(t, "1", rec.Result().Header.Get("x-max-right"))
			})
			t.Run("Shared Via NamespaceTeam admin", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"list": "14"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test14"`)
				assert.Equal(t, "2", rec.Result().Header.Get("x-max-right"))
			})

			t.Run("Shared Via NamespaceUser readonly", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"list": "15"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test15"`)
				assert.Equal(t, "0", rec.Result().Header.Get("x-max-right"))
			})
			t.Run("Shared Via NamespaceUser write", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"list": "16"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test16"`)
				assert.Equal(t, "1", rec.Result().Header.Get("x-max-right"))
			})
			t.Run("Shared Via NamespaceUser admin", func(t *testing.T) {
				rec, err := testHandler.testReadOneWithUser(nil, map[string]string{"list": "17"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test17"`)
				assert.Equal(t, "2", rec.Result().Header.Get("x-max-right"))
			})
		})
	})
	t.Run("Update", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			// Check the list was loaded successfully afterwards, see testReadOneWithUser
			rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"list": "1"}, `{"title":"TestLoremIpsum"}`)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			// The description should not be updated but returned correctly
			assert.Contains(t, rec.Body.String(), `description":"Lorem Ipsum`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"list": "9999"}, `{"title":"TestLoremIpsum"}`)
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeListDoesNotExist)
		})
		t.Run("Normal with updating the description", func(t *testing.T) {
			rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"list": "1"}, `{"title":"TestLoremIpsum","description":"Lorem Ipsum dolor sit amet"}`)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			assert.Contains(t, rec.Body.String(), `"description":"Lorem Ipsum dolor sit amet`)
		})
		t.Run("Empty title", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"list": "1"}, `{"title":""}`)
			assert.Error(t, err)
			assert.Contains(t, err.(*echo.HTTPError).Message.(models.ValidationHTTPError).InvalidFields, "title: non zero value required")
		})
		t.Run("Title too long", func(t *testing.T) {
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"list": "1"}, `{"title":"Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea taki"}`)
			assert.Error(t, err)
			assert.Contains(t, err.(*echo.HTTPError).Message.(models.ValidationHTTPError).InvalidFields[0], "does not validate as runelength(1|250)")
		})
		t.Run("Rights check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"list": "20"}, `{"title":"TestLoremIpsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"list": "6"}, `{"title":"TestLoremIpsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"list": "7"}, `{"title":"TestLoremIpsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"list": "8"}, `{"title":"TestLoremIpsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"list": "9"}, `{"title":"TestLoremIpsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"list": "10"}, `{"title":"TestLoremIpsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"list": "11"}, `{"title":"TestLoremIpsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})

			t.Run("Shared Via NamespaceTeam readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"list": "12"}, `{"title":"TestLoremIpsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceTeam write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"list": "13"}, `{"title":"TestLoremIpsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via NamespaceTeam admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"list": "14"}, `{"title":"TestLoremIpsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})

			t.Run("Shared Via NamespaceUser readonly", func(t *testing.T) {
				_, err := testHandler.testUpdateWithUser(nil, map[string]string{"list": "15"}, `{"title":"TestLoremIpsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceUser write", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"list": "16"}, `{"title":"TestLoremIpsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via NamespaceUser admin", func(t *testing.T) {
				rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"list": "17"}, `{"title":"TestLoremIpsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
		})
	})
	t.Run("Delete", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"list": "1"})
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testDeleteWithUser(nil, map[string]string{"list": "999"})
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeListDoesNotExist)
		})
		t.Run("Rights check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"list": "20"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"list": "6"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"list": "7"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"list": "8"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"list": "9"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"list": "10"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"list": "11"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})

			t.Run("Shared Via NamespaceTeam readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"list": "12"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceTeam write", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"list": "13"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceTeam admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"list": "14"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})

			t.Run("Shared Via NamespaceUser readonly", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"list": "15"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceUser write", func(t *testing.T) {
				_, err := testHandler.testDeleteWithUser(nil, map[string]string{"list": "16"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceUser admin", func(t *testing.T) {
				rec, err := testHandler.testDeleteWithUser(nil, map[string]string{"list": "17"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})
		})
	})
	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			// Check the list was loaded successfully after update, see testReadOneWithUser
			rec, err := testHandler.testCreateWithUser(nil, map[string]string{"namespace": "1"}, `{"title":"Lorem"}`)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
			assert.Contains(t, rec.Body.String(), `"description":""`)
			assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
			assert.NotContains(t, rec.Body.String(), `"tasks":`)
		})
		t.Run("Normal with description", func(t *testing.T) {
			rec, err := testHandler.testCreateWithUser(nil, map[string]string{"namespace": "1"}, `{"title":"Lorem","description":"Lorem Ipsum"}`)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
			assert.Contains(t, rec.Body.String(), `"description":"Lorem Ipsum"`)
			assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
			assert.NotContains(t, rec.Body.String(), `"tasks":`)
		})
		t.Run("Nonexisting Namespace", func(t *testing.T) {
			_, err := testHandler.testCreateWithUser(nil, map[string]string{"namespace": "999999"}, `{"title":"Lorem"}`)
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeNamespaceDoesNotExist)
		})
		t.Run("Empty title", func(t *testing.T) {
			_, err := testHandler.testCreateWithUser(nil, map[string]string{"namespace": "1"}, `{"title":""}`)
			assert.Error(t, err)
			assert.Contains(t, err.(*echo.HTTPError).Message.(models.ValidationHTTPError).InvalidFields, "title: non zero value required")
		})
		t.Run("Title too long", func(t *testing.T) {
			_, err := testHandler.testCreateWithUser(nil, map[string]string{"namespace": "1"}, `{"title":"Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea taki"}`)
			assert.Error(t, err)
			assert.Contains(t, err.(*echo.HTTPError).Message.(models.ValidationHTTPError).InvalidFields[0], "does not validate as runelength(1|250)")
		})
		t.Run("Rights check", func(t *testing.T) {

			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"namespace": "15"}, `{"title":"Lorem"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})

			t.Run("Shared Via NamespaceTeam readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"namespace": "7"}, `{"title":"Lorem"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceTeam write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"namespace": "8"}, `{"title":"Lorem"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
				assert.NotContains(t, rec.Body.String(), `"tasks":`)
			})
			t.Run("Shared Via NamespaceTeam admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"namespace": "9"}, `{"title":"Lorem"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
				assert.NotContains(t, rec.Body.String(), `"tasks":`)
			})

			t.Run("Shared Via NamespaceUser readonly", func(t *testing.T) {
				_, err := testHandler.testCreateWithUser(nil, map[string]string{"namespace": "10"}, `{"title":"Lorem"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceUser write", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"namespace": "11"}, `{"title":"Lorem"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
				assert.NotContains(t, rec.Body.String(), `"tasks":`)
			})
			t.Run("Shared Via NamespaceUser admin", func(t *testing.T) {
				rec, err := testHandler.testCreateWithUser(nil, map[string]string{"namespace": "12"}, `{"title":"Lorem"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
				assert.NotContains(t, rec.Body.String(), `"tasks":`)
			})
		})
	})
}
