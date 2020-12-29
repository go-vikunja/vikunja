// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package integrations

import (
	"net/url"
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/web/handler"
	"github.com/stretchr/testify/assert"
)

// This tests the following behaviour:
// 1. A namespace should not be editable if it is archived.
//   1. With the exception being to un-archive it.
// 2. A list which belongs to an archived namespace cannot be edited.
// 3. An archived list should not be editable.
//   1. Except for un-archiving it.
// 4. It is not possible to un-archive a list individually if its namespace is archived.
// 5. Creating new lists on an archived namespace should not work.
// 6. Creating new tasks on an archived list should not work.
// 7. Creating new tasks on a list who's namespace is archived should not work.
// 8. Editing tasks on an archived list should not work.
// 9. Editing tasks on a list who's namespace is archived should not work.
// 10. Archived namespaces should not appear in the list with all namespaces.
// 11. Archived lists should not appear in the list with all lists.
// 12. Lists who's namespace is archived should not appear in the list with all lists.
//
// All of this is tested through integration tests because it's not yet clear if this will be implemented directly
// or with some kind of middleware.
//
// Maybe the inheritance of lists from namespaces could be solved with some kind of is_archived_inherited flag -
// that way I'd only need to implement the checking on a list level and update the flag for all lists once the
// namespace is archived. The archived flag would then be used to not accedentially unarchive lists which were
// already individually archived when the namespace was archived.
// Should still test it all though.
//
// Namespace 16 is archived
// List 21 belongs to namespace 16
// List 22 is archived individually

func TestArchived(t *testing.T) {
	testListHandler := webHandlerTest{
		user: &testuser1,
		strFunc: func() handler.CObject {
			return &models.List{}
		},
		t: t,
	}
	testNamespaceHandler := webHandlerTest{
		user: &testuser1,
		strFunc: func() handler.CObject {
			return &models.Namespace{}
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

	t.Run("namespace", func(t *testing.T) {
		t.Run("not editable", func(t *testing.T) {
			_, err := testNamespaceHandler.testUpdateWithUser(nil, map[string]string{"namespace": "16"}, `{"title":"TestIpsum","is_archived":true}`)
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeNamespaceIsArchived)
		})
		t.Run("unarchivable", func(t *testing.T) {
			rec, err := testNamespaceHandler.testUpdateWithUser(nil, map[string]string{"namespace": "16"}, `{"title":"TestIpsum","is_archived":false}`)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"is_archived":false`)
		})
		t.Run("no new lists", func(t *testing.T) {
			_, err := testListHandler.testCreateWithUser(nil, map[string]string{"namespace": "16"}, `{"title":"Lorem"}`)
			assert.Error(t, err)
			assertHandlerErrorCode(t, err, models.ErrCodeNamespaceIsArchived)
		})
		t.Run("should not appear in the list", func(t *testing.T) {
			rec, err := testNamespaceHandler.testReadAllWithUser(nil, nil)
			assert.NoError(t, err)
			assert.NotContains(t, rec.Body.String(), `"title":"Archived testnamespace16"`)
		})
		t.Run("should appear in the list if explicitly requested", func(t *testing.T) {
			rec, err := testNamespaceHandler.testReadAllWithUser(url.Values{"is_archived": []string{"true"}}, nil)
			assert.NoError(t, err)
			assert.Contains(t, rec.Body.String(), `"title":"Archived testnamespace16"`)
		})
	})

	t.Run("list", func(t *testing.T) {

		taskTests := func(taskID string, errCode int, t *testing.T) {
			t.Run("task", func(t *testing.T) {
				t.Run("edit task", func(t *testing.T) {
					_, err := testTaskHandler.testUpdateWithUser(nil, map[string]string{"listtask": taskID}, `{"title":"TestIpsum"}`)
					assert.Error(t, err)
					assertHandlerErrorCode(t, err, errCode)
				})
				t.Run("delete", func(t *testing.T) {
					_, err := testTaskHandler.testDeleteWithUser(nil, map[string]string{"listtask": taskID})
					assert.Error(t, err)
					assertHandlerErrorCode(t, err, errCode)
				})
				t.Run("add new labels", func(t *testing.T) {
					_, err := testLabelHandler.testCreateWithUser(nil, map[string]string{"listtask": taskID}, `{"label_id":1}`)
					assert.Error(t, err)
					assertHandlerErrorCode(t, err, errCode)
				})
				t.Run("remove lables", func(t *testing.T) {
					_, err := testLabelHandler.testDeleteWithUser(nil, map[string]string{"listtask": taskID, "label": "4"})
					assert.Error(t, err)
					assertHandlerErrorCode(t, err, errCode)
				})
				t.Run("add assignees", func(t *testing.T) {
					_, err := testAssigneeHandler.testCreateWithUser(nil, map[string]string{"listtask": taskID}, `{"user_id":3}`)
					assert.Error(t, err)
					assertHandlerErrorCode(t, err, errCode)
				})
				t.Run("remove assignees", func(t *testing.T) {
					_, err := testAssigneeHandler.testDeleteWithUser(nil, map[string]string{"listtask": taskID, "user": "2"})
					assert.Error(t, err)
					assertHandlerErrorCode(t, err, errCode)
				})
				t.Run("add relation", func(t *testing.T) {
					_, err := testRelationHandler.testCreateWithUser(nil, map[string]string{"task": taskID}, `{"other_task_id":1,"relation_kind":"related"}`)
					assert.Error(t, err)
					assertHandlerErrorCode(t, err, errCode)
				})
				t.Run("remove relation", func(t *testing.T) {
					_, err := testRelationHandler.testDeleteWithUser(nil, map[string]string{"task": taskID}, `{"other_task_id":2,"relation_kind":"related"}`)
					assert.Error(t, err)
					assertHandlerErrorCode(t, err, errCode)
				})
				t.Run("add comment", func(t *testing.T) {
					_, err := testCommentHandler.testCreateWithUser(nil, map[string]string{"task": taskID}, `{"comment":"Lorem"}`)
					assert.Error(t, err)
					assertHandlerErrorCode(t, err, errCode)
				})
				t.Run("remove comment", func(t *testing.T) {
					var commentID = "15"
					if taskID == "36" {
						commentID = "16"
					}
					_, err := testCommentHandler.testDeleteWithUser(nil, map[string]string{"task": taskID, "commentid": commentID})
					assert.Error(t, err)
					assertHandlerErrorCode(t, err, errCode)
				})
			})
		}

		// The list belongs to an archived namespace
		t.Run("archived namespace", func(t *testing.T) {
			t.Run("not editable", func(t *testing.T) {
				_, err := testListHandler.testUpdateWithUser(nil, map[string]string{"list": "21"}, `{"title":"TestIpsum","is_archived":true}`)
				assert.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrCodeNamespaceIsArchived)
			})
			t.Run("no new tasks", func(t *testing.T) {
				_, err := testTaskHandler.testCreateWithUser(nil, map[string]string{"list": "21"}, `{"title":"Lorem"}`)
				assert.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrCodeNamespaceIsArchived)
			})
			t.Run("not unarchivable", func(t *testing.T) {
				_, err := testListHandler.testUpdateWithUser(nil, map[string]string{"list": "21"}, `{"title":"LoremIpsum","is_archived":false}`)
				assert.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrCodeNamespaceIsArchived)
			})

			taskTests("35", models.ErrCodeNamespaceIsArchived, t)
		})
		// The list itself is archived
		t.Run("archived individually", func(t *testing.T) {
			t.Run("not editable", func(t *testing.T) {
				_, err := testListHandler.testUpdateWithUser(nil, map[string]string{"list": "22"}, `{"title":"TestIpsum","is_archived":true}`)
				assert.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrCodeListIsArchived)
			})
			t.Run("no new tasks", func(t *testing.T) {
				_, err := testTaskHandler.testCreateWithUser(nil, map[string]string{"list": "22"}, `{"title":"Lorem"}`)
				assert.Error(t, err)
				assertHandlerErrorCode(t, err, models.ErrCodeListIsArchived)
			})
			t.Run("unarchivable", func(t *testing.T) {
				rec, err := testListHandler.testUpdateWithUser(nil, map[string]string{"list": "22"}, `{"title":"LoremIpsum","is_archived":false}`)
				assert.NoError(t, err)
				assert.Contains(t, rec.Body.String(), `"is_archived":false`)
			})

			taskTests("36", models.ErrCodeListIsArchived, t)
		})
	})
}
