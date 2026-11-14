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
	"net/url"
	"strings"

	"code.vikunja.io/api/pkg/models"
)

// DavBasePath is the base url path
const DavBasePath = `/dav`

// ProjectBasePath is the base path for all projects resources
const ProjectBasePath = DavBasePath + `/projects`

// PrincipalBasePath is the base path for all principal resources
const PrincipalBasePath = DavBasePath + `/principals`

func principalPathForUser(username string) string {
	return PrincipalBasePath + `/` + username + `/`
}

// WebDAV requires every Depth: 1 child to live under the collection's own URI.
func taskURL(collectionProjectID int64, task *models.Task) string {
	return collectionHref(collectionProjectID) + `/` + encodeURIPathSegment(task.UID) + `.ics`
}

// url.PathEscape leaves & alone, and caldav-go writes hrefs into XML unescaped.
func encodeURIPathSegment(segment string) string {
	return strings.ReplaceAll(url.PathEscape(segment), "&", "%26")
}
