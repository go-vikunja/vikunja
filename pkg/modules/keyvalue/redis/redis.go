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
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/red"
	"github.com/redis/go-redis/v9"
)

// Storage is a redis implementation of a keyvalue storage
type Storage struct {
	client *redis.Client
}

// NewStorage creates a new redis key value storage
func NewStorage() *Storage {
	red.InitRedis()

	return &Storage{
		client: red.GetRedis(),
	}
}

// Put puts a value into redis
func (s *Storage) Put(key string, value interface{}) (err error) {
	return s.put(key, value, 0)
}

// PutWithTTL puts a value into redis which redis expires after ttl
func (s *Storage) PutWithTTL(key string, value interface{}, ttl time.Duration) (err error) {
	return s.put(key, value, ttl)
}

func marshalValue(value interface{}) (interface{}, error) {
	switch value.(type) {
	case int, int8, int16, int32, int64:
		return value, nil
	default:
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		if err := enc.Encode(value); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
}

func (s *Storage) put(key string, value interface{}, ttl time.Duration) error {
	v, err := marshalValue(value)
	if err != nil {
		return err
	}
	return s.client.Set(context.Background(), key, v, ttl).Err()
}

// PutIfAbsent atomically claims a key until its positive TTL expires.
func (s *Storage) PutIfAbsent(key string, value interface{}, ttl time.Duration) (bool, error) {
	if ttl <= 0 {
		return false, errors.New("keyvalue claim TTL must be positive")
	}
	v, err := marshalValue(value)
	if err != nil {
		return false, err
	}
	return s.client.SetNX(context.Background(), key, v, ttl).Result()
}

// Get retrieves a saved value from redis
func (s *Storage) Get(key string) (value interface{}, exists bool, err error) {
	value, err = s.client.Get(context.Background(), key).Result()
	if err != nil && errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	return value, true, err
}

func (s *Storage) GetWithValue(key string, value interface{}) (exists bool, err error) {
	b, err := s.client.Get(context.Background(), key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}

		return
	}

	var buf bytes.Buffer
	_, err = buf.Write(b)
	if err != nil {
		return
	}

	dec := gob.NewDecoder(&buf)
	err = dec.Decode(value)
	return true, err
}

// Del removed a value from redis
func (s *Storage) Del(key string) (err error) {
	return s.client.Del(context.Background(), key).Err()
}

// IncrBy increases the value saved at key by the amount provided through update
func (s *Storage) IncrBy(key string, update int64) (err error) {
	return s.client.IncrBy(context.Background(), key, update).Err()
}

var incrByWithTTLScript = redis.NewScript(`
local v = redis.call('INCRBY', KEYS[1], ARGV[1])
if redis.call('PTTL', KEYS[1]) == -1 then
	redis.call('PEXPIRE', KEYS[1], ARGV[2])
end
return v
`)

// Scripted so a concurrent first increment can't lose the ttl.
func (s *Storage) IncrByWithTTL(key string, update int64, ttl time.Duration) (int64, error) {
	return incrByWithTTLScript.Run(context.Background(), s.client, []string{key}, update, ttl.Milliseconds()).Int64()
}

// DecrBy decreases the value saved at key by the amount provided through update
func (s *Storage) DecrBy(key string, update int64) (err error) {
	return s.client.DecrBy(context.Background(), key, update).Err()
}

var globEscaper = strings.NewReplacer("\\", "\\\\", "*", "\\*", "?", "\\?", "[", "\\[", "]", "\\]")

// ListKeys returns all keys in redis starting with the given prefix
func (s *Storage) ListKeys(prefix string) ([]string, error) {
	ctx := context.Background()
	// Match the prefix literally like the memory backend does; user-derived prefixes may contain glob chars.
	pattern := globEscaper.Replace(prefix) + "*"
	var cursor uint64
	var keys []string
	// SCAN may return duplicates
	seen := make(map[string]struct{})

	for {
		k, c, err := s.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}
		for _, key := range k {
			if _, has := seen[key]; has {
				continue
			}
			seen[key] = struct{}{}
			keys = append(keys, key)
		}
		cursor = c
		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

// DelPrefix removes all keys in redis which start with the given prefix
func (s *Storage) DelPrefix(prefix string) error {
	ctx := context.Background()
	pattern := globEscaper.Replace(prefix) + "*"
	var cursor uint64

	// UNLINK per page so a large prefix neither blocks Redis nor piles up in memory here.
	for {
		keys, next, err := s.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := s.client.Unlink(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			return nil
		}
	}
}
