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

package models

import (
	"sync"
	"sync/atomic"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"
)

// The resolved access map of a user is reused across requests until something that feeds the grants CTE
// changes: projects (owner, parent), users_projects, team_projects, team_members. The xorm After* hooks on
// those models call invalidate*; xorm runs them after the transaction committed, so a concurrent reader
// can never re-fill the cache with pre-commit data under the new generation.
//
// Within one process invalidation is exact. Across replicas the generation is published through keyvalue
// and pulled at most once per second (a round trip per request would cost more than the query), and every
// entry expires after projectAccessCacheTTL as the bound for replicas without a shared keyvalue store.
const (
	projectAccessCacheTTL          = 30 * time.Second
	projectAccessCacheMaxSize      = 10000
	projectAccessGenerationKey     = "project-access-generation"
	projectAccessGenerationSyncGap = time.Second
)

type projectAccessEntry struct {
	access     *projectAccess
	generation int64
	resolved   time.Time
}

type projectAccessCacheStore struct {
	mu      sync.Mutex
	entries map[int64]projectAccessEntry

	generation atomic.Int64
	// bookkeeping for the shared counter: increments we published ourselves are not foreign writes
	ownWrites  atomic.Int64
	lastShared atomic.Int64
	lastSync   atomic.Int64
}

var projectAccessCache = &projectAccessCacheStore{entries: map[int64]projectAccessEntry{}}

func init() {
	db.RegisterGlobalCacheReset(projectAccessCache.reset)
}

func (c *projectAccessCacheStore) reset() {
	c.mu.Lock()
	c.entries = map[int64]projectAccessEntry{}
	c.mu.Unlock()
	c.generation.Add(1)
}

func (c *projectAccessCacheStore) get(userID int64) (*projectAccess, bool) {
	gen := c.currentGeneration()
	c.mu.Lock()
	defer c.mu.Unlock()
	e, has := c.entries[userID]
	if !has {
		return nil, false
	}
	if e.generation != gen || time.Since(e.resolved) > projectAccessCacheTTL {
		delete(c.entries, userID)
		return nil, false
	}
	return e.access, true
}

func (c *projectAccessCacheStore) put(userID int64, pa *projectAccess) {
	gen := c.currentGeneration()
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= projectAccessCacheMaxSize {
		c.entries = map[int64]projectAccessEntry{}
	}
	c.entries[userID] = projectAccessEntry{access: pa, generation: gen, resolved: time.Now()}
}

// invalidateUser drops one user's entry. Other replicas only learn about it through the TTL.
func (c *projectAccessCacheStore) invalidateUser(userID int64) {
	if userID == 0 {
		c.invalidateAll()
		return
	}
	c.mu.Lock()
	delete(c.entries, userID)
	c.mu.Unlock()
}

// invalidateAll starts a new generation here and publishes it for other replicas.
func (c *projectAccessCacheStore) invalidateAll() {
	c.generation.Add(1)
	if !keyvalue.Initialized() {
		return
	}
	c.ownWrites.Add(1)
	if err := keyvalue.IncrBy(projectAccessGenerationKey, 1); err != nil {
		log.Debugf("could not publish project access generation: %s", err)
	}
}

// currentGeneration folds in generations published by other replicas, at most once per sync gap.
func (c *projectAccessCacheStore) currentGeneration() int64 {
	now := time.Now().UnixNano()
	last := c.lastSync.Load()
	if now-last < int64(projectAccessGenerationSyncGap) || !c.lastSync.CompareAndSwap(last, now) || !keyvalue.Initialized() {
		return c.generation.Load()
	}
	var shared int64
	exists, err := keyvalue.GetWithValue(projectAccessGenerationKey, &shared)
	if err != nil || !exists {
		return c.generation.Load()
	}
	own := c.ownWrites.Swap(0)
	// More increments than we published ourselves means another replica invalidated.
	if shared-c.lastShared.Swap(shared) > own {
		c.generation.Add(1)
	}
	return c.generation.Load()
}

// Hooks. xorm calls them after commit inside a transaction, immediately otherwise.

func (p *Project) AfterInsert() {
	// A new child is reachable for everyone with access to the parent.
	if p.ParentProjectID != nil {
		projectAccessCache.invalidateAll()
		return
	}
	projectAccessCache.invalidateUser(p.OwnerID)
}

func (p *Project) AfterUpdate() {
	// Only a parent change moves the subtree between access trees; a bean without a parent cannot have changed it.
	if p.ParentProjectID != nil {
		projectAccessCache.invalidateAll()
		return
	}
	projectAccessCache.invalidateUser(p.OwnerID)
}

func (p *Project) AfterDelete() { projectAccessCache.invalidateAll() }

func (lu *ProjectUser) AfterInsert() { projectAccessCache.invalidateUser(lu.UserID) }
func (lu *ProjectUser) AfterUpdate() { projectAccessCache.invalidateUser(lu.UserID) }
func (lu *ProjectUser) AfterDelete() { projectAccessCache.invalidateUser(lu.UserID) }

func (tl *TeamProject) AfterInsert() { projectAccessCache.invalidateAll() }
func (tl *TeamProject) AfterUpdate() { projectAccessCache.invalidateAll() }
func (tl *TeamProject) AfterDelete() { projectAccessCache.invalidateAll() }

func (tm *TeamMember) AfterInsert() { projectAccessCache.invalidateUser(tm.UserID) }
func (tm *TeamMember) AfterUpdate() { projectAccessCache.invalidateUser(tm.UserID) }
func (tm *TeamMember) AfterDelete() { projectAccessCache.invalidateUser(tm.UserID) }

func (t *Team) AfterDelete() { projectAccessCache.invalidateAll() }
