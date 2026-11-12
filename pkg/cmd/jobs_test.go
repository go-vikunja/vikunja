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

package cmd

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJobReplayRejectsInvalidSelectorsBeforeInitialization(t *testing.T) {
	for _, test := range []struct {
		args  []string
		error string
	}{
		{[]string{}, "provide a positive job ID"},
		{[]string{"abc"}, "positive integer"},
		{[]string{"0"}, "positive integer"},
		{[]string{"1", "--queue", "mail"}, "cannot be combined"},
		{[]string{"--queue", "mail", "--limit", "0"}, "between 1 and 1000"},
		{[]string{"--queue", "mail", "--limit", "1001"}, "between 1 and 1000"},
		{[]string{"--parked-after", "yesterday"}, "invalid parked-after"},
		{[]string{"--parked-before", "tomorrow"}, "invalid parked-before"},
		{[]string{"--parked-after", "2026-11-13T00:00:00Z", "--parked-before", "2026-11-12T00:00:00Z"}, "must precede"},
	} {
		command := newJobReplayCommand()
		command.SetOut(io.Discard)
		command.SetErr(io.Discard)
		command.SetArgs(test.args)
		assert.ErrorContains(t, command.Execute(), test.error)
	}
}
