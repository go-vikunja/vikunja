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

type projectUrgencyListBody struct {
	Body Paginated[models.ProjectUrgencyWeight]
}

type projectUrgencyUpdateBody struct {
	models.ProjectUrgencyWeights
}

func registerProjectUrgencyRoutes(api huma.API) {
	tags := []string{"project_urgency"}

	Register(api, huma.Operation{
		OperationID: "project-urgency-list",
		Summary:     "List the urgency weights of a project",
		Description: "Returns all urgency weights of the given project. Requires read access to the project; the list is not paginated byt he server but is returned in the standard list envelope.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/urgency_weights",
		Tags:        tags,
	}, projectUrgencyList)

	Register(api, huma.Operation{
		OperationID: "project-urgency-update",
		Summary:     "Update all urgency weights of a project",
		Description: "Replaces a project's urgency weights.",
		Method:      http.MethodPut,
		Path:        "/projects/{project}/urgency_weights",
		Tags:        tags,
	}, projectUrgencyUpdate)
}

func init() { AddRouteRegistrar(registerProjectUrgencyRoutes) }

type ProjectUrgencyWeight struct {
	Property models.UrgencyProperty `json:"property"`
	Weight   float64                `json:"weight"`
	Filter   *models.BasicFilter    `json:"filter,omitempty"`
}

func projectUrgencyList(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ListParams
},
) (*projectUrgencyListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.ProjectUrgencyWeights{ProjectID: in.ProjectID}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.(models.ProjectUrgencyWeights)
	if !ok {
		return nil, fmt.Errorf("projectUrgencyWeight.ReadAll returned unexpected type %T (expected []*models.ProjectUrgencyWeight)", result)
	}
	return &projectUrgencyListBody{Body: NewPaginated(items.UrgencyWeights, total, in.Page, in.PerPage)}, nil
}

func projectUrgencyUpdate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	Body      projectUrgencyUpdateBody
},
) (*singleBody[models.ProjectUrgencyWeights], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	urgencyWeights := &in.Body.ProjectUrgencyWeights
	urgencyWeights.ProjectID = in.ProjectID // parent from the path scopes the update
	if err := handler.DoUpdate(ctx, urgencyWeights, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.ProjectUrgencyWeights]{Body: urgencyWeights}, nil
}
