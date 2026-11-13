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
	"code.vikunja.io/api/pkg/log"
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
	postgres := s.Engine().Dialect().URI().DBType == schemas.POSTGRES
	if request.RunAt.IsZero() && !postgres {
		request.RunAt = time.Now()
	}
	job := &Job{
		Queue:       request.Queue,
		Kind:        request.Kind,
		Version:     1,
		Payload:     string(request.Payload),
		State:       StateAvailable,
		MaxAttempts: retryPolicy(request.Queue).Attempts,
		RunAt:       request.RunAt,
	}
	if request.UniqueKey == "" {
		if postgres && request.RunAt.IsZero() {
			s.SetExpr("run_at", postgresClock)
		}
		tx := s.Tx()
		_, err := s.After(func(any) {
			if s.Tx() == tx {
				jobCounters["enqueued"].WithLabelValues(job.Queue, job.Kind).Inc()
			}
		}).Insert(job)
		return err
	}
	hash := sha256.Sum256([]byte(request.Queue + "\x00" + request.Kind + "\x00" + request.UniqueKey))
	key := hex.EncodeToString(hash[:])
	runAtSQL := "?"
	args := []any{job.Queue, job.Kind, job.Payload, job.State, job.MaxAttempts}
	if postgres && request.RunAt.IsZero() {
		runAtSQL = postgresClock
	} else {
		args = append(args, timestamp(job.RunAt))
	}
	args = append(args, key)
	query := "INSERT INTO jobs (queue, kind, version, payload, state, attempt, max_attempts, run_at, unique_key) VALUES (?, ?, 1, ?, ?, 0, ?, " + runAtSQL + ", ?)"
	if s.Engine().Dialect().URI().DBType == schemas.MYSQL {
		query += " ON DUPLICATE KEY UPDATE unique_key = unique_key"
	} else {
		query += " ON CONFLICT (unique_key) DO NOTHING"
	}
	result, err := s.Exec(append([]any{query}, args...)...)
	if err == nil {
		count, countErr := result.RowsAffected()
		if countErr != nil {
			return countErr
		}
		countAfterCommit(s, "enqueued", job.Queue, job.Kind, count)
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
		// A locking CTE prevents rescans from claiming beyond the limit.
		err := s.SQL(postgresClaimSQL, queue, limit, q.owner, lease.Seconds()).Find(&jobs)
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
	expiry := builder.Cond(builder.Gt{"expires_at": timestamp(now)})
	if s.Engine().Dialect().URI().DBType == schemas.POSTGRES {
		expiry = builder.Expr("expires_at > (SELECT " + postgresClock + ")")
	}
	affected, err := s.Where(ownership(job)).And(expiry).Delete(new(Job))
	if err != nil {
		return err
	}
	if affected != 1 {
		return ErrStaleClaim
	}
	countAfterCommit(s, "completed", job.Queue, job.Kind, 1)
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
	if q.engine.Dialect().URI().DBType == schemas.POSTGRES {
		condition, args, err := builder.ToSQL(builder.Or(conditions...))
		if err != nil {
			return 0, err
		}
		return execRows(s, "WITH clock AS (SELECT "+postgresClock+" AS now) UPDATE jobs SET expires_at = clock.now + ? * INTERVAL '1 second' FROM clock WHERE ("+condition+") AND expires_at > clock.now", append([]any{lease.Seconds()}, args...)...)
	}
	return s.Where(builder.Or(conditions...)).And(builder.Gt{"expires_at": timestamp(now)}).Cols("expires_at").Update(&Job{ExpiresAt: now.Add(lease)})
}

func (q *Queue) Fail(ctx context.Context, job Job, now time.Time, cause error, maxAttempts int) error {
	state := StateAvailable
	if job.Attempt >= maxAttempts {
		state = StateFailed
	}
	delay := retryPolicy(job.Queue).Delay(job.Attempt)
	s := q.engine.NewSession()
	defer s.Close()
	db.SetSessionContext(ctx, s)
	count, err := updateFailedClaim(s, job, now, state, delay, cause.Error(), false)
	if err == nil && count == 1 {
		countAfterCommit(s, "failed_attempts", job.Queue, job.Kind, count)
		if state == StateFailed {
			countAfterCommit(s, "parked", job.Queue, job.Kind, count)
			log.Warningf("[Jobs] parked job id=%d queue=%s kind=%s attempt=%d last_error=%q", job.ID, job.Queue, job.Kind, job.Attempt, cause.Error())
		}
	}
	return err
}

func (q *Queue) Rescue(ctx context.Context, now time.Time, limit int) (int64, error) {
	s := q.engine.NewSession()
	defer s.Close()
	db.SetSessionContext(ctx, s)
	var expired []Job
	expiry := builder.Cond(builder.Lte{"expires_at": timestamp(now)})
	if q.engine.Dialect().URI().DBType == schemas.POSTGRES {
		expiry = builder.Expr("expires_at <= (SELECT " + postgresClock + ")")
	}
	if err := s.Where(builder.Eq{"state": StateRunning}).And(expiry).Limit(limit).Find(&expired); err != nil {
		return 0, err
	}
	var total int64
	for _, job := range expired {
		state := StateAvailable
		if job.Attempt >= job.MaxAttempts {
			state = StateFailed
		}
		count, err := updateFailedClaim(s, job, now, state, 0, "worker claim expired", true)
		if err != nil {
			return total, err
		}
		if count == 1 {
			countAfterCommit(s, "rescued", job.Queue, job.Kind, count)
			log.Infof("[Jobs] rescued job id=%d queue=%s kind=%s attempt=%d last_error=%q", job.ID, job.Queue, job.Kind, job.Attempt, "worker claim expired")
			if state == StateFailed {
				countAfterCommit(s, "parked", job.Queue, job.Kind, count)
				log.Warningf("[Jobs] parked job id=%d queue=%s kind=%s attempt=%d last_error=%q", job.ID, job.Queue, job.Kind, job.Attempt, "worker claim expired")
			}
		}
		total += count
	}
	return total, nil
}

func (q *Queue) Replay(ctx context.Context, id int64) error {
	result, err := q.ReplaySelected(ctx, ReplayFilter{ID: id}, false)
	if err != nil {
		return err
	}
	if result.Replayed == 0 {
		return errors.New("parked job not found")
	}
	return nil
}
