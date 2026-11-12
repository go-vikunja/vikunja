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

package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"code.vikunja.io/api/pkg/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/redis/go-redis/v9"
)

type redisEnvelope struct {
	Origin string          `json:"origin"`
	UserID int64           `json:"user_id"`
	Event  string          `json:"event"`
	Data   json.RawMessage `json:"data"`
}

type redisFanout struct {
	client  *redis.Client
	pubsub  *redis.PubSub
	channel string
	origin  string
	cancel  context.CancelFunc
	done    chan struct{}
	once    sync.Once
}

func (h *Hub) StartRedis(ctx context.Context, client *redis.Client, channel string) error {
	if client == nil || channel == "" {
		return errors.New("websocket Redis requires a client and channel")
	}
	h.mu.Lock()
	if h.closed || h.fanout != nil {
		h.mu.Unlock()
		return errors.New("websocket hub already started or closed")
	}
	h.unavailable = true
	h.mu.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	pubsub := client.Subscribe(ctx, channel)
	startupCtx, startupCancel := context.WithTimeout(ctx, 5*time.Second)
	value, err := pubsub.Receive(startupCtx)
	startupCancel()
	if err == nil {
		subscription, ok := value.(*redis.Subscription)
		if !ok || subscription.Kind != "subscribe" || subscription.Channel != channel {
			err = errors.New("redis did not confirm the websocket subscription")
		}
	}
	if err != nil {
		cancel()
		_ = pubsub.Close()
		return fmt.Errorf("subscribe to websocket channel: %w", err)
	}
	fanout := &redisFanout{
		client:  client,
		pubsub:  pubsub,
		channel: channel,
		origin:  watermill.NewUUID(),
		cancel:  cancel,
		done:    make(chan struct{}),
	}
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		cancel()
		_ = pubsub.Close()
		return errors.New("websocket hub closed during startup")
	}
	h.fanout = fanout
	h.unavailable = false
	h.mu.Unlock()

	go fanout.run(ctx, h)
	return nil
}

func (r *redisFanout) publish(userID int64, event string, data any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	envelope, err := json.Marshal(redisEnvelope{
		Origin: r.origin,
		UserID: userID,
		Event:  event,
		Data:   payload,
	})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return r.client.Publish(ctx, r.channel, envelope).Err()
}

func (r *redisFanout) run(ctx context.Context, h *Hub) {
	defer close(r.done)
	defer h.setUnavailable()
	defer r.pubsub.Close()
	pingDone := make(chan struct{})
	go func() {
		defer close(pingDone)
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := r.pubsub.Ping(ctx); err != nil {
					h.setUnavailable()
				}
			}
		}
	}()
	defer func() {
		r.cancel()
		<-pingDone
	}()

	for ctx.Err() == nil {
		value, err := r.pubsub.ReceiveTimeout(ctx, 45*time.Second)
		if err != nil {
			h.setUnavailable()
			if ctx.Err() != nil {
				return
			}
			log.Warningf("WebSocket: Redis subscription interrupted: %v", err)
			timer := time.NewTimer(time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
			continue
		}
		switch value := value.(type) {
		case *redis.Subscription, *redis.Pong:
			h.mu.Lock()
			h.unavailable = false
			h.mu.Unlock()
		case *redis.Message:
			var envelope redisEnvelope
			if json.Unmarshal([]byte(value.Payload), &envelope) != nil ||
				envelope.Origin == r.origin || envelope.UserID <= 0 || !isValidEvent(envelope.Event) {
				continue
			}
			h.publishLocal(envelope.UserID, envelope.Event, envelope.Data)
		}
	}
}

func (r *redisFanout) stop() {
	r.once.Do(func() {
		r.cancel()
		_ = r.pubsub.Close()
	})
	<-r.done
}
