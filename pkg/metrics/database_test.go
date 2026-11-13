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

package metrics

import (
	"os"
	"os/exec"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"github.com/stretchr/testify/require"
)

func TestMetricDatabaseFailureReturnsError(t *testing.T) {
	if os.Getenv("VIKUNJA_TEST_METRIC_DB_FAILURE") == "1" {
		t.Setenv("VIKUNJA_TESTS_USE_CONFIG", "0")
		engine, err := db.CreateTestEngine()
		require.NoError(t, err)
		require.NoError(t, engine.Close())
		_, err = countFromDatabase(UserCountKey)
		require.ErrorContains(t, err, "database is closed")
		return
	}

	executable, err := os.Executable()
	require.NoError(t, err)
	command := exec.CommandContext(t.Context(), executable, "-test.run=^TestMetricDatabaseFailureReturnsError$")
	command.Env = append(os.Environ(), "VIKUNJA_TEST_METRIC_DB_FAILURE=1")
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))
}
