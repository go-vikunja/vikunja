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

package caldav

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/xorm"
)

// Project.CanWrite reports archival through its error return; collapse that into a plain denial.
func denyArchived(can bool, err error) (bool, error) {
	if isError[models.ErrProjectIsArchived](err) {
		return false, nil
	}
	return can, err
}

func (c collectionTarget) isPseudo() bool {
	return models.IsPseudoProjectID(c.project.ID)
}

func (c collectionTarget) canRead(s *xorm.Session, u *user.User) (bool, error) {
	can, _, err := c.project.CanRead(s, u)
	return can, err
}

// Read-only shares and pseudo collections are both the model's call, not the route's.
func (c collectionTarget) canWrite(s *xorm.Session, u *user.User) (bool, error) {
	return denyArchived(c.project.CanWrite(s, u))
}

// GHSA-48ch-p4gq-x46x
func (c collectionTarget) holds(s *xorm.Session, u *user.User, task *models.Task) (bool, error) {
	if c.isPseudo() {
		return c.contains(s, u, task.ID)
	}
	return task.ProjectID == c.project.ID, nil
}

// A pseudo collection has no permissions: membership gates, the task's project decides.
func (c collectionTarget) checkWrite(s *xorm.Session, u *user.User, task *models.Task) error {
	if !c.isPseudo() {
		canWrite, err := c.canWrite(s, u)
		if err != nil {
			return err
		}
		if !canWrite {
			return errForbidden
		}
		return nil
	}

	if task == nil {
		return errNotFound
	}
	in, err := c.holds(s, u, task)
	if err != nil {
		return err
	}
	if !in {
		return errNotFound
	}
	return nil
}

// SQLite bind-variable limit is 999.
var membershipQueryChunkSize = 500

// Candidates must be access-gated first: the favorites arm of ReadAll is not project-scoped.
func (c collectionTarget) containsAll(s *xorm.Session, u *user.User, taskIDs []int64) (map[int64]bool, error) {
	members := make(map[int64]bool, len(taskIDs))
	if len(taskIDs) == 0 {
		return members, nil
	}

	for candidates := range slices.Chunk(taskIDs, membershipQueryChunkSize) {
		ids := make([]string, 0, len(candidates))
		for _, id := range candidates {
			ids = append(ids, strconv.FormatInt(id, 10))
		}

		tc := models.TaskCollection{
			ProjectID: c.project.ID,
			Filter:    "id in " + strings.Join(ids, ","),
		}
		result, _, _, err := tc.ReadAll(s, u, "", 0, -1)
		if err != nil {
			// Unreadable filter: a miss for the whole collection, not a 500 aborting the REPORT.
			if models.IsErrGenericForbidden(err) || models.IsErrSavedFilterDoesNotExist(err) {
				return map[int64]bool{}, nil
			}
			return nil, err
		}

		tasks, ok := result.([]*models.Task)
		if !ok {
			return nil, fmt.Errorf("unexpected task collection result type %T", result)
		}
		for _, t := range tasks {
			members[t.ID] = true
		}
	}

	return members, nil
}

func (c collectionTarget) contains(s *xorm.Session, u *user.User, taskID int64) (bool, error) {
	members, err := c.containsAll(s, u, []int64{taskID})
	if err != nil {
		return false, err
	}
	return members[taskID], nil
}
