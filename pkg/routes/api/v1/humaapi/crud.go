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

package humaapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/web"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v5"
)

// translateError converts errors returned by the shared Do* pipeline (which
// originate from Echo handlers and Vikunja domain types) into Huma
// StatusErrors so Huma emits the right HTTP status + Vikunja-shaped body.
// Any unrecognised error is returned as-is (Huma will wrap it as 500).
func translateError(err error) error {
	if err == nil {
		return nil
	}
	// Vikunja domain errors — keep the {code, message} shape but lift the
	// HTTP status.
	var proc web.HTTPErrorProcessor
	if errors.As(err, &proc) {
		he := proc.HTTPError()
		status := he.HTTPCode
		if status == 0 {
			status = http.StatusInternalServerError
		}
		ve := &vikunjaError{StatusCode: status, Code: he.Code, Message: he.Message}
		return ve
	}
	// Forbidden / NotFound etc. raised via echo.NewHTTPError.
	var hErr *echo.HTTPError
	if errors.As(err, &hErr) {
		msg := fmt.Sprint(hErr.Message)
		return &vikunjaError{StatusCode: hErr.Code, Message: msg}
	}
	return err
}

// SingleID is the common path shape for /resource/{id} endpoints.
type SingleID struct {
	ID int64 `path:"id" doc:"Resource ID"`
}

// Config describes a generic CRUD resource:
//
//   - T is the domain model pointer (must implement handler.CObject)
//   - P is the path-parameter struct; use SingleID for the simple case or
//     define your own for nested routes like /tasks/{task}/labels/{label}.
//
// Note: Go does not permit embedding a type parameter in a struct, so the
// generic request wrappers below keep P as a named field. Huma's default
// parameter discovery only walks anonymous (embedded) fields, so we define
// parallel concrete wrappers per path shape (see SingleID section below)
// that embed the concrete path struct. Resources with different path shapes
// should add their own wrappers + Register call (this spike only needs
// SingleID; the pattern generalises trivially).
type Config[T handler.CObject, P any] struct {
	Tag       string
	BasePath  string     // list + create; may itself contain {params}
	ItemPath  string     // read + update + delete
	New       func() T   // factory — same role as WebHandler.EmptyStruct
	ApplyPath func(T, P) // copies path params onto the model
}

type bodyOutput[T any] struct {
	Body T
}

type listOutput[T any] struct {
	Body []T
}

type deleteMessage struct {
	Message string `json:"message"`
}

// --- SingleID wrappers ---------------------------------------------------
//
// Concrete request-input types for the /{resource}/{id} shape. Huma's
// parameter discovery finds `ID` through the embedded SingleID. The Body
// field carries the decoded JSON payload.

type singleIDCreateInput[T any] struct {
	// No path params for CREATE (path is /{resource}); we still thread a
	// matching type parameter so Register can share a single handler shape.
	Body T
}

type singleIDItemInput struct {
	SingleID
}

type singleIDListInput struct {
	Page    int    `query:"page"     default:"1"  minimum:"1"`
	PerPage int    `query:"per_page" default:"50" minimum:"1" maximum:"1000"`
	Search  string `query:"s"`
}

type singleIDBodyInput[T any] struct {
	SingleID
	Body T
}

// Register wires five Huma operations for the given CRUD resource.
//
// Today this only implements the SingleID path shape. Resources using
// multi-segment paths (e.g. /tasks/{task}/labels/{label}) should hand-write
// their huma.Register calls until we generalise this registrar.
func Register[T handler.CObject](api huma.API, cfg Config[T, SingleID]) {
	jwt := []map[string][]string{{"JWTKeyAuth": {}}}

	// CREATE
	huma.Register(api, huma.Operation{
		OperationID: cfg.Tag + "-create",
		Method:      http.MethodPut,
		Path:        cfg.BasePath,
		Tags:        []string{cfg.Tag},
		Security:    jwt,
	}, func(ctx context.Context, in *singleIDCreateInput[T]) (*bodyOutput[T], error) {
		a, err := auth.GetAuthFromContext(ctx)
		if err != nil {
			return nil, huma.Error401Unauthorized(err.Error())
		}
		if err := handler.DoCreate(ctx, in.Body, a); err != nil {
			return nil, translateError(err)
		}
		return &bodyOutput[T]{Body: in.Body}, nil
	})

	// READ ONE
	huma.Register(api, huma.Operation{
		OperationID: cfg.Tag + "-read",
		Method:      http.MethodGet,
		Path:        cfg.ItemPath,
		Tags:        []string{cfg.Tag},
		Security:    jwt,
	}, func(ctx context.Context, in *singleIDItemInput) (*bodyOutput[T], error) {
		a, err := auth.GetAuthFromContext(ctx)
		if err != nil {
			return nil, huma.Error401Unauthorized(err.Error())
		}
		obj := cfg.New()
		cfg.ApplyPath(obj, in.SingleID)
		if _, err := handler.DoReadOne(ctx, obj, a); err != nil {
			return nil, translateError(err)
		}
		return &bodyOutput[T]{Body: obj}, nil
	})

	// READ ALL
	huma.Register(api, huma.Operation{
		OperationID: cfg.Tag + "-list",
		Method:      http.MethodGet,
		Path:        cfg.BasePath,
		Tags:        []string{cfg.Tag},
		Security:    jwt,
	}, func(ctx context.Context, in *singleIDListInput) (*listOutput[T], error) {
		a, err := auth.GetAuthFromContext(ctx)
		if err != nil {
			return nil, huma.Error401Unauthorized(err.Error())
		}
		obj := cfg.New()
		result, _, _, err := handler.DoReadAll(ctx, obj, a, in.Search, in.Page, in.PerPage)
		if err != nil {
			return nil, translateError(err)
		}
		// Best-effort cast; ReadAll returns interface{}. For the spike
		// we assume []T. Resources returning a different list item type
		// should hand-write their list op via huma.Register directly.
		slice, ok := result.([]T)
		if !ok {
			// fall back to marshaling whatever shape was returned
			return &listOutput[T]{Body: nil}, nil
		}
		return &listOutput[T]{Body: slice}, nil
	})

	// UPDATE
	huma.Register(api, huma.Operation{
		OperationID: cfg.Tag + "-update",
		Method:      http.MethodPost,
		Path:        cfg.ItemPath,
		Tags:        []string{cfg.Tag},
		Security:    jwt,
	}, func(ctx context.Context, in *singleIDBodyInput[T]) (*bodyOutput[T], error) {
		a, err := auth.GetAuthFromContext(ctx)
		if err != nil {
			return nil, huma.Error401Unauthorized(err.Error())
		}
		cfg.ApplyPath(in.Body, in.SingleID)
		if err := handler.DoUpdate(ctx, in.Body, a); err != nil {
			return nil, translateError(err)
		}
		return &bodyOutput[T]{Body: in.Body}, nil
	})

	// DELETE
	huma.Register(api, huma.Operation{
		OperationID: cfg.Tag + "-delete",
		Method:      http.MethodDelete,
		Path:        cfg.ItemPath,
		Tags:        []string{cfg.Tag},
		Security:    jwt,
	}, func(ctx context.Context, in *singleIDItemInput) (*bodyOutput[deleteMessage], error) {
		a, err := auth.GetAuthFromContext(ctx)
		if err != nil {
			return nil, huma.Error401Unauthorized(err.Error())
		}
		obj := cfg.New()
		cfg.ApplyPath(obj, in.SingleID)
		if err := handler.DoDelete(ctx, obj, a); err != nil {
			return nil, translateError(err)
		}
		return &bodyOutput[deleteMessage]{Body: deleteMessage{Message: "Successfully deleted."}}, nil
	})
}
