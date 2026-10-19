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

package log

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeHeaders(t *testing.T) {
	h := http.Header{
		"Authorization":       {"Basic dXNlcjpodW50ZXIy"},
		"Proxy-Authorization": {"Bearer tk_abc"},
		"Cookie":              {"session=abc"},
		"X-Api-Token":         {"tk_def"},
		"Depth":               {"1"},
		"Content-Type":        {"text/xml"},
	}

	t.Run("fmt", func(t *testing.T) {
		out := fmt.Sprintf("%v", SafeHeaders(h))
		for _, secret := range []string{"dXNlcjpodW50ZXIy", "tk_abc", "session=abc", "tk_def"} {
			assert.NotContains(t, out, secret)
		}
		assert.Contains(t, out, "Authorization:[Basic [REDACTED]]")
		assert.Contains(t, out, "Proxy-Authorization:[Bearer [REDACTED]]")
		assert.Contains(t, out, "Depth:[1]")
		assert.Contains(t, out, "Content-Type:[text/xml]")
	})

	t.Run("slog", func(t *testing.T) {
		buf := &bytes.Buffer{}
		slog.New(slog.NewTextHandler(buf, nil)).LogAttrs(context.Background(), slog.LevelInfo, "", slog.Any("headers", SafeHeaders(h)))
		assert.NotContains(t, buf.String(), "dXNlcjpodW50ZXIy")
		assert.Contains(t, buf.String(), "Depth:[1]")
	})

	t.Run("does not mutate input", func(t *testing.T) {
		_ = SafeHeaders(h).String()
		assert.Equal(t, "Basic dXNlcjpodW50ZXIy", h.Get("Authorization"))
	})
}
