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

package oauth2server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testResource struct{}

func (testResource) ResourceURL() string { return "https://vikunja.example/api/v2/mcp" }

func (testResource) ScopePermissions(Scope) models.APIPermissions { return nil }

// metadataHost serves documents over TLS; tests reach it through its own trusted client.
type metadataHost struct {
	server   *httptest.Server
	requests atomic.Int32
}

func newMetadataHost(t *testing.T, handler func(host *metadataHost, w http.ResponseWriter, r *http.Request)) *metadataHost {
	t.Helper()
	host := &metadataHost{}
	host.server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host.requests.Add(1)
		handler(host, w, r)
	}))
	t.Cleanup(host.server.Close)
	// A later host can reuse the port, and so the cached client id.
	t.Cleanup(func() { _ = keyvalue.DelPrefix("oauth_client_metadata_") })
	return host
}

func (h *metadataHost) clientID() string { return h.server.URL + "/oauth/client.json" }

func (h *metadataHost) oauthServer() *Server {
	return NewServer(testResource{}, NewMetadataFetcher(h.server.Client()))
}

func (h *metadataHost) document(overrides map[string]any) map[string]any {
	doc := map[string]any{
		"client_id":   h.clientID(),
		"client_name": "Example MCP Client",
		"redirect_uris": []string{
			"http://127.0.0.1:3000/callback",
			"https://app.example.com/callback",
		},
		"grant_types":                []string{"authorization_code"},
		"response_types":             []string{"code"},
		"token_endpoint_auth_method": "none",
	}
	for k, v := range overrides {
		if v == nil {
			delete(doc, k)
			continue
		}
		doc[k] = v
	}
	return doc
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(w).Encode(v))
}

// A nil override removes the field.
func serveDocument(t *testing.T, overrides map[string]any) *metadataHost {
	t.Helper()
	return newMetadataHost(t, func(host *metadataHost, w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, host.document(overrides))
	})
}

func resolveMetadataClient(t *testing.T, host *metadataHost) MetadataClient {
	t.Helper()
	client, err := host.oauthServer().resolveClient(t.Context(), host.clientID(), unmeteredFetch)
	require.NoError(t, err)
	metadataClient, ok := client.(MetadataClient)
	require.True(t, ok, "%T", client)
	return metadataClient
}

func requireInvalidMetadata(t *testing.T, host *metadataHost, reason string) {
	t.Helper()
	_, err := host.oauthServer().resolveClient(t.Context(), host.clientID(), unmeteredFetch)
	var invalid *models.ErrOAuthInvalidClientMetadata
	require.ErrorAs(t, err, &invalid)
	assert.Contains(t, invalid.Reason, reason)
}

func requireUnavailable(t *testing.T, host *metadataHost) {
	t.Helper()
	_, err := host.oauthServer().resolveClient(t.Context(), host.clientID(), unmeteredFetch)
	assert.True(t, models.IsErrOAuthClientMetadataUnavailable(err), "%v", err)
}

func TestResolveMetadataClient(t *testing.T) {
	t.Run("resolves a valid document", func(t *testing.T) {
		var accept string
		host := newMetadataHost(t, func(host *metadataHost, w http.ResponseWriter, r *http.Request) {
			accept = r.Header.Get("Accept")
			writeJSON(t, w, host.document(nil))
		})
		client := resolveMetadataClient(t, host)
		assert.Equal(t, "application/json", accept)
		assert.Equal(t, host.clientID(), client.ID())
		assert.Equal(t, "Example MCP Client", client.tokenTitle())
		assert.True(t, client.scopedOnly())
		assert.Equal(t, AuthorizeClient{
			ID:   host.clientID(),
			Kind: ClientKindMetadataDocument,
			Name: "Example MCP Client",
			Host: strings.TrimPrefix(host.server.URL, "https://"),
		}, client.describe())
	})

	t.Run("accepts minimal documents and json media types", func(t *testing.T) {
		host := newMetadataHost(t, func(host *metadataHost, w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/client-metadata+json; charset=utf-8")
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
				"client_id":     host.clientID(),
				"client_name":   "Minimal",
				"redirect_uris": []string{"http://localhost/cb"},
			}))
		})
		resolveMetadataClient(t, host)
	})

	t.Run("caches the document", func(t *testing.T) {
		host := serveDocument(t, nil)
		srv := host.oauthServer()
		for range 3 {
			_, err := srv.resolveClient(t.Context(), host.clientID(), unmeteredFetch)
			require.NoError(t, err)
		}
		assert.Equal(t, int32(1), host.requests.Load())
	})

	t.Run("does not cache failures", func(t *testing.T) {
		host := newMetadataHost(t, func(host *metadataHost, w http.ResponseWriter, _ *http.Request) {
			if host.requests.Load() == 1 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			writeJSON(t, w, host.document(nil))
		})
		requireUnavailable(t, host)
		resolveMetadataClient(t, host)
		assert.Equal(t, int32(2), host.requests.Load())
	})

	t.Run("does not follow redirects", func(t *testing.T) {
		var followed atomic.Bool
		host := newMetadataHost(t, func(host *metadataHost, w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/elsewhere.json" {
				followed.Store(true)
				writeJSON(t, w, host.document(nil))
				return
			}
			http.Redirect(w, r, "/elsewhere.json", http.StatusFound)
		})
		requireUnavailable(t, host)
		assert.False(t, followed.Load())
	})

	t.Run("rejects a non-200 response", func(t *testing.T) {
		host := newMetadataHost(t, func(host *metadataHost, w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(host.document(nil))
		})
		requireUnavailable(t, host)
	})

	t.Run("rejects a document that is not served as JSON", func(t *testing.T) {
		for _, contentType := range []string{
			"text/html",
			"text/plain",
			"",
			"application/jsonp",
		} {
			host := newMetadataHost(t, func(host *metadataHost, w http.ResponseWriter, _ *http.Request) {
				raw, err := json.Marshal(host.document(nil))
				require.NoError(t, err)
				w.Header().Set("Content-Type", contentType)
				_, _ = w.Write(raw)
			})
			requireInvalidMetadata(t, host, "not served as JSON")
		}
	})

	t.Run("rejects a document over 5 KiB", func(t *testing.T) {
		host := serveDocument(t, map[string]any{"padding": strings.Repeat("x", maxMetadataDocumentBytes)})
		requireInvalidMetadata(t, host, "larger than 5 KiB")
	})

	t.Run("accepts a document just under 5 KiB", func(t *testing.T) {
		host := newMetadataHost(t, func(host *metadataHost, w http.ResponseWriter, _ *http.Request) {
			doc := host.document(nil)
			raw, err := json.Marshal(doc)
			require.NoError(t, err)
			doc["padding"] = strings.Repeat("x", maxMetadataDocumentBytes-len(raw)-len(`,"padding":""`))
			raw, err = json.Marshal(doc)
			require.NoError(t, err)
			require.Len(t, raw, maxMetadataDocumentBytes)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(raw)
		})
		resolveMetadataClient(t, host)
	})

	t.Run("rejects malformed JSON", func(t *testing.T) {
		for _, body := range []string{
			"{",
			"[]",
			`"https://app.example.com"`,
			`{"client_id": 1}`,
			`{"redirect_uris": "http://127.0.0.1/cb"}`,
		} {
			host := newMetadataHost(t, func(_ *metadataHost, w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(body))
			})
			requireInvalidMetadata(t, host, "not a JSON object")
		}
	})

	t.Run("rejects a client_id mismatch", func(t *testing.T) {
		host := newMetadataHost(t, func(host *metadataHost, w http.ResponseWriter, _ *http.Request) {
			writeJSON(t, w, host.document(map[string]any{"client_id": host.clientID() + "/"}))
		})
		requireInvalidMetadata(t, host, "does not match")

		host = serveDocument(t, map[string]any{"client_id": nil})
		requireInvalidMetadata(t, host, "does not match")
	})

	t.Run("requires client_name and redirect_uris", func(t *testing.T) {
		for name, overrides := range map[string]map[string]any{
			"client_name is missing":   {"client_name": nil},
			"client_name is blank":     {"client_name": " \u200e\t"},
			"redirect_uris is missing": {"redirect_uris": nil},
			"redirect_uris is empty":   {"redirect_uris": []string{}},
		} {
			host := serveDocument(t, overrides)
			_, err := host.oauthServer().resolveClient(t.Context(), host.clientID(), unmeteredFetch)
			assert.True(t, models.IsErrOAuthInvalidClientMetadata(err), "%s: %v", name, err)
		}
	})

	t.Run("rejects confidential clients", func(t *testing.T) {
		for name, overrides := range map[string]map[string]any{
			"client_secret":            {"client_secret": "s3cret"},
			"client_secret_expires_at": {"client_secret_expires_at": 0},
			"secret auth method":       {"token_endpoint_auth_method": "client_secret_basic"},
			"key auth method":          {"token_endpoint_auth_method": "private_key_jwt"},
			"supported without none": {
				"token_endpoint_auth_method":            nil,
				"token_endpoint_auth_methods_supported": []string{"private_key_jwt"},
			},
			"supported with a secret method": {
				"token_endpoint_auth_method":            nil,
				"token_endpoint_auth_methods_supported": []string{"none", "client_secret_post"},
			},
			"single method not in supported": {
				"token_endpoint_auth_method":            "private_key_jwt",
				"token_endpoint_auth_methods_supported": []string{"none"},
			},
		} {
			host := serveDocument(t, overrides)
			_, err := host.oauthServer().resolveClient(t.Context(), host.clientID(), unmeteredFetch)
			assert.True(t, models.IsErrOAuthInvalidClientMetadata(err), "%s: %v", name, err)
		}
	})

	t.Run("accepts public clients", func(t *testing.T) {
		for _, overrides := range []map[string]any{
			{"token_endpoint_auth_method": nil},
			{
				"token_endpoint_auth_method":            nil,
				"token_endpoint_auth_methods_supported": []string{"private_key_jwt", "none"},
			},
			{
				"token_endpoint_auth_method":            "none",
				"token_endpoint_auth_methods_supported": []string{"none"},
			},
		} {
			resolveMetadataClient(t, serveDocument(t, overrides))
		}
	})

	t.Run("rejects unsupported grant and response types", func(t *testing.T) {
		requireInvalidMetadata(t, serveDocument(t, map[string]any{"grant_types": []string{"client_credentials"}}), "grant_types")
		requireInvalidMetadata(t, serveDocument(t, map[string]any{"grant_types": []string{}}), "grant_types")
		requireInvalidMetadata(t, serveDocument(t, map[string]any{"response_types": []string{"token"}}), "response_types")
		resolveMetadataClient(t, serveDocument(t, map[string]any{"grant_types": []string{"refresh_token", "authorization_code"}}))
	})

	t.Run("rejects jwks together with jwks_uri", func(t *testing.T) {
		requireInvalidMetadata(t, serveDocument(t, map[string]any{
			"jwks":     map[string]any{"keys": []any{}},
			"jwks_uri": "https://app.example.com/jwks.json",
		}), "jwks")
	})

	t.Run("bounds and cleans the client name", func(t *testing.T) {
		client := resolveMetadataClient(t, serveDocument(t, map[string]any{"client_name": "  Evil\u202e moc.elgoog\n" + strings.Repeat("é", 200)}))
		name := client.describe().Name
		assert.NotContains(t, name, "\u202e")
		assert.NotContains(t, name, "\n")
		assert.True(t, strings.HasPrefix(name, "Evil moc.elgoog é"), name)
		assert.Len(t, []rune(name), maxClientNameRunes)
	})

	t.Run("the production client refuses non-routable hosts", func(t *testing.T) {
		config.OutgoingRequestsAllowNonRoutableIPs.Set(false)
		host := serveDocument(t, nil)
		srv := NewServer(testResource{}, NewMetadataFetcher(utils.NewSSRFSafeHTTPClient()))
		_, err := srv.resolveClient(t.Context(), host.clientID(), unmeteredFetch)
		assert.True(t, models.IsErrOAuthClientMetadataUnavailable(err), "%v", err)
		assert.Equal(t, int32(0), host.requests.Load())
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestResolveServerIdentifierAsClient(t *testing.T) {
	publicURL := config.ServicePublicURL.GetString()
	config.ServicePublicURL.Set("https://vikunja.example/sub/")
	t.Cleanup(func() { config.ServicePublicURL.Set(publicURL) })

	var fetched []string
	srv := NewServer(testResource{}, NewMetadataFetcher(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		fetched = append(fetched, r.URL.String())
		return nil, errors.New("offline")
	})}))

	_, err := srv.resolveClient(t.Context(), "https://vikunja.example/sub", unmeteredFetch)
	assert.True(t, models.IsErrOAuthClientIDIsServer(err), "%v", err)
	assert.Empty(t, fetched)

	// Simple string comparison: anything else is an ordinary client id URL.
	_, err = srv.resolveClient(t.Context(), "https://vikunja.example/sub/", unmeteredFetch)
	assert.True(t, models.IsErrOAuthClientMetadataUnavailable(err), "%v", err)
	assert.Equal(t, []string{"https://vikunja.example/sub/"}, fetched)
}

func TestMetadataCacheTTL(t *testing.T) {
	for header, want := range map[string]time.Duration{
		"":                          defaultMetadataCacheTTL,
		"public, max-age=7200":      2 * time.Hour,
		`max-age="1800"`:            30 * time.Minute,
		"max-age=60":                minMetadataCacheTTL,
		"max-age=99999999999999999": maxMetadataCacheTTL,
		"max-age=-5":                defaultMetadataCacheTTL,
		"max-age=abc":               defaultMetadataCacheTTL,
		"no-store":                  minMetadataCacheTTL,
		"max-age=7200, no-cache":    minMetadataCacheTTL,
	} {
		assert.Equal(t, want, metadataCacheTTL(header), header)
	}
}

func TestPreviewMetadataClient(t *testing.T) {
	host := serveDocument(t, nil)
	srv := host.oauthServer()
	request := func(overrides map[string]string) *AuthorizeRequest {
		req := &AuthorizeRequest{
			ResponseType:        "code",
			ClientID:            host.clientID(),
			RedirectURI:         "http://127.0.0.1:41234/callback",
			CodeChallenge:       "challenge",
			CodeChallengeMethod: "S256",
		}
		if v, ok := overrides["scope"]; ok {
			req.Scope = v
		}
		return req
	}

	t.Run("keeps a requested read scope", func(t *testing.T) {
		preview, err := srv.Preview(t.Context(), request(map[string]string{"scope": "mcp:read"}), 1)
		require.NoError(t, err)
		assert.Equal(t, []Scope{ScopeMCPRead}, preview.Scopes)
	})

	t.Run("does not fetch for a request failing cheap checks", func(t *testing.T) {
		fresh := serveDocument(t, nil)
		req := request(nil)
		req.ClientID = fresh.clientID()
		req.CodeChallengeMethod = "plain"
		_, err := fresh.oauthServer().Preview(t.Context(), req, 1)
		assert.True(t, models.IsErrOAuthMissingPKCE(err), "%v", err)
		assert.Equal(t, int32(0), fresh.requests.Load())
	})
}
