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

package oauth2server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateClient(t *testing.T) {
	assert.True(t, ValidateClient("vikunja-flutter"))
	assert.False(t, ValidateClient("unknown-client"))
	assert.False(t, ValidateClient(""))
}

func TestValidateRedirectURI(t *testing.T) {
	assert.True(t, ValidateRedirectURI("vikunja-flutter", "vikunja://callback"))
	assert.False(t, ValidateRedirectURI("vikunja-flutter", "https://evil.com/callback"))
	assert.False(t, ValidateRedirectURI("unknown-client", "vikunja://callback"))
}
