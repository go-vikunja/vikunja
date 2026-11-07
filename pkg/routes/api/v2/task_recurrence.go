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

func RegisterTaskRecurrenceRoutes(api huma.API) {
	Register(api, huma.Operation{
		OperationID: "tasks-recurrence",
		Summary:     "Complete or undo a recurring task occurrence",
		Description: "Requires task write access. Completing returns a short-lived undo token. Send that token to restore the previous occurrence if the task has not changed.",
		Method:      http.MethodPost,
		Path:        "/tasks/{task}/recurrence",
		Tags:        []string{"tasks"},
		Errors:      []int{http.StatusConflict},
	}, tasksRecurrence)
}

func init() { AddRouteRegistrar(RegisterTaskRecurrenceRoutes) }

func tasksRecurrence(ctx context.Context, in *struct {
	TaskID int64 `path:"task" doc:"The numeric ID of the recurring task."`
	Body   struct {
		UndoToken string `json:"undo_token,omitempty" doc:"The token returned by completion. Omit to complete the occurrence."`
	}
}) (*singleBody[models.TaskRecurrence], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	recurrence := &models.TaskRecurrence{TaskID: in.TaskID, UndoToken: in.Body.UndoToken}
	if err := handler.DoCreate(ctx, recurrence, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.TaskRecurrence]{Body: recurrence}, nil
}
