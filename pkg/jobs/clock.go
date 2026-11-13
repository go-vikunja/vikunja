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
	"time"

	"xorm.io/builder"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

const postgresClock = "clock_timestamp() AT TIME ZONE 'UTC'"

const postgresClaimSQL = `WITH clock AS (SELECT clock_timestamp() AT TIME ZONE 'UTC' AS now),
 candidates AS (
 SELECT id FROM jobs WHERE queue = ? AND state = 'available' AND run_at <= (SELECT now FROM clock)
 ORDER BY run_at, id LIMIT ? FOR UPDATE SKIP LOCKED
 )
 UPDATE jobs SET state = 'running', attempt = attempt + 1,
 claimed_at = (SELECT now FROM clock), claim_owner = ?,
 expires_at = (SELECT now FROM clock) + ? * INTERVAL '1 second'
 FROM candidates WHERE jobs.id = candidates.id
 RETURNING jobs.*`

func execRows(s *xorm.Session, query string, args ...any) (int64, error) {
	result, err := s.Exec(append([]any{query}, args...)...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func updateFailedClaim(s *xorm.Session, job Job, now time.Time, state string, delay time.Duration, lastError string, expired bool) (int64, error) {
	if s.Engine().Dialect().URI().DBType == schemas.POSTGRES {
		query := `WITH clock AS (SELECT clock_timestamp() AT TIME ZONE 'UTC' AS now)
   UPDATE jobs SET state = ?, run_at = clock.now + ? * INTERVAL '1 second', last_error = ?,
   finished_at = CASE WHEN ? = 'failed' THEN clock.now ELSE NULL END
   FROM clock WHERE id = ? AND jobs.state = 'running' AND claim_owner = ? AND attempt = ?`
		if expired {
			query += " AND expires_at <= clock.now"
		}
		return execRows(s, query, state, delay.Seconds(), lastError, state, job.ID, job.ClaimOwner, job.Attempt)
	}
	finished := time.Time{}
	if state == StateFailed {
		finished = now
	}
	s.Where(ownership(job))
	if expired {
		s.And(builder.Lte{"expires_at": timestamp(now)})
	}
	return s.Cols("state", "run_at", "last_error", "finished_at").Update(&Job{State: state, RunAt: now.Add(delay), LastError: lastError, FinishedAt: finished})
}
