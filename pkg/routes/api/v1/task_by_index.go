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

package v1

// GetTaskByProjectIndex is a doc-only stub: swag allows one @Router per func
// and Task.ReadOne already owns /tasks/{id}, so the by-index route needs its
// own function to host the second annotation. The route is wired directly to
// taskHandler.ReadOneWeb in routes.go.
//
// @Summary Get one task by its per-project index
// @Description Returns a single task identified by its per-project index. Useful when resolving human-readable references like "PROJ-42" to a canonical task object. Note that task indexes are reassigned when a task is moved between projects, so long-lived references should use the returned task id instead.
// @tags task
// @Accept json
// @Produce json
// @Param project path int true "The project ID"
// @Param index path int true "The task's per-project index"
// @Param expand query string false "If set to `subtasks`, Vikunja will fetch only tasks which do not have subtasks and then in a second step, will fetch all of these subtasks. This may result in more tasks than the pagination limit being returned, but all subtasks will be present in the response. You can only set this to `subtasks`."
// @Security JWTKeyAuth
// @Success 200 {object} models.Task "The task"
// @Failure 400 {object} web.HTTPError "Invalid project ID or index"
// @Failure 403 {object} web.HTTPError "The user does not have access to the task"
// @Failure 404 {object} models.Message "Task not found"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/tasks/by-index/{index} [get]
func GetTaskByProjectIndex() {} //nolint:unused
