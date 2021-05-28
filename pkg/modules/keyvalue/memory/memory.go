// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package memory

import (
	"reflect"
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

func (s *Storage) GetWithValue(key string, value interface{}) (exists bool, err error) {
	v, exists, err := s.Get(key)
	if !exists {
		return exists, err
	}

	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Ptr {
		panic("some: check must be a pointer")
	}

	val.Elem().Set(reflect.ValueOf(v))
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
