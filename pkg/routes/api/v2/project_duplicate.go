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

// RegisterProjectDuplicateRoutes wires the project-duplicate action onto the Huma API.
//
// ProjectDuplicate is a CRUDable Create, so the handler reuses handler.DoCreate
// (its CanCreate enforces access); the only custom part is taking ProjectID from
// the path rather than the request body.
func RegisterProjectDuplicateRoutes(api huma.API) {
	tags := []string{"projects"}

	Register(api, huma.Operation{
		OperationID: "projects-duplicate",
		Summary:     "Duplicate a project",
		Description: "Deep-copies a project — its tasks, files, kanban data, assignees, comments, attachments, labels, relations and backgrounds — into a new project owned by the authenticated user. User/team/link shares are only copied when duplicate_shares is set to true. The user needs read access to the source project, plus write access to the parent project when one is given. The copy is placed under parent_project_id (top level if omitted). Returns the duplicate in duplicated_project.",
		Method:      http.MethodPost,
		Path:        "/projects/{projectid}/duplicate",
		Tags:        tags,
	}, projectsDuplicate)
}

func init() { AddRouteRegistrar(RegisterProjectDuplicateRoutes) }

func projectsDuplicate(ctx context.Context, in *struct {
	ProjectID int64 `path:"projectid" doc:"The numeric id of the project to duplicate."`
	Body      models.ProjectDuplicate
}) (*singleBody[models.ProjectDuplicate], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	pd := &in.Body
	pd.ProjectID = in.ProjectID
	if err := handler.DoCreate(ctx, pd, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.ProjectDuplicate]{Body: pd}, nil
}
