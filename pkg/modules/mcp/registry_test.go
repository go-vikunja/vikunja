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
	"testing"

	"code.vikunja.io/api/pkg/web/handler"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetRegistry clears the package-level registry so each test starts from
// a clean slate. Tests that mutate the registry should call this at the top.
func resetRegistry(t *testing.T) {
	t.Helper()
	registryMu.Lock()
	defer registryMu.Unlock()
	resources = nil
	toolIndex = map[string]toolRef{}
}

func TestOpPermission(t *testing.T) {
	cases := map[Op]string{
		OpCreate:  "create",
		OpReadOne: "read_one",
		OpReadAll: "read_all",
		OpUpdate:  "update",
		OpDelete:  "delete",
	}
	for op, want := range cases {
		assert.Equalf(t, want, op.Permission(), "Permission() for op %d", op)
	}
}

func TestOpToolSuffix(t *testing.T) {
	cases := map[Op]string{
		OpCreate:  "create",
		OpReadOne: "read_one",
		OpReadAll: "read_all",
		OpUpdate:  "update",
		OpDelete:  "delete",
	}
	for op, want := range cases {
		assert.Equalf(t, want, op.ToolSuffix(), "ToolSuffix() for op %d", op)
	}
}

func TestOpUnknownPermission(t *testing.T) {
	// Combined bitmasks and zero values have no defined permission string.
	assert.Empty(t, Op(0).Permission())
	assert.Empty(t, (OpCreate | OpReadOne).Permission())
}

func TestRegisterAppends(t *testing.T) {
	resetRegistry(t)

	r := Resource{
		Name:        "stubs",
		Description: "test resource",
		EmptyStruct: func() handler.CObject { return &stubCObject{} },
		Ops:         OpCreate | OpReadOne,
		Inputs: map[Op]any{
			OpCreate:  &struct{}{},
			OpReadOne: &struct{}{},
		},
	}
	require.NoError(t, Register(r))

	got, ok := lookupResource("stubs")
	require.True(t, ok)
	assert.Equal(t, "stubs", got.Name)
}

func TestRegisterDuplicateName(t *testing.T) {
	resetRegistry(t)

	r := Resource{
		Name:        "stubs",
		EmptyStruct: func() handler.CObject { return &stubCObject{} },
		Ops:         OpReadOne,
		Inputs:      map[Op]any{OpReadOne: &struct{}{}},
	}
	require.NoError(t, Register(r))
	err := Register(r)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestRegisterMissingInputForOp(t *testing.T) {
	resetRegistry(t)

	r := Resource{
		Name:        "stubs",
		EmptyStruct: func() handler.CObject { return &stubCObject{} },
		Ops:         OpCreate | OpReadOne,
		// Missing input wrapper for OpReadOne.
		Inputs: map[Op]any{OpCreate: &struct{}{}},
	}
	err := Register(r)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "input")
}

func TestRegisterEmptyName(t *testing.T) {
	resetRegistry(t)

	err := Register(Resource{
		EmptyStruct: func() handler.CObject { return &stubCObject{} },
		Ops:         OpReadOne,
		Inputs:      map[Op]any{OpReadOne: &struct{}{}},
	})
	require.Error(t, err)
}

func TestRegisterRequiresEmptyStruct(t *testing.T) {
	resetRegistry(t)

	err := Register(Resource{
		Name:   "stubs",
		Ops:    OpReadOne,
		Inputs: map[Op]any{OpReadOne: &struct{}{}},
	})
	require.Error(t, err)
}

func TestToolNameResolver(t *testing.T) {
	resetRegistry(t)

	require.NoError(t, Register(Resource{
		Name:        "projects",
		EmptyStruct: func() handler.CObject { return &stubCObject{} },
		Ops:         OpCreate | OpReadOne | OpReadAll | OpUpdate | OpDelete,
		Inputs: map[Op]any{
			OpCreate:  &struct{}{},
			OpReadOne: &struct{}{},
			OpReadAll: &struct{}{},
			OpUpdate:  &struct{}{},
			OpDelete:  &struct{}{},
		},
	}))

	require.NoError(t, Register(Resource{
		Name:        "task_comments",
		EmptyStruct: func() handler.CObject { return &stubCObject{} },
		Ops:         OpReadAll,
		Inputs:      map[Op]any{OpReadAll: &struct{}{}},
	}))

	tests := []struct {
		toolName string
		resource string
		op       Op
	}{
		{"projects_create", "projects", OpCreate},
		{"projects_read_one", "projects", OpReadOne},
		{"projects_read_all", "projects", OpReadAll},
		{"projects_update", "projects", OpUpdate},
		{"projects_delete", "projects", OpDelete},
		{"task_comments_read_all", "task_comments", OpReadAll},
	}
	for _, tc := range tests {
		ref, ok := lookupTool(tc.toolName)
		require.Truef(t, ok, "tool %s should be resolvable", tc.toolName)
		assert.Equal(t, tc.resource, ref.resource.Name, "tool %s", tc.toolName)
		assert.Equal(t, tc.op, ref.op, "tool %s", tc.toolName)
	}

	_, ok := lookupTool("nonexistent_tool")
	assert.False(t, ok)

	// `task_comments_read_all` must resolve to (task_comments, read_all),
	// not to (task, comments_read_all) or any naive underscore split.
	ref, ok := lookupTool("task_comments_read_all")
	require.True(t, ok)
	assert.Equal(t, "task_comments", ref.resource.Name)
	assert.Equal(t, OpReadAll, ref.op)
}

func TestRegisterOnlyExposesEnabledOps(t *testing.T) {
	resetRegistry(t)

	require.NoError(t, Register(Resource{
		Name:        "stubs",
		EmptyStruct: func() handler.CObject { return &stubCObject{} },
		Ops:         OpReadOne | OpReadAll,
		Inputs: map[Op]any{
			OpReadOne: &struct{}{},
			OpReadAll: &struct{}{},
		},
	}))

	_, ok := lookupTool("stubs_read_one")
	assert.True(t, ok)
	_, ok = lookupTool("stubs_read_all")
	assert.True(t, ok)

	// Ops that weren't enabled in the bitmask must not appear.
	_, ok = lookupTool("stubs_create")
	assert.False(t, ok)
	_, ok = lookupTool("stubs_delete")
	assert.False(t, ok)
}

func TestAllOps(t *testing.T) {
	// AllOps must enumerate exactly the five supported ops so the registry
	// and the dispatcher walk the same list.
	want := []Op{OpCreate, OpReadOne, OpReadAll, OpUpdate, OpDelete}
	assert.Equal(t, want, AllOps())
}
