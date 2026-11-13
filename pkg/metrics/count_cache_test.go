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

package metrics

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"code.vikunja.io/api/pkg/modules/keyvalue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricCountsCoalesceMisses(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "failure"}[fail], func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				keyvalue.InitStorage()
				release := make(chan struct{})
				var calls atomic.Int64
				var done sync.WaitGroup
				unavailable := errors.New("count unavailable")
				call := func() {
					value, err := getCachedCount(TaskCountKey, func() (int64, error) {
						calls.Add(1)
						<-release
						if fail {
							return 0, unavailable
						}
						return 42, nil
					})
					if fail {
						assert.ErrorIs(t, err, unavailable)
					} else {
						if err != nil {
							t.Errorf("getCachedCount: %v", err)
							return
						}
						assert.Equal(t, int64(42), value)
					}
				}
				for range 4 {
					done.Go(call)
				}
				synctest.Wait()
				for range 4 {
					done.Go(call)
				}
				synctest.Wait()
				assert.Equal(t, int64(1), calls.Load(), "overlapping scrapes must share one count")
				close(release)
				done.Wait()
				value, err := getCachedCount(TaskCountKey, func() (int64, error) { calls.Add(1); return 43, nil })
				require.NoError(t, err)
				if fail {
					assert.Equal(t, int64(43), value)
					assert.Equal(t, int64(2), calls.Load())
				} else {
					assert.Equal(t, int64(42), value)
					assert.Equal(t, int64(1), calls.Load())
				}
			})
		})
	}
}

func TestMetricCountsDelayedCallerRechecksCache(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		keyvalue.InitStorage()
		var calls int
		load := func() (int64, error) { calls++; return int64(calls), nil }
		value, err := getCachedCount(TaskCountKey, load)
		require.NoError(t, err)
		assert.Equal(t, int64(1), value)
		time.Sleep(countCacheTTL - time.Nanosecond)
		value, err = getCachedCount(TaskCountKey, load)
		require.NoError(t, err)
		assert.Equal(t, int64(1), value)
		time.Sleep(time.Nanosecond)
		value, err = getCachedCount(TaskCountKey, load)
		require.NoError(t, err)
		assert.Equal(t, int64(2), value)
		assert.Equal(t, 2, calls)
	})
}

func TestMetricCountsDifferentKeysDoNotBlock(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		keyvalue.InitStorage()
		release := make(chan struct{})
		var done sync.WaitGroup
		done.Go(func() {
			_, err := getCachedCount(TaskCountKey, func() (int64, error) { <-release; return 42, nil })
			assert.NoError(t, err)
		})
		synctest.Wait()
		value, err := getCachedCount(UserCountKey, func() (int64, error) { return 2, nil })
		close(release)
		done.Wait()
		require.NoError(t, err)
		assert.Equal(t, int64(2), value)
	})
}
