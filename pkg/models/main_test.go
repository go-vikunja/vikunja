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

// mockProjectService provides a test implementation that uses the original model logic
// This prevents import cycles while allowing model tests to continue working
type mockProjectService struct{}

func (m *mockProjectService) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int, isArchived bool, expand ProjectExpandable) (projects []*Project, resultCount int, totalItems int64, err error) {
	// Use the original GetAllRawProjects function
	prs, resultCount, totalItems, err := GetAllRawProjects(s, a, search, page, perPage, isArchived)
	if err != nil {
		return nil, 0, 0, err
	}

	_, is := a.(*LinkSharing)
	if is {
		// If we're dealing with a link share, just return the projects
		return prs, resultCount, totalItems, nil
	}

	// Add project details (favorite state, among other things)
	err = AddProjectDetails(s, prs, a)
	if err != nil {
		return nil, 0, 0, err
	}

	if expand == ProjectExpandableRights {
		var doer *user.User
		doer, err = user.GetFromAuth(a)
		if err != nil {
			return nil, 0, 0, err
		}
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
	// Use the original CreateProject function
	err := CreateProject(s, project, u, true, true)
	if err != nil {
		return nil, err
	}

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
