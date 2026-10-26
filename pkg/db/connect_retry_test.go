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

package db

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPingWithRetry(t *testing.T) {
	t.Run("single attempt when retry interval is disabled", func(t *testing.T) {
		attempts := 0
		err := pingWithRetry(func() error {
			attempts++
			return errors.New("connection refused")
		}, 0, 0)

		require.Error(t, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("retries until the database responds", func(t *testing.T) {
		attempts := 0
		start := time.Now()
		err := pingWithRetry(func() error {
			attempts++
			if attempts < 3 {
				return errors.New("connection refused")
			}
			return nil
		}, 1, 50)

		require.NoError(t, err)
		assert.Equal(t, 3, attempts)
		assert.GreaterOrEqual(t, time.Since(start), 2*time.Millisecond)
	})
}
