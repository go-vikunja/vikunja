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

package routes

import (
	"net"
	"net/http"
	"strings"

	"code.vikunja.io/api/pkg/log"

	"github.com/labstack/echo/v5"
)

// A peer connected through a unix socket has no IP, so Go sets RemoteAddr to "@".
// Echo's extractors cannot parse that and give up before looking at any proxy
// header, leaving the client IP empty. Loopback is the honest stand-in: the peer
// is by definition on the same host.
const unixSocketPeerAddr = "127.0.0.1:0"

// newIPExtractor builds the extractor used to determine the client IP.
// Echo's default RealIP() trusts X-Forwarded-For and X-Real-IP unconditionally,
// which allows attackers to bypass IP-based rate limits.
// See: https://echo.labstack.com/docs/ip-address
func newIPExtractor(method, trustedProxies string) echo.IPExtractor {
	switch method {
	case "xff":
		trustOptions := parseTrustedProxies(trustedProxies)
		log.Debugf("IP extraction: X-Forwarded-For with %d trusted proxy ranges", len(trustOptions))
		return withUnixSocketPeer(echo.ExtractIPFromXFFHeader(trustOptions...))
	case "realip":
		trustOptions := parseTrustedProxies(trustedProxies)
		log.Debugf("IP extraction: X-Real-IP with %d trusted proxy ranges", len(trustOptions))
		return withUnixSocketPeer(echo.ExtractIPFromRealIPHeader(trustOptions...))
	default:
		log.Debugf("IP extraction: direct (TCP remote address)")
		return withUnixSocketPeer(echo.ExtractIPDirect())
	}
}

func withUnixSocketPeer(extractor echo.IPExtractor) echo.IPExtractor {
	return func(req *http.Request) string {
		if remoteAddrHasIP(req.RemoteAddr) {
			return extractor(req)
		}

		withPeer := *req
		withPeer.RemoteAddr = unixSocketPeerAddr
		return extractor(&withPeer)
	}
}

func remoteAddrHasIP(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	return net.ParseIP(host) != nil
}

func parseTrustedProxies(proxies string) []echo.TrustOption {
	if proxies == "" {
		return nil
	}

	var options []echo.TrustOption
	for _, cidr := range strings.Split(proxies, ",") {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			log.Warningf("Invalid trusted proxy CIDR %q: %v", cidr, err)
			continue
		}
		options = append(options, echo.TrustIPRange(ipNet))
	}
	return options
}
