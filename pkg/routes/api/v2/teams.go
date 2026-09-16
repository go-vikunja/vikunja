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

// models.Team.ReadAll returns []*models.Team, so that's the element type.
type teamListBody struct {
	Body Paginated[*models.Team]
}

// RegisterTeamRoutes wires Team CRUD onto the Huma API.
func RegisterTeamRoutes(api huma.API) {
	tags := []string{"teams"}

	Register(api, huma.Operation{
		OperationID: "teams-list",
		Summary:     "List teams",
		Description: "Returns the teams the authenticated user is a member of, paginated. Set include_public=true to also surface public teams the user is not a member of, where the instance has public teams enabled.",
		Method:      http.MethodGet,
		Path:        "/teams",
		Tags:        tags,
	}, teamsList)

	Register(api, huma.Operation{
		OperationID: "teams-read",
		Summary:     "Get a team",
		Description: "Returns a single team the user is a member of. Sends an ETag; pass it as If-None-Match on a later read to get a 304 Not Modified.",
		Method:      http.MethodGet,
		Path:        "/teams/{id}",
		Tags:        tags,
	}, teamsRead)

	Register(api, huma.Operation{
		OperationID: "teams-create",
		Summary:     "Create a team",
		Description: "Creates a team; the authenticated user becomes its first member and an admin of it.",
		Method:      http.MethodPost,
		Path:        "/teams",
		Tags:        tags,
	}, teamsCreate)

	Register(api, huma.Operation{
		OperationID: "teams-update",
		Summary:     "Update a team",
		Description: "Replaces a team's fields — only a team admin may update it. Use PATCH for a partial update.",
		Method:      http.MethodPut,
		Path:        "/teams/{id}",
		Tags:        tags,
	}, teamsUpdate)

	Register(api, huma.Operation{
		OperationID: "teams-delete",
		Summary:     "Delete a team",
		Description: "Deletes a team and revokes the access it granted to all of its members. Only a team admin may delete it.",
		Method:      http.MethodDelete,
		Path:        "/teams/{id}",
		Tags:        tags,
	}, teamsDelete)
}

func init() { AddRouteRegistrar(RegisterTeamRoutes) }

func teamsList(ctx context.Context, in *struct {
	ListParams
	// IncludePublic mirrors the model's include_public query param; bound
	// onto the model below so ReadAll can honor it (gated by the instance
	// public-teams setting).
	IncludePublic bool   `query:"include_public" doc:"Also include public teams the user is not a member of. Only honored when public teams are enabled on the instance."`
	Format        string `query:"format" enum:"html,markdown" doc:"How rich-text fields are exchanged. See the API description."`
}) (*teamListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.Team{IncludePublic: in.IncludePublic}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*models.Team)
	if !ok {
		return nil, fmt.Errorf("teams.ReadAll returned unexpected type %T (expected []*models.Team)", result)
	}
	for _, team := range items {
		convertToMarkdown(ctx, &team.Description)
	}
	return &teamListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

type teamReadBody struct {
	models.Team
	MaxPermission models.Permission `json:"max_permission" readOnly:"true" doc:"The maximum permission the requesting user has on this team (0=read, 2=admin). Teams have no write tier."`
}

func teamsRead(ctx context.Context, in *struct {
	ID     int64  `path:"id"`
	Format string `query:"format" enum:"html,markdown" doc:"How rich-text fields are exchanged. See the API description."`
	conditional.Params
}) (*singleReadBody[teamReadBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	team := &models.Team{ID: in.ID}
	maxPermission, err := handler.DoReadOne(ctx, team, a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	body := &teamReadBody{Team: *team, MaxPermission: models.Permission(maxPermission)}
	convertToMarkdown(ctx, &body.Description)
	return conditionalReadResponse(&in.Params, body, team.Updated, maxPermission)
}

func teamsCreate(ctx context.Context, in *struct {
	Format string `query:"format" enum:"html,markdown" doc:"How rich-text fields are exchanged. See the API description."`
	Body   models.Team
}) (*singleBody[models.Team], error) {
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
	return &singleBody[models.Team]{Body: &in.Body}, nil
}

// Body matches the read shape so AutoPatch's GET→PUT echo of max_permission validates.
func teamsUpdate(ctx context.Context, in *struct {
	ID     int64  `path:"id"`
	Format string `query:"format" enum:"html,markdown" doc:"How rich-text fields are exchanged. See the API description."`
	Body   teamReadBody
}) (*singleBody[models.Team], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	team := &in.Body.Team
	team.ID = in.ID // URL wins over body
	if err := convertToHTML(ctx, &team.Description); err != nil {
		return nil, translateDomainError(err)
	}
	if err := handler.DoUpdate(ctx, team, a); err != nil {
		return nil, translateDomainError(err)
	}
	convertToMarkdown(ctx, &team.Description)
	return &singleBody[models.Team]{Body: team}, nil
}

func teamsDelete(ctx context.Context, in *struct {
	ID int64 `path:"id"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.Team{ID: in.ID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
