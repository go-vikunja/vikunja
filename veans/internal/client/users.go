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

package client

import (
	"context"
	"errors"
	"net/http"

	"code.vikunja.io/veans/internal/output"
)

// CreateBotUser provisions a bot user via PUT /bots. The username must be
// prefixed `bot-` (Vikunja enforces this). The caller becomes the bot's
// owner, which is what allows them to mint API tokens for the bot via
// PUT /tokens with owner_id.
//
// On Vikunja versions that predate the /bots endpoint, the server returns
// 404, which we surface as BOT_USERS_UNAVAILABLE so init can fail fast with
// a clear message.
func (c *Client) CreateBotUser(ctx context.Context, username, name string) (*BotUser, error) {
	var out BotUser
	err := c.Do(ctx, "PUT", "/bots", nil, &BotUserCreate{Username: username, Name: name}, &out)
	if err != nil {
		var oe *output.Error
		if errors.As(err, &oe) && oe.Code == output.CodeNotFound {
			return nil, output.Wrap(output.CodeBotUsersUnavailable, err,
				"this Vikunja instance does not expose /bots — upgrade to a newer version")
		}
		return nil, err
	}
	return &out, nil
}

// ListBotUsers returns all bot users owned by the authenticated user.
func (c *Client) ListBotUsers(ctx context.Context) ([]*BotUser, error) {
	var out []*BotUser
	if err := c.Do(ctx, "GET", "/bots", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// statusCheck pulls the HTTP status off an error for callers that need to
// distinguish 404-on-/bots from other failures. Currently unused outside this
// file, but kept for symmetry.
var _ = http.StatusNotFound
