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
	"fmt"

	"golang.org/x/sys/unix"
)

func checkDiskSpace(path string) CheckResult {
	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return CheckResult{
			Name:   "Disk space",
			Passed: false,
			Error:  err.Error(),
		}
	}

	// Available space in bytes
	availableBytes := stat.Bavail * uint64(stat.Bsize)
	availableGB := float64(availableBytes) / (1024 * 1024 * 1024)

	return CheckResult{
		Name:   "Disk space",
		Passed: true,
		Value:  fmt.Sprintf("%.1f GB available", availableGB),
	}
}
