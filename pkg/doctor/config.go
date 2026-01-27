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

package doctor

import (
	"code.vikunja.io/api/pkg/config"

	"github.com/spf13/viper"
)

// CheckConfig returns configuration checks.
func CheckConfig() CheckGroup {
	results := []CheckResult{
		checkConfigFile(),
		checkPublicURL(),
		checkJWTSecret(),
	}

	// Only show CORS details if CORS is enabled
	if config.CorsEnable.GetBool() {
		results = append(results, checkCORS())
	}

	return CheckGroup{
		Name:    "Configuration",
		Results: results,
	}
}

func checkConfigFile() CheckResult {
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		return CheckResult{
			Name:   "Config file",
			Passed: true,
			Value:  "none (using defaults/environment)",
		}
	}

	return CheckResult{
		Name:   "Config file",
		Passed: true,
		Value:  configFile,
	}
}

func checkPublicURL() CheckResult {
	publicURL := config.ServicePublicURL.GetString()
	if publicURL == "" {
		return CheckResult{
			Name:   "Public URL",
			Passed: false,
			Error:  "not configured (required for many features)",
		}
	}

	return CheckResult{
		Name:   "Public URL",
		Passed: true,
		Value:  publicURL,
	}
}

func checkJWTSecret() CheckResult {
	// We can't check the actual value, but we can check if it's the default length
	// which would indicate it was auto-generated
	secret := config.ServiceJWTSecret.GetString()

	// Auto-generated secrets are 64 hex characters (32 bytes)
	if len(secret) == 64 {
		return CheckResult{
			Name:   "JWT secret",
			Passed: true,
			Value:  "configured (auto-generated)",
		}
	}

	return CheckResult{
		Name:   "JWT secret",
		Passed: true,
		Value:  "configured",
	}
}

func checkCORS() CheckResult {
	origins := config.CorsOrigins.GetStringSlice()

	result := CheckResult{
		Name:   "CORS origins",
		Passed: true,
		Value:  "",
	}

	if len(origins) == 0 {
		result.Value = "none configured"
		return result
	}

	// Show first origin in the value, rest as additional lines
	result.Value = origins[0]
	if len(origins) > 1 {
		result.Lines = append(result.Lines, origins[1:]...)
	}

	return result
}
