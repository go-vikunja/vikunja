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
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

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
	testProjectHandler := webHandlerTest{
		user: &testuser1,
		strFunc: func() handler.CObject {
			return &models.Project{}
		},
		t: t,
	}
	testTaskHandler := webHandlerTest{
		user: &testuser1,
		strFunc: func() handler.CObject {
			return &models.Task{}
		},
		t: t,
	}
	testLabelHandler := webHandlerTest{
		user: &testuser1,
		strFunc: func() handler.CObject {
			return &models.LabelTask{}
		},
		t: t,
	}
	testAssigneeHandler := webHandlerTest{
		user: &testuser1,
		strFunc: func() handler.CObject {
			return &models.TaskAssginee{}
		},
		t: t,
	}
	testRelationHandler := webHandlerTest{
		user: &testuser1,
		strFunc: func() handler.CObject {
			return &models.TaskRelation{}
		},
		t: t,
	}
	testCommentHandler := webHandlerTest{
		user: &testuser1,
		strFunc: func() handler.CObject {
			return &models.TaskComment{}
		},
		t: t,
	}

	taskTests := func(taskID string, errCode int, t *testing.T) {
		t.Run("task", func(t *testing.T) {
			t.Run("edit task", func(t *testing.T) {
				_, err := testTaskHandler.testUpdateWithUser(nil, map[string]string{"projecttask": taskID}, `{"title":"TestIpsum"}`)
				require.Error(t, err)
				assertHandlerErrorCode(t, err, errCode)
			})
			t.Run("delete", func(t *testing.T) {
				_, err := testTaskHandler.testDeleteWithUser(nil, map[string]string{"projecttask": taskID})
				require.Error(t, err)
				assertHandlerErrorCode(t, err, errCode)
			})
			t.Run("add new labels", func(t *testing.T) {
				_, err := testLabelHandler.testCreateWithUser(nil, map[string]string{"projecttask": taskID}, `{"label_id":1}`)
				require.Error(t, err)
				assertHandlerErrorCode(t, err, errCode)
			})
			t.Run("remove labels", func(t *testing.T) {
				_, err := testLabelHandler.testDeleteWithUser(nil, map[string]string{"projecttask": taskID, "label": "4"})
				require.Error(t, err)
				assertHandlerErrorCode(t, err, errCode)
			})
			t.Run("add assignees", func(t *testing.T) {
				_, err := testAssigneeHandler.testCreateWithUser(nil, map[string]string{"projecttask": taskID}, `{"user_id":3}`)
				require.Error(t, err)
				assertHandlerErrorCode(t, err, errCode)
			})
			t.Run("remove assignees", func(t *testing.T) {
				_, err := testAssigneeHandler.testDeleteWithUser(nil, map[string]string{"projecttask": taskID, "user": "2"})
				require.Error(t, err)
				assertHandlerErrorCode(t, err, errCode)
			})
			t.Run("add relation", func(t *testing.T) {
				_, err := testRelationHandler.testCreateWithUser(nil, map[string]string{"task": taskID}, `{"other_task_id":1,"relation_kind":"related"}`)
				require.Error(t, err)
				assertHandlerErrorCode(t, err, errCode)
			})
			t.Run("remove relation", func(t *testing.T) {
				_, err := testRelationHandler.testDeleteWithUser(nil, map[string]string{"task": taskID}, `{"other_task_id":2,"relation_kind":"related"}`)
				require.Error(t, err)
				assertHandlerErrorCode(t, err, errCode)
			})
			t.Run("add comment", func(t *testing.T) {
				_, err := testCommentHandler.testCreateWithUser(nil, map[string]string{"task": taskID}, `{"comment":"Lorem"}`)
				require.Error(t, err)
				assertHandlerErrorCode(t, err, errCode)
			})
			t.Run("remove comment", func(t *testing.T) {
				var commentID = "15"
				if taskID == "36" {
					commentID = "16"
				}
				_, err := testCommentHandler.testDeleteWithUser(nil, map[string]string{"task": taskID, "commentid": commentID})
				require.Error(t, err)
				assertHandlerErrorCode(t, err, errCode)
			})
		})
	}

	// The project belongs to an archived parent project
	t.Run("archived parent project", func(t *testing.T) {
		t.Run("not editable", func(t *testing.T) {
			_, err := testProjectHandler.testUpdateWithUser(nil, map[string]string{"project": "21"}, `{"title":"TestIpsum","is_archived":true}`)
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeProjectIsArchived)
		})
		t.Run("no new tasks", func(t *testing.T) {
			_, err := testTaskHandler.testCreateWithUser(nil, map[string]string{"project": "21"}, `{"title":"Lorem"}`)
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeProjectIsArchived)
		})
		t.Run("not unarchivable", func(t *testing.T) {
			_, err := testProjectHandler.testUpdateWithUser(nil, map[string]string{"project": "21"}, `{"title":"LoremIpsum","is_archived":false}`)
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeProjectIsArchived)
		})

		taskTests("35", models.ErrCodeProjectIsArchived, t)
	})
	// The project itself is archived
	t.Run("archived individually", func(t *testing.T) {
		t.Run("not editable", func(t *testing.T) {
			_, err := testProjectHandler.testUpdateWithUser(nil, map[string]string{"project": "22"}, `{"title":"TestIpsum","is_archived":true}`)
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeProjectIsArchived)
		})
		t.Run("no new tasks", func(t *testing.T) {
			_, err := testTaskHandler.testCreateWithUser(nil, map[string]string{"project": "22"}, `{"title":"Lorem"}`)
			require.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeProjectIsArchived)
		})
		t.Run("unarchivable", func(t *testing.T) {
			rec, err := testProjectHandler.testUpdateWithUser(nil, map[string]string{"project": "22"}, `{"title":"LoremIpsum","is_archived":false}`)
			require.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"is_archived":false`)
		})

		taskTests("36", models.ErrCodeProjectIsArchived, t)
	})
}
