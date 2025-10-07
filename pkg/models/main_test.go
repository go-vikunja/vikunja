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
	"fmt"
	"os"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// mockProjectTeamService provides a test implementation that uses the original model logic
// This prevents import cycles while allowing model tests to continue working
type mockProjectTeamService struct{}

func (m *mockProjectTeamService) Create(s *xorm.Session, teamProject *TeamProject, doer web.Auth) error {
	// Call the original model logic directly (before delegation was added)
	// Check if the permissions are valid
	if err := teamProject.Permission.IsValid(); err != nil {
		return err
	}

	// Check if the team exists
	_, err := GetTeamByID(s, teamProject.TeamID)
	if err != nil {
		return err
	}

	// Check if the project exists
	l, err := GetProjectSimpleByID(s, teamProject.ProjectID)
	if err != nil {
		return err
	}

	// Check if the team is already on the project
	exists, err := s.Where("team_id = ?", teamProject.TeamID).
		And("project_id = ?", teamProject.ProjectID).
		Get(&TeamProject{})
	if err != nil {
		return err
	}
	if exists {
		return ErrTeamAlreadyHasAccess{teamProject.TeamID, teamProject.ProjectID}
	}

	// Insert the new team
	teamProject.ID = 0
	_, err = s.Insert(teamProject)
	if err != nil {
		return err
	}

	// Note: Skipping event dispatch and UpdateProjectLastUpdated in test mock
	// to keep it simple and avoid additional dependencies
	return UpdateProjectLastUpdated(s, l)
}

func (m *mockProjectTeamService) Delete(s *xorm.Session, teamProject *TeamProject) error {
	// Check if the team exists
	_, err := GetTeamByID(s, teamProject.TeamID)
	if err != nil {
		return err
	}

	// Check if the team has access to the project
	has, err := s.
		Where("team_id = ? AND project_id = ?", teamProject.TeamID, teamProject.ProjectID).
		Get(&TeamProject{})
	if err != nil {
		return err
	}
	if !has {
		return ErrTeamDoesNotHaveAccessToProject{TeamID: teamProject.TeamID, ProjectID: teamProject.ProjectID}
	}

	// Delete the relation
	_, err = s.Where("team_id = ?", teamProject.TeamID).
		And("project_id = ?", teamProject.ProjectID).
		Delete(&TeamProject{})
	if err != nil {
		return err
	}

	return UpdateProjectLastUpdated(s, &Project{ID: teamProject.ProjectID})
}

func (m *mockProjectTeamService) GetAll(s *xorm.Session, projectID int64, doer web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	// Check if the user can read the project
	l := &Project{ID: projectID}
	canRead, _, err := l.CanRead(s, doer)
	if err != nil {
		return nil, 0, 0, err
	}
	if !canRead {
		return nil, 0, 0, ErrNeedToHaveProjectReadAccess{ProjectID: projectID, UserID: doer.GetID()}
	}

	limit, start := getLimitFromPageIndex(page, perPage)

	// Get the teams
	all := []*TeamWithPermission{}
	query := s.
		Table("teams").
		Join("INNER", "team_projects", "team_id = teams.id").
		Where("team_projects.project_id = ?", projectID)

	if search != "" {
		query = query.Where("teams.name LIKE ?", "%"+search+"%")
	}

	if limit > 0 {
		query = query.Limit(limit, start)
	}
	err = query.Find(&all)
	if err != nil {
		return nil, 0, 0, err
	}

	teams := []*Team{}
	for i := range all {
		teams = append(teams, &all[i].Team)
	}

	err = AddMoreInfoToTeams(s, teams)
	if err != nil {
		return nil, 0, 0, err
	}

	totalItems, err = s.
		Table("teams").
		Join("INNER", "team_projects", "team_id = teams.id").
		Where("team_projects.project_id = ?", projectID).
		Where("teams.name LIKE ?", "%"+search+"%").
		Count(&TeamWithPermission{})
	if err != nil {
		return nil, 0, 0, err
	}

	return all, len(all), totalItems, err
}

func (m *mockProjectTeamService) Update(s *xorm.Session, teamProject *TeamProject) error {
	// Check if the permission is valid
	if err := teamProject.Permission.IsValid(); err != nil {
		return err
	}

	_, err := s.
		Where("project_id = ? AND team_id = ?", teamProject.ProjectID, teamProject.TeamID).
		Cols("permission").
		Update(teamProject)
	if err != nil {
		return err
	}

	return UpdateProjectLastUpdated(s, &Project{ID: teamProject.ProjectID})
}

// mockProjectUserService provides a test implementation that uses the original model logic
// This prevents import cycles while allowing model tests to continue working
type mockProjectUserService struct{}

func (m *mockProjectUserService) Create(s *xorm.Session, projectUser *ProjectUser, doer *user.User) error {
	// Call the original model logic directly (before delegation was added)
	// Check if the permission is valid
	if err := projectUser.Permission.IsValid(); err != nil {
		return err
	}

	// Check if the project exists
	project, err := GetProjectSimpleByID(s, projectUser.ProjectID)
	if err != nil {
		return err
	}

	// Check if the user exists
	targetUser, err := user.GetUserByUsername(s, projectUser.Username)
	if err != nil {
		return err
	}
	projectUser.UserID = targetUser.ID

	// Check if the user already has access or is owner of that project
	// We explicitly DON'T check for teams here
	if project.OwnerID == projectUser.UserID {
		return ErrUserAlreadyHasAccess{UserID: projectUser.UserID, ProjectID: projectUser.ProjectID}
	}

	exist, err := s.Where("project_id = ? AND user_id = ?", projectUser.ProjectID, projectUser.UserID).
		Get(&ProjectUser{})
	if err != nil {
		return err
	}
	if exist {
		return ErrUserAlreadyHasAccess{UserID: projectUser.UserID, ProjectID: projectUser.ProjectID}
	}

	// Insert the new project-user relation
	projectUser.ID = 0
	_, err = s.Insert(projectUser)
	if err != nil {
		return err
	}

	// Update project's last updated timestamp
	return UpdateProjectLastUpdated(s, project)
}

func (m *mockProjectUserService) Delete(s *xorm.Session, projectUser *ProjectUser) error {
	if projectUser.UserID == 0 {
		// Check if the user exists
		targetUser, err := user.GetUserByUsername(s, projectUser.Username)
		if err != nil {
			return err
		}
		projectUser.UserID = targetUser.ID
	}

	// Check if the user has access to the project
	has, err := s.
		Where("user_id = ? AND project_id = ?", projectUser.UserID, projectUser.ProjectID).
		Get(&ProjectUser{})
	if err != nil {
		return err
	}
	if !has {
		return ErrUserDoesNotHaveAccessToProject{ProjectID: projectUser.ProjectID, UserID: projectUser.UserID}
	}

	// Delete the project-user relation
	_, err = s.
		Where("user_id = ? AND project_id = ?", projectUser.UserID, projectUser.ProjectID).
		Delete(&ProjectUser{})
	if err != nil {
		return err
	}

	// Update project's last updated timestamp
	return UpdateProjectLastUpdated(s, &Project{ID: projectUser.ProjectID})
}

func (m *mockProjectUserService) GetAll(s *xorm.Session, projectID int64, doer *user.User, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	// Check if the user has access to the project
	project := &Project{ID: projectID}
	canRead, _, err := project.CanRead(s, doer)
	if err != nil {
		return nil, 0, 0, err
	}
	if !canRead {
		return nil, 0, 0, ErrNeedToHaveProjectReadAccess{UserID: doer.ID, ProjectID: projectID}
	}

	limit, start := getLimitFromPageIndex(page, perPage)

	// Get all users with their permissions
	all := []*UserWithPermission{}
	query := s.
		Select("users.*, users_projects.permission").
		Join("INNER", "users_projects", "users_projects.user_id = users.id").
		Where("users_projects.project_id = ?", projectID)

	if search != "" {
		query = query.Where("users.username LIKE ?", "%"+search+"%")
	}

	if limit > 0 {
		query = query.Limit(limit, start)
	}
	err = query.Find(&all)
	if err != nil {
		return nil, 0, 0, err
	}

	// Obfuscate all user emails for privacy
	for _, u := range all {
		u.Email = ""
	}

	// Get total count
	totalItems, err = s.
		Join("INNER", "users_projects", "user_id = users.id").
		Where("users_projects.project_id = ?", projectID).
		Where("users.username LIKE ?", "%"+search+"%").
		Count(&UserWithPermission{})
	if err != nil {
		return nil, 0, 0, err
	}

	return all, len(all), totalItems, nil
}

func (m *mockProjectUserService) Update(s *xorm.Session, projectUser *ProjectUser) error {
	if projectUser.UserID == 0 {
		// Check if the user exists
		targetUser, err := user.GetUserByUsername(s, projectUser.Username)
		if err != nil {
			return err
		}
		projectUser.UserID = targetUser.ID
	}

	// Check if the permission is valid
	if err := projectUser.Permission.IsValid(); err != nil {
		return err
	}

	// Update the permission
	_, err := s.
		Where("project_id = ? AND user_id = ?", projectUser.ProjectID, projectUser.UserID).
		Cols("permission").
		Update(projectUser)
	if err != nil {
		return err
	}

	// Update project's last updated timestamp
	return UpdateProjectLastUpdated(s, &Project{ID: projectUser.ProjectID})
}

// mockFavoriteService provides a test implementation that uses the original model logic
// This prevents import cycles while allowing model tests to continue working
type mockFavoriteService struct{}

func (m *mockFavoriteService) AddToFavorite(s *xorm.Session, entityID int64, a web.Auth, kind FavoriteKind) error {
	u, err := user.GetFromAuth(a)
	if err != nil {
		// Only error GetFromAuth is if it's a link share and we want to ignore that
		return nil
	}

	fav := &Favorite{
		EntityID: entityID,
		UserID:   u.ID,
		Kind:     kind,
	}

	_, err = s.Insert(fav)
	return err
}

func (m *mockFavoriteService) RemoveFromFavorite(s *xorm.Session, entityID int64, a web.Auth, kind FavoriteKind) error {
	u, err := user.GetFromAuth(a)
	if err != nil {
		// Only error GetFromAuth is if it's a link share and we want to ignore that
		return nil
	}

	_, err = s.
		Where("entity_id = ? AND user_id = ? AND kind = ?", entityID, u.ID, kind).
		Delete(&Favorite{})
	return err
}

func (m *mockFavoriteService) IsFavorite(s *xorm.Session, entityID int64, a web.Auth, kind FavoriteKind) (bool, error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		// Only error GetFromAuth is if it's a link share and we want to ignore that
		return false, nil
	}

	return s.
		Where("entity_id = ? AND user_id = ? AND kind = ?", entityID, u.ID, kind).
		Exist(&Favorite{})
}

func (m *mockFavoriteService) GetFavoritesMap(s *xorm.Session, entityIDs []int64, a web.Auth, kind FavoriteKind) (map[int64]bool, error) {
	favorites := make(map[int64]bool)
	u, err := user.GetFromAuth(a)
	if err != nil {
		// Only error GetFromAuth is if it's a link share and we want to ignore that
		return favorites, nil
	}

	if len(entityIDs) == 0 {
		return favorites, nil
	}

	// Simple implementation: check each ID individually
	for _, id := range entityIDs {
		exists, err := s.
			Where("entity_id = ? AND user_id = ? AND kind = ?", id, u.ID, kind).
			Exist(&Favorite{})
		if err != nil {
			return nil, err
		}
		if exists {
			favorites[id] = true
		}
	}

	return favorites, nil
}

// mockProjectService provides a test implementation that uses direct logic
// This prevents import cycles while allowing model tests to continue working
// Updated to not call model helper functions per T011A-PART2 requirements
type mockProjectService struct{}

func (m *mockProjectService) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int, isArchived bool, expand ProjectExpandable) (projects []*Project, resultCount int, totalItems int64, err error) {
	// Replicate the core logic without calling model helpers

	// Check if we're dealing with a share auth
	shareAuth, is := a.(*LinkSharing)
	if is {
		project, err := GetProjectSimpleByID(s, shareAuth.ProjectID)
		if err != nil {
			return nil, 0, 0, err
		}
		projects := []*Project{project}

		// Add details manually for share auth
		if AddProjectDetailsFunc != nil {
			err = AddProjectDetailsFunc(s, projects, a)
		}
		if err == nil && len(projects) > 0 {
			projects[0].ParentProjectID = 0
		}
		return projects, 0, 0, err
	}

	doer, err := user.GetFromAuth(a)
	if err != nil {
		return nil, 0, 0, err
	}

	// Get raw projects using the low-level function
	prs, resultCount, totalItems, err := getRawProjectsForUser(
		s,
		&ProjectOptions{
			Search:      search,
			User:        doer,
			Page:        page,
			PerPage:     perPage,
			GetArchived: isArchived,
		})
	if err != nil {
		return nil, 0, 0, err
	}

	// Add saved filters
	savedFiltersProject, err := getSavedFilterProjects(s, doer, search)
	if err != nil {
		return nil, 0, 0, err
	}
	if len(savedFiltersProject) > 0 {
		prs = append(prs, savedFiltersProject...)
	}

	// Add project details using the function variable if set
	if AddProjectDetailsFunc != nil {
		err = AddProjectDetailsFunc(s, prs, a)
		if err != nil {
			return nil, 0, 0, err
		}
	}

	// Handle permission expansion
	if expand == ProjectExpandableRights {
		err = AddMaxPermissionToProjects(s, prs, doer)
		if err != nil {
			return nil, 0, 0, err
		}
	} else {
		for _, pr := range prs {
			pr.MaxPermission = PermissionUnknown
		}
	}

	return prs, resultCount, totalItems, err
}

func (m *mockProjectService) Create(s *xorm.Session, project *Project, u *user.User) (*Project, error) {
	// Replicate the core logic without calling model helper CreateProject

	err := project.CheckIsArchived(s)
	if err != nil {
		return nil, err
	}

	project.ID = 0
	project.OwnerID = u.ID
	project.Owner = u

	err = project.validate(s, project)
	if err != nil {
		return nil, err
	}

	project.HexColor = utils.NormalizeHex(project.HexColor)

	_, err = s.Insert(project)
	if err != nil {
		return nil, err
	}

	project.Position = calculateDefaultPosition(project.ID, project.Position)
	_, err = s.Where("id = ?", project.ID).Update(project)
	if err != nil {
		return nil, err
	}

	if project.IsFavorite {
		if err := AddToFavorites(s, project.ID, u, FavoriteKindProject); err != nil {
			return nil, err
		}
	}

	// Create default views for tests
	err = CreateDefaultViewsForProject(s, project, u, true, true)
	if err != nil {
		return nil, err
	}

	err = events.Dispatch(&ProjectCreatedEvent{
		Project: project,
		Doer:    u,
	})
	if err != nil {
		return nil, err
	}

	// Return full project with details
	fullProject, err := GetProjectSimpleByID(s, project.ID)
	if err != nil {
		return nil, err
	}

	err = fullProject.ReadOne(s, u)
	if err != nil {
		return nil, err
	}

	return fullProject, nil
}

func (m *mockProjectService) Delete(s *xorm.Session, projectID int64, a web.Auth) error {
	// Replicate the core delete logic for tests

	// Load the project
	project, err := GetProjectSimpleByID(s, projectID)
	if err != nil {
		return err
	}

	// Check if this is a default project
	isDefaultProject, err := project.IsDefaultProject(s)
	if err != nil {
		return err
	}

	// No one can delete a default project (not even the owner)
	if isDefaultProject {
		return &ErrCannotDeleteDefaultProject{ProjectID: project.ID}
	}

	// Permission check - simplified for tests
	// Check if auth is a link share
	shareAuth, isShare := a.(*LinkSharing)
	if isShare {
		// Link shares can only delete if they have admin permission and it's their project
		if !(project.ID == shareAuth.ProjectID && shareAuth.Permission == PermissionAdmin) {
			return ErrGenericForbidden{}
		}
	} else {
		// Get user for regular auth
		doerUser, err := GetUserOrLinkShareUser(s, a)
		if err != nil {
			return err
		}

		// Owner has full permissions
		if project.OwnerID != doerUser.ID {
			// For non-owners, check if they have admin rights
			can, err := project.CanWrite(s, a)
			if err != nil {
				return err
			}
			if !can {
				return ErrGenericForbidden{}
			}
		}
	}

	// Delete all tasks on that project
	tasks := []*Task{}
	err = s.Where("project_id = ?", project.ID).Find(&tasks)
	if err != nil {
		return err
	}

	// Get user for task deletion
	u, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		err = task.Delete(s, u)
		if err != nil {
			return err
		}
	}

	// Delete background file if exists
	if project.BackgroundFileID != 0 {
		_, err := s.Where("id = ?", project.BackgroundFileID).Delete(&files.File{})
		if err != nil {
			return err
		}
	}

	// Delete related project entities
	views := []*ProjectView{}
	err = s.Where("project_id = ?", project.ID).Find(&views)
	if err != nil {
		return err
	}

	viewIDs := []int64{}
	for _, v := range views {
		viewIDs = append(viewIDs, v.ID)
	}

	if len(viewIDs) > 0 {
		// Delete buckets associated with these views
		_, err = s.In("project_view_id", viewIDs).Delete(&Bucket{})
		if err != nil {
			return err
		}

		// Delete the views themselves
		_, err = s.In("id", viewIDs).Delete(&ProjectView{})
		if err != nil {
			return err
		}
	}

	// Remove from favorites
	err = RemoveFromFavorite(s, project.ID, u, FavoriteKindProject)
	if err != nil {
		return err
	}

	// Delete link sharing
	_, err = s.Where("project_id = ?", project.ID).Delete(&LinkSharing{})
	if err != nil {
		return err
	}

	// Delete project users
	_, err = s.Where("project_id = ?", project.ID).Delete(&ProjectUser{})
	if err != nil {
		return err
	}

	// Delete team projects
	_, err = s.Where("project_id = ?", project.ID).Delete(&TeamProject{})
	if err != nil {
		return err
	}

	// Delete the project itself
	_, err = s.ID(project.ID).Delete(&Project{})
	if err != nil {
		return err
	}

	// Dispatch project deleted event
	err = events.Dispatch(&ProjectDeletedEvent{
		Project: project,
		Doer:    u,
	})
	if err != nil {
		return err
	}

	// Delete child projects recursively
	childProjects := []*Project{}
	err = s.Where("parent_project_id = ?", project.ID).Find(&childProjects)
	if err != nil {
		return err
	}

	for _, child := range childProjects {
		err = m.Delete(s, child.ID, u)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *mockProjectService) DeleteForce(s *xorm.Session, projectID int64, a web.Auth) error {
	// DeleteForce is the same as Delete but allows deleting default projects
	// Load the project
	project, err := GetProjectSimpleByID(s, projectID)
	if err != nil {
		return err
	}

	// Check if this is a default project
	isDefaultProject, err := project.IsDefaultProject(s)
	if err != nil {
		return err
	}

	// If we're deleting a default project, remove it as default first
	if isDefaultProject {
		_, err = s.Where("default_project_id = ?", project.ID).
			Cols("default_project_id").
			Update(&user.User{DefaultProjectID: 0})
		if err != nil {
			return err
		}
	}

	// Permission check - simplified for tests
	// Check if auth is a link share
	shareAuth, isShare := a.(*LinkSharing)
	if isShare {
		// Link shares can only delete if they have admin permission and it's their project
		if !(project.ID == shareAuth.ProjectID && shareAuth.Permission == PermissionAdmin) {
			return ErrGenericForbidden{}
		}
	} else {
		// Get user for regular auth
		doerUser, err := GetUserOrLinkShareUser(s, a)
		if err != nil {
			return err
		}

		// Owner has full permissions
		if project.OwnerID != doerUser.ID {
			// For non-owners, check if they have admin rights
			can, err := project.CanWrite(s, a)
			if err != nil {
				return err
			}
			if !can {
				return ErrGenericForbidden{}
			}
		}
	}

	// Delete all tasks on that project
	tasks := []*Task{}
	err = s.Where("project_id = ?", project.ID).Find(&tasks)
	if err != nil {
		return err
	}

	// Get user for task deletion
	u, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		err = task.Delete(s, u)
		if err != nil {
			return err
		}
	}

	// Delete background file if exists
	if project.BackgroundFileID != 0 {
		_, err := s.Where("id = ?", project.BackgroundFileID).Delete(&files.File{})
		if err != nil {
			return err
		}
	}

	// Delete related project entities
	views := []*ProjectView{}
	err = s.Where("project_id = ?", project.ID).Find(&views)
	if err != nil {
		return err
	}

	viewIDs := []int64{}
	for _, v := range views {
		viewIDs = append(viewIDs, v.ID)
	}

	if len(viewIDs) > 0 {
		// Delete buckets associated with these views
		_, err = s.In("project_view_id", viewIDs).Delete(&Bucket{})
		if err != nil {
			return err
		}

		// Delete the views themselves
		_, err = s.In("id", viewIDs).Delete(&ProjectView{})
		if err != nil {
			return err
		}
	}

	// Remove from favorites
	err = RemoveFromFavorite(s, project.ID, u, FavoriteKindProject)
	if err != nil {
		return err
	}

	// Delete link sharing
	_, err = s.Where("project_id = ?", project.ID).Delete(&LinkSharing{})
	if err != nil {
		return err
	}

	// Delete project users
	_, err = s.Where("project_id = ?", project.ID).Delete(&ProjectUser{})
	if err != nil {
		return err
	}

	// Delete team projects
	_, err = s.Where("project_id = ?", project.ID).Delete(&TeamProject{})
	if err != nil {
		return err
	}

	// Delete the project itself
	_, err = s.ID(project.ID).Delete(&Project{})
	if err != nil {
		return err
	}

	// Dispatch project deleted event
	err = events.Dispatch(&ProjectDeletedEvent{
		Project: project,
		Doer:    u,
	})
	if err != nil {
		return err
	}

	// Delete child projects recursively
	childProjects := []*Project{}
	err = s.Where("parent_project_id = ?", project.ID).Find(&childProjects)
	if err != nil {
		return err
	}

	for _, child := range childProjects {
		err = m.DeleteForce(s, child.ID, a)
		if err != nil {
			return err
		}
	}

	return nil
}

func setupTime() {
	var err error
	loc, err := time.LoadLocation("GMT")
	if err != nil {
		fmt.Printf("Error setting up time: %s", err)
		os.Exit(1)
	}
	testCreatedTime, err = time.ParseInLocation(time.RFC3339Nano, "2018-12-01T15:13:12.0+00:00", loc)
	if err != nil {
		fmt.Printf("Error setting up time: %s", err)
		os.Exit(1)
	}
	testCreatedTime = testCreatedTime.In(loc)
	testUpdatedTime, err = time.ParseInLocation(time.RFC3339Nano, "2018-12-02T15:13:12.0+00:00", loc)
	if err != nil {
		fmt.Printf("Error setting up time: %s", err)
		os.Exit(1)
	}
	testUpdatedTime = testUpdatedTime.In(loc)
}

func TestMain(m *testing.M) {

	setupTime()

	// Initialize logger for tests
	log.InitLogger()

	// Set default config
	config.InitDefaultConfig()
	// We need to set the root path even if we're not using the config, otherwise fixtures are not loaded correctly
	config.ServiceRootpath.Set(os.Getenv("VIKUNJA_SERVICE_ROOTPATH"))

	i18n.Init()

	// Some tests use the file engine, so we'll need to initialize that
	files.InitTests()

	user.InitTests()

	SetupTests()

	// Register a mock ProjectService provider for model tests
	// This avoids import cycle with services package
	RegisterProjectService(func() interface {
		ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int, isArchived bool, expand ProjectExpandable) (projects []*Project, resultCount int, totalItems int64, err error)
		Create(s *xorm.Session, project *Project, u *user.User) (*Project, error)
		Delete(s *xorm.Session, projectID int64, a web.Auth) error
		DeleteForce(s *xorm.Session, projectID int64, a web.Auth) error
	} {
		// Return a mock that delegates to the original model methods
		// This preserves backward compatibility for model tests
		return &mockProjectService{}
	})

	// Register a mock ProjectTeamService provider for model tests
	// This avoids import cycle with services package
	RegisterProjectTeamService(func() interface {
		Create(s *xorm.Session, teamProject *TeamProject, doer web.Auth) error
		Delete(s *xorm.Session, teamProject *TeamProject) error
		GetAll(s *xorm.Session, projectID int64, doer web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error)
		Update(s *xorm.Session, teamProject *TeamProject) error
	} {
		// Return a mock that delegates to the original model methods
		// This preserves backward compatibility for model tests
		return &mockProjectTeamService{}
	})

	// Register a mock ProjectUserService provider for model tests
	// This avoids import cycle with services package
	RegisterProjectUserService(func() interface {
		Create(s *xorm.Session, projectUser *ProjectUser, doer *user.User) error
		Delete(s *xorm.Session, projectUser *ProjectUser) error
		GetAll(s *xorm.Session, projectID int64, doer *user.User, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error)
		Update(s *xorm.Session, projectUser *ProjectUser) error
	} {
		// Return a mock that delegates to the original model methods
		// This preserves backward compatibility for model tests
		return &mockProjectUserService{}
	})

	// Register a mock FavoriteService provider for model tests
	// This avoids import cycle with services package
	RegisterFavoriteService(func() interface {
		AddToFavorite(s *xorm.Session, entityID int64, a web.Auth, kind FavoriteKind) error
		RemoveFromFavorite(s *xorm.Session, entityID int64, a web.Auth, kind FavoriteKind) error
		IsFavorite(s *xorm.Session, entityID int64, a web.Auth, kind FavoriteKind) (bool, error)
		GetFavoritesMap(s *xorm.Session, entityIDs []int64, a web.Auth, kind FavoriteKind) (map[int64]bool, error)
	} {
		// Return a mock that delegates to the original model methods
		// This preserves backward compatibility for model tests
		return &mockFavoriteService{}
	})

	// Set up a mock for the GetUsersOrLinkSharesFromIDsFunc for model tests,
	// as they should not depend on the services package.
	GetUsersOrLinkSharesFromIDsFunc = func(s *xorm.Session, ids []int64) (map[int64]*user.User, error) {
		usersMap := make(map[int64]*user.User)
		var userIDs []int64
		var linkShareIDs []int64
		for _, id := range ids {
			if id < 0 {
				linkShareIDs = append(linkShareIDs, id*-1)
				continue
			}
			userIDs = append(userIDs, id)
		}

		if len(userIDs) > 0 {
			var err error
			usersMap, err = user.GetUsersByIDs(s, userIDs)
			if err != nil {
				return nil, err
			}
		}

		if len(linkShareIDs) == 0 {
			return usersMap, nil
		}

		shares, err := GetLinkSharesByIDs(s, linkShareIDs)
		if err != nil {
			return nil, err
		}

		for _, share := range shares {
			usersMap[share.ID*-1] = share.ToUser()
		}

		return usersMap, nil
	}

	// Set up a mock for AddMoreInfoToTasksFunc for model tests,
	// as they should not depend on the services package.
	AddMoreInfoToTasksFunc = func(s *xorm.Session, taskMap map[int64]*Task, a web.Auth, view *ProjectView, expand []TaskCollectionExpandable) error {
		// This is a minimal mock that just returns nil - no additional task details are added in tests
		// Individual tests can override this if they need specific behavior
		return nil
	}

	events.Fake()

	os.Exit(m.Run())
}
