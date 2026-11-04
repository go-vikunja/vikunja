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
	"time"

	"xorm.io/xorm"
)

func advanceProjectPurgeMarker(s *xorm.Session, projectID int64) error {
	project := &Project{}
	found, err := lockingSession(s).ID(projectID).Get(project)
	if err != nil {
		return err
	}
	if !found {
		return ErrProjectDoesNotExist{ID: projectID}
	}

	marker := time.Now().UTC().Truncate(time.Second)
	// Whole seconds keep purge generations distinct across database precisions.
	if !marker.After(project.LastPurgedAt) {
		marker = project.LastPurgedAt.Add(time.Second)
	}
	_, err = s.ID(projectID).Cols("last_purged_at").NoAutoTime().Update(&Project{
		LastPurgedAt: marker,
	})
	return err
}
