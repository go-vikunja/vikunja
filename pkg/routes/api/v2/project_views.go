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

// projectViewListBody is the list-response envelope. models.ProjectView.ReadAll
// returns []*models.ProjectView, so that's the element type.
type projectViewListBody struct {
	Body Paginated[*models.ProjectView]
}

// RegisterProjectViewRoutes wires the nested ProjectView CRUD onto the Huma API.
// Every operation binds two path params: {project} → ProjectID and {view} → ID.
// This is the reference shape every nested sub-resource copies.
func RegisterProjectViewRoutes(api huma.API) {
	tags := []string{"project_views"}

	Register(api, huma.Operation{
		OperationID: "project-views-list",
		Summary:     "List the views of a project",
		Description: "Returns all views of the given project. Requires read access to the project; the list is not paginated by the server but is returned in the standard list envelope.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/views",
		Tags:        tags,
	}, projectViewsList)

	Register(api, huma.Operation{
		OperationID: "project-views-read",
		Summary:     "Get a single view of a project",
		Description: "Returns one view of a project. The view must belong to the project in the path. Sends an ETag; pass it as If-None-Match on a later read to get a 304 Not Modified.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/views/{view}",
		Tags:        tags,
	}, projectViewsRead)

	Register(api, huma.Operation{
		OperationID: "project-views-create",
		Summary:     "Create a view in a project",
		Description: "Creates a view in the given project. The parent project is taken from the URL, not the body. Only project admins may create a view.",
		Method:      http.MethodPost,
		Path:        "/projects/{project}/views",
		Tags:        tags,
	}, projectViewsCreate)

	Register(api, huma.Operation{
		OperationID: "project-views-update",
		Summary:     "Update a view of a project",
		Description: "Replaces a project view's fields. The view must belong to the project in the path, and only project admins may update it. Use PATCH for a partial update.",
		Method:      http.MethodPut,
		Path:        "/projects/{project}/views/{view}",
		Tags:        tags,
	}, projectViewsUpdate)

	Register(api, huma.Operation{
		OperationID: "project-views-delete",
		Summary:     "Delete a view of a project",
		Description: "Deletes a project view along with its buckets and task positions. Only project admins may delete it.",
		Method:      http.MethodDelete,
		Path:        "/projects/{project}/views/{view}",
		Tags:        tags,
	}, projectViewsDelete)
}

func init() { AddRouteRegistrar(RegisterProjectViewRoutes) }

func projectViewsList(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ListParams
}) (*projectViewListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.ProjectView{ProjectID: in.ProjectID}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*models.ProjectView)
	if !ok {
		return nil, fmt.Errorf("projectViews.ReadAll returned unexpected type %T (expected []*models.ProjectView)", result)
	}
	return &projectViewListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

type projectViewReadBody struct {
	models.ProjectView
	MaxPermission models.Permission `json:"max_permission" readOnly:"true" doc:"The maximum permission the requesting user has on this view (0=read, 1=read/write, 2=admin)."`
}

func projectViewsRead(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ID        int64 `path:"view"`
	conditional.Params
}) (*singleReadBody[projectViewReadBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	// ReadOne resolves the view via GetProjectViewByIDAndProject, which needs
	// both ids — the parent project scopes the lookup.
	view := &models.ProjectView{ID: in.ID, ProjectID: in.ProjectID}
	maxPermission, err := handler.DoReadOne(ctx, view, a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	body := &projectViewReadBody{ProjectView: *view, MaxPermission: models.Permission(maxPermission)}
	return conditionalReadResponse(&in.Params, body, view.Updated, maxPermission)
}

func projectViewsCreate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	Body      models.ProjectView
}) (*singleBody[models.ProjectView], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.ProjectID = in.ProjectID // URL wins over body
	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.ProjectView]{Body: &in.Body}, nil
}

// Body matches the read shape so AutoPatch's GET→PUT echo of max_permission validates.
func projectViewsUpdate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ID        int64 `path:"view"`
	Body      projectViewReadBody
}) (*singleBody[models.ProjectView], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	view := &in.Body.ProjectView
	view.ID = in.ID               // URL wins over body
	view.ProjectID = in.ProjectID // parent from the path scopes the update
	if err := handler.DoUpdate(ctx, view, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.ProjectView]{Body: view}, nil
}

func projectViewsDelete(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ID        int64 `path:"view"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.ProjectView{ID: in.ID, ProjectID: in.ProjectID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
