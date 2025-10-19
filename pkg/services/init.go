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

	// Register ProjectService provider to avoid import cycles
	models.RegisterProjectService(func() interface {
		ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int, isArchived bool, expand models.ProjectExpandable) (projects []*models.Project, resultCount int, totalItems int64, err error)
		Create(s *xorm.Session, project *models.Project, u *user.User) (*models.Project, error)
		Delete(s *xorm.Session, projectID int64, a web.Auth) error
		DeleteForce(s *xorm.Session, projectID int64, a web.Auth) error
		GetByIDSimple(s *xorm.Session, projectID int64) (*models.Project, error)
		GetByIDs(s *xorm.Session, projectIDs []int64) ([]*models.Project, error)
		GetMapByIDs(s *xorm.Session, projectIDs []int64) (map[int64]*models.Project, error)
	} {
		// Return an adapter that bridges the interface
		return &projectServiceAdapter{service: NewProjectService(nil)}
	})

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

	// Register ProjectUsersService provider to avoid import cycles
	models.RegisterProjectUserService(func() interface {
		Create(s *xorm.Session, projectUser *models.ProjectUser, doer *user.User) error
		Delete(s *xorm.Session, projectUser *models.ProjectUser) error
		GetAll(s *xorm.Session, projectID int64, doer *user.User, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error)
		Update(s *xorm.Session, projectUser *models.ProjectUser) error
	} {
		// Return an adapter that bridges the interface mismatch
		return &projectUserServiceAdapter{service: NewProjectUserService(nil)}
	})

	// Register FavoriteService provider to avoid import cycles
	models.RegisterFavoriteService(func() interface {
		AddToFavorite(s *xorm.Session, entityID int64, a web.Auth, kind models.FavoriteKind) error
		RemoveFromFavorite(s *xorm.Session, entityID int64, a web.Auth, kind models.FavoriteKind) error
		IsFavorite(s *xorm.Session, entityID int64, a web.Auth, kind models.FavoriteKind) (bool, error)
		GetFavoritesMap(s *xorm.Session, entityIDs []int64, a web.Auth, kind models.FavoriteKind) (map[int64]bool, error)
	} {
		// Return an adapter that bridges the interface
		return &favoriteServiceAdapter{service: NewFavoriteService(nil)}
	})

	// Register LabelService provider to avoid import cycles
	models.RegisterLabelService(func() interface {
		Create(s *xorm.Session, label *models.Label, u *user.User) error
		Update(s *xorm.Session, label *models.Label, u *user.User) error
		Delete(s *xorm.Session, label *models.Label, u *user.User) error
		GetAll(s *xorm.Session, u *user.User, search string, page int, perPage int) (interface{}, int, int64, error)
		GetByID(s *xorm.Session, labelID int64) (*models.Label, error)
	} {
		// Return an adapter that bridges the interface
		return &labelServiceAdapter{service: NewLabelService(nil)}
	})

	// Register TeamService provider to avoid import cycles
	models.RegisterTeamService(&teamServiceAdapter{service: NewTeamService(nil)})

	// Register APITokenService provider to avoid import cycles
	models.RegisterAPITokenService(func() interface {
		Create(s *xorm.Session, token *models.APIToken, u *user.User) error
		GetAll(s *xorm.Session, u *user.User, search string, page int, perPage int) ([]*models.APIToken, int, int64, error)
		GetByID(s *xorm.Session, id int64) (*models.APIToken, error)
		Delete(s *xorm.Session, id int64, u *user.User) error
	} {
		// Return an adapter that bridges the interface
		return &apiTokenServiceAdapter{service: NewAPITokenService(nil)}
	})

	// Register ReactionsService provider to avoid import cycles
	models.RegisterReactionsService(&reactionsServiceAdapter{service: NewReactionsService(nil)})

	// Register ProjectViewService provider to avoid import cycles
	models.RegisterProjectViewService(&projectViewServiceAdapter{service: NewProjectViewService(nil)})

	// Register TaskService provider to avoid import cycles
	models.RegisterTaskService(&taskServiceAdapter{service: NewTaskService(nil)})

	// Register LabelTaskService provider to avoid import cycles
	models.RegisterLabelTaskService(&labelTaskServiceAdapter{service: NewLabelService(nil)})

	// Register BulkTaskService provider to avoid import cycles
	models.RegisterBulkTaskService(&bulkTaskServiceAdapter{service: NewBulkTaskService(nil)})

	// Register ProjectDuplicateService provider to avoid import cycles
	models.RegisterProjectDuplicateService(&projectDuplicateServiceAdapter{service: NewProjectDuplicateService(nil)})

	// Register UserExportService function for dependency injection
	models.ExportUserDataFunc = func(s *xorm.Session, u *user.User) error {
		service := NewUserExportService(nil)
		return service.ExportUserData(s, u)
	}

	// Initialize KanbanService to wire up bucket-related model functions
	InitKanbanService()

	// Initialize PermissionService for permission delegation
	// This enables models to delegate permission checks to the service layer
	// NOTE: Permission methods will be migrated incrementally in Phase 4.1 tasks
	InitPermissionService()
}

// projectServiceAdapter adapts ProjectService to the interface expected by models
type projectServiceAdapter struct {
	service *ProjectService
}

func (a *projectServiceAdapter) ReadAll(s *xorm.Session, auth web.Auth, search string, page int, perPage int, isArchived bool, expand models.ProjectExpandable) (projects []*models.Project, resultCount int, totalItems int64, err error) {
	return a.service.ReadAll(s, auth, search, page, perPage, isArchived, expand)
}

func (a *projectServiceAdapter) Create(s *xorm.Session, project *models.Project, u *user.User) (*models.Project, error) {
	return a.service.Create(s, project, u)
}

func (a *projectServiceAdapter) Delete(s *xorm.Session, projectID int64, auth web.Auth) error {
	return a.service.Delete(s, projectID, auth)
}

func (a *projectServiceAdapter) DeleteForce(s *xorm.Session, projectID int64, auth web.Auth) error {
	return a.service.DeleteForce(s, projectID, auth)
}

func (a *projectServiceAdapter) GetByIDSimple(s *xorm.Session, projectID int64) (*models.Project, error) {
	return a.service.GetByIDSimple(s, projectID)
}

func (a *projectServiceAdapter) GetByIDs(s *xorm.Session, projectIDs []int64) ([]*models.Project, error) {
	return a.service.GetByIDs(s, projectIDs)
}

func (a *projectServiceAdapter) GetMapByIDs(s *xorm.Session, projectIDs []int64) (map[int64]*models.Project, error) {
	return a.service.GetMapByIDs(s, projectIDs)
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

// projectUserServiceAdapter adapts ProjectUsersService to the interface expected by models
type projectUserServiceAdapter struct {
	service *ProjectUsersService
}

func (a *projectUserServiceAdapter) Create(s *xorm.Session, projectUser *models.ProjectUser, doer *user.User) error {
	return a.service.Create(s, projectUser, doer)
}

func (a *projectUserServiceAdapter) Delete(s *xorm.Session, projectUser *models.ProjectUser) error {
	return a.service.Delete(s, projectUser)
}

func (a *projectUserServiceAdapter) GetAll(s *xorm.Session, projectID int64, doer *user.User, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	// Call service layer directly - no conversion needed
	users, rc, ti, err := a.service.GetAll(s, projectID, doer, search, page, perPage)
	return users, rc, ti, err
}

func (a *projectUserServiceAdapter) Update(s *xorm.Session, projectUser *models.ProjectUser) error {
	return a.service.Update(s, projectUser)
}

// favoriteServiceAdapter adapts FavoriteService to the interface expected by models
type favoriteServiceAdapter struct {
	service *FavoriteService
}

func (a *favoriteServiceAdapter) AddToFavorite(s *xorm.Session, entityID int64, auth web.Auth, kind models.FavoriteKind) error {
	return a.service.AddToFavorite(s, entityID, auth, kind)
}

func (a *favoriteServiceAdapter) RemoveFromFavorite(s *xorm.Session, entityID int64, auth web.Auth, kind models.FavoriteKind) error {
	return a.service.RemoveFromFavorite(s, entityID, auth, kind)
}

func (a *favoriteServiceAdapter) IsFavorite(s *xorm.Session, entityID int64, auth web.Auth, kind models.FavoriteKind) (bool, error) {
	return a.service.IsFavorite(s, entityID, auth, kind)
}

func (a *favoriteServiceAdapter) GetFavoritesMap(s *xorm.Session, entityIDs []int64, auth web.Auth, kind models.FavoriteKind) (map[int64]bool, error) {
	return a.service.GetFavoritesMap(s, entityIDs, auth, kind)
}

// labelServiceAdapter adapts LabelService to the interface expected by models
type labelServiceAdapter struct {
	service *LabelService
}

func (a *labelServiceAdapter) Create(s *xorm.Session, label *models.Label, u *user.User) error {
	return a.service.Create(s, label, u)
}

func (a *labelServiceAdapter) Update(s *xorm.Session, label *models.Label, u *user.User) error {
	return a.service.Update(s, label, u)
}

func (a *labelServiceAdapter) Delete(s *xorm.Session, label *models.Label, u *user.User) error {
	return a.service.Delete(s, label, u)
}

func (a *labelServiceAdapter) GetAll(s *xorm.Session, u *user.User, search string, page int, perPage int) (interface{}, int, int64, error) {
	return a.service.GetAll(s, u, search, page, perPage)
}

func (a *labelServiceAdapter) GetByID(s *xorm.Session, labelID int64) (*models.Label, error) {
	return a.service.GetByID(s, labelID)
}

// apiTokenServiceAdapter adapts APITokenService to the interface expected by models
type apiTokenServiceAdapter struct {
	service *APITokenService
}

func (a *apiTokenServiceAdapter) Create(s *xorm.Session, token *models.APIToken, u *user.User) error {
	return a.service.Create(s, token, u)
}

func (a *apiTokenServiceAdapter) GetAll(s *xorm.Session, u *user.User, search string, page int, perPage int) ([]*models.APIToken, int, int64, error) {
	return a.service.GetAll(s, u, search, page, perPage)
}

func (a *apiTokenServiceAdapter) GetByID(s *xorm.Session, id int64) (*models.APIToken, error) {
	return a.service.GetByID(s, id)
}

func (a *apiTokenServiceAdapter) Delete(s *xorm.Session, id int64, u *user.User) error {
	return a.service.Delete(s, id, u)
}

// reactionsServiceAdapter adapts ReactionsService to the interface expected by models
type reactionsServiceAdapter struct {
	service *ReactionsService
}

func (a *reactionsServiceAdapter) Create(s *xorm.Session, reaction *models.Reaction, auth web.Auth) error {
	return a.service.Create(s, reaction, auth)
}

func (a *reactionsServiceAdapter) Delete(s *xorm.Session, entityID int64, userID int64, value string, entityKind models.ReactionKind) error {
	return a.service.Delete(s, entityID, userID, value, entityKind)
}

func (a *reactionsServiceAdapter) GetAll(s *xorm.Session, entityID int64, entityKind models.ReactionKind) (models.ReactionMap, error) {
	return a.service.GetAll(s, entityID, entityKind)
}

func (a *reactionsServiceAdapter) CanRead(s *xorm.Session, entityID int64, entityKind models.ReactionKind, auth web.Auth) (bool, int, error) {
	return a.service.CanRead(s, entityID, entityKind, auth)
}

func (a *reactionsServiceAdapter) CanCreate(s *xorm.Session, entityID int64, entityKind models.ReactionKind, auth web.Auth) (bool, error) {
	return a.service.CanCreate(s, entityID, entityKind, auth)
}

func (a *reactionsServiceAdapter) CanDelete(s *xorm.Session, entityID int64, entityKind models.ReactionKind, auth web.Auth) (bool, error) {
	return a.service.CanDelete(s, entityID, entityKind, auth)
}

// projectViewServiceAdapter adapts ProjectViewService to the interface expected by models
type projectViewServiceAdapter struct {
	service *ProjectViewService
}

func (a *projectViewServiceAdapter) Create(s *xorm.Session, pv *models.ProjectView, auth web.Auth, createBacklogBucket bool, addExistingTasksToView bool) error {
	return a.service.Create(s, pv, auth, createBacklogBucket, addExistingTasksToView)
}

func (a *projectViewServiceAdapter) Update(s *xorm.Session, pv *models.ProjectView) error {
	return a.service.Update(s, pv)
}

func (a *projectViewServiceAdapter) Delete(s *xorm.Session, viewID int64, projectID int64) error {
	return a.service.Delete(s, viewID, projectID)
}

func (a *projectViewServiceAdapter) GetAll(s *xorm.Session, projectID int64, auth web.Auth) (views []*models.ProjectView, totalCount int64, err error) {
	return a.service.GetAll(s, projectID, auth)
}

func (a *projectViewServiceAdapter) GetByIDAndProject(s *xorm.Session, viewID, projectID int64) (view *models.ProjectView, err error) {
	return a.service.GetByIDAndProject(s, viewID, projectID)
}

func (a *projectViewServiceAdapter) GetByID(s *xorm.Session, id int64) (view *models.ProjectView, err error) {
	return a.service.GetByID(s, id)
}

func (a *projectViewServiceAdapter) CreateDefaultViewsForProject(s *xorm.Session, project *models.Project, auth web.Auth, createBacklogBucket bool, createDefaultListFilter bool) error {
	return a.service.CreateDefaultViewsForProject(s, project, auth, createBacklogBucket, createDefaultListFilter)
}

// taskServiceAdapter adapts TaskService to the interface expected by models
type taskServiceAdapter struct {
	service *TaskService
}

func (a *taskServiceAdapter) Create(s *xorm.Session, task *models.Task, u *user.User, updateAssignees bool, setBucket bool) error {
	_, err := a.service.CreateWithOptions(s, task, u, updateAssignees, setBucket, false)
	return err
}

func (a *taskServiceAdapter) Update(s *xorm.Session, task *models.Task, u *user.User) (*models.Task, error) {
	return a.service.Update(s, task, u)
}

func (a *taskServiceAdapter) Delete(s *xorm.Session, task *models.Task, auth web.Auth) error {
	return a.service.Delete(s, task, auth)
}

func (a *taskServiceAdapter) GetByID(s *xorm.Session, taskID int64, u *user.User) (*models.Task, error) {
	return a.service.GetByID(s, taskID, u)
}

func (a *taskServiceAdapter) GetByIDSimple(s *xorm.Session, taskID int64) (*models.Task, error) {
	return a.service.GetByIDSimple(s, taskID)
}

func (a *taskServiceAdapter) GetByIDs(s *xorm.Session, ids []int64) ([]*models.Task, error) {
	return a.service.GetByIDs(s, ids)
}

// Permission methods - T-PERM-010
func (a *taskServiceAdapter) CanCreateAssignee(s *xorm.Session, taskID int64, auth web.Auth) (bool, error) {
	return a.service.CanCreateAssignee(s, taskID, auth)
}

func (a *taskServiceAdapter) CanDeleteAssignee(s *xorm.Session, taskID int64, auth web.Auth) (bool, error) {
	return a.service.CanDeleteAssignee(s, taskID, auth)
}

func (a *taskServiceAdapter) CanCreateRelation(s *xorm.Session, taskID int64, otherTaskID int64, relationKind models.RelationKind, auth web.Auth) (bool, error) {
	return a.service.CanCreateRelation(s, taskID, otherTaskID, relationKind, auth)
}

func (a *taskServiceAdapter) CanDeleteRelation(s *xorm.Session, taskID int64, auth web.Auth) (bool, error) {
	return a.service.CanDeleteRelation(s, taskID, auth)
}

func (a *taskServiceAdapter) CanUpdatePosition(s *xorm.Session, taskID int64, auth web.Auth) (bool, error) {
	return a.service.CanUpdatePosition(s, taskID, auth)
}

// labelTaskServiceAdapter adapts LabelService to the interface expected by models for label-task operations
type labelTaskServiceAdapter struct {
	service *LabelService
}

func (a *labelTaskServiceAdapter) AddLabelToTask(s *xorm.Session, labelID, taskID int64, auth web.Auth) error {
	return a.service.AddLabelToTask(s, labelID, taskID, auth)
}

func (a *labelTaskServiceAdapter) RemoveLabelFromTask(s *xorm.Session, labelID, taskID int64, auth web.Auth) error {
	return a.service.RemoveLabelFromTask(s, labelID, taskID, auth)
}

func (a *labelTaskServiceAdapter) UpdateTaskLabels(s *xorm.Session, taskID int64, newLabels []*models.Label, auth web.Auth) error {
	return a.service.UpdateTaskLabels(s, taskID, newLabels, auth)
}

func (a *labelTaskServiceAdapter) GetLabelsByTaskIDs(s *xorm.Session, opts *models.LabelByTaskIDsOptions) ([]*models.LabelWithTaskID, int, int64, error) {
	// Convert from models.LabelByTaskIDsOptions to services.GetLabelsByTaskIDsOptions
	serviceOpts := &GetLabelsByTaskIDsOptions{
		User:                opts.User,
		Search:              opts.Search,
		Page:                opts.Page,
		PerPage:             opts.PerPage,
		TaskIDs:             opts.TaskIDs,
		GetUnusedLabels:     opts.GetUnusedLabels,
		GroupByLabelIDsOnly: opts.GroupByLabelIDsOnly,
		GetForUser:          opts.GetForUser,
	}
	return a.service.GetLabelsByTaskIDs(s, serviceOpts)
}

// bulkTaskServiceAdapter adapts BulkTaskService to the interface expected by models
type bulkTaskServiceAdapter struct {
	service *BulkTaskService
}

func (a *bulkTaskServiceAdapter) GetTasksByIDs(s *xorm.Session, taskIDs []int64) ([]*models.Task, error) {
	return a.service.GetTasksByIDs(s, taskIDs)
}

func (a *bulkTaskServiceAdapter) CanUpdate(s *xorm.Session, taskIDs []int64, auth web.Auth) (bool, error) {
	return a.service.CanUpdate(s, taskIDs, auth)
}

func (a *bulkTaskServiceAdapter) Update(s *xorm.Session, taskIDs []int64, taskUpdate *models.Task, assignees []*user.User, auth web.Auth) error {
	return a.service.Update(s, taskIDs, taskUpdate, assignees, auth)
}

// projectDuplicateServiceAdapter adapts ProjectDuplicateService to the interface expected by models
type projectDuplicateServiceAdapter struct {
	service *ProjectDuplicateService
}

func (a *projectDuplicateServiceAdapter) Duplicate(s *xorm.Session, projectID int64, parentProjectID int64, u *user.User) (*models.Project, error) {
	return a.service.Duplicate(s, projectID, parentProjectID, u)
}

func (a *projectDuplicateServiceAdapter) CanCreate(s *xorm.Session, projectID int64, parentProjectID int64, auth web.Auth) (bool, error) {
	return a.service.CanCreate(s, projectID, parentProjectID, auth)
}

// teamServiceAdapter adapts TeamService to the interface expected by models
type teamServiceAdapter struct {
	service *TeamService
}

func (a *teamServiceAdapter) Create(s *xorm.Session, team *models.Team, doer *user.User, firstUserShouldBeAdmin bool) (*models.Team, error) {
	return a.service.Create(s, team, doer, firstUserShouldBeAdmin)
}

func (a *teamServiceAdapter) GetByID(s *xorm.Session, teamID int64) (*models.Team, error) {
	return a.service.GetByID(s, teamID)
}

func (a *teamServiceAdapter) GetAll(s *xorm.Session, auth web.Auth, search string, page int, perPage int, includePublic bool) (teams []*models.Team, resultCount int, totalItems int64, err error) {
	return a.service.GetAll(s, auth, search, page, perPage, includePublic)
}

func (a *teamServiceAdapter) Update(s *xorm.Session, team *models.Team) (*models.Team, error) {
	return a.service.Update(s, team)
}

func (a *teamServiceAdapter) Delete(s *xorm.Session, teamID int64, doer web.Auth) error {
	return a.service.Delete(s, teamID, doer)
}

func (a *teamServiceAdapter) AddDetailsToTeams(s *xorm.Session, teams []*models.Team) error {
	return a.service.AddDetailsToTeams(s, teams)
}

func (a *teamServiceAdapter) CanRead(s *xorm.Session, teamID int64, auth web.Auth) (bool, int, error) {
	return a.service.CanRead(s, teamID, auth)
}

func (a *teamServiceAdapter) CanCreate(s *xorm.Session, auth web.Auth) (bool, error) {
	return a.service.CanCreate(s, auth)
}

func (a *teamServiceAdapter) CanUpdate(s *xorm.Session, teamID int64, auth web.Auth) (bool, error) {
	return a.service.CanUpdate(s, teamID, auth)
}

func (a *teamServiceAdapter) CanDelete(s *xorm.Session, teamID int64, auth web.Auth) (bool, error) {
	return a.service.CanDelete(s, teamID, auth)
}

func (a *teamServiceAdapter) IsAdmin(s *xorm.Session, teamID int64, auth web.Auth) (bool, error) {
	return a.service.IsAdmin(s, teamID, auth)
}

func (a *teamServiceAdapter) CanCreateTeamMember(s *xorm.Session, teamID int64, auth web.Auth) (bool, error) {
	return a.service.CanCreateTeamMember(s, teamID, auth)
}

func (a *teamServiceAdapter) CanDeleteTeamMember(s *xorm.Session, teamID int64, username string, auth web.Auth) (bool, error) {
	return a.service.CanDeleteTeamMember(s, teamID, username, auth)
}

func (a *teamServiceAdapter) CanUpdateTeamMember(s *xorm.Session, teamID int64, auth web.Auth) (bool, error) {
	return a.service.CanUpdateTeamMember(s, teamID, auth)
}
