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
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/errorreport"
	"code.vikunja.io/api/pkg/jobs"
	"code.vikunja.io/api/pkg/log"
	vmetrics "code.vikunja.io/api/pkg/metrics"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/metrics"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"xorm.io/xorm"
)

type handlerTracker struct {
	idle    *sync.Cond
	pending int
}

var activeHandlers = handlerTracker{idle: sync.NewCond(&sync.Mutex{})}

func (h *handlerTracker) start() {
	h.idle.L.Lock()
	h.pending++
	h.idle.L.Unlock()
}

func (h *handlerTracker) finish() {
	h.idle.L.Lock()
	defer h.idle.L.Unlock()
	h.pending--
	if h.pending == 0 {
		h.idle.Broadcast()
	}
}

// WaitForPendingHandlers drains in-flight handlers before test fixtures are reset.
func WaitForPendingHandlers() {
	if r := currentRuntime(); r != nil {
		r.waitForPendingHandlers()
	}
	jobs.WaitForPendingHandlers()
	activeHandlers.idle.L.Lock()
	for activeHandlers.pending > 0 {
		activeHandlers.idle.Wait()
	}
	activeHandlers.idle.L.Unlock()
	if r := currentRuntime(); r != nil {
		r.waitForPendingHandlers()
	}
}

// Event represents the event interface used by all events
type Event interface {
	Name() string
}

// MetadataSkipErrorReporting marks a message whose failure is a user
// configuration problem (an unreachable webhook target, say) rather than a bug,
// so parking it in the poison queue must not page us. The poison middleware
// republishes the same message, so metadata a handler sets survives.
const MetadataSkipErrorReporting = "skip_error_reporting"

type messageHandleFailedError struct {
	Metadata message.Metadata
}

func shouldReportPoisonedMessage(meta message.Metadata) bool {
	return meta.Get(MetadataSkipErrorReporting) != "true"
}

func (m *messageHandleFailedError) Error() string {
	return fmt.Sprintf("Failed to handle message: %v", m.Metadata)
}

// InitEvents runs the event router until it is shut down.
func InitEvents() error {
	if err := Start(context.Background()); err != nil {
		return err
	}
	r := currentRuntime()
	<-r.done
	return r.runErr
}

// Start returns after all listeners have subscribed.
func Start(ctx context.Context) error {
	ready, err := startEvents(ctx, false)
	if err != nil {
		return err
	}
	r := currentRuntime()
	select {
	case <-ready:
		return nil
	case <-r.done:
		return r.runErr
	case <-ctx.Done():
		return ctx.Err()
	}
}

func startEvents(ctx context.Context, testing bool) (<-chan struct{}, error) {
	limit := config.EventsMaxConcurrentHandlers.GetInt()
	if limit < 1 {
		return nil, fmt.Errorf("events.maxconcurrenthandlers must be positive")
	}
	logger := log.NewWatermillLogger(config.LogEnabled.GetBool(), config.LogEvents.GetString(), config.LogEventsLevel.GetString(), config.LogFormat.GetString())
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}
	if !testing {
		metricsBuilder := metrics.NewPrometheusMetricsBuilder(vmetrics.GetRegistry(), "", "")
		metricsBuilder.AddPrometheusRouterMetrics(router)
	}
	channel := gochannel.NewGoChannel(gochannel.Config{OutputChannelBuffer: 1024}, logger)
	ctx, cancel := context.WithCancel(ctx)
	r := newEventRuntime(router, channel, cancel)
	setRuntime(r)

	semaphore := make(chan struct{}, limit)
	router.AddMiddleware(func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			if message.HandlerNameFromCtx(msg.Context()) != "poison.logger" { //nolint:contextcheck // Handler identity belongs to the message context.
				defer r.finish(msg.UUID)
			}
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			activeHandlers.start()
			defer activeHandlers.finish()
			return h(msg)
		}
	})

	if !testing {
		poison, err := middleware.PoisonQueue(channel, "poison")
		if err != nil {
			cancel()
			return nil, err
		}
		router.AddMiddleware(poison)
		router.AddConsumerHandler("poison.logger", "poison", channel, func(msg *message.Message) error {
			log.Errorf("Error while handling message %s, %v", msg.UUID, msg.Metadata)
			if config.SentryEnabled.GetBool() && shouldReportPoisonedMessage(msg.Metadata) {
				failure := &messageHandleFailedError{Metadata: msg.Metadata}
				sentry.WithScope(func(scope *sentry.Scope) {
					errorreport.ApplyFingerprint(scope, failure,
						"message_handle_failed",
						msg.Metadata.Get(middleware.PoisonedHandlerKey),
						errorreport.Normalize(msg.Metadata.Get(middleware.ReasonForPoisonedKey)),
					)
					sentry.CaptureException(failure)
				})
			}
			return nil
		})
	}
	retry := middleware.Retry{
		MaxRetries:          5,
		InitialInterval:     100 * time.Millisecond,
		MaxInterval:         time.Hour,
		Multiplier:          2,
		RandomizationFactor: 1,
		Logger:              logger,
	}
	if testing {
		retry.MaxRetries = 3
		retry.InitialInterval = 50 * time.Millisecond
		retry.MaxInterval = time.Second
		retry.RandomizationFactor = 0
	}
	router.AddMiddleware(
		retry.Middleware,
		middleware.Recoverer,
	)
	for topic, handlers := range listeners {
		for _, handler := range handlers {
			router.AddConsumerHandler(topic+"."+handler.Name(), topic, channel, handler.Handle)
		}
	}

	go func() {
		r.runErr = router.Run(ctx)
		r.mu.Lock()
		r.closed = true
		r.mu.Unlock()
		close(r.done)
	}()
	return router.Running(), nil
}

func InitEventsForTesting(ctx context.Context) (<-chan struct{}, error) {
	return startEvents(ctx, true)
}

// Dispatch dispatches an event
func Dispatch(event Event) error {
	return DispatchWithContext(context.Background(), event)
}

// DispatchWithContext dispatches an event and copies request metadata from the
// context (see WithRequestMeta) onto the message metadata, so listeners can
// attribute the event to the originating HTTP request.
func DispatchWithContext(ctx context.Context, event Event) error {
	if !isUnderTest && len(durableListeners[event.Name()]) > 0 {
		s := db.NewSession()
		defer s.Close()
		db.SetSessionContext(ctx, s)
		if err := enqueueDurable(s, event); err != nil {
			return err
		}
		if err := s.Commit(); err != nil {
			return err
		}
	}
	return dispatchVolatile(ctx, event)
}

func dispatchVolatile(ctx context.Context, event Event) error {
	if isUnderTest {
		dispatchedTestEvents = append(dispatchedTestEvents, event)
		return nil
	}

	r := currentRuntime()
	if r == nil {
		return fmt.Errorf("event system not initialized: call InitEvents or InitEventsForTesting before dispatching")
	}

	content, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), content)
	if meta := RequestMetaFromContext(ctx); meta != nil {
		if meta.IP != "" {
			msg.Metadata.Set(MetadataKeyIP, meta.IP)
		}
		if meta.UserAgent != "" {
			msg.Metadata.Set(MetadataKeyUserAgent, meta.UserAgent)
		}
		if meta.RequestID != "" {
			msg.Metadata.Set(MetadataKeyRequestID, meta.RequestID)
		}
	}
	return r.publish(event.Name(), msg)
}

type pendingEventQueue struct {
	mu        sync.Mutex
	events    []Event
	callbacks []func() error
}

var pendingEvents sync.Map // map[any]*pendingEventQueue

func pendingQueue(key any) *pendingEventQueue {
	val, _ := pendingEvents.LoadOrStore(key, &pendingEventQueue{})
	return val.(*pendingEventQueue)
}

// DispatchOnCommit stores an event to be dispatched later, after a transaction commits.
// The key should be the *xorm.Session pointer associated with the transaction.
// Call DispatchPending(key) after s.Commit() to actually dispatch the events.
// Call CleanupPending(key) on rollback to discard them.
func DispatchOnCommit(key any, event Event) error {
	session, _ := key.(*xorm.Session)
	if err := enqueueDurable(session, event); err != nil {
		return err
	}
	queue := pendingQueue(key)
	queue.mu.Lock()
	queue.events = append(queue.events, event)
	queue.mu.Unlock()
	return nil
}

// OnCommit registers fn to run synchronously in DispatchPending(key), before the pending
// events are published. CleanupPending(key) discards it. Errors are logged.
func OnCommit(key any, fn func() error) {
	queue := pendingQueue(key)
	queue.mu.Lock()
	queue.callbacks = append(queue.callbacks, fn)
	queue.mu.Unlock()
}

// DispatchPending runs the OnCommit callbacks and dispatches all events accumulated for the
// given key, then removes them.
// Call this after s.Commit(). Safe to call even if nothing was registered.
// Request metadata on the context (see WithRequestMeta) is copied onto each message.
// Failures are logged; remaining callbacks and events still run.
func DispatchPending(ctx context.Context, key any) {
	jobs.Wake()
	val, ok := pendingEvents.LoadAndDelete(key)
	if !ok {
		return
	}
	queue := val.(*pendingEventQueue)
	// No need to lock here since we've already removed it from the map
	// and this key won't receive new events
	for _, fn := range queue.callbacks {
		if err := fn(); err != nil {
			log.Errorf("Failed to run on-commit callback: %v", err)
		}
	}
	for _, event := range queue.events {
		if err := dispatchVolatile(ctx, event); err != nil {
			log.Errorf("Failed to dispatch event %s: %v", event.Name(), err)
		}
	}
}

// CleanupPending discards all pending events and callbacks for the given key without
// running them. Call this when a transaction is rolled back.
func CleanupPending(key any) {
	pendingEvents.Delete(key)
}
