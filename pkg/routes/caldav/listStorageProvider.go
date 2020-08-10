// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package caldav

import (
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	user2 "code.vikunja.io/api/pkg/user"
	"github.com/samedi/caldav-go/data"
	"github.com/samedi/caldav-go/errs"
	"strconv"
	"strings"
	"time"
)

// DavBasePath is the base url path
const DavBasePath = `/dav/`

// ListBasePath is the base path for all lists resources
const ListBasePath = DavBasePath + `lists`

// VikunjaCaldavListStorage represents a list storage
type VikunjaCaldavListStorage struct {
	// Used when handling a list
	list *models.List
	// Used when handling a single task, like updating
	task *models.Task
	// The current user
	user        *user2.User
	isPrincipal bool
	isEntry     bool // Entry level handling should only return a link to the principal url
}

// GetResources returns either all lists, links to the principal, or only one list, depending on the request
func (vcls *VikunjaCaldavListStorage) GetResources(rpath string, withChildren bool) ([]data.Resource, error) {

	// It looks like we need to have the same handler for returning both the calendar home set and the user principal
	// Since the client seems to ignore the whatever is being returned in the first request and just makes a second one
	// to the same url but requesting the calendar home instead
	// The problem with this is caldav-go just return whatever ressource it gets and making that the requested path
	// And for us here, there is no easy (I can think of at least one hacky way) to figure out if the client is requesting
	// the home or principal url. Ough.

	// Ok, maybe the problem is more the client making a request to /dav/ and getting a response which says
	// something like "hey, for /dav/lists, the calendar home is /dav/lists", but the client expects a
	// response to go something like "hey, for /dav/, the calendar home is /dav/lists" since it requested /dav/
	// and not /dav/lists. I'm not sure if thats a bug in the client or in caldav-go.

	if vcls.isEntry {
		r := data.NewResource(rpath, &VikunjaListResourceAdapter{
			isPrincipal:  true,
			isCollection: true,
		})
		return []data.Resource{r}, nil
	}

	// If the request wants the principal url, we'll return that and nothing else
	if vcls.isPrincipal {
		r := data.NewResource(DavBasePath+`/lists/`, &VikunjaListResourceAdapter{
			isPrincipal:  true,
			isCollection: true,
		})
		return []data.Resource{r}, nil
	}

	// If vcls.list.ID is != 0, this means the user is doing a PROPFIND request to /lists/:list
	// Which means we need to get only one list
	if vcls.list != nil && vcls.list.ID != 0 {
		rr, err := vcls.getListRessource(true)
		if err != nil {
			return nil, err
		}
		r := data.NewResource(rpath, &rr)
		r.Name = vcls.list.Title
		return []data.Resource{r}, nil
	}

	// Otherwise get all lists
	thelists, _, _, err := vcls.list.ReadAll(vcls.user, "", -1, 50)
	if err != nil {
		return nil, err
	}
	lists := thelists.([]*models.List)

	var resources []data.Resource
	for _, l := range lists {
		rr := VikunjaListResourceAdapter{
			list:         l,
			isCollection: true,
		}
		r := data.NewResource(ListBasePath+"/"+strconv.FormatInt(l.ID, 10), &rr)
		r.Name = l.Title
		resources = append(resources, r)
	}

	return resources, nil
}

// GetResourcesByList fetches a list of resources from a slice of paths
func (vcls *VikunjaCaldavListStorage) GetResourcesByList(rpaths []string) ([]data.Resource, error) {

	// Parse the set of resourcepaths into usable uids
	// A path looks like this: /dav/lists/10/a6eb526d5748a5c499da202fe74f36ed1aea2aef.ics
	// So we split the url in parts, take the last one and strip the ".ics" at the end
	var uids []string
	for _, path := range rpaths {
		parts := strings.Split(path, "/")
		uid := []rune(parts[4]) // The 4th part is the id with ".ics" suffix
		endlen := len(uid) - 4  // ".ics" are 4 bytes
		uids = append(uids, string(uid[:endlen]))
	}

	// GetTasksByUIDs...
	// Parse these into ressources...
	tasks, err := models.GetTasksByUIDs(uids)
	if err != nil {
		return nil, err
	}

	var resources []data.Resource
	for _, t := range tasks {
		rr := VikunjaListResourceAdapter{
			task: t,
		}
		r := data.NewResource(getTaskURL(t), &rr)
		r.Name = t.Title
		resources = append(resources, r)
	}

	return resources, nil
}

// GetResourcesByFilters fetches a list of resources with a filter
func (vcls *VikunjaCaldavListStorage) GetResourcesByFilters(rpath string, filters *data.ResourceFilter) ([]data.Resource, error) {

	// If we already have a list saved, that means the user is making a REPORT request to find out if
	// anything changed, in that case we need to return all tasks.
	// That list is coming from a previous "getListRessource" in L177
	if vcls.list.Tasks != nil {
		var resources []data.Resource
		for _, t := range vcls.list.Tasks {
			rr := VikunjaListResourceAdapter{
				list:         vcls.list,
				task:         t,
				isCollection: false,
			}
			r := data.NewResource(getTaskURL(t), &rr)
			r.Name = t.Title
			resources = append(resources, r)
		}
		return resources, nil
	}

	// This is used to get all
	rr, err := vcls.getListRessource(false)
	if err != nil {
		return nil, err
	}
	r := data.NewResource(rpath, &rr)
	r.Name = vcls.list.Title
	return []data.Resource{r}, nil
	// For now, filtering is disabled.
	//return vcls.GetResources(rpath, false)
}

func getTaskURL(task *models.Task) string {
	return ListBasePath + "/" + strconv.FormatInt(task.ListID, 10) + `/` + task.UID + `.ics`
}

// GetResource fetches a single resource
func (vcls *VikunjaCaldavListStorage) GetResource(rpath string) (*data.Resource, bool, error) {

	// If the task is not nil, we need to get the task and not the list
	if vcls.task != nil {
		// save and override the updated unix date to not break any later etag checks
		updated := vcls.task.Updated
		task, err := models.GetTaskSimple(&models.Task{ID: vcls.task.ID, UID: vcls.task.UID})
		if err != nil {
			if models.IsErrTaskDoesNotExist(err) {
				return nil, false, errs.ResourceNotFoundError
			}
			return nil, false, err
		}

		vcls.task = &task
		if updated.Unix() > 0 {
			vcls.task.Updated = updated
		}

		rr := VikunjaListResourceAdapter{
			list: vcls.list,
			task: &task,
		}
		r := data.NewResource(rpath, &rr)
		return &r, true, nil
	}

	// Otherwise get the list with all tasks
	rr, err := vcls.getListRessource(true)
	if err != nil {
		return nil, false, err
	}
	r := data.NewResource(rpath, &rr)
	return &r, true, nil
}

// GetShallowResource gets a ressource without childs
// Since Vikunja has no children, this is the same as GetResource
func (vcls *VikunjaCaldavListStorage) GetShallowResource(rpath string) (*data.Resource, bool, error) {
	// Since Vikunja has no childs, this just returns the same as GetResource()
	// FIXME: This should just get the list with no tasks whatsoever, nothing else
	return vcls.GetResource(rpath)
}

// CreateResource creates a new resource
func (vcls *VikunjaCaldavListStorage) CreateResource(rpath, content string) (*data.Resource, error) {

	vTask, err := parseTaskFromVTODO(content)
	if err != nil {
		return nil, err
	}

	vTask.ListID = vcls.list.ID

	// Check the rights
	canCreate, err := vTask.CanCreate(vcls.user)
	if err != nil {
		return nil, err
	}
	if !canCreate {
		return nil, errs.ForbiddenError
	}

	// Create the task
	err = vTask.Create(vcls.user)
	if err != nil {
		return nil, err
	}

	// Build up the proper response
	rr := VikunjaListResourceAdapter{
		list: vcls.list,
		task: vTask,
	}
	r := data.NewResource(rpath, &rr)
	return &r, nil
}

// UpdateResource updates a resource
func (vcls *VikunjaCaldavListStorage) UpdateResource(rpath, content string) (*data.Resource, error) {

	vTask, err := parseTaskFromVTODO(content)
	if err != nil {
		return nil, err
	}

	// At this point, we already have the right task in vcls.task, so we can use that ID directly
	vTask.ID = vcls.task.ID

	// Check the rights
	canUpdate, err := vTask.CanUpdate(vcls.user)
	if err != nil {
		return nil, err
	}
	if !canUpdate {
		return nil, errs.ForbiddenError
	}

	// Update the task
	err = vTask.Update()
	if err != nil {
		return nil, err
	}

	// Build up the proper response
	rr := VikunjaListResourceAdapter{
		list: vcls.list,
		task: vTask,
	}
	r := data.NewResource(rpath, &rr)
	return &r, nil
}

// DeleteResource deletes a resource
func (vcls *VikunjaCaldavListStorage) DeleteResource(rpath string) error {
	if vcls.task != nil {
		// Check the rights
		canDelete, err := vcls.task.CanDelete(vcls.user)
		if err != nil {
			return err
		}
		if !canDelete {
			return errs.ForbiddenError
		}

		// Delete it
		return vcls.task.Delete()
	}

	return nil
}

// VikunjaListResourceAdapter holds the actual resource
type VikunjaListResourceAdapter struct {
	list *models.List
	task *models.Task

	isPrincipal  bool
	isCollection bool
}

// IsCollection checks if the resoure in the adapter is a collection
func (vlra *VikunjaListResourceAdapter) IsCollection() bool {
	// If the discovery does not work, setting this to true makes it work again.
	return vlra.isCollection
}

// CalculateEtag returns the etag of a resource
func (vlra *VikunjaListResourceAdapter) CalculateEtag() string {

	// If we're updating a task, the client sends the etag of the list instead of the one from the task.
	// And therefore, updating the task fails since these etags don't match.
	// To fix that, we use this extra field to determine if we're currently updating a task and return the
	// etag of the list instead.
	//if vlra.list != nil {
	//	return `"` + strconv.FormatInt(vlra.list.ID, 10) + `-` + strconv.FormatInt(vlra.list.Updated, 10) + `"`
	//}

	// Return the etag of a task if we have one
	if vlra.task != nil {
		return `"` + strconv.FormatInt(vlra.task.ID, 10) + `-` + strconv.FormatInt(vlra.task.Updated.Unix(), 10) + `"`
	}
	// This also returns the etag of the list, and not of the task,
	// which becomes problematic because the client uses this etag (= the one from the list) to make
	// Requests to update a task. These do not match and thus updating a task fails.
	return `"` + strconv.FormatInt(vlra.list.ID, 10) + `-` + strconv.FormatInt(vlra.list.Updated.Unix(), 10) + `"`
}

// GetContent returns the content string of a resource (a task in our case)
func (vlra *VikunjaListResourceAdapter) GetContent() string {
	if vlra.list != nil && vlra.list.Tasks != nil {
		return getCaldavTodosForTasks(vlra.list)
	}

	if vlra.task != nil {
		list := models.List{Tasks: []*models.Task{vlra.task}}
		return getCaldavTodosForTasks(&list)
	}

	return ""
}

// GetContentSize is the size of a caldav content
func (vlra *VikunjaListResourceAdapter) GetContentSize() int64 {
	return int64(len(vlra.GetContent()))
}

// GetModTime returns when the resource was last modified
func (vlra *VikunjaListResourceAdapter) GetModTime() time.Time {
	if vlra.task != nil {
		return vlra.task.Updated
	}

	if vlra.list != nil {
		return vlra.list.Updated
	}

	return time.Time{}
}

func (vcls *VikunjaCaldavListStorage) getListRessource(isCollection bool) (rr VikunjaListResourceAdapter, err error) {
	can, _, err := vcls.list.CanRead(vcls.user)
	if err != nil {
		return
	}
	if !can {
		log.Errorf("User %v tried to access a caldav resource (List %v) which they are not allowed to access", vcls.user.Username, vcls.list.ID)
		return rr, models.ErrUserDoesNotHaveAccessToList{ListID: vcls.list.ID}
	}
	err = vcls.list.ReadOne()
	if err != nil {
		return
	}

	rr = VikunjaListResourceAdapter{
		list:         vcls.list,
		isCollection: isCollection,
	}

	return
}
