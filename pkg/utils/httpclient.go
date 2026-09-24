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
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/version"

	"code.dny.dev/ssrf"
	"golang.org/x/net/http/httpproxy"
)

// NewSSRFSafeHTTPClient returns an *http.Client with SSRF protection applied.
// It blocks connections to non-globally-routable IP addresses (loopback,
// private ranges, link-local, etc.) unless outgoingrequests.allownonroutableips
// is set to true. It routes requests through the proxy from outgoingrequests
// config, falling back to the HTTP_PROXY, HTTPS_PROXY and NO_PROXY env vars.
//
// Deprecated webhooks.* config keys are migrated to outgoingrequests.* at
// config init time (see config.InitDefaultConfig), so this function only
// reads the new keys.
func NewSSRFSafeHTTPClient() *http.Client {
	client := &http.Client{
		Timeout: time.Duration(config.OutgoingRequestsTimeoutSeconds.GetInt()) * time.Second,
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	// Not http.ProxyFromEnvironment: it caches the env on first use.
	transport.Proxy = proxyFromEnv(httpproxy.FromEnvironment().ProxyFunc())

	proxyURL := config.OutgoingRequestsProxyURL.GetString()
	proxyPassword := config.OutgoingRequestsProxyPassword.GetString()

	if proxyURL != "" && proxyPassword != "" {
		parsedURL, _ := url.Parse(proxyURL)
		transport.Proxy = http.ProxyURL(parsedURL)
		transport.ProxyConnectHeader = http.Header{
			"Proxy-Authorization": []string{"Basic " + base64.StdEncoding.EncodeToString([]byte("vikunja:"+proxyPassword))},
			"User-Agent":          []string{"Vikunja/" + version.Version},
		}
	}

	if !config.OutgoingRequestsAllowNonRoutableIPs.GetBool() {
		guardProxiedDials(transport)
	}

	client.Transport = transport
	return client
}

func proxyFromEnv(proxyFunc func(*url.URL) (*url.URL, error)) func(*http.Request) (*url.URL, error) {
	return func(req *http.Request) (*url.URL, error) {
		return proxyFunc(req.URL)
	}
}

// guardProxiedDials applies the SSRF guard to every dial except the one to the
// admin-configured proxy, which usually lives on a private network. Proxied
// targets are resolved by the proxy, so filtering them is the proxy's job.
func guardProxiedDials(transport *http.Transport) {
	var proxyAddrs sync.Map
	proxy := transport.Proxy
	transport.Proxy = func(req *http.Request) (*url.URL, error) {
		u, err := proxy(req)
		if u != nil {
			proxyAddrs.Store(proxyDialAddr(u), struct{}{})
		}
		return u, err
	}

	dialer := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
	guarded := &net.Dialer{
		Timeout:   dialer.Timeout,
		KeepAlive: dialer.KeepAlive,
		Control:   ssrf.New(ssrf.WithAnyPort()).Safe,
	}
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		if _, isProxy := proxyAddrs.Load(addr); isProxy {
			return dialer.DialContext(ctx, network, addr)
		}
		return guarded.DialContext(ctx, network, addr)
	}
}

// proxyDialAddr mirrors the address net/http dials for a proxy URL.
func proxyDialAddr(u *url.URL) string {
	port := u.Port()
	if port == "" {
		port = map[string]string{"http": "80", "https": "443", "socks5": "1080", "socks5h": "1080"}[u.Scheme]
	}
	return net.JoinHostPort(u.Hostname(), port)
}
