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
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	vmetrics "code.vikunja.io/api/pkg/metrics"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/metrics"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

var pubsub *gochannel.GoChannel

// activeHandlers tracks in-flight event handler goroutines so the test
// endpoint can wait for them to finish before truncating tables.
var activeHandlers sync.WaitGroup

// WaitForPendingHandlers blocks until all currently in-flight event handler
// goroutines have completed (including retries). This is intended for the
// testing endpoint to avoid connection starvation: async handlers from the
// previous test can hold SQLite connections, starving the next test's seed
// request.
func WaitForPendingHandlers() {
	activeHandlers.Wait()
}

// Event represents the event interface used by all events
type Event interface {
	Name() string
}

type messageHandleFailedError struct {
	Metadata message.Metadata
}

func (m *messageHandleFailedError) Error() string {
	return fmt.Sprintf("Failed to handle message: %v", m.Metadata)
}

// InitEvents sets up everything needed to work with events
func InitEvents() (err error) {
	logger := log.NewWatermillLogger(config.LogEnabled.GetBool(), config.LogEvents.GetString(), config.LogEventsLevel.GetString(), config.LogFormat.GetString())

	router, err := message.NewRouter(
		message.RouterConfig{},
		logger,
	)
	if err != nil {
		return err
	}

	metricsBuilder := metrics.NewPrometheusMetricsBuilder(vmetrics.GetRegistry(), "", "")
	metricsBuilder.AddPrometheusRouterMetrics(router)

	pubsub = gochannel.NewGoChannel(
		gochannel.Config{
			OutputChannelBuffer: 1024,
		},
		logger,
	)

	poison, err := middleware.PoisonQueue(pubsub, "poison")
	if err != nil {
		return err
	}
	router.AddConsumerHandler("poison.logger", "poison", pubsub, func(msg *message.Message) error {
		meta := ""
		for s, m := range msg.Metadata {
			meta += s + "=" + m + ", "
		}
		log.Errorf("Error while handling message %s, %s payload=%s", msg.UUID, meta, string(msg.Payload))

		if config.SentryEnabled.GetBool() {
			sentry.CaptureException(&messageHandleFailedError{
				Metadata: msg.Metadata,
			})
		}
		return nil
	})

	// handlerTracker is a middleware that tracks in-flight handlers via the
	// activeHandlers WaitGroup. It wraps the entire processing chain
	// (including retries) so WaitForPendingHandlers() can drain all work.
	handlerTracker := func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			activeHandlers.Add(1)
			defer activeHandlers.Done()
			return h(msg)
		}
	}

	router.AddMiddleware(
		handlerTracker,
		poison,
		middleware.Retry{
			MaxRetries:          5,
			InitialInterval:     time.Millisecond * 100,
			MaxInterval:         time.Hour,
			Multiplier:          2,
			MaxElapsedTime:      0,
			RandomizationFactor: 1,
			Logger:              logger,
		}.Middleware,
		middleware.Recoverer,
	)

	for topic, funcs := range listeners {
		for _, handler := range funcs {
			router.AddConsumerHandler(topic+"."+handler.Name(), topic, pubsub, handler.Handle)
		}
	}

	return router.Run(context.Background())
}

// Dispatch dispatches an event
func Dispatch(event Event) error {
	if isUnderTest {
		dispatchedTestEvents = append(dispatchedTestEvents, event)
		return nil
	}

	content, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), content)
	return pubsub.Publish(event.Name(), msg)
}

// pendingEventQueue holds the pending events and a mutex for thread-safe access
type pendingEventQueue struct {
	mu     sync.Mutex
	events []Event
}

var pendingEvents sync.Map // map[any]*pendingEventQueue

// DispatchOnCommit stores an event to be dispatched later, after a transaction commits.
// The key should be the *xorm.Session pointer associated with the transaction.
// Call DispatchPending(key) after s.Commit() to actually dispatch the events.
// Call CleanupPending(key) on rollback to discard them.
func DispatchOnCommit(key any, event Event) {
	val, _ := pendingEvents.LoadOrStore(key, &pendingEventQueue{})
	queue := val.(*pendingEventQueue)
	queue.mu.Lock()
	queue.events = append(queue.events, event)
	queue.mu.Unlock()
}

// DispatchPending dispatches all events accumulated for the given key and removes them.
// Call this after s.Commit(). Safe to call even if no events were registered.
// If any event fails to dispatch, the error is logged but remaining events are still dispatched.
func DispatchPending(key any) {
	val, ok := pendingEvents.LoadAndDelete(key)
	if !ok {
		return
	}
	queue := val.(*pendingEventQueue)
	// No need to lock here since we've already removed it from the map
	// and this key won't receive new events
	for _, event := range queue.events {
		if err := Dispatch(event); err != nil {
			log.Errorf("Failed to dispatch event %s: %v", event.Name(), err)
		}
	}
}

// CleanupPending discards all pending events for the given key without dispatching.
// Call this when a transaction is rolled back.
func CleanupPending(key any) {
	pendingEvents.Delete(key)
}
