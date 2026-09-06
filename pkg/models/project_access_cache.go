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
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"
)

const (
	projectAccessCacheTTL       = 30 * time.Second
	projectAccessCacheKeyPrefix = "project_access_"
)

func init() {
	db.RegisterGlobalCacheReset(invalidateAllProjectAccess)
}

func projectAccessCacheKey(userID int64) string {
	return projectAccessCacheKeyPrefix + strconv.FormatInt(userID, 10)
}

func getCachedProjectAccess(userID int64) (map[int64]Permission, bool) {
	if !keyvalue.Initialized() {
		return nil, false
	}

	var permissions map[int64]Permission
	exists, err := keyvalue.GetWithValue(projectAccessCacheKey(userID), &permissions)
	if err != nil {
		log.Debugf("could not read cached project access for user %d: %s", userID, err)
		return nil, false
	}
	return permissions, exists
}

func cacheProjectAccess(userID int64, permissions map[int64]Permission) {
	if !keyvalue.Initialized() {
		return
	}

	if err := keyvalue.PutWithTTL(projectAccessCacheKey(userID), permissions, projectAccessCacheTTL); err != nil {
		log.Debugf("could not cache project access for user %d: %s", userID, err)
	}
}

func invalidateProjectAccess(userID int64) {
	if userID == 0 {
		invalidateAllProjectAccess()
		return
	}

	if !keyvalue.Initialized() {
		return
	}

	if err := keyvalue.Del(projectAccessCacheKey(userID)); err != nil {
		log.Errorf("could not invalidate cached project access for user %d: %s", userID, err)
	}
}

func invalidateAllProjectAccess() {
	if !keyvalue.Initialized() {
		return
	}

	if err := keyvalue.DelPrefix(projectAccessCacheKeyPrefix); err != nil {
		log.Errorf("could not invalidate cached project access: %s", err)
	}
}

// Hooks. xorm calls them after commit inside a transaction, immediately otherwise.

func (p *Project) AfterInsert() {
	// A new child is reachable for everyone with access to the parent.
	if p.ParentProjectID != nil {
		invalidateAllProjectAccess()
		return
	}
	invalidateProjectAccess(p.OwnerID)
}

func (p *Project) AfterUpdate() {
	// Only a parent change moves the subtree between access trees; a bean without a parent cannot have changed it.
	if p.ParentProjectID != nil {
		invalidateAllProjectAccess()
		return
	}
	invalidateProjectAccess(p.OwnerID)
}

func (p *Project) AfterDelete() { invalidateAllProjectAccess() }

func (lu *ProjectUser) AfterInsert() { invalidateProjectAccess(lu.UserID) }
func (lu *ProjectUser) AfterUpdate() { invalidateProjectAccess(lu.UserID) }
func (lu *ProjectUser) AfterDelete() { invalidateProjectAccess(lu.UserID) }

func (tl *TeamProject) AfterInsert() { invalidateAllProjectAccess() }
func (tl *TeamProject) AfterUpdate() { invalidateAllProjectAccess() }
func (tl *TeamProject) AfterDelete() { invalidateAllProjectAccess() }

func (tm *TeamMember) AfterInsert() { invalidateProjectAccess(tm.UserID) }
func (tm *TeamMember) AfterUpdate() { invalidateProjectAccess(tm.UserID) }
func (tm *TeamMember) AfterDelete() { invalidateProjectAccess(tm.UserID) }

func (t *Team) AfterDelete() { invalidateAllProjectAccess() }
