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

// bucketListBody is the list-response envelope. models.Bucket.ReadAll returns
// []*models.Bucket, so that's the element type.
type bucketListBody struct {
	Body Paginated[*models.Bucket]
}

// RegisterBucketRoutes wires the nested kanban-bucket CRUD onto the Huma API.
// Buckets live under /projects/{project}/views/{view}/buckets; every operation
// binds {project} → ProjectID and {view} → ProjectViewID, the write operations
// additionally {bucket} → ID. There is intentionally no read-one route
// (mirroring v1: the Bucket model has no ReadOne/CanRead), so AutoPatch
// synthesises no PATCH either.
func RegisterBucketRoutes(api huma.API) {
	tags := []string{"projects"}

	Register(api, huma.Operation{
		OperationID: "buckets-list",
		Summary:     "List the buckets of a kanban view",
		Description: "Returns all kanban buckets of a project view, ordered by position. Requires read access to the project. The list is not paginated by the server but is returned in the standard list envelope. To get the buckets together with their tasks, use the buckets/tasks endpoint instead.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/views/{view}/buckets",
		Tags:        tags,
	}, bucketsList)

	Register(api, huma.Operation{
		OperationID: "buckets-create",
		Summary:     "Create a bucket in a kanban view",
		Description: "Creates a kanban bucket in the given project view. The project and view come from the URL, not the body. Requires write access to the project.",
		Method:      http.MethodPost,
		Path:        "/projects/{project}/views/{view}/buckets",
		Tags:        tags,
	}, bucketsCreate)

	Register(api, huma.Operation{
		OperationID: "buckets-update",
		Summary:     "Update a bucket of a kanban view",
		Description: "Replaces a kanban bucket's title, limit and position. The bucket is identified by the URL, which also scopes it to the project and view. Requires write access to the project.",
		Method:      http.MethodPut,
		Path:        "/projects/{project}/views/{view}/buckets/{bucket}",
		Tags:        tags,
	}, bucketsUpdate)

	Register(api, huma.Operation{
		OperationID: "buckets-delete",
		Summary:     "Delete a bucket of a kanban view",
		Description: "Deletes a kanban bucket and moves its tasks to the view's default bucket; no tasks are deleted. You cannot delete the last bucket of a view (rejected with 412). Requires write access to the project.",
		Method:      http.MethodDelete,
		Path:        "/projects/{project}/views/{view}/buckets/{bucket}",
		Tags:        tags,
	}, bucketsDelete)
}

func init() { AddRouteRegistrar(RegisterBucketRoutes) }

func bucketsList(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ViewID    int64 `path:"view"`
	ListParams
}) (*bucketListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.Bucket{ProjectID: in.ProjectID, ProjectViewID: in.ViewID}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	buckets, ok := result.([]*models.Bucket)
	if !ok {
		return nil, fmt.Errorf("buckets.ReadAll returned unexpected type %T (expected []*models.Bucket)", result)
	}
	return &bucketListBody{Body: NewPaginated(buckets, total, in.Page, in.PerPage)}, nil
}

func bucketsCreate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ViewID    int64 `path:"view"`
	Body      models.Bucket
}) (*singleBody[models.Bucket], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	b := &in.Body
	b.ProjectID = in.ProjectID  // URL wins over body
	b.ProjectViewID = in.ViewID // URL wins over body
	if err := handler.DoCreate(ctx, b, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.Bucket]{Body: b}, nil
}

func bucketsUpdate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ViewID    int64 `path:"view"`
	BucketID  int64 `path:"bucket"`
	Body      models.Bucket
}) (*singleBody[models.Bucket], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	b := &in.Body
	b.ID = in.BucketID          // URL wins over body
	b.ProjectID = in.ProjectID  // URL wins over body
	b.ProjectViewID = in.ViewID // URL wins over body
	if err := handler.DoUpdate(ctx, b, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.Bucket]{Body: b}, nil
}

func bucketsDelete(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ViewID    int64 `path:"view"`
	BucketID  int64 `path:"bucket"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.Bucket{ID: in.BucketID, ProjectID: in.ProjectID, ProjectViewID: in.ViewID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
