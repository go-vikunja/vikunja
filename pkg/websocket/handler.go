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

package websocket

import (
	"context"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"

	"github.com/coder/websocket"
	"github.com/labstack/echo/v5"
)

var (
	globalHub *Hub
	// Computed once at startup so the wildcard warning is not logged per connection.
	originPatterns []string
)

// InitHub creates the global hub. Must be called once at startup.
func InitHub() {
	globalHub = NewHub()
	originPatterns = allowedOriginPatterns(config.CorsOrigins.GetStringSlice(), config.ServicePublicURL.GetString())
}

// allowedOriginPatterns derives the websocket origin patterns from the configured CORS origins.
// "*" is a supported cors.origins value for REST, but as a websocket origin pattern it matches every
// host and turns origin verification off, so wildcard hosts are dropped. The public url is always
// allowed to keep same-origin deployments working without extra configuration.
func allowedOriginPatterns(corsOrigins []string, publicURL string) []string {
	patterns := make([]string, 0, len(corsOrigins)+1)
	for _, origin := range corsOrigins {
		if isWildcardHost(origin) {
			log.Warningf("WebSocket: ignoring cors.origins entry %q, it would disable origin verification. Configure the origins your frontend is served from explicitly.", origin)
			continue
		}
		patterns = append(patterns, origin)
	}

	if own := schemeAndHost(publicURL); own != "" && !slices.Contains(patterns, own) {
		patterns = append(patterns, own)
	}

	return patterns
}

// isWildcardHost reports whether the pattern matches any host, with or without a scheme prefix.
func isWildcardHost(pattern string) bool {
	host := pattern
	if _, after, found := strings.Cut(pattern, "://"); found {
		host = after
	}
	return host == "*"
}

func schemeAndHost(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

// GetHub returns the global hub.
func GetHub() *Hub {
	return globalHub
}

// UpgradeHandler is the Echo handler for WebSocket upgrades at /api/v1/ws.
// The upgrade happens without authentication - auth is done via the first message.
func UpgradeHandler(c *echo.Context) error {
	if globalHub == nil {
		log.Errorf("WebSocket: hub not initialized")
		return echo.NewHTTPError(http.StatusServiceUnavailable, "WebSocket hub not initialized")
	}

	ws, err := websocket.Accept(c.Response(), c.Request(), &websocket.AcceptOptions{
		OriginPatterns: originPatterns,
	})
	if err != nil {
		log.Errorf("WebSocket: upgrade failed: %v", err)
		return nil // Accept already wrote the error response
	}

	conn := NewConnection(ws, globalHub)

	ctx, cancel := context.WithCancel(context.Background())

	go conn.WriteLoop(ctx, cancel)
	go conn.ReadLoop(ctx, cancel)

	return nil
}
