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
	"context"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

// createAs creates a task through the normal path as the given user.
func createAs(t *testing.T, as *user.User, task *Task) (*Task, error) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	err := task.Create(s, as)
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, err
	}
	require.NoError(t, s.Commit())
	events.DispatchPending(context.Background(), s)
	return task, nil
}

func assigneeIDs(t *testing.T, taskID int64) []int64 {
	t.Helper()
	s := db.NewReadSession()
	defer s.Close()
	var rows []*TaskAssginee
	require.NoError(t, s.Where("task_id = ?", taskID).Find(&rows))
	ids := make([]int64, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.UserID)
	}
	return ids
}

func TestAutoAssignOwnInbox(t *testing.T) {
	t.Run("a task created in the own inbox without assignee is assigned to the creator", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")

		task, err := createAs(t, jane, &Task{Title: "Mine", ProjectID: jane.DefaultProjectID})

		require.NoError(t, err)
		assert.Equal(t, []int64{jane.ID}, assigneeIDs(t, task.ID))
		db.AssertExists(t, "tasks", map[string]interface{}{"id": task.ID, "created_by_id": jane.ID}, false)
	})
	t.Run("a task created in somebody's inbox in their name is not auto-assigned, so it notifies nobody", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		boss := registerPerson(t, "boss", "boss@corp.com")
		jane := registerPerson(t, "jane", "jane@corp.com")
		makeAdmin(t, boss.ID, true)

		s := db.NewSession()
		defer s.Close()
		task := &Task{Title: "In her name", ProjectID: jane.DefaultProjectID, assignorOverride: jane}
		require.NoError(t, createTask(s, task, boss, true, true))
		require.NoError(t, s.Commit())

		assert.Empty(t, assigneeIDs(t, task.ID))
	})
	t.Run("an explicit assignee is kept and nobody is added", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")

		task, err := createAs(t, jane, &Task{Title: "Mine", ProjectID: jane.DefaultProjectID, Assignees: []*user.User{jane}})

		require.NoError(t, err)
		assert.Equal(t, []int64{jane.ID}, assigneeIDs(t, task.ID))
	})
	t.Run("a task in a shared project stays unassigned", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		joe := registerPerson(t, "joe", "joe@corp.com")
		other := createProjectFor(t, joe, "Shared")
		shareProject(t, other.ID, jane.ID, PermissionWrite)

		task, err := createAs(t, jane, &Task{Title: "Shared task", ProjectID: other.ID})

		require.NoError(t, err)
		assert.Empty(t, assigneeIDs(t, task.ID))
	})
	t.Run("the creator's second project is not an inbox", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		second := createProjectFor(t, jane, "Second")

		task, err := createAs(t, jane, &Task{Title: "Elsewhere", ProjectID: second.ID})

		require.NoError(t, err)
		assert.Empty(t, assigneeIDs(t, task.ID))
	})
	t.Run("the assignee event names the creator as doer, which the listener skips", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		events.ClearDispatchedEvents()

		_, err := createAs(t, jane, &Task{Title: "Mine", ProjectID: jane.DefaultProjectID})

		require.NoError(t, err)
		evts := events.GetDispatchedEvents((&TaskAssigneeCreatedEvent{}).Name())
		require.Len(t, evts, 1)
		e := evts[0].(*TaskAssigneeCreatedEvent)
		assert.Equal(t, jane.ID, e.Doer.ID)
		assert.Equal(t, jane.ID, e.Assignee.ID)
	})
}

func TestSetAssignor(t *testing.T) {
	t.Run("a bot may create a task in the name of somebody else", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		joe := registerPerson(t, "joe", "joe@corp.com")
		events.ClearDispatchedEvents()

		res, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", AssignorEmail: "joe@corp.com", Title: "From Joe"})

		require.NoError(t, err)
		assert.Equal(t, joe.ID, res.Assignor.ID)
		db.AssertExists(t, "tasks", map[string]interface{}{"id": res.Task.ID, "created_by_id": joe.ID}, false)
		db.AssertExists(t, "task_assignees", map[string]interface{}{"task_id": res.Task.ID, "user_id": jane.ID}, false)
		created := events.GetDispatchedEvents((&TaskCreatedEvent{}).Name())
		require.Len(t, created, 1)
		assert.Equal(t, bot.ID, created[0].(*TaskCreatedEvent).Doer.ID, "the bot is still the doer")
		// The assignor follows what they handed out.
		db.AssertExists(t, "subscriptions", map[string]interface{}{"entity_id": res.Task.ID, "user_id": joe.ID}, false)
	})
	t.Run("a normal user may not name another assignor", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		joe := registerPerson(t, "joe", "joe@corp.com")
		before := countRows(t, &Task{}, builder.Expr("1 = 1"))

		s := db.NewSession()
		defer s.Close()
		task := &Task{Title: "Forged", ProjectID: jane.DefaultProjectID, assignorOverride: joe}
		err := createTask(s, task, jane, true, true)

		require.Error(t, err)
		assert.IsType(t, ErrGenericForbidden{}, err)
		_ = s.Rollback()
		assert.Equal(t, before, countRows(t, &Task{}, builder.Expr("1 = 1")))
	})
	t.Run("an instance admin may", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		joe := registerPerson(t, "joe", "joe@corp.com")
		s := db.NewSession()
		_, err := s.ID(jane.ID).Cols("is_admin").Update(&user.User{IsAdmin: true})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		s = db.NewSession()
		defer s.Close()
		task := &Task{Title: "By admin", ProjectID: jane.DefaultProjectID, assignorOverride: joe}
		require.NoError(t, createTask(s, task, jane, true, true))
		require.NoError(t, s.Commit())

		db.AssertExists(t, "tasks", map[string]interface{}{"id": task.ID, "created_by_id": joe.ID}, false)
	})
	t.Run("a bot of an ordinary user may not name another assignor", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		ordinary := botByID(t, testBotID) // its owner is no admin
		joe := registerPerson(t, "joe", "joe@corp.com")
		jane := registerPerson(t, "jane", "jane@corp.com")
		shareProject(t, jane.DefaultProjectID, ordinary.ID, PermissionWrite)

		s := db.NewSession()
		defer s.Close()
		task := &Task{Title: "Forged", ProjectID: jane.DefaultProjectID, assignorOverride: joe}
		err := createTask(s, task, ordinary, true, true)

		require.Error(t, err)
		assert.IsType(t, ErrGenericForbidden{}, err)
	})
	t.Run("naming yourself is always fine", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")

		s := db.NewSession()
		defer s.Close()
		task := &Task{Title: "Self", ProjectID: jane.DefaultProjectID, assignorOverride: jane}
		require.NoError(t, createTask(s, task, jane, true, true))
	})
}

func TestAssignTaskByEmail_AssignorDefaults(t *testing.T) {
	t.Run("without a project the assignee is the assignor", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		jane := registerPerson(t, "jane", "jane@corp.com")

		res, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "T"})

		require.NoError(t, err)
		assert.Equal(t, jane.ID, res.Assignor.ID)
		db.AssertExists(t, "tasks", map[string]interface{}{"id": res.Task.ID, "created_by_id": jane.ID}, false)
	})
	t.Run("with a project the owner of the project is the assignor, and null is the same as omitted", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		joe := registerPerson(t, "joe", "joe@corp.com")
		project := createProjectFor(t, joe, "Joe's project")
		shareProject(t, project.ID, bot.ID, PermissionWrite)
		shareProject(t, project.ID, jane.ID, PermissionRead)

		res, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "T", ProjectID: project.ID, AssignorEmail: ""})

		require.NoError(t, err)
		assert.Equal(t, joe.ID, res.Assignor.ID)
		db.AssertExists(t, "tasks", map[string]interface{}{"id": res.Task.ID, "created_by_id": joe.ID, "project_id": project.ID}, false)
	})
	t.Run("an explicit assignor wins over the project owner", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		joe := registerPerson(t, "joe", "joe@corp.com")
		ann := registerPerson(t, "ann", "ann@corp.com")
		project := createProjectFor(t, joe, "Joe's project")
		shareProject(t, project.ID, bot.ID, PermissionWrite)
		shareProject(t, project.ID, jane.ID, PermissionRead)

		res, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", AssignorEmail: "ann@corp.com", Title: "T", ProjectID: project.ID})

		require.NoError(t, err)
		assert.Equal(t, ann.ID, res.Assignor.ID, "the assignor needs no access to the project")
	})
	t.Run("an unresolvable assignor is refused and nothing is created", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		registerPerson(t, "jane", "jane@corp.com")
		before := countRows(t, &Task{}, builder.Expr("1 = 1"))

		_, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", AssignorEmail: "nobody@corp.com", Title: "T"})
		require.Error(t, err)
		assert.True(t, IsErrAssignorNotFound(err), "%v", err)

		_, err = assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", AssignorEmail: "user17@example.com", Title: "T"})
		require.Error(t, err)
		assert.True(t, IsErrAssignorNotActive(err), "%v", err)

		assert.Equal(t, before, countRows(t, &Task{}, builder.Expr("1 = 1")))
	})
	t.Run("a bot that owns the project cannot be the default assignor", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		project := createProjectFor(t, bot, "Bot project")
		shareProject(t, project.ID, jane.ID, PermissionRead)

		_, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "T", ProjectID: project.ID})

		require.Error(t, err)
		assert.True(t, IsErrAssignorNotActive(err), "%v", err)
	})
	t.Run("a retry returns the original assignor", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		registerPerson(t, "jane", "jane@corp.com")
		joe := registerPerson(t, "joe", "joe@corp.com")
		registerPerson(t, "ann", "ann@corp.com")
		first, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", AssignorEmail: "joe@corp.com", Title: "T", ExternalID: "X"})
		require.NoError(t, err)

		again, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", AssignorEmail: "ann@corp.com", Title: "T", ExternalID: "X"})

		require.NoError(t, err)
		assert.False(t, again.Created)
		assert.Equal(t, first.Task.ID, again.Task.ID)
		assert.Equal(t, joe.ID, again.Assignor.ID)
	})
}

func TestAssignmentFilter(t *testing.T) {
	readAll := func(t *testing.T, as *user.User, f TaskAssignmentFilter, filter string) []*Task {
		t.Helper()
		s := db.NewReadSession()
		defer s.Close()
		tc := &TaskCollection{Filter: filter}
		require.NoError(t, tc.SetAssignmentFilter(f))
		res, _, _, err := tc.ReadAll(s, as, "", 1, 100)
		require.NoError(t, err)
		tasks, ok := res.([]*Task)
		require.True(t, ok)
		return tasks
	}
	titles := func(tasks []*Task) []string {
		out := make([]string, 0, len(tasks))
		for _, t := range tasks {
			out = append(out, t.Title)
		}
		return out
	}

	setup := func(t *testing.T) (jane, joe *user.User) {
		db.LoadAndAssertFixtures(t)
		jane = registerPerson(t, "jane", "jane@corp.com")
		joe = registerPerson(t, "joe", "joe@corp.com")
		shared := createProjectFor(t, joe, "Shared")
		shareProject(t, shared.ID, jane.ID, PermissionWrite)

		// jane: own inbox (auto-assigned), assigned to her in a shared project, one she gave to joe, one unassigned in the shared project
		mustCreate(t, jane, &Task{Title: "J inbox", ProjectID: jane.DefaultProjectID})
		mustCreate(t, joe, &Task{Title: "Joe to Jane", ProjectID: shared.ID, Assignees: []*user.User{jane}})
		mustCreate(t, jane, &Task{Title: "Jane to Joe", ProjectID: shared.ID, Assignees: []*user.User{joe}})
		mustCreate(t, jane, &Task{Title: "Unassigned", ProjectID: shared.ID})
		mustCreate(t, jane, &Task{Title: "Jane to both", ProjectID: shared.ID, Assignees: []*user.User{jane, joe}})
		mustCreate(t, joe, &Task{Title: "Joe inbox", ProjectID: joe.DefaultProjectID})
		return jane, joe
	}

	t.Run("mine: only tasks the user is an assignee of, across the inbox and shared projects", func(t *testing.T) {
		jane, _ := setup(t)
		got := titles(readAll(t, jane, TaskAssignmentMine, ""))
		assert.ElementsMatch(t, []string{"J inbox", "Joe to Jane", "Jane to both"}, got)
	})
	t.Run("assigned by me: created by the user and assigned to somebody else, even next to themselves", func(t *testing.T) {
		jane, _ := setup(t)
		got := titles(readAll(t, jane, TaskAssignmentAssignedByMe, ""))
		assert.ElementsMatch(t, []string{"Jane to Joe", "Jane to both"}, got)
	})
	t.Run("self-only and unassigned tasks are not assigned-by-me", func(t *testing.T) {
		jane, _ := setup(t)
		got := titles(readAll(t, jane, TaskAssignmentAssignedByMe, ""))
		assert.NotContains(t, got, "J inbox")
		assert.NotContains(t, got, "Unassigned")
	})
	t.Run("the lists never show tasks of projects the user cannot read", func(t *testing.T) {
		jane, joe := setup(t)
		// Joe assigns a task to jane in his private inbox: she cannot read it, so she cannot see it.
		s := db.NewSession()
		_, err := s.Insert(&TaskAssginee{TaskID: taskIDByTitle(t, "Joe inbox"), UserID: jane.ID})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		assert.NotContains(t, titles(readAll(t, jane, TaskAssignmentMine, "")), "Joe inbox")
		assert.Contains(t, titles(readAll(t, joe, TaskAssignmentMine, "")), "Joe inbox")
	})
	t.Run("a different user gets their own lists", func(t *testing.T) {
		_, joe := setup(t)
		mine := titles(readAll(t, joe, TaskAssignmentMine, ""))
		assert.ElementsMatch(t, []string{"Jane to Joe", "Jane to both", "Joe inbox"}, mine)
		assert.ElementsMatch(t, []string{"Joe to Jane"}, titles(readAll(t, joe, TaskAssignmentAssignedByMe, "")))
	})
	t.Run("combines with a filter", func(t *testing.T) {
		jane, _ := setup(t)
		done := mustCreate(t, jane, &Task{Title: "J done", ProjectID: jane.DefaultProjectID, Done: true})
		require.NotZero(t, done.ID)

		open := titles(readAll(t, jane, TaskAssignmentMine, "done = false"))

		assert.NotContains(t, open, "J done")
		assert.Contains(t, open, "J inbox")
	})
	t.Run("a subtask that matches is kept as a root when its parent does not", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		joe := registerPerson(t, "joe", "joe@corp.com")
		shared := createProjectFor(t, joe, "Shared")
		shareProject(t, shared.ID, jane.ID, PermissionWrite)
		parent := mustCreate(t, joe, &Task{Title: "Parent of joe", ProjectID: shared.ID, Assignees: []*user.User{joe}})
		child := mustCreate(t, joe, &Task{Title: "Child for jane", ProjectID: shared.ID, Assignees: []*user.User{jane}})
		s := db.NewSession()
		require.NoError(t, (&TaskRelation{TaskID: child.ID, OtherTaskID: parent.ID, RelationKind: RelationKindParenttask}).Create(s, joe))
		require.NoError(t, s.Commit())
		s.Close()

		rs := db.NewReadSession()
		defer rs.Close()
		tc := &TaskCollection{Expand: []TaskCollectionExpandable{TaskCollectionExpandSubtasks}}
		require.NoError(t, tc.SetAssignmentFilter(TaskAssignmentMine))
		res, _, _, err := tc.ReadAll(rs, jane, "", 1, 100)
		require.NoError(t, err)

		got := titles(res.([]*Task))
		assert.Contains(t, got, "Child for jane", "its parent is not in the list, so it is a root")
		assert.NotContains(t, got, "Parent of joe")
	})
	t.Run("a link share matches nothing", func(t *testing.T) {
		assert.Equal(t, "1 = 0", mustSQL(t, assignmentCondition(TaskAssignmentMine, 0)))
		assert.Equal(t, "1 = 0", mustSQL(t, assignmentCondition(TaskAssignmentAssignedByMe, -3)))
	})
	t.Run("no filter adds no condition and an invalid value is refused", func(t *testing.T) {
		assert.Nil(t, assignmentCondition(TaskAssignmentNone, 5))
		require.Error(t, (&TaskCollection{}).SetAssignmentFilter("everyone"))
	})
}

func mustSQL(t *testing.T, c builder.Cond) string {
	t.Helper()
	sql, _, err := builder.ToSQL(c)
	require.NoError(t, err)
	return sql
}

func mustCreate(t *testing.T, as *user.User, task *Task) *Task {
	t.Helper()
	created, err := createAs(t, as, task)
	require.NoError(t, err)
	return created
}

func taskIDByTitle(t *testing.T, title string) int64 {
	t.Helper()
	s := db.NewReadSession()
	defer s.Close()
	task := &Task{}
	has, err := s.Where("title = ?", title).Get(task)
	require.NoError(t, err)
	require.True(t, has, title)
	return task.ID
}

func createProjectFor(t *testing.T, owner *user.User, title string) *Project {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	p := &Project{Title: title}
	require.NoError(t, p.Create(s, owner))
	require.NoError(t, s.Commit())
	return p
}

func shareProject(t *testing.T, projectID, userID int64, permission Permission) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	_, err := s.Insert(&ProjectUser{ProjectID: projectID, UserID: userID, Permission: permission})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

func TestImportAssignedRows(t *testing.T) {
	run := func(t *testing.T, as *user.User, rows ...*ImportAssignedRow) []ImportRowResult {
		t.Helper()
		for i, r := range rows {
			if r.Number == 0 {
				r.Number = i + 1
			}
		}
		return ImportAssignedRows(context.Background(), as, rows)
	}
	row := func(title string) *ImportAssignedRow { return &ImportAssignedRow{Task: Task{Title: title}} }

	t.Run("a normal user's rows go to their own inbox and are assigned to them", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")

		res := run(t, jane, row("A"))

		require.NoError(t, res[0].Err)
		db.AssertExists(t, "tasks", map[string]interface{}{"id": res[0].TaskID, "project_id": jane.DefaultProjectID, "created_by_id": jane.ID}, false)
		assert.Equal(t, []int64{jane.ID}, assigneeIDs(t, res[0].TaskID))
	})
	t.Run("an existing project by id needs write access, and the owner is not forced as assignor for a normal user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		joe := registerPerson(t, "joe", "joe@corp.com")
		shared := createProjectFor(t, joe, "Shared")
		private := createProjectFor(t, joe, "Private")
		shareProject(t, shared.ID, jane.ID, PermissionWrite)

		ok := row("In shared")
		ok.ProjectID = shared.ID
		denied := row("In private")
		denied.ProjectID = private.ID
		res := run(t, jane, ok, denied)

		require.NoError(t, res[0].Err)
		db.AssertExists(t, "tasks", map[string]interface{}{"id": res[0].TaskID, "project_id": shared.ID, "created_by_id": jane.ID}, false)
		require.Error(t, res[1].Err)
		assert.Zero(t, countRows(t, &Task{}, builder.Eq{"title": "In private"}))
	})
	t.Run("a normal user does not learn who has an account unless they allow to be found by email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		joe := registerPerson(t, "joe", "joe@corp.com")
		own := createProjectFor(t, jane, "Janes")
		shareProject(t, own.ID, joe.ID, PermissionRead)

		hidden := row("Hidden assignee")
		hidden.ProjectID = own.ID
		hidden.AssigneeEmail = "joe@corp.com"
		hiddenAssignor := row("Hidden assignor")
		hiddenAssignor.AssignorEmail = "joe@corp.com"
		unknown := row("Unknown")
		unknown.ProjectID = own.ID
		unknown.AssigneeEmail = "nobody@corp.com"
		res := run(t, jane, hidden, hiddenAssignor, unknown)

		// All three look the same: not found.
		require.Error(t, res[0].Err)
		assert.True(t, IsErrAssigneeEmailNotFound(res[0].Err), "%v", res[0].Err)
		require.Error(t, res[1].Err)
		assert.True(t, IsErrAssignorNotFound(res[1].Err), "%v", res[1].Err)
		require.Error(t, res[2].Err)
		assert.True(t, IsErrAssigneeEmailNotFound(res[2].Err))

		// Once joe is discoverable the assignee works (he can read the project).
		s := db.NewSession()
		_, err := s.ID(joe.ID).Cols("discoverable_by_email").Update(&user.User{DiscoverableByEmail: true})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()
		visible := row("Visible assignee")
		visible.ProjectID = own.ID
		visible.AssigneeEmail = "joe@corp.com"
		require.NoError(t, run(t, jane, visible)[0].Err)

		// An admin is not limited by it.
		boss := registerPerson(t, "boss", "boss@corp.com")
		registerPerson(t, "quiet", "quiet@corp.com")
		makeAdmin(t, boss.ID, true)
		quiet := row("Quiet")
		quiet.AssigneeEmail = "quiet@corp.com"
		require.NoError(t, run(t, boss, quiet)[0].Err)
	})
	t.Run("a normal user cannot fill somebody else's inbox or name a foreign assignor", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		joe := registerPerson(t, "joe", "joe@corp.com")
		s := db.NewSession()
		_, err := s.ID(joe.ID).Cols("discoverable_by_email").Update(&user.User{DiscoverableByEmail: true})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		toOther := row("To joe")
		toOther.AssigneeEmail = "joe@corp.com"
		foreign := row("Foreign assignor")
		foreign.AssignorEmail = "joe@corp.com"
		self := row("Self assignor")
		self.AssignorEmail = "JANE@corp.com"
		res := run(t, jane, toOther, foreign, self)

		require.Error(t, res[0].Err)
		assert.IsType(t, ErrGenericForbidden{}, res[0].Err)
		require.Error(t, res[1].Err)
		require.NoError(t, res[2].Err)
		assert.Zero(t, countRows(t, &Task{}, builder.In("title", "To joe", "Foreign assignor")))
	})
	t.Run("an admin may fill another inbox and name the assignor", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		boss := registerPerson(t, "boss", "boss@corp.com")
		jane := registerPerson(t, "jane", "jane@corp.com")
		ann := registerPerson(t, "ann", "ann@corp.com")
		s := db.NewSession()
		_, err := s.ID(boss.ID).Cols("is_admin").Update(&user.User{IsAdmin: true})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		r := row("For jane")
		r.AssigneeEmail = "jane@corp.com"
		r.AssignorEmail = "ann@corp.com"
		noAssignor := row("For jane too")
		noAssignor.AssigneeEmail = "jane@corp.com"
		res := run(t, boss, r, noAssignor)

		require.NoError(t, res[0].Err)
		db.AssertExists(t, "tasks", map[string]interface{}{"id": res[0].TaskID, "project_id": jane.DefaultProjectID, "created_by_id": ann.ID}, false)
		require.NoError(t, res[1].Err)
		db.AssertExists(t, "tasks", map[string]interface{}{"id": res[1].TaskID, "created_by_id": jane.ID}, false)
	})
	t.Run("one bad row does not stop the others", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")

		bad := row("Bad")
		bad.AssigneeEmail = "nobody@corp.com"
		bad2 := row("")
		res := run(t, jane, row("First"), bad, bad2, row("Last"))

		require.NoError(t, res[0].Err)
		require.Error(t, res[1].Err)
		assert.True(t, IsErrAssigneeEmailNotFound(res[1].Err))
		require.Error(t, res[2].Err, "an empty title is rejected")
		require.NoError(t, res[3].Err)
		assert.Equal(t, int64(2), countRows(t, &Task{}, builder.In("title", "First", "Last")))
	})
	t.Run("labels are created once for the importer and attached", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")

		a := row("A")
		a.Task.Labels = []*Label{{Title: "urgent"}}
		b := row("B")
		b.Task.Labels = []*Label{{Title: "urgent"}, {Title: "  "}}
		res := run(t, jane, a, b)

		require.NoError(t, res[0].Err)
		require.NoError(t, res[1].Err)
		assert.Equal(t, int64(1), countRows(t, &Label{}, builder.And(builder.Eq{"title": "urgent"}, builder.Eq{"created_by_id": jane.ID})))
		assert.Equal(t, int64(2), countRows(t, &LabelTask{}, builder.In("task_id", res[0].TaskID, res[1].TaskID)))
	})
	t.Run("an assignee who cannot read the given project is refused", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		joe := registerPerson(t, "joe", "joe@corp.com")
		mine := createProjectFor(t, jane, "Janes")
		s := db.NewSession()
		_, err := s.ID(joe.ID).Cols("discoverable_by_email").Update(&user.User{DiscoverableByEmail: true})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		r := row("To joe")
		r.ProjectID = mine.ID
		r.AssigneeEmail = "joe@corp.com"
		res := run(t, jane, r)

		require.Error(t, res[0].Err)
		assert.True(t, IsErrUserDoesNotHaveAccessToProject(res[0].Err), "%v", res[0].Err)
		assert.Zero(t, countRows(t, &Task{}, builder.Eq{"title": "To joe"}))
	})
}
