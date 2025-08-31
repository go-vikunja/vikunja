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

package openid

import (
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/keyvalue"
)

func TestGetAllProvidersTypeSafety(t *testing.T) {
	// Clean up any existing providers
	defer func() {
		CleanupSavedOpenIDProviders()
	}()

	t.Run("should handle []interface{} without panic", func(t *testing.T) {
		// Setup config with OpenID enabled
		config.AuthOpenIDEnabled.Set(true)

		// Mock the config value to be []interface{} which causes the original panic
		configValue := []interface{}{
			map[string]interface{}{
				"name":         "test-provider",
				"authurl":      "https://example.com/auth",
				"clientid":     "test-client",
				"clientsecret": "test-secret",
			},
		}
		config.AuthOpenIDProviders.Set(configValue)

		// Clear keyvalue cache to force reading from config
		_ = keyvalue.Del("openid_providers")

		// This should not panic, but should handle gracefully and return empty
		providers, err := GetAllProviders()

		// Should return empty providers since the config format is invalid
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if len(providers) != 0 {
			t.Errorf("Expected empty providers list, got: %d", len(providers))
		}
	})

	t.Run("should handle other invalid types without panic", func(t *testing.T) {
		// Setup config with OpenID enabled
		config.AuthOpenIDEnabled.Set(true)

		// Mock the config value to be a string (another invalid type)
		configValue := "invalid-config"
		config.AuthOpenIDProviders.Set(configValue)

		// Clear keyvalue cache to force reading from config
		_ = keyvalue.Del("openid_providers")

		// This should not panic
		providers, err := GetAllProviders()

		// Should return empty providers since the config format is invalid
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if len(providers) != 0 {
			t.Errorf("Expected empty providers list, got: %d", len(providers))
		}
	})
}
