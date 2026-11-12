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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

// EnqueueBatch shares the caller's transaction and inserts at most 100 rows per statement.
func EnqueueBatch(s *xorm.Session, requests []Request) error {
	for _, request := range requests {
		if request.Queue == "" || request.Kind == "" || !json.Valid(request.Payload) {
			return errors.New("job requires a queue, kind and JSON payload")
		}
	}
	const batchSize = 100
	now := time.Now()
	for start := 0; start < len(requests); {
		end := start + 1
		for end < min(start+batchSize, len(requests)) && requests[end].Queue == requests[start].Queue && requests[end].Kind == requests[start].Kind {
			end++
		}
		batch := requests[start:end]
		start = end
		values := make([]string, 0, len(batch))
		args := make([]interface{}, 0, len(batch)*7+1)
		args = append(args, "")
		for _, request := range batch {
			at := request.RunAt
			if at.IsZero() {
				at = now
			}
			var key interface{}
			if request.UniqueKey != "" {
				hash := sha256.Sum256([]byte(request.Queue + "\x00" + request.Kind + "\x00" + request.UniqueKey))
				key = hex.EncodeToString(hash[:])
			}
			values = append(values, "(?, ?, 1, ?, ?, 0, ?, ?, ?)")
			args = append(args, request.Queue, request.Kind, string(request.Payload), StateAvailable, retryPolicy(request.Queue).Attempts, timestamp(at), key)
		}
		query := "INSERT INTO jobs (queue, kind, version, payload, state, attempt, max_attempts, run_at, unique_key) VALUES " + strings.Join(values, ",")
		if s.Engine().Dialect().URI().DBType == schemas.MYSQL {
			query += " ON DUPLICATE KEY UPDATE unique_key = unique_key"
		} else {
			query += " ON CONFLICT (unique_key) DO NOTHING"
		}
		args[0] = query
		result, err := s.Exec(args...)
		if err != nil {
			return err
		}
		count, err := result.RowsAffected()
		if err != nil {
			return err
		}
		countAfterCommit(s, "enqueued", batch[0].Queue, batch[0].Kind, count)
		for _, request := range batch {
			wakeAfterCommit(s, request.Queue)
		}
	}
	return nil
}
