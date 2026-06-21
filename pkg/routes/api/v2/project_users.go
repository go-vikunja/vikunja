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

// ReadAll yields the shared users, not the join rows: []*models.UserWithPermission.
type projectUserListBody struct {
	Body Paginated[*models.UserWithPermission]
}

// RegisterProjectUserRoutes registers the project<->user share routes. {user} is
// the username (a string), not a numeric id; there is no read-one.
func RegisterProjectUserRoutes(api huma.API) {
	tags := []string{"sharing"}

	Register(api, huma.Operation{
		OperationID: "project-users-list",
		Summary:     "List the users a project is shared with",
		Description: "Returns the users that have direct access to the project, with their permission. Requires read access to the project; team shares are not included. Pass q to filter by username.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/users",
		Tags:        tags,
	}, projectUsersList)

	Register(api, huma.Operation{
		OperationID: "project-users-create",
		Summary:     "Share a project with a user",
		Description: "Grants a user access to the project. The user is named by username in the body. Only project admins may share; the project owner cannot be added.",
		Method:      http.MethodPost,
		Path:        "/projects/{project}/users",
		Tags:        tags,
	}, projectUsersCreate)

	Register(api, huma.Operation{
		OperationID: "project-users-update",
		Summary:     "Update a user's permission on a project",
		Description: "Changes the permission a user has on the project; only the permission field is updated. The user is identified by username in the path. Only project admins may update a share.",
		Method:      http.MethodPut,
		Path:        "/projects/{project}/users/{user}",
		Tags:        tags,
	}, projectUsersUpdate)

	Register(api, huma.Operation{
		OperationID: "project-users-delete",
		Summary:     "Remove a user's access to a project",
		Description: "Revokes a user's direct access to the project, identified by username in the path. Only project admins may do this.",
		Method:      http.MethodDelete,
		Path:        "/projects/{project}/users/{user}",
		Tags:        tags,
	}, projectUsersDelete)
}

func init() { AddRouteRegistrar(RegisterProjectUserRoutes) }

func projectUsersList(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	ListParams
}) (*projectUserListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.ProjectUser{ProjectID: in.ProjectID}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*models.UserWithPermission)
	if !ok {
		return nil, fmt.Errorf("projectUsers.ReadAll returned unexpected type %T (expected []*models.UserWithPermission)", result)
	}
	return &projectUserListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

func projectUsersCreate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project"`
	Body      models.ProjectUser
}) (*singleBody[models.ProjectUser], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.ProjectID = in.ProjectID // URL wins over body
	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.ProjectUser]{Body: &in.Body}, nil
}

func projectUsersUpdate(ctx context.Context, in *struct {
	ProjectID int64  `path:"project"`
	Username  string `path:"user"`
	Body      models.ProjectUser
}) (*singleBody[models.ProjectUser], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	// Update only persists permission; the user and project come from the path.
	lu := &in.Body
	lu.ProjectID = in.ProjectID
	lu.Username = in.Username
	if err := handler.DoUpdate(ctx, lu, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.ProjectUser]{Body: lu}, nil
}

func projectUsersDelete(ctx context.Context, in *struct {
	ProjectID int64  `path:"project"`
	Username  string `path:"user"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.ProjectUser{ProjectID: in.ProjectID, Username: in.Username}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
