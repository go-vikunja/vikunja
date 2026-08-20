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
	"errors"
	"net/http"

	"code.vikunja.io/api/pkg/modules/linkpreview"

	"github.com/danielgtaylor/huma/v2"
)

type linkPreviewBody struct {
	Body linkpreview.Preview
}

// RegisterLinkPreviewRoutes wires the link-preview endpoint onto the Huma API.
func RegisterLinkPreviewRoutes(api huma.API) {
	Register(api, huma.Operation{
		OperationID: "link-preview",
		Summary:     "Get a link preview",
		Description: "Fetches an external http(s) URL and returns its OpenGraph/meta preview (title, description, image, site name, favicon). Outbound requests are SSRF-protected. Requires authentication.",
		Method:      http.MethodGet,
		Path:        "/link-preview",
		Tags:        []string{"link-preview"},
	}, linkPreview)
}

func init() { AddRouteRegistrar(RegisterLinkPreviewRoutes) }

func linkPreview(ctx context.Context, in *struct {
	URL string `query:"url" doc:"The absolute http(s) URL to preview."`
}) (*linkPreviewBody, error) {
	if _, err := authFromCtx(ctx); err != nil {
		return nil, err
	}

	preview, err := linkpreview.GetPreview(ctx, in.URL)
	switch {
	case err == nil:
		return &linkPreviewBody{Body: *preview}, nil
	case errors.Is(err, linkpreview.ErrInvalidURL):
		return nil, huma.Error422UnprocessableEntity(err.Error())
	case errors.Is(err, linkpreview.ErrTimeout):
		return nil, huma.Error504GatewayTimeout("timed out fetching url")
	default:
		return nil, huma.Error502BadGateway("could not fetch url", err)
	}
}
