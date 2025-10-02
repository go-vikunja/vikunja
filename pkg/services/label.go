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
	"xorm.io/builder"
	"xorm.io/xorm"
)

type LabelService struct {
	DB *xorm.Engine
}

func NewLabelService(db *xorm.Engine) *LabelService {
	return &LabelService{DB: db}
}

func (ls *LabelService) Create(s *xorm.Session, label *models.Label, u *user.User) error {
	if u == nil {
		return ErrAccessDenied
	}
	label.CreatedByID = u.ID
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
