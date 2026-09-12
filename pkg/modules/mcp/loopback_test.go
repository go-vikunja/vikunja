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
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/humabridge"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withTestCaller(t *testing.T) context.Context {
	t.Helper()
	Init(newTestAPI(t), "")
	prev := routeAuthorizer
	routeAuthorizer = func(*models.APIToken, string, string) bool { return true }
	t.Cleanup(func() { routeAuthorizer = prev })
	caller := httptest.NewRequest(http.MethodPost, "/api/v2/mcp", nil)
	caller.Header.Set("Authorization", "Bearer tk_test")
	caller.RemoteAddr = "203.0.113.9:4242"
	ec := echo.New().NewContext(caller, httptest.NewRecorder())
	ec.Set("api_token", &models.APIToken{ID: 1})
	return context.WithValue(context.Background(), humabridge.EchoContextKey, ec)
}
func echoResult(t *testing.T, res any) map[string]any {
	t.Helper()
	m, ok := res.(map[string]any)
	require.True(t, ok, "%T", res)
	return m
}
func TestCallTool_ListForwardsAuthAndFormat(t *testing.T) {
	ctx := withTestCaller(t)
	res, err := callTool(ctx, "things_list", json.RawMessage(`{"q":"x","expand":["a","b"]}`))
	require.NoError(t, err)
	m := echoResult(t, res)
	assert.Equal(t, "Bearer tk_test", m["auth"])
	q := m["query"].(map[string]any)
	assert.Equal(t, "x", q["q"])
	assert.Empty(t, q["format"])
	assert.Equal(t, "a|b", q["expand"])
	res, err = callTool(ctx, "things_list", json.RawMessage(`{"format":"markdown"}`))
	require.NoError(t, err)
	assert.Equal(t, "markdown", echoResult(t, res)["query"].(map[string]any)["format"])
}
func TestCallTool_NullParamIsOmitted(t *testing.T) {
	res, err := callTool(withTestCaller(t), "things_list", json.RawMessage(`{"q":"x","expand":null}`))
	require.NoError(t, err)
	q := echoResult(t, res)["query"].(map[string]any)
	assert.Equal(t, "x", q["q"])
	assert.Empty(t, q["expand"])
}
func TestCallTool_CreateSendsJSONBody(t *testing.T) {
	res, err := callTool(withTestCaller(t), "things_create", json.RawMessage(`{"title":"hi","done":true}`))
	require.NoError(t, err)
	m := echoResult(t, res)
	assert.Equal(t, "application/json", m["content_type"])
	assert.Equal(t, "hi", m["body"].(map[string]any)["title"])
}
func TestCallTool_UpdateIsMergePatch(t *testing.T) {
	res, err := callTool(withTestCaller(t), "things_update", json.RawMessage(`{"id":7,"done":true}`))
	require.NoError(t, err)
	m := echoResult(t, res)
	assert.Equal(t, http.MethodPut, m["method"])
	body := m["body"].(map[string]any)
	assert.Equal(t, "stored", body["title"])
	assert.Equal(t, "keep me", body["description"])
	assert.Equal(t, true, body["done"])
}
func TestCallTool_DeleteReturnsOK(t *testing.T) {
	res, err := callTool(withTestCaller(t), "things_delete", json.RawMessage(`{"id":1}`))
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"ok": true}, res)
}
func TestCallTool_HTTPErrorIsToolError(t *testing.T) {
	_, err := callTool(withTestCaller(t), "things_read", json.RawMessage(`{"id":404}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404 Not Found")
	assert.Contains(t, err.Error(), "no such thing")
}
func TestCallTool_UnauthorizedMentionsScopes(t *testing.T) {
	_, err := callTool(withTestCaller(t), "things_read", json.RawMessage(`{"id":401}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401 Unauthorized")
	assert.Contains(t, err.Error(), "scope")
}
func TestCallTool_UnknownArgumentRejectedBeforeDispatch(t *testing.T) {
	_, err := callTool(withTestCaller(t), "things_read", json.RawMessage(`{"id":1,"bogus":1}`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bogus")
}
func TestCallTool_ScopeDenied(t *testing.T) {
	ctx := withTestCaller(t)
	routeAuthorizer = func(*models.APIToken, string, string) bool { return false }
	_, err := callTool(ctx, "things_read", json.RawMessage(`{"id":1}`))
	require.ErrorIs(t, err, errScopeDenied)
}
func TestCallTool_UnknownTool(t *testing.T) {
	_, err := callTool(withTestCaller(t), "nope", nil)
	require.ErrorIs(t, err, errToolNotFound)
}
func TestCallTool_ClientAddress(t *testing.T) {
	ctx := withTestCaller(t)
	caller := echoContextFrom(ctx).Request()
	caller.Host = "vikunja.example.com"
	caller.Header.Add("X-Forwarded-For", "203.0.113.9")
	caller.Header.Add("X-Forwarded-For", "198.51.100.4")
	caller.Header.Set("X-Request-Id", "req-1")
	tl, _ := findTool("things_read")
	req, err := tl.newRequest(ctx, caller, map[string]json.RawMessage{"id": json.RawMessage(`1`)})
	require.NoError(t, err)
	assert.Equal(t, caller.RemoteAddr, req.RemoteAddr)
	assert.Equal(t, "vikunja.example.com", req.Host)
	assert.Equal(t, []string{
		"203.0.113.9",
		"198.51.100.4",
	}, req.Header.Values("X-Forwarded-For"))
	assert.Equal(t, "req-1", req.Header.Get("X-Request-Id"))
}
