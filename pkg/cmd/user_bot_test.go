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

package cmd

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runBotCmd executes `vikunja user bot ...` against the test DB. FullInit is
// skipped: it would read the config and migrate a real database.
func runBotCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	db.LoadAndAssertFixtures(t)

	for _, c := range []*cobra.Command{userBotCreateCmd, userBotListCmd, userBotDeleteCmd, userBotTokenCreateCmd, userBotTokenListCmd, userBotTokenRevokeCmd} {
		c.PreRun = nil
	}
	// Flag vars persist across Execute calls.
	botFlagAdmin, botFlagScopes, botFlagPreset, botFlagExpires, botFlagTitle = false, "", "", botDefaultExpiry, ""

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs(append([]string{"user", "bot"}, args...))
	err := rootCmd.Execute()
	return out.String(), err
}

func TestUserBotCreate(t *testing.T) {
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	t.Cleanup(license.ResetForTests)

	t.Run("prints the token on the last line", func(t *testing.T) {
		out, err := runBotCmd(t, "create", "bot-ci", "--admin", "--scopes", "admin:users_list", "--preset", "provisioning", "--expires", "90d", "--title", "ci")
		require.NoError(t, err, out)

		lines := strings.Split(strings.TrimSpace(out), "\n")
		token := lines[len(lines)-1]
		assert.True(t, strings.HasPrefix(token, models.APITokenPrefix), out)
		assert.Contains(t, out, `Created instance admin bot "bot-ci"`)

		s := db.NewSession()
		defer s.Close()
		stored, owner, err := models.ValidateTokenAndGetOwner(s, token)
		require.NoError(t, err)
		require.NotNil(t, stored, "printed token must authenticate")
		assert.True(t, owner.IsInstanceBot)
		assert.True(t, owner.IsAdmin)
		assert.Equal(t, "ci", stored.Title)
		assert.ElementsMatch(t, []string{"users_list", "users_create", "users_set_status", "users_delete"}, stored.APIPermissions["admin"])
		assert.WithinDuration(t, time.Now().AddDate(0, 0, 90), stored.ExpiresAt, time.Minute)
	})

	t.Run("without --admin", func(t *testing.T) {
		_, err := runBotCmd(t, "create", "bot-ci", "--scopes", "admin:users_list")
		require.ErrorContains(t, err, "--admin")
	})

	t.Run("non-admin scope", func(t *testing.T) {
		_, err := runBotCmd(t, "create", "bot-ci", "--admin", "--scopes", "tasks:read_all")
		require.Error(t, err)
		assert.True(t, models.IsErrInstanceBotScopeNotAllowed(err), err.Error())
	})

	t.Run("no scopes", func(t *testing.T) {
		_, err := runBotCmd(t, "create", "bot-ci", "--admin")
		require.ErrorContains(t, err, "at least one scope")
	})

	t.Run("license off", func(t *testing.T) {
		license.ResetForTests()
		defer license.SetForTests([]license.Feature{license.FeatureAdminPanel})
		_, err := runBotCmd(t, "create", "bot-ci", "--admin", "--scopes", "admin:users_list")
		require.ErrorContains(t, err, "license")
	})
}

func TestUserBotTokenAndLifecycle(t *testing.T) {
	license.SetForTests([]license.Feature{license.FeatureAdminPanel})
	t.Cleanup(license.ResetForTests)

	t.Run("token create for fixture bot", func(t *testing.T) {
		out, err := runBotCmd(t, "token", "create", "bot-instance-provisioner", "--scopes", "admin:users_list", "--expires", "2099-01-01T00:00:00Z")
		require.NoError(t, err, out)
		lines := strings.Split(strings.TrimSpace(out), "\n")
		assert.True(t, strings.HasPrefix(lines[len(lines)-1], models.APITokenPrefix), out)
	})

	t.Run("token create refuses owned bots", func(t *testing.T) {
		_, err := runBotCmd(t, "token", "create", "bot-owner-a-assistant", "--scopes", "admin:users_list")
		require.ErrorContains(t, err, "not an instance bot")
	})

	t.Run("token list and revoke", func(t *testing.T) {
		out, err := runBotCmd(t, "token", "list", "bot-instance-provisioner")
		require.NoError(t, err)
		assert.Contains(t, out, "instance bot admin token")

		out, err = runBotCmd(t, "token", "revoke", "10")
		require.NoError(t, err, out)

		s := db.NewSession()
		defer s.Close()
		exists, err := s.ID(10).Exist(&models.APIToken{})
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("revoke refuses human tokens", func(t *testing.T) {
		_, err := runBotCmd(t, "token", "revoke", "1")
		require.ErrorContains(t, err, "not belong to an instance bot")
	})

	t.Run("list and delete", func(t *testing.T) {
		out, err := runBotCmd(t, "list")
		require.NoError(t, err)
		assert.Contains(t, out, "bot-instance-provisioner")
		assert.NotContains(t, out, "bot-owner-a-assistant")

		out, err = runBotCmd(t, "delete", "bot-instance-provisioner")
		require.NoError(t, err, out)

		s := db.NewSession()
		defer s.Close()
		exists, err := s.ID(26).Exist(&user.User{})
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestParseBotExpiry(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)

	got, err := parseBotExpiry("90d", now)
	require.NoError(t, err)
	assert.Equal(t, now.AddDate(0, 0, 90), got)

	got, err = parseBotExpiry("1y", now)
	require.NoError(t, err)
	assert.Equal(t, now.AddDate(1, 0, 0), got)

	got, err = parseBotExpiry("2030-01-02T03:04:05Z", now)
	require.NoError(t, err)
	assert.Equal(t, time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC), got)

	for _, bad := range []string{"", "never", "0d", "-1y", "1w", "2020-01-01T00:00:00Z"} {
		_, err := parseBotExpiry(bad, now)
		assert.Error(t, err, bad)
	}
}

func TestParseBotScopes(t *testing.T) {
	perms, err := parseBotScopes("admin:users_list, admin:users_list", "provisioning")
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"users_list", "users_create", "users_set_status", "users_delete"}, perms["admin"])

	_, err = parseBotScopes("users_list", "")
	require.ErrorContains(t, err, "group:permission")

	_, err = parseBotScopes("", "nope")
	require.ErrorContains(t, err, "unknown preset")
}
