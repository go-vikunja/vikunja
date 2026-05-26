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

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// stubCObject is a test double for handler.CObject that records which method
// was invoked by the dispatcher. Each instance must be checked individually,
// because handler.Do* runs against a fresh EmptyStruct() per call.
type stubCObject struct {
	ID    int64 `json:"id"`
	Title string

	// called records the most recent CRUD method invoked on this instance.
	called string
	// returnErr is returned from the next CRUD method invoked. Permission
	// checks always allow access; failure scenarios are exercised by the
	// model layer in the integration tests.
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

// stubTracker tracks the *last* instance handed out by EmptyStruct so the
// test can inspect which method was invoked after the dispatcher has run.
type stubTracker struct {
	last    *stubCObject
	nextErr error
}

func (s *stubTracker) empty() handler.CObject {
	o := &stubCObject{returnErr: s.nextErr}
	s.last = o
	return o
}

// stubInput is the wrapper type used by the dispatcher tests for every op.
// In the real registry each op has its own wrapper type; for testing the
// dispatcher we only need something that unmarshal+ApplyTo work against.
type stubInput struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Search  string `json:"search,omitempty"`
	Page    int    `json:"page,omitempty"`
	PerPage int    `json:"per_page,omitempty"`
}

// ApplyTo copies wrapper fields onto the model. This is the seam Task 4 will
// fill in for real resources; for now the dispatcher tests provide their own
// implementation via the inputAdapter interface so we can verify dispatch
// without depending on the (still-absent) per-resource adapter.
func (i *stubInput) ApplyTo(dst handler.CObject) error {
	s, ok := dst.(*stubCObject)
	if !ok {
		return errors.New("stubInput: unexpected target type")
	}
	s.ID = i.ID
	s.Title = i.Title
	return nil
}

// ReadAllParams exposes the pagination fields to the dispatcher. The real
// wrappers in Task 4 follow the same shape; the dispatcher reads these
// without depending on the concrete struct.
func (i *stubInput) ReadAllParams() (string, int, int) {
	return i.Search, i.Page, i.PerPage
}

// newAuthedCtx returns a context with a test user and an API token that
// authorizes every (resource, op) on the "stubs" resource — sufficient for
// the dispatcher's wiring tests. Scope-denied scenarios are covered in
// scope_test.go with explicitly narrower tokens.
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

// installStubCRUD swaps the dispatcher's Do* function set with test doubles
// that drive the model's CRUD methods directly (no xorm session). It
// returns a teardown that restores the original handler.Do* set. Tests
// that need to verify dispatch routing without standing up the DB should
// call this at the top.
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
		Name:        "stubs",
		EmptyStruct: tracker.empty,
		Ops:         OpReadOne,
		Inputs:      map[Op]any{OpReadOne: &stubInput{}},
	}))

	// Attach an authorising token but no user — the scope check passes,
	// the user lookup inside dispatchPrepared fails. Ordering matters: the
	// scope check runs first so callers without a token never reach the
	// user check.
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
		Name:        "stubs",
		EmptyStruct: tracker.empty,
		Ops:         OpCreate,
		Inputs:      map[Op]any{OpCreate: &stubInput{}},
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
		Name:        "stubs",
		EmptyStruct: tracker.empty,
		Ops:         OpReadOne,
		Inputs:      map[Op]any{OpReadOne: &stubInput{}},
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
	tracker := &stubTracker{}
	require.NoError(t, Register(Resource{
		Name:        "stubs",
		EmptyStruct: tracker.empty,
		Ops:         OpReadAll,
		Inputs:      map[Op]any{OpReadAll: &stubInput{}},
	}))

	out, err := Dispatch(newAuthedCtx(t), "stubs_read_all", json.RawMessage(`{"search":"foo","page":2,"per_page":50}`))
	require.NoError(t, err)
	require.NotNil(t, tracker.last)
	assert.Equal(t, "ReadAll", tracker.last.called)
	// The stub's ReadAll echoes the search/page/per_page so we can confirm
	// the dispatcher threaded the wrapper's pagination fields through.
	assert.Equal(t, []string{"foo"}, out)
}

func TestDispatchCallsUpdate(t *testing.T) {
	resetRegistry(t)
	installStubCRUD(t)
	tracker := &stubTracker{}
	require.NoError(t, Register(Resource{
		Name:        "stubs",
		EmptyStruct: tracker.empty,
		Ops:         OpUpdate,
		Inputs:      map[Op]any{OpUpdate: &stubInput{}},
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
		Name:        "stubs",
		EmptyStruct: tracker.empty,
		Ops:         OpDelete,
		Inputs:      map[Op]any{OpDelete: &stubInput{}},
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
		Name:        "stubs",
		EmptyStruct: tracker.empty,
		Ops:         OpReadOne,
		Inputs:      map[Op]any{OpReadOne: &stubInput{}},
	}))

	_, err := Dispatch(newAuthedCtx(t), "stubs_read_one", json.RawMessage(`{"id":1}`))
	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr)
}

func TestDispatchInvalidJSON(t *testing.T) {
	resetRegistry(t)
	tracker := &stubTracker{}
	require.NoError(t, Register(Resource{
		Name:        "stubs",
		EmptyStruct: tracker.empty,
		Ops:         OpReadOne,
		Inputs:      map[Op]any{OpReadOne: &stubInput{}},
	}))

	_, err := Dispatch(newAuthedCtx(t), "stubs_read_one", json.RawMessage(`{not json`))
	require.Error(t, err)
}

func TestDispatchUnsupportedOpForResource(t *testing.T) {
	resetRegistry(t)
	tracker := &stubTracker{}
	require.NoError(t, Register(Resource{
		Name:        "stubs",
		EmptyStruct: tracker.empty,
		Ops:         OpReadOne, // only read_one is registered
		Inputs:      map[Op]any{OpReadOne: &stubInput{}},
	}))

	// stubs_create was never registered, so it must be tool-not-found.
	_, err := Dispatch(newAuthedCtx(t), "stubs_create", json.RawMessage(`{}`))
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrToolNotFound)
}
