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

package utils

import (
	"time"

	"code.vikunja.io/api/pkg/log"
)

// RetryWithBackoff executes the given function up to 3 times with exponential backoff.
// Delays between retries are 1s, 2s, 4s (total max wait: 7s).
// The name parameter is used for logging to identify what operation is being retried.
// Returns nil on success, or the last error after all retries are exhausted.
func RetryWithBackoff(name string, fn func() error) error {
	const maxRetries = 3
	baseDelay := 1 * time.Second

	var err error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		if attempt < maxRetries {
			delay := baseDelay * time.Duration(1<<(attempt-1)) // exponential: 1s, 2s, 4s
			log.Warningf("%s not available (attempt %d/%d), retrying in %v: %s",
				name, attempt, maxRetries, delay, err)
			time.Sleep(delay)
		}
	}

	log.Errorf("%s not available after %d attempts: %s", name, maxRetries, err)
	return err
}
