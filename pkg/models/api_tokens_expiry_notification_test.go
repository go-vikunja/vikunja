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
	"time"

	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
)

func TestAPITokenExpiringWeekNotification(t *testing.T) {
	u := &user.User{ID: 1, Name: "Test User"}
	token := &APIToken{ID: 42, Title: "My Token", ExpiresAt: time.Now().Add(7 * 24 * time.Hour)}

	n := &APITokenExpiringWeekNotification{User: u, Token: token}

	t.Run("Name", func(t *testing.T) {
		assert.Equal(t, "api_token.expiring.week", n.Name())
	})

	t.Run("SubjectID", func(t *testing.T) {
		assert.Equal(t, int64(42), n.SubjectID())
	})

	t.Run("ToDB", func(t *testing.T) {
		assert.NotNil(t, n.ToDB())
	})

	t.Run("ToMail", func(t *testing.T) {
		mail := n.ToMail("en")
		assert.NotNil(t, mail)
	})
}

func TestAPITokenExpiringDayNotification(t *testing.T) {
	u := &user.User{ID: 1, Name: "Test User"}
	token := &APIToken{ID: 99, Title: "CI Token", ExpiresAt: time.Now().Add(24 * time.Hour)}

	n := &APITokenExpiringDayNotification{User: u, Token: token}

	t.Run("Name", func(t *testing.T) {
		assert.Equal(t, "api_token.expiring.day", n.Name())
	})

	t.Run("SubjectID", func(t *testing.T) {
		assert.Equal(t, int64(99), n.SubjectID())
	})

	t.Run("ToDB", func(t *testing.T) {
		assert.NotNil(t, n.ToDB())
	})

	t.Run("ToMail", func(t *testing.T) {
		mail := n.ToMail("en")
		assert.NotNil(t, mail)
	})
}
