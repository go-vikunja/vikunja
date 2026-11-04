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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// GHSA-hjx8-qv73-f7cm
func TestProjectUserRemovalDeletesWebhooks(t *testing.T) {
	for api, wantStatus := range map[string]int{
		"v1": http.StatusOK,
		"v2": http.StatusNoContent,
	} {
		t.Run(api, func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)

			// user1 has write access to project 10, owned by user6.
			s := db.NewSession()
			defer s.Close()
			w := &models.Webhook{
				TargetURL:   "https://example.com/removed-user",
				Events:      []string{"task.updated"},
				ProjectID:   10,
				CreatedByID: 1,
			}
			_, err = s.Insert(w)
			require.NoError(t, err)
			require.NoError(t, s.Commit())

			rec := apiTokenReq(e, http.MethodDelete, "/api/"+api+"/projects/10/users/user1", userJWT(t, 6), "")
			assert.Equal(t, wantStatus, rec.Code, rec.Body.String())
			db.AssertMissing(t, "webhooks", map[string]interface{}{"id": w.ID})
			db.AssertExists(t, "webhooks", map[string]interface{}{"id": 3}, false)
		})
	}
}

type accessLossV2Fixture struct {
	webhook   *models.Webhook
	linkShare *models.LinkSharing
}

// user1 has write access to project 10, owned by user6.
func insertUser1DataInProject10(t *testing.T) *accessLossV2Fixture {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	f := &accessLossV2Fixture{
		webhook: &models.Webhook{
			TargetURL:   "https://example.com/removed-user",
			Events:      []string{"task.updated"},
			ProjectID:   10,
			CreatedByID: 1,
		},
		linkShare: &models.LinkSharing{
			Hash:        "access-loss-v2",
			ProjectID:   10,
			SharingType: models.SharingTypeWithoutPassword,
			SharedByID:  1,
		},
	}
	_, err := s.Insert(f.webhook)
	require.NoError(t, err)
	_, err = s.Insert(f.linkShare)
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	return f
}

func TestProjectUserRemovalCleanupOptions(t *testing.T) {
	t.Run("defaults delete webhooks and keep link shares", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		f := insertUser1DataInProject10(t)

		rec := apiTokenReq(e, http.MethodDelete, "/api/v2/projects/10/users/user1", userJWT(t, 6), "")
		assert.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
		db.AssertMissing(t, "webhooks", map[string]interface{}{"id": f.webhook.ID})
		db.AssertExists(t, "link_shares", map[string]interface{}{"id": f.linkShare.ID}, false)
	})
	t.Run("keep webhooks and revoke link shares", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		f := insertUser1DataInProject10(t)

		rec := apiTokenReq(e, http.MethodDelete, "/api/v2/projects/10/users/user1?keep_webhooks=true&revoke_link_shares=true", userJWT(t, 6), "")
		assert.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
		db.AssertExists(t, "webhooks", map[string]interface{}{"id": f.webhook.ID, "created_by_id": 6}, false)
		db.AssertMissing(t, "link_shares", map[string]interface{}{"id": f.linkShare.ID})
	})
}

func TestTeamAccessRemovalCleanupOptions(t *testing.T) {
	setup := func(t *testing.T) (*accessLossV2Fixture, *models.Team) {
		t.Helper()
		f := insertUser1DataInProject10(t)
		s := db.NewSession()
		defer s.Close()
		team := &models.Team{
			Name:        "access loss v2",
			CreatedByID: 6,
		}
		_, err := s.Insert(team)
		require.NoError(t, err)
		_, err = s.Insert(&models.TeamMember{
			TeamID: team.ID,
			UserID: 6,
			Admin:  true,
		})
		require.NoError(t, err)
		_, err = s.Insert(&models.TeamMember{
			TeamID: team.ID,
			UserID: 1,
		})
		require.NoError(t, err)
		_, err = s.Insert(&models.TeamProject{
			TeamID:     team.ID,
			ProjectID:  10,
			Permission: models.PermissionWrite,
		})
		require.NoError(t, err)
		_, err = s.Where("project_id = 10 AND user_id = 1").Delete(&models.ProjectUser{})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		return f, team
	}

	for name, path := range map[string]func(team *models.Team) string{
		"remove member":     func(team *models.Team) string { return fmt.Sprintf("/api/v2/teams/%d/members/user1", team.ID) },
		"remove team share": func(team *models.Team) string { return fmt.Sprintf("/api/v2/projects/10/teams/%d", team.ID) },
		"delete team":       func(team *models.Team) string { return fmt.Sprintf("/api/v2/teams/%d", team.ID) },
	} {
		t.Run(name, func(t *testing.T) {
			e, err := setupTestEnv()
			require.NoError(t, err)
			f, team := setup(t)

			rec := apiTokenReq(e, http.MethodDelete, path(team)+"?keep_webhooks=true&revoke_link_shares=true", userJWT(t, 6), "")
			assert.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
			db.AssertExists(t, "webhooks", map[string]interface{}{"id": f.webhook.ID, "created_by_id": 6}, false)
			db.AssertMissing(t, "link_shares", map[string]interface{}{"id": f.linkShare.ID})
		})
	}

	t.Run("a member leaving cannot keep their webhooks", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		f, team := setup(t)

		path := fmt.Sprintf("/api/v2/teams/%d/members/user1?keep_webhooks=true", team.ID)
		rec := apiTokenReq(e, http.MethodDelete, path, userJWT(t, 1), "")
		assert.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
		db.AssertMissing(t, "webhooks", map[string]interface{}{"id": f.webhook.ID})
	})
}

func TestProjectMoveCleanupOptions(t *testing.T) {
	setup := func(t *testing.T) (child *models.Project, f *accessLossV2Fixture) {
		t.Helper()
		s := db.NewSession()
		defer s.Close()
		owner := &user.User{ID: 6}
		parent := &models.Project{Title: "move parent"}
		require.NoError(t, parent.Create(s, owner))
		parentID := parent.ID
		child = &models.Project{
			Title:           "move child",
			ParentProjectID: &parentID,
		}
		require.NoError(t, child.Create(s, owner))
		_, err := s.Insert(&models.ProjectUser{
			ProjectID:  parent.ID,
			UserID:     1,
			Permission: models.PermissionWrite,
		})
		require.NoError(t, err)
		f = &accessLossV2Fixture{
			webhook: &models.Webhook{
				TargetURL:   "https://example.com/move",
				Events:      []string{"task.updated"},
				ProjectID:   child.ID,
				CreatedByID: 1,
			},
			linkShare: &models.LinkSharing{
				Hash:        "access-loss-move",
				ProjectID:   child.ID,
				SharingType: models.SharingTypeWithoutPassword,
				SharedByID:  1,
			},
		}
		_, err = s.Insert(f.webhook)
		require.NoError(t, err)
		_, err = s.Insert(f.linkShare)
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		return child, f
	}
	patch := func(t *testing.T, e *echo.Echo, child *models.Project, body string) *httptest.ResponseRecorder {
		t.Helper()
		return humaRequest(t, e, http.MethodPatch, fmt.Sprintf("/api/v2/projects/%d", child.ID), body, humaTokenFor(t, &user.User{ID: 6}), "application/merge-patch+json")
	}

	t.Run("confirm returns a conflict and changes nothing", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		child, f := setup(t)

		rec := patch(t, e, child, `{"parent_project_id":0,"confirm_access_loss":true}`)
		require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())
		var body struct {
			Code       int               `json:"code"`
			I18nParams map[string]string `json:"i18n_params"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, models.ErrCodeAccessLossNeedsConfirmation, body.Code)
		assert.Equal(t, map[string]string{"webhooks": "1", "link_shares": "1"}, body.I18nParams)
		db.AssertExists(t, "projects", map[string]interface{}{"id": child.ID, "parent_project_id": *child.ParentProjectID}, false)
		db.AssertExists(t, "webhooks", map[string]interface{}{"id": f.webhook.ID, "created_by_id": 1}, false)
	})
	t.Run("choices apply through PATCH", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		child, f := setup(t)

		rec := patch(t, e, child, `{"parent_project_id":0,"keep_webhooks":true,"revoke_link_shares":true}`)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.NotContains(t, rec.Body.String(), "keep_webhooks")
		db.AssertExists(t, "webhooks", map[string]interface{}{"id": f.webhook.ID, "created_by_id": 6}, false)
		db.AssertMissing(t, "link_shares", map[string]interface{}{"id": f.linkShare.ID})
	})
	t.Run("without options a move keeps today's defaults", func(t *testing.T) {
		e, err := setupTestEnv()
		require.NoError(t, err)
		child, f := setup(t)

		rec := patch(t, e, child, `{"parent_project_id":0}`)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		db.AssertMissing(t, "webhooks", map[string]interface{}{"id": f.webhook.ID})
		db.AssertExists(t, "link_shares", map[string]interface{}{"id": f.linkShare.ID}, false)
	})
}
