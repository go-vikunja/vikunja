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
	"errors"
	"fmt"
	"strings"

	"code.vikunja.io/api/pkg/config"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v5"
)

func NewCanonicalAPI() (huma.API, error) {
	config.InitDefaultConfig()
	config.AuthLocalEnabled.Set(true)
	config.AuthOpenIDEnabled.Set(true)
	config.ServiceEnableRegistration.Set(true)
	config.ServiceEnableLinkSharing.Set(true)
	config.ServiceEnableTotp.Set(true)
	config.ServiceEnableTaskAttachments.Set(true)
	config.ServiceEnableTaskComments.Set(true)
	config.WebhooksEnabled.Set(true)
	config.BackgroundsEnabled.Set(true)
	config.BackgroundsUploadEnabled.Set(true)
	config.BackgroundsUnsplashEnabled.Set(true)
	config.MigrationTodoistEnable.Set(true)
	config.MigrationTrelloEnable.Set(true)
	config.MigrationMicrosoftTodoEnable.Set(true)
	config.ServiceTestingtoken.Set("")
	config.ServicePublicURL.Set("")

	e := echo.New()
	api := NewAPI(e, e.Group(GroupPrefix))
	RegisterAll(api)
	document := api.OpenAPI()
	document.Servers = []*huma.Server{{URL: GroupPrefix}}

	callback := document.Paths["/auth/openid/{provider}/callback"]
	if callback == nil || callback.Post == nil {
		return nil, errors.New("canonical API is missing POST /auth/openid/{provider}/callback")
	}
	for path := range document.Paths {
		if strings.HasPrefix(path, "/test/") {
			return nil, fmt.Errorf("canonical API unexpectedly contains %s", path)
		}
	}

	return api, nil
}
