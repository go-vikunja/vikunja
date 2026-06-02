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
	"code.vikunja.io/api/pkg/modules/avatar"

	"github.com/danielgtaylor/huma/v2"
)

// avatarResponse carries raw image bytes plus the runtime Content-Type. Huma writes
// the []byte Body straight to the wire; the header:"Content-Type" field overrides
// content negotiation so the provider's actual mime type reaches the client.
type avatarResponse struct {
	ContentType string `header:"Content-Type"`
	Body        []byte
}

type avatarInput struct {
	Username string `path:"username" doc:"The username of the user whose avatar to fetch."`
	Size     int64  `query:"size" default:"250" minimum:"1" doc:"Desired avatar edge length in pixels. Clamped to the server's configured maximum if larger; providers that render fixed-size images may ignore it."`
}

func init() { AddRouteRegistrar(RegisterAvatarRoutes) }

// RegisterAvatarRoutes wires the avatar binary endpoint onto the Huma API.
func RegisterAvatarRoutes(api huma.API) {
	Register(api, huma.Operation{
		OperationID: "avatar-get",
		Summary:     "Get a user's avatar",
		Description: "Returns the user's avatar as raw image bytes. The Content-Type is chosen at " +
			"runtime by the user's avatar provider (gravatar, initials, marble, an uploaded image, " +
			"or the default placeholder). An unknown username is not an error — the default " +
			"placeholder avatar is returned. Authenticated like every other endpoint.",
		Method: http.MethodGet,
		Path:   "/avatar/{username}",
		Tags:   []string{"user"},
		// Spell out the binary response; a bare []byte Body would otherwise be
		// modeled as a base64 JSON string instead of binary image data.
		Responses: map[string]*huma.Response{
			"200": {
				Description: "The avatar image bytes. The Content-Type header carries the actual image type.",
				Content: map[string]*huma.MediaType{
					"application/octet-stream": {
						Schema: &huma.Schema{Type: huma.TypeString, Format: "binary"},
					},
				},
			},
		},
	}, avatarGet)
}

func avatarGet(ctx context.Context, in *avatarInput) (*avatarResponse, error) {
	// Authenticated but no per-user check — any authenticated caller may view any
	// avatar (matching v1); authFromCtx just surfaces a clean 401 if auth is missing.
	if _, err := authFromCtx(ctx); err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	// Avatar resolution (user lookup, provider selection, size clamping) is
	// shared with the v1 handler; only the transport differs.
	a, mimeType, err := avatar.GetAvatarForUsername(s, in.Username, in.Size)
	if err != nil {
		return nil, translateDomainError(err)
	}

	return &avatarResponse{ContentType: mimeType, Body: a}, nil
}
