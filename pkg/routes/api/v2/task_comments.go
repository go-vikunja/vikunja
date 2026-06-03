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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/conditional"
)

type taskCommentListBody struct {
	Body Paginated[*models.TaskComment]
}

// RegisterTaskCommentRoutes wires the nested TaskComment CRUD onto the Huma API.
//
// The feature gate is checked here, not in the central wiring: the registrar
// runs at RegisterAll time after the config has loaded, so a disabled instance
// registers no comment routes at all.
func RegisterTaskCommentRoutes(api huma.API) {
	if !config.ServiceEnableTaskComments.GetBool() {
		return
	}

	tags := []string{"task_comments"}

	Register(api, huma.Operation{
		OperationID: "task-comments-list",
		Summary:     "List the comments of a task",
		Description: "Returns the comments of the given task, paginated. Requires read access to the task. Pass order_by=desc to sort newest-first (default is oldest-first).",
		Method:      http.MethodGet,
		Path:        "/tasks/{task}/comments",
		Tags:        tags,
	}, taskCommentsList)

	Register(api, huma.Operation{
		OperationID: "task-comments-read",
		Summary:     "Get a single comment of a task",
		Description: "Returns one comment of a task. The comment must belong to the task in the path. Sends an ETag; pass it as If-None-Match on a later read to get a 304 Not Modified.",
		Method:      http.MethodGet,
		Path:        "/tasks/{task}/comments/{commentid}",
		Tags:        tags,
	}, taskCommentsRead)

	Register(api, huma.Operation{
		OperationID: "task-comments-create",
		Summary:     "Create a comment on a task",
		Description: "Adds a comment to the given task. The parent task is taken from the URL, not the body, and the author is the authenticated user. Requires write access to the task.",
		Method:      http.MethodPost,
		Path:        "/tasks/{task}/comments",
		Tags:        tags,
	}, taskCommentsCreate)

	Register(api, huma.Operation{
		OperationID: "task-comments-update",
		Summary:     "Update a comment of a task",
		Description: "Replaces a comment's text. The comment must belong to the task in the path, and only its author may update it. Use PATCH for a partial update.",
		Method:      http.MethodPut,
		Path:        "/tasks/{task}/comments/{commentid}",
		Tags:        tags,
	}, taskCommentsUpdate)

	Register(api, huma.Operation{
		OperationID: "task-comments-delete",
		Summary:     "Delete a comment of a task",
		Description: "Deletes a comment of a task. The comment must belong to the task in the path, and only its author may delete it.",
		Method:      http.MethodDelete,
		Path:        "/tasks/{task}/comments/{commentid}",
		Tags:        tags,
	}, taskCommentsDelete)
}

func init() { AddRouteRegistrar(RegisterTaskCommentRoutes) }

func taskCommentsList(ctx context.Context, in *struct {
	TaskID  int64  `path:"task"`
	OrderBy string `query:"order_by" enum:"asc,desc" default:"asc" doc:"Sort order by creation time: 'asc' (oldest first, default) or 'desc' (newest first)."`
	ListParams
}) (*taskCommentListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.TaskComment{TaskID: in.TaskID, OrderBy: in.OrderBy}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*models.TaskComment)
	if !ok {
		return nil, fmt.Errorf("taskComments.ReadAll returned unexpected type %T (expected []*models.TaskComment)", result)
	}
	return &taskCommentListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

func taskCommentsRead(ctx context.Context, in *struct {
	TaskID int64 `path:"task"`
	ID     int64 `path:"commentid"`
	conditional.Params
}) (*singleReadBody[models.TaskComment], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	// TaskID scopes the lookup to the parent task, guarding against reading a
	// comment of one task through another (IDOR).
	comment := &models.TaskComment{ID: in.ID, TaskID: in.TaskID}
	if _, err := handler.DoReadOne(ctx, comment, a); err != nil {
		return nil, translateDomainError(err)
	}
	// PreconditionFailed wants the unquoted etag; the response header uses the RFC 9110 quoted form.
	etag := fmt.Sprintf("%d-%d", comment.ID, comment.Updated.UnixNano())
	if in.HasConditionalParams() {
		if err := in.PreconditionFailed(etag, comment.Updated); err != nil {
			return nil, err
		}
	}
	return &singleReadBody[models.TaskComment]{ETag: `"` + etag + `"`, Body: comment}, nil
}

func taskCommentsCreate(ctx context.Context, in *struct {
	TaskID int64 `path:"task"`
	Body   models.TaskComment
}) (*singleBody[models.TaskComment], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.TaskID = in.TaskID // URL wins over body
	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.TaskComment]{Body: &in.Body}, nil
}

func taskCommentsUpdate(ctx context.Context, in *struct {
	TaskID int64 `path:"task"`
	ID     int64 `path:"commentid"`
	Body   models.TaskComment
}) (*singleBody[models.TaskComment], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.ID = in.ID         // URL wins over body
	in.Body.TaskID = in.TaskID // parent from the path scopes the update
	if err := handler.DoUpdate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.TaskComment]{Body: &in.Body}, nil
}

func taskCommentsDelete(ctx context.Context, in *struct {
	TaskID int64 `path:"task"`
	ID     int64 `path:"commentid"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.TaskComment{ID: in.ID, TaskID: in.TaskID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
