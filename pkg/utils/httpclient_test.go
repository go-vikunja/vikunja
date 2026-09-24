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

package utils

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"

	"code.dny.dev/ssrf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSSRFSafeHTTPClient(t *testing.T) {
	t.Run("returns a non-nil client", func(t *testing.T) {
		client := NewSSRFSafeHTTPClient()
		assert.NotNil(t, client)
	})

	t.Run("can reach a routable test server", func(t *testing.T) {
		config.OutgoingRequestsAllowNonRoutableIPs.Set("true")
		defer config.OutgoingRequestsAllowNonRoutableIPs.Set("false")

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewSSRFSafeHTTPClient()
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
		require.NoError(t, err)
		resp, err := client.Do(req) //nolint:gosec // testing SSRF-safe client
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("blocks non-routable IPs when config is false", func(t *testing.T) {
		config.OutgoingRequestsAllowNonRoutableIPs.Set("false")
		client := NewSSRFSafeHTTPClient()

		// Attempt to connect to localhost (non-routable)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://127.0.0.1:1/test", nil)
		require.NoError(t, err)
		_, err = client.Do(req) //nolint:bodyclose,gosec // testing SSRF-safe client
		require.Error(t, err)
	})

	t.Run("has default timeout from config", func(t *testing.T) {
		config.OutgoingRequestsTimeoutSeconds.Set("30")
		client := NewSSRFSafeHTTPClient()
		assert.Equal(t, 30*time.Second, client.Timeout)
	})

	t.Run("respects custom timeout config", func(t *testing.T) {
		config.OutgoingRequestsTimeoutSeconds.Set("15")
		defer config.OutgoingRequestsTimeoutSeconds.Set("30")

		client := NewSSRFSafeHTTPClient()
		assert.Equal(t, 15*time.Second, client.Timeout)
	})

	t.Run("timeout fires on slow server", func(t *testing.T) {
		config.OutgoingRequestsAllowNonRoutableIPs.Set("true")
		config.OutgoingRequestsTimeoutSeconds.Set("1")
		defer config.OutgoingRequestsAllowNonRoutableIPs.Set("false")
		defer config.OutgoingRequestsTimeoutSeconds.Set("30")

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(3 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewSSRFSafeHTTPClient()
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
		require.NoError(t, err)
		_, err = client.Do(req) //nolint:bodyclose,gosec
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Client.Timeout")
	})

	t.Run("allows non-routable IPs when config is true", func(t *testing.T) {
		config.OutgoingRequestsAllowNonRoutableIPs.Set("true")
		defer config.OutgoingRequestsAllowNonRoutableIPs.Set("false")

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewSSRFSafeHTTPClient()
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
		require.NoError(t, err)
		resp, err := client.Do(req) //nolint:gosec // testing SSRF-safe client
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

const redirectToProxyPath = "/redirect-to-proxy"

type fakeProxy struct {
	*httptest.Server
	hits      atomic.Int32
	proxyAuth atomic.Value
}

// newFakeProxy answers every request itself, so a hit proves the client routed through it.
func newFakeProxy(t *testing.T) *fakeProxy {
	t.Helper()
	p := &fakeProxy{}
	p.proxyAuth.Store("")
	p.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.hits.Add(1)
		p.proxyAuth.Store(r.Header.Get("Proxy-Authorization"))
		if r.URL.Path == redirectToProxyPath {
			http.Redirect(w, r, p.URL, http.StatusTemporaryRedirect)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(p.Close)
	return p
}

func setProxyConfig(t *testing.T, proxyURL, password string) {
	t.Helper()
	config.OutgoingRequestsProxyURL.Set(proxyURL)
	config.OutgoingRequestsProxyPassword.Set(password)
	t.Cleanup(func() {
		config.OutgoingRequestsProxyURL.Set("")
		config.OutgoingRequestsProxyPassword.Set("")
	})
}

// setProxyEnv sets both cases because httpproxy prefers the lowercase variables.
func setProxyEnv(t *testing.T, httpProxy string) {
	t.Helper()
	for name, value := range map[string]string{
		"HTTP_PROXY":  httpProxy,
		"HTTPS_PROXY": "",
		"NO_PROXY":    "",
	} {
		t.Setenv(name, value)
		t.Setenv(strings.ToLower(name), value)
	}
}

func get(t *testing.T, client *http.Client, target string) error {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, target, nil)
	require.NoError(t, err)
	resp, err := client.Do(req) //nolint:gosec // testing outgoing clients
	if err == nil {
		resp.Body.Close()
	}
	return err
}

func TestNewSSRFSafeHTTPClientProxy(t *testing.T) {
	config.OutgoingRequestsAllowNonRoutableIPs.Set("false")
	const proxiedTarget = "http://vikunja-proxy-test.invalid/"

	t.Run("uses HTTP_PROXY from the environment", func(t *testing.T) {
		proxy := newFakeProxy(t)
		t.Setenv("HTTP_PROXY", proxy.URL)
		t.Setenv("NO_PROXY", "")

		require.NoError(t, get(t, NewSSRFSafeHTTPClient(), proxiedTarget))
		assert.Equal(t, int32(1), proxy.hits.Load())
	})

	t.Run("uses the configured proxy on a non-routable address", func(t *testing.T) {
		proxy := newFakeProxy(t)
		setProxyConfig(t, proxy.URL, "")

		require.NoError(t, get(t, NewSSRFSafeHTTPClient(), proxiedTarget))
		assert.Equal(t, int32(1), proxy.hits.Load())
		assert.Empty(t, proxy.proxyAuth.Load())
	})

	t.Run("prefers the configured proxy over the environment", func(t *testing.T) {
		envProxy := newFakeProxy(t)
		proxy := newFakeProxy(t)
		t.Setenv("HTTP_PROXY", envProxy.URL)
		setProxyConfig(t, proxy.URL, "")

		require.NoError(t, get(t, NewSSRFSafeHTTPClient(), proxiedTarget))
		assert.Equal(t, int32(1), proxy.hits.Load())
		assert.Equal(t, int32(0), envProxy.hits.Load())
	})

	t.Run("authenticates as vikunja with the proxy password", func(t *testing.T) {
		proxy := newFakeProxy(t)
		setProxyConfig(t, proxy.URL, "secret")

		require.NoError(t, get(t, NewSSRFSafeHTTPClient(), proxiedTarget))
		assert.Equal(t, "Basic "+base64.StdEncoding.EncodeToString([]byte("vikunja:secret")), proxy.proxyAuth.Load())
	})

	t.Run("authenticates with credentials from the proxy url", func(t *testing.T) {
		proxy := newFakeProxy(t)
		proxyURL, err := url.Parse(proxy.URL)
		require.NoError(t, err)
		proxyURL.User = url.UserPassword("alice", "hunter2")
		setProxyConfig(t, proxyURL.String(), "")

		require.NoError(t, get(t, NewSSRFSafeHTTPClient(), proxiedTarget))
		assert.Equal(t, "Basic "+base64.StdEncoding.EncodeToString([]byte("alice:hunter2")), proxy.proxyAuth.Load())
	})

	t.Run("fails closed on an invalid proxy url", func(t *testing.T) {
		config.OutgoingRequestsAllowNonRoutableIPs.Set("true")
		defer config.OutgoingRequestsAllowNonRoutableIPs.Set("false")
		target := newFakeProxy(t)
		setProxyConfig(t, "not a url", "")

		require.Error(t, get(t, NewSSRFSafeHTTPClient(), target.URL))
		assert.Equal(t, int32(0), target.hits.Load())
	})

	t.Run("still blocks direct non-routable targets when a proxy is set", func(t *testing.T) {
		proxy := newFakeProxy(t)
		target := newFakeProxy(t)
		t.Setenv("HTTP_PROXY", proxy.URL)

		// Loopback targets bypass the env proxy, so this dial is direct.
		require.Error(t, get(t, NewSSRFSafeHTTPClient(), target.URL))
		assert.Equal(t, int32(0), target.hits.Load())
	})

	t.Run("uses a scheme-less HTTP_PROXY on a non-routable address", func(t *testing.T) {
		proxy := newFakeProxy(t)
		setProxyEnv(t, strings.TrimPrefix(proxy.URL, "http://"))

		require.NoError(t, get(t, NewSSRFSafeHTTPClient(), proxiedTarget))
		assert.Equal(t, int32(1), proxy.hits.Load())
	})

	t.Run("blocks direct requests to the proxy address", func(t *testing.T) {
		proxy := newFakeProxy(t)
		setProxyEnv(t, proxy.URL)
		client := NewSSRFSafeHTTPClient()

		require.NoError(t, get(t, client, proxiedTarget))
		require.ErrorIs(t, get(t, client, proxy.URL), ssrf.ErrProhibitedIP)
		assert.Equal(t, int32(1), proxy.hits.Load())
	})

	t.Run("blocks redirects to the proxy address", func(t *testing.T) {
		proxy := newFakeProxy(t)
		setProxyEnv(t, proxy.URL)

		require.ErrorIs(t, get(t, NewSSRFSafeHTTPClient(), "http://vikunja-proxy-test.invalid"+redirectToProxyPath), ssrf.ErrProhibitedIP)
		assert.Equal(t, int32(1), proxy.hits.Load())
	})

	t.Run("blocks direct requests to the proxy address for an unproxied scheme", func(t *testing.T) {
		proxy := newFakeProxy(t)
		setProxyEnv(t, proxy.URL)
		client := NewSSRFSafeHTTPClient()
		proxyURL, err := url.Parse(proxy.URL)
		require.NoError(t, err)

		require.NoError(t, get(t, client, proxiedTarget))
		require.ErrorIs(t, get(t, client, "https://"+proxyURL.Host+"/"), ssrf.ErrProhibitedIP)
		assert.Equal(t, int32(1), proxy.hits.Load())
	})
}

func TestProxyDialAddr(t *testing.T) {
	for raw, want := range map[string]string{
		"http://proxy":               "proxy:80",
		"https://proxy":              "proxy:443",
		"socks5://proxy":             "proxy:1080",
		"http://proxy:3128":          "proxy:3128",
		"http://[2001:db8::1]":       "[2001:db8::1]:80",
		"https://[2001:db8::1]:8443": "[2001:db8::1]:8443",
		"http://bücher.example":      "xn--bcher-kva.example:80",
	} {
		t.Run(raw, func(t *testing.T) {
			u, err := url.Parse(raw)
			require.NoError(t, err)
			assert.Equal(t, want, proxyDialAddr(u))
		})
	}
}

func TestNewHTTPClient(t *testing.T) {
	t.Run("reaches non-routable targets", func(t *testing.T) {
		config.OutgoingRequestsAllowNonRoutableIPs.Set("false")
		target := newFakeProxy(t)

		require.NoError(t, get(t, NewHTTPClient(), target.URL))
		assert.Equal(t, int32(1), target.hits.Load())
	})

	t.Run("uses the configured proxy", func(t *testing.T) {
		proxy := newFakeProxy(t)
		setProxyConfig(t, proxy.URL, "")

		require.NoError(t, get(t, NewHTTPClient(), "http://vikunja-proxy-test.invalid/"))
		assert.Equal(t, int32(1), proxy.hits.Load())
	})

	t.Run("has timeout from config", func(t *testing.T) {
		config.OutgoingRequestsTimeoutSeconds.Set("15")
		defer config.OutgoingRequestsTimeoutSeconds.Set("30")

		assert.Equal(t, 15*time.Second, NewHTTPClient().Timeout)
	})
}
