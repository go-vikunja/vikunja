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
	"code.vikunja.io/api/pkg/metrics"
	_ "code.vikunja.io/web" // For swaggerdocs generation
)

// Delete implements the delete method for listTask
// @Summary Delete a task
// @Description Deletes a task from a list. This does not mean "mark it done".
// @tags task
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Task ID"
// @Success 200 {object} models.Message "The created task object."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid task ID provided."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{id} [delete]
func (i *ListTask) Delete() (err error) {

	// Check if it exists
	_, err = GetListTaskByID(i.ID)
	if err != nil {
		return
	}

	if _, err = x.ID(i.ID).Delete(ListTask{}); err != nil {
		return err
	}

	metrics.UpdateCount(-1, metrics.TaskCountKey)
	return
}
