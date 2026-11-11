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
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

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
