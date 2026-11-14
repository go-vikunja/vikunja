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
	"errors"
	"net/url"
	"strings"

	"code.vikunja.io/api/pkg/models"
	user2 "code.vikunja.io/api/pkg/user"
	"github.com/samedi/caldav-go/data"
)

// DavBasePath is the base url path
const DavBasePath = `/dav`

// ProjectBasePath is the base path for all projects resources
const ProjectBasePath = DavBasePath + `/projects`

// PrincipalBasePath is the base path for all principal resources
const PrincipalBasePath = DavBasePath + `/principals`

// ProjectHomeSetPath is the CalDAV home-set path Apple clients use after discovery.
const ProjectHomeSetPath = ProjectBasePath + `/`

// VikunjaCaldavProjectStorage represents a project storage
type VikunjaCaldavProjectStorage struct {
	// Used when handling a project
	project *models.ProjectWithTasksAndBuckets
	// Used when handling a single task, like updating
	task *models.Task
	// The current user
	user *user2.User
}

// GetResources only satisfies data.Storage.
func (vcls *VikunjaCaldavProjectStorage) GetResources(string, bool) ([]data.Resource, error) {
	return nil, errNotThroughLibrary
}

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

// caldav-go only answers the methods nobody handles, with 501, and never asks the storage.
var errNotThroughLibrary = errors.New("caldav: not served through caldav-go")

// GetResourcesByList only satisfies data.Storage.
func (vcls *VikunjaCaldavProjectStorage) GetResourcesByList([]string) ([]data.Resource, error) {
	return nil, errNotThroughLibrary
}

// GetResourcesByFilters only satisfies data.Storage.
func (vcls *VikunjaCaldavProjectStorage) GetResourcesByFilters(string, *data.ResourceFilter) ([]data.Resource, error) {
	return nil, errNotThroughLibrary
}

// GetResource only satisfies data.Storage.
func (vcls *VikunjaCaldavProjectStorage) GetResource(string) (*data.Resource, bool, error) {
	return nil, false, errNotThroughLibrary
}

// GetShallowResource only satisfies data.Storage.
func (vcls *VikunjaCaldavProjectStorage) GetShallowResource(string) (*data.Resource, bool, error) {
	return nil, false, errNotThroughLibrary
}

// CreateResource only satisfies data.Storage.
func (vcls *VikunjaCaldavProjectStorage) CreateResource(string, string) (*data.Resource, error) {
	return nil, errNotThroughLibrary
}

// UpdateResource only satisfies data.Storage.
func (vcls *VikunjaCaldavProjectStorage) UpdateResource(string, string) (*data.Resource, error) {
	return nil, errNotThroughLibrary
}

// DeleteResource only satisfies data.Storage.
func (vcls *VikunjaCaldavProjectStorage) DeleteResource(string) error {
	return errNotThroughLibrary
}
