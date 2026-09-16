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

package models

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

type subtaskAccessEnv struct {
	s           *xorm.Session
	owner       *user.User
	reader      *user.User
	shared      *Project
	crossShared *Project
	private     *Project
	taskA       *Task // in shared, root of the subtask tree
}

func setupSubtaskAccessEnv(t *testing.T) *subtaskAccessEnv {
	t.Helper()
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()

	owner := &user.User{ID: 1}
	reader := &user.User{ID: 2}

	shared := &Project{Title: "subtask-access shared"}
	require.NoError(t, shared.Create(s, owner))
	_, err := s.Insert(&ProjectUser{UserID: reader.ID, ProjectID: shared.ID, Permission: PermissionWrite})
	require.NoError(t, err)

	crossShared := &Project{Title: "subtask-access cross shared"}
	require.NoError(t, crossShared.Create(s, owner))
	_, err = s.Insert(&ProjectUser{UserID: reader.ID, ProjectID: crossShared.ID, Permission: PermissionRead})
	require.NoError(t, err)

	private := &Project{Title: "subtask-access private"}
	require.NoError(t, private.Create(s, owner))

	taskA := &Task{Title: "subtask-access root", ProjectID: shared.ID}
	require.NoError(t, taskA.Create(s, owner))
	require.NoError(t, s.Commit())

	return &subtaskAccessEnv{
		s:           s,
		owner:       owner,
		reader:      reader,
		shared:      shared,
		crossShared: crossShared,
		private:     private,
		taskA:       taskA,
	}
}

func (e *subtaskAccessEnv) newTask(t *testing.T, title string, project *Project) *Task {
	t.Helper()
	task := &Task{Title: title, ProjectID: project.ID}
	require.NoError(t, task.Create(e.s, e.owner))
	return task
}

func (e *subtaskAccessEnv) insertSubtaskRelation(t *testing.T, parent, child *Task) {
	t.Helper()
	_, err := e.s.Insert(&TaskRelation{
		TaskID:       parent.ID,
		OtherTaskID:  child.ID,
		RelationKind: RelationKindSubtask,
		CreatedByID:  e.owner.ID,
	})
	require.NoError(t, err)
	_, err = e.s.Insert(&TaskRelation{
		TaskID:       child.ID,
		OtherTaskID:  parent.ID,
		RelationKind: RelationKindParenttask,
		CreatedByID:  e.owner.ID,
	})
	require.NoError(t, err)
	require.NoError(t, e.s.Commit())
}

func (e *subtaskAccessEnv) expandForReader(t *testing.T) []*Task {
	t.Helper()
	tasks, _, _, err := getRawTasksForProjects(e.s, []*Project{e.shared}, e.reader, &taskSearchOptions{
		expand: []TaskCollectionExpandable{TaskCollectionExpandSubtasks},
	})
	require.NoError(t, err)
	return tasks
}

func taskIDsOf(tasks []*Task) map[int64]bool {
	ids := make(map[int64]bool, len(tasks))
	for _, t := range tasks {
		ids[t.ID] = true
	}
	return ids
}

// Guards inaccessible subtask traversal (GHSA-3hc7-r24j-rpwc).
func TestExpandSubtasks_PrivateSubtreeExcluded(t *testing.T) {
	env := setupSubtaskAccessEnv(t)
	defer env.s.Close()

	privateSub := env.newTask(t, "private subtask", env.private)
	privateSubSub := env.newTask(t, "private sub-subtask", env.private)
	publicSub := env.newTask(t, "public subtask", env.shared)
	env.insertSubtaskRelation(t, env.taskA, privateSub)
	env.insertSubtaskRelation(t, privateSub, privateSubSub)
	env.insertSubtaskRelation(t, env.taskA, publicSub)

	tasks := env.expandForReader(t)
	ids := taskIDsOf(tasks)
	assert.Contains(t, ids, env.taskA.ID, "the root task must be returned")
	assert.Contains(t, ids, publicSub.ID, "a readable subtask must be returned")
	assert.NotContains(t, ids, privateSub.ID, "a subtask the caller cannot read must not be returned")
	assert.NotContains(t, ids, privateSubSub.ID, "the subtree behind an inaccessible task must not be returned")
}

func TestExpandSubtasks_CrossProjectChildIncluded(t *testing.T) {
	env := setupSubtaskAccessEnv(t)
	defer env.s.Close()

	crossSub := env.newTask(t, "cross-project subtask", env.crossShared)
	env.insertSubtaskRelation(t, env.taskA, crossSub)

	tasks := env.expandForReader(t)
	ids := taskIDsOf(tasks)
	assert.Contains(t, ids, env.taskA.ID)
	assert.Contains(t, ids, crossSub.ID, "a readable subtask in another project must be returned")
}

func TestExpandSubtasks_CycleTerminates(t *testing.T) {
	env := setupSubtaskAccessEnv(t)
	defer env.s.Close()

	loopB := env.newTask(t, "cycle b", env.shared)
	env.insertSubtaskRelation(t, env.taskA, loopB)
	_, err := env.s.Insert(&TaskRelation{
		TaskID:       loopB.ID,
		OtherTaskID:  env.taskA.ID,
		RelationKind: RelationKindSubtask,
		CreatedByID:  env.owner.ID,
	})
	require.NoError(t, err)
	require.NoError(t, env.s.Commit())

	tasks := env.expandForReader(t)
	counts := map[int64]int{}
	for _, t := range tasks {
		counts[t.ID]++
	}
	assert.Equal(t, 1, counts[env.taskA.ID], "a task in a relation cycle must be returned exactly once")
	assert.Equal(t, 1, counts[loopB.ID], "a task in a relation cycle must be returned exactly once")
}

func TestExpandSubtasks_FavoritedInaccessibleParentKeepsChildRoot(t *testing.T) {
	env := setupSubtaskAccessEnv(t)
	defer env.s.Close()

	favoritedParent := env.newTask(t, "favorited parent", env.private)
	child := env.newTask(t, "accessible child", env.shared)

	_, err := env.s.Insert(&TaskRelation{
		TaskID:       child.ID,
		OtherTaskID:  favoritedParent.ID,
		RelationKind: RelationKindParenttask,
		CreatedByID:  env.owner.ID,
	})
	require.NoError(t, err)
	_, err = env.s.Insert(&Favorite{
		UserID:   env.reader.ID,
		Kind:     FavoriteKindTask,
		EntityID: favoritedParent.ID,
	})
	require.NoError(t, err)
	require.NoError(t, env.s.Commit())

	plain, _, _, err := getRawTasksForProjects(env.s, []*Project{env.shared}, env.reader, &taskSearchOptions{})
	require.NoError(t, err)
	require.Contains(t, taskIDsOf(plain), child.ID)

	expanded, _, _, err := getRawTasksForProjects(env.s, []*Project{env.shared}, env.reader, &taskSearchOptions{
		expand: []TaskCollectionExpandable{TaskCollectionExpandSubtasks},
	})
	require.NoError(t, err)
	assert.Contains(t, taskIDsOf(expanded), child.ID,
		"an accessible child of an inaccessible favorited parent must remain a root and be returned")
	assert.NotContains(t, taskIDsOf(expanded), favoritedParent.ID,
		"the inaccessible favorited parent itself must never be returned")
}
