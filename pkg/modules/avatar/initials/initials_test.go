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

package initials

import (
	"testing"

	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAvatar(t *testing.T) {
	t.Run("generates valid SVG with name", func(t *testing.T) {
		provider := &Provider{}

		testUser := &user.User{
			ID:       1,
			Name:     "John Doe",
			Username: "johndoe",
		}

		avatar, mimeType, err := provider.GetAvatar(testUser, 128)
		require.NoError(t, err)
		assert.NotNil(t, avatar)
		assert.Equal(t, "image/svg+xml", mimeType)

		// Verify it's valid SVG
		svgString := string(avatar)
		assert.Contains(t, svgString, "<svg")
		assert.Contains(t, svgString, "viewBox=\"0 0 100 100\"")
		assert.Contains(t, svgString, "width=\"128\"")
		assert.Contains(t, svgString, "height=\"128\"")
		// Should contain "J" from "John"
		assert.Contains(t, svgString, ">J<")
		// Should have a background color
		assert.Contains(t, svgString, "fill=\"#")
		// Should have bold font weight
		assert.Contains(t, svgString, "font-weight=\"bold\"")
	})

	t.Run("generates valid SVG with username when name is empty", func(t *testing.T) {
		provider := &Provider{}

		testUser := &user.User{
			ID:       2,
			Name:     "", // Empty name
			Username: "jane_smith",
		}

		avatar, mimeType, err := provider.GetAvatar(testUser, 64)
		require.NoError(t, err)
		assert.NotNil(t, avatar)
		assert.Equal(t, "image/svg+xml", mimeType)

		// Verify it's valid SVG
		svgString := string(avatar)
		assert.Contains(t, svgString, "<svg")
		assert.Contains(t, svgString, "width=\"64\"")
		assert.Contains(t, svgString, "height=\"64\"")
		// Should contain "J" from "jane_smith"
		assert.Contains(t, svgString, ">J<")
	})

	t.Run("uses consistent colors based on user ID", func(t *testing.T) {
		provider := &Provider{}

		testUser := &user.User{
			ID:       0,
			Name:     "Alice",
			Username: "alice",
		}

		avatar1, _, err := provider.GetAvatar(testUser, 100)
		require.NoError(t, err)

		avatar2, _, err := provider.GetAvatar(testUser, 200)
		require.NoError(t, err)

		// Both should use the same colors (user ID 0 -> index 0)
		svg1 := string(avatar1)
		svg2 := string(avatar2)

		// Should have the first background color (index 0)
		assert.Contains(t, svg1, `fill="#e0f8d9"`)
		assert.Contains(t, svg2, `fill="#e0f8d9"`)
		// Should have the first text color (index 0)
		assert.Contains(t, svg1, `fill="#005f00"`)
		assert.Contains(t, svg2, `fill="#005f00"`)
	})

	t.Run("escapes special characters", func(t *testing.T) {
		provider := &Provider{}

		testUser := &user.User{
			ID:       3,
			Name:     "<script>",
			Username: "hacker",
		}

		avatar, mimeType, err := provider.GetAvatar(testUser, 50)
		require.NoError(t, err)
		assert.Equal(t, "image/svg+xml", mimeType)

		svgString := string(avatar)
		// Should escape the < character
		assert.NotContains(t, svgString, "><script><")
		assert.Contains(t, svgString, "&lt;")
	})

	t.Run("handles different sizes", func(t *testing.T) {
		provider := &Provider{}

		testUser := &user.User{
			ID:       4,
			Name:     "Bob",
			Username: "bob",
		}

		testCases := []struct {
			size         int64
			expectedSize string
		}{
			{32, "32"},
			{64, "64"},
			{128, "128"},
			{256, "256"},
			{512, "512"},
		}

		for _, tc := range testCases {
			avatar, mimeType, err := provider.GetAvatar(testUser, tc.size)
			require.NoError(t, err)
			assert.Equal(t, "image/svg+xml", mimeType)

			svgString := string(avatar)
			assert.Contains(t, svgString, `width="`+tc.expectedSize+`"`)
			assert.Contains(t, svgString, `height="`+tc.expectedSize+`"`)
			assert.Contains(t, svgString, ">B<")
		}
	})

	t.Run("returns error for user without name or username", func(t *testing.T) {
		provider := &Provider{}

		testUser := &user.User{
			ID:       5,
			Name:     "",
			Username: "",
		}

		_, _, err := provider.GetAvatar(testUser, 100)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no name or username")
	})
}

func TestFlushCache(t *testing.T) {
	provider := &Provider{}
	testUser := &user.User{
		ID:       999,
		Name:     "Test",
		Username: "test",
	}

	// FlushCache should be a no-op and not return an error
	err := provider.FlushCache(testUser)
	assert.NoError(t, err)
}
