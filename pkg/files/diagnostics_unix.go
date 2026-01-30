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

//go:build !windows

package files

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"code.vikunja.io/api/pkg/utils"
)

// storageDiagnosticInfo gathers process/directory identity and user namespace
// status. It is best-effort: any failure is silently omitted.
func storageDiagnosticInfo(basePath string) string {
	var parts []string

	uid := os.Getuid()
	gid := os.Getgid()
	parts = append(parts, fmt.Sprintf("process uid=%d gid=%d", uid, gid))

	info, err := os.Stat(basePath)
	if err == nil {
		if stat, ok := info.Sys().(*syscall.Stat_t); ok {
			parts = append(parts, fmt.Sprintf("dir owner uid=%d gid=%d", stat.Uid, stat.Gid))
		}
	}

	if utils.IsUserNamespaceActive() {
		summary := utils.UIDMappingSummary()
		parts = append(parts, fmt.Sprintf("user namespace ACTIVE (%s)", summary))

		if hostUID, ok := utils.MapToHostUID(uid); ok {
			parts = append(parts, fmt.Sprintf("process host uid=%d", hostUID))
		}
	}

	result := "[" + strings.Join(parts, ", ") + "]"

	if utils.IsUserNamespaceActive() {
		hostUID, ok := utils.MapToHostUID(uid)
		if ok {
			result += fmt.Sprintf(
				"\n  Hint: A user namespace is active (common in rootless Docker). "+
					"The process appears as uid %d inside the container but maps to uid %d on the host. "+
					"Ensure the host directory is owned by uid %d, or run the container with --user 0:0.",
				uid, hostUID, hostUID)
		}
	}

	return result
}
