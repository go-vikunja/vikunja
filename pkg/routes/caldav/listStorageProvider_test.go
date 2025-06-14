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

package caldav

// This file tests logic related to handling tasks in CALDAV format

import (
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/samedi/caldav-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Check logic related to creating sub-tasks
func TestSubTask_Create(t *testing.T) {
	u := &user.User{
		ID:       15,
		Username: "user15",
		Email:    "user15@example.com",
	}

	config.InitDefaultConfig()
	files.InitTests()
	user.InitTests()
	models.SetupTests()

	//
	// Create a subtask
	//
	t.Run("create", func(t *testing.T) {

		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		const taskUID = "uid_child1"
		const taskContent = `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:Project 36 for Caldav tests
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:uid_child1
DTSTAMP:20230301T073337Z
SUMMARY:Caldav child task 1
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
RELATED-TO;RELTYPE=PARENT:uid-caldav-test-parent-task
END:VTODO
END:VCALENDAR`

		storage := &VikunjaCaldavProjectStorage{
			project: &models.ProjectWithTasksAndBuckets{Project: models.Project{ID: 36}},
			task:    &models.Task{UID: taskUID},
			user:    u,
		}

		// Create the subtask:
		taskResource, err := storage.CreateResource(taskUID, taskContent)
		require.NoError(t, err)

		// Check that the result CALDAV contains the relation:
		content, _ := taskResource.GetContentData()
		assert.Contains(t, content, "UID:"+taskUID)
		assert.Contains(t, content, "RELATED-TO;RELTYPE=PARENT:uid-caldav-test-parent-task")

		// Get the task from the DB:
		tasks, err := models.GetTasksByUIDs(s, []string{taskUID}, u)
		require.NoError(t, err)
		task := tasks[0]

		// Check that the parent-child relationship is present:
		assert.Len(t, task.RelatedTasks[models.RelationKindParenttask], 1)
		parentTask := task.RelatedTasks[models.RelationKindParenttask][0]
		assert.Equal(t, "uid-caldav-test-parent-task", parentTask.UID)
	})

	//
	// Create a subtask on a subtask, i.e. create a grand-child
	//
	t.Run("create grandchild on child task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		const taskUIDChild = "uid_child1"
		const taskContentChild = `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:Project 36 for Caldav tests
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:uid_child1
DTSTAMP:20230301T073337Z
SUMMARY:Caldav child task 1
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
RELATED-TO;RELTYPE=PARENT:uid-caldav-test-parent-task
END:VTODO
END:VCALENDAR`

		storage := &VikunjaCaldavProjectStorage{
			project: &models.ProjectWithTasksAndBuckets{Project: models.Project{ID: 36}},
			task:    &models.Task{UID: taskUIDChild},
			user:    u,
		}

		// Create the subtask:
		_, err := storage.CreateResource(taskUIDChild, taskContentChild)
		require.NoError(t, err)

		const taskUID = "uid_grand_child1"
		const taskContent = `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:Project 36 for Caldav tests
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:uid_grand_child1
DTSTAMP:20230301T073337Z
SUMMARY:Caldav grand child task 1
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
RELATED-TO;RELTYPE=PARENT:uid_child1
END:VTODO
END:VCALENDAR`

		storage = &VikunjaCaldavProjectStorage{
			project: &models.ProjectWithTasksAndBuckets{Project: models.Project{ID: 36}},
			task:    &models.Task{UID: taskUID},
			user:    u,
		}

		// Create the task:
		var taskResource *data.Resource
		taskResource, err = storage.CreateResource(taskUID, taskContent)
		require.NoError(t, err)

		// Check that the result CALDAV contains the relation:
		content, _ := taskResource.GetContentData()
		assert.Contains(t, content, "UID:"+taskUID)
		assert.Contains(t, content, "RELATED-TO;RELTYPE=PARENT:uid_child1")

		// Get the task from the DB:
		tasks, err := models.GetTasksByUIDs(s, []string{taskUID}, u)
		require.NoError(t, err)
		task := tasks[0]

		// Check that the parent-child relationship of the grandchildren is present:
		assert.Len(t, task.RelatedTasks[models.RelationKindParenttask], 1)
		parentTask := task.RelatedTasks[models.RelationKindParenttask][0]
		assert.Equal(t, "uid_child1", parentTask.UID)

		// Get the child task and check that it now has a parent and a child:
		tasks, err = models.GetTasksByUIDs(s, []string{"uid_child1"}, u)
		require.NoError(t, err)
		task = tasks[0]
		assert.Len(t, task.RelatedTasks[models.RelationKindParenttask], 1)
		parentTask = task.RelatedTasks[models.RelationKindParenttask][0]
		assert.Equal(t, "uid-caldav-test-parent-task", parentTask.UID)
		assert.Len(t, task.RelatedTasks[models.RelationKindSubtask], 1)
		childTask := task.RelatedTasks[models.RelationKindSubtask][0]
		assert.Equal(t, taskUID, childTask.UID)
	})

	//
	// Create a subtask on a parent that we don't know anything about (yet)
	//
	t.Run("create subtask on unknown parent", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Create a subtask:
		const taskUID = "uid_child1"
		const taskContent = `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:Project 36 for Caldav tests
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:uid_child1
DTSTAMP:20230301T073337Z
SUMMARY:Caldav child task 1
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
RELATED-TO;RELTYPE=PARENT:uid-caldav-test-parent-doesnt-exist-yet
END:VTODO
END:VCALENDAR`

		storage := &VikunjaCaldavProjectStorage{
			project: &models.ProjectWithTasksAndBuckets{Project: models.Project{ID: 36}},
			task:    &models.Task{UID: taskUID},
			user:    u,
		}

		// Create the task:
		taskResource, err := storage.CreateResource(taskUID, taskContent)
		require.NoError(t, err)

		// Check that the result CALDAV contains the relation:
		content, _ := taskResource.GetContentData()
		assert.Contains(t, content, "UID:"+taskUID)
		assert.Contains(t, content, "RELATED-TO;RELTYPE=PARENT:uid-caldav-test-parent-doesnt-exist-yet")

		// Get the task from the DB:
		tasks, err := models.GetTasksByUIDs(s, []string{taskUID}, u)
		require.NoError(t, err)
		task := tasks[0]

		// Check that the parent-child relationship is present:
		assert.Len(t, task.RelatedTasks[models.RelationKindParenttask], 1)
		parentTask := task.RelatedTasks[models.RelationKindParenttask][0]
		assert.Equal(t, "uid-caldav-test-parent-doesnt-exist-yet", parentTask.UID)

		// Check that the non-existent parent task was created in the process:
		tasks, err = models.GetTasksByUIDs(s, []string{"uid-caldav-test-parent-doesnt-exist-yet"}, u)
		require.NoError(t, err)
		task = tasks[0]
		assert.Equal(t, "uid-caldav-test-parent-doesnt-exist-yet", task.UID)
	})
}

// Logic related to editing tasks and subtasks
func TestSubTask_Update(t *testing.T) {
	u := &user.User{
		ID:       15,
		Username: "user15",
		Email:    "user15@example.com",
	}

	//
	// Edit a subtask and check that the relations are not gone
	//
	t.Run("edit subtask", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Edit the subtask:
		const taskUID = "uid-caldav-test-child-task"
		const taskContent = `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:Project 36 for Caldav tests
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:uid-caldav-test-child-task
DTSTAMP:20230301T073337Z
SUMMARY:Child task for Caldav Test (edited)
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
RELATED-TO;RELTYPE=PARENT:uid-caldav-test-parent-task
END:VTODO
END:VCALENDAR`
		tasks, err := models.GetTasksByUIDs(s, []string{taskUID}, u)
		require.NoError(t, err)
		task := tasks[0]
		storage := &VikunjaCaldavProjectStorage{
			project: &models.ProjectWithTasksAndBuckets{Project: models.Project{ID: 36}},
			task:    task,
			user:    u,
		}

		// Edit the task:
		taskResource, err := storage.UpdateResource(taskUID, taskContent)
		require.NoError(t, err)

		// Check that the result CALDAV still contains the relation:
		content, _ := taskResource.GetContentData()
		assert.Contains(t, content, "UID:"+taskUID)
		assert.Contains(t, content, "RELATED-TO;RELTYPE=PARENT:uid-caldav-test-parent-task")

		// Get the task from the DB:
		tasks, err = models.GetTasksByUIDs(s, []string{taskUID}, u)
		require.NoError(t, err)
		task = tasks[0]

		// Check that the parent-child relationship is still present:
		assert.Len(t, task.RelatedTasks[models.RelationKindParenttask], 1)
		parentTask := task.RelatedTasks[models.RelationKindParenttask][0]
		assert.Equal(t, "uid-caldav-test-parent-task", parentTask.UID)
	})

	//
	// Edit a parent task and check that the subtasks are still linked
	//
	t.Run("edit parent", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Edit the parent task:
		const taskUID = "uid-caldav-test-parent-task"
		const taskContent = `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:Project 36 for Caldav tests
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:uid-caldav-test-parent-task
DTSTAMP:20230301T073337Z
SUMMARY:Parent task for Caldav Test (edited)
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
RELATED-TO;RELTYPE=CHILD:uid-caldav-test-child-task
RELATED-TO;RELTYPE=CHILD:uid-caldav-test-child-task-2
END:VTODO
END:VCALENDAR`
		tasks, err := models.GetTasksByUIDs(s, []string{taskUID}, u)
		require.NoError(t, err)
		task := tasks[0]
		storage := &VikunjaCaldavProjectStorage{
			project: &models.ProjectWithTasksAndBuckets{Project: models.Project{ID: 36}},
			task:    task,
			user:    u,
		}

		// Edit the task:
		_, err = storage.UpdateResource(taskUID, taskContent)
		require.NoError(t, err)

		// Get the task from the DB:
		tasks, err = models.GetTasksByUIDs(s, []string{taskUID}, u)
		require.NoError(t, err)
		task = tasks[0]

		// Check that the subtasks are still linked:
		assert.Len(t, task.RelatedTasks[models.RelationKindSubtask], 2)
		existingSubTask := task.RelatedTasks[models.RelationKindSubtask][0]
		assert.Equal(t, "uid-caldav-test-child-task", existingSubTask.UID)
		existingSubTask = task.RelatedTasks[models.RelationKindSubtask][1]
		assert.Equal(t, "uid-caldav-test-child-task-2", existingSubTask.UID)
	})

	//
	// Edit a subtask and change its parent
	//
	t.Run("edit subtask change parent", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Edit the subtask:
		const taskUID = "uid-caldav-test-child-task"
		const taskContent = `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:Project 36 for Caldav tests
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:uid-caldav-test-child-task
DTSTAMP:20230301T073337Z
SUMMARY:Child task for Caldav Test (edited)
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
RELATED-TO;RELTYPE=PARENT:uid-caldav-test-parent-task-2
END:VTODO
END:VCALENDAR`
		tasks, err := models.GetTasksByUIDs(s, []string{taskUID}, u)
		require.NoError(t, err)
		task := tasks[0]
		storage := &VikunjaCaldavProjectStorage{
			project: &models.ProjectWithTasksAndBuckets{Project: models.Project{ID: 36}},
			task:    task,
			user:    u,
		}

		// Edit the task:
		taskResource, err := storage.UpdateResource(taskUID, taskContent)
		require.NoError(t, err)

		// Check that the result CALDAV contains the new relation:
		content, _ := taskResource.GetContentData()
		assert.Contains(t, content, "UID:"+taskUID)
		assert.Contains(t, content, "RELATED-TO;RELTYPE=PARENT:uid-caldav-test-parent-task-2")

		// Get the task from the DB:
		tasks, err = models.GetTasksByUIDs(s, []string{taskUID}, u)
		require.NoError(t, err)
		task = tasks[0]

		// Check that the parent-child relationship has changed to the new parent:
		assert.Len(t, task.RelatedTasks[models.RelationKindParenttask], 1)
		parentTask := task.RelatedTasks[models.RelationKindParenttask][0]
		assert.Equal(t, "uid-caldav-test-parent-task-2", parentTask.UID)

		// Get the previous parent from the DB and check that its previous child is gone:
		tasks, err = models.GetTasksByUIDs(s, []string{"uid-caldav-test-parent-task"}, u)
		require.NoError(t, err)
		task = tasks[0]
		assert.Len(t, task.RelatedTasks[models.RelationKindSubtask], 1)
		// We're gone, but our former sibling is still there:
		formerSiblingSubTask := task.RelatedTasks[models.RelationKindSubtask][0]
		assert.Equal(t, "uid-caldav-test-child-task-2", formerSiblingSubTask.UID)
	})

	//
	// Edit a subtask and remove its parent
	//
	t.Run("edit subtask remove parent", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Edit the subtask:
		const taskUID = "uid-caldav-test-child-task"
		const taskContent = `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:Project 36 for Caldav tests
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:uid-caldav-test-child-task
DTSTAMP:20230301T073337Z
SUMMARY:Child task for Caldav Test (edited)
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
END:VTODO
END:VCALENDAR`
		tasks, err := models.GetTasksByUIDs(s, []string{taskUID}, u)
		require.NoError(t, err)
		task := tasks[0]
		storage := &VikunjaCaldavProjectStorage{
			project: &models.ProjectWithTasksAndBuckets{Project: models.Project{ID: 36}},
			task:    task,
			user:    u,
		}

		// Edit the task:
		taskResource, err := storage.UpdateResource(taskUID, taskContent)
		require.NoError(t, err)

		// Check that the result CALDAV contains the new relation:
		content, _ := taskResource.GetContentData()
		assert.Contains(t, content, "UID:"+taskUID)
		assert.NotContains(t, content, "RELATED-TO;RELTYPE=PARENT:uid-caldav-test-parent-task")

		// Get the task from the DB:
		tasks, err = models.GetTasksByUIDs(s, []string{taskUID}, u)
		require.NoError(t, err)
		task = tasks[0]

		// Check that the parent-child relationship is gone:
		assert.Empty(t, task.RelatedTasks[models.RelationKindParenttask])

		// Get the previous parent from the DB and check that its child is gone:
		tasks, err = models.GetTasksByUIDs(s, []string{"uid-caldav-test-parent-task"}, u)
		require.NoError(t, err)
		task = tasks[0]
		// We're gone, but our former sibling is still there:
		assert.Len(t, task.RelatedTasks[models.RelationKindSubtask], 1)
		formerSiblingSubTask := task.RelatedTasks[models.RelationKindSubtask][0]
		assert.Equal(t, "uid-caldav-test-child-task-2", formerSiblingSubTask.UID)
	})
}
