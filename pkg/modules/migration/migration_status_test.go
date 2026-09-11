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
	t.Cleanup(func() { config.MigrationClaimTimeout.Set("5m") })

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
	assert.Nil(t, fetched.HeartbeatAt, "the fallback to started_at only applies to rows that never beat")
	assert.False(t, fetched.FinishedAt.IsZero())
	assert.Nil(t, fetched.ActiveUserID)
}

func TestClaimMigrationRecentClaimIsNotTakenOver(t *testing.T) {
	clearMigrationStatus(t)
	u1 := getTestUser(t, 1)

	config.MigrationClaimTimeout.Set("24h")
	t.Cleanup(func() { config.MigrationClaimTimeout.Set("5m") })

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

func setClaimTimestamps(t *testing.T, statusID int64, startedAt time.Time, heartbeatAt *time.Time) {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	_, err := s.Where("id = ?", statusID).
		Cols("started_at", "heartbeat_at").
		Update(&Status{StartedAt: startedAt, HeartbeatAt: heartbeatAt})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

// The regression this feature exists to prevent: a big import runs longer than the timeout
// and must keep its slot instead of being run a second time in parallel.
func TestClaimMigrationRecentHeartbeatKeepsLongRunningClaim(t *testing.T) {
	clearMigrationStatus(t)
	u1 := getTestUser(t, 1)

	config.MigrationClaimTimeout.Set("1h")
	t.Cleanup(func() { config.MigrationClaimTimeout.Set("5m") })

	status, err := ClaimMigration(&testMigrator{"todoist"}, u1)
	require.NoError(t, err)

	beatAt := time.Now()
	setClaimTimestamps(t, status.ID, time.Now().Add(-24*time.Hour), &beatAt)

	_, err = ClaimMigration(&testMigrator{"csv"}, u1)
	assertIsAlreadyRunning(t, err, "todoist")

	fetched, err := GetMigrationStatusByID(status.ID)
	require.NoError(t, err)
	assert.True(t, fetched.FinishedAt.IsZero())
	require.NotNil(t, fetched.ActiveUserID)
	assert.Equal(t, u1.ID, *fetched.ActiveUserID)
}

func TestClaimMigrationTakesOverStaleHeartbeat(t *testing.T) {
	clearMigrationStatus(t)
	u1 := getTestUser(t, 1)

	config.MigrationClaimTimeout.Set("1h")
	t.Cleanup(func() { config.MigrationClaimTimeout.Set("5m") })

	status, err := ClaimMigration(&testMigrator{"todoist"}, u1)
	require.NoError(t, err)

	// started_at alone would keep this claim; only the dead heartbeat releases it.
	beatAt := time.Now().Add(-2 * time.Hour)
	setClaimTimestamps(t, status.ID, time.Now(), &beatAt)

	newStatus, err := ClaimMigration(&testMigrator{"csv"}, u1)
	require.NoError(t, err)
	assert.Equal(t, "csv", newStatus.MigratorName)

	fetched, err := GetMigrationStatusByID(status.ID)
	require.NoError(t, err)
	assert.False(t, fetched.FinishedAt.IsZero())
	assert.Nil(t, fetched.ActiveUserID)
}

func TestHeartbeatInterval(t *testing.T) {
	assert.Equal(t, 30*time.Second, heartbeatInterval(24*time.Hour))
	assert.Equal(t, 30*time.Second, heartbeatInterval(5*time.Minute))
	assert.Equal(t, 6*time.Second, heartbeatInterval(time.Minute))
	assert.Equal(t, time.Second, heartbeatInterval(2*time.Second))
	assert.Equal(t, time.Second, heartbeatInterval(time.Millisecond))
}

func TestStartRunStopEndsTheHeartbeat(t *testing.T) {
	clearMigrationStatus(t)
	u1 := getTestUser(t, 1)

	config.MigrationClaimTimeout.Set("10s")
	t.Cleanup(func() { config.MigrationClaimTimeout.Set("5m") })

	status, err := ClaimMigration(&testMigrator{"todoist"}, u1)
	require.NoError(t, err)
	require.Nil(t, status.HeartbeatAt)

	stop := StartRun(status.ID)
	t.Cleanup(stop)

	var lastBeat time.Time
	require.Eventually(t, func() bool {
		fetched, err := GetMigrationStatusByID(status.ID)
		require.NoError(t, err)
		if fetched.HeartbeatAt == nil {
			return false
		}
		lastBeat = *fetched.HeartbeatAt
		return true
	}, 5*time.Second, 100*time.Millisecond, "the running migration never recorded a heartbeat")

	stop()
	afterStop, err := GetMigrationStatusByID(status.ID)
	require.NoError(t, err)
	require.NotNil(t, afterStop.HeartbeatAt)

	time.Sleep(3 * heartbeatInterval(10*time.Second))

	settled, err := GetMigrationStatusByID(status.ID)
	require.NoError(t, err)
	require.NotNil(t, settled.HeartbeatAt)
	assert.Equal(t, afterStop.HeartbeatAt.UnixNano(), settled.HeartbeatAt.UnixNano(), "stop() must end the ticker")
	assert.GreaterOrEqual(t, afterStop.HeartbeatAt.UnixNano(), lastBeat.UnixNano())
}

// Without stale release there is nothing to prove being alive to, so no beat is written.
func TestStartRunDoesNotBeatWithoutStaleRelease(t *testing.T) {
	clearMigrationStatus(t)
	u1 := getTestUser(t, 1)

	config.MigrationClaimTimeout.Set("0")
	t.Cleanup(func() { config.MigrationClaimTimeout.Set("5m") })

	status, err := ClaimMigration(&testMigrator{"todoist"}, u1)
	require.NoError(t, err)

	stop := StartRun(status.ID)
	time.Sleep(2 * minHeartbeatInterval)
	stop()

	fetched, err := GetMigrationStatusByID(status.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched.HeartbeatAt)
}
