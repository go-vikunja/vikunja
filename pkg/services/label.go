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

package services

import (
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/builder"
	"xorm.io/xorm"
)

type LabelService struct {
	DB             *xorm.Engine
	ProjectService *ProjectService
}

func NewLabelService(db *xorm.Engine) *LabelService {
	return &LabelService{
		DB:             db,
		ProjectService: NewProjectService(db),
	}
}

func (ls *LabelService) Create(s *xorm.Session, label *models.Label, u *user.User) error {
	if u == nil {
		return ErrAccessDenied
	}
	label.ID = 0
	label.HexColor = utils.NormalizeHex(label.HexColor)
	label.CreatedByID = u.ID
	label.CreatedBy = u
	_, err := s.Insert(label)
	return err
}

func (ls *LabelService) Get(s *xorm.Session, id int64, u *user.User) (*models.Label, error) {
	label := new(models.Label)
	has, err := s.ID(id).Get(label)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrLabelNotFound
	}

	can, err := ls.Can(s, label, u).Read()
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, ErrAccessDenied
	}

	return label, nil
}

type LabelPermissions struct {
	s     *xorm.Session
	label *models.Label
	user  *user.User
}

func (ls *LabelService) Can(s *xorm.Session, label *models.Label, u *user.User) *LabelPermissions {
	return &LabelPermissions{s: s, label: label, user: u}
}

func (lp *LabelPermissions) Read() (bool, error) {
	if lp.user == nil {
		return false, nil
	}
	return lp.label.CreatedByID == lp.user.ID, nil
}

func (lp *LabelPermissions) Write() (bool, error) {
	if lp.user == nil {
		return false, nil
	}
	return lp.label.CreatedByID == lp.user.ID, nil
}

func (lp *LabelPermissions) ReadAll() (bool, error) {
	if lp.user == nil {
		return false, nil
	}
	return true, nil
}

func (ls *LabelService) GetAll(s *xorm.Session, u *user.User, search string, page int, perPage int) (interface{}, int, int64, error) {
	can, err := ls.Can(s, nil, u).ReadAll()
	if err != nil {
		return nil, 0, 0, err
	}
	if !can {
		return nil, 0, 0, ErrAccessDenied
	}

	// Build the query conditions
	// We group by label ID to avoid duplicate labels (same label on multiple tasks)
	var groupBy = "labels.id"
	var selectStmt = "labels.*"

	// Get all labels associated with tasks the user has access to
	var labels []*models.LabelWithTaskID
	cond := builder.And(builder.NotNull{"label_tasks.label_id"})

	// Get project IDs the user has access to
	projects, err := ls.getProjectIDsForUser(s, u.ID)
	if err != nil {
		return nil, 0, 0, err
	}

	// Filter labels by tasks in accessible projects
	cond = builder.And(builder.In("label_tasks.task_id",
		builder.
			Select("id").
			From("tasks").
			Where(builder.In("project_id", projects)),
	), cond)

	// Include unused labels created by the user
	cond = builder.Or(cond, builder.Eq{"labels.created_by_id": u.ID})

	// Handle search by IDs or text
	ids := []int64{}
	searchTerms := []string{search}

	for _, searchTerm := range searchTerms {
		searchTerm = strings.Trim(searchTerm, " ")
		if searchTerm == "" {
			continue
		}

		vals := strings.Split(searchTerm, ",")
		for _, val := range vals {
			v, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				log.Debugf("Label search string part '%s' is not a number: %s", val, err)
				continue
			}
			ids = append(ids, v)
		}
	}

	if len(ids) > 0 {
		cond = builder.And(cond, builder.In("labels.id", ids))
	} else if search != "" {
		searchTerm := strings.Trim(search, " ")
		if searchTerm != "" {
			cond = builder.And(cond, db.ILIKE("labels.title", searchTerm))
		}
	}

	// Apply pagination
	limit, start := getLimitFromPageIndex(page, perPage)

	query := s.Table("labels").
		Select(selectStmt).
		Join("LEFT", "label_tasks", "label_tasks.label_id = labels.id").
		Where(cond).
		GroupBy(groupBy).
		OrderBy("labels.id ASC")
	if limit > 0 {
		query = query.Limit(limit, start)
	}
	err = query.Find(&labels)
	if err != nil {
		return nil, 0, 0, err
	}

	if len(labels) == 0 {
		return nil, 0, 0, nil
	}

	// Get all created by users
	var userids []int64
	for _, l := range labels {
		userids = append(userids, l.CreatedByID)
	}
	users := make(map[int64]*user.User)
	if len(userids) > 0 {
		err = s.In("id", userids).Find(&users)
		if err != nil {
			return nil, 0, 0, err
		}
	}

	// Obfuscate all user emails
	for _, u := range users {
		u.Email = ""
	}

	// Put it all together
	for in, l := range labels {
		if createdBy, has := users[l.CreatedByID]; has {
			labels[in].CreatedBy = createdBy
		}
	}

	// Get the total number of entries
	totalEntries, err := s.Table("labels").
		Select("count(DISTINCT labels.id)").
		Join("LEFT", "label_tasks", "label_tasks.label_id = labels.id").
		Where(cond).
		Count(&models.Label{})
	if err != nil {
		return nil, 0, 0, err
	}

	return labels, len(labels), totalEntries, err
}

// Helper function to get project IDs for a user
func (ls *LabelService) getProjectIDsForUser(s *xorm.Session, userID int64) ([]int64, error) {
	fullUser, err := user.GetUserByID(s, userID)
	if err != nil {
		return nil, err
	}

	projects, _, err := models.GetAllProjectsForUser(s, fullUser.ID, &models.ProjectOptions{
		User: &user.User{ID: userID},
	})
	if err != nil {
		return nil, err
	}

	projectIDs := make([]int64, 0, len(projects))
	for _, project := range projects {
		projectIDs = append(projectIDs, project.ID)
	}

	return projectIDs, nil
}

func (ls *LabelService) Update(s *xorm.Session, label *models.Label, u *user.User) error {
	// Load the existing label to get the CreatedByID for permission checking
	existingLabel := &models.Label{ID: label.ID}
	exists, err := s.Get(existingLabel)
	if err != nil {
		return err
	}
	if !exists {
		return ErrAccessDenied
	}

	can, err := ls.Can(s, existingLabel, u).Write()
	if err != nil {
		return err
	}
	if !can {
		return ErrAccessDenied
	}

	_, err = s.ID(label.ID).AllCols().Update(label)
	return err
}

func (ls *LabelService) Delete(s *xorm.Session, label *models.Label, u *user.User) error {
	can, err := ls.Can(s, label, u).Write()
	if err != nil {
		return err
	}
	if !can {
		return ErrAccessDenied
	}

	_, err = s.Delete(label)
	return err
}

// GetLabelsByTaskIDsOptions is a struct to not clutter the function with too many optional parameters.
type GetLabelsByTaskIDsOptions struct {
	User                web.Auth
	Search              []string
	Page                int
	PerPage             int
	TaskIDs             []int64
	GetUnusedLabels     bool
	GroupByLabelIDsOnly bool
	GetForUser          bool
}

// GetLabelsByTaskIDs is a helper function to get all labels for a set of tasks
// Used when getting all labels for one task as well when getting all labels
func (ls *LabelService) GetLabelsByTaskIDs(s *xorm.Session, opts *GetLabelsByTaskIDsOptions) (ls2 []*models.LabelWithTaskID, resultCount int, totalEntries int64, err error) {
	linkShare, isLinkShareAuth := opts.User.(*models.LinkSharing)

	// We still need the task ID when we want to get all labels for a task, but because of this, we get the same label
	// multiple times when it is associated to more than one task.
	// Because of this whole thing, we need this extra switch here to only group by Task IDs if needed.
	var groupBy = "labels.id,label_tasks.task_id"
	var selectStmt = "labels.*, label_tasks.task_id"
	if opts.GroupByLabelIDsOnly {
		groupBy = "labels.id"
		selectStmt = "labels.*"
	}

	// Get all labels associated with these tasks
	var labels []*models.LabelWithTaskID
	cond := builder.And(builder.NotNull{"label_tasks.label_id"})
	if len(opts.TaskIDs) > 0 && !opts.GetForUser {
		cond = builder.And(builder.In("label_tasks.task_id", opts.TaskIDs), cond)
	}
	if opts.GetForUser {
		var projectIDs []int64
		if isLinkShareAuth {
			projectIDs = []int64{linkShare.ProjectID}
		} else {
			fullUser, err := user.GetUserByID(s, opts.User.GetID())
			if err != nil {
				return nil, 0, 0, err
			}
			projects, _, err := models.GetAllProjectsForUser(s, fullUser.ID, &models.ProjectOptions{
				User: &user.User{ID: opts.User.GetID()},
			})
			if err != nil {
				return nil, 0, 0, err
			}
			projectIDs = make([]int64, 0, len(projects))
			for _, project := range projects {
				projectIDs = append(projectIDs, project.ID)
			}
		}

		cond = builder.And(builder.In("label_tasks.task_id",
			builder.
				Select("id").
				From("tasks").
				Where(builder.In("project_id", projectIDs)),
		), cond)
	}
	if opts.GetUnusedLabels && !isLinkShareAuth {
		cond = builder.Or(cond, builder.Eq{"labels.created_by_id": opts.User.GetID()})
	}

	ids := []int64{}

	for _, search := range opts.Search {
		search = strings.Trim(search, " ")
		if search == "" {
			continue
		}

		vals := strings.Split(search, ",")
		for _, val := range vals {
			v, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				log.Debugf("Label search string part '%s' is not a number: %s", val, err)
				continue
			}
			ids = append(ids, v)
		}
	}

	if len(ids) > 0 {
		cond = builder.And(cond, builder.In("labels.id", ids))
	} else if len(opts.Search) > 0 {
		var searchcond builder.Cond
		for _, search := range opts.Search {
			search = strings.Trim(search, " ")
			if search == "" {
				continue
			}

			searchcond = builder.Or(searchcond, db.ILIKE("labels.title", search))
		}

		cond = builder.And(cond, searchcond)
	}

	limit, start := getLimitFromPageIndex(opts.Page, opts.PerPage)

	query := s.Table("labels").
		Select(selectStmt).
		Join("LEFT", "label_tasks", "label_tasks.label_id = labels.id").
		Where(cond).
		GroupBy(groupBy).
		OrderBy("labels.id ASC")
	if limit > 0 {
		query = query.Limit(limit, start)
	}
	err = query.Find(&labels)
	if err != nil {
		return nil, 0, 0, err
	}

	if len(labels) == 0 {
		return nil, 0, 0, nil
	}

	// Get all created by users
	var userids []int64
	for _, l := range labels {
		userids = append(userids, l.CreatedByID)
	}
	users := make(map[int64]*user.User)
	if len(userids) > 0 {
		err = s.In("id", userids).Find(&users)
		if err != nil {
			return nil, 0, 0, err
		}
	}

	// Obfuscate all user emails
	for _, u := range users {
		u.Email = ""
	}

	// Put it all together
	for in, l := range labels {
		if createdBy, has := users[l.CreatedByID]; has {
			labels[in].CreatedBy = createdBy
		}
	}

	// Get the total number of entries
	totalEntries, err = s.Table("labels").
		Select("count(DISTINCT labels.id)").
		Join("LEFT", "label_tasks", "label_tasks.label_id = labels.id").
		Where(cond).
		Count(&models.Label{})
	if err != nil {
		return nil, 0, 0, err
	}

	return labels, len(labels), totalEntries, err
}

// HasAccessToLabel checks if a user has access to a specific label
func (ls *LabelService) HasAccessToLabel(s *xorm.Session, labelID int64, a web.Auth) (bool, error) {
	if a == nil {
		return false, nil
	}

	label := &models.Label{ID: labelID}
	exists, err := s.Get(label)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, models.ErrLabelDoesNotExist{LabelID: labelID}
	}

	// Check if user created the label
	if label.CreatedByID == a.GetID() {
		return true, nil
	}

	// Check if label is associated with tasks in user's accessible projects
	linkShare, isLinkShare := a.(*models.LinkSharing)

	var projectIDs []int64
	if isLinkShare {
		projectIDs = []int64{linkShare.ProjectID}
	} else {
		fullUser, err := user.GetUserByID(s, a.GetID())
		if err != nil {
			return false, err
		}
		projects, _, err := models.GetAllProjectsForUser(s, fullUser.ID, &models.ProjectOptions{
			User: &user.User{ID: a.GetID()},
		})
		if err != nil {
			return false, err
		}
		projectIDs = make([]int64, 0, len(projects))
		for _, project := range projects {
			projectIDs = append(projectIDs, project.ID)
		}
	}

	cond := builder.In("label_tasks.task_id",
		builder.
			Select("id").
			From("tasks").
			Where(builder.In("project_id", projectIDs)),
	)

	exists, err = s.Table("labels").
		Select("label_tasks.*").
		Join("LEFT", "label_tasks", "label_tasks.label_id = labels.id").
		Where(cond).
		And("labels.id = ?", labelID).
		Exist(&models.LabelTask{})

	return exists, err
}

// IsLabelOwner checks if the user is the owner of the label
func (ls *LabelService) IsLabelOwner(s *xorm.Session, labelID int64, a web.Auth) (bool, error) {
	if a == nil {
		return false, nil
	}

	if _, is := a.(*models.LinkSharing); is {
		return false, nil
	}

	label := &models.Label{ID: labelID}
	exists, err := s.Get(label)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, models.ErrLabelDoesNotExist{LabelID: labelID}
	}

	return label.CreatedByID == a.GetID(), nil
}

// AddLabelToTask adds a label to a task
func (ls *LabelService) AddLabelToTask(s *xorm.Session, labelID, taskID int64, a web.Auth) error {
	// Check if the label exists and user has access
	hasAccess, err := ls.HasAccessToLabel(s, labelID, a)
	if err != nil {
		return err
	}
	if !hasAccess {
		u, _ := a.(*user.User)
		if u != nil {
			return models.ErrUserHasNoAccessToLabel{LabelID: labelID, UserID: u.ID}
		}
		return ErrAccessDenied
	}

	// Check if user can write to the task
	task := &models.Task{ID: taskID}
	exists, err := s.Get(task)
	if err != nil {
		return err
	}
	if !exists {
		return models.ErrTaskDoesNotExist{ID: taskID}
	}

	canUpdate, err := task.CanUpdate(s, a)
	if err != nil {
		return err
	}
	if !canUpdate {
		return ErrAccessDenied
	}

	// Check if the label is already added
	exists, err = s.Exist(&models.LabelTask{LabelID: labelID, TaskID: taskID})
	if err != nil {
		return err
	}
	if exists {
		return models.ErrLabelIsAlreadyOnTask{LabelID: labelID, TaskID: taskID}
	}

	// Add the label to the task
	_, err = s.Insert(&models.LabelTask{LabelID: labelID, TaskID: taskID})
	return err
}

// RemoveLabelFromTask removes a label from a task
func (ls *LabelService) RemoveLabelFromTask(s *xorm.Session, labelID, taskID int64, a web.Auth) error {
	// Check if user can write to the task
	task := &models.Task{ID: taskID}
	exists, err := s.Get(task)
	if err != nil {
		return err
	}
	if !exists {
		return models.ErrTaskDoesNotExist{ID: taskID}
	}

	canUpdate, err := task.CanUpdate(s, a)
	if err != nil {
		return err
	}
	if !canUpdate {
		return ErrAccessDenied
	}

	// Remove the label from the task
	_, err = s.Delete(&models.LabelTask{LabelID: labelID, TaskID: taskID})
	return err
}

// UpdateTaskLabels updates all labels on a task at once
func (ls *LabelService) UpdateTaskLabels(s *xorm.Session, taskID int64, newLabels []*models.Label, a web.Auth) error {
	// Get the task
	task := &models.Task{ID: taskID}
	exists, err := s.Get(task)
	if err != nil {
		return err
	}
	if !exists {
		return models.ErrTaskDoesNotExist{ID: taskID}
	}

	// Check permissions
	canUpdate, err := task.CanUpdate(s, a)
	if err != nil {
		return err
	}
	if !canUpdate {
		return ErrAccessDenied
	}

	// Get current labels
	currentLabels, _, _, err := ls.GetLabelsByTaskIDs(s, &GetLabelsByTaskIDsOptions{
		TaskIDs: []int64{taskID},
	})
	if err != nil {
		return err
	}

	// If we don't have any new labels, delete everything right away
	if len(newLabels) == 0 && len(currentLabels) > 0 {
		_, err = s.Where("task_id = ?", taskID).Delete(&models.LabelTask{})
		return err
	}

	// If we didn't change anything (from 0 to zero) don't do anything
	if len(newLabels) == 0 && len(currentLabels) == 0 {
		return nil
	}

	// Make a hashmap of the new labels for easier comparison
	newLabelsMap := make(map[int64]*models.Label, len(newLabels))
	for _, newLabel := range newLabels {
		newLabelsMap[newLabel.ID] = newLabel
	}

	// Get old labels to delete
	var labelsToDelete []int64
	oldLabelsMap := make(map[int64]*models.Label, len(currentLabels))
	for _, oldLabel := range currentLabels {
		oldLabelsMap[oldLabel.ID] = &oldLabel.Label
		if newLabelsMap[oldLabel.ID] == nil {
			labelsToDelete = append(labelsToDelete, oldLabel.ID)
		}
	}

	// Delete all labels not passed
	if len(labelsToDelete) > 0 {
		_, err = s.In("label_id", labelsToDelete).
			And("task_id = ?", taskID).
			Delete(&models.LabelTask{})
		if err != nil {
			return err
		}
	}

	// Loop through our labels and add new ones
	for _, l := range newLabels {
		// Check if the label is already added on the task
		if oldLabelsMap[l.ID] != nil {
			continue
		}

		// Check if the user has access to the label
		hasAccess, err := ls.HasAccessToLabel(s, l.ID, a)
		if err != nil {
			return err
		}
		if !hasAccess {
			u, _ := a.(*user.User)
			if u != nil {
				return models.ErrUserHasNoAccessToLabel{LabelID: l.ID, UserID: u.ID}
			}
			return ErrAccessDenied
		}

		// Insert it
		_, err = s.Insert(&models.LabelTask{
			LabelID: l.ID,
			TaskID:  taskID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
