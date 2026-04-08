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
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/version"

	"code.dny.dev/ssrf"
)

// NewSSRFSafeHTTPClient returns an *http.Client with SSRF protection applied.
// It blocks connections to non-globally-routable IP addresses (loopback,
// private ranges, link-local, etc.) unless outgoingrequests.allownonroutableips
// is set to true. It also configures proxy settings from outgoingrequests config.
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
