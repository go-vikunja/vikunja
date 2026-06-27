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

import "context"

// CreateToken mints an API token. If t.OwnerID is non-zero, the token is
// minted FOR that user — the caller must be the bot's owner (i.e. created
// the bot in step 8 of init).
func (c *Client) CreateToken(ctx context.Context, t *APIToken) (*APIToken, error) {
	var out APIToken
	if err := c.Do(ctx, "POST", "/tokens", nil, t, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
