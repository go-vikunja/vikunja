// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licence as published by
// the Free Software Foundation, either version 3 of the Licence, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licence for more details.
//
// You should have received a copy of the GNU Affero General Public Licence
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"crypto/sha256"
	"sync"
	"time"
)

// PBKDF2 over the raw token costs ~3 ms of CPU, and API clients send the same token on every request.
// Remembering the digest of a verified token skips that; the row is still loaded by id on every request,
// so revocation and expiry apply immediately.
const (
	verifiedTokenTTL     = 10 * time.Minute
	verifiedTokenMaxSize = 10000
)

type verifiedToken struct {
	id       int64
	hash     string
	verified time.Time
}

type verifiedTokenCache struct {
	mu      sync.Mutex
	entries map[[sha256.Size]byte]verifiedToken
}

var verifiedTokens = &verifiedTokenCache{entries: map[[sha256.Size]byte]verifiedToken{}}

func (c *verifiedTokenCache) get(token string) (verifiedToken, bool) {
	key := sha256.Sum256([]byte(token))
	c.mu.Lock()
	defer c.mu.Unlock()
	v, has := c.entries[key]
	if !has {
		return verifiedToken{}, false
	}
	if time.Since(v.verified) > verifiedTokenTTL {
		delete(c.entries, key)
		return verifiedToken{}, false
	}
	return v, true
}

func (c *verifiedTokenCache) put(token string, t *APIToken) {
	key := sha256.Sum256([]byte(token))
	c.mu.Lock()
	defer c.mu.Unlock()
	// A full cache means more distinct tokens than any real deployment has; dropping everything is cheaper than LRU bookkeeping.
	if len(c.entries) >= verifiedTokenMaxSize {
		c.entries = map[[sha256.Size]byte]verifiedToken{}
	}
	c.entries[key] = verifiedToken{id: t.ID, hash: t.TokenHash, verified: time.Now()}
}

func (c *verifiedTokenCache) forget(token string) {
	key := sha256.Sum256([]byte(token))
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}
