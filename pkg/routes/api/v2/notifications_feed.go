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

package apiv2

import (
	"context"
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/modules/humabridge"
	"code.vikunja.io/api/pkg/routes/feeds"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v5"
)

// RegisterNotificationsFeedRoutes wires the Atom notifications feed onto the
// Huma API. It documents HTTP Basic auth (a feeds-scoped API token) because
// feed readers can't carry a bearer header.
func RegisterNotificationsFeedRoutes(api huma.API) {
	Register(api, huma.Operation{
		OperationID: "notifications-atom-feed",
		Summary:     "Notifications Atom feed",
		Description: "Returns the authenticated user's latest notifications as an Atom feed. Authenticated with HTTP Basic auth: the username is the token owner and the password is a feeds-scoped Vikunja API token (tk_ prefix) — password and LDAP credentials are rejected because feed URLs are commonly shared or cached. Fetching the feed does not mark notifications as read.",
		Method:      http.MethodGet,
		Path:        "/notifications.atom",
		Tags:        []string{"service"},
		// This op carries its own HTTP Basic auth instead of the global bearer
		// schemes; the path is in unauthenticatedAPIPaths so the JWT middleware
		// lets it through and the handler authenticates itself.
		Security: []map[string][]string{{"BasicAuth": {}}},
		Responses: map[string]*huma.Response{
			"200": {
				Description: "The notifications Atom feed.",
				Content: map[string]*huma.MediaType{
					"application/atom+xml": {
						Schema: &huma.Schema{Type: huma.TypeString, Format: "binary"},
					},
				},
			},
		},
	}, notificationsAtomFeed)
}

func init() { AddRouteRegistrar(RegisterNotificationsFeedRoutes) }

// notificationsAtomFeed authenticates with HTTP Basic (sharing the feeds
// validator) and streams the Atom feed; there is no handler.Do* for a non-JSON
// body and the auth can't ride the group's JWT middleware.
func notificationsAtomFeed(ctx context.Context, _ *struct{}) (*huma.StreamResponse, error) {
	c, ok := ctx.Value(humabridge.EchoContextKey).(*echo.Context)
	if !ok {
		return nil, huma.Error500InternalServerError("could not resolve request context")
	}

	username, password, ok := (*c).Request().BasicAuth()
	if !ok {
		return nil, basicAuthChallenge(c)
	}

	s := db.NewSession()
	defer s.Close()

	u, err := feeds.AuthenticateFeedToken(s, username, password)
	if err != nil {
		return nil, translateDomainError(err)
	}
	if u == nil {
		return nil, basicAuthChallenge(c)
	}

	atom, err := feeds.BuildNotificationsAtomFeed(s, u)
	if err != nil {
		return nil, translateDomainError(err)
	}

	return &huma.StreamResponse{Body: func(hctx huma.Context) {
		ec := humaecho.Unwrap(hctx)
		(*ec).Response().Header().Set(echo.HeaderContentType, feeds.AtomContentType)
		_, _ = (*ec).Response().Write([]byte(atom))
	}}, nil
}

// basicAuthChallenge returns a 401 carrying a WWW-Authenticate Basic challenge,
// mirroring v1's BasicAuth middleware so feed readers prompt for credentials.
func basicAuthChallenge(c *echo.Context) error {
	(*c).Response().Header().Set(echo.HeaderWWWAuthenticate, `Basic realm="Restricted"`)
	return huma.Error401Unauthorized(http.StatusText(http.StatusUnauthorized))
}
