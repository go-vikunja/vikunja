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

package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// Each instance must be checked individually: handler.Do* runs against a fresh Model() per call.
type stubCObject struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`

	called string
	// Permission checks always allow; denial scenarios live in the integration tests.
	returnErr error
}

func (s *stubCObject) CanRead(_ *xorm.Session, _ web.Auth) (bool, int, error) {
	return true, 0, nil
}
func (s *stubCObject) CanDelete(_ *xorm.Session, _ web.Auth) (bool, error) { return true, nil }
func (s *stubCObject) CanUpdate(_ *xorm.Session, _ web.Auth) (bool, error) { return true, nil }
func (s *stubCObject) CanCreate(_ *xorm.Session, _ web.Auth) (bool, error) { return true, nil }

func (s *stubCObject) Create(_ *xorm.Session, _ web.Auth) error {
	s.called = "Create"
	return s.returnErr
}
func (s *stubCObject) ReadOne(_ *xorm.Session, _ web.Auth) error {
	s.called = "ReadOne"
	return s.returnErr
}
func (s *stubCObject) ReadAll(_ *xorm.Session, _ web.Auth, search string, page, perPage int) (any, int, int64, error) {
	s.called = "ReadAll"
	return []string{search}, page, int64(perPage), s.returnErr
}
func (s *stubCObject) Update(_ *xorm.Session, _ web.Auth) error {
	s.called = "Update"
	return s.returnErr
}
func (s *stubCObject) Delete(_ *xorm.Session, _ web.Auth) error {
	s.called = "Delete"
	return s.returnErr
}

// stubTracker keeps the last instance EmptyStruct handed out, for post-dispatch inspection.
type stubTracker struct {
	last    *stubCObject
	nextErr error
}

func (s *stubTracker) empty() handler.CObject {
	o := &stubCObject{returnErr: s.nextErr}
	s.last = o
	return o
}

// The token authorizes every op on "stubs"; scope denial is covered in scope_test.go.
func newAuthedCtx(t *testing.T) context.Context {
	t.Helper()
	u := &user.User{ID: 42}
	token := &models.APIToken{
		APIPermissions: models.APIPermissions{
			"stubs": []string{"create", "read_one", "read_all", "update", "delete"},
		},
	}
	ctx := WithUser(context.Background(), u)
	return WithToken(ctx, token)
}

// installStubCRUD drives the model's CRUD methods directly so routing tests need no database.
func installStubCRUD(t *testing.T) {
	t.Helper()
	saved := crud
	crud = crudFuncs{
		doCreate: func(_ context.Context, obj handler.CObject, a web.Auth) error {
			return obj.Create(nil, a)
		},
		doReadOne: func(_ context.Context, obj handler.CObject, a web.Auth) (int, error) {
			return 0, obj.ReadOne(nil, a)
		},
		doReadAll: func(_ context.Context, obj handler.CObject, a web.Auth, search string, page, perPage int) (any, int, int64, error) {
			return obj.ReadAll(nil, a, search, page, perPage)
		},
		doUpdate: func(_ context.Context, obj handler.CObject, a web.Auth) error {
			return obj.Update(nil, a)
		},
		doDelete: func(_ context.Context, obj handler.CObject, a web.Auth) error {
			return obj.Delete(nil, a)
		},
	}
	t.Cleanup(func() { crud = saved })
}

func TestDispatchToolNotFound(t *testing.T) {
	resetRegistry(t)

	_, err := Dispatch(newAuthedCtx(t), "missing_tool", json.RawMessage(`{}`))
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrToolNotFound)
}

func TestDispatchNoUser(t *testing.T) {
	resetRegistry(t)
	tracker := &stubTracker{}
	require.NoError(t, Register(Resource{
		Name:  "stubs",
		Model: tracker.empty,
		Ops:   OpReadOne,
	}))

	// Ordering matters: the scope check runs before the user lookup.
	token := &models.APIToken{
		APIPermissions: models.APIPermissions{
			"stubs": []string{"read_one"},
		},
	}
	ctx := WithToken(context.Background(), token)

	_, err := Dispatch(ctx, "stubs_read_one", json.RawMessage(`{"id":1}`))
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNoUserInContext)
}

func TestDispatchCallsCreate(t *testing.T) {
	resetRegistry(t)
	installStubCRUD(t)
	tracker := &stubTracker{}
	require.NoError(t, Register(Resource{
		Name:  "stubs",
		Model: tracker.empty,
		Ops:   OpCreate,
	}))

	_, err := Dispatch(newAuthedCtx(t), "stubs_create", json.RawMessage(`{"title":"hello"}`))
	require.NoError(t, err)
	require.NotNil(t, tracker.last)
	assert.Equal(t, "Create", tracker.last.called)
	assert.Equal(t, "hello", tracker.last.Title)
}

func TestDispatchCallsReadOne(t *testing.T) {
	resetRegistry(t)
	installStubCRUD(t)
	tracker := &stubTracker{}
	require.NoError(t, Register(Resource{
		Name:  "stubs",
		Model: tracker.empty,
		Ops:   OpReadOne,
	}))

	out, err := Dispatch(newAuthedCtx(t), "stubs_read_one", json.RawMessage(`{"id":7}`))
	require.NoError(t, err)
	require.NotNil(t, tracker.last)
	assert.Equal(t, "ReadOne", tracker.last.called)
	assert.Equal(t, int64(7), tracker.last.ID)
	// ReadOne returns the (now-populated) model directly.
	assert.Same(t, tracker.last, out)
}

func TestDispatchCallsReadAll(t *testing.T) {
	resetRegistry(t)
	installStubCRUD(t)
	config.InitDefaultConfig()
	tracker := &stubTracker{}
	require.NoError(t, Register(Resource{
		Name:  "stubs",
		Model: tracker.empty,
		Ops:   OpReadAll,
	}))

	out, err := Dispatch(newAuthedCtx(t), "stubs_read_all", json.RawMessage(`{"search":"foo","page":2,"per_page":25}`))
	require.NoError(t, err)
	require.NotNil(t, tracker.last)
	assert.Equal(t, "ReadAll", tracker.last.called)
	// The stub's ReadAll echoes search/page/per_page back.
	env := requireReadAllResult(t, out)
	assert.Equal(t, []string{"foo"}, env.Items)
	assert.Equal(t, 2, env.Page)
	assert.Equal(t, 25, env.PerPage)
	assert.Equal(t, 2, env.ResultCount)
	assert.Equal(t, int64(25), env.TotalItems)
}

func requireReadAllResult(t *testing.T, out any) *readAllResult {
	t.Helper()
	env, ok := out.(*readAllResult)
	require.Truef(t, ok, "read_all must return an envelope, got %T", out)
	return env
}

func dispatchReadAll(t *testing.T, rawArgs string) (any, error) {
	t.Helper()
	resetRegistry(t)
	installStubCRUD(t)
	config.InitDefaultConfig()
	tracker := &stubTracker{}
	require.NoError(t, Register(Resource{
		Name:  "stubs",
		Model: tracker.empty,
		Ops:   OpReadAll,
	}))
	return Dispatch(newAuthedCtx(t), "stubs_read_all", json.RawMessage(rawArgs))
}

func TestDispatchReadAllPaginationDefaults(t *testing.T) {
	// Passing page/per_page through as zero makes the models drop the LIMIT clause.
	out, err := dispatchReadAll(t, `{}`)
	require.NoError(t, err)
	env := requireReadAllResult(t, out)
	assert.Equal(t, 1, env.Page)
	assert.Equal(t, config.ServiceMaxItemsPerPage.GetInt(), env.PerPage)
}

func TestDispatchReadAllPerPageClampedToMax(t *testing.T) {
	out, err := dispatchReadAll(t, `{"per_page":1000000}`)
	require.NoError(t, err)
	env := requireReadAllResult(t, out)
	assert.Equal(t, config.ServiceMaxItemsPerPage.GetInt(), env.PerPage)
}

func TestDispatchReadAllRejectsNegativePagination(t *testing.T) {
	_, err := dispatchReadAll(t, `{"page":-1}`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `invalid value for "page"`)

	_, err = dispatchReadAll(t, `{"per_page":-1}`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `invalid value for "per_page"`)
}

func TestDispatchGatedResourceIsNotFound(t *testing.T) {
	// A disabled resource is hidden from tools/list and find_action;
	// do_action must not be able to name it either.
	resetRegistry(t)
	installStubCRUD(t)
	tracker := &stubTracker{}
	require.NoError(t, Register(Resource{
		Name:  "stubs",
		Model: tracker.empty,
		Ops:   OpReadOne,
		Gate:  func() bool { return false },
	}))

	_, err := Dispatch(newAuthedCtx(t), "stubs_read_one", json.RawMessage(`{"id":1}`))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrToolNotFound)
	assert.Empty(t, tracker.last.called, "a gated resource must never reach its model")
}

// Schema derivation skips anonymous fields, so only "amount" becomes an argument.
type validatedStub struct {
	Amount int64 `json:"amount" valid:"range(0|10)"`
	stubCObject
}

func TestDispatchValidatesTagRules(t *testing.T) {
	// `valid:` tags are enforced by echo's validator in REST; MCP has to run
	// them itself or writes bypass them entirely.
	resetRegistry(t)
	installStubCRUD(t)
	var last *validatedStub
	require.NoError(t, Register(Resource{
		Name: "stubs",
		Model: func() handler.CObject {
			last = &validatedStub{}
			return last
		},
		Ops: OpCreate,
	}))

	_, err := Dispatch(newAuthedCtx(t), "stubs_create", json.RawMessage(`{"amount":50}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "amount")
	require.NotNil(t, last)
	assert.Empty(t, last.called, "validation must run before the model is touched")

	_, err = Dispatch(newAuthedCtx(t), "stubs_create", json.RawMessage(`{"amount":5}`))
	require.NoError(t, err)
	assert.Equal(t, "Create", last.called)
}

func TestDispatchCallsUpdate(t *testing.T) {
	resetRegistry(t)
	installStubCRUD(t)
	tracker := &stubTracker{}
	require.NoError(t, Register(Resource{
		Name:  "stubs",
		Model: tracker.empty,
		Ops:   OpUpdate,
	}))

	_, err := Dispatch(newAuthedCtx(t), "stubs_update", json.RawMessage(`{"id":3,"title":"new"}`))
	require.NoError(t, err)
	require.NotNil(t, tracker.last)
	assert.Equal(t, "Update", tracker.last.called)
	assert.Equal(t, int64(3), tracker.last.ID)
	assert.Equal(t, "new", tracker.last.Title)
}

func TestDispatchCallsDelete(t *testing.T) {
	resetRegistry(t)
	installStubCRUD(t)
	tracker := &stubTracker{}
	require.NoError(t, Register(Resource{
		Name:  "stubs",
		Model: tracker.empty,
		Ops:   OpDelete,
	}))

	_, err := Dispatch(newAuthedCtx(t), "stubs_delete", json.RawMessage(`{"id":9}`))
	require.NoError(t, err)
	require.NotNil(t, tracker.last)
	assert.Equal(t, "Delete", tracker.last.called)
	assert.Equal(t, int64(9), tracker.last.ID)
}

func TestDispatchModelErrorPropagates(t *testing.T) {
	resetRegistry(t)
	installStubCRUD(t)
	wantErr := errors.New("simulated model error")
	tracker := &stubTracker{nextErr: wantErr}
	require.NoError(t, Register(Resource{
		Name:  "stubs",
		Model: tracker.empty,
		Ops:   OpReadOne,
	}))

	_, err := Dispatch(newAuthedCtx(t), "stubs_read_one", json.RawMessage(`{"id":1}`))
	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr)
}

func TestDispatchInvalidJSON(t *testing.T) {
	resetRegistry(t)
	tracker := &stubTracker{}
	require.NoError(t, Register(Resource{
		Name:  "stubs",
		Model: tracker.empty,
		Ops:   OpReadOne,
	}))

	_, err := Dispatch(newAuthedCtx(t), "stubs_read_one", json.RawMessage(`{not json`))
	require.Error(t, err)
}

func TestDispatchUnsupportedOpForResource(t *testing.T) {
	resetRegistry(t)
	tracker := &stubTracker{}
	require.NoError(t, Register(Resource{
		Name:  "stubs",
		Model: tracker.empty,
		Ops:   OpReadOne, // only read_one is registered
	}))

	// stubs_create was never registered, so it must be tool-not-found.
	_, err := Dispatch(newAuthedCtx(t), "stubs_create", json.RawMessage(`{}`))
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrToolNotFound)
}
