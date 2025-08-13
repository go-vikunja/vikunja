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
	"slices"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/caldav"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	user2 "code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"github.com/samedi/caldav-go/data"
	"github.com/samedi/caldav-go/errs"
	"xorm.io/xorm"
)

// DavBasePath is the base url path
const DavBasePath = `/dav/`

// ProjectBasePath is the base path for all projects resources
const ProjectBasePath = DavBasePath + `projects`

// VikunjaCaldavProjectStorage represents a project storage
type VikunjaCaldavProjectStorage struct {
	// Used when handling a project
	project *models.ProjectWithTasksAndBuckets
	// Used when handling a single task, like updating
	task *models.Task
	// The current user
	user        *user2.User
	isPrincipal bool
	isEntry     bool // Entry level handling should only return a link to the principal url
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

	if vcls.isEntry {
		r := data.NewResource(rpath, &VikunjaProjectResourceAdapter{
			isPrincipal:  true,
			isCollection: true,
		})
		return []data.Resource{r}, nil
	}

	// If the request wants the principal url, we'll return that and nothing else
	if vcls.isPrincipal {
		r := data.NewResource(DavBasePath+`/projects/`, &VikunjaProjectResourceAdapter{
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
		r := data.NewResource(rpath, &rr)
		r.Name = vcls.project.Title

		// If the request is withChildren (Depth: 1), we need to return all tasks of the project
		if withChildren {
			resources := []data.Resource{r}

			// Check if there are tasks to iterate over
			if vcls.project.Tasks != nil {
				for i := range vcls.project.Tasks {
					taskResource := VikunjaProjectResourceAdapter{
						project:      vcls.project,
						task:         &vcls.project.Tasks[i].Task,
						isCollection: false,
					}
					addTaskResource(&vcls.project.Tasks[i].Task, &taskResource, &resources)
				}
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
		resources = append(resources, r)
	}

	return resources, nil
}

// GetResourcesByList fetches a list of resources from a slice of paths
func (vcls *VikunjaCaldavProjectStorage) GetResourcesByList(rpaths []string) (resources []data.Resource, err error) {

	// Parse the set of resourcepaths into usable uids
	// A path looks like this: /dav/projects/10/a6eb526d5748a5c499da202fe74f36ed1aea2aef.ics
	// So we split the url in parts, take the last one and strip the ".ics" at the end
	var uids []string
	for _, path := range rpaths {
		parts := strings.Split(path, "/")
		if len(parts) < 5 {
			continue
		}
		uids = append(uids, strings.TrimSuffix(parts[4], ".ics"))
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
	err = s.Commit()
	if err != nil {
		return
	}

	for _, t := range tasks {
		rr := VikunjaProjectResourceAdapter{
			task: t,
		}
		addTaskResource(t, &rr, &resources)
	}

	return
}

// GetResourcesByFilters fetches a project of resources with a filter
func (vcls *VikunjaCaldavProjectStorage) GetResourcesByFilters(rpath string, _ *data.ResourceFilter) ([]data.Resource, error) {

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
			r := data.NewResource(getTaskURL(&vcls.project.Tasks[i].Task), &rr)
			r.Name = vcls.project.Tasks[i].Title
			resources = append(resources, r)
		}
		return resources, nil
	}

	// This is used to get all
	rr, err := vcls.getProjectRessource(false)
	if err != nil {
		return nil, err
	}
	r := data.NewResource(rpath, &rr)
	r.Name = vcls.project.Title
	return []data.Resource{r}, nil
	// For now, filtering is disabled.
	// return vcls.GetResources(rpath, false)
}

func getTaskURL(task *models.Task) string {
	return ProjectBasePath + "/" + strconv.FormatInt(task.ProjectID, 10) + `/` + task.UID + `.ics`
}

// GetResource fetches a single resource
func (vcls *VikunjaCaldavProjectStorage) GetResource(rpath string) (*data.Resource, bool, error) {

	// If the task is not nil, we need to get the task and not the project
	if vcls.task != nil {
		s := db.NewSession()
		defer s.Close()

		// save and override the updated unix date to not break any later etag checks
		updated := vcls.task.Updated
		tasks, err := models.GetTasksByUIDs(s, []string{vcls.task.UID}, vcls.user)
		if err != nil {
			_ = s.Rollback()
			if models.IsErrTaskDoesNotExist(err) {
				return nil, false, errs.ResourceNotFoundError
			}
			return nil, false, err
		}
		if err := s.Commit(); err != nil {
			return nil, false, err
		}

		if len(tasks) < 1 {
			return nil, false, errs.ResourceNotFoundError
		}
		vcls.task = tasks[0]

		if updated.Unix() > 0 {
			vcls.task.Updated = updated
		}

		rr := VikunjaProjectResourceAdapter{
			project: vcls.project,
			task:    vcls.task,
		}
		r := data.NewResource(rpath, &rr)
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

// CreateResource creates a new resource
func (vcls *VikunjaCaldavProjectStorage) CreateResource(rpath, content string) (*data.Resource, error) {

	s := db.NewSession()
	defer s.Close()

	vTask, err := caldav.ParseTaskFromVTODO(content)
	if err != nil {
		return nil, err
	}

	vTask.ProjectID = vcls.project.ID

	// Check the permissions
	canCreate, err := vTask.CanCreate(s, vcls.user)
	if err != nil {
		return nil, err
	}
	if !canCreate {
		return nil, errs.ForbiddenError
	}

	// Create the task
	err = vTask.Create(s, vcls.user)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	vcls.task.ID = vTask.ID
	err = persistLabels(s, vcls.user, vcls.task, vTask.Labels)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	vcls.task.ProjectID = vcls.project.ID
	err = persistRelations(s, vcls.user, vcls.task, vTask.RelatedTasks)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	if err := s.Commit(); err != nil {
		return nil, err
	}

	// Build up the proper response
	rr := VikunjaProjectResourceAdapter{
		project: vcls.project,
		task:    vTask,
	}
	r := data.NewResource(rpath, &rr)
	return &r, nil
}

// UpdateResource updates a resource
func (vcls *VikunjaCaldavProjectStorage) UpdateResource(rpath, content string) (*data.Resource, error) {

	vTask, err := caldav.ParseTaskFromVTODO(content)
	if err != nil {
		return nil, err
	}

	// At this point, we already have the right task in vcls.task, so we can use that ID directly
	vTask.ID = vcls.task.ID

	// Explicitly set the ProjectID in case the task now belongs to a different project:
	vTask.ProjectID = vcls.project.ID
	vcls.task.ProjectID = vcls.project.ID

	s := db.NewSession()
	defer s.Close()

	// Check the permissions
	canUpdate, err := vTask.CanUpdate(s, vcls.user)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}
	if !canUpdate {
		_ = s.Rollback()
		return nil, errs.ForbiddenError
	}

	// Update the task
	err = vTask.Update(s, vcls.user)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	err = persistLabels(s, vcls.user, vcls.task, vTask.Labels)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	err = persistRelations(s, vcls.user, vcls.task, vTask.RelatedTasks)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	if err := s.Commit(); err != nil {
		return nil, err
	}

	// Build up the proper response
	rr := VikunjaProjectResourceAdapter{
		project: vcls.project,
		task:    vTask,
	}
	r := data.NewResource(rpath, &rr)
	return &r, nil
}

// DeleteResource deletes a resource
func (vcls *VikunjaCaldavProjectStorage) DeleteResource(_ string) error {
	if vcls.task != nil {
		s := db.NewSession()
		defer s.Close()

		// Check the permissions
		canDelete, err := vcls.task.CanDelete(s, vcls.user)
		if err != nil {
			_ = s.Rollback()
			return err
		}
		if !canDelete {
			return errs.ForbiddenError
		}

		// Delete it
		err = vcls.task.Delete(s, vcls.user)
		if err != nil {
			_ = s.Rollback()
			return err
		}

		return s.Commit()
	}

	return nil
}

func persistLabels(s *xorm.Session, a web.Auth, task *models.Task, labels []*models.Label) (err error) {

	labelTitles := []string{}

	for _, label := range labels {
		labelTitles = append(labelTitles, label.Title)
	}

	u := &user2.User{
		ID: a.GetID(),
	}

	// Using readall ensures the current user has the permission to see the labels they provided via caldav.
	existingLabels, _, _, err := models.GetLabelsByTaskIDs(s, &models.LabelByTaskIDsOptions{
		Search:              labelTitles,
		User:                u,
		GetForUser:          true,
		GetUnusedLabels:     true,
		GroupByLabelIDsOnly: true,
	})
	if err != nil {
		return err
	}

	labelMap := make(map[string]*models.Label)
	for i := range existingLabels {
		labelMap[existingLabels[i].Title] = &existingLabels[i].Label
	}

	for _, label := range labels {
		if l, has := labelMap[label.Title]; has {
			*label = *l
			continue
		}

		err = label.Create(s, a)
		if err != nil {
			return err
		}
	}

	// Create the label <-> task relation
	return task.UpdateTaskLabels(s, a, labels)
}

func removeStaleRelations(s *xorm.Session, a web.Auth, task *models.Task, newRelations map[models.RelationKind][]*models.Task) (err error) {

	// Get the existing task with details:
	existingTask := &models.Task{ID: task.ID}
	// FIXME: Optimize to get only required attributes (ie. RelatedTasks).
	err = existingTask.ReadOne(s, a)
	if err != nil {
		return
	}

	for relationKind, relatedTasks := range existingTask.RelatedTasks {

		for _, relatedTask := range relatedTasks {
			relationInNewList := slices.ContainsFunc(newRelations[relationKind], func(newRelation *models.Task) bool { return newRelation.UID == relatedTask.UID })

			if !relationInNewList {
				rel := models.TaskRelation{
					TaskID:       task.ID,
					OtherTaskID:  relatedTask.ID,
					RelationKind: relationKind,
				}
				err = rel.Delete(s, a)
				if err != nil {
					return
				}
			}
		}
	}

	return
}

// Persist new relations provided by the VTODO entry:
func persistRelations(s *xorm.Session, a web.Auth, task *models.Task, newRelations map[models.RelationKind][]*models.Task) (err error) {

	err = removeStaleRelations(s, a, task, newRelations)
	if err != nil {
		return err
	}

	// Ensure the current relations exist:
	for relationType, relatedTasksInVTODO := range newRelations {
		// Persist each relation independently:
		for _, relatedTaskInVTODO := range relatedTasksInVTODO {

			var relatedTask *models.Task
			createDummy := false

			// Get the task from the DB:
			relatedTaskInDB, err := models.GetTaskSimpleByUUID(s, relatedTaskInVTODO.UID)
			if err != nil {
				relatedTask = relatedTaskInVTODO
				createDummy = true
			} else {
				relatedTask = relatedTaskInDB
			}

			// If the related task doesn't exist, create a dummy one now in the same list.
			// It'll probably be populated right after in a following request.
			// In the worst case, this was an error by the client and we are left with
			// this dummy task to clean up.
			if createDummy {
				relatedTask.ProjectID = task.ProjectID
				relatedTask.Title = "DUMMY-UID-" + relatedTask.UID
				err = relatedTask.Create(s, a)
				if err != nil {
					return err
				}
			}

			// Create the relation:
			rel := models.TaskRelation{
				TaskID:       task.ID,
				OtherTaskID:  relatedTask.ID,
				RelationKind: relationType,
			}
			err = rel.Create(s, a)
			if err != nil && !models.IsErrRelationAlreadyExists(err) {
				return err
			}
		}
	}
	return err
}

// VikunjaProjectResourceAdapter holds the actual resource
type VikunjaProjectResourceAdapter struct {
	project      *models.ProjectWithTasksAndBuckets
	projectTasks []*models.TaskWithComments
	task         *models.Task

	isPrincipal  bool
	isCollection bool
}

// IsCollection checks if the resoure in the adapter is a collection
func (vlra *VikunjaProjectResourceAdapter) IsCollection() bool {
	// If the discovery does not work, setting this to true makes it work again.
	return vlra.isCollection
}

// CalculateEtag returns the etag of a resource
func (vlra *VikunjaProjectResourceAdapter) CalculateEtag() string {

	// If we're updating a task, the client sends the etag of the project instead of the one from the task.
	// And therefore, updating the task fails since these etags don't match.
	// To fix that, we use this extra field to determine if we're currently updating a task and return the
	// etag of the project instead.
	// if vlra.project != nil {
	//	 return `"` + strconv.FormatInt(vlra.project.ID, 10) + `-` + strconv.FormatInt(vlra.project.Updated, 10) + `"`
	// }

	// Return the etag of a task if we have one
	if vlra.task != nil {
		return `"` + strconv.FormatInt(vlra.task.ID, 10) + `-` + strconv.FormatInt(vlra.task.Updated.Unix(), 10) + `"`
	}

	if vlra.project == nil {
		return ""
	}

	// This also returns the etag of the project, and not of the task,
	// which becomes problematic because the client uses this etag (= the one from the project) to make
	// Requests to update a task. These do not match and thus updating a task fails.
	return `"` + strconv.FormatInt(vlra.project.ID, 10) + `-` + strconv.FormatInt(vlra.project.Updated.Unix(), 10) + `"`
}

// GetContent returns the content string of a resource (a task in our case)
func (vlra *VikunjaProjectResourceAdapter) GetContent() string {
	if vlra.project != nil && vlra.projectTasks != nil {
		return caldav.GetCaldavTodosForTasks(vlra.project, vlra.projectTasks)
	}

	if vlra.task != nil {
		project := models.ProjectWithTasksAndBuckets{Tasks: []*models.TaskWithComments{{Task: *vlra.task}}}
		return caldav.GetCaldavTodosForTasks(&project, project.Tasks)
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

	can, _, err := vcls.project.CanRead(s, vcls.user)
	if err != nil {
		_ = s.Rollback()
		return
	}
	if !can {
		_ = s.Rollback()
		log.Errorf("User %v tried to access a caldav resource (Project %v) which they are not allowed to access", vcls.user.Username, vcls.project.ID)
		return rr, models.ErrUserDoesNotHaveAccessToProject{ProjectID: vcls.project.ID, UserID: vcls.user.ID}
	}
	err = vcls.project.ReadOne(s, vcls.user)
	if err != nil {
		_ = s.Rollback()
		return
	}

	projectTasks := vcls.project.Tasks
	if projectTasks == nil {
		tk := models.TaskCollection{
			ProjectID: vcls.project.ID,
		}
		iface, _, _, err := tk.ReadAll(s, vcls.user, "", 0, -1)
		if err != nil {
			_ = s.Rollback()
			return rr, err
		}
		tasks, ok := iface.([]*models.Task)
		if !ok {
			panic("Tasks returned from TaskCollection.ReadAll are not []*models.Task!")
		}

		for _, t := range tasks {
			projectTasks = append(projectTasks, &models.TaskWithComments{Task: *t})
		}
		vcls.project.Tasks = projectTasks
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

func addTaskResource(task *models.Task, rr *VikunjaProjectResourceAdapter, resources *[]data.Resource) {
	taskResourceInstance := data.NewResource(getTaskURL(task), rr)
	taskResourceInstance.Name = task.Title
	*resources = append(*resources, taskResourceInstance)
}
