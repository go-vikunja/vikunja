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

func TestAsDataURI(t *testing.T) {
	testUser := &user.User{
		ID:       1,
		Username: "testuser",
		Name:     "Test User",
		Email:    "test@example.com",
	}

	// Table-driven test for SVG providers
	testCases := []struct {
		name     string
		provider Provider
	}{
		{
			name:     "Initials Provider",
			provider: &initials.Provider{},
		},
		{
			name:     "Marble Provider",
			provider: &marble.Provider{},
		},
		{
			name:     "Empty Provider",
			provider: &empty.Provider{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.provider.AsDataURI(testUser, 64)
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if !strings.HasPrefix(result, "data:image/svg+xml;base64,") {
				t.Errorf("Expected data URI with SVG base64, got: %s", result)
			}
			// Basic sanity check for reasonable length
			if len(result) < 50 {
				t.Errorf("Expected longer data URI, got: %s", result)
			}
		})
	}

	t.Run("Gravatar Provider - Base64 Format", func(t *testing.T) {
		// Skip this test as it requires keyvalue store initialization
		// and network access to gravatar service
		t.Skip("Gravatar provider test requires full application setup")
	})
}
