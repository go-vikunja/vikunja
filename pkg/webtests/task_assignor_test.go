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
	"encoding/json"
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTasksAssign_Assignor(t *testing.T) {
	t.Run("without an assignor and without a project the assignee is the assignor", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := botToken(t, models.APIPermissions{"tasks": {"assign"}})

		res := humaRequest(t, e, http.MethodPost, assignPath, `{"assignee_email":"user16@example.com","title":"T"}`, token, "")

		require.Equal(t, http.StatusCreated, res.Code, res.Body.String())
		var out struct {
			Assignor struct {
				ID int64 `json:"id"`
			} `json:"assignor"`
		}
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &out))
		assert.Equal(t, int64(16), out.Assignor.ID)
	})
	t.Run("an explicit assignor is stored as the creator, null counts as omitted", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := botToken(t, models.APIPermissions{"tasks": {"assign"}})

		res := humaRequest(t, e, http.MethodPost, assignPath, `{"assignee_email":"user16@example.com","assignor_email":"user1@example.com","title":"From user1"}`, token, "")
		require.Equal(t, http.StatusCreated, res.Code, res.Body.String())
		var out assignResponse
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &out))
		db.AssertExists(t, "tasks", map[string]interface{}{"id": out.Task.ID, "created_by_id": 1}, false)

		null := humaRequest(t, e, http.MethodPost, assignPath, `{"assignee_email":"user16@example.com","assignor_email":null,"title":"Null assignor"}`, token, "")
		assert.Equal(t, http.StatusCreated, null.Code, null.Body.String())
	})
	t.Run("an unknown or inactive assignor is refused with its own code", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := botToken(t, models.APIPermissions{"tasks": {"assign"}})

		missing := humaRequest(t, e, http.MethodPost, assignPath, `{"assignee_email":"user16@example.com","assignor_email":"nobody@example.com","title":"T"}`, token, "")
		assert.Equal(t, http.StatusNotFound, missing.Code, missing.Body.String())
		assert.Contains(t, missing.Body.String(), "4046")

		inactive := humaRequest(t, e, http.MethodPost, assignPath, `{"assignee_email":"user16@example.com","assignor_email":"user17@example.com","title":"T"}`, token, "")
		assert.Equal(t, http.StatusConflict, inactive.Code, inactive.Body.String())
		assert.Contains(t, inactive.Body.String(), "4045")
	})
}

func TestTasksList_Assignment(t *testing.T) {
	type row struct {
		Title     string
		Assignees []int64
	}
	list := func(t *testing.T, query string, u *user.User) (int, []row) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		res := humaRequest(t, e, http.MethodGet, "/api/v2/tasks"+query, "", humaTokenFor(t, u), "")
		var out struct {
			Items []struct {
				Title     string `json:"title"`
				Assignees []struct {
					ID int64 `json:"id"`
				} `json:"assignees"`
			} `json:"items"`
		}
		_ = json.Unmarshal(res.Body.Bytes(), &out)
		rows := make([]row, 0, len(out.Items))
		for _, it := range out.Items {
			r := row{Title: it.Title}
			for _, a := range it.Assignees {
				r.Assignees = append(r.Assignees, a.ID)
			}
			rows = append(rows, r)
		}
		return res.Code, rows
	}

	s := db.NewSession()
	defer s.Close()
	u, err := user.GetUserByID(s, 1)
	require.NoError(t, err)

	t.Run("mine only returns tasks assigned to the user", func(t *testing.T) {
		code, rows := list(t, "?assignment=mine&per_page=200", u)
		require.Equal(t, http.StatusOK, code)

		require.NotEmpty(t, rows, "fixture user1 has tasks assigned")
		for _, r := range rows {
			assert.Contains(t, r.Assignees, int64(1), r.Title)
		}
	})
	t.Run("assigned_by_me is accepted", func(t *testing.T) {
		code, _ := list(t, "?assignment=assigned_by_me", u)
		assert.Equal(t, http.StatusOK, code)
	})
	t.Run("an unknown value is refused before it reaches the database", func(t *testing.T) {
		code, _ := list(t, "?assignment=everyone", u)
		assert.Equal(t, http.StatusUnprocessableEntity, code)
	})
	t.Run("the project endpoints do not take it", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		res := humaRequest(t, e, http.MethodGet, "/api/v2/projects/1/tasks?assignment=mine", "", humaTokenFor(t, u), "")
		assert.Equal(t, http.StatusOK, res.Code, "an unknown query parameter is ignored there")
	})
}
