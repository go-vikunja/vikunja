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
	"context"
	"errors"
	"testing"

	"code.vikunja.io/api/pkg/log"

	"github.com/ThreeDotsLabs/watermill/message"
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
	DispatchPending(context.Background(), key)

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

	DispatchPending(context.Background(), key)

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
	DispatchPending(context.Background(), key)

	assert.Equal(t, 0, CountDispatchedEvents("test.event"))
}

func TestDispatchPendingNoEvents(t *testing.T) {
	Fake()

	key := new(int)

	// Should be a no-op
	DispatchPending(context.Background(), key)

	// Verify no events were dispatched
	assert.Equal(t, 0, CountDispatchedEvents("test.event"))
}

func TestOnCommit(t *testing.T) {
	log.InitLogger()

	t.Run("runs on DispatchPending before events", func(t *testing.T) {
		Fake()
		key := new(int)
		var calls []int
		var dispatchedBeforeCallback int

		OnCommit(key, func() error {
			dispatchedBeforeCallback = CountDispatchedEvents("test.event")
			calls = append(calls, 1)
			return errors.New("failing callback must not stop the others")
		})
		DispatchOnCommit(key, &testEvent{})
		OnCommit(key, func() error {
			calls = append(calls, 2)
			return nil
		})
		assert.Empty(t, calls)

		DispatchPending(context.Background(), key)

		assert.Equal(t, []int{1, 2}, calls)
		assert.Zero(t, dispatchedBeforeCallback)
		assert.Equal(t, 1, CountDispatchedEvents("test.event"))

		DispatchPending(context.Background(), key)
		assert.Equal(t, []int{1, 2}, calls, "must run only once")
	})
	t.Run("dropped by CleanupPending", func(t *testing.T) {
		Fake()
		key := new(int)
		called := false

		OnCommit(key, func() error {
			called = true
			return nil
		})
		CleanupPending(key)
		DispatchPending(context.Background(), key)

		assert.False(t, called)
	})
}

func TestShouldReportPoisonedMessage(t *testing.T) {
	t.Run("no metadata", func(t *testing.T) {
		assert.True(t, shouldReportPoisonedMessage(message.Metadata{}))
	})
	t.Run("unrelated metadata", func(t *testing.T) {
		assert.True(t, shouldReportPoisonedMessage(message.Metadata{"reason_poisoned": "boom"}))
	})
	t.Run("flag set", func(t *testing.T) {
		assert.False(t, shouldReportPoisonedMessage(message.Metadata{MetadataSkipErrorReporting: "true"}))
	})
	t.Run("flag set to something else", func(t *testing.T) {
		assert.True(t, shouldReportPoisonedMessage(message.Metadata{MetadataSkipErrorReporting: "false"}))
	})
}
