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
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func insertExpiringToken(t *testing.T, ownerID int64, expiresAt time.Time) *APIToken {
	s := db.NewSession()
	defer s.Close()

	token := &APIToken{
		Title:          "Expiring token",
		TokenSalt:      "salt",
		TokenHash:      "hash" + expiresAt.String(),
		TokenLastEight: "12345678",
		APIPermissions: APIPermissions{"tasks": {"read"}},
		ExpiresAt:      expiresAt,
		OwnerID:        ownerID,
	}
	_, err := s.Insert(token)
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	return token
}

func expiryNotificationsFor(t *testing.T, userID int64, name string, tokenID int64) []*notifications.DatabaseNotification {
	s := db.NewSession()
	defer s.Close()

	got, err := notifications.GetNotificationsForNameAndUser(s, userID, name, tokenID)
	require.NoError(t, err)
	return got
}

func TestCheckForExpiringAPITokens(t *testing.T) {
	t.Run("sends 7-day notification", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		notifications.Fake()
		t.Cleanup(notifications.Unfake)

		now := time.Now()
		s := db.NewSession()
		defer s.Close()

		token := &APIToken{
			Title:          "Test 7-day token",
			TokenSalt:      "salt1",
			TokenHash:      "uniquehash7day",
			TokenLastEight: "test1234",
			APIPermissions: APIPermissions{"tasks": {"read"}},
			ExpiresAt:      now.Add(6 * 24 * time.Hour),
			OwnerID:        1,
		}
		_, err := s.Insert(token)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		checkForExpiringAPITokensAt(now)

		notifications.AssertSent(t, &APITokenExpiringWeekNotification{})
	})

	t.Run("sends only 1-day notification for token expiring within 24h", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		notifications.Fake()
		t.Cleanup(notifications.Unfake)

		now := time.Now()
		s := db.NewSession()
		defer s.Close()

		token := &APIToken{
			Title:          "Test 1-day token",
			TokenSalt:      "salt2",
			TokenHash:      "uniquehash1day",
			TokenLastEight: "test5678",
			APIPermissions: APIPermissions{"tasks": {"read"}},
			ExpiresAt:      now.Add(20 * time.Hour),
			OwnerID:        1,
		}
		_, err := s.Insert(token)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		checkForExpiringAPITokensAt(now)

		notifications.AssertSent(t, &APITokenExpiringDayNotification{})
		notifications.AssertNotSent(t, &APITokenExpiringWeekNotification{})
	})

	t.Run("does not send for tokens expiring in 30 days", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		notifications.Fake()
		t.Cleanup(notifications.Unfake)

		now := time.Now()
		s := db.NewSession()
		defer s.Close()

		token := &APIToken{
			Title:          "Far future token",
			TokenSalt:      "salt3",
			TokenHash:      "uniquehash30day",
			TokenLastEight: "test9012",
			APIPermissions: APIPermissions{"tasks": {"read"}},
			ExpiresAt:      now.Add(30 * 24 * time.Hour),
			OwnerID:        1,
		}
		_, err := s.Insert(token)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		checkForExpiringAPITokensAt(now)

		// The existing fixture tokens expire in 2099, so no notifications should be sent
		// for our 30-day token either
		notifications.AssertNotSent(t, &APITokenExpiringWeekNotification{})
		notifications.AssertNotSent(t, &APITokenExpiringDayNotification{})
	})

	t.Run("does not send for already expired tokens", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		notifications.Fake()
		t.Cleanup(notifications.Unfake)

		now := time.Now()
		s := db.NewSession()
		defer s.Close()

		token := &APIToken{
			Title:          "Expired token",
			TokenSalt:      "salt4",
			TokenHash:      "uniquehashexpired",
			TokenLastEight: "testexp1",
			APIPermissions: APIPermissions{"tasks": {"read"}},
			ExpiresAt:      now.Add(-24 * time.Hour),
			OwnerID:        1,
		}
		_, err := s.Insert(token)
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		checkForExpiringAPITokensAt(now)

		notifications.AssertNotSent(t, &APITokenExpiringWeekNotification{})
		notifications.AssertNotSent(t, &APITokenExpiringDayNotification{})
	})

	// Bot 23 is owned by user 21, see users.yml.
	const botID, botOwnerID = 23, 21

	t.Run("bot token notifies the bot owner, not the bot", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		now := time.Now()
		token := insertExpiringToken(t, botID, now.Add(6*24*time.Hour))

		checkForExpiringAPITokensAt(now)

		n := &APITokenExpiringWeekNotification{}
		assert.Len(t, expiryNotificationsFor(t, botOwnerID, n.Name(), token.ID), 1)
		assert.Empty(t, expiryNotificationsFor(t, botID, n.Name(), token.ID))
	})

	t.Run("bot token names the bot in the notification", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		now := time.Now()
		token := insertExpiringToken(t, botID, now.Add(20*time.Hour))

		checkForExpiringAPITokensAt(now)

		n := &APITokenExpiringDayNotification{}
		got := expiryNotificationsFor(t, botOwnerID, n.Name(), token.ID)
		require.Len(t, got, 1)
		payload, ok := got[0].Notification.(map[string]interface{})
		require.True(t, ok)
		bot, ok := payload["bot"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "bot-owner-a-assistant", bot["username"])
	})

	t.Run("human token still notifies its own owner", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		now := time.Now()
		token := insertExpiringToken(t, botOwnerID, now.Add(6*24*time.Hour))

		checkForExpiringAPITokensAt(now)

		n := &APITokenExpiringWeekNotification{}
		got := expiryNotificationsFor(t, botOwnerID, n.Name(), token.ID)
		require.Len(t, got, 1)
		payload, ok := got[0].Notification.(map[string]interface{})
		require.True(t, ok)
		assert.NotContains(t, payload, "bot")
	})

	// Fixture 26 is the instance bot; 1 and 17 (disabled) get promoted per test.
	const instanceBotID = 26

	promote := func(t *testing.T, ids ...int64) {
		s := db.NewSession()
		defer s.Close()
		for _, id := range ids {
			_, err := s.ID(id).Cols("is_admin").Update(&user.User{IsAdmin: true})
			require.NoError(t, err)
		}
		require.NoError(t, s.Commit())
	}

	t.Run("instance bot token notifies every active human admin", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		promote(t, 1, 2, 17)

		now := time.Now()
		token := insertExpiringToken(t, instanceBotID, now.Add(6*24*time.Hour))

		checkForExpiringAPITokensAt(now)

		n := &APITokenExpiringWeekNotification{}
		for _, adminID := range []int64{1, 2} {
			got := expiryNotificationsFor(t, adminID, n.Name(), token.ID)
			require.Len(t, got, 1, "admin %d", adminID)
			payload, ok := got[0].Notification.(map[string]interface{})
			require.True(t, ok)
			bot, ok := payload["bot"].(map[string]interface{})
			require.True(t, ok)
			assert.Equal(t, "bot-instance-provisioner", bot["username"])
		}
		assert.Empty(t, expiryNotificationsFor(t, 17, n.Name(), token.ID), "disabled admins cannot act on it")
		assert.Empty(t, expiryNotificationsFor(t, instanceBotID, n.Name(), token.ID))
	})

	t.Run("instance bot token is not sent to non-admins or owned-bot owners", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		promote(t, 1)

		now := time.Now()
		token := insertExpiringToken(t, instanceBotID, now.Add(6*24*time.Hour))

		checkForExpiringAPITokensAt(now)

		n := &APITokenExpiringWeekNotification{}
		assert.Len(t, expiryNotificationsFor(t, 1, n.Name(), token.ID), 1)
		assert.Empty(t, expiryNotificationsFor(t, botOwnerID, n.Name(), token.ID))
	})

	t.Run("bot token notification is not repeated", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		now := time.Now()
		token := insertExpiringToken(t, botID, now.Add(6*24*time.Hour))

		checkForExpiringAPITokensAt(now)
		checkForExpiringAPITokensAt(now.Add(time.Hour))

		n := &APITokenExpiringWeekNotification{}
		assert.Len(t, expiryNotificationsFor(t, botOwnerID, n.Name(), token.ID), 1)
	})
}
