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
	"context"
	"fmt"
	"slices"

	"code.vikunja.io/api/pkg/caldav"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/richtext"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// findTask resolves uid through c. Tasks the user cannot read or that are not in c are a
// miss, which is (nil, nil).
func findTask(s *xorm.Session, u *user.User, c collectionTarget, uid string) (*models.Task, error) {
	tasks, err := models.GetTasksByUIDs(s, []string{uid}, u)
	if err != nil {
		if models.IsErrTaskDoesNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, nil
	}

	task := tasks[0]
	if task.ProjectID == c.project.ID {
		return task, nil
	}
	// The URL project must match the task's real project (GHSA-48ch-p4gq-x46x).
	if !c.isPseudo() {
		return nil, nil
	}

	// tasks.uid has no unique index, so the first candidate may be a same-UID sibling.
	for _, candidate := range tasks {
		in, err := c.contains(s, u, candidate.ID)
		if err != nil {
			return nil, err
		}
		if in {
			return candidate, nil
		}
	}
	return nil, nil
}

// readCollectionTasks loads c.project and its tasks.
func readCollectionTasks(s *xorm.Session, u *user.User, c collectionTarget) ([]*models.TaskWithComments, error) {
	if _, err := c.checkRead(s, u); err != nil {
		return nil, err
	}
	if err := c.project.ReadOne(s, u); err != nil {
		return nil, err
	}

	if c.project.Tasks != nil {
		return c.project.Tasks, nil
	}

	tk := models.TaskCollection{ProjectID: c.project.ID}
	iface, _, _, err := tk.ReadAll(s, u, "", 0, -1)
	if err != nil {
		return nil, err
	}
	tasks, ok := iface.([]*models.Task)
	if !ok {
		return nil, fmt.Errorf("unexpected task collection result type %T", iface)
	}

	var projectTasks []*models.TaskWithComments
	for _, t := range tasks {
		projectTasks = append(projectTasks, &models.TaskWithComments{Task: *t})
	}
	c.project.Tasks = projectTasks
	return projectTasks, nil
}

// readableCollection is a collection the user was checked to be able to read.
type readableCollection struct {
	collectionTarget
}

func (c collectionTarget) checkRead(s *xorm.Session, u *user.User) (readableCollection, error) {
	can, err := c.canRead(s, u)
	if err != nil {
		return readableCollection{}, err
	}
	if !can {
		log.Debugf("User %v tried to access a caldav resource (Project %v) which they are not allowed to access", u.Username, c.project.ID)
		return readableCollection{}, errNotFound
	}
	return readableCollection{c}, nil
}

// writeTransaction commits fn's writes and dispatches their events, or drops both.
func writeTransaction(fn func(s *xorm.Session) error) error {
	s := db.NewSession()
	defer s.Close()

	if err := fn(s); err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return err
	}
	if err := s.Commit(); err != nil {
		events.CleanupPending(s)
		return err
	}
	events.DispatchPending(context.Background(), s)
	return nil
}

func parseTaskBody(content string) (*models.Task, caldav.ParsedVTODOProperties, error) {
	vTask, props, err := caldav.ParseTaskFromVTODO(content)
	if err != nil {
		log.Errorf("[CALDAV] Failed to parse VTODO: %v", err)
		log.Debugf("[CALDAV] VTODO content that failed to parse: %s", content)
		return nil, props, fmt.Errorf("%w: %w", errForbidden, err)
	}
	return vTask, props, nil
}

func createTask(s *xorm.Session, u *user.User, c readableCollection, vTask *models.Task) error {
	vTask.ProjectID = c.project.ID

	// A client PUTting a UID that already exists is syncing a stale href - typically after
	// the task moved to another project or out of this collection. Creating a second task
	// would silently fork it (#3482), and answering 403 would strand the client on a
	// permanent denial, so 404 and let it resync to the real collection.
	if vTask.UID != "" {
		existing, err := models.GetTasksByUIDs(s, []string{vTask.UID}, u)
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			log.Debugf("[CALDAV] Refusing to create task with already existing UID %s in project %d", vTask.UID, c.project.ID)
			return errNotFound
		}
	}

	canCreate, err := denyArchived(vTask.CanCreate(s, u))
	if err != nil {
		log.Errorf("[CALDAV] Permission check failed in createTask for user %s, project %d: %v", u.Username, c.project.ID, err)
		return err
	}
	if !canCreate {
		log.Warningf("[CALDAV] User %s does not have permission to create task in project %d", u.Username, c.project.ID)
		return errForbidden
	}

	// Inbound CalDAV descriptions are markdown; store them as canonical HTML.
	if err := applyDescriptionFromMarkdown(s, u, vTask, ""); err != nil {
		log.Errorf("[CALDAV] Failed to convert description in createTask: %v", err)
		return err
	}

	if err := vTask.Create(s, u); err != nil {
		log.Errorf("[CALDAV] Failed to create task in createTask: %v, task: %+v", err, vTask)
		return err
	}

	// Without labels of its own: UpdateTaskLabels diffs the new labels against them.
	created := &models.Task{ID: vTask.ID, ProjectID: vTask.ProjectID}
	if err := persistLabels(s, u, created, vTask.Labels); err != nil {
		log.Errorf("[CALDAV] Failed to persist labels in createTask: %v, labels: %+v", err, vTask.Labels)
		return err
	}
	if err := persistRelations(s, u, created, vTask.RelatedTasks); err != nil {
		log.Errorf("[CALDAV] Failed to persist relations in createTask: %v, relations: %+v", err, vTask.RelatedTasks)
		return err
	}
	return nil
}

// updateTask overlays what vTask expressed onto stored, which must come from findTask in s.
func updateTask(s *xorm.Session, u *user.User, c readableCollection, stored, vTask *models.Task, props caldav.ParsedVTODOProperties) (*models.Task, error) {
	if err := c.checkWrite(s, u, stored); err != nil {
		return nil, err
	}

	vTask.ID = stored.ID
	vTask.ProjectID = c.project.ID
	if c.isPseudo() {
		// A pseudo id is no project to move the task into; leave it where it lives.
		vTask.ProjectID = stored.ProjectID
	}

	canUpdate, err := denyArchived(vTask.CanUpdate(s, u))
	if err != nil {
		log.Errorf("[CALDAV] Permission check failed in updateTask for user %s, task %d: %v", u.Username, stored.ID, err)
		return nil, err
	}
	if !canUpdate {
		log.Warningf("[CALDAV] User %s does not have permission to update task %d", u.Username, stored.ID)
		return nil, errForbidden
	}

	// GHSA-48ch-p4gq-x46x: the collection must hold the task it writes to.
	in, err := c.holds(s, u, stored)
	if err != nil {
		return nil, err
	}
	if !in {
		return nil, errNotFound
	}

	// Inbound markdown → canonical HTML, kept verbatim when unchanged.
	if err := applyDescriptionFromMarkdown(s, u, vTask, stored.Description); err != nil {
		log.Errorf("[CALDAV] Failed to convert description in updateTask: %v", err)
		return nil, err
	}

	// Overlay only what the VTODO expressed; everything else keeps its stored value.
	base := *stored
	if props.Title {
		base.Title = vTask.Title
	}
	if props.Description {
		base.Description = vTask.Description
	}
	if props.Done {
		base.Done = vTask.Done
	}
	if props.Priority {
		base.Priority = vTask.Priority
	}
	if props.DueDate {
		base.DueDate = vTask.DueDate
	}
	if props.StartDate {
		base.StartDate = vTask.StartDate
	}
	if props.EndDate {
		base.EndDate = vTask.EndDate
	}
	if props.Color {
		base.HexColor = vTask.HexColor
	}
	// nil = no VALARM, keep stored reminders; non-nil (even empty) = client owns the alarm set.
	if vTask.Reminders != nil {
		base.Reminders = vTask.Reminders
	}

	if err := base.Update(s, u); err != nil {
		log.Errorf("[CALDAV] Failed to update task in updateTask: %v, task: %+v", err, base)
		return nil, err
	}

	// nil = CATEGORIES absent, leave labels alone; non-nil (even empty) = explicit set.
	if vTask.Labels != nil {
		if err := persistLabels(s, u, &base, vTask.Labels); err != nil {
			log.Errorf("[CALDAV] Failed to persist labels in updateTask: %v, labels: %+v", err, vTask.Labels)
			return nil, err
		}
		base.Labels = vTask.Labels
	}

	// RELATED-TO has no explicit-clear form, so relations can be added or replaced via CalDAV, never removed.
	if len(vTask.RelatedTasks) > 0 {
		if err := persistRelations(s, u, &base, vTask.RelatedTasks); err != nil {
			log.Errorf("[CALDAV] Failed to persist relations in updateTask: %v, relations: %+v", err, vTask.RelatedTasks)
			return nil, err
		}
		base.RelatedTasks = vTask.RelatedTasks
	}

	// base, not vTask: Update wrote back the Updated timestamp the ETag is built from.
	return &base, nil
}

func deleteTask(s *xorm.Session, u *user.User, c readableCollection, task *models.Task) error {
	if err := c.checkWrite(s, u, task); err != nil {
		return err
	}

	canDelete, err := denyArchived(task.CanDelete(s, u))
	if err != nil {
		return err
	}
	if !canDelete {
		return errForbidden
	}
	return task.Delete(s, u)
}

// applyDescriptionFromMarkdown converts a task's inbound CalDAV description
// (markdown) to canonical HTML, rebuilding @mentions. Unchanged markdown keeps the
// stored HTML verbatim, so a no-op read-modify-write doesn't churn it or move Updated.
func applyDescriptionFromMarkdown(s *xorm.Session, viewer *user.User, vTask *models.Task, storedHTML string) error {
	if !richtext.Changed(storedHTML, vTask.Description) {
		vTask.Description = storedHTML
		return nil
	}

	htmlDesc, err := richtext.MarkdownToHTMLWithMentions(s, viewer, vTask.Description)
	if err != nil {
		return err
	}
	vTask.Description = htmlDesc
	return nil
}

func persistLabels(s *xorm.Session, a web.Auth, task *models.Task, labels []*models.Label) (err error) {

	labelTitles := []string{}

	for _, label := range labels {
		labelTitles = append(labelTitles, label.Title)
	}

	u := &user.User{
		ID: a.GetID(),
	}

	// Using readall ensures the current user has the permission to see the labels they provided via caldav.
	existingLabels, _, _, err := models.GetLabelsForUser(s, u, labelTitles, 0, 0)
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
				// Skip forbidden relations without aborting the sync.
				canDelete, canDeleteErr := rel.CanDelete(s, a)
				if canDeleteErr != nil {
					return canDeleteErr
				}
				if !canDelete {
					continue
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

// Bounds the tasks a single VTODO can mint via unknown RELATED-TO UIDs.
const maxPlaceholdersPerVTODO = 50

// Persist new relations provided by the VTODO entry:
func persistRelations(s *xorm.Session, a web.Auth, task *models.Task, newRelations map[models.RelationKind][]*models.Task) (err error) {

	err = removeStaleRelations(s, a, task, newRelations)
	if err != nil {
		return err
	}

	// Inaccessible and missing UIDs must remain indistinguishable.
	allUIDs := make([]string, 0)
	dedupedRelations := make(map[models.RelationKind][]*models.Task, len(newRelations))
	for relationType, relatedTasksInVTODO := range newRelations {
		seen := make(map[string]bool, len(relatedTasksInVTODO))
		deduped := make([]*models.Task, 0, len(relatedTasksInVTODO))
		for _, relatedTaskInVTODO := range relatedTasksInVTODO {
			if seen[relatedTaskInVTODO.UID] {
				continue
			}
			seen[relatedTaskInVTODO.UID] = true
			deduped = append(deduped, relatedTaskInVTODO)
			allUIDs = append(allUIDs, relatedTaskInVTODO.UID)
		}
		dedupedRelations[relationType] = deduped
	}
	accessibleTasksByUID := make(map[string]*models.Task)
	if len(allUIDs) > 0 {
		accessibleTasks, err := models.GetTasksByUIDs(s, allUIDs, a)
		if err != nil {
			return err
		}
		for _, accessibleTask := range accessibleTasks {
			accessibleTasksByUID[accessibleTask.UID] = accessibleTask
		}
	}

	placeholders := 0

	// Ensure the current relations exist:
	for relationType, relatedTasksInVTODO := range dedupedRelations {
		// Persist each relation independently:
		for _, relatedTaskInVTODO := range relatedTasksInVTODO {

			var relatedTask *models.Task
			createDummy := false

			// Get the task from the DB:
			if accessible, has := accessibleTasksByUID[relatedTaskInVTODO.UID]; has {
				relatedTask = accessible
			} else {
				relatedTask = relatedTaskInVTODO
				createDummy = true
			}

			// Placeholders stay in the caller's project so CanCreate applies.
			if createDummy {
				// Dropped, not rejected: the client keeps syncing the rest of the VTODO.
				if placeholders >= maxPlaceholdersPerVTODO {
					continue
				}
				placeholders++
				relatedTask = &models.Task{
					ProjectID: task.ProjectID,
					Title:     "DUMMY-UID-" + relatedTaskInVTODO.UID,
					UID:       relatedTaskInVTODO.UID,
				}
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
			canCreate, canCreateErr := rel.CanCreate(s, a)
			if canCreateErr != nil {
				return canCreateErr
			}
			if !canCreate {
				continue
			}
			createErr := rel.Create(s, a)
			if createErr != nil && !models.IsErrRelationAlreadyExists(createErr) {
				return createErr
			}
		}
	}
	return nil
}
