// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
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

package caldav

// This file tests logic related to handling tasks in CALDAV format

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a new echo.Context for testing
func newTestContext(method, path string, body string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, "text/calendar; charset=utf-8") // Common for PUT/POST
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// Check logic related to creating sub-tasks
func TestSubTask_Create(t *testing.T) {
	currentUser := &user.User{ // Renamed from u to currentUser for clarity
		ID:       15,
		Username: "user15",
		Email:    "user15@example.com",
	}

	config.InitDefaultConfig()
	log.InitLogger()
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

		const projectID = 36
		const taskUID = "uid_child1"
		var taskURL = "/dav/projects/" + strconv.FormatInt(projectID, 10) + "/" + taskUID + ".ics"
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

		c, rec := newTestContext(http.MethodPut, taskURL, taskContent)
		// Set user in context, similar to how middleware would
		c.Set("userBasicAuth", currentUser)
		c.SetParamNames("project", "task")
		c.SetParamValues(strconv.FormatInt(projectID, 10), taskUID+".ics")


		err := UpsertTaskFromICS(c, currentUser, projectID, taskUID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code) // Or StatusNoContent if update

		// To check the CalDAV content, we would now need to call FetchTaskAsICS
		fetchCtx, fetchRec := newTestContext(http.MethodGet, taskURL, "")
		fetchCtx.Set("userBasicAuth", currentUser)
		fetchCtx.SetParamNames("project", "task")
		fetchCtx.SetParamValues(strconv.FormatInt(projectID, 10), taskUID+".ics")

		err = FetchTaskAsICS(fetchCtx, currentUser, projectID, taskUID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, fetchRec.Code)
		
		content := fetchRec.Body.String()
		assert.Contains(t, content, "UID:"+taskUID)
		assert.Contains(t, content, "RELATED-TO;RELTYPE=PARENT:uid-caldav-test-parent-task")

		// Get the task from the DB:
		tasks, err := models.GetTasksByUIDs(s, []string{taskUID}, currentUser)
		require.NoError(t, err)
		require.Len(t, tasks, 1, "Task should be found in DB")
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

		const projectID = 36
		const parentTaskUID = "uid-caldav-test-parent-task" // Assuming this exists from fixtures
		const childTaskUID = "uid_child1"
		var childTaskURL = "/dav/projects/" + strconv.FormatInt(projectID, 10) + "/" + childTaskUID + ".ics"
		var childTaskContent = `BEGIN:VCALENDAR
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
RELATED-TO;RELTYPE=PARENT:` + parentTaskUID + `
END:VTODO
END:VCALENDAR`

		// Create the child task first
		cChild, recChild := newTestContext(http.MethodPut, childTaskURL, childTaskContent)
		cChild.Set("userBasicAuth", currentUser)
		cChild.SetParamNames("project", "task")
		cChild.SetParamValues(strconv.FormatInt(projectID, 10), childTaskUID+".ics")
		err := UpsertTaskFromICS(cChild, currentUser, projectID, childTaskUID)
		require.NoError(t, err)
		require.True(t, recChild.Code == http.StatusCreated || recChild.Code == http.StatusNoContent)


		const grandChildTaskUID = "uid_grand_child1"
		var grandChildTaskURL = "/dav/projects/" + strconv.FormatInt(projectID, 10) + "/" + grandChildTaskUID + ".ics"
		var grandChildTaskContent = `BEGIN:VCALENDAR
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
RELATED-TO;RELTYPE=PARENT:` + childTaskUID + `
END:VTODO
END:VCALENDAR`

		cGChild, recGChild := newTestContext(http.MethodPut, grandChildTaskURL, grandChildTaskContent)
		cGChild.Set("userBasicAuth", currentUser)
		cGChild.SetParamNames("project", "task")
		cGChild.SetParamValues(strconv.FormatInt(projectID, 10), grandChildTaskUID+".ics")
		err = UpsertTaskFromICS(cGChild, currentUser, projectID, grandChildTaskUID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, recGChild.Code)


		// Check that the result CALDAV contains the relation:
		fetchGCtx, fetchGRec := newTestContext(http.MethodGet, grandChildTaskURL, "")
		fetchGCtx.Set("userBasicAuth", currentUser)
		fetchGCtx.SetParamNames("project", "task")
		fetchGCtx.SetParamValues(strconv.FormatInt(projectID, 10), grandChildTaskUID+".ics")
		err = FetchTaskAsICS(fetchGCtx, currentUser, projectID, grandChildTaskUID)
		require.NoError(t, err)
		gContent := fetchGRec.Body.String()
		assert.Contains(t, gContent, "UID:"+grandChildTaskUID)
		assert.Contains(t, gContent, "RELATED-TO;RELTYPE=PARENT:"+childTaskUID)

		// Get the grandchild task from the DB:
		tasks, err := models.GetTasksByUIDs(s, []string{grandChildTaskUID}, currentUser)
		require.NoError(t, err)
		require.Len(t, tasks, 1)
		task := tasks[0]

		// Check that the parent-child relationship of the grandchildren is present:
		assert.Len(t, task.RelatedTasks[models.RelationKindParenttask], 1)
		parentTask := task.RelatedTasks[models.RelationKindParenttask][0]
		assert.Equal(t, childTaskUID, parentTask.UID)

		// Get the child task and check that it now has a parent and a child:
		childTasks, err := models.GetTasksByUIDs(s, []string{childTaskUID}, currentUser)
		require.NoError(t, err)
		require.Len(t, childTasks, 1)
		task = childTasks[0]
		assert.Len(t, task.RelatedTasks[models.RelationKindParenttask], 1)
		parentTask = task.RelatedTasks[models.RelationKindParenttask][0]
		assert.Equal(t, parentTaskUID, parentTask.UID)

		assert.Len(t, task.RelatedTasks[models.RelationKindSubtask], 1)
		gcTask := task.RelatedTasks[models.RelationKindSubtask][0]
		assert.Equal(t, grandChildTaskUID, gcTask.UID)
	})

	//
	// Create a subtask on a parent that we don't know anything about (yet)
	//
	t.Run("create subtask on unknown parent", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		const projectID = 36
		const taskUID = "uid_child1_unknown_parent" // Make UID unique for this test run
		const unknownParentUID = "uid-caldav-test-parent-doesnt-exist-yet"
		var taskURL = "/dav/projects/" + strconv.FormatInt(projectID, 10) + "/" + taskUID + ".ics"
		var taskContent = `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:Project 36 for Caldav tests
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:` + taskUID + `
DTSTAMP:20230301T073337Z
SUMMARY:Caldav child task 1
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
RELATED-TO;RELTYPE=PARENT:` + unknownParentUID + `
END:VTODO
END:VCALENDAR`

		c, rec := newTestContext(http.MethodPut, taskURL, taskContent)
		c.Set("userBasicAuth", currentUser)
		c.SetParamNames("project", "task")
		c.SetParamValues(strconv.FormatInt(projectID, 10), taskUID+".ics")

		err := UpsertTaskFromICS(c, currentUser, projectID, taskUID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		// Check that the result CALDAV contains the relation:
		fetchCtx, fetchRec := newTestContext(http.MethodGet, taskURL, "")
		fetchCtx.Set("userBasicAuth", currentUser)
		fetchCtx.SetParamNames("project", "task")
		fetchCtx.SetParamValues(strconv.FormatInt(projectID, 10), taskUID+".ics")
		err = FetchTaskAsICS(fetchCtx, currentUser, projectID, taskUID)
		require.NoError(t, err)
		content := fetchRec.Body.String()

		assert.Contains(t, content, "UID:"+taskUID)
		assert.Contains(t, content, "RELATED-TO;RELTYPE=PARENT:"+unknownParentUID)

		// Get the task from the DB:
		tasks, err := models.GetTasksByUIDs(s, []string{taskUID}, currentUser)
		require.NoError(t, err)
		require.Len(t, tasks, 1)
		task := tasks[0]

		// Check that the parent-child relationship is present:
		assert.Len(t, task.RelatedTasks[models.RelationKindParenttask], 1)
		parentTask := task.RelatedTasks[models.RelationKindParenttask][0]
		assert.Equal(t, unknownParentUID, parentTask.UID)

		// Check that the non-existent parent task was created in the process:
		parentTasks, err := models.GetTasksByUIDs(s, []string{unknownParentUID}, currentUser)
		require.NoError(t, err)
		require.Len(t, parentTasks, 1, "Dummy parent task should have been created")
		createdParentTask := parentTasks[0]
		assert.Equal(t, unknownParentUID, createdParentTask.UID)
		assert.Equal(t, projectID, int(createdParentTask.ProjectID)) // Should be in the same project
		assert.Equal(t, "DUMMY-UID-"+unknownParentUID, createdParentTask.Title)
	})
}

// Logic related to editing tasks and subtasks
func TestSubTask_Update(t *testing.T) {
	currentUser := &user.User{ // Renamed from u to currentUser
		ID:       15,
		Username: "user15",
		Email:    "user15@example.com",
	}
	// Init calls are already in TestSubTask_Create, if tests run in sequence they might not be needed
	// but it's safer to have them if tests can run independently.
	// config.InitDefaultConfig()
	// files.InitTests()
	// user.InitTests()
	// models.SetupTests()


	//
	// Edit a subtask and check that the relations are not gone
	//
	t.Run("edit subtask", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		const projectID = 36
		const taskUID = "uid-caldav-test-child-task" // Exists in fixtures
		const parentUID = "uid-caldav-test-parent-task" // Exists in fixtures
		var taskURL = "/dav/projects/" + strconv.FormatInt(projectID, 10) + "/" + taskUID + ".ics"
		var taskContent = `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:Project 36 for Caldav tests
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:` + taskUID + `
DTSTAMP:20230301T073337Z
SUMMARY:Child task for Caldav Test (edited)
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
RELATED-TO;RELTYPE=PARENT:` + parentUID + `
END:VTODO
END:VCALENDAR`
		// Fetch existing task to pass to Upsert (not strictly needed by Upsert but good for context)
		// tasks, err := models.GetTasksByUIDs(s, []string{taskUID}, currentUser)
		// require.NoError(t, err)
		// require.Len(t, tasks, 1)
		// existingTask := tasks[0]

		c, rec := newTestContext(http.MethodPut, taskURL, taskContent)
		c.Set("userBasicAuth", currentUser)
		c.SetParamNames("project", "task")
		c.SetParamValues(strconv.FormatInt(projectID, 10), taskUID+".ics")

		err := UpsertTaskFromICS(c, currentUser, projectID, taskUID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, rec.Code) // Update should be 204

		// Check that the result CALDAV still contains the relation:
		fetchCtx, fetchRec := newTestContext(http.MethodGet, taskURL, "")
		fetchCtx.Set("userBasicAuth", currentUser)
		fetchCtx.SetParamNames("project", "task")
		fetchCtx.SetParamValues(strconv.FormatInt(projectID, 10), taskUID+".ics")
		err = FetchTaskAsICS(fetchCtx, currentUser, projectID, taskUID)
		require.NoError(t, err)
		content := fetchRec.Body.String()

		assert.Contains(t, content, "UID:"+taskUID)
		assert.Contains(t, content, "SUMMARY:Child task for Caldav Test (edited)")
		assert.Contains(t, content, "RELATED-TO;RELTYPE=PARENT:"+parentUID)

		// Get the task from the DB:
		tasksDB, err := models.GetTasksByUIDs(s, []string{taskUID}, currentUser)
		require.NoError(t, err)
		require.Len(t, tasksDB, 1)
		taskFromDB := tasksDB[0]
		assert.Equal(t, "Child task for Caldav Test (edited)", taskFromDB.Title)


		// Check that the parent-child relationship is still present:
		assert.Len(t, taskFromDB.RelatedTasks[models.RelationKindParenttask], 1)
		parentTask := taskFromDB.RelatedTasks[models.RelationKindParenttask][0]
		assert.Equal(t, parentUID, parentTask.UID)
	})

	//
	// Edit a parent task and check that the subtasks are still linked
	//
	t.Run("edit parent", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		const projectID = 36
		const taskUID = "uid-caldav-test-parent-task" // Exists in fixtures
		const childUID1 = "uid-caldav-test-child-task" // Exists in fixtures
		const childUID2 = "uid-caldav-test-child-task-2" // Exists in fixtures
		var taskURL = "/dav/projects/" + strconv.FormatInt(projectID, 10) + "/" + taskUID + ".ics"
		var taskContent = `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:Project 36 for Caldav tests
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:` + taskUID + `
DTSTAMP:20230301T073337Z
SUMMARY:Parent task for Caldav Test (edited)
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
RELATED-TO;RELTYPE=CHILD:` + childUID1 + `
RELATED-TO;RELTYPE=CHILD:` + childUID2 + `
END:VTODO
END:VCALENDAR`

		c, rec := newTestContext(http.MethodPut, taskURL, taskContent)
		c.Set("userBasicAuth", currentUser)
		c.SetParamNames("project", "task")
		c.SetParamValues(strconv.FormatInt(projectID, 10), taskUID+".ics")

		err := UpsertTaskFromICS(c, currentUser, projectID, taskUID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, rec.Code)


		// Get the task from the DB:
		tasksDB, err := models.GetTasksByUIDs(s, []string{taskUID}, currentUser)
		require.NoError(t, err)
		require.Len(t, tasksDB, 1)
		taskFromDB := tasksDB[0]
		assert.Equal(t, "Parent task for Caldav Test (edited)", taskFromDB.Title)

		// Check that the subtasks are still linked:
		assert.Len(t, taskFromDB.RelatedTasks[models.RelationKindSubtask], 2)
		
		foundChild1 := false
		foundChild2 := false
		for _, subTask := range taskFromDB.RelatedTasks[models.RelationKindSubtask] {
			if subTask.UID == childUID1 {
				foundChild1 = true
			}
			if subTask.UID == childUID2 {
				foundChild2 = true
			}
		}
		assert.True(t, foundChild1, "Child task 1 should be linked")
		assert.True(t, foundChild2, "Child task 2 should be linked")
	})

	//
	// Edit a subtask and change its parent
	//
	t.Run("edit subtask change parent", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		const projectID = 36
		const taskUID = "uid-caldav-test-child-task" // Exists in fixtures
		const originalParentUID = "uid-caldav-test-parent-task" // Exists in fixtures
		const newParentUID = "uid-caldav-test-parent-task-2" // Exists in fixtures
		var taskURL = "/dav/projects/" + strconv.FormatInt(projectID, 10) + "/" + taskUID + ".ics"

		var taskContent = `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:Project 36 for Caldav tests
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:` + taskUID + `
DTSTAMP:20230301T073337Z
SUMMARY:Child task for Caldav Test (parent changed)
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
RELATED-TO;RELTYPE=PARENT:` + newParentUID + `
END:VTODO
END:VCALENDAR`

		c, rec := newTestContext(http.MethodPut, taskURL, taskContent)
		c.Set("userBasicAuth", currentUser)
		c.SetParamNames("project", "task")
		c.SetParamValues(strconv.FormatInt(projectID, 10), taskUID+".ics")

		err := UpsertTaskFromICS(c, currentUser, projectID, taskUID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// Check that the result CALDAV contains the new relation:
		fetchCtx, fetchRec := newTestContext(http.MethodGet, taskURL, "")
		fetchCtx.Set("userBasicAuth", currentUser)
		fetchCtx.SetParamNames("project", "task")
		fetchCtx.SetParamValues(strconv.FormatInt(projectID, 10), taskUID+".ics")
		err = FetchTaskAsICS(fetchCtx, currentUser, projectID, taskUID)
		require.NoError(t, err)
		content := fetchRec.Body.String()

		assert.Contains(t, content, "UID:"+taskUID)
		assert.Contains(t, content, "RELATED-TO;RELTYPE=PARENT:"+newParentUID)
		assert.NotContains(t, content, "RELATED-TO;RELTYPE=PARENT:"+originalParentUID)


		// Get the task from the DB:
		tasksDB, err := models.GetTasksByUIDs(s, []string{taskUID}, currentUser)
		require.NoError(t, err)
		require.Len(t, tasksDB, 1)
		taskFromDB := tasksDB[0]

		// Check that the parent-child relationship has changed to the new parent:
		assert.Len(t, taskFromDB.RelatedTasks[models.RelationKindParenttask], 1)
		parentTask := taskFromDB.RelatedTasks[models.RelationKindParenttask][0]
		assert.Equal(t, newParentUID, parentTask.UID)

		// Get the previous parent from the DB and check that its previous child is gone:
		originalParentTasks, err := models.GetTasksByUIDs(s, []string{originalParentUID}, currentUser)
		require.NoError(t, err)
		require.Len(t, originalParentTasks, 1)
		originalParentFromDB := originalParentTasks[0]
		
		foundOriginalChild := false
		for _, subTask := range originalParentFromDB.RelatedTasks[models.RelationKindSubtask] {
			if subTask.UID == taskUID {
				foundOriginalChild = true
				break
			}
		}
		assert.False(t, foundOriginalChild, "Task should no longer be a child of the original parent")
		// Check that the sibling is still there
		assert.Len(t, originalParentFromDB.RelatedTasks[models.RelationKindSubtask], 1) 
		formerSiblingSubTask := originalParentFromDB.RelatedTasks[models.RelationKindSubtask][0]
		assert.Equal(t, "uid-caldav-test-child-task-2", formerSiblingSubTask.UID)
	})

	//
	// Edit a subtask and remove its parent
	//
	t.Run("edit subtask remove parent", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		const projectID = 36
		const taskUID = "uid-caldav-test-child-task" // Exists in fixtures
		const originalParentUID = "uid-caldav-test-parent-task" // Exists in fixtures
		var taskURL = "/dav/projects/" + strconv.FormatInt(projectID, 10) + "/" + taskUID + ".ics"
		var taskContent = `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:Project 36 for Caldav tests
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:` + taskUID + `
DTSTAMP:20230301T073337Z
SUMMARY:Child task for Caldav Test (no parent)
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
END:VTODO
END:VCALENDAR`
		
		c, rec := newTestContext(http.MethodPut, taskURL, taskContent)
		c.Set("userBasicAuth", currentUser)
		c.SetParamNames("project", "task")
		c.SetParamValues(strconv.FormatInt(projectID, 10), taskUID+".ics")

		err := UpsertTaskFromICS(c, currentUser, projectID, taskUID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, rec.Code)

		// Check that the result CALDAV contains no parent relation
		fetchCtx, fetchRec := newTestContext(http.MethodGet, taskURL, "")
		fetchCtx.Set("userBasicAuth", currentUser)
		fetchCtx.SetParamNames("project", "task")
		fetchCtx.SetParamValues(strconv.FormatInt(projectID, 10), taskUID+".ics")
		err = FetchTaskAsICS(fetchCtx, currentUser, projectID, taskUID)
		require.NoError(t, err)
		content := fetchRec.Body.String()

		assert.Contains(t, content, "UID:"+taskUID)
		assert.NotContains(t, content, "RELATED-TO;RELTYPE=PARENT")

		// Get the task from the DB:
		tasksDB, err := models.GetTasksByUIDs(s, []string{taskUID}, currentUser)
		require.NoError(t, err)
		require.Len(t, tasksDB, 1)
		taskFromDB := tasksDB[0]

		// Check that the parent-child relationship is gone:
		assert.Empty(t, taskFromDB.RelatedTasks[models.RelationKindParenttask])

		// Get the previous parent from the DB and check that its child is gone:
		originalParentTasks, err := models.GetTasksByUIDs(s, []string{originalParentUID}, currentUser)
		require.NoError(t, err)
		require.Len(t, originalParentTasks, 1)
		originalParentFromDB := originalParentTasks[0]

		foundOriginalChild := false
		for _, subTask := range originalParentFromDB.RelatedTasks[models.RelationKindSubtask] {
			if subTask.UID == taskUID {
				foundOriginalChild = true
				break
			}
		}
		assert.False(t, foundOriginalChild, "Task should no longer be a child of the original parent")
		// We're gone, but our former sibling is still there:
		assert.Len(t, originalParentFromDB.RelatedTasks[models.RelationKindSubtask], 1)
		formerSiblingSubTask := originalParentFromDB.RelatedTasks[models.RelationKindSubtask][0]
		assert.Equal(t, "uid-caldav-test-child-task-2", formerSiblingSubTask.UID)
	})
}
