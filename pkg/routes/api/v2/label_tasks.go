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
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
)

// Element type is *models.LabelWithTaskID because that's what
// models.LabelTask.ReadAll returns; TaskID is json:"-", so the wire shape
// matches plain Label.
type labelTaskListBody struct {
	Body Paginated[*models.LabelWithTaskID]
}

// RegisterLabelTaskRoutes wires the nested labels-on-a-task routes onto the
// Huma API: list, attach and detach. There is no read-one or update — a
// label-task is just a relation, so it has no max_permission.
func RegisterLabelTaskRoutes(api huma.API) {
	tags := []string{"labels"}

	Register(api, huma.Operation{
		OperationID: "task-labels-list",
		Summary:     "List the labels on a task",
		Description: "Returns the labels attached to the given task, paginated. Requires read access to the task.",
		Method:      http.MethodGet,
		Path:        "/tasks/{projecttask}/labels",
		Tags:        tags,
	}, labelTasksList)

	Register(api, huma.Operation{
		OperationID: "task-labels-create",
		Summary:     "Add a label to a task",
		Description: "Attaches an existing label to the given task. Requires write access to the task and access to the label. Fails if the label is already on the task.",
		Method:      http.MethodPost,
		Path:        "/tasks/{projecttask}/labels",
		Tags:        tags,
	}, labelTasksCreate)

	Register(api, huma.Operation{
		OperationID: "task-labels-delete",
		Summary:     "Remove a label from a task",
		Description: "Detaches a label from the given task. Requires write access to the task.",
		Method:      http.MethodDelete,
		Path:        "/tasks/{projecttask}/labels/{label}",
		Tags:        tags,
	}, labelTasksDelete)
}

func init() { AddRouteRegistrar(RegisterLabelTaskRoutes) }

func labelTasksList(ctx context.Context, in *struct {
	TaskID int64 `path:"projecttask"`
	ListParams
}) (*labelTaskListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.LabelTask{TaskID: in.TaskID}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*models.LabelWithTaskID)
	if !ok {
		return nil, fmt.Errorf("labelTasks.ReadAll returned unexpected type %T (expected []*models.LabelWithTaskID)", result)
	}
	return &labelTaskListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

func labelTasksCreate(ctx context.Context, in *struct {
	TaskID int64 `path:"projecttask"`
	Body   models.LabelTask
}) (*singleBody[models.LabelTask], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.TaskID = in.TaskID // parent from the path, not the body
	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.LabelTask]{Body: &in.Body}, nil
}

func labelTasksDelete(ctx context.Context, in *struct {
	TaskID  int64 `path:"projecttask"`
	LabelID int64 `path:"label"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.LabelTask{TaskID: in.TaskID, LabelID: in.LabelID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
