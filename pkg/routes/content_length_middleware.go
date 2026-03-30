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

package routes

import (
	"encoding/json"
	"strconv"

	"github.com/labstack/echo/v5"
)

// ContentLengthJSONSerializer wraps Echo's default JSON serializer to always
// set the Content-Length header on JSON responses. This is a server-side
// mitigation for a known macOS curl bug: when piping curl output to another
// program (e.g. curl | jq), the receiving program can get empty stdin if the
// response uses chunked transfer encoding without a Content-Length header.
//
// The default serializer uses json.NewEncoder which streams directly to the
// response writer, so Go's HTTP server cannot pre-calculate Content-Length
// and falls back to chunked encoding. This serializer marshals to a byte
// slice first, sets Content-Length, then writes the bytes.
type ContentLengthJSONSerializer struct{}

// Serialize marshals the target to JSON bytes, sets Content-Length, then writes
// the response. This ensures the Content-Length header is always present.
func (s ContentLengthJSONSerializer) Serialize(c *echo.Context, target any, indent string) error {
	var data []byte
	var err error

	if indent != "" {
		data, err = json.MarshalIndent(target, "", indent)
	} else {
		data, err = json.Marshal(target)
	}
	if err != nil {
		return err
	}

	// Append newline for consistency with encoding/json.Encoder behavior
	data = append(data, '\n')

	c.Response().Header().Set("Content-Length", strconv.Itoa(len(data)))
	_, err = c.Response().Write(data)
	return err
}

// Deserialize decodes JSON from the request body into the target.
func (s ContentLengthJSONSerializer) Deserialize(c *echo.Context, target any) error {
	return json.NewDecoder(c.Request().Body).Decode(target)
}
