//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/web"
	"time"
)

// ReadAll gets all labels a user can use
// @Summary Get all labels a user has access to
// @Description Returns all labels which are either created by the user or associated with a task the user has at least read-access to.
// @tags labels
// @Accept json
// @Produce json
// @Param p query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param s query string false "Search labels by label text."
// @Security JWTKeyAuth
// @Success 200 {array} models.Label "The labels"
// @Failure 500 {object} models.Message "Internal error"
// @Router /labels [get]
func (l *Label) ReadAll(search string, a web.Auth, page int) (ls interface{}, err error) {
	u, err := getUserWithError(a)
	if err != nil {
		return nil, err
	}

	// Get all tasks
	taskIDs, err := getUserTaskIDs(u)
	if err != nil {
		return nil, err
	}

	return getLabelsByTaskIDs(search, u, page, taskIDs, true)
}

// ReadOne gets one label
// @Summary Gets one label
// @Description Returns one label by its ID.
// @tags labels
// @Accept json
// @Produce json
// @Param id path int true "Label ID"
// @Security JWTKeyAuth
// @Success 200 {object} models.Label "The label"
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the label"
// @Failure 404 {object} code.vikunja.io/web.HTTPError "Label not found"
// @Failure 500 {object} models.Message "Internal error"
// @Router /labels/{id} [get]
func (l *Label) ReadOne() (err error) {
	label, err := getLabelByIDSimple(l.ID)
	if err != nil {
		return err
	}
	*l = *label

	user, err := GetUserByID(l.CreatedByID)
	if err != nil {
		return err
	}

	l.CreatedBy = &user
	return
}

func getLabelByIDSimple(labelID int64) (*Label, error) {
	label := Label{}
	exists, err := x.ID(labelID).Get(&label)
	if err != nil {
		return &label, err
	}

	if !exists {
		return &Label{}, ErrLabelDoesNotExist{labelID}
	}
	return &label, err
}

// Helper method to get all task ids a user has
func getUserTaskIDs(u *User) (taskIDs []int64, err error) {
	tasks, err := GetTasksByUser("", u, -1, SortTasksByUnsorted, time.Unix(0, 0), time.Unix(0, 0))
	if err != nil {
		return nil, err
	}

	// make a slice of task ids
	for _, t := range tasks {
		taskIDs = append(taskIDs, t.ID)
	}

	return
}
