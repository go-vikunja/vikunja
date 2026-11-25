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
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo/v5"
	"xorm.io/xorm"
)

const (
	maxPropPatchBodySize       = 64 * 1024
	maxPropPatchProperties     = 100
	maxPropPatchNamespaceBytes = 256
)

var errTooManyProperties = errors.New("too many properties")

var calendarColorProp = xml.Name{Space: "http://apple.com/ns/ical/", Local: "calendar-color"}

func normalizeCalendarColor(value string) (hex string, ok bool) {
	hex = strings.TrimPrefix(strings.TrimSpace(value), "#")
	if (len(hex) != 6 && len(hex) != 8) || !govalidator.IsHexadecimal(hex) {
		return "", false
	}
	return hex[:6], true
}

type propPatchInstruction struct {
	name   xml.Name
	remove bool
	value  string
}

// The last instruction per property wins. Namespace length and property count
// are capped because properties are echoed back.
func parsePropertyUpdate(body string) ([]*propPatchInstruction, error) {
	d := xml.NewDecoder(strings.NewReader(body))
	seen := map[xml.Name]*propPatchInstruction{}
	var instructions []*propPatchInstruction
	var current *propPatchInstruction
	var path []string
	for {
		tok, err := d.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if len(path) == 0 && t.Name.Local != "propertyupdate" {
				return nil, fmt.Errorf("expected <propertyupdate>, got <%s>", t.Name.Local)
			}
			if len(path) == 3 && (path[1] == "set" || path[1] == "remove") && path[2] == "prop" {
				if len(t.Name.Space) > maxPropPatchNamespaceBytes {
					return nil, fmt.Errorf("namespace of <%s> too long", t.Name.Local)
				}
				current = seen[t.Name]
				if current == nil {
					if len(instructions) == maxPropPatchProperties {
						return nil, errTooManyProperties
					}
					current = &propPatchInstruction{name: t.Name}
					seen[t.Name] = current
					instructions = append(instructions, current)
				}
				current.remove = path[1] == "remove"
				current.value = ""
			}
			path = append(path, t.Name.Local)
		case xml.CharData:
			if len(path) == 4 && current != nil {
				current.value += string(t)
			}
		case xml.EndElement:
			path = path[:len(path)-1]
			if len(path) == 3 {
				current = nil
			}
			if len(path) == 0 {
				return instructions, nil
			}
		}
	}
}

type propPatchTarget int

const (
	propPatchCollection propPatchTarget = iota
	propPatchTask
)

type propPatchResult struct {
	name   xml.Name
	status int
}

// Only a collection's calendar-color is settable; everything else gets
// sabre/dav's 403 "polite refusal", which clients tolerate unlike caldav-go's 501.
func handlePropPatch(c *echo.Context, body string, project *models.ProjectWithTasksAndBuckets, u *user.User, target propPatchTarget) error {
	if len(body) > maxPropPatchBodySize {
		return c.String(http.StatusRequestEntityTooLarge, "Request Entity Too Large")
	}

	s := db.NewSession()
	defer s.Close()

	can, _, err := project.CanRead(s, u)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error").Wrap(err)
	}
	if !can {
		return c.String(http.StatusForbidden, "Forbidden")
	}

	instructions, err := parsePropertyUpdate(body)
	if errors.Is(err, errTooManyProperties) {
		return c.String(http.StatusRequestEntityTooLarge, "Request Entity Too Large")
	}
	if err != nil {
		return c.String(http.StatusBadRequest, "Bad Request")
	}

	results, err := applyPropPatch(s, instructions, &project.Project, u, target)
	if err != nil {
		_ = s.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error").Wrap(err)
	}
	if err := s.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error").Wrap(err)
	}

	return writePropPatchResponse(c, results)
}

// applyPropPatch is all-or-nothing as RFC 4918 §9.2 demands: one refused
// property turns every acceptable one into a 424.
func applyPropPatch(s *xorm.Session, instructions []*propPatchInstruction, project *models.Project, u *user.User, target propPatchTarget) ([]propPatchResult, error) {
	results := make([]propPatchResult, 0, len(instructions))
	var color *string
	refused := false
	for _, in := range instructions {
		status := http.StatusForbidden
		if in.name == calendarColorProp && target == propPatchCollection {
			hex, ok := "", true
			if !in.remove {
				hex, ok = normalizeCalendarColor(in.value)
			}
			if ok {
				color = &hex
				status = http.StatusOK
			}
		}
		refused = refused || status != http.StatusOK
		results = append(results, propPatchResult{name: in.name, status: status})
	}

	if color != nil {
		can, err := canSetCalendarColor(s, project, u)
		if err != nil {
			return nil, err
		}
		refused = refused || !can
		if !can {
			setStatus(results, http.StatusOK, http.StatusForbidden)
		}
	}

	if refused {
		setStatus(results, http.StatusOK, http.StatusFailedDependency)
		return results, nil
	}
	if color == nil {
		return results, nil
	}

	// iOS re-sends the color on every launch; an Update would bump the ctag and fire webhooks.
	if project.HexColor == *color {
		return results, nil
	}

	// ReadOne fills in the favorite state, which Update would otherwise reset.
	if err := project.ReadOne(s, u); err != nil {
		return nil, err
	}
	project.HexColor = *color
	return results, project.Update(s, u)
}

func setStatus(results []propPatchResult, from, to int) {
	for i := range results {
		if results[i].status == from {
			results[i].status = to
		}
	}
}

func canSetCalendarColor(s *xorm.Session, project *models.Project, u *user.User) (bool, error) {
	// A saved filter's owner can update it, but it has no color to store.
	if models.IsPseudoProjectID(project.ID) {
		return false, nil
	}
	return denyArchived(project.CanUpdate(s, u))
}

func writePropPatchResponse(c *echo.Context, results []propPatchResult) error {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="utf-8"?>` + "\n")
	sb.WriteString(`<D:multistatus xmlns:D="DAV:">` + "\n")
	sb.WriteString("  <D:response>\n")
	sb.WriteString("    <D:href>" + xmlEscape(c.Request().URL.Path) + "</D:href>\n")
	for _, r := range results {
		sb.WriteString("    <D:propstat>\n")
		sb.WriteString("      <D:prop>" + propTag(r.name) + "</D:prop>\n")
		fmt.Fprintf(&sb, "      <D:status>HTTP/1.1 %d %s</D:status>\n", r.status, http.StatusText(r.status))
		sb.WriteString("    </D:propstat>\n")
	}
	sb.WriteString("  </D:response>\n")
	sb.WriteString("</D:multistatus>\n")

	c.Response().Header().Set("Content-Type", "application/xml; charset=utf-8")
	c.Response().WriteHeader(http.StatusMultiStatus)
	_, err := io.WriteString(c.Response(), sb.String())
	return err
}

func propTag(name xml.Name) string {
	if name.Space == "" {
		return "<" + name.Local + "/>"
	}
	return `<` + name.Local + ` xmlns="` + xmlEscape(name.Space) + `"/>`
}
