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
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// InitializeDependencies wires up service layer dependencies with the models layer
// This must be called during application initialization to enable service layer functionality
func InitializeDependencies() {
	// Initialize user mentions service
	mentionsService := NewUserMentionsService()

	// Inject the service function into models to avoid import cycles
	models.NotifyMentionedUsersFunc = func(
		sess *xorm.Session,
		subject interface {
			CanRead(s *xorm.Session, a web.Auth) (bool, int, error)
		},
		text string,
		notification notifications.NotificationWithSubject,
	) (users map[int64]*user.User, err error) {
		return mentionsService.NotifyMentionedUsers(sess, subject, text, notification)
	}

	// Register ProjectTeamService provider to avoid import cycles
	models.RegisterProjectTeamService(func() interface {
		Create(s *xorm.Session, teamProject *models.TeamProject, doer web.Auth) error
		Delete(s *xorm.Session, teamProject *models.TeamProject) error
		GetAll(s *xorm.Session, projectID int64, doer web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error)
		Update(s *xorm.Session, teamProject *models.TeamProject) error
	} {
		// Return an adapter that bridges the interface mismatch
		return &projectTeamServiceAdapter{service: NewProjectTeamService(nil)}
	})
}

// projectTeamServiceAdapter adapts ProjectTeamService to the interface expected by models
type projectTeamServiceAdapter struct {
	service *ProjectTeamService
}

func (a *projectTeamServiceAdapter) Create(s *xorm.Session, teamProject *models.TeamProject, doer web.Auth) error {
	return a.service.Create(s, teamProject, doer)
}

func (a *projectTeamServiceAdapter) Delete(s *xorm.Session, teamProject *models.TeamProject) error {
	return a.service.Delete(s, teamProject)
}

func (a *projectTeamServiceAdapter) GetAll(s *xorm.Session, projectID int64, doer web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	// Call service layer directly - no conversion needed
	teams, rc, ti, err := a.service.GetAll(s, projectID, doer, search, page, perPage)
	return teams, rc, ti, err
}

func (a *projectTeamServiceAdapter) Update(s *xorm.Session, teamProject *models.TeamProject) error {
	return a.service.Update(s, teamProject)
}
