//   Vikunja is a todo-list application to facilitate your life.
//   Copyright 2019 Vikunja and contributors. All rights reserved.
//
//   This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU General Public License for more details.
//
//   You should have received a copy of the GNU General Public License
//   along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
			rec, err := testHandler.testReadAll(nil, nil)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Test1`)
			assert.NotContains(t, rec.Body.String(), `Test2`)
			assert.Contains(t, rec.Body.String(), `Test3`) // Shared directly via users_list
			assert.Contains(t, rec.Body.String(), `Test4`) // Shared via namespace
			assert.NotContains(t, rec.Body.String(), `Test5`)
		})
		t.Run("Search", func(t *testing.T) {
			rec, err := testHandler.testReadAll(url.Values{"s": []string{"Test1"}}, nil)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Test1`)
			assert.NotContains(t, rec.Body.String(), `Test2`)
			assert.NotContains(t, rec.Body.String(), `Test3`)
			assert.NotContains(t, rec.Body.String(), `Test4`)
			assert.NotContains(t, rec.Body.String(), `Test5`)
		})
	})
	t.Run("ReadOne", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testReadOne(nil, map[string]string{"list": "1"})
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Test1"`)
			assert.NotContains(t, rec.Body.String(), `"title":"Test2"`)
			assert.Contains(t, rec.Body.String(), `"owner":{"id":1,"username":"user1",`)
			assert.NotContains(t, rec.Body.String(), `"owner":{"id":2,"username":"user2",`)
			assert.Contains(t, rec.Body.String(), `"tasks":[{"id":1,"text":"task #1",`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testReadOne(nil, map[string]string{"list": "9999"})
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeListDoesNotExist)
		})
		t.Run("Rights check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user3
				_, err := testHandler.testReadOne(nil, map[string]string{"list": "2"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `You don't have the right to see this`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"list": "6"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test6"`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"list": "7"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test7"`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"list": "8"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test8"`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"list": "9"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test9"`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"list": "10"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test10"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"list": "11"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test11"`)
			})

			t.Run("Shared Via NamespaceTeam readonly", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"list": "12"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test12"`)
			})
			t.Run("Shared Via NamespaceTeam write", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"list": "13"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test13"`)
			})
			t.Run("Shared Via NamespaceTeam admin", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"list": "14"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test14"`)
			})

			t.Run("Shared Via NamespaceUser readonly", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"list": "15"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test15"`)
			})
			t.Run("Shared Via NamespaceUser write", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"list": "16"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test16"`)
			})
			t.Run("Shared Via NamespaceUser admin", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"list": "17"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Test17"`)
			})
		})
	})
	t.Run("Update", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			// Check the list was loaded successfully afterwards, see testReadOne
			rec, err := testHandler.testUpdate(nil, map[string]string{"list": "1"}, `{"title":"TestLoremIpsum"}`)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			// The description should not be updated but returned correctly
			assert.Contains(t, rec.Body.String(), `description":"Lorem Ipsum`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testUpdate(nil, map[string]string{"list": "9999"}, `{"title":"TestLoremIpsum"}`)
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeListDoesNotExist)
		})
		t.Run("Normal with updating the description", func(t *testing.T) {
			rec, err := testHandler.testUpdate(nil, map[string]string{"list": "1"}, `{"title":"TestLoremIpsum","description":"Lorem Ipsum dolor sit amet"}`)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			assert.Contains(t, rec.Body.String(), `"description":"Lorem Ipsum dolor sit amet`)
		})
		t.Run("Empty title", func(t *testing.T) {
			_, err := testHandler.testUpdate(nil, map[string]string{"list": "1"}, `{"title":""}`)
			assert.Error(t, err)
			assert.Contains(t, err.(*echo.HTTPError).Message.(models.ValidationHTTPError).InvalidFields, "title: non zero value required")
		})
		t.Run("Almost empty title", func(t *testing.T) {
			_, err := testHandler.testUpdate(nil, map[string]string{"list": "1"}, `{"title":"nn"}`)
			assert.Error(t, err)
			assert.Contains(t, err.(*echo.HTTPError).Message.(models.ValidationHTTPError).InvalidFields[0], "does not validate as runelength(3|250)")
		})
		t.Run("Title too long", func(t *testing.T) {
			_, err := testHandler.testUpdate(nil, map[string]string{"list": "1"}, `{"title":"Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea taki"}`)
			assert.Error(t, err)
			assert.Contains(t, err.(*echo.HTTPError).Message.(models.ValidationHTTPError).InvalidFields[0], "does not validate as runelength(3|250)")
		})
		t.Run("Rights check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user3
				_, err := testHandler.testUpdate(nil, map[string]string{"list": "2"}, `{"title":"TestLoremIpsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testUpdate(nil, map[string]string{"list": "6"}, `{"title":"TestLoremIpsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"list": "7"}, `{"title":"TestLoremIpsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"list": "8"}, `{"title":"TestLoremIpsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testUpdate(nil, map[string]string{"list": "9"}, `{"title":"TestLoremIpsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"list": "10"}, `{"title":"TestLoremIpsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"list": "11"}, `{"title":"TestLoremIpsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})

			t.Run("Shared Via NamespaceTeam readonly", func(t *testing.T) {
				_, err := testHandler.testUpdate(nil, map[string]string{"list": "12"}, `{"title":"TestLoremIpsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceTeam write", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"list": "13"}, `{"title":"TestLoremIpsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via NamespaceTeam admin", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"list": "14"}, `{"title":"TestLoremIpsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})

			t.Run("Shared Via NamespaceUser readonly", func(t *testing.T) {
				_, err := testHandler.testUpdate(nil, map[string]string{"list": "15"}, `{"title":"TestLoremIpsum"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceUser write", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"list": "16"}, `{"title":"TestLoremIpsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
			t.Run("Shared Via NamespaceUser admin", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"list": "17"}, `{"title":"TestLoremIpsum"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"TestLoremIpsum"`)
			})
		})
	})
	t.Run("Delete", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testDelete(nil, map[string]string{"list": "1"})
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testDelete(nil, map[string]string{"list": "999"})
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeListDoesNotExist)
		})
		t.Run("Rights check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user3
				_, err := testHandler.testDelete(nil, map[string]string{"list": "2"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"list": "6"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"list": "7"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testDelete(nil, map[string]string{"list": "8"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"list": "9"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"list": "10"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testDelete(nil, map[string]string{"list": "11"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})

			t.Run("Shared Via NamespaceTeam readonly", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"list": "12"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceTeam write", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"list": "13"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceTeam admin", func(t *testing.T) {
				rec, err := testHandler.testDelete(nil, map[string]string{"list": "14"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})

			t.Run("Shared Via NamespaceUser readonly", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"list": "15"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceUser write", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"list": "16"})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceUser admin", func(t *testing.T) {
				rec, err := testHandler.testDelete(nil, map[string]string{"list": "17"})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"message":"Successfully deleted."`)
			})
		})
	})
	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			// Check the list was loaded successfully after update, see testReadOne
			rec, err := testHandler.testCreate(nil, map[string]string{"namespace": "1"}, `{"title":"Lorem"}`)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
			assert.Contains(t, rec.Body.String(), `"description":""`)
			assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
			assert.Contains(t, rec.Body.String(), `"tasks":null`)
		})
		t.Run("Normal with description", func(t *testing.T) {
			rec, err := testHandler.testCreate(nil, map[string]string{"namespace": "1"}, `{"title":"Lorem","description":"Lorem Ipsum"}`)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
			assert.Contains(t, rec.Body.String(), `"description":"Lorem Ipsum"`)
			assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
			assert.Contains(t, rec.Body.String(), `"tasks":null`)
		})
		t.Run("Nonexisting Namespace", func(t *testing.T) {
			_, err := testHandler.testCreate(nil, map[string]string{"namespace": "999999"}, `{"title":"Lorem"}`)
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeNamespaceDoesNotExist)
		})
		t.Run("Empty title", func(t *testing.T) {
			_, err := testHandler.testCreate(nil, map[string]string{"namespace": "1"}, `{"title":""}`)
			assert.Error(t, err)
			assert.Contains(t, err.(*echo.HTTPError).Message.(models.ValidationHTTPError).InvalidFields, "title: non zero value required")
		})
		t.Run("Almost empty title", func(t *testing.T) {
			_, err := testHandler.testCreate(nil, map[string]string{"namespace": "1"}, `{"title":"nn"}`)
			assert.Error(t, err)
			assert.Contains(t, err.(*echo.HTTPError).Message.(models.ValidationHTTPError).InvalidFields[0], "does not validate as runelength(3|250)")
		})
		t.Run("Title too long", func(t *testing.T) {
			_, err := testHandler.testCreate(nil, map[string]string{"namespace": "1"}, `{"title":"Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. Stet clita kasd gubergren, no sea taki"}`)
			assert.Error(t, err)
			assert.Contains(t, err.(*echo.HTTPError).Message.(models.ValidationHTTPError).InvalidFields[0], "does not validate as runelength(3|250)")
		})
		t.Run("Rights check", func(t *testing.T) {

			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user3
				_, err := testHandler.testCreate(nil, map[string]string{"namespace": "3"}, `{"title":"Lorem"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})

			t.Run("Shared Via NamespaceTeam readonly", func(t *testing.T) {
				_, err := testHandler.testCreate(nil, map[string]string{"namespace": "7"}, `{"title":"Lorem"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceTeam write", func(t *testing.T) {
				rec, err := testHandler.testCreate(nil, map[string]string{"namespace": "8"}, `{"title":"Lorem"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
				assert.Contains(t, rec.Body.String(), `"tasks":null`)
			})
			t.Run("Shared Via NamespaceTeam admin", func(t *testing.T) {
				rec, err := testHandler.testCreate(nil, map[string]string{"namespace": "9"}, `{"title":"Lorem"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
				assert.Contains(t, rec.Body.String(), `"tasks":null`)
			})

			t.Run("Shared Via NamespaceUser readonly", func(t *testing.T) {
				_, err := testHandler.testCreate(nil, map[string]string{"namespace": "10"}, `{"title":"Lorem"}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceUser write", func(t *testing.T) {
				rec, err := testHandler.testCreate(nil, map[string]string{"namespace": "11"}, `{"title":"Lorem"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
				assert.Contains(t, rec.Body.String(), `"tasks":null`)
			})
			t.Run("Shared Via NamespaceUser admin", func(t *testing.T) {
				rec, err := testHandler.testCreate(nil, map[string]string{"namespace": "12"}, `{"title":"Lorem"}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"title":"Lorem"`)
				assert.Contains(t, rec.Body.String(), `"description":""`)
				assert.Contains(t, rec.Body.String(), `"owner":{"id":1`)
				assert.Contains(t, rec.Body.String(), `"tasks":null`)
			})
		})
	})
}
