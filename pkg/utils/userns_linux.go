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

//go:build linux

package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

// UIDMapEntry represents a single line from /proc/self/uid_map.
// Fields use int64 to avoid overflow on 32-bit architectures where the
// trivial mapping count (4294967295) exceeds math.MaxInt32.
type UIDMapEntry struct {
	InsideUID  int64
	OutsideUID int64
	Count      int64
}

var (
	uidMapOnce    sync.Once
	uidMapEntries []UIDMapEntry
	uidMapErr     error
)

func loadUIDMap() {
	data, err := os.ReadFile("/proc/self/uid_map")
	if err != nil {
		uidMapErr = err
		return
	}
	uidMapEntries, uidMapErr = parseUIDMap(string(data))
}

func parseUIDMap(content string) ([]UIDMapEntry, error) {
	var entries []UIDMapEntry
	for _, line := range strings.Split(strings.TrimSpace(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			return nil, fmt.Errorf("unexpected uid_map line: %q", line)
		}
		inside, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing inside uid %q: %w", fields[0], err)
		}
		outside, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing outside uid %q: %w", fields[1], err)
		}
		count, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing count %q: %w", fields[2], err)
		}
		entries = append(entries, UIDMapEntry{
			InsideUID:  inside,
			OutsideUID: outside,
			Count:      count,
		})
	}
	return entries, nil
}

func isTrivialMapping(entries []UIDMapEntry) bool {
	if len(entries) != 1 {
		return len(entries) == 0
	}
	e := entries[0]
	return e.InsideUID == 0 && e.OutsideUID == 0 && e.Count == 4294967295
}

func mapContainerUID(entries []UIDMapEntry, containerUID int64) (hostUID int64, ok bool) {
	for _, e := range entries {
		if containerUID >= e.InsideUID && containerUID < e.InsideUID+e.Count {
			return e.OutsideUID + (containerUID - e.InsideUID), true
		}
	}
	return 0, false
}

func uidMappingSummaryFromEntries(entries []UIDMapEntry) string {
	var parts []string
	for _, e := range entries {
		if e.Count == 1 {
			parts = append(parts, fmt.Sprintf("%d→%d", e.InsideUID, e.OutsideUID))
		} else {
			parts = append(parts, fmt.Sprintf("%d-%d→%d-%d",
				e.InsideUID, e.InsideUID+e.Count-1,
				e.OutsideUID, e.OutsideUID+e.Count-1))
		}
	}
	return strings.Join(parts, ", ")
}

// IsUserNamespaceActive returns true if the process is running inside a
// Linux user namespace with non-trivial UID remapping (e.g., rootless Docker).
func IsUserNamespaceActive() bool {
	uidMapOnce.Do(loadUIDMap)
	if uidMapErr != nil {
		return false
	}
	return !isTrivialMapping(uidMapEntries)
}

// GetUIDMapping returns a copy of the parsed uid_map entries.
// The returned slice is safe to modify without affecting cached state.
func GetUIDMapping() ([]UIDMapEntry, error) {
	uidMapOnce.Do(loadUIDMap)
	if uidMapEntries == nil {
		return nil, uidMapErr
	}
	out := make([]UIDMapEntry, len(uidMapEntries))
	copy(out, uidMapEntries)
	return out, uidMapErr
}

// MapToHostUID maps a container UID to the corresponding host UID.
// Returns mapped=false if no mapping covers that UID.
func MapToHostUID(containerUID int64) (hostUID int64, mapped bool) {
	uidMapOnce.Do(loadUIDMap)
	if uidMapErr != nil {
		return 0, false
	}
	return mapContainerUID(uidMapEntries, containerUID)
}

// UIDMappingSummary returns a human-readable summary of the UID mapping,
// e.g., "0→1001, 1-65536→101001-166536".
func UIDMappingSummary() string {
	uidMapOnce.Do(loadUIDMap)
	if uidMapErr != nil {
		return ""
	}
	return uidMappingSummaryFromEntries(uidMapEntries)
}
