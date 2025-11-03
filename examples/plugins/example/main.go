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

package main

import (
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/plugins"
	"code.vikunja.io/api/pkg/user"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"
)

type ExamplePlugin struct{}

func (p *ExamplePlugin) Name() string    { return "example" }
func (p *ExamplePlugin) Version() string { return "1.0.0" }
func (p *ExamplePlugin) Init() error {
	log.Infof("example plugin initialized")

	events.RegisterListener((&models.TaskCreatedEvent{}).Name(), &TestListener{})

	return nil
}
func (p *ExamplePlugin) Shutdown() error { return nil }

// RegisterAuthenticatedRoutes implements the AuthenticatedRouterPlugin interface
func (p *ExamplePlugin) RegisterAuthenticatedRoutes(g *echo.Group) {
	g.GET("/user-info", handleUserInfo)

	log.Infof("example plugin authenticated routes registered")
}

// RegisterUnauthenticatedRoutes implements the UnauthenticatedRouterPlugin interface
func (p *ExamplePlugin) RegisterUnauthenticatedRoutes(g *echo.Group) {
	g.GET("/status", handleStatus)

	log.Infof("example plugin unauthenticated routes registered")
}

// Authenticated route handlers
func handleUserInfo(c echo.Context) error {

	s := db.NewSession()
	defer s.Close()

	// Get the authenticated user from context
	u, err := user.GetCurrentUserFromDB(s, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}

	p := &ExamplePlugin{}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Hello from example plugin!",
		"user":    u,
		"plugin":  p.Name(),
		"version": p.Version(),
	})
}

// Unauthenticated route handlers
func handleStatus(c echo.Context) error {

	p := &ExamplePlugin{}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"plugin":  p.Name(),
		"version": p.Version(),
		"message": "Example plugin is running",
	})
}

func NewPlugin() plugins.Plugin { return &ExamplePlugin{} }

type TestListener struct{}

func (t *TestListener) Handle(msg *message.Message) error {
	log.Infof("TestListener received message: %s", string(msg.Payload))
	return nil
}

func (t *TestListener) Name() string {
	return "TestListener"
}
