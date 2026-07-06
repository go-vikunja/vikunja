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
	"encoding/json"
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenAuthorizes_PermissionPresent(t *testing.T) {
	token := &models.APIToken{
		APIPermissions: models.APIPermissions{
			"projects": []string{"read_one", "read_all"},
		},
	}

	r := &Resource{Name: "projects"}

	assert.True(t, tokenAuthorizes(token, r.Name, OpReadOne))
	assert.True(t, tokenAuthorizes(token, r.Name, OpReadAll))
}

func TestTokenAuthorizes_PermissionAbsent(t *testing.T) {
	token := &models.APIToken{
		APIPermissions: models.APIPermissions{
			"projects": []string{"read_one"},
		},
	}

	r := &Resource{Name: "projects"}

	assert.False(t, tokenAuthorizes(token, r.Name, OpCreate))
	assert.False(t, tokenAuthorizes(token, r.Name, OpUpdate))
	assert.False(t, tokenAuthorizes(token, r.Name, OpDelete))
}

func TestTokenAuthorizes_NoGroup(t *testing.T) {
	token := &models.APIToken{
		APIPermissions: models.APIPermissions{
			"mcp": []string{"access"},
		},
	}

	assert.False(t, tokenAuthorizes(token, "projects", OpReadOne))
	assert.False(t, tokenAuthorizes(token, "projects", OpCreate))
}

func TestTokenAuthorizes_NilPermissionsMap(t *testing.T) {
	// A token with nil APIPermissions should never authorize anything.
	token := &models.APIToken{APIPermissions: nil}

	assert.False(t, tokenAuthorizes(token, "projects", OpReadOne))
}

func TestTokenAuthorizes_NilToken(t *testing.T) {
	// Defensive: a nil token (should never happen in practice because the
	// entry handler always sets one) must not panic.
	assert.False(t, tokenAuthorizes(nil, "projects", OpReadOne))
}

func TestTokenAuthorizes_FullScopes(t *testing.T) {
	token := &models.APIToken{
		APIPermissions: models.APIPermissions{
			"projects": []string{"create", "read_one", "read_all", "update", "delete"},
		},
	}

	for _, op := range AllOps() {
		assert.Truef(t, tokenAuthorizes(token, "projects", op), "op %s should be authorized", op.ToolSuffix())
	}
}

func TestDispatchScopeDenied(t *testing.T) {
	resetRegistry(t)
	installStubCRUD(t)
	tracker := &stubTracker{}
	require.NoError(t, Register(Resource{
		Name:  "stubs",
		Model: tracker.empty,
		Ops:   OpCreate | OpReadOne,
	}))

	// Token has read_one but not create.
	token := &models.APIToken{
		APIPermissions: models.APIPermissions{
			"stubs": []string{"read_one"},
		},
	}
	ctx := WithToken(newAuthedCtx(t), token)

	_, err := Dispatch(ctx, "stubs_create", json.RawMessage(`{"title":"x"}`))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrScopeDenied)
	// The denied call must not have invoked Do*. (Register reflects the
	// model type once at registration time, so an instance exists — what
	// matters is that no CRUD method ran on it.)
	assert.Empty(t, tracker.last.called, "Do* must not run for a denied scope")
}

func TestDispatchScopeDenied_NoTokenInContext(t *testing.T) {
	// Without a token in context, the scope check has nothing to authorize
	// against. The dispatcher should treat a missing token as denied
	// (defensive — the entry handler always sets one in production).
	resetRegistry(t)
	installStubCRUD(t)
	tracker := &stubTracker{}
	require.NoError(t, Register(Resource{
		Name:  "stubs",
		Model: tracker.empty,
		Ops:   OpReadOne,
	}))

	// User in context but no token — the scope check must still deny.
	u := &user.User{ID: 42}
	ctx := WithUser(t.Context(), u)
	_, err := Dispatch(ctx, "stubs_read_one", json.RawMessage(`{"id":1}`))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrScopeDenied)
	assert.Empty(t, tracker.last.called)
}

func TestDispatchDeleteScopeDenied(t *testing.T) {
	// Delete is the most destructive op; make sure the scope check gates it
	// like every other op.
	resetRegistry(t)
	installStubCRUD(t)
	tracker := &stubTracker{}
	require.NoError(t, Register(Resource{
		Name:  "stubs",
		Model: tracker.empty,
		Ops:   OpDelete,
	}))

	token := &models.APIToken{
		APIPermissions: models.APIPermissions{
			"stubs": []string{"read_one"}, // delete not allowed
		},
	}
	ctx := WithToken(newAuthedCtx(t), token)

	_, err := Dispatch(ctx, "stubs_delete", json.RawMessage(`{"id":1}`))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrScopeDenied)
	assert.Empty(t, tracker.last.called)
}

func TestDispatchScopeAllowed(t *testing.T) {
	// Positive control: with the right scope, dispatch reaches the stub.
	resetRegistry(t)
	installStubCRUD(t)
	tracker := &stubTracker{}
	require.NoError(t, Register(Resource{
		Name:  "stubs",
		Model: tracker.empty,
		Ops:   OpReadOne,
	}))

	token := &models.APIToken{
		APIPermissions: models.APIPermissions{
			"stubs": []string{"read_one"},
		},
	}
	ctx := WithToken(newAuthedCtx(t), token)

	_, err := Dispatch(ctx, "stubs_read_one", json.RawMessage(`{"id":1}`))
	require.NoError(t, err)
	require.NotNil(t, tracker.last)
	assert.Equal(t, "ReadOne", tracker.last.called)
}
