// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	isUnderTest          bool
	dispatchedTestEvents []Event
)

// Fake sets up the "test mode" of the events package. Typically you'd call this function in the TestMain function
// in the package you're testing. It will prevent any events from being fired, instead they will be recorded and be
// available for assertions.
func Fake() {
	isUnderTest = true
	dispatchedTestEvents = nil
}

// AssertDispatched asserts an event has been dispatched.
func AssertDispatched(t *testing.T, event Event) {
	var found bool
	for _, testEvent := range dispatchedTestEvents {
		if event.Name() == testEvent.Name() {
			found = true
			break
		}
	}

	assert.True(t, found, "Failed to assert "+event.Name()+" has been dispatched.")
}
