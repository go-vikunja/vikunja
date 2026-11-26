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

package webtests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth/oauth2server"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/routes"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	metadataClientName     = "Example MCP Client"
	metadataClientRedirect = "http://127.0.0.1:3000/callback"
)

// metadataClientEnv serves a valid document at /client.json and a broken one at /broken.json.
func metadataClientEnv(t *testing.T) (e *echo.Echo, clientID string, host *httptest.Server) {
	t.Helper()
	host = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		doc := map[string]any{
			"client_id":   "https://" + r.Host + r.URL.Path,
			"client_name": metadataClientName,
			"client_uri":  "https://app.example.com",
			"redirect_uris": []string{
				metadataClientRedirect,
				"https://app.example.com/callback",
			},
		}
		switch r.URL.Path {
		case "/client.json":
		case "/broken.json":
			doc["client_secret"] = "s3cret"
		default:
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(doc)
	}))
	t.Cleanup(host.Close)

	e, err := setupTestEnv(routes.WithOAuthMetadataHTTPClient(host.Client()))
	require.NoError(t, err)
	return e, host.URL + "/client.json", host
}

func TestOAuthMetadataDocumentClient(t *testing.T) {
	e, clientID, host := metadataClientEnv(t)
	userToken := humaTokenFor(t, &testuser1)
	// The loopback port is chosen at runtime and need not match the registered one.
	redirect := "http://127.0.0.1:45678/callback"
	params := map[string]string{
		"client_id":    clientID,
		"redirect_uri": redirect,
	}
	exchangeParams := url.Values{
		"client_id":    {clientID},
		"redirect_uri": {redirect},
	}

	t.Run("preview shows the client name and verified host", func(t *testing.T) {
		res := authorizePreview(t, e, userToken, params)
		require.Equal(t, http.StatusOK, res.code, string(res.body))
		assert.JSONEq(t, `{
			"client": {
				"id": "`+clientID+`",
				"kind": "metadata_document",
				"name": "Example MCP Client",
				"host": "`+strings.TrimPrefix(host.URL, "https://")+`",
				"uri": "https://app.example.com"
			},
			"redirect_uri": "`+redirect+`",
			"scopes": ["mcp:read", "mcp:write"],
			"scoped": true
		}`, stripSchema(t, res.body))
	})

	t.Run("authorize and exchange yield a scoped token titled with the client name", func(t *testing.T) {
		authz := scopedAuthorize(t, e, "v2", map[string]string{
			"client_id":    clientID,
			"redirect_uri": redirect,
			"scope":        "mcp:read",
			"resource":     mcpResourceURL,
		})
		require.Equal(t, http.StatusOK, authz.code, string(authz.body))
		var authorized oauth2server.AuthorizeResponse
		require.NoError(t, json.Unmarshal(authz.body, &authorized))
		assert.Equal(t, "mcp:read", authorized.Scope)
		assert.Equal(t, redirect, authorized.RedirectURI)

		tok := scopedExchange(t, e, "v2", authorized.Code, exchangeParams)
		require.Equal(t, http.StatusOK, tok.code, string(tok.body))
		var raw map[string]any
		require.NoError(t, json.Unmarshal(tok.body, &raw))
		assert.NotContains(t, raw, "refresh_token")
		assert.Equal(t, "mcp:read", raw["scope"])

		stored := storedToken(t, raw["access_token"].(string))
		assert.Equal(t, metadataClientName, stored.Title)
		assert.Equal(t, clientID, stored.OAuthClientID)
		assert.Equal(t, mcpResourceURL, stored.OAuthResource)
		assert.True(t, stored.HasMCPAccess())
		assert.False(t, hasNonReadPermission(stored.APIPermissions), "%v", stored.APIPermissions)
	})

	t.Run("without scope or resource it still gets a scoped token, never a session", func(t *testing.T) {
		authz := scopedAuthorize(t, e, "v2", params)
		require.Equal(t, http.StatusOK, authz.code, string(authz.body))
		var authorized oauth2server.AuthorizeResponse
		require.NoError(t, json.Unmarshal(authz.body, &authorized))
		assert.Equal(t, "mcp:read mcp:write", authorized.Scope)

		tok := scopedExchange(t, e, "v2", authorized.Code, exchangeParams)
		require.Equal(t, http.StatusOK, tok.code, string(tok.body))
		var raw map[string]any
		require.NoError(t, json.Unmarshal(tok.body, &raw))
		assert.Regexp(t, `^tk_`, raw["access_token"])
		assert.NotContains(t, raw, "refresh_token")
		assert.Equal(t, metadataClientName, storedToken(t, raw["access_token"].(string)).Title)
	})

	t.Run("the token endpoint never issues a session for an unscoped code of a URL client", func(t *testing.T) {
		s := db.NewSession()
		code, err := models.CreateOAuthCode(s, &models.OAuthCode{
			UserID:              1,
			ClientID:            clientID,
			RedirectURI:         redirect,
			CodeChallenge:       pkceChallenge(scopedVerifier),
			CodeChallengeMethod: "S256",
		})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		s.Close()

		tok := scopedExchange(t, e, "v2", code, exchangeParams)
		require.Equal(t, http.StatusBadRequest, tok.code, string(tok.body))
		assert.Equal(t, models.ErrCodeOAuthInvalidScope, tok.errorCode(t))
		assert.NotContains(t, string(tok.body), "access_token")
	})

	t.Run("accepts a listed https redirect", func(t *testing.T) {
		res := authorizePreview(t, e, userToken, map[string]string{
			"client_id":    clientID,
			"redirect_uri": "https://app.example.com/callback",
		})
		assert.Equal(t, http.StatusOK, res.code, string(res.body))
	})

	t.Run("rejects an unlisted redirect URI", func(t *testing.T) {
		for _, redirect := range []string{
			"https://app.example.com/callback/",
			"https://evil.example/callback",
			"http://127.0.0.1:3000/other",
			"http://localhost:3000/callback",
			"vikunja-flutter://callback",
		} {
			for _, version := range []string{"v1", "v2"} {
				authz := scopedAuthorize(t, e, version, map[string]string{
					"client_id":    clientID,
					"redirect_uri": redirect,
				})
				require.Equal(t, http.StatusBadRequest, authz.code, "%s %s: %s", version, redirect, authz.body)
				assert.Equal(t, models.ErrCodeOAuthInvalidRedirectURI, authz.errorCode(t), redirect)
			}
		}
	})

	t.Run("rejects URL client ids that are not https metadata URLs", func(t *testing.T) {
		before := countOAuthCodes(t)
		for _, id := range []string{
			strings.Replace(clientID, "https://", "http://", 1),
			host.URL,
			clientID + "#frag",
		} {
			for _, res := range []*httpResult{
				authorizePreview(t, e, userToken, map[string]string{
					"client_id":    id,
					"redirect_uri": metadataClientRedirect,
				}),
				scopedAuthorize(t, e, "v2", map[string]string{
					"client_id":    id,
					"redirect_uri": metadataClientRedirect,
				}),
			} {
				require.Equal(t, http.StatusBadRequest, res.code, "%s: %s", id, res.body)
				assert.Equal(t, models.ErrCodeOAuthInvalidClientID, res.errorCode(t), id)
			}
		}
		assert.Equal(t, before, countOAuthCodes(t))
	})

	t.Run("reports an unreachable document", func(t *testing.T) {
		res := authorizePreview(t, e, userToken, map[string]string{
			"client_id":    host.URL + "/missing.json",
			"redirect_uri": metadataClientRedirect,
		})
		require.Equal(t, http.StatusBadGateway, res.code, string(res.body))
		assert.Equal(t, models.ErrCodeOAuthClientMetadataUnavailable, res.errorCode(t))
	})

	t.Run("reports an invalid document", func(t *testing.T) {
		res := authorizePreview(t, e, userToken, map[string]string{
			"client_id":    host.URL + "/broken.json",
			"redirect_uri": metadataClientRedirect,
		})
		require.Equal(t, http.StatusBadRequest, res.code, string(res.body))
		assert.Equal(t, models.ErrCodeOAuthInvalidClientMetadata, res.errorCode(t))
		assert.Contains(t, string(res.body), "client secret")
	})

	t.Run("legacy clients keep the session flow", func(t *testing.T) {
		res := authorizePreview(t, e, userToken, nil)
		require.Equal(t, http.StatusOK, res.code, string(res.body))
		var preview oauth2server.AuthorizePreview
		require.NoError(t, json.Unmarshal(res.body, &preview))
		assert.Equal(t, oauth2server.ClientKindLegacy, preview.Client.Kind)
		assert.False(t, preview.Scoped)

		authz := scopedAuthorize(t, e, "v2", nil)
		require.Equal(t, http.StatusOK, authz.code, string(authz.body))
		var authorized oauth2server.AuthorizeResponse
		require.NoError(t, json.Unmarshal(authz.body, &authorized))
		tok := scopedExchange(t, e, "v2", authorized.Code, nil)
		require.Equal(t, http.StatusOK, tok.code, string(tok.body))
		assert.Contains(t, string(tok.body), "refresh_token")
	})
}

// countingMetadataEnv serves a valid document at every path and counts the fetches.
func countingMetadataEnv(t *testing.T) (*echo.Echo, *httptest.Server, *atomic.Int32) {
	t.Helper()
	requests := &atomic.Int32{}
	host := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"client_id":     "https://" + r.Host + r.URL.Path,
			"client_name":   metadataClientName,
			"redirect_uris": []string{metadataClientRedirect},
		})
	}))
	t.Cleanup(host.Close)
	e, err := setupTestEnv(routes.WithOAuthMetadataHTTPClient(host.Client()))
	require.NoError(t, err)
	return e, host, requests
}

func TestOAuthMetadataDocumentFetchBudget(t *testing.T) {
	e, host, requests := countingMetadataEnv(t)
	userToken := humaTokenFor(t, &testuser1)
	clientID := host.URL + "/client.json"
	params := func(id string) map[string]string {
		return map[string]string{
			"client_id":    id,
			"redirect_uri": metadataClientRedirect,
		}
	}

	authz := scopedAuthorize(t, e, "v2", params(clientID))
	require.Equal(t, http.StatusOK, authz.code, string(authz.body))
	var authorized oauth2server.AuthorizeResponse
	require.NoError(t, json.Unmarshal(authz.body, &authorized))

	var limited *httpResult
	for i := 0; i < 1000 && limited == nil; i++ {
		res := authorizePreview(t, e, userToken, params(host.URL+"/client-"+strconv.Itoa(i)+".json"))
		if res.code != http.StatusOK {
			limited = res
		}
	}
	require.NotNil(t, limited, "the fetch budget never ran out")

	t.Run("exhausted budget answers 429 without fetching", func(t *testing.T) {
		before := requests.Load()
		for _, res := range []*httpResult{
			limited,
			authorizePreview(t, e, userToken, params(host.URL+"/one-more.json")),
			scopedAuthorize(t, e, "v2", params(host.URL+"/one-more.json")),
		} {
			require.Equal(t, http.StatusTooManyRequests, res.code, string(res.body))
			assert.Equal(t, models.ErrCodeOAuthClientMetadataFetchLimit, res.errorCode(t))
		}
		assert.Equal(t, before, requests.Load())
	})

	t.Run("cached clients still work", func(t *testing.T) {
		res := authorizePreview(t, e, userToken, params(clientID))
		assert.Equal(t, http.StatusOK, res.code, string(res.body))
	})

	t.Run("budget is per user", func(t *testing.T) {
		res := authorizePreview(t, e, humaTokenFor(t, &testuser2), params(host.URL+"/one-more.json"))
		assert.Equal(t, http.StatusOK, res.code, string(res.body))
	})

	t.Run("the token exchange still re-resolves the client", func(t *testing.T) {
		require.NoError(t, keyvalue.DelPrefix("oauth_client_metadata_"))
		before := requests.Load()
		tok := scopedExchange(t, e, "v2", authorized.Code, url.Values{
			"client_id":    {clientID},
			"redirect_uri": {metadataClientRedirect},
		})
		require.Equal(t, http.StatusOK, tok.code, string(tok.body))
		assert.Equal(t, before+1, requests.Load())
	})
}

func TestOAuthMetadataDocumentWithoutPublicURL(t *testing.T) {
	e, host, requests := countingMetadataEnv(t)
	previous := config.ServicePublicURL.GetString()
	config.ServicePublicURL.Set("")
	t.Cleanup(func() { config.ServicePublicURL.Set(previous) })

	res := authorizePreview(t, e, humaTokenFor(t, &testuser1), map[string]string{
		"client_id":    host.URL + "/client.json",
		"redirect_uri": metadataClientRedirect,
	})
	require.Equal(t, http.StatusBadRequest, res.code, string(res.body))
	assert.Equal(t, models.ErrCodeOAuthInvalidTarget, res.errorCode(t))
	assert.Equal(t, int32(0), requests.Load())
}
