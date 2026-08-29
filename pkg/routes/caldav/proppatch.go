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

// propertyUpdateXML models RFC 4918 §9.2 <propertyupdate> just enough to
// know which properties the client tried to change; we never apply them.
type propertyUpdateXML struct {
	XMLName xml.Name         `xml:"propertyupdate"`
	Set     []propUpdateItem `xml:"set"`
	Remove  []propUpdateItem `xml:"remove"`
}

type propUpdateItem struct {
	Prop propNames `xml:"prop"`
}

type propNames struct {
	Names []struct {
		XMLName xml.Name
	} `xml:",any"`
}

func parsePropertyUpdate(body string) ([]xml.Name, error) {
	var pu propertyUpdateXML
	if err := xml.Unmarshal([]byte(body), &pu); err != nil {
		return nil, err
	}

	var names []xml.Name
	for _, item := range pu.Set {
		for _, n := range item.Prop.Names {
			names = append(names, n.XMLName)
		}
	}
	for _, item := range pu.Remove {
		for _, n := range item.Prop.Names {
			names = append(names, n.XMLName)
		}
	}
	return names, nil
}

// handlePropPatch answers PROPPATCH (RFC 4918 §9.2) without applying any
// change: Vikunja doesn't support client-set collection metadata, so every
// submitted property is refused with 403, the same "polite refusal" shape
// sabre/dav uses for protected properties. Clients treat this as best-effort
// and keep syncing instead of aborting on the 501 caldav-go falls back to.
func handlePropPatch(c *echo.Context, body string, project *models.ProjectWithTasksAndBuckets, u *user.User) error {
	s := db.NewSession()
	defer s.Close()

	can, _, err := project.CanRead(s, u)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error").Wrap(err)
	}
	if !can {
		return c.String(http.StatusForbidden, "Forbidden")
	}

	props, err := parsePropertyUpdate(body)
	if err != nil {
		return c.String(http.StatusBadRequest, "Bad Request")
	}

	return writePropPatchResponse(c, props)
}

func writePropPatchResponse(c *echo.Context, props []xml.Name) error {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="utf-8"?>` + "\n")
	sb.WriteString(`<D:multistatus xmlns:D="DAV:">` + "\n")
	sb.WriteString("  <D:response>\n")
	sb.WriteString("    <D:href>" + xmlEscape(c.Request().URL.Path) + "</D:href>\n")
	for _, p := range props {
		sb.WriteString("    <D:propstat>\n")
		sb.WriteString("      <D:prop>" + propTag(p) + "</D:prop>\n")
		sb.WriteString("      <D:status>HTTP/1.1 403 Forbidden</D:status>\n")
		sb.WriteString("    </D:propstat>\n")
	}
	sb.WriteString("  </D:response>\n")
	sb.WriteString("</D:multistatus>\n")

	c.Response().Header().Set("Content-Type", "application/xml; charset=utf-8")
	c.Response().WriteHeader(http.StatusMultiStatus)
	_, err := fmt.Fprint(c.Response(), sb.String())
	return err
}

func propTag(name xml.Name) string {
	if name.Space == "" {
		return "<" + name.Local + "/>"
	}
	return `<` + name.Local + ` xmlns="` + xmlEscape(name.Space) + `"/>`
}
