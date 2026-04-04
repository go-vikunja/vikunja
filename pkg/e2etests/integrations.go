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

package e2etests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/routes"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
)

var registerListenersOnce sync.Once

// Test users matching the fixture data in pkg/db/fixtures/users.yml
var (
	testuser1 = user.User{
		ID:       1,
		Username: "user1",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Email:    "user1@example.com",
		Issuer:   "local",
	}
)

// setupE2ETestEnv initializes the full application environment with real events.
// Unlike setupTestEnv in pkg/webtests/, this does NOT call events.Fake(),
// so events are dispatched through the real Watermill router to registered listeners.
func setupE2ETestEnv(ctx context.Context) (e *echo.Echo, err error) {
	config.InitDefaultConfig()
	config.ServicePublicURL.Set("https://localhost")
	config.WebhooksEnabled.Set(true)
	config.WebhooksAllowNonRoutableIPs.Set(true)

	log.InitLogger()

	files.InitTests()
	user.InitTests()
	models.SetupTests() //nolint:contextcheck
	keyvalue.InitStorage()

	err = db.LoadFixtures()
	if err != nil {
		return
	}

	// Register all listeners (including webhook listener) before starting the router.
	// This must happen before InitEventsForTesting because the router wires up
	// all listeners that were registered via events.RegisterListener().
	// Use sync.Once because RegisterListeners appends to the global registry
	// and calling it multiple times would stack duplicate handlers.
	registerListenersOnce.Do(models.RegisterListeners)

	// Start the real watermill event system. InitEventsForTesting initializes
	// pubsub and starts the router in a background goroutine, returning a
	// channel that closes once the router is ready.
	ready, err := events.InitEventsForTesting(ctx)
	if err != nil {
		return
	}

	// user.InitTests() calls events.Fake() which sets isUnderTest=true and
	// prevents real event dispatch. Undo that now that pubsub is initialized.
	events.Unfake()

	// Wait for the router to be ready before proceeding.
	<-ready

	e = routes.NewEcho()     //nolint:contextcheck
	routes.RegisterRoutes(e) //nolint:contextcheck
	return
}

// createRequest builds an httptest request and echo context, mirroring webtests.createRequest
func createRequest(e *echo.Echo, method string, payload string, queryParam url.Values, urlParams map[string]string) (c *echo.Context, rec *httptest.ResponseRecorder) {
	req := httptest.NewRequestWithContext(context.Background(), method, "/", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.URL.RawQuery = queryParam.Encode()
	rec = httptest.NewRecorder()

	c = e.NewContext(req, rec)
	if len(urlParams) > 0 {
		pathValues := make(echo.PathValues, 0, len(urlParams))
		for name, value := range urlParams {
			pathValues = append(pathValues, echo.PathValue{Name: name, Value: value})
		}
		c.SetPathValues(pathValues)
	}
	return
}

// addUserTokenToContext creates a JWT for the user and sets it on the echo context
func addUserTokenToContext(t *testing.T, u *user.User, c *echo.Context) {
	token, err := auth.NewUserJWTAuthtoken(u, "test-session-id")
	require.NoError(t, err)
	tken, err := jwt.Parse(token, func(_ *jwt.Token) (interface{}, error) {
		return []byte(config.ServiceSecret.GetString()), nil
	})
	require.NoError(t, err)
	c.Set("user", tken)
}

// testUpdateWithUser performs a POST (update) request as the given user
func testUpdateWithUser(e *echo.Echo, t *testing.T, u *user.User, urlParams map[string]string, payload string) (rec *httptest.ResponseRecorder, err error) {
	c, rec := createRequest(e, http.MethodPost, payload, nil, urlParams)
	addUserTokenToContext(t, u, c)

	hndl := handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return &models.Task{}
		},
	}
	err = hndl.UpdateWeb(c)
	return
}
