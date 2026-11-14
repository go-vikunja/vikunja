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

	"code.vikunja.io/api/pkg/caldav"
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
	user   *user2.User
	target target
}

func (vcls *VikunjaCaldavProjectStorage) collection() collectionTarget {
	return collectionTarget{project: vcls.project}
}

// GetResources returns either all projects, links to the principal, or only one project, depending on the request
func (vcls *VikunjaCaldavProjectStorage) GetResources(rpath string, withChildren bool) ([]data.Resource, error) {

	// It looks like we need to have the same handler for returning both the calendar home set and the user principal
	// Since the client seems to ignore the whatever is being returned in the first request and just makes a second one
	// to the same url but requesting the calendar home instead
	// The problem with this is caldav-go just return whatever resource it gets and making that the requested path
	// And for us here, there is no easy (I can think of at least one hacky way) to figure out if the client is requesting
	// the home or principal url. Ough.

	// Ok, maybe the problem is more the client making a request to /dav/ and getting a response which says
	// something like "hey, for /dav/projects, the calendar home is /dav/projects", but the client expects a
	// response to go something like "hey, for /dav/, the calendar home is /dav/projects" since it requested /dav/
	// and not /dav/projects. I'm not sure if thats a bug in the client or in caldav-go.

	switch vcls.target.(type) {
	case entryTarget:
		r := data.NewResource(withTrailingSlash(rpath), &VikunjaProjectResourceAdapter{
			isPrincipal:  true,
			isCollection: true,
		})
		return []data.Resource{r}, nil
	case principalTarget:
		// The principal answers with the home set and nothing else.
		r := data.NewResource(ProjectHomeSetPath, &VikunjaProjectResourceAdapter{
			isPrincipal:  true,
			isCollection: true,
		})
		return []data.Resource{r}, nil
	}

	// If vcls.project.ID is != 0, this means the user is doing a PROPFIND request to /projects/:project
	// Which means we need to get only one project
	if vcls.project != nil && vcls.project.ID != 0 {
		rr, err := vcls.getProjectRessource(true)
		if err != nil {
			return nil, err
		}
		r := data.NewResource(vcls.hrefFor(rpath), &rr)
		r.Name = vcls.project.Title
		r.Color = caldav.AppleColor(vcls.project.HexColor)

		// If the request is withChildren (Depth: 1), we need to return all tasks of the project
		if withChildren {
			resources := []data.Resource{r}

			emitted := make(map[string]bool, len(vcls.project.Tasks))
			for i := range vcls.project.Tasks {
				task := &vcls.project.Tasks[i].Task
				// tasks.uid has no unique index, so two same-UID tasks would map to one href.
				href := taskURL(vcls.project.ID, task)
				if emitted[href] {
					continue
				}
				emitted[href] = true

				taskResource := VikunjaProjectResourceAdapter{
					project:      vcls.project,
					task:         task,
					isCollection: false,
				}
				addTaskResource(vcls.project.ID, task, &taskResource, &resources)
			}
			return resources, nil
		}
		return []data.Resource{r}, nil
	}

	s := db.NewSession()
	defer s.Close()

	// Otherwise get all projects
	theprojects, _, _, err := vcls.project.ReadAll(s, vcls.user, "", -1, 50)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}
	if err := s.Commit(); err != nil {
		return nil, err
	}
	projects := theprojects.([]*models.Project)

	if !withChildren {
		r := data.NewResource(withTrailingSlash(rpath), &VikunjaProjectResourceAdapter{
			isPrincipal:  true,
			isCollection: true,
		})
		return []data.Resource{r}, nil
	}

	var resources []data.Resource
	for _, l := range projects {
		rr := VikunjaProjectResourceAdapter{
			project: &models.ProjectWithTasksAndBuckets{
				Project: *l,
			},
			isCollection: true,
		}
		r := data.NewResource(ProjectBasePath+"/"+strconv.FormatInt(l.ID, 10), &rr)
		r.Name = l.Title
		r.Color = caldav.AppleColor(l.HexColor)
		resources = append(resources, r)
	}

	return resources, nil
}

func withTrailingSlash(path string) string {
	if path == "" {
		return ProjectHomeSetPath
	}
	if strings.HasSuffix(path, "/") {
		return path
	}
	return path + "/"
}

func principalPathForUser(username string) string {
	return withTrailingSlash(PrincipalBasePath + `/` + username)
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
	return ProjectBasePath + "/" + strconv.FormatInt(collectionProjectID, 10) + `/` + encodeURIPathSegment(task.UID) + `.ics`
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

		// save and override the updated unix date to not break any later etag checks
		updated := vcls.task.Updated
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

		if updated.Unix() > 0 {
			vcls.task.Updated = updated
		}

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

// Writes never reach caldav-go: PUT and DELETE are served in-house.
var errWriteThroughLibrary = errors.New("caldav: writes do not go through caldav-go")

// CreateResource only satisfies data.Storage.
func (vcls *VikunjaCaldavProjectStorage) CreateResource(string, string) (*data.Resource, error) {
	return nil, errWriteThroughLibrary
}

// UpdateResource only satisfies data.Storage.
func (vcls *VikunjaCaldavProjectStorage) UpdateResource(string, string) (*data.Resource, error) {
	return nil, errWriteThroughLibrary
}

// DeleteResource only satisfies data.Storage.
func (vcls *VikunjaCaldavProjectStorage) DeleteResource(string) error {
	return errWriteThroughLibrary
}

// VikunjaProjectResourceAdapter holds the actual resource
type VikunjaProjectResourceAdapter struct {
	project      *models.ProjectWithTasksAndBuckets
	projectTasks []*models.TaskWithComments
	task         *models.Task

	isPrincipal  bool
	isCollection bool

	// Rendered once per request: a multistatus response asks for it per requested property.
	content *string
}

// IsCollection checks if the resoure in the adapter is a collection
func (vlra *VikunjaProjectResourceAdapter) IsCollection() bool {
	// If the discovery does not work, setting this to true makes it work again.
	return vlra.isCollection
}

// IsCalendar is false for the home sets, which only contain the project calendars.
func (vlra *VikunjaProjectResourceAdapter) IsCalendar() bool {
	return !vlra.isPrincipal
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

	projectTasks, err := readCollectionTasks(s, vcls.user, vcls.collection())
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
