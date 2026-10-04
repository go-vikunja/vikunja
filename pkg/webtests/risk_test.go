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

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type riskJSON struct {
	ID          int64   `json:"id"`
	ProjectID   int64   `json:"project_id"`
	Title       string  `json:"title"`
	Probability int     `json:"probability"`
	Impact      int     `json:"impact"`
	Score       int     `json:"score"`
	Rating      string  `json:"rating"`
	Status      string  `json:"status"`
	Resolution  string  `json:"resolution"`
	ClosedAt    *string `json:"closed_at"`
}

type riskListJSON struct {
	Items      []riskJSON `json:"items"`
	Total      int64      `json:"total"`
	TotalPages int64      `json:"total_pages"`
}

// user1 owns project 1, user13 has no access to it.
func riskCall(t *testing.T, e *echo.Echo, userID int64, method, path, body string) (int, string) {
	t.Helper()
	res := humaRequest(t, e, method, path, body, humaTokenFor(t, userByID(t, userID)), "")
	return res.Code, res.Body.String()
}

func createRiskVia(t *testing.T, e *echo.Echo, userID, projectID int64, body string) riskJSON {
	t.Helper()
	code, out := riskCall(t, e, userID, http.MethodPost, "/api/v2/projects/"+itoa(projectID)+"/risks", body)
	require.Equal(t, http.StatusCreated, code, out)
	var risk riskJSON
	require.NoError(t, json.Unmarshal([]byte(out), &risk))
	return risk
}

func TestRisksAPI(t *testing.T) {
	t.Run("create, read, update, change the status, list the history, delete", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		risk := createRiskVia(t, e, 1, 1, `{"title":"Supplier delay","probability":4,"impact":5,"category":"Schedule"}`)
		assert.Equal(t, int64(1), risk.ProjectID)
		assert.Equal(t, "open", risk.Status)
		assert.Equal(t, 20, risk.Score)
		assert.Equal(t, "critical", risk.Rating)
		path := "/api/v2/risks/" + itoa(risk.ID)

		code, out := riskCall(t, e, 1, http.MethodGet, path, "")
		require.Equal(t, http.StatusOK, code, out)
		assert.Contains(t, out, `"max_permission"`)

		code, out = riskCall(t, e, 1, http.MethodPut, path, `{"title":"Supplier delayed","probability":2,"impact":2,"status":"closed","project_id":3}`)
		require.Equal(t, http.StatusOK, code, out)
		var updated riskJSON
		require.NoError(t, json.Unmarshal([]byte(out), &updated))
		assert.Equal(t, "Supplier delayed", updated.Title)
		assert.Equal(t, "low", updated.Rating)
		assert.Equal(t, "open", updated.Status, "the status is not editable here")
		assert.Equal(t, int64(1), updated.ProjectID, "nor is the project")

		code, out = riskCall(t, e, 1, http.MethodPost, path+"/status", `{"status":"closed","note":"Delivered"}`)
		require.Equal(t, http.StatusOK, code, out)
		var closed riskJSON
		require.NoError(t, json.Unmarshal([]byte(out), &closed))
		assert.Equal(t, "closed", closed.Status)
		assert.Equal(t, "Delivered", closed.Resolution)
		assert.NotNil(t, closed.ClosedAt)

		code, out = riskCall(t, e, 1, http.MethodPost, path+"/status", `{"status":"open"}`)
		require.Equal(t, http.StatusOK, code, out)
		var reopened riskJSON
		require.NoError(t, json.Unmarshal([]byte(out), &reopened))
		assert.Equal(t, "open", reopened.Status)
		assert.Nil(t, reopened.ClosedAt)
		assert.Empty(t, reopened.Resolution)

		code, out = riskCall(t, e, 1, http.MethodGet, path+"/history", "")
		require.Equal(t, http.StatusOK, code, out)
		var history struct {
			Items []struct {
				FromStatus string `json:"from_status"`
				ToStatus   string `json:"to_status"`
				Note       string `json:"note"`
			} `json:"items"`
		}
		require.NoError(t, json.Unmarshal([]byte(out), &history))
		require.Len(t, history.Items, 3)
		assert.Equal(t, "open", history.Items[0].ToStatus, "newest first")
		assert.Equal(t, "closed", history.Items[0].FromStatus)
		assert.Equal(t, "Delivered", history.Items[1].Note)

		code, out = riskCall(t, e, 1, http.MethodDelete, path, "")
		assert.Equal(t, http.StatusNoContent, code, out)
		code, _ = riskCall(t, e, 1, http.MethodGet, path, "")
		assert.Equal(t, http.StatusNotFound, code)
	})
	t.Run("PATCH is a partial update", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		risk := createRiskVia(t, e, 1, 1, `{"title":"Keep me","probability":4,"impact":4}`)

		code, out := riskCall(t, e, 1, http.MethodPatch, "/api/v2/risks/"+itoa(risk.ID), `{"impact":1}`)

		require.Equal(t, http.StatusOK, code, out)
		var patched riskJSON
		require.NoError(t, json.Unmarshal([]byte(out), &patched))
		assert.Equal(t, "Keep me", patched.Title)
		assert.Equal(t, 4, patched.Probability)
		assert.Equal(t, 1, patched.Impact)
	})
	t.Run("lists over all projects and over one, with filters", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		createRiskVia(t, e, 1, 1, `{"title":"Critical one","probability":5,"impact":5,"category":"Budget"}`)
		low := createRiskVia(t, e, 1, 1, `{"title":"Low one","probability":1,"impact":1}`)
		_, out := riskCall(t, e, 1, http.MethodPost, "/api/v2/risks/"+itoa(low.ID)+"/status", `{"status":"closed"}`)
		require.NotEmpty(t, out)

		list := func(query string) riskListJSON {
			code, body := riskCall(t, e, 1, http.MethodGet, "/api/v2/risks"+query, "")
			require.Equal(t, http.StatusOK, code, body)
			var l riskListJSON
			require.NoError(t, json.Unmarshal([]byte(body), &l))
			return l
		}

		all := list("")
		assert.GreaterOrEqual(t, all.Total, int64(3), "the fixture risk and the two new ones")
		assert.Equal(t, "Critical one", all.Items[0].Title, "worst first")

		assert.Len(t, list("?status=closed").Items, 1)
		assert.Len(t, list("?status=closed&status=open").Items, int(all.Total))
		assert.Equal(t, "Critical one", list("?rating=critical").Items[0].Title)
		assert.Len(t, list("?rating=critical").Items, 1)
		assert.Len(t, list("?category=BUDGET").Items, 1)
		assert.Len(t, list("?q=low").Items, 1)
		assert.Len(t, list("?project_id=1&status=closed").Items, 1)
		assert.Empty(t, list("?project_id=13").Items, "a project of nobody's")
		assert.Equal(t, "Low one", list("?sort_by=score&order_by=asc&status=closed").Items[0].Title)

		perPage := list("?per_page=2&page=1")
		assert.Len(t, perPage.Items, 2)
		assert.Equal(t, all.Total, perPage.Total)

		code, body := riskCall(t, e, 1, http.MethodGet, "/api/v2/projects/1/risks?status=open", "")
		require.Equal(t, http.StatusOK, code, body)
		var perProject riskListJSON
		require.NoError(t, json.Unmarshal([]byte(body), &perProject))
		for _, item := range perProject.Items {
			assert.Equal(t, int64(1), item.ProjectID)
			assert.Equal(t, "open", item.Status)
		}
	})
	t.Run("bad filter values are refused, not answered with an empty list", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		for _, query := range []string{"?status=nonsense", "?rating=huge", "?sort_by=password", "?order_by=sideways", "?overdue=maybe"} {
			code, body := riskCall(t, e, 1, http.MethodGet, "/api/v2/risks"+query, "")
			assert.Equal(t, http.StatusUnprocessableEntity, code, "%s: %s", query, body)
		}
	})
	t.Run("validation on create", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		for name, body := range map[string]string{
			"empty title":     `{"title":""}`,
			"probability 6":   `{"title":"x","probability":6}`,
			"impact 6":        `{"title":"x","impact":6}`,
			"unknown owner":   `{"title":"x","owner_id":99999}`,
			"no title at all": `{}`,
		} {
			code, out := riskCall(t, e, 1, http.MethodPost, "/api/v2/projects/1/risks", body)
			assert.GreaterOrEqual(t, code, 400, "%s: %s", name, out)
			assert.Less(t, code, 500, "%s: %s", name, out)
		}
		code, out := riskCall(t, e, 1, http.MethodPost, "/api/v2/risks/1/status", `{"status":"finished"}`)
		assert.Equal(t, http.StatusUnprocessableEntity, code, out)
	})
	t.Run("a user without access gets nothing and can change nothing", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		risk := createRiskVia(t, e, 1, 1, `{"title":"Private"}`)
		path := "/api/v2/risks/" + itoa(risk.ID)

		code, out := riskCall(t, e, 13, http.MethodGet, path, "")
		assert.Equal(t, http.StatusForbidden, code, out)
		code, out = riskCall(t, e, 13, http.MethodPut, path, `{"title":"Hijacked","probability":3,"impact":3}`)
		assert.Equal(t, http.StatusForbidden, code, out)
		code, out = riskCall(t, e, 13, http.MethodPost, path+"/status", `{"status":"closed"}`)
		assert.Equal(t, http.StatusForbidden, code, out)
		code, out = riskCall(t, e, 13, http.MethodDelete, path, "")
		assert.Equal(t, http.StatusForbidden, code, out)
		code, out = riskCall(t, e, 13, http.MethodGet, path+"/history", "")
		assert.Equal(t, http.StatusForbidden, code, out)
		code, out = riskCall(t, e, 13, http.MethodPost, "/api/v2/projects/1/risks", `{"title":"Planted"}`)
		assert.Equal(t, http.StatusForbidden, code, out)

		code, out = riskCall(t, e, 13, http.MethodGet, "/api/v2/projects/1/risks", "")
		require.Equal(t, http.StatusOK, code, out)
		var l riskListJSON
		require.NoError(t, json.Unmarshal([]byte(out), &l))
		assert.Empty(t, l.Items)
		code, out = riskCall(t, e, 13, http.MethodGet, "/api/v2/risks", "")
		require.Equal(t, http.StatusOK, code, out)
		require.NoError(t, json.Unmarshal([]byte(out), &l))
		for _, item := range l.Items {
			assert.NotEqual(t, risk.ID, item.ID, "a risk of a project nobody shared with user13")
		}

		db.AssertExists(t, "risks", map[string]interface{}{"id": risk.ID, "title": "Private", "status": "open"}, false)
		assert.Zero(t, countTasksTitled(t, "Planted"))
	})
	t.Run("a missing risk is a 404", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)

		for _, call := range []struct{ method, path, body string }{
			{http.MethodGet, "/api/v2/risks/99999", ""},
			{http.MethodPut, "/api/v2/risks/99999", `{"title":"x","probability":3,"impact":3}`},
			{http.MethodDelete, "/api/v2/risks/99999", ""},
			{http.MethodPost, "/api/v2/risks/99999/status", `{"status":"closed"}`},
		} {
			code, out := riskCall(t, e, 1, call.method, call.path, call.body)
			assert.Equal(t, http.StatusNotFound, code, "%s %s: %s", call.method, call.path, out)
		}
	})
	t.Run("an unauthenticated caller is refused", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		res := humaRequest(t, e, http.MethodGet, "/api/v2/risks", "", "", "")
		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})
	t.Run("the routes are advertised for API tokens", func(t *testing.T) {
		_, err := setupTestEnv()
		require.NoError(t, err)

		routes := models.GetAPITokenRoutes()
		require.Contains(t, routes, "risks", "not in the catch-all group; got %v", keysOf(routes))
		assert.Equal(t, http.MethodGet, routes["risks"]["read_all"].Method)
		assert.Equal(t, http.MethodGet, routes["risks"]["read_one"].Method)
		assert.NoError(t, models.PermissionsAreValid(models.APIPermissions{"risks": {"read_all", "read_one"}}))
	})
}

func keysOf[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
