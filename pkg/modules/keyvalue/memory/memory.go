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

package memory

import (
	"reflect"
	"strings"
	"sync"

	e "code.vikunja.io/api/pkg/modules/keyvalue/error"
)

// Storage is the memory implementation of a storage backend
type Storage struct {
	store map[string]interface{}
	mutex sync.Mutex
}

// NewStorage creates a new memory storage
func NewStorage() *Storage {
	s := &Storage{}
	s.store = make(map[string]interface{})
	return s
}

// Put puts a value into the memory storage
func (s *Storage) Put(key string, value interface{}) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	val := reflect.ValueOf(value)
	// Make sure to store the underlying value when value is a pointer to a value
	if val.Kind() == reflect.Ptr {
		s.store[key] = val.Elem().Interface()
		return nil
	}

	s.store[key] = value
	return nil
}

// Get retrieves a saved value from memory storage
func (s *Storage) Get(key string) (value interface{}, exists bool, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	value, exists = s.store[key]
	return
}

func (s *Storage) GetWithValue(key string, ptr interface{}) (exists bool, err error) {
	stored, exists, err := s.Get(key)
	if !exists {
		return exists, err
	}

	val := reflect.ValueOf(ptr)
	if val.Kind() != reflect.Ptr {
		panic("value must be a pointer")
	}
	if val.IsNil() {
		panic("pointer must not be a nil-pointer")
	}

	val.Elem().Set(reflect.ValueOf(stored))

	return exists, err
}

// Del removes a saved value from a memory storage
func (s *Storage) Del(key string) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.store, key)
	return nil
}

// IncrBy increases the value saved at key by the amount provided through update
// It assumes the value saved for the key either does not exist or has a type of int64
func (s *Storage) IncrBy(key string, update int64) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.store[key]
	if !exists {
		s.store[key] = int64(0)
	}

	val, is := s.store[key].(int64)
	if !is {
		return &e.ErrValueHasWrongType{Key: key, ExpectedValue: "int64"}
	}
	s.store[key] = val + update
	return nil
}

// DecrBy decreases the value saved at key by the amount provided through update
// It assumes the value saved for the key either does not exist or has a type of int64
func (s *Storage) DecrBy(key string, update int64) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.store[key]
	if !exists {
		s.store[key] = int64(0)
	}

	val, is := s.store[key].(int64)
	if !is {
		return &e.ErrValueHasWrongType{Key: key, ExpectedValue: "int64"}
	}
	s.store[key] = val - update
	return nil
}

// ListKeys returns all keys in the storage which start with the given prefix
func (s *Storage) ListKeys(prefix string) ([]string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	keys := make([]string, 0)
	for k := range s.store {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}

	return keys, nil
}

// DelPrefix removes all keys which start with the given prefix
func (s *Storage) DelPrefix(prefix string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for k := range s.store {
		if strings.HasPrefix(k, prefix) {
			delete(s.store, k)
		}
	}

	return nil
}
