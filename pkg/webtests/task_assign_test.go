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
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const assignPath = "/api/v2/tasks/assign"

// botToken creates an API token for fixture bot 23 (owned by user 21) with the given permissions.
// The owner is made an instance admin: only bots of admins may act in the name of other people.
func botToken(t *testing.T, permissions models.APIPermissions) string {
	t.Helper()
	return botTokenForOwner(t, true, permissions)
}

func botTokenForOwner(t *testing.T, ownerIsAdmin bool, permissions models.APIPermissions) string {
	t.Helper()
	s := db.NewSession()
	defer s.Close()

	_, err := s.ID(21).Cols("is_admin").Update(&user.User{IsAdmin: ownerIsAdmin})
	require.NoError(t, err)

	owner, err := user.GetUserByID(s, 21)
	require.NoError(t, err)
	token := &models.APIToken{
		Title:          "integration",
		APIPermissions: permissions,
		ExpiresAt:      time.Now().Add(time.Hour),
		OwnerID:        23,
	}
	require.NoError(t, token.Create(s, owner))
	require.NoError(t, s.Commit())
	return token.Token
}

type assignResponse struct {
	Created bool `json:"created"`
	Task    struct {
		ID        int64  `json:"id"`
		ProjectID int64  `json:"project_id"`
		Title     string `json:"title"`
	} `json:"task"`
	Assignee struct {
		ID    int64  `json:"id"`
		Email string `json:"email"`
	} `json:"assignee"`
}

func postAssign(t *testing.T, e *echo.Echo, token, body string) (int, string, assignResponse) {
	t.Helper()
	res := humaRequest(t, e, http.MethodPost, assignPath, body, token, "")
	var out assignResponse
	_ = json.Unmarshal(res.Body.Bytes(), &out)
	return res.Code, res.Body.String(), out
}

func TestTasksAssign(t *testing.T) {
	t.Run("the permission is advertised as tasks.assign", func(t *testing.T) {
		_, err := setupTestEnv()
		require.NoError(t, err)

		routes := models.GetAPITokenRoutes()
		require.Contains(t, routes, "tasks")
		detail := routes["tasks"]["assign"]
		require.NotNil(t, detail, "tasks.assign must exist; got groups %v", routes["tasks"])
		assert.Equal(t, http.MethodPost, detail.Method)
		assert.Equal(t, assignPath, detail.Path)
		require.NoError(t, models.PermissionsAreValid(models.APIPermissions{"tasks": {"assign"}}))
	})
	t.Run("a bot token with the permission creates a task in the assignee's inbox", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := botToken(t, models.APIPermissions{"tasks": {"assign"}})

		code, body, out := postAssign(t, e, token, `{"assignee_email":"USER16@example.com","title":"Approve invoice 4711","priority":2,"external_id":"INV-4711"}`)

		require.Equal(t, http.StatusCreated, code, body)
		assert.True(t, out.Created)
		assert.Equal(t, "Approve invoice 4711", out.Task.Title)
		assert.Equal(t, int64(16), out.Assignee.ID)
		assert.Empty(t, out.Assignee.Email, "the email is not echoed back")
		assert.NotContains(t, body, "user16@example.com")
		db.AssertExists(t, "tasks", map[string]interface{}{"id": out.Task.ID, "created_by_id": 23, "project_id": out.Task.ProjectID}, false)
		db.AssertExists(t, "task_assignees", map[string]interface{}{"task_id": out.Task.ID, "user_id": 16}, false)
	})
	t.Run("a retry with the same external_id returns 200 and the same task", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := botToken(t, models.APIPermissions{"tasks": {"assign"}})
		payload := `{"assignee_email":"user16@example.com","title":"Once","external_id":"REF-1"}`

		_, _, first := postAssign(t, e, token, payload)
		code, body, second := postAssign(t, e, token, payload)

		require.Equal(t, http.StatusOK, code, body)
		assert.False(t, second.Created)
		assert.Equal(t, first.Task.ID, second.Task.ID)
	})
	t.Run("errors carry the documented status and code", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := botToken(t, models.APIPermissions{"tasks": {"assign"}})

		for name, tc := range map[string]struct {
			body   string
			status int
			code   string
		}{
			"unknown email":    {`{"assignee_email":"nobody@example.com","title":"T"}`, http.StatusNotFound, "4040"},
			"disabled user":    {`{"assignee_email":"user17@example.com","title":"T"}`, http.StatusConflict, "4041"},
			"no usable inbox":  {`{"assignee_email":"user2@example.com","title":"T"}`, http.StatusConflict, "4043"},
			"empty title":      {`{"assignee_email":"user16@example.com","title":""}`, http.StatusUnprocessableEntity, ""},
			"missing email":    {`{"title":"T"}`, http.StatusUnprocessableEntity, ""},
			"bad priority":     {`{"assignee_email":"user16@example.com","title":"T","priority":9}`, http.StatusUnprocessableEntity, ""},
			"unknown project":  {`{"assignee_email":"user16@example.com","title":"T","project_id":99999}`, http.StatusNotFound, ""},
			"project no write": {`{"assignee_email":"user16@example.com","title":"T","project_id":1}`, http.StatusForbidden, ""},
		} {
			code, body, _ := postAssign(t, e, token, tc.body)
			assert.Equal(t, tc.status, code, "%s: %s", name, body)
			if tc.code != "" {
				assert.Contains(t, body, tc.code, name)
			}
		}
	})
	t.Run("a bot token without the permission is refused", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := botToken(t, models.APIPermissions{"tasks": {"read_all"}})

		code, body, _ := postAssign(t, e, token, `{"assignee_email":"user16@example.com","title":"T"}`)

		assert.Equal(t, http.StatusUnauthorized, code, body)
		assert.Zero(t, countTasksTitled(t, "T"))
	})
	t.Run("the token of a bot an ordinary user owns is refused", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := botTokenForOwner(t, false, models.APIPermissions{"tasks": {"assign"}})

		code, body, _ := postAssign(t, e, token, `{"assignee_email":"user16@example.com","title":"Spam"}`)

		assert.Equal(t, http.StatusForbidden, code, body)
		assert.Zero(t, countTasksTitled(t, "Spam"))
	})
	t.Run("a human user's token with the permission is refused", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		s := db.NewSession()
		human, err := user.GetUserByID(s, 1)
		require.NoError(t, err)
		token := &models.APIToken{Title: "human", APIPermissions: models.APIPermissions{"tasks": {"assign"}}, ExpiresAt: time.Now().Add(time.Hour)}
		require.NoError(t, token.Create(s, human))
		require.NoError(t, s.Commit())
		s.Close()

		code, body, _ := postAssign(t, e, token.Token, `{"assignee_email":"user16@example.com","title":"Human"}`)

		assert.Equal(t, http.StatusForbidden, code, body)
		assert.Zero(t, countTasksTitled(t, "Human"))
	})
	t.Run("a logged-in human and an anonymous caller are refused", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		s := db.NewSession()
		human, err := user.GetUserByID(s, 1)
		require.NoError(t, err)
		s.Close()

		code, body, _ := postAssign(t, e, humaTokenFor(t, human), `{"assignee_email":"user16@example.com","title":"Jwt"}`)
		assert.Equal(t, http.StatusForbidden, code, body)

		anon, _, _ := postAssign(t, e, "", `{"assignee_email":"user16@example.com","title":"Anon"}`)
		assert.Equal(t, http.StatusUnauthorized, anon)
		assert.Zero(t, countTasksTitled(t, "Jwt")+countTasksTitled(t, "Anon"))
	})
	t.Run("the endpoint does not shadow reading a task by id", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		s := db.NewSession()
		u, err := user.GetUserByID(s, 1)
		require.NoError(t, err)
		s.Close()

		res := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/1", "", humaTokenFor(t, u), "")

		assert.Equal(t, http.StatusOK, res.Code, res.Body.String())
	})
	t.Run("the bot cannot read what it created", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		token := botToken(t, models.APIPermissions{"tasks": {"assign", "read_one", "delete"}})
		_, _, out := postAssign(t, e, token, `{"assignee_email":"user16@example.com","title":"Secret"}`)
		require.NotZero(t, out.Task.ID)

		read := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/"+itoa(out.Task.ID), "", token, "")
		assert.GreaterOrEqual(t, read.Code, 400, "the bot has no access to the inbox: %s", read.Body.String())
		del := humaRequest(t, e, http.MethodDelete, "/api/v2/tasks/"+itoa(out.Task.ID), "", token, "")
		assert.GreaterOrEqual(t, del.Code, 400, del.Body.String())
		db.AssertExists(t, "tasks", map[string]interface{}{"id": out.Task.ID}, false)
	})
}

func countTasksTitled(t *testing.T, title string) int64 {
	t.Helper()
	s := db.NewReadSession()
	defer s.Close()
	n, err := s.Where("title = ?", title).Count(&models.Task{})
	require.NoError(t, err)
	return n
}

func itoa(n int64) string {
	b, _ := json.Marshal(n)
	return string(b)
}
