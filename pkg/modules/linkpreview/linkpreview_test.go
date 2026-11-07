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

package linkpreview

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"

	"code.dny.dev/ssrf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	log.InitLogger()
	keyvalue.InitStorage()
	os.Exit(m.Run())
}

func TestLinkPreviewCache(t *testing.T) {
	previous := config.OutgoingRequestsAllowNonRoutableIPs.GetBool()
	config.OutgoingRequestsAllowNonRoutableIPs.Set(true)
	t.Cleanup(func() { config.OutgoingRequestsAllowNonRoutableIPs.Set(previous) })

	t.Run("reuses metadata across fragments and returns independent values", func(t *testing.T) {
		var requests atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			requests.Add(1)
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<title>Cached article</title>`))
		}))
		defer server.Close()

		first, err := GetPreview(context.Background(), server.URL+"/article#first")
		require.NoError(t, err)
		first.Title = "Changed by caller"
		second, err := GetPreview(context.Background(), server.URL+"/article#second")
		require.NoError(t, err)
		assert.Equal(t, "Cached article", second.Title)
		assert.EqualValues(t, 1, requests.Load())

		_, err = GetPreview(context.Background(), server.URL+"/article?different=1")
		require.NoError(t, err)
		assert.EqualValues(t, 2, requests.Load())
	})

	t.Run("fetches again after expiry", func(t *testing.T) {
		keyvalue.InitStorage()
		var requests atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			requests.Add(1)
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<title>Article</title>`))
		}))
		defer server.Close()

		preview, err := GetPreview(context.Background(), server.URL)
		require.NoError(t, err)
		keys, err := keyvalue.ListKeys("link_preview_")
		require.NoError(t, err)
		require.Len(t, keys, 1)
		require.NoError(t, keyvalue.PutWithTTL(keys[0], *preview, time.Nanosecond))
		time.Sleep(time.Millisecond)

		_, err = GetPreview(context.Background(), server.URL)
		require.NoError(t, err)
		assert.EqualValues(t, 2, requests.Load())
	})

	t.Run("does not cache failures", func(t *testing.T) {
		var requests atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if requests.Add(1) == 1 {
				w.WriteHeader(http.StatusBadGateway)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<title>Recovered</title>`))
		}))
		defer server.Close()

		_, err := GetPreview(context.Background(), server.URL)
		require.Error(t, err)
		preview, err := GetPreview(context.Background(), server.URL)
		require.NoError(t, err)
		assert.Equal(t, "Recovered", preview.Title)
		assert.EqualValues(t, 2, requests.Load())
	})

	t.Run("validates cached URLs against the current public URL", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/redirect" {
				http.Redirect(w, r, "/article", http.StatusFound)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<title>Article</title>`))
		}))
		defer server.Close()
		_, err := GetPreview(context.Background(), server.URL+"/article")
		require.NoError(t, err)
		_, err = GetPreview(context.Background(), server.URL+"/redirect")
		require.NoError(t, err)

		previousURL := config.ServicePublicURL.GetString()
		config.ServicePublicURL.Set(server.URL + "/article")
		t.Cleanup(func() { config.ServicePublicURL.Set(previousURL) })
		for _, path := range []string{"/article", "/redirect"} {
			_, err = GetPreview(context.Background(), server.URL+path)
			var invalid *ErrInvalidURL
			require.ErrorAs(t, err, &invalid, path)
		}
	})
}

func TestLinkPreview(t *testing.T) {
	previous := config.OutgoingRequestsAllowNonRoutableIPs.GetBool()
	config.OutgoingRequestsAllowNonRoutableIPs.Set(true)
	t.Cleanup(func() { config.OutgoingRequestsAllowNonRoutableIPs.Set(previous) })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/redirect":
			http.Redirect(w, r, "/pages/article", http.StatusFound)
		case "/pages/article":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`<html><head>
				<title>Fallback</title><meta name="description" content="Fallback description">
				<meta property="og:title" content="Article &amp; more">
				<meta property="og:description" content="Article description">
				<meta property="og:site_name" content="Example">
				<meta property="og:url" content="https://untrusted.example/">
				<meta property="og:image" content="../image.jpg">
				<link rel="icon" href="/icon.png">
			</head></html>`))
		case "/fallback":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<title>Page title</title><meta name="description" content="Page description">`))
		case "/unsafe":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<meta property="og:image" content="javascript:alert(1)"><link rel="icon" href="data:image/png,abc">`))
		case "/large":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(strings.Repeat("a", 2*1024*1024+1)))
		case "/image":
			w.Header().Set("Content-Type", "image/png")
		case "/internal":
			http.Redirect(w, r, "https://tasks.example.com/tasks/1", http.StatusFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	t.Run("OG metadata and relative assets use redirect destination", func(t *testing.T) {
		preview, err := GetPreview(context.Background(), server.URL+"/redirect")
		require.NoError(t, err)
		assert.Equal(t, "Article & more", preview.Title)
		assert.Equal(t, "Article description", preview.Description)
		assert.Equal(t, "Example", preview.SiteName)
		assert.Equal(t, server.URL+"/pages/article", preview.URL)
		assert.Equal(t, server.URL+"/image.jpg", preview.Image)
		assert.Equal(t, server.URL+"/icon.png", preview.Favicon)
	})
	t.Run("HTML fallbacks", func(t *testing.T) {
		preview, err := GetPreview(context.Background(), server.URL+"/fallback")
		require.NoError(t, err)
		assert.Equal(t, "Page title", preview.Title)
		assert.Equal(t, "Page description", preview.Description)
	})
	t.Run("unsafe assets are omitted", func(t *testing.T) {
		preview, err := GetPreview(context.Background(), server.URL+"/unsafe")
		require.NoError(t, err)
		assert.Empty(t, preview.Image)
		assert.Empty(t, preview.Favicon)
	})
	t.Run("rejects oversized, non HTML and failed responses", func(t *testing.T) {
		for _, path := range []string{"/large", "/image", "/missing"} {
			_, err := GetPreview(context.Background(), server.URL+path)
			var unavailable *ErrPreviewUnavailable
			require.ErrorAs(t, err, &unavailable, path)
		}
	})
	t.Run("rejects redirect to this instance", func(t *testing.T) {
		previousURL := config.ServicePublicURL.GetString()
		config.ServicePublicURL.Set("https://tasks.example.com/")
		t.Cleanup(func() { config.ServicePublicURL.Set(previousURL) })
		_, err := GetPreview(context.Background(), server.URL+"/internal")
		var unavailable *ErrPreviewUnavailable
		require.ErrorAs(t, err, &unavailable)
	})
	t.Run("respects cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := GetPreview(ctx, server.URL+"/fallback")
		require.ErrorIs(t, err, context.Canceled)
	})
}

func TestLinkPreviewURL(t *testing.T) {
	previous := config.ServicePublicURL.GetString()
	config.ServicePublicURL.Set("https://tasks.example.com/app/")
	t.Cleanup(func() { config.ServicePublicURL.Set(previous) })
	for _, target := range []string{
		"https://tasks.example.com/app",
		"https://tasks.example.com/app/tasks/1",
		"https://TASKS.EXAMPLE.COM/app/tasks/1",
		"https://tasks.example.com:443/app/tasks/1",
	} {
		_, err := previewURL(target)
		var invalid *ErrInvalidURL
		require.ErrorAs(t, err, &invalid, target)
	}
	for _, target := range []string{
		"https://tasks.example.com/application/",
		"https://tasks.example.com.evil/app/tasks/1",
	} {
		_, err := previewURL(target)
		require.NoError(t, err, target)
	}
}

func TestLinkPreviewSSRF(t *testing.T) {
	previous := config.OutgoingRequestsAllowNonRoutableIPs.GetBool()
	config.OutgoingRequestsAllowNonRoutableIPs.Set(false)
	t.Cleanup(func() { config.OutgoingRequestsAllowNonRoutableIPs.Set(previous) })
	_, err := GetPreview(context.Background(), "http://127.0.0.1:1/")
	require.ErrorIs(t, err, ssrf.ErrProhibitedIP)
}
