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
)

const (
	projectAccessCacheTTL       = 30 * time.Second
	projectAccessCacheKeyPrefix = "project-access-"
)

func projectAccessCacheKey(userID int64) string {
	return projectAccessCacheKeyPrefix + strconv.FormatInt(userID, 10)
}

func invalidateProjectAccess(userID int64) {
	if userID == 0 {
		invalidateAllProjectAccess()
		return
	}
	db.InvalidateShared(projectAccessCacheKey(userID))
}

func invalidateAllProjectAccess() {
	db.InvalidateSharedPrefix(projectAccessCacheKeyPrefix)
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
