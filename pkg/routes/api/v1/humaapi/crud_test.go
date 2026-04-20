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

package humaapi

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/modules/humaecho5"
	"code.vikunja.io/api/pkg/web"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// fakeObj is a minimal CObject for compile-level spec-generation tests only.
// It is not exercised at runtime.
type fakeObj struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

// Satisfy handler.CObject (web.CRUDable + web.Permissions) — implementations
// are unreachable in this test because we only inspect the OpenAPI spec.

// web.CRUDable
func (*fakeObj) Create(_ *xorm.Session, _ web.Auth) error { panic("unused") }
func (*fakeObj) ReadOne(_ *xorm.Session, _ web.Auth) error {
	panic("unused")
}
func (*fakeObj) ReadAll(_ *xorm.Session, _ web.Auth, _ string, _ int, _ int) (interface{}, int, int64, error) {
	panic("unused")
}
func (*fakeObj) Update(_ *xorm.Session, _ web.Auth) error { panic("unused") }
func (*fakeObj) Delete(_ *xorm.Session, _ web.Auth) error { panic("unused") }

// web.Permissions
func (*fakeObj) CanRead(_ *xorm.Session, _ web.Auth) (bool, int, error) { panic("unused") }
func (*fakeObj) CanDelete(_ *xorm.Session, _ web.Auth) (bool, error)    { panic("unused") }
func (*fakeObj) CanUpdate(_ *xorm.Session, _ web.Auth) (bool, error)    { panic("unused") }
func (*fakeObj) CanCreate(_ *xorm.Session, _ web.Auth) (bool, error)    { panic("unused") }

// TestRegisterEmitsFiveOperations confirms that a call to Register
// produces the expected OpenAPI path entries for a simple id-based
// resource. We don't invoke the handlers here; we only inspect the spec.
func TestRegisterEmitsFiveOperations(t *testing.T) {
	e := echo.New()
	api := humaecho5.New(e, huma.DefaultConfig("spike", "0.0.1"))

	Register(api, Config[*fakeObj, SingleID]{
		Tag:       "fakes",
		BasePath:  "/fakes",
		ItemPath:  "/fakes/{id}",
		New:       func() *fakeObj { return &fakeObj{} },
		ApplyPath: func(o *fakeObj, p SingleID) { o.ID = p.ID },
	})

	spec := api.OpenAPI()
	require.NotNil(t, spec.Paths["/fakes"])
	require.NotNil(t, spec.Paths["/fakes/{id}"])

	// Five ops across two paths: list+create on base, read+update+delete on item
	ops := 0
	for _, p := range spec.Paths {
		if p.Get != nil {
			ops++
		}
		if p.Put != nil {
			ops++
		}
		if p.Post != nil {
			ops++
		}
		if p.Delete != nil {
			ops++
		}
	}
	assert.Equal(t, 5, ops)

	// Also prove the spec is 3.1.x
	req := httptest.NewRequest("GET", "/openapi.json", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	var doc map[string]any
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&doc))
	require.Contains(t, doc["openapi"].(string), "3.1")
}
