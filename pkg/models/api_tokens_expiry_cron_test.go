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
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/mail"
	"code.vikunja.io/api/pkg/notifications"
	"xorm.io/xorm/contexts"

	"github.com/stretchr/testify/require"
)

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
}

type tokenNotificationInsertFailure struct{ enabled atomic.Bool }

func (h *tokenNotificationInsertFailure) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
	if h.enabled.Load() && strings.HasPrefix(c.SQL, "INSERT") && strings.Contains(c.SQL, "notifications") {
		return c.Ctx, errors.New("notification insert failed")
	}
	return c.Ctx, nil
}

func (*tokenNotificationInsertFailure) AfterProcess(c *contexts.ContextHook) error { return c.Err }

func expiringTokenForAtomicityTest(t *testing.T) (*APIToken, time.Time) {
	t.Helper()
	db.LoadAndAssertFixtures(t)
	notifications.Unfake()
	mail.Fake()
	t.Cleanup(mail.ResetSent)
	enabled := config.MailerEnabled.GetBool()
	config.MailerEnabled.Set(true)
	t.Cleanup(func() { config.MailerEnabled.Set(enabled) })
	now := time.Now().UTC().Truncate(time.Second)
	token := &APIToken{Title: t.Name(), TokenHash: t.Name(), OwnerID: 1, ExpiresAt: now.Add(6 * 24 * time.Hour), APIPermissions: APIPermissions{"tasks": {"read"}}}
	s := db.NewSession()
	defer s.Close()
	_, err := s.Insert(token)
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	return token, now
}

func TestTokenExpiryRollsBackMailWhenNotificationFails(t *testing.T) {
	token, now := expiringTokenForAtomicityTest(t)
	engine, err := db.CreateDBEngine()
	require.NoError(t, err)
	hook := &tokenNotificationInsertFailure{}
	hook.enabled.Store(true)
	t.Cleanup(func() { hook.enabled.Store(false) })
	engine.AddHook(hook)
	checkForExpiringAPITokensAt(now)
	require.Empty(t, mail.SentMails(), "mail must not commit when its notification marker fails")
	hook.enabled.Store(false)
	checkForExpiringAPITokensAt(now.Add(time.Hour))
	require.Len(t, mail.SentMails(), 1, "rollback must leave the warning eligible for the next check")
	s := db.NewSession()
	defer s.Close()
	rows, err := notifications.GetNotificationsForNameAndUser(s, 1, (&APITokenExpiringWeekNotification{}).Name(), token.ID)
	require.NoError(t, err)
	require.Len(t, rows, 1)
}
