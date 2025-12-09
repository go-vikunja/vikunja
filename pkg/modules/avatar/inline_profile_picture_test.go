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

package avatar

import (
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/modules/avatar/empty"
	"code.vikunja.io/api/pkg/modules/avatar/initials"
	"code.vikunja.io/api/pkg/modules/avatar/marble"
	"code.vikunja.io/api/pkg/user"
)

func TestInlineProfilePicture(t *testing.T) {
	testUser := &user.User{
		ID:       1,
		Username: "testuser",
		Name:     "Test User",
		Email:    "test@example.com",
	}

	t.Run("Initials Provider", func(t *testing.T) {
		provider := &initials.Provider{}
		result, err := provider.InlineProfilePicture(testUser, 64)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if !strings.Contains(result, "<svg") {
			t.Errorf("Expected SVG content, got: %s", result)
		}
		if !strings.Contains(result, "T") { // Should contain the initial "T"
			t.Errorf("Expected initial 'T' in SVG, got: %s", result)
		}
	})

	t.Run("Marble Provider", func(t *testing.T) {
		provider := &marble.Provider{}
		result, err := provider.InlineProfilePicture(testUser, 64)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if !strings.Contains(result, "<svg") {
			t.Errorf("Expected SVG content, got: %s", result)
		}
		if !strings.Contains(result, "mask__marble") {
			t.Errorf("Expected marble-specific SVG content, got: %s", result)
		}
	})

	t.Run("Empty Provider", func(t *testing.T) {
		provider := &empty.Provider{}
		result, err := provider.InlineProfilePicture(testUser, 64)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if !strings.Contains(result, "<svg") {
			t.Errorf("Expected SVG content, got: %s", result)
		}
		if !strings.Contains(result, "<?xml") {
			t.Errorf("Expected XML declaration in SVG, got: %s", result)
		}
	})

	t.Run("Gravatar Provider - Base64 Format", func(t *testing.T) {
		// Skip this test as it requires keyvalue store initialization
		// and network access to gravatar service
		t.Skip("Gravatar provider test requires full application setup")
	})
}
