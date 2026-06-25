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

// RegisterLabelTaskBulkRoutes wires the bulk label-replacement action onto the
// Huma API. The model op is a CRUDable Create (handler.DoCreate, whose
// CanCreate enforces write access to the task), but the verb is PUT because the
// operation replaces the task's whole label set — the idempotent PUT semantics
// describe it more honestly than POST.
func RegisterLabelTaskBulkRoutes(api huma.API) {
	tags := []string{"labels"}

	Register(api, huma.Operation{
		OperationID: "task-labels-bulk-replace",
		Summary:     "Replace all labels on a task",
		Description: "Sets the task's labels to exactly the provided list: labels not in the list are removed, missing ones are added, unchanged ones are left alone. Requires write access to the task, and you must be able to see every label you attach. Returns the resulting label set.",
		Method:      http.MethodPut,
		Path:        "/tasks/{projecttask}/labels/bulk",
		Tags:        tags,
	}, labelTasksBulkReplace)
}

func init() { AddRouteRegistrar(RegisterLabelTaskBulkRoutes) }

func labelTasksBulkReplace(ctx context.Context, in *struct {
	TaskID int64 `path:"projecttask" doc:"The numeric id of the task whose labels to replace."`
	Body   models.LabelTaskBulk
}) (*singleBody[models.LabelTaskBulk], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.TaskID = in.TaskID // parent from the path, not the body
	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.LabelTaskBulk]{Body: &in.Body}, nil
}
