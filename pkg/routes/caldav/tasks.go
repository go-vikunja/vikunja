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

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/xorm"
)

// findTask resolves uid through c. Tasks the user cannot read or that are not in c are a
// miss, which is (nil, nil).
func findTask(s *xorm.Session, u *user.User, c collectionTarget, uid string) (*models.Task, error) {
	tasks, err := models.GetTasksByUIDs(s, []string{uid}, u)
	if err != nil {
		if models.IsErrTaskDoesNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, nil
	}

	task := tasks[0]
	// Pinned: the empty project of /dav/projects//<uid>.ics skips the check, see resolveTask.
	if c.project.ID == 0 || task.ProjectID == c.project.ID {
		return task, nil
	}
	// The URL project must match the task's real project (GHSA-48ch-p4gq-x46x).
	if !c.isPseudo() {
		return nil, nil
	}

	// tasks.uid has no unique index, so the first candidate may be a same-UID sibling.
	for _, candidate := range tasks {
		in, err := c.contains(s, u, candidate.ID)
		if err != nil {
			return nil, err
		}
		if in {
			return candidate, nil
		}
	}
	return nil, nil
}

// readCollectionTasks loads c.project and its tasks.
func readCollectionTasks(s *xorm.Session, u *user.User, c collectionTarget) ([]*models.TaskWithComments, error) {
	can, err := c.canRead(s, u)
	if err != nil {
		return nil, err
	}
	if !can {
		log.Debugf("User %v tried to access a caldav resource (Project %v) which they are not allowed to access", u.Username, c.project.ID)
		return nil, errNotFound
	}
	if err := c.project.ReadOne(s, u); err != nil {
		return nil, err
	}

	if c.project.Tasks != nil {
		return c.project.Tasks, nil
	}

	tk := models.TaskCollection{ProjectID: c.project.ID}
	iface, _, _, err := tk.ReadAll(s, u, "", 0, -1)
	if err != nil {
		return nil, err
	}
	tasks, ok := iface.([]*models.Task)
	if !ok {
		return nil, fmt.Errorf("unexpected task collection result type %T", iface)
	}

	var projectTasks []*models.TaskWithComments
	for _, t := range tasks {
		projectTasks = append(projectTasks, &models.TaskWithComments{Task: *t})
	}
	c.project.Tasks = projectTasks
	return projectTasks, nil
}
