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
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

const redacted = "[REDACTED]"

var sensitiveHeaderNameParts = []string{"auth", "cookie", "token", "secret", "password", "key"}

// SafeHeaders masks credential headers when logged, via %v or slog.
type SafeHeaders http.Header

func (h SafeHeaders) String() string {
	return fmt.Sprint(map[string][]string(h.redact()))
}

func (h SafeHeaders) LogValue() slog.Value {
	return slog.AnyValue(map[string][]string(h.redact()))
}

func (h SafeHeaders) redact() http.Header {
	out := make(http.Header, len(h))
	for name, values := range h {
		if !isSensitiveHeader(name) {
			out[name] = values
			continue
		}
		masked := make([]string, len(values))
		for i, v := range values {
			masked[i] = redacted
			// Keep the scheme (Basic/Bearer), it's useful for debugging and not secret.
			if scheme, _, ok := strings.Cut(v, " "); ok && strings.HasSuffix(strings.ToLower(name), "authorization") {
				masked[i] = scheme + " " + redacted
			}
		}
		out[name] = masked
	}
	return out
}

func isSensitiveHeader(name string) bool {
	name = strings.ToLower(name)
	for _, part := range sensitiveHeaderNameParts {
		if strings.Contains(name, part) {
			return true
		}
	}
	return false
}
