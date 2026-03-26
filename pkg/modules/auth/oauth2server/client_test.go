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

func TestValidateRedirectURI(t *testing.T) {
	t.Run("accepts vikunja-flutter scheme", func(t *testing.T) {
		assert.True(t, ValidateRedirectURI("vikunja-flutter://callback"))
	})
	t.Run("accepts vikunja-desktop scheme", func(t *testing.T) {
		assert.True(t, ValidateRedirectURI("vikunja-desktop://auth"))
	})
	t.Run("rejects https scheme", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI("https://evil.com/callback"))
	})
	t.Run("rejects http scheme", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI("http://localhost/callback"))
	})
	t.Run("rejects javascript scheme", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI("javascript:alert(1)"))
	})
	t.Run("rejects data scheme", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI("data:text/html,<script>alert(1)</script>"))
	})
	t.Run("rejects non-vikunja custom scheme", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI("myapp://callback"))
	})
	t.Run("rejects empty URI", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI(""))
	})
}
