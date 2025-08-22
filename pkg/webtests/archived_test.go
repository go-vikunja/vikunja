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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This tests the following behaviour:
// 2. A project which belongs to an archived project cannot be edited.
// 3. An archived project should not be editable.
//   1. Except for un-archiving it.
// 4. It is not possible to un-archive a project individually if its parent project is archived.
// 5. Creating new child projects in an archived project should not work.
// 6. Creating new tasks on an archived project should not work.
// 7. Creating new tasks on a project whose parent project is archived should not work.
// 8. Editing tasks on an archived project should not work.
// 9. Editing tasks on a project whose parent project is archived should not work.
// 11. Archived projects should not appear in the list with all projects.
// 12. Projects whose parent project is archived should not appear in the project with all projects.
//
// All of this is tested through web tests because it's not yet clear if this will be implemented directly
// or with some kind of middleware.
//
// Maybe the inheritance of projects from parents could be solved with some kind of is_archived_inherited flag -
// that way I'd only need to implement the checking on a project level and update the flag for all projects once the
// project is archived. The archived flag would then be used to not accedentially unarchive projects which were
// already individually archived when the parent project was archived.
//
// Project 21 belongs to project 16
// Project 22 is archived individually

func TestArchived(t *testing.T) {
	th := NewTestHelper(t)
	th.Login(t, &testuser1)

	// TODO: A previous version of this test file contained a comprehensive suite of tests
	// to ensure that tasks within archived projects could not be modified in any way.
	// These tests were removed during a refactoring and need to be re-added.
	//
	// The missing tests, which should be run for both "archived parent project" and
	// "archived individually" scenarios, include:
	//   - Editing a task
	//   - Deleting a task
	//   - Adding new labels to a task
	//   - Removing labels from a task
	//   - Adding assignees to a task
	//   - Removing assignees from a task
	//   - Adding a relation to a task
	//   - Removing a relation from a task
	//   - Adding a comment to a task
	//   - Removing a comment from a task

	// The project belongs to an archived parent project
	t.Run("archived parent project", func(t *testing.T) {
		t.Run("not editable", func(t *testing.T) {
			_, err := th.Request(t, "POST", "/api/v1/projects/21", strings.NewReader(`{"title":"TestIpsum","is_archived":true}`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), `"code":403`)
		})
		t.Run("no new tasks", func(t *testing.T) {
			_, err := th.Request(t, "PUT", "/api/v1/projects/21/tasks", strings.NewReader(`{"title":"Lorem"}`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), `"code":403`)
		})
		t.Run("not unarchivable", func(t *testing.T) {
			_, err := th.Request(t, "POST", "/api/v1/projects/21", strings.NewReader(`{"title":"LoremIpsum","is_archived":false}`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), `"code":403`)
		})

	})
	// The project itself is archived
	t.Run("archived individually", func(t *testing.T) {
		t.Run("not editable", func(t *testing.T) {
			_, err := th.Request(t, "POST", "/api/v1/projects/22", strings.NewReader(`{"title":"TestIpsum","is_archived":true}`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), `"code":403`)
		})
		t.Run("no new tasks", func(t *testing.T) {
			_, err := th.Request(t, "PUT", "/api/v1/projects/22/tasks", strings.NewReader(`{"title":"Lorem"}`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), `"code":403`)
		})
		t.Run("unarchivable", func(t *testing.T) {
			resp, err := th.Request(t, "POST", "/api/v1/projects/22", strings.NewReader(`{"title":"LoremIpsum","is_archived":false}`))
			require.NoError(t, err)
			assert.Contains(t, resp.Body.String(), `"is_archived":false`)
		})
	})
}
