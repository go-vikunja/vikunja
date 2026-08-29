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

// GetCached returns a value memoized for the lifetime of s. It always misses
// once s has written anything, so a memo can never outlive the data it derives from.
func GetCached[T any](s *xorm.Session, key string) (value T, found bool) {
	c, ok := cacheForSession(s)
	if !ok {
		return value, false
	}
	v, has := c.get(key)
	if !has {
		return value, false
	}
	value, found = v.(T)
	return value, found
}

// SetCached memoizes value under key for the lifetime of s.
func SetCached[T any](s *xorm.Session, key string, value T) {
	if c, ok := cacheForSession(s); ok {
		c.set(key, value)
	}
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
