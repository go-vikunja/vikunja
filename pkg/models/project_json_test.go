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

package models

import (
	"encoding/json"
	"testing"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

// TestProjectParentProjectIDIsAlwaysSerialized covers the Go-built projects that
// never touch the projects table, so nothing else would set ParentProjectID.
func TestProjectParentProjectIDIsAlwaysSerialized(t *testing.T) {
	t.Run("favorites pseudo project", func(t *testing.T) {
		b, err := json.Marshal(&FavoritesPseudoProject)
		require.NoError(t, err)
		assert.Contains(t, string(b), `"parent_project_id":0`)
	})
	t.Run("saved filter pseudo project", func(t *testing.T) {
		b, err := json.Marshal((&SavedFilter{ID: 71, Title: "my filter"}).ToProject())
		require.NoError(t, err)
		assert.Contains(t, string(b), `"parent_project_id":0`)
	})
}

// TestProjectAfterLoadNormalizesNullParent covers the other source of a nil
// parent: parent_project_id is a nullable column and rows predating it were
// never backfilled, so those used to serialize without the field at all.
func TestProjectAfterLoadNormalizesNullParent(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	_, err := s.Exec("UPDATE projects SET parent_project_id = NULL WHERE id = 1")
	require.NoError(t, err)

	project, exists, err := getProjectSimple(s, builder.Eq{"id": 1})
	require.NoError(t, err)
	require.True(t, exists)
	require.NotNil(t, project.ParentProjectID)
	assert.Equal(t, int64(0), *project.ParentProjectID)

	b, err := json.Marshal(project)
	require.NoError(t, err)
	assert.Contains(t, string(b), `"parent_project_id":0`)
}

// TestProjectUnmarshalJSONKeepsNilParent guards the request-side half of
// GHSA-44v6-7fxq-vgf4: an omitted parent_project_id must stay distinguishable
// from an explicit 0, which is what Update gates the Admin check on.
func TestProjectUnmarshalJSONKeepsNilParent(t *testing.T) {
	t.Run("omitted", func(t *testing.T) {
		var p Project
		require.NoError(t, json.Unmarshal([]byte(`{"title":"x"}`), &p))
		assert.Nil(t, p.ParentProjectID)
	})
	t.Run("explicit 0", func(t *testing.T) {
		var p Project
		require.NoError(t, json.Unmarshal([]byte(`{"title":"x","parent_project_id":0}`), &p))
		require.NotNil(t, p.ParentProjectID)
		assert.Equal(t, int64(0), *p.ParentProjectID)
	})
}
