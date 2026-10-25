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
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// resolveTaskZone returns the time zone a change to a task in the project projectID is computed
// in: the zone of the user making it, re-fetched
// from the database; then the zone of the owner of that project (its own, never a parent's); then
// the server's. A name that is empty or does not load is skipped, and so is a lookup that fails,
// with a log line. It never returns nil and never fails.
//
// A link share has no zone and falls through to the owner. A JWT user carries no zone in its
// claims, so it is re-fetched. A disabled or locked user keeps its stored zone; a user that cannot
// be loaded falls through to the owner.
func resolveTaskZone(s *xorm.Session, a web.Auth, projectID int64) *time.Location {
	if loc, ok := doerZone(s, a, projectID); ok {
		log.Debugf("Task zone for project %d: %s, from the user", projectID, loc)
		return loc
	}

	if loc, ok := projectOwnerZone(s, projectID); ok {
		log.Debugf("Task zone for project %d: %s, from the owner", projectID, loc)
		return loc
	}

	loc := config.GetTimeZone()
	log.Debugf("Task zone for project %d: %s, from the server", projectID, loc)
	return loc
}

// doerZone returns the stored zone of the user behind a, when it has a usable one. The user is
// loaded from the database; a disabled or locked user's row counts, any other failed lookup is
// logged and skipped (never the auth object's own fields). A link share has no zone.
func doerZone(s *xorm.Session, a web.Auth, projectID int64) (*time.Location, bool) {
	if a == nil {
		return nil, false
	}
	doer, err := GetUserOrLinkShareUser(s, a)
	if err != nil && !user.IsErrUserStatusError(err) {
		log.Warningf("Task zone for project %d: could not load the user %d: %v", projectID, a.GetID(), err)
		return nil, false
	}
	if doer == nil || doer.ID <= 0 {
		return nil, false
	}
	return loadZone(doer.Timezone)
}

// projectOwnerZone returns the zone of the owner of the project projectID, when it has a usable one.
func projectOwnerZone(s *xorm.Session, projectID int64) (*time.Location, bool) {
	p, err := GetProjectSimpleByID(s, projectID)
	if err != nil {
		log.Warningf("Task zone for project %d: could not load the project: %v", projectID, err)
		return nil, false
	}
	owner, err := user.GetUserByID(s, p.OwnerID)
	if err != nil && !user.IsErrUserStatusError(err) {
		log.Warningf("Task zone for project %d: could not load its owner %d: %v", projectID, p.OwnerID, err)
		return nil, false
	}
	return loadZone(owner.Timezone)
}

// zoneCache holds every zone loadZone has loaded, by name (*time.Location values).
var zoneCache sync.Map

// loadZone returns the zone named name, loading it once. An empty name, "Local" (the process
// zone, which must never decide a task's dates) and a name time.LoadLocation rejects are not usable.
func loadZone(name string) (*time.Location, bool) {
	if name == "" || name == "Local" {
		return nil, false
	}
	if loc, ok := zoneCache.Load(name); ok {
		return loc.(*time.Location), true
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		log.Debugf("Time zone %q is not usable: %v", name, err)
		return nil, false
	}
	actual, _ := zoneCache.LoadOrStore(name, loc)
	return actual.(*time.Location), true
}
