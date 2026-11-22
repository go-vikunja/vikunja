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

package migration

import (
	"errors"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedactSecrets(t *testing.T) {
	tests := map[string]string{
		"https://api.trello.com/1/members/me?key=abc&token=s3cret":         "https://api.trello.com/1/members/me?key=[REDACTED]&token=[REDACTED]",
		"https://example.com/?access_token=s3cret&page=2":                  "https://example.com/?access_token=[REDACTED]&page=2",
		"https://example.com/?Client_Secret=s3cret#frag":                   "https://example.com/?Client_Secret=[REDACTED]#frag",
		`Get "https://example.com/?token=s3cret": dial tcp: no such host`:  `Get "https://example.com/?token=[REDACTED]": dial tcp: no such host`,
		"https://graph.microsoft.com/v1.0/me/todo/lists?$skiptoken=abc123": "https://graph.microsoft.com/v1.0/me/todo/lists?$skiptoken=abc123",
		"duplicate key=value in map":                                       "duplicate key=value in map",
	}
	for in, want := range tests {
		assert.Equal(t, want, redactSecrets(in), in)
	}
}

func TestRedactError(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		require.NoError(t, RedactError(nil))
	})
	t.Run("without secrets keeps the error", func(t *testing.T) {
		err := fmt.Errorf("wrapped: %w", ErrResponseTooLarge)
		assert.Same(t, err, RedactError(err))
	})
	t.Run("wrapped url error", func(t *testing.T) {
		err := fmt.Errorf("HTTP request failure on https://api.trello.com/1/members/me/boards: %w", &url.Error{
			Op:  "Get",
			URL: "https://api.trello.com/1/members/me/boards?key=appkey&token=usertoken",
			Err: errors.New("dial tcp: lookup api.trello.com: no such host"),
		})

		redacted := RedactError(err)
		assert.NotContains(t, redacted.Error(), "usertoken")
		assert.NotContains(t, redacted.Error(), "appkey")
		assert.Contains(t, redacted.Error(), "no such host")
	})
}
