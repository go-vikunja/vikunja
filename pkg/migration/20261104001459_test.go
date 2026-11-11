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

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/require"
)

func TestMigrationUserIdentities20261104001459(t *testing.T) {
	engine, err := db.CreateTestEngine()
	require.NoError(t, err)
	table := userIdentities20261104001459{}
	require.NoError(t, engine.DropTables(table))
	require.NoError(t, createUserIdentities20261104001459(engine))
	t.Cleanup(func() {
		require.NoError(t, engine.DropTables(table))
	})

	insert := func(userID int64, issuer, subject string) error {
		_, err := engine.Insert(&userIdentities20261104001459{
			UserID:  userID,
			Issuer:  issuer,
			Subject: subject,
		})
		return err
	}

	// Max lengths must fit the composite unique index on MySQL.
	require.NoError(t, insert(1, strings.Repeat("i", 250), strings.Repeat("s", 255)))
	require.NoError(t, insert(1, "https://a.example.com", "sub-1"))
	require.NoError(t, insert(2, "https://b.example.com", "sub-1"))

	require.Error(t, insert(2, "https://a.example.com", "sub-1"), "one IdP account must belong to one user only")
	require.Error(t, insert(1, "https://a.example.com", "sub-2"), "a user must have at most one identity per issuer")
	require.NoError(t, insert(3, "https://a.example.com", "SUB-1"), "subjects are case-sensitive")
	require.NoError(t, insert(1, "https://A.example.com", "sub-1"), "issuers are case-sensitive")
}
