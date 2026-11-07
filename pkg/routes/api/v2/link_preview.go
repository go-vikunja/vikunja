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

	"code.vikunja.io/api/pkg/modules/linkpreview"

	"github.com/danielgtaylor/huma/v2"
)

func init() { AddRouteRegistrar(RegisterLinkPreviewRoutes) }

func RegisterLinkPreviewRoutes(api huma.API) {
	Register(api, huma.Operation{
		OperationID: "link-preview-read",
		Summary:     "Get a link preview",
		Description: "Fetches Open Graph metadata with HTML title and description fallbacks. Requires authentication. Only external HTTP(S) URLs without credentials are accepted; this instance's public URL is excluded, including redirects. Outgoing-request restrictions apply. HTML is limited to 2 MiB and fetching to eight seconds. Returns 422 for an invalid URL, 502 for an unavailable preview, or 504 for a timeout.",
		Method:      http.MethodGet,
		Path:        "/link-preview",
		Tags:        []string{"link previews"},
	}, linkPreviewRead)
}

func linkPreviewRead(ctx context.Context, in *struct {
	URL string `query:"url" required:"true" maxLength:"4096" doc:"External HTTP(S) URL to preview."`
}) (*singleBody[linkpreview.Preview], error) {
	if _, err := authFromCtx(ctx); err != nil {
		return nil, err
	}
	preview, err := linkpreview.GetPreview(ctx, in.URL)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[linkpreview.Preview]{Body: preview}, nil
}
