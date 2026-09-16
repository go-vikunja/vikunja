//go:build !windows

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
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckDirectoryOwnership_MatchingUIDIgnoresGroup(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root does not get an ownership match result")
	}

	info, err := os.Stat(t.TempDir())
	require.NoError(t, err)
	stat, ok := info.Sys().(*syscall.Stat_t)
	require.True(t, ok)
	stat.Gid = ^uint32(0)

	results := checkDirectoryOwnership(info)

	for _, result := range results {
		if result.Name == "Ownership match" {
			assert.True(t, result.Passed, result.Error)
		}
	}
}
