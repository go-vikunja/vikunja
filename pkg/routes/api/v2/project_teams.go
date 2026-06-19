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

// ReadAll returns []*models.TeamWithPermission, not the bare relation.
type projectTeamListBody struct {
	Body Paginated[*models.TeamWithPermission]
}

// RegisterProjectTeamRoutes wires the team<->project share CRUD onto Huma.
// There is no read-one operation (v1 has none either).
func RegisterProjectTeamRoutes(api huma.API) {
	tags := []string{"sharing"}

	Register(api, huma.Operation{
		OperationID: "project-teams-list",
		Summary:     "List the teams a project is shared with",
		Description: "Returns the teams that have access to the project, each with the permission they were granted. Requires read access to the project.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/teams",
		Tags:        tags,
	}, projectTeamsList)

	Register(api, huma.Operation{
		OperationID: "project-teams-create",
		Summary:     "Share a project with a team",
		Description: "Gives a team access to the project at the requested permission. Only project admins may share. Fails if the team already has access.",
		Method:      http.MethodPost,
		Path:        "/projects/{project}/teams",
		Tags:        tags,
	}, projectTeamsCreate)

	Register(api, huma.Operation{
		OperationID: "project-teams-update",
		Summary:     "Update a team's permission on a project",
		Description: "Changes the permission a team has on the project; only the permission is writable. Only project admins may update a share.",
		Method:      http.MethodPut,
		Path:        "/projects/{project}/teams/{team}",
		Tags:        tags,
	}, projectTeamsUpdate)

	Register(api, huma.Operation{
		OperationID: "project-teams-delete",
		Summary:     "Remove a team from a project",
		Description: "Revokes a team's access to the project. Only project admins may remove a share.",
		Method:      http.MethodDelete,
		Path:        "/projects/{project}/teams/{team}",
		Tags:        tags,
	}, projectTeamsDelete)
}

func init() { AddRouteRegistrar(RegisterProjectTeamRoutes) }

func projectTeamsList(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ListParams
}) (*projectTeamListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.TeamProject{ProjectID: in.ProjectID}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*models.TeamWithPermission)
	if !ok {
		return nil, fmt.Errorf("projectTeams.ReadAll returned unexpected type %T (expected []*models.TeamWithPermission)", result)
	}
	return &projectTeamListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

func projectTeamsCreate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	Body      models.TeamProject
}) (*singleBody[models.TeamProject], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.ProjectID = in.ProjectID // URL wins over body
	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.TeamProject]{Body: &in.Body}, nil
}

func projectTeamsUpdate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	TeamID    int64 `path:"team"`
	Body      models.TeamProject
}) (*singleBody[models.TeamProject], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	tp := &in.Body
	tp.ProjectID = in.ProjectID // URL wins over body
	tp.TeamID = in.TeamID
	if err := handler.DoUpdate(ctx, tp, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.TeamProject]{Body: tp}, nil
}

func projectTeamsDelete(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	TeamID    int64 `path:"team"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.TeamProject{ProjectID: in.ProjectID, TeamID: in.TeamID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
