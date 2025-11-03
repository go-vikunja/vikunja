//   Vikunja is a to-do list application to facilitate your life.
//   Copyright 2018-present Vikunja and contributors. All rights reserved.
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

package webtests

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func Test${MODEL}(t *testing.T) {
	testHandler := webHandlerTest{
		user: &testuser1,
		strFunc: func() handler.CObject {
			return &models.${MODEL}{}
		},
		t: t,
	}
	t.Run("ReadAll", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testReadAll(nil, nil)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), ``)
			assert.NotContains(t, rec.Body.String(), ``)
		})
		t.Run("Search", func(t *testing.T) {
			rec, err := testHandler.testReadAll(url.Values{"s": []string{""}}, nil)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), ``)
			assert.NotContains(t, rec.Body.String(), ``)
		})
	})
	t.Run("ReadOne", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testReadOne(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), ``)
			assert.NotContains(t, rec.Body.String(), ``)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testReadOne(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCode)
		})
		t.Run("Rights check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user3
				_, err := testHandler.testReadOne(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `You don't have the right to see this`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})

			t.Run("Shared Via NamespaceTeam readonly", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})
			t.Run("Shared Via NamespaceTeam write", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})
			t.Run("Shared Via NamespaceTeam admin", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})

			t.Run("Shared Via NamespaceUser readonly", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})
			t.Run("Shared Via NamespaceUser write", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})
			t.Run("Shared Via NamespaceUser admin", func(t *testing.T) {
				rec, err := testHandler.testReadOne(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})
		})
	})
	t.Run("Update", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testUpdate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), ``)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testUpdate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCode)
		})
		t.Run("Rights check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				_, err := testHandler.testUpdate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testUpdate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testUpdate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})

			t.Run("Shared Via NamespaceTeam readonly", func(t *testing.T) {
				_, err := testHandler.testUpdate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceTeam write", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})
			t.Run("Shared Via NamespaceTeam admin", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})

			t.Run("Shared Via NamespaceUser readonly", func(t *testing.T) {
				_, err := testHandler.testUpdate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceUser write", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})
			t.Run("Shared Via NamespaceUser admin", func(t *testing.T) {
				rec, err := testHandler.testUpdate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})
		})
	})
	t.Run("Delete", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testDelete(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), ``)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testDelete(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCode)
		})
		t.Run("Rights check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := testHandler.testDelete(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := testHandler.testDelete(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})

			t.Run("Shared Via NamespaceTeam readonly", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceTeam write", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceTeam admin", func(t *testing.T) {
				rec, err := testHandler.testDelete(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})

			t.Run("Shared Via NamespaceUser readonly", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceUser write", func(t *testing.T) {
				_, err := testHandler.testDelete(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceUser admin", func(t *testing.T) {
				rec, err := testHandler.testDelete(nil, map[string]string{"${URL_PLACEHOLDER}": ""})
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})
		})
	})
	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := testHandler.testCreate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), ``)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := testHandler.testCreate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCode)
		})
		t.Run("Rights check", func(t *testing.T) {

			t.Run("Forbidden", func(t *testing.T) {
				_, err := testHandler.testCreate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})

			t.Run("Shared Via NamespaceTeam readonly", func(t *testing.T) {
				_, err := testHandler.testCreate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceTeam write", func(t *testing.T) {
				rec, err := testHandler.testCreate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})
			t.Run("Shared Via NamespaceTeam admin", func(t *testing.T) {
				rec, err := testHandler.testCreate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})

			t.Run("Shared Via NamespaceUser readonly", func(t *testing.T) {
				_, err := testHandler.testCreate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
				assert.Error(t, err)
				assert.Contains(t, err.(*echo.HTTPError).Message, `Forbidden`)
			})
			t.Run("Shared Via NamespaceUser write", func(t *testing.T) {
				rec, err := testHandler.testCreate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})
			t.Run("Shared Via NamespaceUser admin", func(t *testing.T) {
				rec, err := testHandler.testCreate(nil, map[string]string{"${URL_PLACEHOLDER}": ""}, `{}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), ``)
			})
		})
	})
}
