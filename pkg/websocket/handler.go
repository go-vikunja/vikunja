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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"

	"github.com/coder/websocket"
	"github.com/labstack/echo/v5"
)

var globalHub *Hub

// InitHub creates the global hub. Must be called once at startup.
func InitHub() {
	globalHub = NewHub()
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
		OriginPatterns: config.CorsOrigins.GetStringSlice(),
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
