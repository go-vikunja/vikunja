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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"github.com/danielgtaylor/huma/v2"
)

type adminTeamListBody struct {
	Body Paginated[models.InviteLinkTeam]
}

func init() { AddRouteRegistrar(RegisterAdminTeamRoutes) }

func RegisterAdminTeamRoutes(api huma.API) {
	Register(api, huma.Operation{OperationID: "admin-teams-list", Summary: "List local teams for invitations", Description: "Includes teams the admin does not belong to. Excludes externally managed teams. Requires admin_panel and user_invites.", Method: http.MethodGet, Path: "/admin/teams", Tags: []string{"admin"}}, adminTeamsList)
}

func adminTeamsList(_ context.Context, in *ListParams) (*adminTeamListBody, error) {
	s := db.NewSession()
	defer s.Close()
	teams, total, err := models.ListTeamsAsAdmin(s, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &adminTeamListBody{Body: NewPaginated(teams, total, in.Page, in.PerPage)}, nil
}
