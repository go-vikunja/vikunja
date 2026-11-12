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

package redis

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestRedisPutIfAbsentAcrossClients(t *testing.T) {
	addr := os.Getenv("VIKUNJA_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("set VIKUNJA_TEST_REDIS_ADDR to run Redis integration tests")
	}
	clients := []*redis.Client{
		redis.NewClient(&redis.Options{Addr: addr}),
		redis.NewClient(&redis.Options{Addr: addr}),
	}
	key := fmt.Sprintf("test:keyvalue:%d", time.Now().UnixNano())
	ctx := context.Background()
	t.Cleanup(func() {
		_ = clients[0].Del(ctx, key).Err()
		for _, client := range clients {
			_ = client.Close()
		}
	})

	type result struct {
		claimed bool
		err     error
	}
	const attempts = 64
	results := make(chan result, attempts)
	start := make(chan struct{})
	for i := range attempts {
		go func() {
			<-start
			s := &Storage{client: clients[i%len(clients)]}
			claimed, err := s.PutIfAbsent(key, "first", time.Minute)
			results <- result{claimed, err}
		}()
	}
	close(start)
	winners := 0
	for range attempts {
		result := <-results
		require.NoError(t, result.err)
		if result.claimed {
			winners++
		}
	}
	require.Equal(t, 1, winners)

	s := &Storage{client: clients[1]}
	claimed, err := s.PutIfAbsent(key, "second", time.Hour)
	require.NoError(t, err)
	require.False(t, claimed)
	var value string
	exists, err := s.GetWithValue(key, &value)
	require.NoError(t, err)
	require.True(t, exists)
	require.Equal(t, "first", value)
	require.LessOrEqual(t, clients[0].PTTL(ctx, key).Val(), time.Minute)

	require.NoError(t, clients[0].PExpire(ctx, key, time.Millisecond).Err())
	require.Eventually(t, func() bool {
		return clients[0].Exists(ctx, key).Val() == 0
	}, time.Second, time.Millisecond)
	claimed, err = s.PutIfAbsent(key, "second", time.Minute)
	require.NoError(t, err)
	require.True(t, claimed)
}

func TestRedisPutIfAbsentUnavailable(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:0"})
	require.NoError(t, client.Close())
	s := &Storage{client: client}
	claimed, err := s.PutIfAbsent("claim", true, time.Minute)
	require.Error(t, err)
	require.False(t, claimed)
}
