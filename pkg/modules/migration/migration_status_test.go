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

package migration

import (
	"sync"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testMigrator struct {
	name string
}

func (t *testMigrator) Name() string { return t.name }

func getTestUser(t *testing.T, id int64) *user.User {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	u, err := user.GetUserByID(s, id)
	require.NoError(t, err)
	return u
}

func clearMigrationStatus(t *testing.T) {
	t.Helper()
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()
	_, err := s.Where("1 = 1").Delete(&Status{})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

func assertIsAlreadyRunning(t *testing.T, err error, migratorName string) {
	t.Helper()
	require.Error(t, err)
	var e *ErrMigrationAlreadyRunning
	require.ErrorAs(t, err, &e, "expected ErrMigrationAlreadyRunning, got %v", err)
	assert.Equal(t, migratorName, e.MigratorName)
}

func TestClaimMigrationSerializesPerUser(t *testing.T) {
	clearMigrationStatus(t)
	u1 := getTestUser(t, 1)

	status, err := ClaimMigration(&testMigrator{"todoist"}, u1)
	require.NoError(t, err)
	assert.Equal(t, "todoist", status.MigratorName)
	require.NotNil(t, status.ActiveUserID)
	assert.Equal(t, u1.ID, *status.ActiveUserID)

	_, err = ClaimMigration(&testMigrator{"todoist"}, u1)
	assertIsAlreadyRunning(t, err, "todoist")

	_, err = ClaimMigration(&testMigrator{"csv"}, u1)
	assertIsAlreadyRunning(t, err, "todoist")

	status2, err := ClaimMigration(&testMigrator{"csv"}, getTestUser(t, 2))
	require.NoError(t, err)
	assert.Equal(t, "csv", status2.MigratorName)
}

func TestClaimMigrationConcurrentSameUserOnlyOneWins(t *testing.T) {
	clearMigrationStatus(t)
	u1 := getTestUser(t, 1)

	winnerStatus, losses := claimConcurrently(t, u1, 10)

	require.Len(t, losses, 9, "exactly one concurrent claim must win")
	for _, err := range losses {
		assertIsAlreadyRunning(t, err, "todoist")
	}
	require.NoError(t, FinishMigration(winnerStatus))
}

// A second import arriving right after the first must lose with the domain error, not
// with whatever the driver reports while the winning claim is still being written.
func TestClaimMigrationConcurrentAfterClaimIsAlreadyRunning(t *testing.T) {
	clearMigrationStatus(t)
	u1 := getTestUser(t, 1)

	_, err := ClaimMigration(&testMigrator{"todoist"}, u1)
	require.NoError(t, err)

	winner, losses := claimConcurrently(t, u1, 10)

	require.Nil(t, winner)
	require.Len(t, losses, 10)
	for _, err := range losses {
		assertIsAlreadyRunning(t, err, "todoist")
	}
}

func claimConcurrently(t *testing.T, u *user.User, attempts int) (winner *Status, losses []error) {
	t.Helper()

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s, err := ClaimMigration(&testMigrator{"todoist"}, u)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				losses = append(losses, err)
				return
			}
			winner = s
		}()
	}
	wg.Wait()

	return winner, losses
}

func TestClaimMigrationReleasesAfterFinish(t *testing.T) {
	clearMigrationStatus(t)
	u1 := getTestUser(t, 1)

	status, err := ClaimMigration(&testMigrator{"todoist"}, u1)
	require.NoError(t, err)

	require.NoError(t, FinishMigration(status))

	fetched, err := GetMigrationStatusByID(status.ID)
	require.NoError(t, err)
	assert.False(t, fetched.FinishedAt.IsZero())
	assert.Nil(t, fetched.ActiveUserID)

	_, err = ClaimMigration(&testMigrator{"csv"}, u1)
	require.NoError(t, err)

	s := db.NewSession()
	defer s.Close()
	count, err := s.Where("user_id = ?", u1.ID).Count(&Status{})
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestClaimMigrationBlocksOnLegacyUnfinishedRow(t *testing.T) {
	clearMigrationStatus(t)
	u1 := getTestUser(t, 1)

	legacy := &Status{
		UserID:       u1.ID,
		MigratorName: "trello",
		StartedAt:    time.Now(),
	}
	s := db.NewSession()
	_, err := s.Insert(legacy)
	require.NoError(t, err)
	require.NoError(t, s.Commit())

	_, err = ClaimMigration(&testMigrator{"csv"}, u1)
	assertIsAlreadyRunning(t, err, "trello")
}

func TestClaimMigrationTakesOverStaleClaim(t *testing.T) {
	clearMigrationStatus(t)
	u1 := getTestUser(t, 1)

	config.MigrationClaimTimeout.Set("24h")
	t.Cleanup(func() { config.MigrationClaimTimeout.Set("24h") })

	status, err := ClaimMigration(&testMigrator{"todoist"}, u1)
	require.NoError(t, err)

	_, err = ClaimMigration(&testMigrator{"csv"}, u1)
	assertIsAlreadyRunning(t, err, "todoist")

	config.MigrationClaimTimeout.Set("1ms")
	s := db.NewSession()
	_, err = s.Where("id = ?", status.ID).
		Cols("started_at").
		Update(&Status{StartedAt: time.Now().Add(-time.Hour)})
	require.NoError(t, err)
	require.NoError(t, s.Commit())

	newStatus, err := ClaimMigration(&testMigrator{"csv"}, u1)
	require.NoError(t, err)
	assert.Equal(t, "csv", newStatus.MigratorName)

	fetched, err := GetMigrationStatusByID(status.ID)
	require.NoError(t, err)
	assert.False(t, fetched.FinishedAt.IsZero())
	assert.Nil(t, fetched.ActiveUserID)
}

func TestClaimMigrationRecentClaimIsNotTakenOver(t *testing.T) {
	clearMigrationStatus(t)
	u1 := getTestUser(t, 1)

	config.MigrationClaimTimeout.Set("24h")
	t.Cleanup(func() { config.MigrationClaimTimeout.Set("24h") })

	status, err := ClaimMigration(&testMigrator{"todoist"}, u1)
	require.NoError(t, err)

	s := db.NewSession()
	_, err = s.Where("id = ?", status.ID).
		Cols("started_at").
		Update(&Status{StartedAt: time.Now().Add(-time.Hour)})
	require.NoError(t, err)
	require.NoError(t, s.Commit())

	_, err = ClaimMigration(&testMigrator{"csv"}, u1)
	assertIsAlreadyRunning(t, err, "todoist")
}
