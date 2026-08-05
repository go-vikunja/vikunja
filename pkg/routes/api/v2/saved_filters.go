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
	"github.com/danielgtaylor/huma/v2/conditional"
)

// RegisterSavedFilterRoutes wires saved filter CRUD onto the Huma API.
// No list operation, by design — v1 has none either.
func RegisterSavedFilterRoutes(api huma.API) {
	tags := []string{"filters"}

	Register(api, huma.Operation{
		OperationID: "filters-read",
		Summary:     "Get a saved filter",
		Description: "Returns a single saved filter. Only the owner may see it. Sends an ETag; pass it as If-None-Match on a later read to get a 304 Not Modified.",
		Method:      http.MethodGet,
		Path:        "/filters/{filter}",
		Tags:        tags,
	}, savedFiltersRead)

	Register(api, huma.Operation{
		OperationID: "filters-create",
		Summary:     "Create a saved filter",
		Description: "Creates a saved filter; the authenticated user becomes its owner. The filter query is validated before it is stored.",
		Method:      http.MethodPost,
		Path:        "/filters",
		Tags:        tags,
	}, savedFiltersCreate)

	Register(api, huma.Operation{
		OperationID: "filters-update",
		Summary:     "Update a saved filter",
		Description: "Replaces all of a saved filter's fields — only the owner may update it. Use PATCH for a partial update.",
		Method:      http.MethodPut,
		Path:        "/filters/{filter}",
		Tags:        tags,
	}, savedFiltersUpdate)

	Register(api, huma.Operation{
		OperationID: "filters-delete",
		Summary:     "Delete a saved filter",
		Description: "Deletes a saved filter. Only the owner may delete it.",
		Method:      http.MethodDelete,
		Path:        "/filters/{filter}",
		Tags:        tags,
	}, savedFiltersDelete)
}

func init() { AddRouteRegistrar(RegisterSavedFilterRoutes) }

type savedFilterReadBody struct {
	models.SavedFilter
	MaxPermission models.Permission `json:"max_permission" readOnly:"true" doc:"The maximum permission the requesting user has on this saved filter (0=read, 1=read/write, 2=admin). Filters are owner-only, so this is always 2 for a successful read."`
}

func savedFiltersRead(ctx context.Context, in *struct {
	ID     int64  `path:"filter"`
	Format string `query:"format" enum:"html,markdown" doc:"How rich-text fields are exchanged. See the API description."`
	conditional.Params
}) (*singleReadBody[savedFilterReadBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	filter := &models.SavedFilter{ID: in.ID}
	maxPermission, err := handler.DoReadOne(ctx, filter, a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	body := &savedFilterReadBody{SavedFilter: *filter, MaxPermission: models.Permission(maxPermission)}
	convertToMarkdown(ctx, &body.Description)
	return conditionalReadResponse(&in.Params, body, filter.Updated, maxPermission)
}

func savedFiltersCreate(ctx context.Context, in *struct {
	Format string `query:"format" enum:"html,markdown" doc:"How rich-text fields are exchanged. See the API description."`
	Body   models.SavedFilter
}) (*singleBody[models.SavedFilter], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := convertToHTML(ctx, &in.Body.Description); err != nil {
		return nil, translateDomainError(err)
	}
	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	convertToMarkdown(ctx, &in.Body.Description)
	return &singleBody[models.SavedFilter]{Body: &in.Body}, nil
}

// Body matches the read shape so AutoPatch's GET→PUT echo of max_permission validates.
func savedFiltersUpdate(ctx context.Context, in *struct {
	ID     int64  `path:"filter"`
	Format string `query:"format" enum:"html,markdown" doc:"How rich-text fields are exchanged. See the API description."`
	Body   savedFilterReadBody
}) (*singleBody[models.SavedFilter], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	filter := &in.Body.SavedFilter
	filter.ID = in.ID // URL wins over body
	if err := convertToHTML(ctx, &filter.Description); err != nil {
		return nil, translateDomainError(err)
	}
	if err := handler.DoUpdate(ctx, filter, a); err != nil {
		return nil, translateDomainError(err)
	}
	convertToMarkdown(ctx, &filter.Description)
	return &singleBody[models.SavedFilter]{Body: filter}, nil
}

func savedFiltersDelete(ctx context.Context, in *struct {
	ID int64 `path:"filter"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.SavedFilter{ID: in.ID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
