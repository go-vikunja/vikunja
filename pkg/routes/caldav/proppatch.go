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

package caldav

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"github.com/labstack/echo/v5"
)

const (
	maxPropPatchBodySize       = 64 * 1024
	maxPropPatchProperties     = 100
	maxPropPatchNamespaceBytes = 256
)

var errTooManyProperties = fmt.Errorf("%w: too many properties", errBodyTooLarge)

// parsePropertyUpdate returns the distinct property names under
// <propertyupdate><set|remove><prop>. Properties are echoed back with their
// namespace, so namespace length and property count are capped to bound the
// response size.
func parsePropertyUpdate(body string) ([]xml.Name, error) {
	d := xml.NewDecoder(strings.NewReader(body))
	seen := map[xml.Name]struct{}{}
	var names []xml.Name
	var path []string
	for {
		tok, err := d.Token()
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errMalformedBody, err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if len(path) == 0 && t.Name.Local != "propertyupdate" {
				return nil, fmt.Errorf("%w: expected <propertyupdate>, got <%s>", errMalformedBody, t.Name.Local)
			}
			if len(path) == 3 && (path[1] == "set" || path[1] == "remove") && path[2] == "prop" {
				if len(t.Name.Space) > maxPropPatchNamespaceBytes {
					return nil, fmt.Errorf("%w: namespace of <%s> too long", errMalformedBody, t.Name.Local)
				}
				if _, ok := seen[t.Name]; !ok {
					if len(names) == maxPropPatchProperties {
						return nil, errTooManyProperties
					}
					seen[t.Name] = struct{}{}
					names = append(names, t.Name)
				}
			}
			path = append(path, t.Name.Local)
		case xml.EndElement:
			path = path[:len(path)-1]
			if len(path) == 0 {
				return names, nil
			}
		}
	}
}

// handlePropPatch answers PROPPATCH (RFC 4918 §9.2) without applying any
// change: Vikunja doesn't support client-set collection metadata, so every
// submitted property is refused with 403, the same "polite refusal" shape
// sabre/dav uses for protected properties. Clients treat this as best-effort
// and keep syncing instead of aborting on the 501 caldav-go falls back to.
func handlePropPatch(c *echo.Context, body string, project *models.ProjectWithTasksAndBuckets, u *user.User) error {
	if len(body) > maxPropPatchBodySize {
		return writeError(c, errBodyTooLarge)
	}

	s := db.NewSession()
	defer s.Close()

	can, _, err := project.CanRead(s, u)
	if err != nil {
		return writeError(c, err)
	}
	if !can {
		return writeError(c, errForbidden)
	}

	props, err := parsePropertyUpdate(body)
	if err != nil {
		return writeError(c, err)
	}

	return writePropPatchResponse(c, props)
}

func writePropPatchResponse(c *echo.Context, props []xml.Name) error {
	r := propstatResponse{href: c.Request().URL.Path, propstats: make([]propstat, len(props))}
	for i, p := range props {
		r.propstats[i] = propstat{status: http.StatusForbidden, props: []prop{{name: p}}}
	}
	return writeMultistatus(c, multistatus{responses: []msResponse{r}})
}
