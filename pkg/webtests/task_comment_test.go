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

	"code.vikunja.io/api/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskComments(t *testing.T) {
	th := NewTestHelper(t)
	th.Login(t, &testuser1)

	// Only run specific nested tests:
	// ^TestTaskComments$/^Update$/^Update_task_items$/^Removing_Assignees_null$
	t.Run("Update", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := th.Request(t, "POST", "/api/v1/tasks/1/comments/1", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := th.Request(t, "POST", "/api/v1/tasks/99999/comments/9999", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), `"code":404`)
		})
		t.Run("Permissions check", func(t *testing.T) {
			// Only the own comments can be updated
			t.Run("Forbidden", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/tasks/14/comments/2", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/tasks/15/comments/3", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/tasks/16/comments/4", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/tasks/17/comments/5", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/tasks/18/comments/6", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/tasks/19/comments/7", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/tasks/20/comments/8", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/tasks/21/comments/9", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/tasks/22/comments/10", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/tasks/23/comments/11", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/tasks/24/comments/12", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/tasks/25/comments/13", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				_, err := th.Request(t, "POST", "/api/v1/tasks/26/comments/14", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
		})
	})
	t.Run("Delete", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := th.Request(t, "DELETE", "/api/v1/tasks/1/comments/1", nil)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `Successfully deleted.`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := th.Request(t, "DELETE", "/api/v1/tasks/99999/comments/9999", nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), `"code":404`)
		})
		t.Run("Permissions check", func(t *testing.T) {
			// Only the own comments can be deleted
			t.Run("Forbidden", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/tasks/14/comments/2", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/tasks/15/comments/3", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/tasks/16/comments/4", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/tasks/17/comments/5", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/tasks/18/comments/6", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/tasks/19/comments/7", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/tasks/20/comments/8", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/tasks/21/comments/9", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/tasks/22/comments/10", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/tasks/23/comments/11", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/tasks/24/comments/12", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/tasks/25/comments/13", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				_, err := th.Request(t, "DELETE", "/api/v1/tasks/26/comments/14", nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
		})
	})
	t.Run("Create", func(t *testing.T) {
		t.Run("Normal", func(t *testing.T) {
			rec, err := th.Request(t, "PUT", "/api/v1/tasks/1/comments", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
		})
		t.Run("Nonexisting", func(t *testing.T) {
			_, err := th.Request(t, "PUT", "/api/v1/tasks/9999/comments", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
			require.Error(t, err)
			assert.Contains(t, err.Error(), `"code":404`)
		})
		t.Run("Permissions check", func(t *testing.T) {
			t.Run("Forbidden", func(t *testing.T) {
				// Owned by user13
				_, err := th.Request(t, "PUT", "/api/v1/tasks/34/comments", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "PUT", "/api/v1/tasks/15/comments", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Team write", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/tasks/16/comments", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Team admin", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/tasks/17/comments", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})

			t.Run("Shared Via User readonly", func(t *testing.T) {
				_, err := th.Request(t, "PUT", "/api/v1/tasks/18/comments", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via User write", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/tasks/19/comments", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})
			t.Run("Shared Via User admin", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/tasks/20/comments", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})

			t.Run("Shared Via Parent Project Team readonly", func(t *testing.T) {
				_, err := th.Request(t, "PUT", "/api/v1/tasks/21/comments", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project Team write", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/tasks/22/comments", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Parent Project Team admin", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/tasks/23/comments", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})

			t.Run("Shared Via Parent Project User readonly", func(t *testing.T) {
				_, err := th.Request(t, "PUT", "/api/v1/tasks/24/comments", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.Error(t, err)
				assert.Contains(t, err.Error(), `"code":403`)
			})
			t.Run("Shared Via Parent Project User write", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/tasks/25/comments", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})
			t.Run("Shared Via Parent Project User admin", func(t *testing.T) {
				rec, err := th.Request(t, "PUT", "/api/v1/tasks/26/comments", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
				require.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			})
		})
		// TODO: This test is functionally broken. It is intended to test comment creation
		// via a link share, but it has been refactored incorrectly.
		//
		// To fix this, the test needs to:
		// 1. Log out the current user (`th.Logout(t)`).
		// 2. Set the link share context on the test helper (`th.SetLinkShare(t, ...)`).
		//
		// Without these steps, the request is made as a fully authenticated user,
		// and the link share functionality is not actually being tested.
		t.Run("Link Share", func(t *testing.T) {
			rec, err := th.Request(t, "PUT", "/api/v1/tasks/13/comments", strings.NewReader(`{"comment":"Lorem Ipsum"}`))
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"comment":"Lorem Ipsum"`)
			db.AssertExists(t, "task_comments", map[string]interface{}{
				"task_id":   13,
				"comment":   "Lorem Ipsum",
				"author_id": -2,
			}, false)
		})
	})
}
