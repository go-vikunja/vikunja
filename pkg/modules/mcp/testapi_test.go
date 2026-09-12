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

package mcp

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/autopatch"
	"github.com/danielgtaylor/huma/v2/humatest"
)

type testThing struct {
	ID          int64       `json:"id" readOnly:"true"`
	Title       string      `json:"title" valid:"required" minLength:"1"`
	Description string      `json:"description"`
	Done        bool        `json:"done"`
	Owner       *testOwner  `json:"owner" readOnly:"true"`
	Reminders   []*testWhen `json:"reminders"`
}
type testOwner struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
type testWhen struct {
	At string `json:"at" format:"date-time"`
}
type testThingBody struct{ Body *testThing }
type echoBody struct {
	Body struct {
		Method string            `json:"method"`
		Path   string            `json:"path"`
		Query  map[string]string `json:"query"`
		Auth   string            `json:"auth"`
		CT     string            `json:"content_type"`
		Body   *testThing        `json:"body,omitempty"`
	}
}

func newTestAPI(t *testing.T) huma.API {
	t.Helper()
	cfg := huma.DefaultConfig("test", "1")
	cfg.FieldsOptionalByDefault = true
	_, api := humatest.New(t, cfg)
	huma.Register(api, huma.Operation{
		OperationID: "things-list",
		Method:      http.MethodGet,
		Path:        "/things",
		Summary:     "List things",
	},
		func(_ context.Context, in *struct {
			Q      string   `query:"q"`
			Page   int      `query:"page" default:"1" minimum:"1"`
			Expand []string `query:"expand,explode" enum:"a,b"`
			Format string   `query:"format" enum:"html,markdown"`
			Auth   string   `header:"Authorization"`
		}) (*echoBody, error) {
			out := &echoBody{}
			out.Body.Method, out.Body.Path, out.Body.Auth = http.MethodGet, "/things", in.Auth
			out.Body.Query = map[string]string{
				"q":      in.Q,
				"format": in.Format,
				"expand": strings.Join(in.Expand, "|"),
			}
			return out, nil
		})
	huma.Register(api, huma.Operation{
		OperationID: "things-read",
		Method:      http.MethodGet,
		Path:        "/things/{id}",
	}, func(_ context.Context, in *struct {
		ID int64 `path:"id"`
	}) (*testThingBody, error) {
		if in.ID == 404 {
			return nil, huma.Error404NotFound("no such thing")
		}
		if in.ID == 401 {
			return nil, huma.Error401Unauthorized("Unauthorized")
		}
		return &testThingBody{
			Body: &testThing{
				ID:          in.ID,
				Title:       "stored",
				Description: "keep me",
				Owner:       &testOwner{ID: 1},
			},
		}, nil
	})
	huma.Register(api, huma.Operation{
		OperationID: "things-create",
		Method:      http.MethodPost,
		Path:        "/things",
	}, func(_ context.Context, in *struct {
		Body testThing
		CT   string `header:"Content-Type"`
	}) (*echoBody, error) {
		out := &echoBody{}
		out.Body.Method, out.Body.Path, out.Body.CT = http.MethodPost, "/things", in.CT
		out.Body.Body = &in.Body
		return out, nil
	})
	huma.Register(api, huma.Operation{
		OperationID: "things-update",
		Method:      http.MethodPut,
		Path:        "/things/{id}",
	}, func(_ context.Context, in *struct {
		ID   int64 `path:"id"`
		Body testThing
		CT   string `header:"Content-Type"`
	}) (*echoBody, error) {
		out := &echoBody{}
		out.Body.Method, out.Body.Path, out.Body.CT = http.MethodPut, "/things/"+strconv.FormatInt(in.ID, 10), in.CT
		out.Body.Body = &in.Body
		return out, nil
	})
	huma.Register(api, huma.Operation{
		OperationID: "things-delete",
		Method:      http.MethodDelete,
		Path:        "/things/{id}",
	}, func(_ context.Context, _ *struct {
		ID int64 `path:"id"`
	}) (*struct{}, error) {
		return nil, nil
	})
	huma.Register(api, huma.Operation{
		OperationID: "tokens-create",
		Method:      http.MethodPost,
		Path:        "/tokens",
	}, func(_ context.Context, _ *struct{ Body testThing }) (*testThingBody, error) { return nil, nil })
	huma.Register(api, huma.Operation{
		OperationID: "things-upload",
		Method:      http.MethodPost,
		Path:        "/things/{id}/file",
	}, func(_ context.Context, _ *struct {
		ID      int64 `path:"id"`
		RawBody huma.MultipartFormFiles[struct{}]
	}) (*struct{}, error) {
		return nil, nil
	})
	autopatch.AutoPatch(api)
	return api
}
