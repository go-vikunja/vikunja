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
	"github.com/danielgtaylor/huma/v2/conditional"
)

// projectListBody is the list-response envelope. models.Project.ReadAll
// returns []*models.Project, so that's the element type.
type projectListBody struct {
	Body Paginated[*models.Project]
}

// RegisterProjectRoutes wires Project CRUD onto the Huma API.
func RegisterProjectRoutes(api huma.API) {
	tags := []string{"projects"}

	Register(api, huma.Operation{
		OperationID: "projects-list",
		Summary:     "List projects",
		Description: "Returns the projects the authenticated user has access to (owned plus shared, with child projects of accessible parents), paginated. Archived projects are excluded unless is_archived=true. Pass expand=permissions to include each project's max_permission for the caller.",
		Method:      http.MethodGet,
		Path:        "/projects",
		Tags:        tags,
	}, projectsList)

	Register(api, huma.Operation{
		OperationID: "projects-read",
		Summary:     "Get a project",
		Description: "Returns a single project the caller can read, including its views and the caller's favorite/subscription state. Resolves the Favorites pseudo-project and saved-filter-backed projects. Sends an ETag; pass it as If-None-Match on a later read to get a 304 Not Modified.",
		Method:      http.MethodGet,
		Path:        "/projects/{id}",
		Tags:        tags,
	}, projectsRead)

	Register(api, huma.Operation{
		OperationID: "projects-create",
		Summary:     "Create a project",
		Description: "Creates a project; the authenticated user becomes its owner. When parent_project_id is set, the caller needs write access to that parent. Default views and a backlog bucket are created automatically.",
		Method:      http.MethodPost,
		Path:        "/projects",
		Tags:        tags,
	}, projectsCreate)

	Register(api, huma.Operation{
		OperationID: "projects-update",
		Summary:     "Update a project",
		Description: "Replaces a project's fields. Requires write access (admin to reparent or delete). Use PATCH for a partial update.",
		Method:      http.MethodPut,
		Path:        "/projects/{id}",
		Tags:        tags,
	}, projectsUpdate)

	Register(api, huma.Operation{
		OperationID: "projects-delete",
		Summary:     "Delete a project",
		Description: "Deletes a project together with its tasks, views, buckets and child projects. Only project admins may delete it.",
		Method:      http.MethodDelete,
		Path:        "/projects/{id}",
		Tags:        tags,
	}, projectsDelete)
}

func init() { AddRouteRegistrar(RegisterProjectRoutes) }

func projectsList(ctx context.Context, in *struct {
	ListParams
	Expand     string `query:"expand" enum:"permissions" doc:"If set to \"permissions\", each returned project includes the max permission the requesting user has on it (max_permission). Currently only \"permissions\" is supported."`
	IsArchived bool   `query:"is_archived" doc:"If true, also returns archived projects."`
}) (*projectListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	p := &models.Project{
		Expand:     models.ProjectExpandable(in.Expand),
		IsArchived: in.IsArchived,
	}
	result, _, total, err := handler.DoReadAll(ctx, p, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*models.Project)
	if !ok {
		return nil, fmt.Errorf("projects.ReadAll returned unexpected type %T (expected []*models.Project)", result)
	}
	return &projectListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

func projectsRead(ctx context.Context, in *struct {
	ID     int64  `path:"id"`
	Expand string `query:"expand" enum:"permissions" doc:"If set to \"permissions\", the project includes the max permission the requesting user has on it (max_permission)."`
	conditional.Params
}) (*singleReadBody[models.Project], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	project := &models.Project{ID: in.ID, Expand: models.ProjectExpandable(in.Expand)}
	maxPermission, err := handler.DoReadOne(ctx, project, a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	// ReadOne doesn't act on Expand itself; the caller's max permission comes
	// from DoReadOne's CanRead result. Surface it on the model only when asked,
	// matching the list operation's expand=permissions behaviour.
	if models.ProjectExpandable(in.Expand) == models.ProjectExpandableRights {
		project.MaxPermission = models.Permission(maxPermission)
	}
	// PreconditionFailed wants the unquoted etag; response header uses RFC 9110 quoted form.
	etag := fmt.Sprintf("%d-%d", project.ID, project.Updated.UnixNano())
	if in.HasConditionalParams() {
		if err := in.PreconditionFailed(etag, project.Updated); err != nil {
			return nil, err
		}
	}
	return &singleReadBody[models.Project]{ETag: `"` + etag + `"`, Body: project}, nil
}

func projectsCreate(ctx context.Context, in *struct {
	Body models.Project
}) (*singleBody[models.Project], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.Project]{Body: &in.Body}, nil
}

func projectsUpdate(ctx context.Context, in *struct {
	ID   int64 `path:"id"`
	Body models.Project
}) (*singleBody[models.Project], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.ID = in.ID // URL wins over body
	if err := handler.DoUpdate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.Project]{Body: &in.Body}, nil
}

func projectsDelete(ctx context.Context, in *struct {
	ID int64 `path:"id"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.Project{ID: in.ID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
