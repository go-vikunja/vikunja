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
	"testing"

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdentitySubjectIsCaseSensitive(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	const issuer = "https://case.service.com"
	_, err := s.Insert(&Identity{
		UserID:  1,
		Issuer:  issuer,
		Subject: "Alice",
	})
	require.NoError(t, err)
	_, err = s.Insert(&Identity{
		UserID:  2,
		Issuer:  issuer,
		Subject: "alice",
	})
	require.NoError(t, err, "subjects differing only in case are distinct identities")

	upper, found, err := GetIdentityByIssuerAndSubject(s, issuer, "Alice")
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, int64(1), upper.UserID)

	lower, found, err := GetIdentityByIssuerAndSubject(s, issuer, "alice")
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, int64(2), lower.UserID)

	_, found, err = GetIdentityByIssuerAndSubject(s, "https://CASE.service.com", "Alice")
	require.NoError(t, err)
	assert.False(t, found, "issuers are compared case-sensitively too")

	_, found, err = GetIdentityByIssuerAndSubject(s, issuer, "Alice ")
	require.NoError(t, err)
	assert.False(t, found, "trailing spaces are significant")
}

func TestGetIdentitiesByUserID(t *testing.T) {
	t.Run("user with identities", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		identities, err := GetIdentitiesByUserID(s, 1)
		require.NoError(t, err)
		require.Len(t, identities, 2)
		assert.Equal(t, int64(1), identities[0].ID)
		assert.Equal(t, int64(2), identities[1].ID)
		for _, identity := range identities {
			assert.Equal(t, int64(1), identity.UserID)
		}
	})
	t.Run("user without identities", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		identities, err := GetIdentitiesByUserID(s, 2)
		require.NoError(t, err)
		assert.Empty(t, identities)
	})
}
