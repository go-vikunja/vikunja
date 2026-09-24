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
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/version"

	"code.dny.dev/ssrf"
	"golang.org/x/net/http/httpproxy"
	"golang.org/x/net/idna"
)

// NewHTTPClient returns a proxy-aware client without the SSRF guard, for admin-configured endpoints.
// Use NewSSRFSafeHTTPClient when users control the target url.
func NewHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = configuredProxy()
	transport.ProxyConnectHeader = http.Header{"User-Agent": []string{"Vikunja/" + version.Version}}

	return &http.Client{
		Timeout:   time.Duration(config.OutgoingRequestsTimeoutSeconds.GetInt()) * time.Second,
		Transport: transport,
	}
}

// NewSSRFSafeHTTPClient blocks non-globally-routable targets unless outgoingrequests.allownonroutableips is set.
//
// Deprecated webhooks.* config keys are migrated to outgoingrequests.* at
// config init time (see config.InitDefaultConfig), so this function only
// reads the new keys.
func NewSSRFSafeHTTPClient() *http.Client {
	client := NewHTTPClient()
	if !config.OutgoingRequestsAllowNonRoutableIPs.GetBool() {
		guardProxiedDials(client.Transport.(*http.Transport), proxyDialAddrs())
	}
	return client
}

func configuredProxy() func(*http.Request) (*url.URL, error) {
	raw := config.OutgoingRequestsProxyURL.GetString()
	if raw == "" {
		// Not http.ProxyFromEnvironment: it caches the env on first use.
		proxyFunc := httpproxy.FromEnvironment().ProxyFunc()
		return func(req *http.Request) (*url.URL, error) {
			return proxyFunc(req.URL)
		}
	}

	proxyURL, err := url.Parse(raw)
	if err != nil || proxyURL.Host == "" {
		// The raw value may contain credentials, keep it out of the error.
		invalid := fmt.Errorf("invalid %s, expected a url like http://host:port", config.OutgoingRequestsProxyURL)
		return func(*http.Request) (*url.URL, error) {
			return nil, invalid
		}
	}

	if password := config.OutgoingRequestsProxyPassword.GetString(); password != "" {
		if proxyURL.User == nil {
			proxyURL.User = url.UserPassword("vikunja", password)
		} else if _, hasPassword := proxyURL.User.Password(); !hasPassword {
			proxyURL.User = url.UserPassword(proxyURL.User.Username(), password)
		}
	}
	return http.ProxyURL(proxyURL)
}

// The proxy dial is exempt: it is admin-chosen, often private, and resolves proxied targets itself.
func guardProxiedDials(transport *http.Transport, proxyAddrs map[string]struct{}) {
	proxy := transport.Proxy
	transport.Proxy = func(req *http.Request) (*url.URL, error) {
		u, err := proxy(req)
		if err != nil || u != nil {
			return u, err
		}
		if _, isProxy := proxyAddrs[proxyDialAddr(req.URL)]; isProxy {
			return nil, fmt.Errorf("direct request to the outgoing proxy: %w", ssrf.ErrProhibitedIP)
		}
		return nil, nil
	}

	dialer := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
	guarded := &net.Dialer{
		Timeout:   dialer.Timeout,
		KeepAlive: dialer.KeepAlive,
		Control:   ssrf.New(ssrf.WithAnyPort()).Safe,
	}
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		if _, isProxy := proxyAddrs[addr]; isProxy {
			return dialer.DialContext(ctx, network, addr)
		}
		return guarded.DialContext(ctx, network, addr)
	}
}

func proxyDialAddrs() map[string]struct{} {
	var proxies []*url.URL
	if raw := config.OutgoingRequestsProxyURL.GetString(); raw != "" {
		if u, err := url.Parse(raw); err == nil && u.Host != "" {
			proxies = append(proxies, u)
		}
	} else {
		env := *httpproxy.FromEnvironment()
		// NO_PROXY must not hide a proxy that other hosts still use.
		env.NoProxy = ""
		proxyFunc := env.ProxyFunc()
		for _, scheme := range []string{"http", "https"} {
			if u, err := proxyFunc(&url.URL{Scheme: scheme, Host: "probe.invalid"}); err == nil && u != nil {
				proxies = append(proxies, u)
			}
		}
	}

	addrs := make(map[string]struct{}, len(proxies))
	for _, u := range proxies {
		addrs[proxyDialAddr(u)] = struct{}{}
	}
	return addrs
}

// proxyDialAddr mirrors the address net/http dials for a url.
func proxyDialAddr(u *url.URL) string {
	host := u.Hostname()
	if strings.IndexFunc(host, func(r rune) bool { return r >= utf8.RuneSelf }) >= 0 {
		if ascii, err := idna.Lookup.ToASCII(host); err == nil {
			host = ascii
		}
	}
	port := u.Port()
	if port == "" {
		port = map[string]string{"http": "80", "https": "443", "socks5": "1080", "socks5h": "1080"}[u.Scheme]
	}
	return net.JoinHostPort(host, port)
}
