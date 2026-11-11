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

	"github.com/asaskevich/govalidator"
	"github.com/stretchr/testify/assert"
)

func TestBotStatusValidator(t *testing.T) {
	type body struct {
		Status  Status  `valid:"bot_status"`
		Pointer *Status `valid:"bot_status"`
	}
	valid := func(s Status, p *Status) bool {
		ok, _ := govalidator.ValidateStruct(body{Status: s, Pointer: p})
		return ok
	}

	assert.True(t, valid(StatusActive, nil))
	assert.True(t, valid(StatusDisabled, new(StatusActive)))
	assert.False(t, valid(StatusAccountLocked, nil))
	assert.False(t, valid(StatusEmailConfirmationRequired, nil))
	assert.False(t, valid(StatusActive, new(StatusAccountLocked)))
}
