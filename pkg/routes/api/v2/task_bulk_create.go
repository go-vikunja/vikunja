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

// RegisterTaskBulkCreateRoutes wires the bulk task creation onto the Huma API.
func RegisterTaskBulkCreateRoutes(api huma.API) {
	tags := []string{"tasks"}

	Register(api, huma.Operation{
		OperationID: "tasks-bulk-create",
		Summary:     "Create multiple tasks",
		Description: "Creates up to 100 tasks in the project from the URL in one atomic request: if any task is invalid, none are created and the error names the failing index. Tasks are created in payload order and land on top of every view in that order. The authenticated user needs write access to the project and becomes the creator of every task. Bucket limits are only enforced for tasks with an explicit bucket_id; tasks left to land in a view's default bucket are not counted against its limit.",
		Method:      http.MethodPost,
		Path:        "/projects/{project}/tasks/bulk",
		Tags:        tags,
	}, tasksBulkCreate)
}

func init() { AddRouteRegistrar(RegisterTaskBulkCreateRoutes) }

func tasksBulkCreate(ctx context.Context, in *struct {
	Project int64  `path:"project" doc:"The numeric id of the project to create the tasks in."`
	Format  string `query:"format" enum:"html,markdown" doc:"How rich-text fields are exchanged. See the API description."`
	Body    models.BulkTaskCreation
}) (*singleBody[models.BulkTaskCreation], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.ProjectID = in.Project // URL wins over body

	descriptions := make([]*string, 0, len(in.Body.Tasks))
	for _, t := range in.Body.Tasks {
		if t != nil {
			descriptions = append(descriptions, &t.Description)
		}
	}
	if err := convertToHTML(ctx, descriptions...); err != nil {
		return nil, translateDomainError(err)
	}

	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}

	convertTasksToMarkdown(ctx, in.Body.Tasks...)
	return &singleBody[models.BulkTaskCreation]{Body: &in.Body}, nil
}
