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
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	user2 "code.vikunja.io/api/pkg/user"
	"github.com/samedi/caldav-go/data"
	"github.com/samedi/caldav-go/errs"
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

func (vcls *VikunjaCaldavProjectStorage) collection() collectionTarget {
	return collectionTarget{project: vcls.project}
}

// GetResources only satisfies data.Storage: PROPFIND is served in-house.
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

// VikunjaProjectResourceAdapter holds the actual resource
type VikunjaProjectResourceAdapter struct {
	project      *models.ProjectWithTasksAndBuckets
	projectTasks []*models.TaskWithComments
	task         *models.Task

	isCollection bool

	// Rendered once per request: a multistatus response asks for it per requested property.
	content *string
}

// IsCollection checks if the resoure in the adapter is a collection
func (vlra *VikunjaProjectResourceAdapter) IsCollection() bool {
	// If the discovery does not work, setting this to true makes it work again.
	return vlra.isCollection
}

// CalculateEtag returns the etag of a resource
func (vlra *VikunjaProjectResourceAdapter) CalculateEtag() string {

	if vlra.task != nil {
		return taskETag(vlra.task)
	}
	if vlra.project == nil {
		return ""
	}
	return collectionETag(vlra.project, vlra.projectTasks)
}

// GetContent returns the content string of a resource (a task in our case)
func (vlra *VikunjaProjectResourceAdapter) GetContent() string {
	if vlra.content == nil {
		content := vlra.renderContent()
		vlra.content = &content
	}
	return *vlra.content
}

func (vlra *VikunjaProjectResourceAdapter) renderContent() string {
	if vlra.project != nil && vlra.projectTasks != nil {
		return renderCollection(vlra.project, vlra.projectTasks)
	}
	if vlra.task != nil {
		return renderTask(vlra.task)
	}
	return ""
}

// GetContentSize is the size of a caldav content
func (vlra *VikunjaProjectResourceAdapter) GetContentSize() int64 {
	return int64(len(vlra.GetContent()))
}

// GetModTime returns when the resource was last modified
func (vlra *VikunjaProjectResourceAdapter) GetModTime() time.Time {
	if vlra.task != nil {
		return vlra.task.Updated
	}

	if vlra.project != nil {
		return vlra.project.Updated
	}

	return time.Time{}
}

func (vcls *VikunjaCaldavProjectStorage) getProjectRessource(isCollection bool) (rr VikunjaProjectResourceAdapter, err error) {
	s := db.NewSession()
	defer s.Close()

	if vcls.project == nil {
		return
	}

	readable, err := vcls.collection().checkRead(s, vcls.user)
	var projectTasks []*models.TaskWithComments
	if err == nil {
		projectTasks, err = readCollectionTasks(s, vcls.user, readable)
	}
	if err != nil {
		_ = s.Rollback()
		if errors.Is(err, errNotFound) {
			// Any other error becomes a 500 in caldav-go, which would tell the client the collection exists.
			return rr, errs.ResourceNotFoundError
		}
		return rr, err
	}

	if err := s.Commit(); err != nil {
		return rr, err
	}

	rr = VikunjaProjectResourceAdapter{
		project:      vcls.project,
		projectTasks: projectTasks,
		isCollection: isCollection,
	}

	return
}
