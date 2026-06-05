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

package apiv2

import (
	"context"
	"fmt"
	"net/http"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
)

// Assignees are returned as the assigned users, not the join rows:
// models.TaskAssginee.ReadAll yields []*user.User, so that's the element type.
type taskAssigneeListBody struct {
	Body Paginated[*user.User]
}

// RegisterTaskAssigneeRoutes wires the nested TaskAssignee create/list/delete
// onto the Huma API. There is no read-one or update, so no max_permission body.
func RegisterTaskAssigneeRoutes(api huma.API) {
	tags := []string{"assignees"}

	Register(api, huma.Operation{
		OperationID: "task-assignees-list",
		Summary:     "List the assignees of a task",
		Description: "Returns the users assigned to the given task, paginated. Requires read access to the task. Pass q to filter assignees by username.",
		Method:      http.MethodGet,
		Path:        "/tasks/{projecttask}/assignees",
		Tags:        tags,
	}, taskAssigneesList)

	Register(api, huma.Operation{
		OperationID: "task-assignees-create",
		Summary:     "Assign a user to a task",
		Description: "Assigns a user to the given task. The parent task is taken from the URL; the assignee is named by user_id in the body. The assignee must have access to the task's project, and the caller needs write access to the task.",
		Method:      http.MethodPost,
		Path:        "/tasks/{projecttask}/assignees",
		Tags:        tags,
	}, taskAssigneesCreate)

	Register(api, huma.Operation{
		OperationID: "task-assignees-delete",
		Summary:     "Remove an assignee from a task",
		Description: "Un-assigns a user from the given task, identified by their user id in the path. Requires write access to the task.",
		Method:      http.MethodDelete,
		Path:        "/tasks/{projecttask}/assignees/{user}",
		Tags:        tags,
	}, taskAssigneesDelete)
}

func init() { AddRouteRegistrar(RegisterTaskAssigneeRoutes) }

func taskAssigneesList(ctx context.Context, in *struct {
	TaskID int64 `path:"projecttask"`
	ListParams
}) (*taskAssigneeListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.TaskAssginee{TaskID: in.TaskID}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*user.User)
	if !ok {
		return nil, fmt.Errorf("taskAssignees.ReadAll returned unexpected type %T (expected []*user.User)", result)
	}
	return &taskAssigneeListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

func taskAssigneesCreate(ctx context.Context, in *struct {
	TaskID int64 `path:"projecttask"`
	Body   models.TaskAssginee
}) (*singleBody[models.TaskAssginee], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.TaskID = in.TaskID // URL wins over body
	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.TaskAssginee]{Body: &in.Body}, nil
}

func taskAssigneesDelete(ctx context.Context, in *struct {
	TaskID int64 `path:"projecttask"`
	UserID int64 `path:"user"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.TaskAssginee{TaskID: in.TaskID, UserID: in.UserID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
