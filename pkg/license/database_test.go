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

package license

import (
	"os"
	"os/exec"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestLicenseDatabaseFailuresReturnErrors(t *testing.T) {
	if operation := os.Getenv("VIKUNJA_TEST_LICENSE_DB_FAILURE"); operation != "" {
		config.InitDefaultConfig()
		log.InitLogger()
		t.Setenv("VIKUNJA_TESTS_USE_CONFIG", "0")
		engine, err := db.CreateTestEngine()
		require.NoError(t, err)
		require.NoError(t, engine.Close())

		switch operation {
		case "user counts":
			_, err = getUserCounts()
		case "cached status":
			_, err = loadCachedStatus()
		case "cache response":
			err = cacheResponse(&Response{Valid: true})
		case "clear cache":
			err = clearCache()
		default:
			t.Fatalf("unknown operation %q", operation)
		}
		require.ErrorContains(t, err, "database is closed")
		return
	}

	executable, err := os.Executable()
	require.NoError(t, err)
	for _, operation := range []string{"user counts", "cached status", "cache response", "clear cache"} {
		t.Run(operation, func(t *testing.T) {
			command := exec.CommandContext(t.Context(), executable, "-test.run=^TestLicenseDatabaseFailuresReturnErrors$")
			command.Env = append(os.Environ(), "VIKUNJA_TEST_LICENSE_DB_FAILURE="+operation)
			output, err := command.CombinedOutput()
			require.NoError(t, err, string(output))
		})
	}
}
