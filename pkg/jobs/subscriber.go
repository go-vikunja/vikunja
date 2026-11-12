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
	"errors"
	"strconv"
	"sync"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"github.com/ThreeDotsLabs/watermill/message"
	"xorm.io/xorm"
)

type Options struct {
	Concurrency int
	Lease       time.Duration
	Timeout     time.Duration
}

type attemptKey struct{}
type activeClaim struct {
	job   Job
	ctx   context.Context
	lease time.Duration
}

type Subscriber struct {
	queue   *Queue
	options map[string]Options
	closing chan struct{}
	once    sync.Once
	wg      sync.WaitGroup
	mu      sync.Mutex
	active  map[int64]activeClaim
}

func NewSubscriber(queue *Queue, options map[string]Options) *Subscriber {
	s := &Subscriber{queue: queue, options: options, closing: make(chan struct{}), active: make(map[int64]activeClaim)}
	s.wg.Add(1)
	go s.heartbeat()
	return s
}

func FromMessage(msg *message.Message) (Job, bool) {
	job, ok := msg.Context().Value(attemptKey{}).(Job)
	return job, ok
}

func (s *Subscriber) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	opts, ok := s.options[topic]
	if !ok || opts.Concurrency < 1 || opts.Lease <= 0 || opts.Timeout <= 0 {
		return nil, errors.New("invalid job queue options")
	}
	out := make(chan *message.Message)
	slots := make(chan struct{}, opts.Concurrency)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer close(out)
		delay := 50 * time.Millisecond
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.closing:
				return
			default:
			}
			batch, err := s.queue.Claim(ctx, topic, cap(slots)-len(slots), time.Now(), opts.Lease)
			if err != nil {
				log.Errorf("Claim jobs from %s: %v", topic, err)
			}
			if len(batch) == 0 {
				timer := time.NewTimer(delay)
				select {
				case <-timer.C:
				case <-wake:
				case <-ctx.Done():
					timer.Stop()
					return
				case <-s.closing:
					timer.Stop()
					return
				}
				timer.Stop()
				delay = min(delay*2, 5*time.Second)
				continue
			}
			delay = 50 * time.Millisecond
			for _, job := range batch {
				slots <- struct{}{}
				attemptCtx, cancel := context.WithTimeout(context.WithValue(ctx, attemptKey{}, job), opts.Timeout)
				msg := message.NewMessage(strconv.FormatInt(job.ID, 10), []byte(job.Payload))
				msg.SetContext(attemptCtx)
				s.mu.Lock()
				s.active[job.ID] = activeClaim{job: job, ctx: attemptCtx, lease: opts.Lease}
				s.mu.Unlock()
				s.wg.Add(1)
				go s.await(ctx, msg, job, cancel, slots)
				select {
				case out <- msg:
				case <-s.closing:
					return
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return out, nil
}

func (s *Subscriber) await(ctx context.Context, msg *message.Message, job Job, cancel context.CancelFunc, slots chan struct{}) {
	defer s.wg.Done()
	defer cancel()
	defer func() { <-slots; s.mu.Lock(); delete(s.active, job.ID); s.mu.Unlock(); Wake() }()
	select {
	case <-msg.Acked():
		session := s.queue.engine.NewSession()
		defer session.Close()
		// Transactional handlers already completed their attempt with the effect.
		if err := Complete(session, job, time.Now()); err != nil && !errors.Is(err, ErrStaleClaim) {
			log.Errorf("Complete job %d: %v", job.ID, err)
		}
	case <-msg.Nacked():
		cause := errors.New(msg.Metadata.Get("job_error"))
		if err := s.queue.Fail(context.WithoutCancel(ctx), job, time.Now(), cause, job.MaxAttempts); err != nil {
			log.Errorf("Record job %d failure: %v", job.ID, err)
		}
	case <-msg.Context().Done():
		// Keep the slot until the cancelled handler returns, bounding stuck work.
		select {
		case <-msg.Acked():
		case <-msg.Nacked():
		case <-s.closing:
		}
	case <-s.closing:
	}
}

func (s *Subscriber) heartbeat() {
	defer s.wg.Done()
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			batches := make(map[time.Duration][]Job)
			s.mu.Lock()
			for _, claim := range s.active {
				if claim.ctx.Err() == nil {
					batches[claim.lease] = append(batches[claim.lease], claim.job)
				}
			}
			s.mu.Unlock()
			for lease, batch := range batches {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				if _, err := s.queue.Renew(ctx, batch, time.Now(), lease); err != nil {
					log.Errorf("Renew job claims: %v", err)
				}
				cancel()
			}
		case <-s.closing:
			return
		}
	}
}

func (s *Subscriber) Close() error {
	s.once.Do(func() { close(s.closing) })
	s.wg.Wait()
	return nil
}

// Commit fences database effects against a rescued or timed-out attempt.
func Commit(session *xorm.Session, msg *message.Message) error {
	if job, durable := FromMessage(msg); durable {
		if err := msg.Context().Err(); err != nil {
			return err
		}
		if err := Complete(session, job, time.Now()); err != nil {
			return err
		}
	}
	return session.Commit()
}

func NewSession(msg *message.Message) *xorm.Session {
	s := db.NewSession()
	db.SetSessionContext(msg.Context(), s)
	return s
}
