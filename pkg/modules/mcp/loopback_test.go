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
	return WithCaller(WithToken(context.Background(), &models.APIToken{ID: 1}), caller)
}
func echoResult(t *testing.T, res any) map[string]any {
	t.Helper()
	m, ok := res.(map[string]any)
	require.True(t, ok, "%T", res)
	return m
}
func TestCallTool_ListForwardsAuthAndDefaultsMarkdown(t *testing.T) {
	ctx := withTestCaller(t)
	res, err := callTool(ctx, "things_list", json.RawMessage(`{"q":"x","expand":["a","b"]}`))
	require.NoError(t, err)
	m := echoResult(t, res)
	assert.Equal(t, "Bearer tk_test", m["auth"])
	q := m["query"].(map[string]any)
	assert.Equal(t, "x", q["q"])
	assert.Equal(t, "markdown", q["format"])
	assert.Equal(t, "a|b", q["expand"])
	res, err = callTool(ctx, "things_list", json.RawMessage(`{"format":"html"}`))
	require.NoError(t, err)
	assert.Equal(t, "html", echoResult(t, res)["query"].(map[string]any)["format"])
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
	var apiErr *apiError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusNotFound, apiErr.status)
	assert.Contains(t, err.Error(), "no such thing")
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
	require.ErrorIs(t, err, ErrScopeDenied)
}
func TestCallTool_UnknownTool(t *testing.T) {
	_, err := callTool(withTestCaller(t), "nope", nil)
	require.ErrorIs(t, err, ErrToolNotFound)
}
func TestCallTool_ClientAddress(t *testing.T) {
	ctx := withTestCaller(t)
	caller := CallerFromContext(ctx)
	caller.Header.Set("X-Forwarded-For", "203.0.113.9")
	tl, _ := findTool("things_read")
	req, err := tl.newRequest(ctx, caller, map[string]json.RawMessage{"id": json.RawMessage(`1`)})
	require.NoError(t, err)
	assert.Equal(t, caller.RemoteAddr, req.RemoteAddr)
	assert.Equal(t, "203.0.113.9", req.Header.Get("X-Forwarded-For"))
}
