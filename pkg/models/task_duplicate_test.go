//go:build unit
// +build unit

package models

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskDuplicate_DeepCopy(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{ID: 1, Username: "user1"}
	auth := web.NewAuthUser(u)

	// Use an existing task from fixtures as the root
	var root Task
	require.True(t, s.ID(1).Get(&root))

	// Add a reminder to the root task (if not already present)
	reminder := &TaskReminder{
		TaskID:   root.ID,
		Reminder: root.Created.Add(3600), // 1 hour after creation
	}
	require.NoError(t, s.Insert(reminder))

	// Duplicate
	dup := &TaskDuplicate{ProjectID: root.ProjectID, TaskID: root.ID}
	require.NoError(t, dup.Create(s, auth))
	assert.NotZero(t, dup.Task.ID)
	assert.Equal(t, root.Title, dup.Task.Title)
	assert.Equal(t, root.Description, dup.Task.Description)
	assert.Equal(t, root.HexColor, dup.Task.HexColor)
	assert.Equal(t, root.PercentDone, dup.Task.PercentDone)
	assert.Len(t, dup.Task.Assignees, len(root.Assignees))
	assert.Len(t, dup.Task.Labels, len(root.Labels))

	// Check reminders are copied
	reminders := []*TaskReminder{}
	require.NoError(t, s.Where("task_id = ?", dup.Task.ID).Find(&reminders))
	assert.NotEmpty(t, reminders)
	assert.WithinDuration(t, reminder.Reminder, reminders[0].Reminder, 0)

	// Check subtask duplicated (if any)
	subtasks := []*TaskRelation{}
	require.NoError(t, s.Where("task_id = ? AND relation_kind = ?", dup.Task.ID, RelationKindSubtask).Find(&subtasks))
	if len(subtasks) > 0 {
		var duplicatedSub Task
		require.True(t, s.ID(subtasks[0].OtherTaskID).Get(&duplicatedSub))
		assert.NotEmpty(t, duplicatedSub.Title)
	}

	// Check follows relation duplicated (if any)
	dedupFollows := []*TaskRelation{}
	require.NoError(t, s.Where("task_id = ? AND relation_kind = ?", dup.Task.ID, RelationKindFollows).Find(&dedupFollows))
	// Just check that the number of follows relations is the same as the original
	origFollows := []*TaskRelation{}
	s.Where("task_id = ? AND relation_kind = ?", root.ID, RelationKindFollows).Find(&origFollows)
	assert.Equal(t, len(origFollows), len(dedupFollows))
}
