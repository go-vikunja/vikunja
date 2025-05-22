//go:build unit
// +build unit

package models

import (
	"testing"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskDuplicate_DeepCopy(t *testing.T) {
	s := setupTestSession(t)
	u := &user.User{ID: 1, Username: "testuser"}
	auth := web.NewAuthUser(u)

	// Create a root task
	root := &Task{
		Title:       "Root Task",
		Description: "Root Desc",
		ProjectID:   1,
		Assignees:   []*user.User{u},
		Labels:      []*Label{{ID: 1, Title: "Label1"}},
		HexColor:    "#ff0000",
		PercentDone: 50,
	}
	require.NoError(t, root.Create(s, auth))

	// Create a subtask
	sub := &Task{
		Title:     "Subtask",
		ProjectID: 1,
	}
	require.NoError(t, sub.Create(s, auth))

	// Relate subtask
	rel := &TaskRelation{
		TaskID:       root.ID,
		OtherTaskID:  sub.ID,
		RelationKind: RelationKindSubtask,
	}
	require.NoError(t, rel.Create(s, auth))

	// Add follows relation
	follower := &Task{
		Title:     "Follower",
		ProjectID: 1,
	}
	require.NoError(t, follower.Create(s, auth))
	followsRel := &TaskRelation{
		TaskID:       root.ID,
		OtherTaskID:  follower.ID,
		RelationKind: RelationKindFollows,
	}
	require.NoError(t, followsRel.Create(s, auth))

	// Duplicate
	dup := &TaskDuplicate{ProjectID: 1, TaskID: root.ID}
	require.NoError(t, dup.Create(s, auth))
	assert.NotZero(t, dup.Task.ID)
	assert.Equal(t, root.Title, dup.Task.Title)
	assert.Equal(t, root.Description, dup.Task.Description)
	assert.Equal(t, root.HexColor, dup.Task.HexColor)
	assert.Equal(t, root.PercentDone, dup.Task.PercentDone)
	assert.Len(t, dup.Task.Assignees, 1)
	assert.Len(t, dup.Task.Labels, 1)

	// Check subtask duplicated
	subtasks := []*TaskRelation{}
	require.NoError(t, s.Where("task_id = ? AND relation_kind = ?", dup.Task.ID, RelationKindSubtask).Find(&subtasks))
	assert.Len(t, subtasks, 1)
	var duplicatedSub Task
	require.True(t, s.ID(subtasks[0].OtherTaskID).Get(&duplicatedSub))
	assert.Equal(t, sub.Title, duplicatedSub.Title)

	// Check follows relation duplicated
	dedupFollows := []*TaskRelation{}
	require.NoError(t, s.Where("task_id = ? AND relation_kind = ?", dup.Task.ID, RelationKindFollows).Find(&dedupFollows))
	assert.Len(t, dedupFollows, 1)
}
