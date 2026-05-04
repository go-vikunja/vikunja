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

// --- Label ---

// labelListBody wraps the paginated list response.
type labelListBody struct {
	Body Paginated[*models.Label]
}

// RegisterLabelRoutes wires Label CRUD operations onto the given Huma API.
// Auth is supplied globally via huma.Config (see NewAPI), so operations
// don't declare Security per-call.
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

// --- handlers ---

func labelsList(ctx context.Context, in *ListParams) (*labelListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.Label{}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	// models.Label.ReadAll reuses the v1 hydration path which returns
	// []*LabelWithTaskID — the wrapper exists to carry task_id internally
	// for other v1 callers, but TaskID is json:"-" so the wire shape is
	// already identical to []*Label. Unwrap here so the v2 surface only
	// references models.Label. Concrete type cast guards against the
	// generic-any silent-empty trap the spike hit.
	hydrated, ok := result.([]*models.LabelWithTaskID)
	if !ok {
		return nil, fmt.Errorf("labels.ReadAll returned unexpected type %T (expected []*models.LabelWithTaskID)", result)
	}
	items := make([]*models.Label, len(hydrated))
	for i, h := range hydrated {
		items[i] = &h.Label
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
	// ETag derives from the ID + last-updated timestamp so any edit
	// invalidates downstream caches. conditional.PreconditionFailed
	// expects the unquoted value; the response header uses the RFC 9110
	// quoted form.
	etag := fmt.Sprintf("%d-%d", label.ID, label.Updated.UnixNano())
	if in.HasConditionalParams() {
		// PreconditionFailed returns a 304 (reads) or 412 (writes) when
		// conditions aren't met; nil means continue.
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
