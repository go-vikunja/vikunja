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
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/utils"
)

const oauthStateTTL = 10 * time.Minute

func oauthStateKey(m MigratorName, state string) string {
	return "migration_oauth_state_" + m.Name() + "_" + state
}

// Without a user-bound state, a victim lured to a callback with someone else's code imports that account.
func AuthURLForUser(m OAuthMigrator, userID int64) (string, error) {
	state, err := utils.CryptoRandomString(32)
	if err != nil {
		return "", err
	}
	// Not an int64: the redis backend stores ints raw but GetWithValue gob-decodes.
	if err := keyvalue.PutWithTTL(oauthStateKey(m, state), strconv.FormatInt(userID, 10), oauthStateTTL); err != nil {
		return "", err
	}
	return m.AuthURL(state), nil
}

func ConsumeOAuthState(m OAuthMigrator, userID int64) error {
	state := m.OAuthState()
	if state == "" {
		return &ErrInvalidOAuthState{}
	}
	key := oauthStateKey(m, state)
	var owner string
	exists, err := keyvalue.GetWithValue(key, &owner)
	if err != nil {
		return err
	}
	if !exists || owner != strconv.FormatInt(userID, 10) {
		return &ErrInvalidOAuthState{}
	}
	return keyvalue.Del(key)
}
