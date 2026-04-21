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

// Paginated is the standard list-response envelope for /api/v2.
type Paginated[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int64 `json:"total_pages"`
}

// --- Label ---

// labelBody is the standard single-Label response envelope (no cache headers).
type labelBody struct {
	Body *models.Label
}

// labelReadBody is the read-operation response envelope, carrying an ETag
// header so clients can issue If-None-Match for subsequent reads.
type labelReadBody struct {
	ETag string `header:"ETag"`
	Body *models.Label
}

// labelListBody wraps the paginated list response. The v1 ReadAll returns
// the hydrated LabelWithTaskID rows (label fields plus the task_id it's
// attached to) and we keep that shape here for contract parity.
type labelListBody struct {
	Body Paginated[*models.LabelWithTaskID]
}

type emptyBody struct{}

// jwtSecurity is the security requirement entry applied to every Label
// operation. Mirrors the "JWTKeyAuth" scheme declared in huma.go.
var jwtSecurity = []map[string][]string{{"JWTKeyAuth": {}}}

// RegisterLabelRoutes wires Label CRUD operations onto the given Huma API.
func RegisterLabelRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "labels-list",
		Method:      http.MethodGet,
		Path:        "/labels",
		Tags:        []string{"labels"},
		Security:    jwtSecurity,
	}, labelsList)

	huma.Register(api, huma.Operation{
		OperationID: "labels-read",
		Method:      http.MethodGet,
		Path:        "/labels/{id}",
		Tags:        []string{"labels"},
		Security:    jwtSecurity,
	}, labelsRead)

	huma.Register(api, huma.Operation{
		OperationID:   "labels-create",
		Method:        http.MethodPost,
		Path:          "/labels",
		Tags:          []string{"labels"},
		Security:      jwtSecurity,
		DefaultStatus: http.StatusCreated,
	}, labelsCreate)

	huma.Register(api, huma.Operation{
		OperationID: "labels-update",
		Method:      http.MethodPut,
		Path:        "/labels/{id}",
		Tags:        []string{"labels"},
		Security:    jwtSecurity,
	}, labelsUpdate)

	huma.Register(api, huma.Operation{
		OperationID:   "labels-delete",
		Method:        http.MethodDelete,
		Path:          "/labels/{id}",
		Tags:          []string{"labels"},
		Security:      jwtSecurity,
		DefaultStatus: http.StatusNoContent,
	}, labelsDelete)
}

// --- handlers ---

func labelsList(ctx context.Context, in *struct {
	Page    int    `query:"page"     default:"1"  minimum:"1"`
	PerPage int    `query:"per_page" default:"50" minimum:"1" maximum:"1000"`
	Q       string `query:"q"`
}) (*labelListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	result, _, total, err := handler.DoReadAll(ctx, &models.Label{}, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	// Concrete type cast — prevents the generic-any silent-empty trap the
	// spike hit, where an `interface{}` slice marshalled to an empty JSON
	// array without a loud failure.
	items, ok := result.([]*models.LabelWithTaskID)
	if !ok {
		return nil, fmt.Errorf("labels.ReadAll returned unexpected type %T (expected []*models.LabelWithTaskID)", result)
	}
	if items == nil {
		items = []*models.LabelWithTaskID{}
	}
	totalPages := int64(0)
	if in.PerPage > 0 {
		totalPages = (total + int64(in.PerPage) - 1) / int64(in.PerPage)
	}
	return &labelListBody{Body: Paginated[*models.LabelWithTaskID]{
		Items:      items,
		Total:      total,
		Page:       in.Page,
		PerPage:    in.PerPage,
		TotalPages: totalPages,
	}}, nil
}

func labelsRead(ctx context.Context, in *struct {
	ID int64 `path:"id"`
	conditional.Params
}) (*labelReadBody, error) {
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
	return &labelReadBody{ETag: `"` + etag + `"`, Body: label}, nil
}

func labelsCreate(ctx context.Context, in *struct {
	Body models.Label
}) (*labelBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &labelBody{Body: &in.Body}, nil
}

func labelsUpdate(ctx context.Context, in *struct {
	ID   int64 `path:"id"`
	Body models.Label
}) (*labelBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.ID = in.ID // URL wins over body
	if err := handler.DoUpdate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &labelBody{Body: &in.Body}, nil
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
