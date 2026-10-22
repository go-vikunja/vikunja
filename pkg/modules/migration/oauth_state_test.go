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
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stateMigrator struct {
	State string
}

func (m *stateMigrator) Name() string               { return "state-test" }
func (m *stateMigrator) Migrate(_ *user.User) error { return nil }
func (m *stateMigrator) AuthURL(state string) string {
	return "https://example.com/auth?state=" + state
}
func (m *stateMigrator) OAuthState() string { return m.State }

func issueState(t *testing.T) string {
	t.Helper()
	u, err := AuthURLForUser(&stateMigrator{}, 1)
	require.NoError(t, err)
	state := u[strings.Index(u, "state=")+len("state="):]
	require.NotEmpty(t, state)
	return state
}

func TestOAuthState(t *testing.T) {
	var invalid *ErrInvalidOAuthState

	t.Run("accepted once for the user it was issued to", func(t *testing.T) {
		state := issueState(t)
		require.NoError(t, ConsumeOAuthState(&stateMigrator{State: state}, 1))
		require.ErrorAs(t, ConsumeOAuthState(&stateMigrator{State: state}, 1), &invalid, "a state must be single-use")
	})

	t.Run("rejected for another user", func(t *testing.T) {
		state := issueState(t)
		require.ErrorAs(t, ConsumeOAuthState(&stateMigrator{State: state}, 2), &invalid)
		assert.NoError(t, ConsumeOAuthState(&stateMigrator{State: state}, 1), "a foreign attempt must not burn the owner's state")
	})

	t.Run("rejected when missing or unknown", func(t *testing.T) {
		require.ErrorAs(t, ConsumeOAuthState(&stateMigrator{}, 1), &invalid)
		require.ErrorAs(t, ConsumeOAuthState(&stateMigrator{State: "forged"}, 1), &invalid)
	})

	t.Run("stored as a type the redis backend gob-encodes", func(t *testing.T) {
		// redis writes ints raw, which GetWithValue can't gob-decode
		stored, exists, err := keyvalue.Get(oauthStateKey(&stateMigrator{}, issueState(t)))
		require.NoError(t, err)
		require.True(t, exists)
		assert.IsType(t, "", stored)
	})

	t.Run("each auth url gets its own state", func(t *testing.T) {
		first, second := issueState(t), issueState(t)
		assert.NotEqual(t, first, second)
	})
}
