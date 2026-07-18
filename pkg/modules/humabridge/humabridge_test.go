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

package humabridge_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/modules/humabridge"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/autopatch"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPrefix = "/api/v2"

func newGroupAPI() (*echo.Echo, huma.API) {
	e := echo.New()
	g := e.Group(testPrefix)
	api := humabridge.NewWithGroup(e, g, testPrefix, huma.DefaultConfig("spike", "0.0.1"))
	return e, api
}

// TestAdapterRoundtrip proves that a Huma operation registered against the
// group-mounted adapter is served by Echo and that the echo.Context is
// retrievable from the handler's context.Context via EchoContextKey.
func TestAdapterRoundtrip(t *testing.T) {
	e, api := newGroupAPI()

	type pingInput struct {
		Name string `path:"name"`
	}
	type pingOutput struct {
		Body struct {
			Echo       string `json:"echo"`
			HasEchoCtx bool   `json:"has_echo_ctx"`
		}
	}

	huma.Register(api, huma.Operation{
		OperationID: "ping",
		Method:      "GET",
		Path:        "/ping/{name}",
	}, func(ctx context.Context, in *pingInput) (*pingOutput, error) {
		_, ok := ctx.Value(humabridge.EchoContextKey).(*echo.Context)
		out := &pingOutput{}
		out.Body.Echo = in.Name
		out.Body.HasEchoCtx = ok
		return out, nil
	})

	req := httptest.NewRequest("GET", testPrefix+"/ping/world", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, 200, rec.Code, "body: %s", rec.Body.String())
	var got struct {
		Echo       string `json:"echo"`
		HasEchoCtx bool   `json:"has_echo_ctx"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, "world", got.Echo)
	assert.True(t, got.HasEchoCtx, "echo.Context not stashed on request ctx")
}

// TestServeHTTPSkipsAlreadyPrefixedPath proves groupPrefixAdapter.ServeHTTP
// leaves a path alone when it already carries the group prefix, instead of
// prepending it again. Internal Huma dispatches (e.g. autopatch) always go
// through api.Adapter().ServeHTTP with an absolute, already-prefixed path,
// so this exercises that call path directly rather than through e.ServeHTTP.
// If the skip guard were removed, the prefix would be doubled to
// "/api/v2/api/v2/ping/world", which the router wouldn't match, and this
// test would fail with a 404.
func TestServeHTTPSkipsAlreadyPrefixedPath(t *testing.T) {
	_, api := newGroupAPI()

	type pingInput struct {
		Name string `path:"name"`
	}
	type pingOutput struct {
		Body struct {
			Echo string `json:"echo"`
		}
	}

	huma.Register(api, huma.Operation{
		OperationID: "ping-skip",
		Method:      "GET",
		Path:        "/ping/{name}",
	}, func(_ context.Context, in *pingInput) (*pingOutput, error) {
		out := &pingOutput{}
		out.Body.Echo = in.Name
		return out, nil
	})

	req := httptest.NewRequest("GET", testPrefix+"/ping/world", nil)
	rec := httptest.NewRecorder()
	api.Adapter().ServeHTTP(rec, req)

	require.Equal(t, 200, rec.Code, "body: %s", rec.Body.String())
	var got struct {
		Echo string `json:"echo"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, "world", got.Echo)
}

// TestOpenAPISpecServed proves Huma serves the OAS 3.1 spec document
// on its configured URL under the group prefix.
func TestOpenAPISpecServed(t *testing.T) {
	e, _ := newGroupAPI()
	req := httptest.NewRequest("GET", testPrefix+"/openapi.json", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, 200, rec.Code)
	assert.Contains(t, rec.Body.String(), `"openapi":"3.1`,
		"expected OAS 3.1 header, got %s", rec.Body.String())
}

// TestAutoPatchUnderGroup exercises the group-prefix rewrite: autopatch's
// synthesised PATCH re-dispatches GET and PUT with paths relative to the
// group, which only resolve if the adapter prepends the prefix again.
func TestAutoPatchUnderGroup(t *testing.T) {
	e, api := newGroupAPI()

	type thing struct {
		Title       string `json:"title"`
		Description string `json:"description,omitempty"`
	}
	stored := thing{Title: "original", Description: "keep me"}

	type thingInput struct {
		ID string `path:"id"`
	}
	type thingBody struct {
		Body thing
	}
	type thingPutInput struct {
		ID   string `path:"id"`
		Body thing
	}

	huma.Register(api, huma.Operation{
		OperationID: "thing-read",
		Method:      "GET",
		Path:        "/things/{id}",
	}, func(_ context.Context, _ *thingInput) (*thingBody, error) {
		return &thingBody{Body: stored}, nil
	})
	huma.Register(api, huma.Operation{
		OperationID: "thing-update",
		Method:      "PUT",
		Path:        "/things/{id}",
	}, func(_ context.Context, in *thingPutInput) (*thingBody, error) {
		stored = in.Body
		return &thingBody{Body: stored}, nil
	})
	autopatch.AutoPatch(api)

	req := httptest.NewRequest("PATCH", testPrefix+"/things/1", strings.NewReader(`{"title":"patched"}`))
	req.Header.Set("Content-Type", "application/merge-patch+json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, 200, rec.Code, "body: %s", rec.Body.String())
	assert.Equal(t, "patched", stored.Title)
	assert.Equal(t, "keep me", stored.Description, "merge patch must leave unrelated fields alone")
}
