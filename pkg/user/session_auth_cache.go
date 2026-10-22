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
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/modules/keyvalue"

	"xorm.io/xorm"
)

// Bounds how long a revocation that skipped InvalidateSessionAuthCache (expiry cron, direct db edits) stays unnoticed.
const SessionAuthCacheTTL = 30 * time.Second

// Never expires: if it reset, a counter climbing back to an old value would revive that value's entries.
func sessionAuthGenerationKey(userID int64) string {
	return "session_auth_gen_" + strconv.FormatInt(userID, 10)
}

// Build it before the db check, so an invalidation during the check orphans the entry.
func SessionAuthCacheKey(userID int64, sessionID string) (string, error) {
	gen, _, err := keyvalue.GetInt64(sessionAuthGenerationKey(userID))
	if err != nil {
		return "", err
	}
	return "session_auth_ok_" + strconv.FormatInt(userID, 10) + "_" + strconv.FormatInt(gen, 10) + "_" + sessionID, nil
}

// Inside a transaction use InvalidateSessionAuthCacheOnCommit instead.
func InvalidateSessionAuthCache(userID int64) error {
	return keyvalue.IncrBy(sessionAuthGenerationKey(userID), 1)
}

// Invalidates again on commit: a request racing the transaction could re-cache the old state.
func InvalidateSessionAuthCacheOnCommit(s *xorm.Session, userID int64) error {
	if err := InvalidateSessionAuthCache(userID); err != nil {
		return err
	}
	events.OnCommit(s, func() error { return InvalidateSessionAuthCache(userID) })
	return nil
}
