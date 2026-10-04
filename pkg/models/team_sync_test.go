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

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/require"
)

func TestMigrateExternalTeamIDs(t *testing.T) {
	t.Run("moves team to new id", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := MigrateExternalTeamIDs(s, "https://some.issuer", map[string]string{"14": "guid-14"})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertExists(t, "teams", map[string]any{"id": 14, "external_id": "guid-14"}, false)
		db.AssertExists(t, "teams", map[string]any{"id": 15, "external_id": "15"}, false)
	})
	t.Run("only touches teams of the issuer", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := MigrateExternalTeamIDs(s, "ldap", map[string]string{"14": "guid-14"})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertExists(t, "teams", map[string]any{"id": 14, "external_id": "14"}, false)
	})
	t.Run("keeps old id when new id is taken", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := MigrateExternalTeamIDs(s, "https://some.issuer", map[string]string{"14": "15"})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertExists(t, "teams", map[string]any{"id": 14, "external_id": "14"}, false)
		db.AssertExists(t, "teams", map[string]any{"id": 15, "external_id": "15"}, false)
	})
}
