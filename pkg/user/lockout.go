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

package user

import (
	"context"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/notifications"
)

const maxFailedLoginAttempts = 10

// lockAccount locks the user, issues a password reset token and sends n.
//
// Must not share the caller's session: login handlers roll back on auth
// failure, which would discard the lockout (GHSA-fgfv-pv97-6cmj).
func lockAccount(user *User, attemptsKey string, n notifications.Notification) {
	log.Infof("Locking user account %d after %d failed login attempts", user.ID, maxFailedLoginAttempts)
	s := db.NewSession()
	defer s.Close()
	defer events.CleanupPending(s)

	if err := RequestUserPasswordResetToken(s, user); err != nil {
		log.Errorf("Could not issue password reset token for user %d during lockout: %s", user.ID, err)
		_ = s.Rollback()
		return
	}
	if err := notifications.Notify(user, n, s); err != nil {
		log.Errorf("Could not send lockout mail to user %d: %s", user.ID, err)
		_ = s.Rollback()
		return
	}
	if err := user.SetStatus(s, StatusAccountLocked); err != nil {
		log.Errorf("Could not lock user %d: %s", user.ID, err)
		_ = s.Rollback()
		return
	}
	events.DispatchOnCommit(s, &AccountLockedEvent{UserID: user.ID})
	if err := s.Commit(); err != nil {
		log.Errorf("Could not commit lockout for user %d: %s", user.ID, err)
		return
	}
	events.DispatchPending(context.Background(), s)

	// Otherwise the first typo after a password reset relocks the account.
	if err := keyvalue.Del(attemptsKey); err != nil {
		log.Errorf("Could not reset failed attempts for user %d after lockout: %s", user.ID, err)
	}
}
