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
	"fmt"
	"math"
	"time"

	"code.vikunja.io/api/pkg/db"
	"xorm.io/builder"
	"xorm.io/xorm/schemas"
)

type ReplayFilter struct {
	ID           int64
	Queue        string
	Kind         string
	ParkedAfter  time.Time
	ParkedBefore time.Time
	Limit        int
}

type ReplayResult struct {
	Matched  int64
	Replayed int64
}

func (f ReplayFilter) Validate() error {
	selected := f.Queue != "" || f.Kind != "" || !f.ParkedAfter.IsZero() || !f.ParkedBefore.IsZero()
	if f.ID < 0 || (f.ID == 0 && !selected) {
		return errors.New("provide a positive job ID or a queue, kind, or parked-at selector")
	}
	if f.ID != 0 && selected {
		return errors.New("job ID cannot be combined with other selectors")
	}
	if f.Limit < 0 || f.Limit > 1000 {
		return errors.New("replay limit must be between 1 and 1000")
	}
	if !f.ParkedAfter.IsZero() && !f.ParkedBefore.IsZero() && !f.ParkedAfter.Before(f.ParkedBefore) {
		return errors.New("parked-after must precede parked-before")
	}
	return nil
}

func (q *Queue) ReplaySelected(ctx context.Context, filter ReplayFilter, dryRun bool) (result ReplayResult, err error) {
	if err = filter.Validate(); err != nil {
		return result, err
	}
	cond := builder.And(builder.Eq{"state": StateFailed})
	if filter.ID != 0 {
		cond = cond.And(builder.Eq{"id": filter.ID})
	}
	if filter.Queue != "" {
		cond = cond.And(builder.Eq{"queue": filter.Queue})
	}
	if filter.Kind != "" {
		cond = cond.And(builder.Eq{"kind": filter.Kind})
	}
	if !filter.ParkedAfter.IsZero() {
		cond = cond.And(builder.Gte{"finished_at": filter.ParkedAfter.UTC().Format("2006-01-02 15:04:05.999999999")})
	}
	if !filter.ParkedBefore.IsZero() {
		cond = cond.And(builder.Lt{"finished_at": filter.ParkedBefore.UTC().Format("2006-01-02 15:04:05.999999999")}, builder.Gt{"finished_at": timestamp(time.Time{})})
	}
	s := q.engine.NewSession()
	defer s.Close()
	db.SetSessionContext(ctx, s)
	result.Matched, err = s.Where(cond).Count(new(Job))
	if err != nil || dryRun || result.Matched == 0 {
		return result, err
	}
	limit := filter.Limit
	if limit == 0 {
		limit = 100
	}
	var rows []Job
	if err = s.Cols("id", "queue", "attempt").Where(cond).OrderBy("id").Limit(limit).Find(&rows); err != nil {
		return result, err
	}
	for _, job := range rows {
		allowance := retryPolicy(job.Queue).Attempts
		if job.Attempt > math.MaxInt-allowance {
			return result, fmt.Errorf("job %d attempt count cannot be increased", job.ID)
		}
		var changed int64
		if q.engine.Dialect().URI().DBType == schemas.POSTGRES {
			s.SetExpr("run_at", postgresClock)
		}
		changed, err = s.ID(job.ID).Where(builder.Eq{"state": StateFailed, "attempt": job.Attempt}).
			Cols("state", "run_at", "max_attempts", "finished_at").Update(&Job{
			State:       StateAvailable,
			RunAt:       time.Now(),
			MaxAttempts: job.Attempt + allowance,
		})
		if err != nil {
			return result, err
		}
		result.Replayed += changed
		if changed > 0 {
			Wake(job.Queue)
		}
	}
	return result, nil
}
