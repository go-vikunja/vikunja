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

// deletedProjectsBody is not paginated: the bin only ever holds a user's own
// projects within the retention window.
type deletedProjectsBody struct {
	Body []*models.Project
}

func RegisterProjectRestoreRoutes(api huma.API) {
	tags := []string{"projects"}

	Register(api, huma.Operation{
		OperationID: "projects-deleted-list",
		Summary:     "List deleted projects",
		Description: "Returns the caller's soft-deleted projects together with the deleted_at timestamp they were removed at. Projects are permanently purged 30 days after that timestamp. Only projects the caller owns are listed — a project shared with the caller stays with its owner's bin.",
		Method:      http.MethodGet,
		Path:        "/projects/deleted",
		Tags:        tags,
	}, projectsDeletedList)

	Register(api, huma.Operation{
		OperationID: "projects-restore",
		Summary:     "Restore a deleted project",
		Description: "Restores a soft-deleted project and every descendant that was soft-deleted along with it. Requires admin permission on the project. Fails with 404 once the project has been permanently purged.",
		Method:      http.MethodPost,
		Path:        "/projects/{id}/restore",
		Tags:        tags,
		// 200, not the POST default 201: this revives an existing project.
		DefaultStatus: http.StatusOK,
	}, projectsRestore)
}

func init() { AddRouteRegistrar(RegisterProjectRestoreRoutes) }

func projectsDeletedList(ctx context.Context, _ *struct{}) (*deletedProjectsBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	// GetDeletedProjects scopes to the caller's own projects, so there is no
	// separate Can* to call here.
	projects, err := models.GetDeletedProjects(s, a)
	if err != nil {
		return nil, translateDomainError(err)
	}

	for _, p := range projects {
		convertToMarkdown(ctx, &p.Description)
	}

	return &deletedProjectsBody{Body: projects}, nil
}

func projectsRestore(ctx context.Context, in *struct {
	ID int64 `path:"id"`
}) (*singleBody[models.Project], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	// RestoreProject checks admin permission itself — the project is invisible
	// to the regular Can* path while it is soft-deleted.
	project, err := models.RestoreProject(s, in.ID, a)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}

	convertToMarkdown(ctx, &project.Description)
	project.MaxPermission = models.PermissionUnknown

	return &singleBody[models.Project]{Body: project}, nil
}
