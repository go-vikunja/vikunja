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

import "net/http"

// preconditions are a request's If-Match and If-None-Match headers. Pinned: like caldav-go,
// If-Match compares only its first value, as an exact string.
type preconditions struct {
	ifMatch     []string
	ifNoneMatch []string
}

func preconditionsOf(h http.Header) preconditions {
	return preconditions{ifMatch: h.Values("If-Match"), ifNoneMatch: h.Values("If-None-Match")}
}

func (p preconditions) ifMatches(etag string) bool {
	return len(p.ifMatch) == 0 || p.ifMatch[0] == "*" || p.ifMatch[0] == etag
}

// allowCreate is for a PUT onto a resource that does not exist yet.
func (p preconditions) allowCreate() bool {
	return len(p.ifMatch) == 0
}

// allowUpdate is for a PUT onto an existing resource.
func (p preconditions) allowUpdate(etag string) bool {
	return p.ifMatches(etag) && (len(p.ifNoneMatch) != 1 || p.ifNoneMatch[0] != "*")
}

func (p preconditions) allowDelete(etag string) bool {
	return p.ifMatches(etag)
}
