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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"code.vikunja.io/api/pkg/db"
	"xorm.io/builder"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

const (
	StateAvailable  = "available"
	StateRunning    = "running"
	StateFailed     = "failed"
	DefaultAttempts = 5
)

var ErrStaleClaim = errors.New("job claim is no longer owned by this attempt")

type Job struct {
	ID          int64     `xorm:"bigint autoincr pk"`
	Queue       string    `xorm:"varchar(100) not null index(claim)"`
	Kind        string    `xorm:"varchar(200) not null"`
	Version     int       `xorm:"not null default 1"`
	Payload     string    `xorm:"longtext not null"`
	State       string    `xorm:"varchar(20) not null index(claim)"`
	Attempt     int       `xorm:"not null default 0"`
	MaxAttempts int       `xorm:"not null default 5"`
	RunAt       time.Time `xorm:"datetime not null index(claim)"`
	ClaimedAt   time.Time `xorm:"datetime null"`
	ClaimOwner  string    `xorm:"varchar(100) null"`
	ExpiresAt   time.Time `xorm:"datetime null index"`
	UniqueKey   *string   `xorm:"varchar(64) null unique"`
	LastError   string    `xorm:"text null"`
	FinishedAt  time.Time `xorm:"datetime null index"`
}

func (*Job) TableName() string { return "jobs" }
func (j *Job) AfterInsert()    { Wake(j.Queue) }
func init()                    { db.RegisterTables([]interface{}{new(Job)}) }

type Request struct {
	Queue     string
	Kind      string
	Payload   json.RawMessage
	UniqueKey string
	RunAt     time.Time
}

func Enqueue(s *xorm.Session, request Request) error {
	if request.Queue == "" || request.Kind == "" || !json.Valid(request.Payload) {
		return errors.New("job requires a queue, kind and JSON payload")
	}
	if request.RunAt.IsZero() {
		request.RunAt = time.Now()
	}
	job := &Job{
		Queue:       request.Queue,
		Kind:        request.Kind,
		Version:     1,
		Payload:     string(request.Payload),
		State:       StateAvailable,
		MaxAttempts: DefaultAttempts,
		RunAt:       request.RunAt,
	}
	if request.UniqueKey == "" {
		_, err := s.Insert(job)
		return err
	}
	hash := sha256.Sum256([]byte(request.Queue + "\x00" + request.Kind + "\x00" + request.UniqueKey))
	key := hex.EncodeToString(hash[:])
	query := "INSERT INTO jobs (queue, kind, version, payload, state, attempt, max_attempts, run_at, unique_key) VALUES (?, ?, 1, ?, ?, 0, ?, ?, ?) ON CONFLICT (unique_key) DO NOTHING"
	if s.Engine().Dialect().URI().DBType == schemas.MYSQL {
		query = "INSERT INTO jobs (queue, kind, version, payload, state, attempt, max_attempts, run_at, unique_key) VALUES (?, ?, 1, ?, ?, 0, ?, ?, ?) ON DUPLICATE KEY UPDATE unique_key = unique_key"
	}
	_, err := s.Exec(query, job.Queue, job.Kind, job.Payload, job.State, job.MaxAttempts, timestamp(job.RunAt), key)
	if err == nil {
		wakeAfterCommit(s, job.Queue)
	}
	return err
}

func timestamp(t time.Time) string { return t.UTC().Format("2006-01-02 15:04:05") }

type Queue struct {
	engine *xorm.Engine
	owner  string
}

var claimMu sync.Mutex

func NewQueue(engine *xorm.Engine, owner string) *Queue { return &Queue{engine: engine, owner: owner} }

func (q *Queue) Claim(ctx context.Context, queue string, limit int, now time.Time, lease time.Duration) ([]Job, error) {
	if limit < 1 {
		return nil, nil
	}
	s := q.engine.NewSession()
	defer s.Close()
	db.SetSessionContext(ctx, s)
	var jobs []Job
	if q.engine.Dialect().URI().DBType == schemas.POSTGRES {
		err := s.SQL(`UPDATE jobs SET state = 'running', attempt = attempt + 1, claimed_at = ?, claim_owner = ?, expires_at = ?
   WHERE id IN (SELECT id FROM jobs WHERE queue = ? AND state = 'available' AND run_at <= ? ORDER BY run_at, id LIMIT ? FOR UPDATE SKIP LOCKED)
   RETURNING *`, timestamp(now), q.owner, timestamp(now.Add(lease)), queue, timestamp(now), limit).Find(&jobs)
		if err != nil {
			return nil, err
		}
		return jobs, nil
	}
	claimMu.Lock()
	defer claimMu.Unlock()
	if err := s.Begin(); err != nil {
		return nil, err
	}
	if err := s.Where(builder.Eq{"queue": queue, "state": StateAvailable}).And(builder.Lte{"run_at": timestamp(now)}).OrderBy("run_at, id").Limit(limit).Find(&jobs); err != nil {
		return nil, err
	}
	for i := range jobs {
		job := &jobs[i]
		job.State = StateRunning
		job.Attempt++
		job.ClaimedAt = now
		job.ClaimOwner = q.owner
		job.ExpiresAt = now.Add(lease)
		if _, err := s.ID(job.ID).Cols("state", "attempt", "claimed_at", "claim_owner", "expires_at").Update(job); err != nil {
			return nil, err
		}
	}
	if err := s.Commit(); err != nil {
		return nil, err
	}
	return jobs, nil
}

func ownership(job Job) builder.Cond {
	return builder.Eq{"id": job.ID, "state": StateRunning, "claim_owner": job.ClaimOwner, "attempt": job.Attempt}
}

func Complete(s *xorm.Session, job Job, now time.Time) error {
	affected, err := s.Where(ownership(job)).And(builder.Gt{"expires_at": timestamp(now)}).Delete(new(Job))
	if err != nil {
		return err
	}
	if affected != 1 {
		return ErrStaleClaim
	}
	return nil
}

func (q *Queue) Renew(ctx context.Context, jobs []Job, now time.Time, lease time.Duration) (int64, error) {
	if len(jobs) == 0 {
		return 0, nil
	}
	conditions := make([]builder.Cond, 0, len(jobs))
	for _, job := range jobs {
		conditions = append(conditions, ownership(job))
	}
	s := q.engine.NewSession()
	defer s.Close()
	db.SetSessionContext(ctx, s)
	return s.Where(builder.Or(conditions...)).And(builder.Gt{"expires_at": timestamp(now)}).Cols("expires_at").Update(&Job{ExpiresAt: now.Add(lease)})
}

func (q *Queue) Fail(ctx context.Context, job Job, now time.Time, cause error, maxAttempts int) error {
	state := StateAvailable
	if job.Attempt >= maxAttempts {
		state = StateFailed
	}
	delay := min(time.Second*time.Duration(1<<min(job.Attempt, 12)), time.Hour)
	s := q.engine.NewSession()
	defer s.Close()
	db.SetSessionContext(ctx, s)
	_, err := s.Where(ownership(job)).Cols("state", "run_at", "last_error").Update(&Job{State: state, RunAt: now.Add(delay), LastError: cause.Error()})
	return err
}

func (q *Queue) Rescue(ctx context.Context, now time.Time, limit int) (int64, error) {
	s := q.engine.NewSession()
	defer s.Close()
	db.SetSessionContext(ctx, s)
	var expired []Job
	if err := s.Where(builder.Eq{"state": StateRunning}).And(builder.Lte{"expires_at": timestamp(now)}).Limit(limit).Find(&expired); err != nil {
		return 0, err
	}
	var total int64
	for _, job := range expired {
		state := StateAvailable
		if job.Attempt >= job.MaxAttempts {
			state = StateFailed
		}
		count, err := s.Where(ownership(job)).And(builder.Lte{"expires_at": timestamp(now)}).Cols("state", "run_at", "last_error").Update(&Job{State: state, RunAt: now, LastError: "worker claim expired"})
		if err != nil {
			return total, err
		}
		total += count
	}
	return total, nil
}

func (q *Queue) Replay(ctx context.Context, id int64) error {
	s := q.engine.NewSession()
	defer s.Close()
	db.SetSessionContext(ctx, s)
	var job Job
	found, err := s.ID(id).Where(builder.Eq{"state": StateFailed}).Get(&job)
	if err != nil {
		return err
	}
	if !found {
		return errors.New("parked job not found")
	}
	_, err = s.ID(id).Where(builder.Eq{"state": StateFailed}).Cols("state", "run_at", "max_attempts").Update(&Job{State: StateAvailable, RunAt: time.Now(), MaxAttempts: job.Attempt + DefaultAttempts})
	return err
}
