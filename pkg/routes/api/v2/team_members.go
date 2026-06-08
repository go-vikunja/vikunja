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

// RegisterTeamMemberRoutes wires team membership management onto the Huma API.
//
// Members are read through the team itself (Team.Members), so there is no
// list/read here — only add, remove and the admin toggle.
func RegisterTeamMemberRoutes(api huma.API) {
	tags := []string{"teams"}

	Register(api, huma.Operation{
		OperationID: "teams-members-add",
		Summary:     "Add a member to a team",
		Description: "Adds a user to a team by username. Only a team admin may add members.",
		Method:      http.MethodPost,
		Path:        "/teams/{team}/members",
		Tags:        tags,
	}, teamMembersAdd)

	Register(api, huma.Operation{
		OperationID: "teams-members-remove",
		Summary:     "Remove a member from a team",
		Description: "Removes a user from a team, revoking the access the team granted them. A team admin may remove anyone; a member may remove themselves. The last member of a team cannot be removed.",
		Method:      http.MethodDelete,
		Path:        "/teams/{team}/members/{user}",
		Tags:        tags,
	}, teamMembersRemove)

	Register(api, huma.Operation{
		OperationID:   "teams-members-toggle-admin",
		Summary:       "Toggle a team member's admin status",
		Description:   "Flips the member's admin flag: an admin becomes a regular member and vice-versa. The request body is ignored. Only a team admin may do this.",
		Method:        http.MethodPost,
		Path:          "/teams/{team}/members/{user}/admin",
		DefaultStatus: http.StatusOK,
		Tags:          tags,
	}, teamMembersToggleAdmin)
}

func init() { AddRouteRegistrar(RegisterTeamMemberRoutes) }

func teamMembersAdd(ctx context.Context, in *struct {
	TeamID int64 `path:"team"`
	Body   models.TeamMember
}) (*singleBody[models.TeamMember], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.TeamID = in.TeamID // URL wins over body
	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.TeamMember]{Body: &in.Body}, nil
}

func teamMembersRemove(ctx context.Context, in *struct {
	TeamID   int64  `path:"team"`
	Username string `path:"user" doc:"The username of the member to remove."`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	tm := &models.TeamMember{TeamID: in.TeamID, Username: in.Username}
	if err := handler.DoDelete(ctx, tm, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

// teamMembersToggleAdmin shares v1's update pipeline (DoUpdate, the core UpdateWeb wraps).
func teamMembersToggleAdmin(ctx context.Context, in *struct {
	TeamID   int64  `path:"team"`
	Username string `path:"user" doc:"The username of the member whose admin status to toggle."`
}) (*singleBody[models.TeamMember], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	tm := &models.TeamMember{TeamID: in.TeamID, Username: in.Username}
	if err := handler.DoUpdate(ctx, tm, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.TeamMember]{Body: tm}, nil
}
