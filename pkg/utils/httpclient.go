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
	"encoding/base64"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/version"

	"code.dny.dev/ssrf"
)

// warnProxyDefeatsSSRFGuardOnce fires at most once: once transport.Proxy is
// set below, the dialer's Control hook (the SSRF guard) only ever sees the
// proxy's own dial - the proxy resolves and connects to the real destination
// itself, entirely outside Go's http.Transport, so the guard can't evaluate
// it. A proxy that isn't itself SSRF-aware (e.g. a generic corporate egress
// proxy, as opposed to something like https://github.com/frain-dev/mole)
// replaces the destination-level check below, it doesn't supplement it.
var warnProxyDefeatsSSRFGuardOnce sync.Once

// NewSSRFSafeHTTPClient returns an *http.Client with SSRF protection applied.
// It blocks connections to non-globally-routable IP addresses (loopback,
// private ranges, link-local, etc.) unless outgoingrequests.allownonroutableips
// is set to true. It also configures proxy settings from outgoingrequests config.
//
// If outgoingrequests.proxyurl is also set, that REPLACES this destination-level
// check rather than supplementing it: once a proxy is configured, the dialer's
// Control hook only ever sees the proxy's own address, never the real
// destination, since the proxy resolves and connects to it independently of
// Go's http.Transport. The configured proxy itself must enforce SSRF policy.
//
// Deprecated webhooks.* config keys are migrated to outgoingrequests.* at
// config init time (see config.InitDefaultConfig), so this function only
// reads the new keys.
func NewSSRFSafeHTTPClient() *http.Client {
	client := &http.Client{
		Timeout: time.Duration(config.OutgoingRequestsTimeoutSeconds.GetInt()) * time.Second,
	}
	transport := &http.Transport{}

	if !config.OutgoingRequestsAllowNonRoutableIPs.GetBool() {
		guardian := ssrf.New(ssrf.WithAnyPort())
		transport.DialContext = (&net.Dialer{
			Control: guardian.Safe,
		}).DialContext
	}

	proxyURL := config.OutgoingRequestsProxyURL.GetString()
	proxyPassword := config.OutgoingRequestsProxyPassword.GetString()

	if proxyURL != "" && proxyPassword != "" {
		if !config.OutgoingRequestsAllowNonRoutableIPs.GetBool() {
			warnProxyDefeatsSSRFGuardOnce.Do(func() {
				log.Warningf("outgoingrequests.proxyurl is configured: this REPLACES, not supplements, the non-routable-IP SSRF guard for webhooks/migrations/avatar downloads - the configured proxy itself must enforce SSRF policy (e.g. a properly configured mole instance, https://github.com/frain-dev/mole), or private/internal destinations will reach it unfiltered")
			})
		}
		parsedURL, _ := url.Parse(proxyURL)
		transport.Proxy = http.ProxyURL(parsedURL)
		transport.ProxyConnectHeader = http.Header{
			"Proxy-Authorization": []string{"Basic " + base64.StdEncoding.EncodeToString([]byte("vikunja:"+proxyPassword))},
			"User-Agent":          []string{"Vikunja/" + version.Version},
		}
	}

	client.Transport = transport
	return client
}
