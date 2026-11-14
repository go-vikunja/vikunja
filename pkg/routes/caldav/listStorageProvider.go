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
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	user2 "code.vikunja.io/api/pkg/user"
	"github.com/samedi/caldav-go/data"
	"github.com/samedi/caldav-go/errs"
	"github.com/samedi/caldav-go/lib"
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

// GetResourcesByList fetches a list of resources from a slice of paths
func (vcls *VikunjaCaldavProjectStorage) GetResourcesByList(rpaths []string) (resources []data.Resource, err error) {

	// Path format: /dav/projects/{projectID}/{uid}.ics.
	// Remember the href's project ID per uid so the consistency check below
	// can drop tasks requested via the wrong project (GHSA-48ch-p4gq-x46x).
	var uids []string
	uidURLProjects := map[string]int64{}
	for _, path := range rpaths {
		parts := strings.Split(path, "/")
		if len(parts) < 5 {
			continue
		}
		// Skip malformed hrefs: without these guards an empty/non-numeric
		// project would bypass the consistency check, and an empty UID
		// would query for tasks with empty/NULL uids.
		if !strings.HasSuffix(parts[4], ".ics") {
			continue
		}
		// echo does not decode path params either, so hrefs travel percent-encoded end to end.
		uid, uerr := url.PathUnescape(strings.TrimSuffix(parts[4], ".ics"))
		if uerr != nil || uid == "" {
			continue
		}
		urlProjectID, perr := strconv.ParseInt(parts[3], 10, 64)
		if perr != nil {
			continue
		}
		uids = append(uids, uid)
		uidURLProjects[uid] = urlProjectID
	}

	if len(uids) == 0 {
		return
	}

	s := db.NewSession()
	defer s.Close()

	// GetTasksByUIDs...
	// Parse these into resources...
	tasks, err := models.GetTasksByUIDs(s, uids, vcls.user)
	if err != nil {
		_ = s.Rollback()
		return
	}

	// One membership lookup for the whole request: a multiget REPORT can name thousands
	// of hrefs and TaskCollection.ReadAll is expensive.
	var candidates []int64
	for _, t := range tasks {
		urlProjectID, ok := uidURLProjects[t.UID]
		if !ok || urlProjectID == t.ProjectID {
			continue
		}
		if !models.IsPseudoProjectID(urlProjectID) || vcls.project == nil || urlProjectID != vcls.project.ID {
			continue
		}
		candidates = append(candidates, t.ID)
	}

	members, err := vcls.collection().containsAll(s, vcls.user, candidates)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	emitted := make(map[string]bool, len(tasks))
	for _, t := range tasks {
		urlProjectID, ok := uidURLProjects[t.UID]
		if !ok {
			continue
		}
		// Closes the URL-path leak for calendar-multiget REPORTs (GHSA-48ch-p4gq-x46x).
		if urlProjectID != t.ProjectID {
			if !models.IsPseudoProjectID(urlProjectID) || vcls.project == nil || urlProjectID != vcls.project.ID {
				continue
			}
			if !members[t.ID] {
				continue
			}
		}
		// tasks.uid has no unique index, so two same-UID tasks would map to one href.
		href := taskURL(urlProjectID, t)
		if emitted[href] {
			continue
		}
		emitted[href] = true
		rr := VikunjaProjectResourceAdapter{
			task: t,
		}
		addTaskResource(urlProjectID, t, &rr, &resources)
	}

	err = s.Commit()
	if err != nil {
		return nil, err
	}

	return
}

// GetResourcesByFilters fetches a project of resources with a filter
func (vcls *VikunjaCaldavProjectStorage) GetResourcesByFilters(rpath string, filter *data.ResourceFilter) ([]data.Resource, error) {

	// If we already have a project saved, that means the user is making a REPORT request to find out if
	// anything changed, in that case we need to return all tasks.
	// That project is coming from a previous "getProjectRessource" in L177
	if vcls.project.Tasks != nil {
		var resources []data.Resource
		for i := range vcls.project.Tasks {
			rr := VikunjaProjectResourceAdapter{
				project:      vcls.project,
				task:         &vcls.project.Tasks[i].Task,
				isCollection: false,
			}
			r := data.NewResource(taskURL(vcls.project.ID, &vcls.project.Tasks[i].Task), &rr)
			r.Name = vcls.project.Tasks[i].Title
			if !matchesTodoFilter(filter, &r) {
				continue
			}
			resources = append(resources, r)
		}
		return resources, nil
	}

	// This is used to get all
	rr, err := vcls.getProjectRessource(false)
	if err != nil {
		return nil, err
	}
	r := data.NewResource(vcls.hrefFor(rpath), &rr)
	r.Name = vcls.project.Title
	return []data.Resource{r}, nil
	// For now, filtering is disabled.
	// return vcls.GetResources(rpath, false)
}

// caldav-go only evaluates time-range on VEVENTs, so a VTODO query with one would match
// nothing. Those keep getting every task, the client narrows them down itself.
func matchesTodoFilter(filter *data.ResourceFilter, r *data.Resource) bool {
	if filter == nil || filter.GetTimeRangeFilter() != nil {
		return true
	}
	return filter.Match(todoResource{r})
}

// caldav-go reports every non-collection resource as a VEVENT.
type todoResource struct {
	*data.Resource
}

func (todoResource) ComponentName() string {
	return lib.VTODO
}

// WebDAV requires every Depth: 1 child to live under the collection's own URI.
func taskURL(collectionProjectID int64, task *models.Task) string {
	return taskHref(collectionProjectID, task.UID)
}

func taskHref(collectionProjectID int64, uid string) string {
	return collectionHref(collectionProjectID) + `/` + encodeURIPathSegment(uid) + `.ics`
}

// url.PathEscape leaves & alone, and caldav-go writes hrefs into XML unescaped.
func encodeURIPathSegment(segment string) string {
	return strings.ReplaceAll(url.PathEscape(segment), "&", "%26")
}

// caldav-go builds rpath from the decoded request path, so echoing it back for a
// task resource would undo the encoding taskURL applied.
func (vcls *VikunjaCaldavProjectStorage) hrefFor(rpath string) string {
	if strings.HasSuffix(rpath, ".ics") && vcls.project != nil && vcls.task != nil && vcls.task.UID != "" {
		return taskURL(vcls.project.ID, vcls.task)
	}
	return rpath
}

// GetResource fetches a single resource
func (vcls *VikunjaCaldavProjectStorage) GetResource(rpath string) (*data.Resource, bool, error) {

	// If the task is not nil, we need to get the task and not the project
	if vcls.task != nil {
		s := db.NewSession()
		defer s.Close()

		task, err := findTask(s, vcls.user, vcls.collection(), vcls.task.UID)
		if err != nil {
			_ = s.Rollback()
			return nil, false, err
		}
		if task == nil {
			_ = s.Rollback()
			return nil, false, errs.ResourceNotFoundError
		}

		if err := s.Commit(); err != nil {
			return nil, false, err
		}

		vcls.task = task

		rr := VikunjaProjectResourceAdapter{
			project: vcls.project,
			task:    vcls.task,
		}
		r := data.NewResource(vcls.hrefFor(rpath), &rr)
		return &r, true, nil
	}

	// Otherwise get the project with all tasks
	rr, err := vcls.getProjectRessource(true)
	if err != nil {
		return nil, false, err
	}
	r := data.NewResource(rpath, &rr)
	return &r, true, nil
}

// GetShallowResource gets a resource without children
// Since Vikunja has no children, this is the same as GetResource
func (vcls *VikunjaCaldavProjectStorage) GetShallowResource(rpath string) (*data.Resource, bool, error) {
	// Since Vikunja has no children, this just returns the same as GetResource()
	// FIXME: This should just get the project with no tasks whatsoever, nothing else
	return vcls.GetResource(rpath)
}

// Only REPORT still reaches caldav-go.
var errNotThroughLibrary = errors.New("caldav: not served through caldav-go")

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

func addTaskResource(collectionProjectID int64, task *models.Task, rr *VikunjaProjectResourceAdapter, resources *[]data.Resource) {
	taskResourceInstance := data.NewResource(taskURL(collectionProjectID, task), rr)
	taskResourceInstance.Name = task.Title
	*resources = append(*resources, taskResourceInstance)
}
