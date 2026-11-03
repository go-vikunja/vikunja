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

type teamProjectListBody struct {
	Body Paginated[*models.Project]
}

func RegisterTeamProjectRoutes(api huma.API) {
	Register(api, huma.Operation{
		OperationID: "teams-projects-list",
		Summary:     "List projects shared with a team that you administer",
		Description: "Returns the projects shared directly with the team on which the caller is admin. Projects that only inherit the share through a parent are not included. Requires team membership. Not paginated: all matching projects are returned on one page.",
		Method:      http.MethodGet,
		Path:        "/teams/{team}/projects",
		Tags:        []string{"teams"},
	}, teamProjectsList)
}

func init() { AddRouteRegistrar(RegisterTeamProjectRoutes) }

func teamProjectsList(ctx context.Context, in *struct {
	TeamID int64 `path:"team"`
}) (*teamProjectListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	s := db.NewSession()
	defer s.Close()

	projects, err := models.GetAdminProjectsSharedWithTeam(s, a, in.TeamID)
	if err != nil {
		return nil, translateDomainError(err)
	}
	total := len(projects)
	return &teamProjectListBody{Body: NewPaginated(projects, int64(total), 1, max(total, 1))}, nil
}
