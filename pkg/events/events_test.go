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

package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testEvent struct{}

func (t *testEvent) Name() string {
	return "test.event"
}

func TestDispatchOnCommit(t *testing.T) {
	Fake()

	// Use a simple key (in real use this is a *xorm.Session)
	key := new(int)

	DispatchOnCommit(key, &testEvent{})

	// Event should NOT be dispatched yet
	assert.Equal(t, 0, CountDispatchedEvents("test.event"))

	// Simulate post-commit dispatch
	DispatchPending(key)

	// Now it should be dispatched
	assert.Equal(t, 1, CountDispatchedEvents("test.event"))
}

func TestDispatchOnCommitMultipleEvents(t *testing.T) {
	Fake()

	key := new(int)

	DispatchOnCommit(key, &testEvent{})
	DispatchOnCommit(key, &testEvent{})
	DispatchOnCommit(key, &testEvent{})

	assert.Equal(t, 0, CountDispatchedEvents("test.event"))

	DispatchPending(key)

	assert.Equal(t, 3, CountDispatchedEvents("test.event"))
}

func TestCleanupPending(t *testing.T) {
	Fake()

	key := new(int)

	DispatchOnCommit(key, &testEvent{})
	DispatchOnCommit(key, &testEvent{})

	// Simulate rollback — discard events
	CleanupPending(key)

	// Dispatching after cleanup should be a no-op
	DispatchPending(key)

	assert.Equal(t, 0, CountDispatchedEvents("test.event"))
}

func TestDispatchPendingNoEvents(t *testing.T) {
	Fake()

	key := new(int)

	// Should be a no-op
	DispatchPending(key)

	// Verify no events were dispatched
	assert.Equal(t, 0, CountDispatchedEvents("test.event"))
}
