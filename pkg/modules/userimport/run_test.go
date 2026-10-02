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

package userimport

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// prepare loads the fixtures. All fixture users are local or from another OpenID provider, so
// the import leaves them alone: the only managed users are the ones a test inserts.
func prepare(t *testing.T) {
	t.Helper()
	db.LoadAndAssertFixtures(t)
	events.ClearDispatchedEvents()
}

func insertTestUser(t *testing.T, u *user.User) *user.User {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	_, err := s.Insert(u)
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	return u
}

// insertEntraUsers inserts e1@corp.com ... eN@corp.com as active Entra users and returns them.
func insertEntraUsers(t *testing.T, n int) []*user.User {
	t.Helper()
	s := db.NewSession()
	defer s.Close()

	users := make([]*user.User, 0, n)
	for i := 1; i <= n; i++ {
		u := &user.User{
			Username: fmt.Sprintf("e%d", i),
			Email:    fmt.Sprintf("e%d@corp.com", i),
			Name:     fmt.Sprintf("Old Name %d", i),
			Issuer:   entraIssuer,
			Subject:  fmt.Sprintf("sub-%d", i),
			Status:   user.StatusActive,
		}
		_, err := s.Insert(u)
		require.NoError(t, err)
		users = append(users, u)
	}
	require.NoError(t, s.Commit())
	return users
}

func insertSession(t *testing.T, userID int64) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	_, err := s.Insert(&models.Session{
		ID:         uuid.NewString(),
		UserID:     userID,
		TokenHash:  strings.ReplaceAll(uuid.NewString(), "-", ""),
		LastActive: time.Now(),
	})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

func userByEmail(t *testing.T, email string) *user.User {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	u := &user.User{}
	found, err := s.Where("email = ?", email).Get(u)
	require.NoError(t, err)
	require.Truef(t, found, "no user with email %s", email)
	return u
}

func countUsers(t *testing.T, cond builder.Cond) int64 {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	n, err := s.Where(cond).Count(&user.User{})
	require.NoError(t, err)
	return n
}

// writeList writes a user list next to the test and returns its path.
func writeList(t *testing.T, rows ...string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "user_list.csv")
	content := csvHeader + "\n" + strings.Join(rows, "\n") + "\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func opts(file string) Options {
	return Options{File: file, MaxDisablePercent: 20}
}

// listing returns rows for e1..eN: id-<i>, mail in mixed case, a new name, title and department.
func listing(from, to int) []string {
	rows := []string{}
	for i := from; i <= to; i++ {
		rows = append(rows, fmt.Sprintf("id-%d,New Name %d,e%d@corp.com,E%d@Corp.com,Title %d,Dept %d,true", i, i, i, i, i, i))
	}
	return rows
}

func TestRun(t *testing.T) {
	t.Run("creates, updates and deactivates in one run", func(t *testing.T) {
		prepare(t)
		insertEntraUsers(t, 12)
		e1, e12 := userByEmail(t, "e1@corp.com"), userByEmail(t, "e12@corp.com")
		insertSession(t, e1.ID)
		insertSession(t, e12.ID)

		rows := append(listing(1, 11), "id-new,Nina New,nina@corp.com,nina@corp.com,Analyst,Finance,true")
		res, err := RunWithOptions(context.Background(), opts(writeList(t, rows...)))
		require.NoError(t, err)

		assert.Equal(t, 1, res.Created)
		assert.Equal(t, 11, res.Updated)
		assert.Equal(t, 1, res.Disabled)
		assert.False(t, res.DryRun)

		// (1) mapped by mail: name, job title and department follow the file, the email stays
		db.AssertExists(t, "users", map[string]interface{}{
			"id":         e1.ID,
			"name":       "New Name 1",
			"job_title":  "Title 1",
			"department": "Dept 1",
			"import_id":  "id-1",
			"email":      "e1@corp.com",
			"status":     int(user.StatusActive),
		}, false)

		// (2) not in the file: deactivated, sessions gone; the listed user keeps theirs
		db.AssertExists(t, "users", map[string]interface{}{"id": e12.ID, "status": int(user.StatusDisabled)}, false)
		db.AssertMissing(t, "sessions", map[string]interface{}{"user_id": e12.ID})
		db.AssertExists(t, "sessions", map[string]interface{}{"user_id": e1.ID}, false)

		// a person nobody mapped to is created, findable, and has an Inbox
		nina := userByEmail(t, "nina@corp.com")
		assert.Equal(t, user.IssuerImport, nina.Issuer)
		assert.Equal(t, "id-new", nina.Subject)
		assert.Equal(t, "id-new", nina.ImportID)
		assert.Equal(t, "Nina New", nina.Name)
		assert.Equal(t, "Analyst", nina.JobTitle)
		assert.Equal(t, "Finance", nina.Department)
		assert.Equal(t, "nina", nina.Username)
		assert.True(t, nina.DiscoverableByName)
		assert.True(t, nina.DiscoverableByEmail)
		assert.Equal(t, user.StatusActive, nina.Status)
		assert.Empty(t, nina.Password)
		db.AssertExists(t, "projects", map[string]interface{}{"owner_id": nina.ID, "title": "Inbox"}, false)

		// exempt: no fixture user was touched, the one disabled local user is still user17
		assert.Equal(t, int64(1), countUsers(t, builder.And(builder.Eq{"issuer": user.IssuerLocal}, builder.Eq{"status": int(user.StatusDisabled)})))
		assert.Zero(t, countUsers(t, builder.And(builder.Eq{"issuer": "https://some.service.com"}, builder.Eq{"status": int(user.StatusDisabled)})),
			"a user of another OpenID provider is not managed")

		// the deactivation is audited as a status change, with no acting user
		statusEvents := events.GetDispatchedEvents((&models.AdminUserStatusChangedEvent{}).Name())
		require.Len(t, statusEvents, 1)
		evt, ok := statusEvents[0].(*models.AdminUserStatusChangedEvent)
		require.True(t, ok)
		assert.Equal(t, e12.ID, evt.User.ID)
		assert.Equal(t, user.StatusActive, evt.OldStatus)
		assert.Equal(t, user.StatusDisabled, evt.NewStatus)
		assert.Nil(t, evt.Doer)
	})
	t.Run("a second run with the same file changes nothing", func(t *testing.T) {
		prepare(t)
		insertEntraUsers(t, 12)
		file := writeList(t, append(listing(1, 11), "id-new,Nina New,nina@corp.com,nina@corp.com,Analyst,Finance,true")...)

		_, err := RunWithOptions(context.Background(), opts(file))
		require.NoError(t, err)
		res, err := RunWithOptions(context.Background(), opts(file))
		require.NoError(t, err)

		assert.Zero(t, res.Created)
		assert.Zero(t, res.Updated)
		assert.Zero(t, res.Disabled)
		assert.Zero(t, res.Reenabled)
	})
	t.Run("an unchanged file is still processed: a user who joined since is deactivated", func(t *testing.T) {
		prepare(t)
		insertEntraUsers(t, 12)
		file := writeList(t, listing(1, 12)...)

		_, err := RunWithOptions(context.Background(), opts(file))
		require.NoError(t, err)

		s := db.NewSession()
		_, err = s.Insert(&user.User{Username: "joiner", Email: "joiner@corp.com", Issuer: entraIssuer, Subject: "sub-joiner", Status: user.StatusActive})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		res, err := RunWithOptions(context.Background(), opts(file))
		require.NoError(t, err)
		assert.Equal(t, 1, res.Disabled)
		db.AssertExists(t, "users", map[string]interface{}{"email": "joiner@corp.com", "status": int(user.StatusDisabled)}, false)
	})
	t.Run("a returner is re-enabled", func(t *testing.T) {
		prepare(t)
		insertEntraUsers(t, 12)
		e3 := userByEmail(t, "e3@corp.com")
		s := db.NewSession()
		_, err := s.ID(e3.ID).Cols("status").Update(&user.User{Status: user.StatusDisabled})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		res, err := RunWithOptions(context.Background(), opts(writeList(t, listing(1, 12)...)))
		require.NoError(t, err)

		assert.Equal(t, 1, res.Reenabled)
		db.AssertExists(t, "users", map[string]interface{}{"id": e3.ID, "status": int(user.StatusActive)}, false)

		statusEvents := events.GetDispatchedEvents((&models.AdminUserStatusChangedEvent{}).Name())
		require.Len(t, statusEvents, 1)
		evt, ok := statusEvents[0].(*models.AdminUserStatusChangedEvent)
		require.True(t, ok)
		assert.Equal(t, user.StatusDisabled, evt.OldStatus)
		assert.Equal(t, user.StatusActive, evt.NewStatus)
	})
	t.Run("accountEnabled false deactivates", func(t *testing.T) {
		prepare(t)
		insertEntraUsers(t, 12)
		rows := listing(1, 12)
		rows[4] = "id-5,Off,e5@corp.com,e5@corp.com,,,false"

		res, err := RunWithOptions(context.Background(), opts(writeList(t, rows...)))
		require.NoError(t, err)

		assert.Equal(t, 1, res.Disabled)
		db.AssertExists(t, "users", map[string]interface{}{"email": "e5@corp.com", "status": int(user.StatusDisabled)}, false)
	})
	t.Run("exempt users are untouched even when absent from the file", func(t *testing.T) {
		prepare(t)
		insertEntraUsers(t, 12)
		e1 := userByEmail(t, "e1@corp.com")
		s := db.NewSession()
		_, err := s.ID(e1.ID).Cols("is_admin").Update(&user.User{IsAdmin: true})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		// e1 is left out of the file
		res, err := RunWithOptions(context.Background(), opts(writeList(t, listing(2, 12)...)))
		require.NoError(t, err)

		assert.Zero(t, res.Disabled)
		db.AssertExists(t, "users", map[string]interface{}{"id": e1.ID, "status": int(user.StatusActive), "is_admin": true}, false)
	})
	t.Run("a listed admin does not get a duplicate account", func(t *testing.T) {
		prepare(t)
		insertEntraUsers(t, 12)
		before := countUsers(t, builder.Expr("1 = 1"))
		e1 := userByEmail(t, "e1@corp.com")
		s := db.NewSession()
		_, err := s.ID(e1.ID).Cols("is_admin").Update(&user.User{IsAdmin: true})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		res, err := RunWithOptions(context.Background(), opts(writeList(t, listing(1, 12)...)))
		require.NoError(t, err)

		assert.Zero(t, res.Created)
		assert.Equal(t, before, countUsers(t, builder.Expr("1 = 1")))
	})
	t.Run("matches by import id when the mail changed, and never edits the email", func(t *testing.T) {
		prepare(t)
		insertEntraUsers(t, 12)
		e2 := userByEmail(t, "e2@corp.com")
		s := db.NewSession()
		_, err := s.ID(e2.ID).Cols("import_id").Update(&user.User{ImportID: "id-2"})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		rows := listing(1, 12)
		rows[1] = "id-2,Renamed Mail,e2@corp.com,new-mail@corp.com,,,true"
		res, err := RunWithOptions(context.Background(), opts(writeList(t, rows...)))
		require.NoError(t, err)

		assert.Zero(t, res.Disabled)
		assert.Zero(t, res.Created)
		db.AssertExists(t, "users", map[string]interface{}{"id": e2.ID, "email": "e2@corp.com", "name": "Renamed Mail", "status": int(user.StatusActive)}, false)
	})

	t.Run("dry run reports the plan and changes nothing", func(t *testing.T) {
		prepare(t)
		insertEntraUsers(t, 12)
		activeBefore := countUsers(t, builder.Eq{"status": int(user.StatusActive)})
		total := countUsers(t, builder.Expr("1 = 1"))

		o := opts(writeList(t, append(listing(1, 11), "id-new,Nina,nina@corp.com,nina@corp.com,,,true")...))
		o.DryRun = true
		res, err := RunWithOptions(context.Background(), o)
		require.NoError(t, err)

		assert.True(t, res.DryRun)
		assert.Equal(t, 1, res.Created)
		assert.Equal(t, 1, res.Disabled)
		assert.Zero(t, events.CountDispatchedEvents((&models.AdminUserStatusChangedEvent{}).Name()), "a dry run queues no events")
		assert.Equal(t, activeBefore, countUsers(t, builder.Eq{"status": int(user.StatusActive)}))
		assert.Equal(t, total, countUsers(t, builder.Expr("1 = 1")))
		db.AssertExists(t, "users", map[string]interface{}{"email": "e1@corp.com", "name": "Old Name 1"}, false)
	})
	t.Run("the safety limit aborts without changing anything", func(t *testing.T) {
		prepare(t)
		insertEntraUsers(t, 12)
		activeBefore := countUsers(t, builder.Eq{"status": int(user.StatusActive)})
		total := countUsers(t, builder.Expr("1 = 1"))

		// only one of twelve is listed: eleven would be deactivated
		res, err := RunWithOptions(context.Background(), opts(writeList(t, append(listing(1, 1), "id-new,Nina,nina@corp.com,nina@corp.com,,,true")...)))

		require.Error(t, err)
		assert.Equal(t, 11, res.Disabled, "the result still reports what was planned")
		assert.Equal(t, activeBefore, countUsers(t, builder.Eq{"status": int(user.StatusActive)}))
		assert.Equal(t, total, countUsers(t, builder.Expr("1 = 1")), "nobody was created either")
		db.AssertExists(t, "users", map[string]interface{}{"email": "e1@corp.com", "name": "Old Name 1"}, false)
	})

	t.Run("a file without people aborts", func(t *testing.T) {
		prepare(t)
		insertEntraUsers(t, 12)
		activeBefore := countUsers(t, builder.Eq{"status": int(user.StatusActive)})

		_, err := RunWithOptions(context.Background(), opts(writeList(t)))

		require.Error(t, err)
		assert.Equal(t, activeBefore, countUsers(t, builder.Eq{"status": int(user.StatusActive)}))
	})
	t.Run("a file with the wrong header aborts", func(t *testing.T) {
		prepare(t)
		insertEntraUsers(t, 12)
		path := filepath.Join(t.TempDir(), "user_list.csv")
		require.NoError(t, os.WriteFile(path, []byte("foo,bar\n1,2\n"), 0o600))

		_, err := RunWithOptions(context.Background(), opts(path))

		require.Error(t, err)
		assert.Zero(t, countUsers(t, builder.Eq{"status": int(user.StatusDisabled), "issuer": entraIssuer}))
	})
	t.Run("a missing file is an error", func(t *testing.T) {
		prepare(t)
		_, err := RunWithOptions(context.Background(), opts(filepath.Join(t.TempDir(), "nope.csv")))
		require.Error(t, err)
	})
	t.Run("a stale file is refused, a fresh one is not", func(t *testing.T) {
		prepare(t)
		insertEntraUsers(t, 12)
		file := writeList(t, listing(1, 12)...)
		old := time.Now().Add(-48 * time.Hour)
		require.NoError(t, os.Chtimes(file, old, old))

		o := opts(file)
		o.MaxFileAge = 26 * time.Hour
		_, err := RunWithOptions(context.Background(), o)
		var stale *ErrStaleFile
		require.ErrorAs(t, err, &stale)

		require.NoError(t, os.Chtimes(file, time.Now(), time.Now()))
		_, err = RunWithOptions(context.Background(), o)
		require.NoError(t, err)
	})
	t.Run("a file that was modified moments ago is refused until it has settled", func(t *testing.T) {
		prepare(t)
		insertEntraUsers(t, 12)
		file := writeList(t, listing(1, 12)...)

		o := opts(file)
		o.SettleTime = time.Minute
		_, err := RunWithOptions(context.Background(), o)
		require.ErrorIs(t, err, ErrFileChanging)
		assert.Zero(t, countUsers(t, builder.Eq{"issuer": entraIssuer, "status": int(user.StatusDisabled)}))

		settled := time.Now().Add(-5 * time.Minute)
		require.NoError(t, os.Chtimes(file, settled, settled))
		_, err = RunWithOptions(context.Background(), o)
		require.NoError(t, err)
	})
	t.Run("readList returns the content of a settled, unchanged file", func(t *testing.T) {
		file := writeList(t, listing(1, 3)...)
		settled := time.Now().Add(-5 * time.Minute)
		require.NoError(t, os.Chtimes(file, settled, settled))

		data, err := readList(Options{File: file, SettleTime: time.Minute, MaxFileAge: 26 * time.Hour})

		require.NoError(t, err)
		assert.Contains(t, string(data), "id-1,New Name 1")
		assert.Contains(t, string(data), "id-3,New Name 3")
	})
	t.Run("readList refuses a directory", func(t *testing.T) {
		_, err := readList(Options{File: t.TempDir()})
		require.Error(t, err)
	})
	t.Run("a run that starts while another one is going is skipped", func(t *testing.T) {
		prepare(t)
		runMu.Lock()
		defer runMu.Unlock()

		_, err := RunWithOptions(context.Background(), opts(writeList(t, listing(1, 1)...)))

		require.ErrorIs(t, err, ErrAlreadyRunning)
	})
	t.Run("a batch of deactivations is atomic and queues no events when it fails", func(t *testing.T) {
		prepare(t)
		users := insertEntraUsers(t, 3)

		// The first deactivation succeeds. Deactivating the last admin then fails in the backstop
		// guard (the fixtures contain no admin at all), which must undo the first one as well.
		lastAdmin := users[2]
		lastAdmin.IsAdmin = true
		err := inTransaction(context.Background(), func(s *xorm.Session) error {
			return applyDisables(s, []*user.User{users[0], lastAdmin})
		})

		require.Error(t, err)
		assert.True(t, user.IsErrLastAdmin(err))
		assert.Zero(t, countUsers(t, builder.Eq{"status": int(user.StatusDisabled), "issuer": entraIssuer}), "the first deactivation was rolled back")
		assert.Zero(t, events.CountDispatchedEvents((&models.AdminUserStatusChangedEvent{}).Name()))
	})
	t.Run("a username held by a disabled user does not break the run", func(t *testing.T) {
		prepare(t)
		insertEntraUsers(t, 12)
		insertTestUser(t, &user.User{
			Username: "jdoe", Email: "old@corp.com", Issuer: entraIssuer, Subject: "sub-old", Status: user.StatusDisabled,
		})

		// A new hire whose sign-in name is the leaver's username.
		rows := append(listing(1, 12), "id-new,New Hire,jdoe@corp.com,new@corp.com,,,true")
		res, err := RunWithOptions(context.Background(), opts(writeList(t, rows...)))
		require.NoError(t, err)

		assert.Equal(t, 1, res.Created)
		assert.Zero(t, res.Failed)
		hire := userByEmail(t, "new@corp.com")
		assert.NotEmpty(t, hire.Username)
		assert.NotEqual(t, "jdoe", hire.Username, "the leaver keeps their username, the hire gets a generated one")
	})
	t.Run("a person that cannot be created is skipped, the others are still created", func(t *testing.T) {
		prepare(t)
		plan := &Plan{Create: []*Person{
			{ID: "id-bad", Mail: "", DisplayName: "No Mail", Enabled: true},
			{ID: "id-good", Mail: "good@corp.com", DisplayName: "Good", Enabled: true},
		}}
		result := &Result{}

		err := apply(context.Background(), plan, result)

		require.NoError(t, err)
		assert.Equal(t, 1, result.Created)
		assert.Equal(t, 1, result.Failed)
		db.AssertExists(t, "users", map[string]interface{}{"email": "good@corp.com", "import_id": "id-good"}, false)
	})
	t.Run("a person whose import id was taken since the plan was built is not created twice", func(t *testing.T) {
		prepare(t)
		insertTestUser(t, &user.User{
			Username: "logged-in", Email: "x-token@corp.com", Issuer: entraIssuer, Subject: "sub-x", ImportID: "id-x", Status: user.StatusActive,
		})
		total := countUsers(t, builder.Expr("1 = 1"))

		var made bool
		err := inTransaction(context.Background(), func(s *xorm.Session) (err error) {
			made, err = createImportedUser(s, &Person{ID: "id-x", Mail: "x@corp.com", DisplayName: "X", Enabled: true})
			return err
		})

		require.NoError(t, err)
		assert.False(t, made)
		assert.Equal(t, total, countUsers(t, builder.Expr("1 = 1")))
	})
	t.Run("many deactivations are applied in several transactions and all audited", func(t *testing.T) {
		prepare(t)
		n := batchSize*2 + 5
		insertEntraUsers(t, n)

		o := opts(writeList(t, listing(1, 1)...))
		o.MaxDisablePercent = 100
		res, err := RunWithOptions(context.Background(), o)
		require.NoError(t, err)

		assert.Equal(t, n-1, res.Disabled)
		assert.Equal(t, int64(n-1), countUsers(t, builder.And(builder.Eq{"issuer": entraIssuer}, builder.Eq{"status": int(user.StatusDisabled)})))
		assert.Equal(t, n-1, events.CountDispatchedEvents((&models.AdminUserStatusChangedEvent{}).Name()))
	})
	t.Run("many updates are applied in several transactions", func(t *testing.T) {
		prepare(t)
		n := batchSize*2 + 5
		insertEntraUsers(t, n)

		res, err := RunWithOptions(context.Background(), opts(writeList(t, listing(1, n)...)))
		require.NoError(t, err)

		assert.Equal(t, n, res.Updated)
		assert.Equal(t, int64(n), countUsers(t, builder.Like{"name", "New Name"}))
	})
}

func TestChunk(t *testing.T) {
	assert.Empty(t, chunk([]int{}, 3))
	assert.Equal(t, [][]int{{1, 2, 3}, {4, 5}}, chunk([]int{1, 2, 3, 4, 5}, 3))
	assert.Equal(t, [][]int{{1, 2}}, chunk([]int{1, 2}, 5))
	assert.Equal(t, [][]int{{1}, {2}}, chunk([]int{1, 2}, 1))
}
