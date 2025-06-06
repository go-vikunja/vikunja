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
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// CanCreate checks if a user can subscribe to an entity
func (sb *Subscription) CanCreate(s *xorm.Session, a web.Auth) (can bool, err error) {
	if _, is := a.(*LinkSharing); is {
		return false, ErrGenericForbidden{}
	}

	sb.EntityType = getEntityTypeFromString(sb.Entity)

	switch sb.EntityType {
	case SubscriptionEntityProject:
		l := &Project{ID: sb.EntityID}
		can, _, err = l.CanRead(s, a)
	case SubscriptionEntityTask:
		t := &Task{ID: sb.EntityID}
		can, _, err = t.CanRead(s, a)
	default:
		return false, &ErrUnknownSubscriptionEntityType{EntityType: sb.EntityType}
	}

	return
}

// CanDelete checks if a user can delete a subscription
func (sb *Subscription) CanDelete(s *xorm.Session, a web.Auth) (can bool, err error) {
	if _, is := a.(*LinkSharing); is {
		return false, ErrGenericForbidden{}
	}

	sb.EntityType = getEntityTypeFromString(sb.Entity)

	realSb := &Subscription{}
	exists, err := s.
		Where("entity_id = ? AND entity_type = ? AND user_id = ?", sb.EntityID, sb.EntityType, a.GetID()).
		Get(realSb)
	if err != nil {
		return false, err
	}

	return exists, nil
}
