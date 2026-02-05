package websocket

import (
	"context"

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
	ws, err := websocket.Accept(c.Response(), c.Request(), &websocket.AcceptOptions{})
	if err != nil {
		log.Errorf("WebSocket: upgrade failed: %v", err)
		return nil // Accept already wrote the error response
	}

	conn := NewConnection(ws, globalHub)

	ctx := context.Background()

	go conn.WriteLoop(ctx)
	go conn.ReadLoop(ctx)

	return nil
}
