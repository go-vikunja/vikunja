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

// RegisterTaskBucketRoutes wires the kanban task-bucket move onto the Huma API.
//
// TaskBucket exposes only Update, so the handler reuses handler.DoUpdate (its
// CanUpdate enforces write access on the bucket's project). The bucket and view
// come from the path; only the task id is read from the body.
func RegisterTaskBucketRoutes(api huma.API) {
	tags := []string{"projects"}

	Register(api, huma.Operation{
		OperationID: "task-bucket-update",
		Summary:     "Place a task in a kanban bucket",
		Description: "Moves a task into the given bucket of a project's kanban view. Requires write access to the project. " +
			"Idempotent: re-sending the same bucket is a no-op. Side effects: moving a task into the view's done bucket marks it done (and out of it un-marks it); a repeating task moved into the done bucket is reopened and routed back to the default bucket instead. " +
			"Moving a task into a bucket that is already at its task limit is rejected with 412. A bucket that does not resolve under the project and view in the path is rejected with 404.",
		Method: http.MethodPut,
		Path:   "/projects/{project}/views/{view}/buckets/{bucket}/tasks",
		Tags:   tags,
	}, taskBucketUpdate)
}

func init() { AddRouteRegistrar(RegisterTaskBucketRoutes) }

func taskBucketUpdate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ViewID    int64 `path:"view"`
	BucketID  int64 `path:"bucket"`
	Body      models.TaskBucket
}) (*singleBody[models.TaskBucket], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	tb := &in.Body
	tb.ProjectID = in.ProjectID  // URL wins over body
	tb.ProjectViewID = in.ViewID // URL wins over body
	tb.BucketID = in.BucketID    // URL wins over body
	if err := handler.DoUpdate(ctx, tb, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.TaskBucket]{Body: tb}, nil
}
