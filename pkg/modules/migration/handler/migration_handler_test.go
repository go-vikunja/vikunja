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

package handler

import (
	"sync/atomic"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubErr struct {
	Msg string `json:"msg"`
}

func (s *stubErr) Error() string { return s.Msg }

// The listener reconstructs migrators, so run counts cannot live on an instance.
var migrateRunCount int64

type stubMigrator struct {
	NameValue  string   `json:"name_value"`
	CheckErr   *stubErr `json:"check_err,omitempty"`
	MigrateErr *stubErr `json:"migrate_err,omitempty"`
	Panic      bool     `json:"panic"`
}

func (m *stubMigrator) Name() string    { return m.NameValue }
func (m *stubMigrator) AuthURL() string { return "" }

func (m *stubMigrator) CheckCredentials() error {
	if m.CheckErr != nil {
		return m.CheckErr
	}
	return nil
}

func (m *stubMigrator) Migrate(_ *user.User) error {
	atomic.AddInt64(&migrateRunCount, 1)
	if m.Panic {
		panic("boom")
	}
	if m.MigrateErr != nil {
		return m.MigrateErr
	}
	return nil
}

func getTestUser(t *testing.T) *user.User {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	u, err := user.GetUserByID(s, 1)
	require.NoError(t, err)
	return u
}

func clearMigrationStatus(t *testing.T) {
	t.Helper()
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	_, err := s.Where("1 = 1").Delete(&migration.Status{})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

func TestStartMigrationClaimsAndDispatchesStatusID(t *testing.T) {
	clearMigrationStatus(t)
	events.ClearDispatchedEvents()
	u := getTestUser(t)

	require.NoError(t, StartMigration(&stubMigrator{NameValue: "stub"}, u))

	dispatched := events.GetDispatchedEvents((&MigrationRequestedEvent{}).Name())
	require.Len(t, dispatched, 1)
	event, ok := dispatched[0].(*MigrationRequestedEvent)
	require.True(t, ok)
	require.NotZero(t, event.MigrationStatusID)

	status, err := migration.GetMigrationStatusByID(event.MigrationStatusID)
	require.NoError(t, err)
	require.NotNil(t, status.ActiveUserID)
	assert.Equal(t, u.ID, *status.ActiveUserID)
}

func TestStartMigrationRefusesSecondMigrationForSameUser(t *testing.T) {
	clearMigrationStatus(t)
	events.ClearDispatchedEvents()
	u := getTestUser(t)

	require.NoError(t, StartMigration(&stubMigrator{NameValue: "stub"}, u))

	err := StartMigration(&stubMigrator{NameValue: "other-stub"}, u)
	var alreadyRunning *migration.ErrMigrationAlreadyRunning
	require.ErrorAs(t, err, &alreadyRunning, "expected ErrMigrationAlreadyRunning, got %v", err)
	assert.Equal(t, "stub", alreadyRunning.MigratorName)
}

func TestStartMigrationCredentialFailureReleasesClaim(t *testing.T) {
	clearMigrationStatus(t)
	events.ClearDispatchedEvents()
	u := getTestUser(t)

	credErr := &stubErr{Msg: "bad credentials"}
	err := StartMigration(&stubMigrator{NameValue: "stub", CheckErr: credErr}, u)
	require.EqualError(t, err, "bad credentials")
	assert.Empty(t, events.GetDispatchedEvents((&MigrationRequestedEvent{}).Name()), "no event must be dispatched on credential failure")

	require.NoError(t, StartMigration(&stubMigrator{NameValue: "stub"}, u))
}

func TestStartMigrationDispatchFailureReleasesClaim(t *testing.T) {
	clearMigrationStatus(t)
	events.ClearDispatchedEvents()
	u := getTestUser(t)

	events.Unfake()
	t.Cleanup(events.Fake)

	err := StartMigration(&stubMigrator{NameValue: "stub"}, u)
	require.Error(t, err)

	_, err = migration.ClaimMigration(&stubMigrator{NameValue: "stub"}, u)
	require.NoError(t, err)
}

func TestMigrationListenerReusesClaimedStatus(t *testing.T) {
	clearMigrationStatus(t)
	u := getTestUser(t)
	RegisterMigratorForEvents(func() migration.Migrator { return &stubMigrator{NameValue: "stub"} })

	status, err := migration.ClaimMigration(&stubMigrator{NameValue: "stub"}, u)
	require.NoError(t, err)

	runCountBefore := atomic.LoadInt64(&migrateRunCount)
	event := &MigrationRequestedEvent{
		Migrator:          &stubMigrator{NameValue: "stub"},
		MigratorKind:      "stub",
		User:              u,
		MigrationStatusID: status.ID,
	}
	events.TestListener(t, event, &MigrationListener{})

	assert.Equal(t, runCountBefore+1, atomic.LoadInt64(&migrateRunCount), "the migration must have run")

	s := db.NewSession()
	defer s.Close()
	count, err := s.Where("user_id = ?", u.ID).Count(&migration.Status{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestMigrationListenerPanicReleasesClaim(t *testing.T) {
	clearMigrationStatus(t)
	u := getTestUser(t)
	RegisterMigratorForEvents(func() migration.Migrator { return &stubMigrator{NameValue: "panic-stub", Panic: true} })

	status, err := migration.ClaimMigration(&stubMigrator{NameValue: "panic-stub"}, u)
	require.NoError(t, err)

	event := &MigrationRequestedEvent{
		Migrator:          &stubMigrator{NameValue: "panic-stub", Panic: true},
		MigratorKind:      "panic-stub",
		User:              u,
		MigrationStatusID: status.ID,
	}
	events.TestListener(t, event, &MigrationListener{})

	fetched, err := migration.GetMigrationStatusByID(status.ID)
	require.NoError(t, err)
	assert.False(t, fetched.FinishedAt.IsZero())
	assert.Nil(t, fetched.ActiveUserID)

	_, err = migration.ClaimMigration(&stubMigrator{NameValue: "panic-stub"}, u)
	require.NoError(t, err)
}

func TestMigrationListenerStaleEventDoesNothing(t *testing.T) {
	clearMigrationStatus(t)
	u := getTestUser(t)
	RegisterMigratorForEvents(func() migration.Migrator { return &stubMigrator{NameValue: "stale-stub"} })

	status, err := migration.ClaimMigration(&stubMigrator{NameValue: "stale-stub"}, u)
	require.NoError(t, err)
	require.NoError(t, migration.FinishMigration(status))

	runCountBefore := atomic.LoadInt64(&migrateRunCount)
	event := &MigrationRequestedEvent{
		Migrator:          &stubMigrator{NameValue: "stale-stub"},
		MigratorKind:      "stale-stub",
		User:              u,
		MigrationStatusID: status.ID,
	}
	events.TestListener(t, event, &MigrationListener{})

	assert.Equal(t, runCountBefore, atomic.LoadInt64(&migrateRunCount), "a stale event for a finished migration must not re-run the migrator")
}
