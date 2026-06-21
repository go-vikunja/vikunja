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
	"net/http"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
)

// RegisterTaskAssigneeBulkRoutes wires the bulk assignee replacement onto the
// Huma API. PUT is the honest verb — the operation replaces the task's whole
// assignee set idempotently — even though the model implements it as a Create.
func RegisterTaskAssigneeBulkRoutes(api huma.API) {
	tags := []string{"assignees"}

	Register(api, huma.Operation{
		OperationID: "task-assignees-bulk",
		Summary:     "Replace all assignees of a task",
		Description: "Replaces the task's full assignee set with the users in the body: users not in the list are unassigned, new ones are added. Pass an empty array to unassign everyone. Each assignee must have access to the task's project, and the caller needs write access to the task.",
		Method:      http.MethodPut,
		Path:        "/tasks/{projecttask}/assignees/bulk",
		Tags:        tags,
	}, taskAssigneesBulk)
}

func init() { AddRouteRegistrar(RegisterTaskAssigneeBulkRoutes) }

func taskAssigneesBulk(ctx context.Context, in *struct {
	TaskID int64 `path:"projecttask"`
	Body   models.BulkAssignees
}) (*singleBody[models.BulkAssignees], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.TaskID = in.TaskID // URL wins over body
	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.BulkAssignees]{Body: &in.Body}, nil
}
