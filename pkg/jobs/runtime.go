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

package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"xorm.io/builder"
)

type registration struct {
	queue  string
	handle func(*message.Message) error
}

var handlers = map[string]registration{}

func Register(kind, queue string, handle func(*message.Message) error) {
	handlers[kind] = registration{queue: queue, handle: handle}
}

type Runtime struct {
	queue  *Queue
	router *message.Router
	cancel context.CancelFunc
	done   chan struct{}
}

var runtimeMu sync.Mutex
var current *Runtime

func queueOptions(queue string) Options {
	opts := Options{Concurrency: 2, Lease: time.Minute, Timeout: 5 * time.Minute}
	prefix := "jobs.queues." + queue + "."
	if value := config.Key(prefix + "concurrency").GetInt(); value > 0 {
		opts.Concurrency = value
	}
	if value := config.Key(prefix + "lease").GetDuration(); value > 0 {
		opts.Lease = value
	}
	if value := config.Key(prefix + "timeout").GetDuration(); value > 0 {
		opts.Timeout = value
	}
	return opts
}

func Start(ctx context.Context) error {
	if !config.JobsEnabled.GetBool() {
		return nil
	}
	engine, err := db.CreateDBEngine()
	if err != nil {
		return err
	}
	logger := log.NewWatermillLogger(config.LogEnabled.GetBool(), config.LogEvents.GetString(), config.LogEventsLevel.GetString(), config.LogFormat.GetString())
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return err
	}
	queue := NewQueue(engine, watermill.NewUUID())
	options := make(map[string]Options)
	for _, handler := range handlers {
		if config.Key("jobs.queues." + handler.queue + ".enabled").GetBool() {
			options[handler.queue] = queueOptions(handler.queue)
		}
	}
	subscriber := NewSubscriber(queue, options) //nolint:contextcheck // Heartbeats use independent deadlines.
	router.AddMiddleware(func(next message.HandlerFunc) message.HandlerFunc {
		recovered := middleware.Recoverer(next)
		return func(msg *message.Message) ([]*message.Message, error) {
			messages, err := recovered(msg)
			if err != nil {
				msg.Metadata.Set("job_error", err.Error())
			}
			return messages, err
		}
	})
	for name := range options {
		router.AddConsumerHandler("jobs."+name, name, subscriber, func(msg *message.Message) error {
			job, _ := FromMessage(msg)
			handler, exists := handlers[job.Kind]
			if !exists || job.Version != 1 {
				return fmt.Errorf("unsupported job kind/version: %s/%d", job.Kind, job.Version)
			}
			return handler.handle(msg)
		})
	}
	ctx, cancel := context.WithCancel(ctx)
	runtime := &Runtime{queue: queue, router: router, cancel: cancel, done: make(chan struct{})}
	runtimeMu.Lock()
	current = runtime
	runtimeMu.Unlock()
	go func() {
		defer close(runtime.done)
		if err := router.Run(ctx); err != nil {
			log.Errorf("Durable job router: %v", err)
		}
	}()
	select {
	case <-router.Running():
		return nil
	case <-runtime.done:
		return fmt.Errorf("durable job router stopped at startup")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func Shutdown(ctx context.Context) error {
	runtimeMu.Lock()
	runtime := current
	current = nil
	runtimeMu.Unlock()
	if runtime == nil {
		return nil
	}
	runtime.cancel()
	select {
	case <-runtime.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func Maintain(ctx context.Context) error {
	engine, err := db.CreateDBEngine()
	if err != nil {
		return err
	}
	queue := NewQueue(engine, "maintenance")
	now := time.Now()
	return drainMaintenance(ctx, 5*time.Second, func(batchCtx context.Context) (int64, error) {
		return queue.Rescue(batchCtx, now, 1000)
	})
}

func drainMaintenance(ctx context.Context, budget time.Duration, batch func(context.Context) (int64, error)) error {
	batchCtx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()
	for batchCtx.Err() == nil {
		count, err := batch(batchCtx)
		if err != nil {
			if batchCtx.Err() != nil {
				return ctx.Err()
			}
			return err
		}
		if count == 0 {
			break
		}
	}
	// Unfinished work resumes next tick when only this phase's budget expires.
	return ctx.Err()
}

func Available() bool {
	runtimeMu.Lock()
	runtime := current
	runtimeMu.Unlock()
	if runtime == nil {
		return !config.JobsEnabled.GetBool()
	}
	select {
	case <-runtime.done:
		return false
	default:
		return true
	}
}

func WaitForPendingHandlers() {
	runtimeMu.Lock()
	runtime := current
	runtimeMu.Unlock()
	if runtime == nil {
		return
	}
	for {
		select {
		case <-runtime.done:
			return
		default:
		}
		count, err := runtime.queue.engine.Where(builder.Or(builder.Eq{"state": StateRunning}, builder.And(builder.Eq{"state": StateAvailable}, builder.Lte{"run_at": timestamp(time.Now())}))).Count(new(Job))
		if err == nil && count == 0 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
}
