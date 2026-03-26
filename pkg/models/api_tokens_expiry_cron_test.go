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

	"github.com/stretchr/testify/require"
)

func TestCheckForExpiringAPITokens(t *testing.T) {
	t.Run("sends 7-day notification", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		notifications.Fake()

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

	t.Run("sends both 1-day and 7-day notification", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		notifications.Fake()

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

		notifications.AssertSent(t, &APITokenExpiringWeekNotification{})
		notifications.AssertSent(t, &APITokenExpiringDayNotification{})
	})

	t.Run("does not send for tokens expiring in 30 days", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		notifications.Fake()

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
}
