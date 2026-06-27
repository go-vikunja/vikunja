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

// RegisterBulkTaskRoutes wires the bulk task update action onto the Huma API.
//
// BulkTask is a CRUDable Update, so the handler reuses handler.DoUpdate; its
// CanUpdate fans the write check out across every project the involved tasks
// belong to, so a single project the user can't write to rejects the request.
func RegisterBulkTaskRoutes(api huma.API) {
	tags := []string{"tasks"}

	Register(api, huma.Operation{
		OperationID: "tasks-bulk-update",
		Summary:     "Bulk update tasks",
		Description: "Applies the fields named in `fields` from `values` to every task in `task_ids`. The user needs write access to every project the involved tasks belong to; if write is missing on even one, the whole request is rejected and nothing is changed. Returns the updated tasks.",
		Method:      http.MethodPut,
		Path:        "/tasks/bulk",
		Tags:        tags,
	}, tasksBulkUpdate)
}

func init() { AddRouteRegistrar(RegisterBulkTaskRoutes) }

func tasksBulkUpdate(ctx context.Context, in *struct {
	Format string `query:"format" enum:"html,markdown" doc:"How rich-text fields are exchanged. See the API description."`
	Body   models.BulkTask
}) (*singleBody[models.BulkTask], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	bt := &in.Body
	if bt.Values != nil {
		if err := convertToHTML(ctx, &bt.Values.Description); err != nil {
			return nil, translateDomainError(err)
		}
	}
	if err := handler.DoUpdate(ctx, bt, a); err != nil {
		return nil, translateDomainError(err)
	}
	// Echo values + updated tasks back in the requested format (values.description
	// was converted to HTML above for persistence).
	convertTasksToMarkdown(ctx, append([]*models.Task{bt.Values}, bt.Tasks...)...)
	return &singleBody[models.BulkTask]{Body: bt}, nil
}
