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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateEmailStatusPersistence(t *testing.T) {
	t.Run("mailer enabled", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		config.MailerEnabled.Set(true)
		defer config.MailerEnabled.Set(false)

		err := UpdateEmail(s, &EmailUpdate{User: &User{ID: 1}, NewEmail: "new1@example.com"})
		require.NoError(t, err)

		err = s.Commit()
		require.NoError(t, err)

		s2 := db.NewSession()
		defer s2.Close()
		updated, err := GetUserWithEmail(s2, &User{ID: 1})
		require.NoError(t, err)
		assert.Equal(t, StatusEmailConfirmationRequired, updated.Status)
		assert.Equal(t, "new1@example.com", updated.Email)
	})

	t.Run("mailer disabled", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		config.MailerEnabled.Set(false)

		err := UpdateEmail(s, &EmailUpdate{User: &User{ID: 2}, NewEmail: "new2@example.com"})
		require.NoError(t, err)

		err = s.Commit()
		require.NoError(t, err)

		s2 := db.NewSession()
		defer s2.Close()
		updated, err := GetUserWithEmail(s2, &User{ID: 2})
		require.NoError(t, err)
		assert.Equal(t, StatusActive, updated.Status)
		assert.Equal(t, "new2@example.com", updated.Email)
	})
}
