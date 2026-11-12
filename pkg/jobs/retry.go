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
	"math/rand/v2"
	"time"

	"code.vikunja.io/api/pkg/config"
)

type RetryPolicy struct {
	Attempts     int
	InitialDelay time.Duration
	MaxDelay     time.Duration
}

func retryPolicy(queue string) RetryPolicy {
	policy := RetryPolicy{
		Attempts:     DefaultAttempts,
		InitialDelay: 2 * time.Second,
		MaxDelay:     time.Hour,
	}
	prefix := "jobs.queues." + queue + "."
	if value := config.Key(prefix + "maxattempts").GetInt(); value > 0 {
		policy.Attempts = value
	}
	if value := config.Key(prefix + "retryinitialdelay").GetDuration(); value > 0 {
		policy.InitialDelay = value
	}
	if value := config.Key(prefix + "retrymaxdelay").GetDuration(); value > 0 {
		policy.MaxDelay = value
	}
	return policy
}

func (p RetryPolicy) Delay(attempt int) time.Duration {
	delay := p.InitialDelay
	for remaining := attempt - 1; remaining > 0 && delay < p.MaxDelay; remaining-- {
		if delay > p.MaxDelay/2 {
			delay = p.MaxDelay
			break
		}
		delay *= 2
	}
	seconds := int64(min(delay, p.MaxDelay) / time.Second)
	if seconds <= 1 {
		return time.Duration(seconds) * time.Second
	}
	// Whole seconds survive every supported database's datetime precision.
	lower := (seconds + 1) / 2
	return time.Duration(lower+rand.Int64N(seconds-lower+1)) * time.Second //nolint:gosec // Retry jitter is not security-sensitive.
}
