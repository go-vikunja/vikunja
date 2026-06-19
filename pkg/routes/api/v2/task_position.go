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

// RegisterTaskPositionRoutes wires the task-position update onto the Huma API.
//
// Setting a position is a plain CRUDable Update, so the handler reuses
// handler.DoUpdate (its CanUpdate delegates to the task's CanUpdate); the only
// custom part is taking TaskID from the path rather than the request body.
func RegisterTaskPositionRoutes(api huma.API) {
	tags := []string{"tasks"}

	Register(api, huma.Operation{
		OperationID: "tasks-position-update",
		Summary:     "Set a task's position in a view",
		Description: "Sets where a task sorts within one of its project's views. The position is per view, so this only affects the view named by project_view_id. Requires write access to the task. Positions below the minimum spacing make the server recalculate every position in the view, so the returned value may differ from the one sent.",
		Method:      http.MethodPut,
		Path:        "/tasks/{task}/position",
		Tags:        tags,
	}, tasksPositionUpdate)
}

func init() { AddRouteRegistrar(RegisterTaskPositionRoutes) }

func tasksPositionUpdate(ctx context.Context, in *struct {
	TaskID int64 `path:"task" doc:"The numeric id of the task whose position to set."`
	Body   models.TaskPosition
}) (*singleBody[models.TaskPosition], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	tp := &in.Body
	tp.TaskID = in.TaskID // URL wins over body
	if err := handler.DoUpdate(ctx, tp, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.TaskPosition]{Body: tp}, nil
}
