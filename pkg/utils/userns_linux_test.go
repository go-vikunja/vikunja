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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseUIDMap_Trivial(t *testing.T) {
	input := "         0          0 4294967295\n"
	entries, err := parseUIDMap(input)
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, UIDMapEntry{InsideUID: 0, OutsideUID: 0, Count: 4294967295}, entries[0])
}

func TestParseUIDMap_RootlessDocker(t *testing.T) {
	input := "         0       1001          1\n         1     101001      65536\n"
	entries, err := parseUIDMap(input)
	require.NoError(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, UIDMapEntry{InsideUID: 0, OutsideUID: 1001, Count: 1}, entries[0])
	assert.Equal(t, UIDMapEntry{InsideUID: 1, OutsideUID: 101001, Count: 65536}, entries[1])
}

func TestParseUIDMap_Empty(t *testing.T) {
	entries, err := parseUIDMap("")
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestIsTrivialMapping(t *testing.T) {
	trivial := []UIDMapEntry{{InsideUID: 0, OutsideUID: 0, Count: 4294967295}}
	assert.True(t, isTrivialMapping(trivial))

	rootless := []UIDMapEntry{
		{InsideUID: 0, OutsideUID: 1001, Count: 1},
		{InsideUID: 1, OutsideUID: 101001, Count: 65536},
	}
	assert.False(t, isTrivialMapping(rootless))

	assert.True(t, isTrivialMapping(nil))
}

func TestMapToHostUID(t *testing.T) {
	entries := []UIDMapEntry{
		{InsideUID: 0, OutsideUID: 1001, Count: 1},
		{InsideUID: 1, OutsideUID: 101001, Count: 65536},
	}

	hostUID, ok := mapContainerUID(entries, 0)
	assert.True(t, ok)
	assert.Equal(t, 1001, hostUID)

	hostUID, ok = mapContainerUID(entries, 1)
	assert.True(t, ok)
	assert.Equal(t, 101001, hostUID)

	hostUID, ok = mapContainerUID(entries, 1000)
	assert.True(t, ok)
	assert.Equal(t, 102000, hostUID)

	_, ok = mapContainerUID(entries, 70000)
	assert.False(t, ok)
}

func TestUIDMappingSummaryString(t *testing.T) {
	entries := []UIDMapEntry{
		{InsideUID: 0, OutsideUID: 1001, Count: 1},
		{InsideUID: 1, OutsideUID: 101001, Count: 65536},
	}

	summary := uidMappingSummaryFromEntries(entries)
	assert.Equal(t, "0→1001, 1-65536→101001-166536", summary)
}

func TestUIDMappingSummaryString_Single(t *testing.T) {
	entries := []UIDMapEntry{
		{InsideUID: 0, OutsideUID: 1001, Count: 1},
	}

	summary := uidMappingSummaryFromEntries(entries)
	assert.Equal(t, "0→1001", summary)
}
