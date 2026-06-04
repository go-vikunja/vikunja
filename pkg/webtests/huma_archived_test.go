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

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaArchived ports the HTTP-observable archived-project behaviours from
// v1's TestArchived (pkg/webtests/archived_test.go) to the v2 routes:
//
//   - A project whose parent is archived cannot be edited and cannot be
//     un-archived individually (the un-archive exception only applies to the
//     self-archived project, not an archived ancestor).
//   - A self-archived project cannot be edited, but CAN be un-archived.
//   - Archiving a non-archived project works.
//
// The archived error maps to HTTP 412 Precondition Failed (domain code
// ErrCodeProjectIsArchived / 3008), unlike v2's create/update validation which
// is 422.
//
// Project 21 belongs to (archived) project 22; project 22 is archived
// individually; project 1 is a normal owned project — all owned by testuser1.
//
// v1's TestArchived also covered "no new tasks", task edits/deletes, label,
// assignee, relation and comment mutations under an archived project. Those
// are NOT ported here because the corresponding task / label-task / assignee /
// relation / comment endpoints do not exist on /api/v2 yet — there is no v2
// HTTP surface to exercise them through. They remain proven by the v1 webtest
// and by the model-level TestCheckIsArchived until those resources are ported.
func TestHumaArchived(t *testing.T) {
	// Each subtest gets a pristine handler: the shared serve() does not reload
	// fixtures per request, so the un-archive/archive mutations below must not
	// leak across subtests (mirrors huma_team_test.go's per-subtest isolation).
	handlerFor := func(u *user.User) *webHandlerTestV2 {
		return &webHandlerTestV2{user: u, basePath: "/api/v2/projects", idParam: "project", t: t}
	}

	// The project belongs to an archived parent project.
	t.Run("archived parent project", func(t *testing.T) {
		t.Run("not editable", func(t *testing.T) {
			testHandler := handlerFor(&testuser1)
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "21"}, `{"title":"TestIpsum","is_archived":true}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusPreconditionFailed, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeProjectIsArchived)
		})
		t.Run("not unarchivable", func(t *testing.T) {
			// The un-archive exception only applies to the self-archived
			// project; here the archived ancestor (22) still blocks it.
			testHandler := handlerFor(&testuser1)
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "21"}, `{"title":"LoremIpsum","is_archived":false}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusPreconditionFailed, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeProjectIsArchived)
		})
	})

	// The project itself is archived.
	t.Run("archived individually", func(t *testing.T) {
		t.Run("not editable", func(t *testing.T) {
			testHandler := handlerFor(&testuser1)
			_, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "22"}, `{"title":"TestIpsum","is_archived":true}`)
			require.Error(t, err)
			assert.Equal(t, http.StatusPreconditionFailed, getHTTPErrorCode(err))
			assertHandlerErrorCode(t, err, models.ErrCodeProjectIsArchived)
		})
		t.Run("unarchivable", func(t *testing.T) {
			testHandler := handlerFor(&testuser1)
			rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "22"}, `{"title":"LoremIpsum","is_archived":false}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"is_archived":false`)
		})
	})

	// Archiving a non-archived project should work.
	t.Run("archive non-archived project", func(t *testing.T) {
		testHandler := handlerFor(&testuser1)
		rec, err := testHandler.testUpdateWithUser(nil, map[string]string{"project": "1"}, `{"title":"Test1","is_archived":true}`)
		require.NoError(t, err)
		assert.Contains(t, rec.Body.String(), `"is_archived":true`)
	})
}
