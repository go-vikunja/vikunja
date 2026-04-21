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
	"github.com/danielgtaylor/huma/v2/conditional"
)

// Element type is *models.LabelWithTaskID because that's what
// models.Label.ReadAll returns; TaskID is json:"-", so the wire shape
// matches plain Label.
type labelListBody struct {
	Body Paginated[*models.LabelWithTaskID]
}

// RegisterLabelRoutes wires Label CRUD onto the Huma API.
func RegisterLabelRoutes(api huma.API) {
	tags := []string{"labels"}

	Register(api, huma.Operation{
		OperationID: "labels-list",
		Method:      http.MethodGet,
		Path:        "/labels",
		Tags:        tags,
	}, labelsList)

	Register(api, huma.Operation{
		OperationID: "labels-read",
		Method:      http.MethodGet,
		Path:        "/labels/{id}",
		Tags:        tags,
	}, labelsRead)

	Register(api, huma.Operation{
		OperationID: "labels-create",
		Method:      http.MethodPost,
		Path:        "/labels",
		Tags:        tags,
	}, labelsCreate)

	Register(api, huma.Operation{
		OperationID: "labels-update",
		Method:      http.MethodPut,
		Path:        "/labels/{id}",
		Tags:        tags,
	}, labelsUpdate)

	Register(api, huma.Operation{
		OperationID: "labels-delete",
		Method:      http.MethodDelete,
		Path:        "/labels/{id}",
		Tags:        tags,
	}, labelsDelete)
}

func labelsList(ctx context.Context, in *ListParams) (*labelListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.Label{}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	items, ok := result.([]*models.LabelWithTaskID)
	if !ok {
		return nil, fmt.Errorf("labels.ReadAll returned unexpected type %T (expected []*models.LabelWithTaskID)", result)
	}
	return &labelListBody{Body: NewPaginated(items, total, in.Page, in.PerPage)}, nil
}

func labelsRead(ctx context.Context, in *struct {
	ID int64 `path:"id"`
	conditional.Params
}) (*singleReadBody[models.Label], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	label := &models.Label{ID: in.ID}
	if _, err := handler.DoReadOne(ctx, label, a); err != nil {
		return nil, translateDomainError(err)
	}
	// PreconditionFailed wants the unquoted etag; response header uses RFC 9110 quoted form.
	etag := fmt.Sprintf("%d-%d", label.ID, label.Updated.UnixNano())
	if in.HasConditionalParams() {
		if err := in.PreconditionFailed(etag, label.Updated); err != nil {
			return nil, err
		}
	}
	return &singleReadBody[models.Label]{ETag: `"` + etag + `"`, Body: label}, nil
}

func labelsCreate(ctx context.Context, in *struct {
	Body models.Label
}) (*singleBody[models.Label], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.Label]{Body: &in.Body}, nil
}

func labelsUpdate(ctx context.Context, in *struct {
	ID   int64 `path:"id"`
	Body models.Label
}) (*singleBody[models.Label], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.ID = in.ID // URL wins over body
	if err := handler.DoUpdate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.Label]{Body: &in.Body}, nil
}

func labelsDelete(ctx context.Context, in *struct {
	ID int64 `path:"id"`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.Label{ID: in.ID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
