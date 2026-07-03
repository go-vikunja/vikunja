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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
)

type adminProjectListBody struct {
	Body Paginated[*models.Project]
}

type adminProjectBody struct {
	Body *models.Project
}

// adminOwnerPatchBody reassigns a project's owner. owner_id is the only field;
// the regular project-update endpoint refuses owner changes.
type adminOwnerPatchBody struct {
	OwnerID int64 `json:"owner_id" minimum:"1" doc:"The numeric ID of the user who should become the project's owner."`
}

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

	Register(api, huma.Operation{
		OperationID: "admin-projects-patch-owner",
		Summary:     "Reassign a project's owner (admin)",
		Description: "Reassigns a project to a new owner — the admin-only escape hatch the regular update endpoint does not allow. The new owner must be an active account that is not scheduled for deletion. Restricted to instance admins on a licensed instance.",
		Method:      http.MethodPatch,
		Path:        "/admin/projects/{id}/owner",
		Tags:        tags,
	}, adminProjectsPatchOwner)
}

func init() { AddRouteRegistrar(RegisterAdminProjectRoutes) }

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

func adminProjectsPatchOwner(ctx context.Context, in *struct {
	ID   int64 `path:"id" doc:"The numeric ID of the project."`
	Body adminOwnerPatchBody
}) (*adminProjectBody, error) {
	if in.ID < 1 {
		return nil, translateDomainError(models.ErrProjectDoesNotExist{ID: in.ID})
	}
	if in.Body.OwnerID < 1 {
		return nil, translateDomainError(models.ErrInvalidData{Message: "invalid body"})
	}

	doer, err := adminDoerFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	p, err := models.ReassignProjectOwner(s, doer, in.ID, in.Body.OwnerID)
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		events.CleanupPending(s)
		return nil, translateDomainError(err)
	}
	events.DispatchPending(ctx, s)
	return &adminProjectBody{Body: p}, nil
}
