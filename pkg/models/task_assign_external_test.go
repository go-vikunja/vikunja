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
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

// Fixture bot 23 is owned by user 21 and has no access to any project.
const testBotID int64 = 23

// makeAdmin sets the instance admin flag of a user.
func makeAdmin(t *testing.T, id int64, admin bool) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	_, err := s.ID(id).Cols("is_admin").Update(&user.User{IsAdmin: admin})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

// getBot returns fixture bot 23. Its owner (user 21) is made an instance admin, because only bots
// of admins may act in the name of other people.
func getBot(t *testing.T) *user.User {
	t.Helper()
	makeAdmin(t, 21, true)
	return botByID(t, testBotID)
}

func botByID(t *testing.T, id int64) *user.User {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	bot, err := user.GetUserByID(s, id)
	require.NoError(t, err)
	require.True(t, bot.IsBot())
	return bot
}

// registerPerson creates a normal user with an Inbox, like a first login or the user import does.
func registerPerson(t *testing.T, username, email string) *user.User {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	u, err := RegisterUser(s, &user.User{
		Username: username,
		Email:    email,
		Password: "12345678",
	}, user.CreateUserOptions{SkipEmailConfirm: true})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	return u
}

func assign(t *testing.T, bot *user.User, in *AssignTaskByEmailInput) (*AssignTaskByEmailResult, error) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	res, err := AssignTaskByEmail(s, bot, in)
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, err
	}
	require.NoError(t, s.Commit())
	events.DispatchPending(context.Background(), s)
	return res, nil
}

func countRows(t *testing.T, bean interface{}, cond builder.Cond) int64 {
	t.Helper()
	s := db.NewReadSession()
	defer s.Close()
	n, err := s.Where(cond).Count(bean)
	require.NoError(t, err)
	return n
}

func TestAssignTaskByEmail_Inbox(t *testing.T) {
	t.Run("creates the task in the assignee's inbox and assigns it", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		events.ClearDispatchedEvents()
		bot := getBot(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		due := time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)

		res, err := assign(t, bot, &AssignTaskByEmailInput{
			AssigneeEmail: "jane@corp.com", Title: "Approve invoice", Description: "<p>Please</p>", DueDate: due, Priority: 3,
		})

		require.NoError(t, err)
		assert.True(t, res.Created)
		assert.Equal(t, jane.DefaultProjectID, res.Task.ProjectID, "the assignee's inbox")
		assert.Equal(t, jane.ID, res.Assignee.ID)
		db.AssertExists(t, "tasks", map[string]interface{}{
			"id": res.Task.ID, "project_id": jane.DefaultProjectID, "title": "Approve invoice",
			"created_by_id": bot.ID, "priority": 3,
		}, false)
		db.AssertExists(t, "task_assignees", map[string]interface{}{"task_id": res.Task.ID, "user_id": jane.ID}, false)
		assert.Positive(t, res.Task.Index)
		assert.NotEmpty(t, res.Task.Identifier)
		db.AssertExists(t, "projects", map[string]interface{}{"id": jane.DefaultProjectID, "title": "Inbox", "owner_id": jane.ID}, false)
	})
	t.Run("queues the normal task and assignment events with the bot as doer", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		events.ClearDispatchedEvents()
		bot := getBot(t)
		registerPerson(t, "jane", "jane@corp.com")
		events.ClearDispatchedEvents()

		_, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "T"})
		require.NoError(t, err)

		created := events.GetDispatchedEvents((&TaskCreatedEvent{}).Name())
		require.Len(t, created, 1)
		assert.Equal(t, bot.ID, created[0].(*TaskCreatedEvent).Doer.ID)
		assigned := events.GetDispatchedEvents((&TaskAssigneeCreatedEvent{}).Name())
		require.Len(t, assigned, 1)
		assert.Equal(t, bot.ID, assigned[0].(*TaskAssigneeCreatedEvent).Doer.ID)
	})
	t.Run("a person who never logged in but was imported is assignable", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		imported := registerPerson(t, "imported", "imported@corp.com")
		s := db.NewSession()
		_, err := s.ID(imported.ID).Cols("issuer", "subject").Update(&user.User{Issuer: user.IssuerImport, Subject: "oid-1"})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		res, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "imported@corp.com", Title: "Welcome"})

		require.NoError(t, err)
		assert.Equal(t, imported.ID, res.Assignee.ID)
	})
	t.Run("a user without a default project gets an inbox on the spot", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		s := db.NewSession()
		_, err := s.ID(16).Cols("default_project_id").Update(&user.User{DefaultProjectID: 0})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		res, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "user16@example.com", Title: "T"})

		require.NoError(t, err)
		db.AssertExists(t, "projects", map[string]interface{}{"id": res.Task.ProjectID, "owner_id": 16, "title": "Inbox"}, false)
		db.AssertExists(t, "users", map[string]interface{}{"id": 16, "default_project_id": res.Task.ProjectID}, false)
	})
	t.Run("a default project the assignee does not own is refused, nothing is written", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		before := countRows(t, &Task{}, builder.Expr("1 = 1"))
		// Fixture user2 has default_project_id 4, which user2 does not own.
		s := db.NewSession()
		p, err := GetProjectSimpleByID(s, 4)
		require.NoError(t, err)
		s.Close()
		require.NotEqual(t, int64(2), p.OwnerID)

		_, err = assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "user2@example.com", Title: "T"})

		require.Error(t, err)
		assert.True(t, IsErrNoUsableInbox(err), "%v", err)
		assert.Equal(t, before, countRows(t, &Task{}, builder.Expr("1 = 1")))
	})
	t.Run("an archived inbox is refused", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		s := db.NewSession()
		_, err := s.ID(jane.DefaultProjectID).Cols("is_archived").Update(&Project{IsArchived: true})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		_, err = assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "T"})

		require.Error(t, err)
		assert.True(t, IsErrNoUsableInbox(err), "%v", err)
	})
}

func TestAssignTaskByEmail_Resolve(t *testing.T) {
	t.Run("email is trimmed and compared case-insensitively", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		jane := registerPerson(t, "jane", "Jane.Doe@Corp.com")

		res, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "  jane.doe@CORP.COM ", Title: "T"})

		require.NoError(t, err)
		assert.Equal(t, jane.ID, res.Assignee.ID)
	})
	t.Run("unknown and empty emails are not found", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		for _, email := range []string{"nobody@corp.com", "", "   "} {
			_, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: email, Title: "T"})
			require.Error(t, err, email)
			assert.True(t, IsErrAssigneeEmailNotFound(err), "%q: %v", email, err)
		}
	})
	t.Run("disabled and locked users are refused", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		for _, email := range []string{"user17@example.com", "user18@example.com"} {
			_, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: email, Title: "T"})
			require.Error(t, err, email)
			assert.True(t, IsErrAssigneeNotActive(err), "%s: %v", email, err)
		}
	})
	t.Run("a bot cannot be the assignee", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		s := db.NewSession()
		_, err := s.ID(24).Cols("email").Update(&user.User{Email: "bot24@corp.com"})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		_, err = assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "bot24@corp.com", Title: "T"})

		require.Error(t, err)
		assert.True(t, IsErrAssigneeEmailNotFound(err))
	})
	t.Run("two users with the same email are ambiguous", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		// user13 and user14 are different users; give them one address (only local users are unique).
		s := db.NewSession()
		_, err := s.ID(14).Cols("email").Update(&user.User{Email: "user13@example.com"})
		require.NoError(t, err)
		_, err = s.ID(13).Cols("email").Update(&user.User{Email: "user13@example.com"})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		_, err = assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "user13@example.com", Title: "T"})

		require.Error(t, err)
		assert.True(t, IsErrAssigneeAmbiguous(err), "%v", err)
	})
}

func TestAssignTaskByEmail_Authorization(t *testing.T) {
	t.Run("a human user may not call it", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		registerPerson(t, "jane", "jane@corp.com")
		s := db.NewSession()
		human, err := user.GetUserByID(s, 1)
		require.NoError(t, err)
		s.Close()
		before := countRows(t, &Task{}, builder.Expr("1 = 1"))

		_, err = assign(t, human, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "T"})

		require.Error(t, err)
		assert.IsType(t, ErrGenericForbidden{}, err)
		assert.Equal(t, before, countRows(t, &Task{}, builder.Expr("1 = 1")))
	})
	t.Run("a bot of an ordinary user may not call it, whoever owns it later", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		registerPerson(t, "jane", "jane@corp.com")
		ordinary := botByID(t, testBotID) // owner user 21 is no admin here
		before := countRows(t, &Task{}, builder.Expr("1 = 1"))

		_, err := assign(t, ordinary, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "Spam"})

		require.Error(t, err)
		assert.IsType(t, ErrGenericForbidden{}, err)
		assert.Equal(t, before, countRows(t, &Task{}, builder.Expr("1 = 1")))

		// An admin owner makes it trusted, and an owner who is disabled makes it untrusted again.
		makeAdmin(t, 21, true)
		_, err = assign(t, botByID(t, testBotID), &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "Fine"})
		require.NoError(t, err)

		s := db.NewSession()
		_, err = s.ID(21).Cols("status").Update(&user.User{Status: user.StatusDisabled})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()
		_, err = assign(t, botByID(t, testBotID), &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "Too late"})
		require.Error(t, err)
	})
	t.Run("a nil caller is refused", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		_, err := assign(t, nil, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "T"})
		require.Error(t, err)
	})
	t.Run("an explicit project needs write access for the bot", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		other := registerPerson(t, "other", "other@corp.com")
		before := countRows(t, &Task{}, builder.Expr("1 = 1"))

		// jane's inbox is not shared with the bot.
		_, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "T", ProjectID: other.DefaultProjectID})
		require.Error(t, err)
		assert.IsType(t, ErrGenericForbidden{}, err)
		assert.Equal(t, before, countRows(t, &Task{}, builder.Expr("1 = 1")))

		// Shared with write access: allowed. The assignee must also be able to read it.
		s := db.NewSession()
		_, err = s.Insert(&ProjectUser{ProjectID: jane.DefaultProjectID, UserID: bot.ID, Permission: PermissionWrite})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()
		res, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "T", ProjectID: jane.DefaultProjectID})
		require.NoError(t, err)
		assert.Equal(t, jane.DefaultProjectID, res.Task.ProjectID)
	})
	t.Run("read-only access is not enough", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		s := db.NewSession()
		_, err := s.Insert(&ProjectUser{ProjectID: jane.DefaultProjectID, UserID: bot.ID, Permission: PermissionRead})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		_, err = assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "T", ProjectID: jane.DefaultProjectID})

		require.Error(t, err)
	})
	t.Run("an assignee who cannot read the explicit project is refused", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		other := registerPerson(t, "other", "other@corp.com")
		s := db.NewSession()
		_, err := s.Insert(&ProjectUser{ProjectID: other.DefaultProjectID, UserID: bot.ID, Permission: PermissionWrite})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()
		before := countRows(t, &Task{}, builder.Expr("1 = 1"))

		_, err = assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "T", ProjectID: other.DefaultProjectID})

		require.Error(t, err)
		assert.True(t, IsErrUserDoesNotHaveAccessToProject(err), "%v", err)
		assert.Equal(t, before, countRows(t, &Task{}, builder.Expr("1 = 1")), "the failed assignment rolls the task back")
		_ = jane
	})
	t.Run("the bot cannot read, update or delete what it created in the inbox", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		registerPerson(t, "jane", "jane@corp.com")
		res, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "T"})
		require.NoError(t, err)

		s := db.NewSession()
		defer s.Close()
		task := &Task{ID: res.Task.ID}
		canRead, _, err := task.CanRead(s, bot)
		require.NoError(t, err)
		assert.False(t, canRead)
		canUpdate, err := task.CanUpdate(s, bot)
		require.NoError(t, err)
		assert.False(t, canUpdate)
		canDelete, err := task.CanDelete(s, bot)
		require.NoError(t, err)
		assert.False(t, canDelete)
	})
}

func TestAssignTaskByEmail_ExternalID(t *testing.T) {
	t.Run("the same external id returns the same task and creates nothing", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		registerPerson(t, "jane", "jane@corp.com")
		in := &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "Invoice", ExternalID: "INV-1"}

		first, err := assign(t, bot, in)
		require.NoError(t, err)
		require.True(t, first.Created)
		events.ClearDispatchedEvents()
		second, err := assign(t, bot, in)

		require.NoError(t, err)
		assert.False(t, second.Created)
		assert.Equal(t, first.Task.ID, second.Task.ID)
		assert.Equal(t, int64(1), countRows(t, &Task{}, builder.Eq{"title": "Invoice"}))
		assert.Equal(t, int64(1), countRows(t, &TaskExternalRef{}, builder.Eq{"external_id": "INV-1"}))
		assert.Zero(t, events.CountDispatchedEvents((&TaskCreatedEvent{}).Name()), "a retry notifies nobody")
	})
	t.Run("another bot may use the same external id", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		makeAdmin(t, 22, true)
		other := botByID(t, 24)
		registerPerson(t, "jane", "jane@corp.com")

		a, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "A", ExternalID: "X"})
		require.NoError(t, err)
		b, err := assign(t, other, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "B", ExternalID: "X"})
		require.NoError(t, err)

		assert.NotEqual(t, a.Task.ID, b.Task.ID)
		assert.True(t, b.Created)
	})
	t.Run("the same external id for another assignee is a conflict", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		registerPerson(t, "jane", "jane@corp.com")
		registerPerson(t, "joe", "joe@corp.com")
		_, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "A", ExternalID: "X"})
		require.NoError(t, err)

		_, err = assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "joe@corp.com", Title: "A", ExternalID: "X"})

		require.Error(t, err)
		assert.True(t, IsErrExternalIDConflict(err), "%v", err)
	})
	t.Run("a deleted task frees its external id", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		registerPerson(t, "jane", "jane@corp.com")
		first, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "A", ExternalID: "X"})
		require.NoError(t, err)
		s := db.NewSession()
		_, err = s.Where("id = ?", first.Task.ID).Delete(&Task{})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		second, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "A again", ExternalID: "X"})

		require.NoError(t, err)
		assert.True(t, second.Created)
		assert.NotEqual(t, first.Task.ID, second.Task.ID)
		assert.Equal(t, int64(1), countRows(t, &TaskExternalRef{}, builder.Eq{"external_id": "X"}))
	})
	t.Run("hard-deleting the task or the bot removes its references and baseline", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)
		jane := registerPerson(t, "jane", "jane@corp.com")
		a, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "A", ExternalID: "A"})
		require.NoError(t, err)
		b, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: "B", ExternalID: "B"})
		require.NoError(t, err)
		s := db.NewSession()
		_, err = s.Insert(&TaskBaseline{TaskID: a.Task.ID, ProjectID: jane.DefaultProjectID, SavedAt: time.Now(), SavedByID: jane.ID})
		require.NoError(t, err)
		require.NoError(t, hardDeleteTask(s, &Task{ID: a.Task.ID}))
		require.NoError(t, s.Commit())
		s.Close()

		assert.Zero(t, countRows(t, &TaskExternalRef{}, builder.Eq{"task_id": a.Task.ID}))
		assert.Zero(t, countRows(t, &TaskBaseline{}, builder.Eq{"task_id": a.Task.ID}))
		assert.Equal(t, int64(1), countRows(t, &TaskExternalRef{}, builder.Eq{"task_id": b.Task.ID}))

		s = db.NewSession()
		defer s.Close()
		botUser := botByID(t, testBotID)
		require.NoError(t, DeleteUser(s, botUser))
		require.NoError(t, s.Commit())
		assert.Zero(t, countRows(t, &TaskExternalRef{}, builder.Eq{"bot_id": testBotID}))
	})
	t.Run("a failed call leaves no reference behind", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		bot := getBot(t)

		_, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "nobody@corp.com", Title: "A", ExternalID: "X"})

		require.Error(t, err)
		assert.Zero(t, countRows(t, &TaskExternalRef{}, builder.Expr("1 = 1")))
	})
}

func TestAssignTaskByEmail_Validation(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	bot := getBot(t)
	registerPerson(t, "jane", "jane@corp.com")
	before := countRows(t, &Task{}, builder.Expr("1 = 1"))

	cases := map[string]*AssignTaskByEmailInput{
		"empty title":         {AssigneeEmail: "jane@corp.com", Title: ""},
		"blank title":         {AssigneeEmail: "jane@corp.com", Title: "   "},
		"long title":          {AssigneeEmail: "jane@corp.com", Title: strings.Repeat("a", 251)},
		"negative priority":   {AssigneeEmail: "jane@corp.com", Title: "T", Priority: -1},
		"priority too high":   {AssigneeEmail: "jane@corp.com", Title: "T", Priority: 6},
		"long external id":    {AssigneeEmail: "jane@corp.com", Title: "T", ExternalID: strings.Repeat("e", 251)},
		"unknown project":     {AssigneeEmail: "jane@corp.com", Title: "T", ProjectID: 99999},
		"title with only 250": nil,
	}
	delete(cases, "title with only 250")
	for name, in := range cases {
		_, err := assign(t, bot, in)
		assert.Error(t, err, name)
	}
	assert.Equal(t, before, countRows(t, &Task{}, builder.Expr("1 = 1")))

	t.Run("a title of exactly 250 characters is fine", func(t *testing.T) {
		res, err := assign(t, bot, &AssignTaskByEmailInput{AssigneeEmail: "jane@corp.com", Title: strings.Repeat("a", 250)})
		require.NoError(t, err)
		assert.Len(t, res.Task.Title, 250)
	})
}
