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

package models

import (
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// TaskAssginee represents an assignment of a user to a task
type TaskAssginee struct {
	ID      int64     `xorm:"bigint autoincr not null unique pk" json:"-"`
	TaskID  int64     `xorm:"bigint INDEX not null" json:"-" param:"projecttask"`
	UserID  int64     `xorm:"bigint INDEX not null" json:"user_id" param:"user"`
	Created time.Time `xorm:"created not null"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// TableName makes a pretty table name
func (*TaskAssginee) TableName() string {
	return "task_assignees"
}

// TaskAssigneeWithUser is a helper type to deal with user joins
type TaskAssigneeWithUser struct {
	TaskID    int64
	user.User `xorm:"extends"`
}

func getRawTaskAssigneesForTasks(s *xorm.Session, taskIDs []int64) (taskAssignees []*TaskAssigneeWithUser, err error) {
	taskAssignees = []*TaskAssigneeWithUser{}
	err = s.Table("task_assignees").
		Select("task_id, users.*").
		In("task_id", taskIDs).
		Join("INNER", "users", "task_assignees.user_id = users.id").
		Find(&taskAssignees)
	return
}

// Create or update a bunch of task assignees
func (t *Task) updateTaskAssignees(s *xorm.Session, assignees []*user.User, doer web.Auth) (err error) {

	// Load the current assignees
	currentAssignees, err := getRawTaskAssigneesForTasks(s, []int64{t.ID})
	if err != nil {
		return err
	}

	t.Assignees = make([]*user.User, 0, len(currentAssignees))
	for i := range currentAssignees {
		t.Assignees = append(t.Assignees, &currentAssignees[i].User)
	}

	// If we don't have any new assignees, delete everything right away. Saves us some hassle.
	if len(assignees) == 0 && len(t.Assignees) > 0 {
		_, err = s.Where("task_id = ?", t.ID).
			Delete(&TaskAssginee{})
		t.setTaskAssignees(assignees)
		return err
	}

	// If we didn't change anything (from 0 to zero) don't do anything.
	if len(assignees) == 0 && len(t.Assignees) == 0 {
		return nil
	}

	// Make a hashmap of the new assignees for easier comparison
	newAssignees := make(map[int64]*user.User, len(assignees))
	for _, newAssignee := range assignees {
		newAssignees[newAssignee.ID] = newAssignee
	}

	// Get old assignees to delete
	var found bool
	var assigneesToDelete []int64
	oldAssignees := make(map[int64]*user.User, len(t.Assignees))
	for _, oldAssignee := range t.Assignees {
		found = false
		if newAssignees[oldAssignee.ID] != nil {
			found = true // If a new assignee is already in the project with old assignees
		}

		// Put all assignees which are only on the old project to the trash
		if !found {
			assigneesToDelete = append(assigneesToDelete, oldAssignee.ID)
		}

		oldAssignees[oldAssignee.ID] = oldAssignee
	}

	// Delete all assignees not passed
	if len(assigneesToDelete) > 0 {
		_, err = s.In("user_id", assigneesToDelete).
			And("task_id = ?", t.ID).
			Delete(&TaskAssginee{})
		if err != nil {
			return err
		}
	}

	// Get the project to perform later checks
	project, err := GetProjectSimpleByID(s, t.ProjectID)
	if err != nil {
		return
	}

	// Loop through our users and add them
	for _, u := range assignees {
		// Check if the user is already assigned and assign him only if not
		if oldAssignees[u.ID] != nil {
			// continue outer loop
			continue
		}

		// Add the new assignee
		err = t.addNewAssigneeByID(s, u.ID, project, doer)
		if err != nil {
			return err
		}
	}

	t.setTaskAssignees(assignees)

	err = updateProjectLastUpdated(s, &Project{ID: t.ProjectID})
	return
}

// Small helper functions to set the new assignees in various places
func (t *Task) setTaskAssignees(assignees []*user.User) {
	if len(assignees) == 0 {
		t.Assignees = nil
		return
	}
	t.Assignees = assignees
}

// Delete a task assignee
// @Summary Delete an assignee
// @Description Un-assign a user from a task.
// @tags assignees
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param taskID path int true "Task ID"
// @Param userID path int true "Assignee user ID"
// @Success 200 {object} models.Message "The assignee was successfully deleted."
// @Failure 403 {object} web.HTTPError "Not allowed to delete the assignee."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/assignees/{userID} [delete]
func (la *TaskAssginee) Delete(s *xorm.Session, a web.Auth) (err error) {
	_, err = s.Delete(&TaskAssginee{TaskID: la.TaskID, UserID: la.UserID})
	if err != nil {
		return err
	}

	err = updateProjectByTaskID(s, la.TaskID)
	if err != nil {
		return err
	}

	doer, _ := user.GetFromAuth(a)
	task, err := GetTaskByIDSimple(s, la.TaskID)
	if err != nil {
		return err
	}

	err = events.Dispatch(&TaskAssigneeDeletedEvent{
		Task:     &task,
		Assignee: &user.User{ID: la.UserID},
		Doer:     doer,
	})
	if err != nil {
		return err
	}
	return events.Dispatch(&TaskUpdatedEvent{
		Task: &task,
		Doer: doer,
	})
}

// Create adds a new assignee to a task
// @Summary Add a new assignee to a task
// @Description Adds a new assignee to a task. The assignee needs to have access to the project, the doer must be able to edit this task.
// @tags assignees
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param assignee body models.TaskAssginee true "The assingee object"
// @Param taskID path int true "Task ID"
// @Success 201 {object} models.TaskAssginee "The created assingee object."
// @Failure 400 {object} web.HTTPError "Invalid assignee object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/assignees [put]
func (la *TaskAssginee) Create(s *xorm.Session, a web.Auth) (err error) {

	// Get the project to perform later checks
	project, err := GetProjectSimpleByTaskID(s, la.TaskID)
	if err != nil {
		return
	}

	task := &Task{ID: la.TaskID}
	return task.addNewAssigneeByID(s, la.UserID, project, a)
}

func (t *Task) addNewAssigneeByID(s *xorm.Session, newAssigneeID int64, project *Project, auth web.Auth) (err error) {
	// Check if the user exists and has access to the project
	newAssignee, err := user.GetUserByID(s, newAssigneeID)
	if err != nil {
		return err
	}
	canRead, _, err := project.CanRead(s, newAssignee)
	if err != nil {
		return err
	}
	if !canRead {
		return ErrUserDoesNotHaveAccessToProject{project.ID, newAssigneeID}
	}

	exist, err := s.
		Where("task_id = ? AND user_id = ?", t.ID, newAssigneeID).
		Exist(&TaskAssginee{})
	if err != nil {
		return err
	}
	if exist {
		return &ErrUserAlreadyAssigned{
			UserID: newAssigneeID,
			TaskID: t.ID,
		}
	}

	_, err = s.Insert(&TaskAssginee{
		TaskID: t.ID,
		UserID: newAssigneeID,
	})
	if err != nil {
		return err
	}

	sub := &Subscription{
		UserID:     newAssigneeID,
		EntityType: SubscriptionEntityTask,
		EntityID:   t.ID,
	}

	err = sub.Create(s, newAssignee)
	if err != nil && !IsErrSubscriptionAlreadyExists(err) {
		return err
	}

	doer, _ := user.GetFromAuth(auth)
	task, err := GetTaskSimple(s, &Task{ID: t.ID})
	if err != nil {
		return err
	}
	err = events.Dispatch(&TaskAssigneeCreatedEvent{
		Task:     &task,
		Assignee: newAssignee,
		Doer:     doer,
	})
	if err != nil {
		return err
	}
	err = events.Dispatch(&TaskUpdatedEvent{
		Task: &task,
		Doer: doer,
	})
	if err != nil {
		return err
	}

	err = updateProjectLastUpdated(s, &Project{ID: t.ProjectID})
	return
}

// ReadAll gets all assignees for a task
// @Summary Get all assignees for a task
// @Description Returns an array with all assignees for this task.
// @tags assignees
// @Accept json
// @Produce json
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search assignees by their username."
// @Param taskID path int true "Task ID"
// @Security JWTKeyAuth
// @Success 200 {array} user.User "The assignees"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/assignees [get]
func (la *TaskAssginee) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	task, err := GetProjectSimpleByTaskID(s, la.TaskID)
	if err != nil {
		return nil, 0, 0, err
	}

	can, _, err := task.CanRead(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !can {
		return nil, 0, 0, ErrGenericForbidden{}
	}
	limit, start := getLimitFromPageIndex(page, perPage)
	var taskAssignees []*user.User
	query := s.Table("task_assignees").
		Select("users.*").
		Join("INNER", "users", "task_assignees.user_id = users.id").
		Where(builder.And(
			builder.Eq{"task_id": la.TaskID},
			db.ILIKE("users.username", search),
		))
	if limit > 0 {
		query = query.Limit(limit, start)
	}
	err = query.Find(&taskAssignees)
	if err != nil {
		return nil, 0, 0, err
	}

	numberOfTotalItems, err = s.Table("task_assignees").
		Select("users.*").
		Join("INNER", "users", "task_assignees.user_id = users.id").
		Where("task_id = ? AND users.username LIKE ?", la.TaskID, "%"+search+"%").
		Count(&user.User{})
	return taskAssignees, len(taskAssignees), numberOfTotalItems, err
}

// BulkAssignees is a helper struct used to update multiple assignees at once.
type BulkAssignees struct {
	// A project with all assignees
	Assignees []*user.User `json:"assignees"`
	TaskID    int64        `json:"-" param:"projecttask"`

	web.CRUDable    `json:"-"`
	web.Permissions `json:"-"`
}

// Create adds new assignees to a task
// @Summary Add multiple new assignees to a task
// @Description Adds multiple new assignees to a task. The assignee needs to have access to the project, the doer must be able to edit this task. Every user not in the project will be unassigned from the task, pass an empty array to unassign everyone.
// @tags assignees
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param assignee body models.BulkAssignees true "The array of assignees"
// @Param taskID path int true "Task ID"
// @Success 201 {object} models.TaskAssginee "The created assingees object."
// @Failure 400 {object} web.HTTPError "Invalid assignee object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/assignees/bulk [post]
func (ba *BulkAssignees) Create(s *xorm.Session, a web.Auth) (err error) {
	task, err := GetTaskByIDSimple(s, ba.TaskID)
	if err != nil {
		return
	}
	assignees, err := getRawTaskAssigneesForTasks(s, []int64{task.ID})
	if err != nil {
		return err
	}
	for i := range assignees {
		task.Assignees = append(task.Assignees, &assignees[i].User)
	}

	err = task.updateTaskAssignees(s, ba.Assignees, a)
	return
}
