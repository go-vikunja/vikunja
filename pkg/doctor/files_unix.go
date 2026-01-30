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
	"os"
	"os/user"
	"strconv"
	"syscall"

	"code.vikunja.io/api/pkg/utils"
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

// isGroupMember reports whether the current process belongs to the given
// group, checking both the primary gid and all supplementary groups.
func isGroupMember(gid int) bool {
	if os.Getgid() == gid {
		return true
	}

	groups, err := os.Getgroups()
	if err != nil {
		return false
	}
	for _, g := range groups {
		if g == gid {
			return true
		}
	}
	return false
}

func checkDirectoryOwnership(info os.FileInfo) []CheckResult {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return []CheckResult{
			{
				Name:   "Directory owner",
				Passed: true,
				Value:  "unable to determine ownership",
			},
		}
	}

	uid := stat.Uid
	gid := stat.Gid

	ownerName := strconv.FormatUint(uint64(uid), 10)
	groupName := strconv.FormatUint(uint64(gid), 10)

	if u, err := user.LookupId(ownerName); err == nil {
		ownerName = u.Username
	}
	if g, err := user.LookupGroupId(groupName); err == nil {
		groupName = g.Name
	}

	currentUID := os.Getuid()

	results := []CheckResult{
		{
			Name:   "Directory owner",
			Passed: true,
			Value:  fmt.Sprintf("%s:%s (uid=%d, gid=%d)", ownerName, groupName, uid, gid),
		},
	}

	nsActive := utils.IsUserNamespaceActive()

	switch {
	case currentUID != 0 && currentUID != int(uid):
		errMsg := fmt.Sprintf(
			"directory owned by uid %d but Vikunja runs as uid %d",
			uid, currentUID,
		)
		if nsActive {
			if hostUID, ok := utils.MapToHostUID(currentUID); ok {
				errMsg += fmt.Sprintf(" (user namespace active, host uid=%d)", hostUID)
			}
		}
		results = append(results, CheckResult{
			Name:   "Ownership match",
			Passed: false,
			Error:  errMsg,
		})
	case currentUID != 0 && !isGroupMember(int(gid)):
		results = append(results, CheckResult{
			Name:   "Ownership match",
			Passed: false,
			Error: fmt.Sprintf(
				"directory owned by gid %d but Vikunja process is not a member of that group",
				gid,
			),
		})
	case currentUID != 0 && nsActive:
		matchResult := CheckResult{
			Name:   "Ownership match",
			Passed: true,
			Value:  fmt.Sprintf("uid %d matches", currentUID),
		}
		if hostUID, ok := utils.MapToHostUID(currentUID); ok {
			matchResult.Lines = []string{
				fmt.Sprintf("WARNING: user namespace active — uid %d maps to host uid %d,", currentUID, hostUID),
				"which may differ from the actual host directory owner.",
				"If writes fail, ensure the host directory is owned by the mapped host uid,",
				"or run the container with --user 0:0.",
			}
		}
		results = append(results, matchResult)
	}

	return results
}
