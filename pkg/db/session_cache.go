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
	"context"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"weak"

	"xorm.io/xorm"
	"xorm.io/xorm/contexts"
)

// xorm.Session has no context getter, so the cache is also kept in a weak map
// keyed by the session pointer; the entry dies with the session.
var sessionCaches sync.Map // weak.Pointer[xorm.Session] -> *sessionCache

type sessionCacheCtxKeyType struct{}

var sessionCacheCtxKey sessionCacheCtxKeyType

type sessionCache struct {
	mu     sync.Mutex
	dirty  bool
	values map[string]any
}

func (c *sessionCache) get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.dirty {
		return nil, false
	}
	v, has := c.values[key]
	return v, has
}

func (c *sessionCache) set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.dirty {
		return
	}
	c.values[key] = value
}

func (c *sessionCache) markDirty() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dirty = true
	clear(c.values)
}

// Sessions created any other way never memoize.
func attachSessionCache(s *xorm.Session) {
	c := &sessionCache{values: map[string]any{}}
	key := weak.Make(s)
	sessionCaches.Store(key, c)
	runtime.AddCleanup(s, func(k weak.Pointer[xorm.Session]) { sessionCaches.Delete(k) }, key)
	s.Context(context.WithValue(context.Background(), sessionCacheCtxKey, c))
}

// SetSessionContext replaces the context of s while preserving its memo. Calling
// s.Context directly instead drops the cache pointer, which leaves the memo
// serving reads that writes on s no longer invalidate.
func SetSessionContext(ctx context.Context, s *xorm.Session) {
	if c, ok := cacheForSession(s); ok {
		ctx = context.WithValue(ctx, sessionCacheCtxKey, c)
	}
	s.Context(ctx)
}

func cacheForSession(s *xorm.Session) (*sessionCache, bool) {
	c, ok := sessionCaches.Load(weak.Make(s))
	if !ok {
		return nil, false
	}
	return c.(*sessionCache), true
}

// For writes that bypass xorm entirely, which writeInvalidationHook never sees.
func invalidateAllSessionCaches() {
	sessionCaches.Range(func(_, v any) bool {
		v.(*sessionCache).markDirty()
		return true
	})
}

// Remember returns the value memoized under key for the lifetime of s, running fetch on a miss.
// The memo always misses once s has written anything, so it can never outlive the data it derives
// from. Nothing is stored when fetch fails. Reference-typed fields of a memoized value stay shared
// with the memo, so callers that hand out pointers must copy on the way out.
func Remember[T any](s *xorm.Session, key string, fetch func() (T, error)) (T, error) {
	c, ok := cacheForSession(s)
	if !ok {
		return fetch()
	}

	if v, has := c.get(key); has {
		if value, is := v.(T); is {
			return value, nil
		}
	}

	value, err := fetch()
	if err != nil {
		var zero T
		return zero, err
	}
	c.set(key, value)
	return value, nil
}

// RememberEach is Remember for id-keyed batch loads: memoized ids are served from the memo and
// fetch runs once for the rest. Ids fetch does not return stay absent from both the memo and the
// result.
func RememberEach[T any](s *xorm.Session, ids []int64, key func(int64) string, fetch func(missing []int64) (map[int64]T, error)) (map[int64]T, error) {
	if len(ids) == 0 {
		return map[int64]T{}, nil
	}

	c, ok := cacheForSession(s)
	if !ok {
		return fetch(ids)
	}

	values := make(map[int64]T, len(ids))
	missing := make([]int64, 0, len(ids))
	for _, id := range ids {
		v, has := c.get(key(id))
		if !has {
			missing = append(missing, id)
			continue
		}
		value, is := v.(T)
		if !is {
			missing = append(missing, id)
			continue
		}
		values[id] = value
	}
	if len(missing) == 0 {
		return values, nil
	}

	loaded, err := fetch(missing)
	if err != nil {
		return nil, err
	}
	for id, value := range loaded {
		c.set(key(id), value)
		values[id] = value
	}
	return values, nil
}

type writeInvalidationHook struct{}

func (writeInvalidationHook) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
	if isWriteStatement(c.SQL) {
		if cache, ok := c.Ctx.Value(sessionCacheCtxKey).(*sessionCache); ok {
			cache.markDirty()
		}
	}
	return c.Ctx, nil
}

func (writeInvalidationHook) AfterProcess(*contexts.ContextHook) error { return nil }

// A WITH statement can modify data, so one naming a write verb counts as a write.
var writeVerbInCTE = regexp.MustCompile(`(?i)\b(insert|update|delete|merge|replace|create|drop|alter|truncate)\b`)

var leadingKeyword = regexp.MustCompile(`^[\s(]*([a-zA-Z]+)`)

// Anything not recognizable as a pure read counts as a write.
func isWriteStatement(sqlStr string) bool {
	m := leadingKeyword.FindStringSubmatch(sqlStr)
	if m == nil {
		return true
	}
	switch strings.ToLower(m[1]) {
	case "select", "show", "pragma", "describe", "desc",
		"begin", "commit", "rollback", "savepoint", "release", "prepare", "deallocate":
		return false
	case "with":
		return writeVerbInCTE.MatchString(sqlStr)
	}
	return true
}
