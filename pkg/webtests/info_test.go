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

package webtests

import (
	"fmt"
	"net/http"
	"testing"

	apiv1 "code.vikunja.io/api/pkg/routes/api/v1"
	"code.vikunja.io/api/pkg/version"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfo(t *testing.T) {
	rec, err := newTestRequest(t, http.MethodGet, apiv1.Info, ``, nil, nil)
	require.NoError(t, err)

	expected := fmt.Sprintf(`{
  "version": "%s",
  "api_min_compatible": "%s",
  "frontend_url": "https://localhost",
  "motd": "",
  "link_sharing_enabled": true,
  "max_file_size": "20MB",
  "max_items_per_page": 50,
  "available_migrators": [
    "vikunja-file",
    "ticktick"
  ],
  "task_attachments_enabled": true,
  "enabled_background_providers": [
    "upload"
  ],
  "totp_enabled": true,
  "legal": {
    "imprint_url": "",
    "privacy_policy_url": ""
  },
  "caldav_enabled": true,
  "auth": {
    "local": {
      "enabled": true,
      "registration_enabled": true
    },
    "ldap": {
      "enabled": false
    },
    "openid_connect": {
      "enabled": false,
      "providers": null
    }
  },
  "email_reminders_enabled": true,
  "user_deletion_enabled": true,
  "task_comments_enabled": true,
  "demo_mode_enabled": false,
  "webhooks_enabled": true,
  "public_teams_enabled": false
}`, version.Version, version.APIMinCompatible)

	assert.JSONEq(t, expected, rec.Body.String())
}
