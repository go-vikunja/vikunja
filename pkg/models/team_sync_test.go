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
	"xorm.io/builder"
)

func TestMigrateExternalTeamIDs(t *testing.T) {
	t.Run("moves team to new id", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := MigrateExternalTeamIDs(s, "https://some.issuer", map[string]string{"14": "guid-14"})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertExists(t, "teams", map[string]any{
			"id":          14,
			"external_id": "guid-14",
		}, false)
		db.AssertExists(t, "teams", map[string]any{
			"id":          15,
			"external_id": "15",
		}, false)
	})
	t.Run("only touches teams of the issuer", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := MigrateExternalTeamIDs(s, "ldap", map[string]string{"14": "guid-14"})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertExists(t, "teams", map[string]any{
			"id":          14,
			"external_id": "14",
		}, false)
	})
	t.Run("keeps old id when new id is taken", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := MigrateExternalTeamIDs(s, "https://some.issuer", map[string]string{"14": "15"})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertExists(t, "teams", map[string]any{
			"id":          14,
			"external_id": "14",
		}, false)
		db.AssertExists(t, "teams", map[string]any{
			"id":          15,
			"external_id": "15",
		}, false)
	})
	t.Run("does nothing for an empty map", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := MigrateExternalTeamIDs(s, "https://some.issuer", map[string]string{})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertExists(t, "teams", map[string]any{
			"id":          14,
			"external_id": "14",
		}, false)
		db.AssertExists(t, "teams", map[string]any{
			"id":          15,
			"external_id": "15",
		}, false)
	})
	t.Run("ignores the new id being taken by another issuer", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		other := &Team{
			Name:        "other issuer",
			CreatedByID: 7,
			ExternalID:  "guid-14",
			Issuer:      "ldap",
		}
		_, err := s.Insert(other)
		require.NoError(t, err)

		err = MigrateExternalTeamIDs(s, "https://some.issuer", map[string]string{"14": "guid-14"})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertExists(t, "teams", map[string]any{
			"id":          14,
			"external_id": "guid-14",
		}, false)
		db.AssertExists(t, "teams", map[string]any{
			"id":          other.ID,
			"external_id": "guid-14",
			"issuer":      "ldap",
		}, false)
	})
	t.Run("moves only one team when two old ids map to the same new id", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := MigrateExternalTeamIDs(s, "https://some.issuer", map[string]string{
			"14": "guid",
			"15": "guid",
		})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertCount(t, "teams", builder.Eq{
			"issuer":      "https://some.issuer",
			"external_id": "guid",
		}, 1)
		db.AssertCount(t, "teams", builder.Or(
			builder.Eq{
				"id":          14,
				"external_id": "14",
			},
			builder.Eq{
				"id":          15,
				"external_id": "15",
			},
		), 1)
	})
	t.Run("moves the free id when another one is taken", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		free := &Team{
			Name:        "free",
			CreatedByID: 7,
			ExternalID:  "old",
			Issuer:      "https://some.issuer",
		}
		_, err := s.Insert(free)
		require.NoError(t, err)

		err = MigrateExternalTeamIDs(s, "https://some.issuer", map[string]string{
			"14":  "15",
			"old": "new",
		})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		db.AssertExists(t, "teams", map[string]any{
			"id":          14,
			"external_id": "14",
		}, false)
		db.AssertExists(t, "teams", map[string]any{
			"id":          15,
			"external_id": "15",
		}, false)
		db.AssertExists(t, "teams", map[string]any{
			"id":          free.ID,
			"external_id": "new",
		}, false)
	})
}
