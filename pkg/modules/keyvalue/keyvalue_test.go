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

package keyvalue

import (
	"errors"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"code.vikunja.io/api/pkg/modules/keyvalue/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRememberReturnsExisting(t *testing.T) {
	store = memory.NewStorage()
	err := Put("foo", "bar")
	require.NoError(t, err)

	called := false
	val, err := Remember("foo", func() (interface{}, error) {
		called = true
		return "baz", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "bar", val)
	assert.False(t, called)
}

func TestRememberComputesAndStores(t *testing.T) {
	store = memory.NewStorage()

	called := 0
	val, err := Remember("foo", func() (interface{}, error) {
		called++
		return "bar", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "bar", val)
	assert.Equal(t, 1, called)

	v, exists, err := Get("foo")
	require.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, "bar", v)

	val, err = Remember("foo", func() (interface{}, error) {
		called++
		return "baz", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "bar", val)
	assert.Equal(t, 1, called)
}

func TestRememberErrorDoesNotStore(t *testing.T) {
	store = memory.NewStorage()

	_, err := Remember("foo", func() (interface{}, error) {
		return nil, errors.New("fail")
	})

	require.Error(t, err)
	_, exists, err2 := Get("foo")
	require.NoError(t, err2)
	assert.False(t, exists)
}

func TestRememberForReturnsCachedWithinTTL(t *testing.T) {
	store = memory.NewStorage()

	called := 0
	fn := func() (int64, error) {
		called++
		return int64(called), nil
	}

	val, err := RememberFor("foo", time.Hour, fn)
	require.NoError(t, err)
	assert.Equal(t, int64(1), val)

	// Still within the TTL, so fn must not be called again.
	val, err = RememberFor("foo", time.Hour, fn)
	require.NoError(t, err)
	assert.Equal(t, int64(1), val)
	assert.Equal(t, 1, called)
}

func TestRememberForRecomputesAfterExpiry(t *testing.T) {
	store = memory.NewStorage()

	// Seed an already-expired value.
	require.NoError(t, Put("foo", expiringValue[int64]{Value: 1, ExpiresAt: time.Now().Add(-time.Minute)}))

	called := 0
	val, err := RememberFor("foo", time.Hour, func() (int64, error) {
		called++
		return 2, nil
	})

	require.NoError(t, err)
	assert.Equal(t, int64(2), val)
	assert.Equal(t, 1, called)
}

func TestRememberForErrorDoesNotStore(t *testing.T) {
	store = memory.NewStorage()

	_, err := RememberFor("foo", time.Hour, func() (int64, error) {
		return 0, errors.New("fail")
	})

	require.Error(t, err)
	_, exists, err2 := Get("foo")
	require.NoError(t, err2)
	assert.False(t, exists)
}

// getWithValueErrorStore simulates a backend that cannot deserialize an existing value
// into the requested type, e.g. a key that held a plain int64 before the cache started
// storing a struct (the pre-refactor metrics counters in Redis).
type getWithValueErrorStore struct {
	*memory.Storage
}

func (s *getWithValueErrorStore) GetWithValue(string, interface{}) (bool, error) {
	return false, errors.New("decode error")
}

func TestRememberForRecomputesWhenStoredValueCannotBeDeserialized(t *testing.T) {
	store = &getWithValueErrorStore{memory.NewStorage()}

	called := 0
	val, err := RememberFor("foo", time.Hour, func() (int64, error) {
		called++
		return 42, nil
	})

	require.NoError(t, err)
	assert.Equal(t, int64(42), val)
	assert.Equal(t, 1, called)
}

func TestPutWithTTL(t *testing.T) {
	run := func(name string, test func(*testing.T)) {
		t.Run(name, func(t *testing.T) { synctest.Test(t, test) })
	}
	run("value exists before the ttl passed", func(t *testing.T) {
		store = memory.NewStorage()
		require.NoError(t, PutWithTTL("foo", "bar", time.Minute))
		time.Sleep(30 * time.Second)

		val, exists, err := Get("foo")
		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, "bar", val)
	})
	run("value is gone after the ttl passed", func(t *testing.T) {
		store = memory.NewStorage()
		require.NoError(t, PutWithTTL("foo", "bar", time.Minute))
		time.Sleep(time.Minute)

		_, exists, err := Get("foo")
		require.NoError(t, err)
		assert.False(t, exists)
	})
	run("expired keys are not listed", func(t *testing.T) {
		store = memory.NewStorage()
		require.NoError(t, PutWithTTL("prefix_expired", "bar", time.Minute))
		require.NoError(t, PutWithTTL("prefix_valid", "bar", time.Hour))
		time.Sleep(time.Minute)

		keys, err := ListKeys("prefix_")
		require.NoError(t, err)
		assert.Equal(t, []string{"prefix_valid"}, keys)
	})
	run("put clears an existing ttl", func(t *testing.T) {
		store = memory.NewStorage()
		require.NoError(t, PutWithTTL("foo", "bar", time.Minute))
		require.NoError(t, Put("foo", "baz"))
		time.Sleep(time.Minute)

		val, exists, err := Get("foo")
		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, "baz", val)
	})
	run("incrementing after ttl expiry starts fresh", func(t *testing.T) {
		store = memory.NewStorage()
		require.NoError(t, PutWithTTL("c", int64(5), time.Minute))
		time.Sleep(time.Minute)
		require.NoError(t, IncrBy("c", 1))

		val, exists, err := Get("c")
		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, int64(1), val)
	})
}

func TestPutIfAbsentConcurrent(t *testing.T) {
	store = memory.NewStorage()
	const attempts = 64
	start := make(chan struct{})
	claimed := make(chan bool, attempts)
	errors := make(chan error, attempts)
	var ready sync.WaitGroup
	ready.Add(attempts)
	for range attempts {
		go func() {
			ready.Done()
			<-start
			ok, err := PutIfAbsent("claim", "first", time.Minute)
			claimed <- ok
			errors <- err
		}()
	}
	ready.Wait()
	close(start)

	winners := 0
	for range attempts {
		require.NoError(t, <-errors)
		if <-claimed {
			winners++
		}
	}
	require.Equal(t, 1, winners)
}

func TestPutIfAbsentPreservesValueUntilExpiry(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		store = memory.NewStorage()
		ok, err := PutIfAbsent("claim", "first", time.Minute)
		require.NoError(t, err)
		require.True(t, ok)

		ok, err = PutIfAbsent("claim", "second", time.Hour)
		require.NoError(t, err)
		require.False(t, ok)
		value, exists, err := Get("claim")
		require.NoError(t, err)
		require.True(t, exists)
		require.Equal(t, "first", value)

		time.Sleep(time.Minute)
		ok, err = PutIfAbsent("claim", "second", time.Minute)
		require.NoError(t, err)
		require.True(t, ok)
	})
}

func TestPutIfAbsentRejectsNonpositiveTTL(t *testing.T) {
	store = memory.NewStorage()
	require.NoError(t, Put("claim", "first"))
	for _, ttl := range []time.Duration{0, -time.Second} {
		ok, err := PutIfAbsent("claim", "second", ttl)
		require.Error(t, err)
		require.False(t, ok)
	}
	value, exists, err := Get("claim")
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, "first", value)
}
