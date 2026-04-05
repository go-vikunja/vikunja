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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrAccountIsBot(t *testing.T) {
	err := &ErrAccountIsBot{UserID: 42}
	assert.True(t, IsErrAccountIsBot(err))
	assert.Equal(t, http.StatusPreconditionFailed, err.HTTPError().HTTPCode)
	assert.Equal(t, 1031, err.HTTPError().Code)
}

func TestErrBotUsersDisabled(t *testing.T) {
	err := &ErrBotUsersDisabled{}
	assert.True(t, IsErrBotUsersDisabled(err))
	assert.Equal(t, http.StatusForbidden, err.HTTPError().HTTPCode)
	assert.Equal(t, 1032, err.HTTPError().Code)
}

func TestErrBotNotOwned(t *testing.T) {
	err := &ErrBotNotOwned{UserID: 7}
	assert.True(t, IsErrBotNotOwned(err))
	assert.Equal(t, http.StatusForbidden, err.HTTPError().HTTPCode)
	assert.Equal(t, 1033, err.HTTPError().Code)
}

func TestErrBotUsernameMustHavePrefix(t *testing.T) {
	err := &ErrBotUsernameMustHavePrefix{Username: "not-a-bot"}
	assert.True(t, IsErrBotUsernameMustHavePrefix(err))
	assert.Equal(t, http.StatusBadRequest, err.HTTPError().HTTPCode)
	assert.Equal(t, 1034, err.HTTPError().Code)
}
