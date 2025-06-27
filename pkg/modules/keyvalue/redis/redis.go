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

	var v interface{}

	switch value.(type) {
	case int:
		v = value
	case int8:
		v = value
	case int16:
		v = value
	case int32:
		v = value
	case int64:
		v = value
	default:
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err = enc.Encode(value)
		if err != nil {
			return err
		}
		return s.client.Set(context.Background(), key, buf.Bytes(), 0).Err()
	}

	return s.client.Set(context.Background(), key, v, 0).Err()
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

// DecrBy decreases the value saved at key by the amount provided through update
func (s *Storage) DecrBy(key string, update int64) (err error) {
	return s.client.DecrBy(context.Background(), key, update).Err()
}

// ListKeys returns all keys in redis starting with the given prefix
func (s *Storage) ListKeys(prefix string) ([]string, error) {
	ctx := context.Background()
	pattern := prefix + "*"
	var cursor uint64
	var keys []string

	for {
		k, c, err := s.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, k...)
		cursor = c
		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

// DelPrefix removes all keys in redis which start with the given prefix
func (s *Storage) DelPrefix(prefix string) error {
	keys, err := s.ListKeys(prefix)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return s.client.Del(context.Background(), keys...).Err()
}
