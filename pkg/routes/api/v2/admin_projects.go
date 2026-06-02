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

type adminProjectListBody struct {
	Body Paginated[*models.Project]
}

func init() { AddRouteRegistrar(RegisterAdminProjectRoutes) }

// Permissions are enforced by the gateV2AdminRoutes path middleware, not per-handler.
func RegisterAdminProjectRoutes(api huma.API) {
	tags := []string{"admin"}

	Register(api, huma.Operation{
		OperationID: "admin-projects-list",
		Summary:     "List all projects (admin)",
		Description: "Returns every project on the instance, including archived ones and projects the caller does not own. Restricted to instance admins on a licensed instance; unlicensed or non-admin callers get a 404, making the endpoint indistinguishable from one that is not registered.",
		Method:      http.MethodGet,
		Path:        "/admin/projects",
		Tags:        tags,
	}, adminProjectsList)
}

func adminProjectsList(ctx context.Context, in *ListParams) (*adminProjectListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.AdminProjectList{}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*models.Project)
	if !ok {
		return nil, fmt.Errorf("AdminProjectList.ReadAll returned unexpected type %T (expected []*models.Project)", result)
	}
	return &adminProjectListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}
