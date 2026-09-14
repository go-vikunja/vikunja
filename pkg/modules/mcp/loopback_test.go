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
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/humabridge"
	apiv2 "code.vikunja.io/api/pkg/routes/api/v2"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCaller(t *testing.T) (*Module, context.Context) {
	t.Helper()
	m := newTestModule(t)
	m.authorize = func(*models.APIToken, string, string) bool { return true }
	caller := httptest.NewRequest(http.MethodPost, "/api/v2/mcp", nil)
	caller.Header.Set("Authorization", "Bearer tk_test")
	caller.RemoteAddr = "203.0.113.9:4242"
	ec := echo.New().NewContext(caller, httptest.NewRecorder())
	ec.Set("api_token", &models.APIToken{ID: 1})
	return m, context.WithValue(context.Background(), humabridge.EchoContextKey, ec)
}
func echoResult(t *testing.T, res any) map[string]any {
	t.Helper()
	m, ok := res.(map[string]any)
	require.True(t, ok, "%T", res)
	return m
}
func TestCallTool_ListForwardsAuthAndFormat(t *testing.T) {
	m, ctx := newTestCaller(t)
	res, err := m.callTool(ctx, "things_list", json.RawMessage(`{"q":"x","expand":["a","b"]}`))
	require.NoError(t, err)
	out := echoResult(t, res)
	assert.Equal(t, "Bearer tk_test", out["auth"])
	q := out["query"].(map[string]any)
	assert.Equal(t, "x", q["q"])
	assert.Empty(t, q["format"])
	assert.Equal(t, "a|b", q["expand"])
	res, err = m.callTool(ctx, "things_list", json.RawMessage(`{"format":"markdown"}`))
	require.NoError(t, err)
	assert.Equal(t, "markdown", echoResult(t, res)["query"].(map[string]any)["format"])
}
func TestCallTool_NullParamIsOmitted(t *testing.T) {
	m, ctx := newTestCaller(t)
	res, err := m.callTool(ctx, "things_list", json.RawMessage(`{"q":"x","expand":null}`))
	require.NoError(t, err)
	q := echoResult(t, res)["query"].(map[string]any)
	assert.Equal(t, "x", q["q"])
	assert.Empty(t, q["expand"])
	res, err = m.callTool(ctx, "things_list", json.RawMessage(`{"q":null}`))
	require.NoError(t, err)
	assert.Empty(t, echoResult(t, res)["query"].(map[string]any)["q"])
}
func TestCallTool_NullBodyFieldClearsIt(t *testing.T) {
	m, ctx := newTestCaller(t)
	res, err := m.callTool(ctx, "things_update", json.RawMessage(`{"id":7,"description":null}`))
	require.NoError(t, err)
	body := echoResult(t, res)["body"].(map[string]any)
	assert.Equal(t, "stored", body["title"])
	assert.Empty(t, body["description"])
}
func TestCallTool_CreateSendsJSONBody(t *testing.T) {
	m, ctx := newTestCaller(t)
	res, err := m.callTool(ctx, "things_create", json.RawMessage(`{"title":"hi","done":true}`))
	require.NoError(t, err)
	out := echoResult(t, res)
	assert.Equal(t, "application/json", out["content_type"])
	assert.Equal(t, "hi", out["body"].(map[string]any)["title"])
}
func TestCallTool_UpdateIsMergePatch(t *testing.T) {
	m, ctx := newTestCaller(t)
	res, err := m.callTool(ctx, "things_update", json.RawMessage(`{"id":7,"done":true}`))
	require.NoError(t, err)
	out := echoResult(t, res)
	assert.Equal(t, http.MethodPut, out["method"])
	body := out["body"].(map[string]any)
	assert.Equal(t, "stored", body["title"])
	assert.Equal(t, "keep me", body["description"])
	assert.Equal(t, true, body["done"])
}
func TestCallTool_DeleteReturnsOK(t *testing.T) {
	m, ctx := newTestCaller(t)
	res, err := m.callTool(ctx, "things_delete", json.RawMessage(`{"id":1}`))
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"ok": true}, res)
}
func TestCallTool_HTTPErrorIsToolError(t *testing.T) {
	m, ctx := newTestCaller(t)
	_, err := m.callTool(ctx, "things_read", json.RawMessage(`{"id":404}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404 Not Found")
	assert.Contains(t, err.Error(), "no such thing")
}
func TestCallTool_UnauthorizedMentionsScopes(t *testing.T) {
	m, ctx := newTestCaller(t)
	_, err := m.callTool(ctx, "things_read", json.RawMessage(`{"id":401}`))
	require.Error(t, err)
	assert.Equal(t, "401 Unauthorized: Unauthorized — "+scopeHint, err.Error())
}
func TestCallTool_UnknownArgumentRejectedBeforeDispatch(t *testing.T) {
	m, ctx := newTestCaller(t)
	_, err := m.callTool(ctx, "things_read", json.RawMessage(`{"id":1,"bogus":1}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bogus")
}
func TestCallTool_ScopeDenied(t *testing.T) {
	m, ctx := newTestCaller(t)
	m.authorize = func(*models.APIToken, string, string) bool { return false }
	_, err := m.callTool(ctx, "things_read", json.RawMessage(`{"id":1}`))
	require.ErrorIs(t, err, errScopeDenied)
}
func TestCallTool_UnknownTool(t *testing.T) {
	m, ctx := newTestCaller(t)
	_, err := m.callTool(ctx, "nope", nil)
	require.ErrorIs(t, err, errToolNotFound)
}
func TestCallTool_ClientAddress(t *testing.T) {
	m, ctx := newTestCaller(t)
	caller := humabridge.EchoContextFrom(ctx).Request()
	caller.Host = "vikunja.example.com"
	caller.Header.Add("X-Forwarded-For", "203.0.113.9")
	caller.Header.Add("X-Forwarded-For", "198.51.100.4")
	caller.Header.Set("X-Request-Id", "req-1")
	caller.Header.Set("Accept-Language", "de")
	caller.Header.Set("X-Forwarded-Proto", "https")
	tl := m.index["things_read"]
	req, err := tl.newRequest(ctx, humabridge.EchoContextFrom(ctx), map[string]json.RawMessage{"id": json.RawMessage(`1`)})
	require.NoError(t, err)
	assert.Equal(t, caller.RemoteAddr, req.RemoteAddr)
	assert.Equal(t, "vikunja.example.com", req.Host)
	assert.Equal(t, []string{
		"203.0.113.9",
		"198.51.100.4",
	}, req.Header.Values("X-Forwarded-For"))
	assert.Equal(t, "req-1", req.Header.Get("X-Request-Id"))
	assert.Equal(t, "vikunja.example.com", req.Header.Get("X-Forwarded-Host"))
	assert.Empty(t, req.Header.Get("Accept-Language"))
	assert.Empty(t, req.Header.Get("X-Forwarded-Proto"))
}
func TestCallTool_UsesTheResponseRequestID(t *testing.T) {
	m, ctx := newTestCaller(t)
	ec := humabridge.EchoContextFrom(ctx)
	ec.Response().Header().Set(echo.HeaderXRequestID, "generated-1")
	ec.Request().Header.Set(echo.HeaderXRequestID, "from-client")
	tl := m.index["things_read"]
	req, err := tl.newRequest(ctx, ec, map[string]json.RawMessage{"id": json.RawMessage(`1`)})
	require.NoError(t, err)
	assert.Equal(t, "generated-1", req.Header.Get(echo.HeaderXRequestID))
}
func TestCallTool_KeepsTheForwardedHost(t *testing.T) {
	m, ctx := newTestCaller(t)
	caller := humabridge.EchoContextFrom(ctx).Request()
	caller.Host = "internal:3456"
	caller.Header.Set("X-Forwarded-Host", "vikunja.example.com")
	tl := m.index["things_read"]
	req, err := tl.newRequest(ctx, humabridge.EchoContextFrom(ctx), map[string]json.RawMessage{"id": json.RawMessage(`1`)})
	require.NoError(t, err)
	assert.Equal(t, "vikunja.example.com", req.Header.Get("X-Forwarded-Host"))
}

func TestNewRequest_PatchSendsFormatAsAHeader(t *testing.T) {
	api := newRichTextAPI(t)
	spec := specFor(t, api, http.MethodPatch, "/notes/{id}")
	tl := &tool{
		op:          api.OpenAPI().Paths["/notes/{id}"].Patch,
		spec:        spec,
		contentType: "application/merge-patch+json",
	}
	_, ctx := newTestCaller(t)
	req, err := tl.newRequest(ctx, humabridge.EchoContextFrom(ctx), map[string]json.RawMessage{
		"id":          json.RawMessage(`5`),
		"format":      json.RawMessage(`"markdown"`),
		"description": json.RawMessage(`"**bold**"`),
	})
	require.NoError(t, err)
	assert.Equal(t, "markdown", req.Header.Get(apiv2.RichTextFormatHeader))
	assert.Equal(t, "/notes/5", req.URL.String())
	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"description":"**bold**"}`, string(body))
}
