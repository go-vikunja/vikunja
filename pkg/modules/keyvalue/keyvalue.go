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
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/keyvalue/memory"
	"code.vikunja.io/api/pkg/modules/keyvalue/redis"
)

// Storage defines an interface for saving key-value pairs
type Storage interface {
	Put(key string, value interface{}) (err error)
	Get(key string) (value interface{}, exists bool, err error)
	GetWithValue(key string, value interface{}) (exists bool, err error)
	Del(key string) (err error)
	IncrBy(key string, update int64) (err error)
	DecrBy(key string, update int64) (err error)
	ListKeys(prefix string) ([]string, error)
	DelPrefix(prefix string) error
}

var store Storage

// InitStorage initializes the configured storage backend
func InitStorage() {
	switch config.KeyvalueType.GetString() {
	case "redis":
		store = redis.NewStorage()
	case "memory":
		fallthrough
	default:
		store = memory.NewStorage()
	}
}

// Put puts a value in the storage backend
func Put(key string, value interface{}) error {
	return store.Put(key, value)
}

// Get returns a value from a storage backend
func Get(key string) (value interface{}, exists bool, err error) {
	return store.Get(key)
}

func GetWithValue(key string, value interface{}) (exists bool, err error) {
	return store.GetWithValue(key, value)
}

// Del removes a save value from a storage backend
func Del(key string) (err error) {
	return store.Del(key)
}

// IncrBy increases a value at key by the amount in update
func IncrBy(key string, update int64) (err error) {
	return store.IncrBy(key, update)
}

// DecrBy increases a value at key by the amount in update
func DecrBy(key string, update int64) (err error) {
	return store.DecrBy(key, update)
}

// ListKeys returns all keys beginning with prefix from the configured backend
func ListKeys(prefix string) ([]string, error) {
	return store.ListKeys(prefix)
}

// DelPrefix deletes all keys with the given prefix in the backend
func DelPrefix(prefix string) error {
	return store.DelPrefix(prefix)
}

// Remember returns the value for a key if it exists.
// If the key is not present, it executes fn to calculate the value,
// stores it and then returns it.
func Remember(key string, fn func() (any, error)) (any, error) {
	val, exists, err := Get(key)
	if err != nil {
		return nil, err
	}
	if exists {
		return val, nil
	}

	val, err = fn()
	if err != nil {
		return nil, err
	}

	if err := Put(key, val); err != nil {
		return nil, err
	}

	return val, nil
}
