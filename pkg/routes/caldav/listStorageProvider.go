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
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	user2 "code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"github.com/samedi/caldav-go/data"
	"github.com/samedi/caldav-go/errs"
	"xorm.io/xorm"
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
		r := data.NewResource(withTrailingSlash(rpath), &VikunjaProjectResourceAdapter{
			isPrincipal:  true,
			isCollection: true,
		})
		return []data.Resource{r}, nil
	}

	// If the request wants the principal url, we'll return that and nothing else
	if vcls.isPrincipal {
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

	if !withChildren {
		r := data.NewResource(withTrailingSlash(rpath), &VikunjaProjectResourceAdapter{
			isPrincipal:  true,
			isCollection: true,
		})
		return []data.Resource{r}, nil
	}

	// Bulk-query the latest task update time per project in a single DB call.
	// This is the key data iOS uses to decide whether to fetch new tasks:
	// if the ctag/etag for a calendar collection does not change, iOS skips it.
	// We open a fresh session here because the previous session's internal state
	// (from the complex CTE in ReadAll) is incompatible with further queries.
	latestTaskTimes := map[int64]time.Time{}
	if len(projects) > 0 {
		// Query the latest task update time per project by selecting the real
		// `updated` column (not a MAX aggregate). Aggregate functions return
		// values in DB-driver-specific formats that xorm cannot reliably map to
		// time.Time across all databases (e.g. PostgreSQL lib/pq returns a
		// time.Time which xorm then serialises as a Go string in a format not
		// matched by any known layout). Selecting the actual column lets xorm use
		// its native time.Time mapping on SQLite, PostgreSQL, and MySQL.
		// Sorting DESC and taking the first row per project in Go gives the max.
		type taskProjectUpdate struct {
			ProjectID int64     `xorm:"project_id"`
			Updated   time.Time `xorm:"updated"`
		}
		projectIDs := make([]int64, len(projects))
		for i, l := range projects {
			projectIDs[i] = l.ID
		}
		var taskUpdates []taskProjectUpdate
		s2 := db.NewSession()
		defer s2.Close()
		if err := s2.
			Table("tasks").
			Select("project_id, updated").
			In("project_id", projectIDs).
			Desc("updated").
			Find(&taskUpdates); err != nil {
			_ = s2.Rollback()
			return nil, err
		}
		if err := s2.Commit(); err != nil {
			return nil, err
		}
		for _, tu := range taskUpdates {
			if _, exists := latestTaskTimes[tu.ProjectID]; !exists {
				latestTaskTimes[tu.ProjectID] = tu.Updated
			}
		}

		// Also factor in the latest CalDAV deletion timestamp per project so that
		// the ctag advances when a task is deleted. Deleted tasks leave no task row
		// to bump max-updated, so without this iOS never sees the ctag change and
		// never sends a sync-collection REPORT to learn about the deletion.
		type deletionProjectTime struct {
			ProjectID int64     `xorm:"project_id"`
			DeletedAt time.Time `xorm:"deleted_at"`
		}
		var deletionTimes []deletionProjectTime
		s3 := db.NewSession()
		defer s3.Close()
		if err := s3.
			Table("task_caldav_deletions").
			Select("project_id, deleted_at").
			In("project_id", projectIDs).
			Desc("deleted_at").
			Find(&deletionTimes); err != nil {
			_ = s3.Rollback()
			return nil, err
		}
		if err := s3.Commit(); err != nil {
			return nil, err
		}
		for _, dt := range deletionTimes {
			if dt.DeletedAt.After(latestTaskTimes[dt.ProjectID]) {
				latestTaskTimes[dt.ProjectID] = dt.DeletedAt
			}
		}
	}

	var resources []data.Resource
	for _, l := range projects {
		rr := VikunjaProjectResourceAdapter{
			project: &models.ProjectWithTasksAndBuckets{
				Project: *l,
			},
			isCollection:     true,
			latestTaskUpdate: latestTaskTimes[l.ID],
		}
		r := data.NewResource(ProjectBasePath+"/"+strconv.FormatInt(l.ID, 10), &rr)
		r.Name = l.Title
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
		uid := strings.TrimSuffix(parts[4], ".ics")
		if uid == "" {
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
	err = s.Commit()
	if err != nil {
		return
	}

	for _, t := range tasks {
		// Closes the URL-path leak for calendar-multiget REPORTs
		// (GHSA-48ch-p4gq-x46x).
		if urlProjectID, ok := uidURLProjects[t.UID]; ok && urlProjectID != t.ProjectID {
			continue
		}
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

		// Reject reads where the URL project (set by TaskHandler in handler.go
		// from the :project param) doesn't match the task's real project
		// (GHSA-48ch-p4gq-x46x).
		if vcls.project != nil && vcls.project.ID != 0 && vcls.task.ProjectID != vcls.project.ID {
			return nil, false, errs.ResourceNotFoundError
		}

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
		log.Errorf("[CALDAV] Failed to parse VTODO in CreateResource: %v", err)
		log.Debugf("[CALDAV] VTODO content that failed to parse: %s", content)
		return nil, err
	}

	vTask.ProjectID = vcls.project.ID

	// Check the permissions
	canCreate, err := vTask.CanCreate(s, vcls.user)
	if err != nil {
		log.Errorf("[CALDAV] Permission check failed in CreateResource for user %s, project %d: %v", vcls.user.Username, vcls.project.ID, err)
		return nil, err
	}
	if !canCreate {
		log.Warningf("[CALDAV] User %s does not have permission to create task in project %d", vcls.user.Username, vcls.project.ID)
		return nil, errs.ForbiddenError
	}

	// Create the task
	err = vTask.Create(s, vcls.user)
	if err != nil {
		log.Errorf("[CALDAV] Failed to create task in CreateResource: %v, task: %+v", err, vTask)
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, err
	}

	vcls.task.ID = vTask.ID
	err = persistLabels(s, vcls.user, vcls.task, vTask.Labels)
	if err != nil {
		log.Errorf("[CALDAV] Failed to persist labels in CreateResource: %v, labels: %+v", err, vTask.Labels)
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, err
	}

	vcls.task.ProjectID = vcls.project.ID
	err = persistRelations(s, vcls.user, vcls.task, vTask.RelatedTasks)
	if err != nil {
		log.Errorf("[CALDAV] Failed to persist relations in CreateResource: %v, relations: %+v", err, vTask.RelatedTasks)
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, err
	}

	if err := s.Commit(); err != nil {
		log.Errorf("[CALDAV] Failed to commit transaction in CreateResource: %v", err)
		events.CleanupPending(s)
		return nil, err
	}

	events.DispatchPending(s)

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
		log.Errorf("[CALDAV] Failed to parse VTODO in UpdateResource: %v", err)
		log.Debugf("[CALDAV] VTODO content that failed to parse: %s", content)
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
		log.Errorf("[CALDAV] Permission check failed in UpdateResource for user %s, task %d: %v", vcls.user.Username, vcls.task.ID, err)
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, err
	}
	if !canUpdate {
		log.Warningf("[CALDAV] User %s does not have permission to update task %d", vcls.user.Username, vcls.task.ID)
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, errs.ForbiddenError
	}

	// Update the task
	err = vTask.Update(s, vcls.user)
	if err != nil {
		log.Errorf("[CALDAV] Failed to update task in UpdateResource: %v, task: %+v", err, vTask)
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, err
	}

	err = persistLabels(s, vcls.user, vcls.task, vTask.Labels)
	if err != nil {
		log.Errorf("[CALDAV] Failed to persist labels in UpdateResource: %v, labels: %+v", err, vTask.Labels)
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, err
	}

	err = persistRelations(s, vcls.user, vcls.task, vTask.RelatedTasks)
	if err != nil {
		log.Errorf("[CALDAV] Failed to persist relations in UpdateResource: %v, relations: %+v", err, vTask.RelatedTasks)
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, err
	}

	if err := s.Commit(); err != nil {
		log.Errorf("[CALDAV] Failed to commit transaction in UpdateResource: %v", err)
		events.CleanupPending(s)
		return nil, err
	}

	events.DispatchPending(s)

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
			events.CleanupPending(s)
			return err
		}
		if !canDelete {
			events.CleanupPending(s)
			return errs.ForbiddenError
		}

		// Delete it
		err = vcls.task.Delete(s, vcls.user)
		if err != nil {
			_ = s.Rollback()
			events.CleanupPending(s)
			return err
		}

		err = s.Commit()
		if err != nil {
			events.CleanupPending(s)
			return err
		}

		events.DispatchPending(s)
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

		// Only process CalDAV-compatible relation kinds (parenttask, subtask).
		// Other kinds (related, blocking, etc.) are never set via CalDAV and
		// should not be removed here.
		if relationKind != models.RelationKindParenttask && relationKind != models.RelationKindSubtask {
			continue
		}

		// For subtask relations: only consider removal if the VTODO explicitly
		// declares RELATED-TO;RELTYPE=CHILD (i.e., subtask kind is a key in
		// newRelations). Subtask relations are often auto-created as inverses
		// when child tasks declare RELATED-TO;RELTYPE=PARENT pointing to this
		// task. Removing them just because this task's VTODO doesn't mention
		// RELATED-TO;RELTYPE=CHILD would break those child-declared links.
		if relationKind == models.RelationKindSubtask {
			if _, hasSubtaskKind := newRelations[models.RelationKindSubtask]; !hasSubtaskKind {
				continue
			}
		}

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

	// latestTaskUpdate is the maximum Updated time across all tasks in the project.
	// It is populated by a bulk query during home-set PROPFIND to avoid loading
	// all tasks just to compute the collection ctag.
	latestTaskUpdate time.Time

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

	if vlra.task != nil {
		return `"` + strconv.FormatInt(vlra.task.ID, 10) + `-` + strconv.FormatInt(vlra.task.Updated.Unix(), 10) + `"`
	}

	if vlra.project == nil {
		return ""
	}

	// For collections, use the latest modification time across all tasks
	// so that the etag (and derived ctag/sync-token) changes whenever
	// any task in the project is added, modified, or deleted.
	latest := vlra.project.Updated
	if vlra.latestTaskUpdate.After(latest) {
		latest = vlra.latestTaskUpdate
	}
	for _, t := range vlra.projectTasks {
		if t.Updated.After(latest) {
			latest = t.Updated
		}
	}

	return `"` + strconv.FormatInt(vlra.project.ID, 10) + `-` + strconv.FormatInt(latest.Unix(), 10) + `"`
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
		latest := vlra.project.Updated
		if vlra.latestTaskUpdate.After(latest) {
			latest = vlra.latestTaskUpdate
		}
		for _, t := range vlra.projectTasks {
			if t.Updated.After(latest) {
				latest = t.Updated
			}
		}
		return latest
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

	// Ensure every task has a UID persisted in the database. Tasks created via
	// the web UI may have UID == "" if they pre-date UID generation in Create().
	// ParseTodos computes a fallback UID on-the-fly, but without a stored UID
	// RecordCaldavTaskDeletion skips the entry and the client never learns about
	// the deletion. Saving the UID here (on first CalDAV access) makes all
	// future deletions trackable.
	for _, t := range projectTasks {
		if t.UID == "" {
			t.UID = caldav.FallbackUID(t.Updated, t.Title)
			if _, err := s.ID(t.ID).Cols("uid").Update(&t.Task); err != nil {
				_ = s.Rollback()
				return rr, err
			}
		}
	}

	if err := s.Commit(); err != nil {
		return rr, err
	}

	// Query the latest deletion timestamp for this project so that a depth=0
	// PROPFIND on the project itself advances its ctag/sync-token when tasks
	// are deleted. Without this the deletion row in task_caldav_deletions is
	// invisible to this code path and iOS never sees the ctag change.
	type deletionTime struct {
		DeletedAt time.Time `xorm:"deleted_at"`
	}
	var dt []deletionTime
	s2 := db.NewSession()
	defer s2.Close()
	if err := s2.
		Table("task_caldav_deletions").
		Select("deleted_at").
		Where("project_id = ?", vcls.project.ID).
		Desc("deleted_at").
		Limit(1).
		Find(&dt); err != nil {
		_ = s2.Rollback()
		return rr, err
	}
	if err := s2.Commit(); err != nil {
		return rr, err
	}
	var latestDeletion time.Time
	if len(dt) > 0 {
		latestDeletion = dt[0].DeletedAt
	}

	rr = VikunjaProjectResourceAdapter{
		project:          vcls.project,
		projectTasks:     projectTasks,
		isCollection:     isCollection,
		latestTaskUpdate: latestDeletion,
	}

	return
}

func addTaskResource(task *models.Task, rr *VikunjaProjectResourceAdapter, resources *[]data.Resource) {
	taskResourceInstance := data.NewResource(getTaskURL(task), rr)
	taskResourceInstance.Name = task.Title
	*resources = append(*resources, taskResourceInstance)
}
