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
	"net/http"
	"strings"
)

// Pinned: writes compare only the first If-Match, exactly, and check If-None-Match only for "*".
type preconditions struct {
	ifMatch     []string
	ifNoneMatch []string
}

func preconditionsOf(h http.Header) preconditions {
	return preconditions{
		ifMatch:     h.Values("If-Match"),
		ifNoneMatch: h.Values("If-None-Match"),
	}
}

func (p preconditions) ifMatches(etag string) bool {
	return len(p.ifMatch) == 0 || p.ifMatch[0] == "*" || p.ifMatch[0] == etag
}

func (p preconditions) allowCreate() bool {
	return len(p.ifMatch) == 0
}

func (p preconditions) allowUpdate(etag string) bool {
	return p.ifMatches(etag) && (len(p.ifNoneMatch) != 1 || p.ifNoneMatch[0] != "*")
}

func (p preconditions) allowDelete(etag string) bool {
	return p.ifMatches(etag)
}

// RFC 9110 §13.1.2: weak comparison over the whole list.
func (p preconditions) notModified(etag string) bool {
	for _, value := range p.ifNoneMatch {
		for _, candidate := range strings.Split(value, ",") {
			candidate = strings.TrimSpace(candidate)
			if candidate == "*" || strings.TrimPrefix(candidate, "W/") == strings.TrimPrefix(etag, "W/") {
				return true
			}
		}
	}
	return false
}
