package humaecho5_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/modules/humaecho5"
	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAdapterRoundtrip proves that a Huma operation registered against the
// v5 adapter is served by Echo and that the echo.Context is retrievable
// from the handler's context.Context via EchoContextKey.
func TestAdapterRoundtrip(t *testing.T) {
	e := echo.New()
	api := humaecho5.New(e, huma.DefaultConfig("spike", "0.0.1"))

	type pingInput struct {
		Name string `path:"name"`
	}
	type pingOutput struct {
		Body struct {
			Echo string `json:"echo"`
			HasEchoCtx bool `json:"has_echo_ctx"`
		}
	}

	huma.Register(api, huma.Operation{
		OperationID: "ping",
		Method:      "GET",
		Path:        "/ping/{name}",
	}, func(ctx context.Context, in *pingInput) (*pingOutput, error) {
		_, ok := ctx.Value(humaecho5.EchoContextKey).(*echo.Context)
		out := &pingOutput{}
		out.Body.Echo = in.Name
		out.Body.HasEchoCtx = ok
		return out, nil
	})

	req := httptest.NewRequest("GET", "/ping/world", nil)
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

// TestOpenAPISpecServed proves Huma serves the OAS 3.1 spec document
// on its configured URL.
func TestOpenAPISpecServed(t *testing.T) {
	e := echo.New()
	_ = humaecho5.New(e, huma.DefaultConfig("spike", "0.0.1"))
	req := httptest.NewRequest("GET", "/openapi.json", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, 200, rec.Code)
	assert.True(t, strings.Contains(rec.Body.String(), `"openapi":"3.1`),
		"expected OAS 3.1 header, got %s", rec.Body.String())
}
