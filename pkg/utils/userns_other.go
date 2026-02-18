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

//go:build !linux

package utils

// UIDMapEntry represents a single line from /proc/self/uid_map.
// On non-Linux platforms this type exists for API compatibility but is never populated.
// Fields use int64 to avoid overflow on 32-bit architectures.
type UIDMapEntry struct {
	InsideUID  int64
	OutsideUID int64
	Count      int64
}

// IsUserNamespaceActive always returns false on non-Linux platforms.
func IsUserNamespaceActive() bool { return false }

// GetUIDMapping returns nil on non-Linux platforms.
func GetUIDMapping() ([]UIDMapEntry, error) { return nil, nil }

// MapToHostUID always returns mapped=false on non-Linux platforms.
func MapToHostUID(_ int64) (int64, bool) { return 0, false }

// UIDMappingSummary returns empty string on non-Linux platforms.
func UIDMappingSummary() string { return "" }
