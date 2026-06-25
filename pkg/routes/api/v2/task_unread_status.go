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

// taskReadBody confirms the mark-read action: the underlying model carries no
// JSON-exposed fields, so it returns a status message rather than a resource.
type taskReadBody struct {
	Body struct {
		Message string `json:"message" readOnly:"true" doc:"A confirmation message."`
	}
}

// RegisterTaskUnreadStatusRoutes wires the mark-task-as-read action onto the Huma API.
//
// Marking a task read clears the caller's unread entry for it, which is what
// drives the per-task "unread" dot shown for mentions and other notifications.
// The model's Update deletes that entry, so the action is idempotent — PUT, not
// POST. It is also unconditional: there is no read entry to clear for a task the
// caller cannot see, so it succeeds as a no-op rather than refusing.
func RegisterTaskUnreadStatusRoutes(api huma.API) {
	tags := []string{"tasks"}

	Register(api, huma.Operation{
		OperationID: "tasks-mark-read",
		Summary:     "Mark a task as read",
		Description: "Clears the authenticated user's unread status for a task, dismissing the unread indicator raised by mentions and other task notifications. Idempotent: marking an already-read or inaccessible task succeeds as a no-op.",
		Method:      http.MethodPut,
		Path:        "/tasks/{projecttask}/read",
		Tags:        tags,
	}, tasksMarkRead)
}

func init() { AddRouteRegistrar(RegisterTaskUnreadStatusRoutes) }

func tasksMarkRead(ctx context.Context, in *struct {
	TaskID int64 `path:"projecttask" doc:"The numeric id of the task to mark as read."`
}) (*taskReadBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	t := &models.TaskUnreadStatus{TaskID: in.TaskID}
	if err := handler.DoUpdate(ctx, t, a); err != nil {
		return nil, translateDomainError(err)
	}
	out := &taskReadBody{}
	out.Body.Message = "success"
	return out, nil
}
