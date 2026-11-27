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
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/models"
)

const DavBasePath = `/dav`

const ProjectBasePath = DavBasePath + `/projects`

const PrincipalBasePath = DavBasePath + `/principals`

func principalPathForUser(username string) string {
	return PrincipalBasePath + `/` + encodeURIPathSegment(username) + `/`
}

func collectionHref(projectID int64) string {
	return ProjectBasePath + "/" + strconv.FormatInt(projectID, 10)
}

// WebDAV requires every Depth: 1 child to live under the collection's own URI.
func taskURL(collectionProjectID int64, task *models.Task) string {
	return collectionHref(collectionProjectID) + `/` + encodeURIPathSegment(task.UID) + `.ics`
}

// & too, which url.PathEscape leaves alone: clients compare the hrefs they stored byte for byte.
func encodeURIPathSegment(segment string) string {
	return strings.ReplaceAll(url.PathEscape(segment), "&", "%26")
}
