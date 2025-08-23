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
	"fmt"
	"math"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// Project is a service for projects.
type Project struct {
	DB *xorm.Engine
}

// Get gets a project by its ID.
func (p *Project) Get(s *xorm.Session, projectID int64, u *user.User) (*models.Project, error) {
	return nil, nil
}

// Update updates a project.
func (p *Project) Update(s *xorm.Session, project *models.Project, u *user.User) (*models.Project, error) {
	// Permission check
	can, err := project.CanUpdate(s, u)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, &models.ErrGenericForbidden{}
	}

	err = p.validate(s, project)
	if err != nil {
		return nil, err
	}

	if project.IsArchived {
		isDefaultProject, err := project.IsDefaultProject(s)
		if err != nil {
			return nil, err
		}

		if isDefaultProject {
			return nil, &models.ErrCannotArchiveDefaultProject{ProjectID: project.ID}
		}
	}

	err = SetArchiveStateForProjectDescendants(s, project.ID, project.IsArchived)
	if err != nil {
		return nil, err
	}

	// We need to specify the cols we want to update here to be able to un-archive projects
	colsToUpdate := []string{
		"title",
		"is_archived",
		"identifier",
		"hex_color",
		"parent_project_id",
		"position",
	}
	if project.Description != "" {
		colsToUpdate = append(colsToUpdate, "description")
	}

	if project.Position < 0.1 {
		err = recalculateProjectPositions(s, project.ParentProjectID)
		if err != nil {
			return nil, err
		}
	}

	wasFavorite, err := models.IsFavorite(s, project.ID, u, models.FavoriteKindProject)
	if err != nil {
		return nil, err
	}
	if project.IsFavorite && !wasFavorite {
		if err := models.AddToFavorites(s, project.ID, u, models.FavoriteKindProject); err != nil {
			return nil, err
		}
	}

	if !project.IsFavorite && wasFavorite {
		if err := models.RemoveFromFavorite(s, project.ID, u, models.FavoriteKindProject); err != nil {
			return nil, err
		}
	}

	project.HexColor = utils.NormalizeHex(project.HexColor)

	_, err = s.
		ID(project.ID).
		Cols(colsToUpdate...).
		Update(project)
	if err != nil {
		return nil, err
	}

	err = events.Dispatch(&models.ProjectUpdatedEvent{
		Project: project,
		Doer:    u,
	})
	if err != nil {
		return nil, err
	}

	l, err := models.GetProjectSimpleByID(s, project.ID)
	if err != nil {
		return nil, err
	}

	*project = *l
	err = project.ReadOne(s, u)
	return project, err
}

func recalculateProjectPositions(s *xorm.Session, parentProjectID int64) (err error) {

	allProjects := []*models.Project{}
	err = s.
		Where("parent_project_id = ?", parentProjectID).
		OrderBy("position asc").
		Find(&allProjects)
	if err != nil {
		return
	}

	maxPosition := math.Pow(2, 32)

	for i, project := range allProjects {

		currentPosition := maxPosition / float64(len(allProjects)) * (float64(i+1))

		_, err = s.Cols("position").
			Where("id = ?", project.ID).
			Update(&models.Project{Position: currentPosition})
		if err != nil {
			return
		}
	}

	return
}

// SetArchiveStateForProjectDescendants uses a recursive CTE to find and set the archived status of all descendant projects.
func SetArchiveStateForProjectDescendants(s *xorm.Session, parentProjectID int64, shouldBeArchived bool) error {
	var descendantIDs []int64
	err := s.SQL(
		`
WITH RECURSIVE descendant_ids (id) AS (
    SELECT id
    FROM projects
    WHERE parent_project_id = ?
    UNION ALL
    SELECT p.id
    FROM projects p
    INNER JOIN descendant_ids di ON p.parent_project_id = di.id
)
SELECT id FROM descendant_ids`,
		parentProjectID,
	).Find(&descendantIDs)
	if err != nil {
		log.Errorf("Error finding descendant projects for parent ID %d: %v", parentProjectID, err)
		return fmt.Errorf("failed to find descendant projects for parent ID %d: %w", parentProjectID, err)
	}

	if len(descendantIDs) == 0 {
		return nil
	}

	_, err = s.In("id", descendantIDs).
		And("is_archived != ?", shouldBeArchived).
		Cols("is_archived").
		Update(&models.Project{IsArchived: shouldBeArchived})
	if err != nil {
		log.Errorf("Error updating is_archived for descendant projects for parent ID %d to %t: %v", parentProjectID, shouldBeArchived, err)
		return fmt.Errorf("failed to update is_archived for descendant projects for parent ID %d to %t: %w", parentProjectID, shouldBeArchived, err)
	}
	return nil
}

// GetByID gets a project by its ID.
func (p *Project) GetByID(s *xorm.Session, projectID int64, u *user.User) (*models.Project, error) {
	project, err := models.GetProjectSimpleByID(s, projectID)
	if err != nil {
		return nil, err
	}

	// Permission check
	if project.OwnerID != u.ID {
		has, err := s.Where("project_id = ? AND user_id = ?", projectID, u.ID).Exist(&models.ProjectUser{})
		if err != nil {
			return nil, err
		}
		if !has {
			// Check team permissions
			has, err = s.
				Table("team_members").
				Join("INNER", "team_projects", "team_members.team_id = team_projects.team_id").
				Where("team_members.user_id = ? AND team_projects.project_id = ?", u.ID, projectID).
				Exist()
			if err != nil {
				return nil, err
			}
			if !has {
				return nil, models.ErrProjectDoesNotExist{ID: projectID}
			}
		}
	}

	if err := models.AddProjectDetails(s, []*models.Project{project}, u); err != nil {
		return nil, err
	}
	return project, nil
}

// getLimitFromPageIndex is a helper function to calculate the limit and offset for a given page and items per page
func getLimitFromPageIndex(page int, perPage int) (limit int, start int) {
	if page == 0 {
		page = 1
	}
	if perPage == 0 {
		perPage = 20
	}
	start = (page - 1) * perPage
	return perPage, start
}

func (p *Project) getUserProjectsStatement(userID int64, search string, getArchived bool) *builder.Builder {
	dialect := db.GetDialect()

	conds := []builder.Cond{
		builder.Or(
			builder.Eq{"tm2.user_id": userID},
			builder.Eq{"ul.user_id": userID},
			builder.Eq{"l.owner_id": userID},
		),
	}

	ids := []int64{}
	if search != "" {
		vals := strings.Split(search, ",")
		for _, val := range vals {
			v, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				log.Debugf("Project search string part '%s' is not a number: %s", val, err)
				continue
			}
			ids = append(ids, v)
		}

		var filterCond builder.Cond
		if len(ids) > 0 {
			filterCond = builder.In("l.id", ids)
		} else {
			filterCond = db.MultiFieldSearchWithTableAlias(
				[]string{
					"title",
					"description",
					"identifier",
				},
				search,
				"l",
			)
		}

		parentCondition := builder.Or(
			builder.IsNull{"l.parent_project_id"},
			builder.Eq{"l.parent_project_id": 0},
			// else check for shared sub projects with a parent
			builder.And(
				builder.Or(
					builder.NotNull{"tm2.user_id"},
					builder.NotNull{"ul.user_id"},
				),
				builder.NotNull{"l.parent_project_id"},
			),
		)
		conds = append(conds, filterCond, parentCondition)
	}

	if !getArchived {
		conds = append(conds,
			builder.And(
				builder.Eq{"l.is_archived": false},
			),
		)
	}

	return builder.Dialect(dialect).
		Select("l.*").
		From("projects", "l").
		Join("LEFT", "team_projects tl", "tl.project_id = l.id").
		Join("LEFT", "team_members tm2", "tm2.team_id = tl.team_id").
		Join("LEFT", "users_projects ul", "ul.project_id = l.id").
		Where(builder.And(conds...)).
		GroupBy("l.id")
}

func (p *Project) getAllProjectsForUserInternal(s *xorm.Session, userID int64, search string, page int, perPage int, isArchived bool) (projects []*models.Project, totalCount int64, err error) {
	limit, start := getLimitFromPageIndex(page, perPage)
	query := p.getUserProjectsStatement(userID, search, isArchived)

	querySQLString, args, err := query.ToSQL()
	if err != nil {
		return nil, 0, err
	}

	var limitSQL string
	if limit > 0 {
		limitSQL = fmt.Sprintf("LIMIT %d OFFSET %d", limit, start)
	}

	baseQuery := querySQLString + `
UNION ALL
SELECT p.* FROM projects p
INNER JOIN all_projects ap ON p.parent_project_id = ap.id`

	columnStr := strings.Join([]string{
		"all_projects.id",
		"all_projects.title",
		"all_projects.description",
		"all_projects.identifier",
		"all_projects.hex_color",
		"all_projects.owner_id",
		"CASE WHEN all_projects.parent_project_id IS NULL THEN 0 ELSE all_projects.parent_project_id END AS parent_project_id",
		"all_projects.is_archived",
		"all_projects.background_file_id",
		"all_projects.background_blur_hash",
		"all_projects.position",
		"all_projects.created",
		"all_projects.updated",
	}, ", ")
	currentProjects := []*models.Project{}
	err = s.SQL(`WITH RECURSIVE all_projects as (`+baseQuery+`)
SELECT DISTINCT `+columnStr+` FROM all_projects
ORDER BY all_projects.position `+limitSQL, args...).Find(&currentProjects)
	if err != nil {
		return
	}

	if len(currentProjects) == 0 {
		return nil, 0, err
	}

	totalCount, err = s.
		SQL(`WITH RECURSIVE all_projects as (`+baseQuery+`)
SELECT COUNT(DISTINCT all_projects.id) FROM all_projects`, args...).
		Count(&models.Project{})
	if err != nil {
		return nil, 0, err
	}

	return currentProjects, totalCount, err
}

func (p *Project) getSavedFiltersForUser(s *xorm.Session, u *user.User, search string) (fs []*models.SavedFilter, err error) {
	var cond builder.Cond = builder.Eq{"owner_id": u.ID}
	if search != "" {
		cond = builder.And(cond, db.MultiFieldSearch([]string{"title", "description"}, search))
	}
	err = s.Where(cond).Find(&fs)
	return
}

func (p *Project) getSavedFilterProjects(s *xorm.Session, doer *user.User, search string) (savedFiltersProjects []*models.Project, err error) {
	savedFilters, err := p.getSavedFiltersForUser(s, doer, search)
	if err != nil {
		return
	}

	if len(savedFilters) == 0 {
		return nil, nil
	}

	for _, filter := range savedFilters {
		filterProject := filter.ToProject()
		filterProject.Owner = doer
		savedFiltersProjects = append(savedFiltersProjects, filterProject)
	}

	return
}

// GetAllForUser returns all projects for a user
func (p *Project) GetAllForUser(s *xorm.Session, u *user.User, search string, page int, perPage int, isArchived bool) (projects []*models.Project, resultCount int, totalItems int64, err error) {
	projects, totalItems, err = p.getAllProjectsForUserInternal(s, u.ID, search, page, perPage, isArchived)
	if err != nil {
		return nil, 0, 0, err
	}
	resultCount = len(projects)

	// Saved Filters
	savedFiltersProject, err := p.getSavedFilterProjects(s, u, search)
	if err != nil {
		return nil, 0, 0, err
	}
	totalItems += int64(len(savedFiltersProject))

	// Favorite projects
	favoriteCount, err := s.
		Where(builder.And(
			builder.Eq{"user_id": u.ID},
			builder.Eq{"kind": models.FavoriteKindTask},
		)).
		Count(&models.Favorite{})
	if err != nil {
		return
	}
	if favoriteCount > 0 {
		totalItems++
	}

	if page == 1 || page == 0 { // Only add to the first page
		if len(savedFiltersProject) > 0 {
			projects = append(projects, savedFiltersProject...)
		}

		if favoriteCount > 0 {
			favoritesProject := &models.Project{}
			*favoritesProject = models.FavoritesPseudoProject
			projects = append(projects, favoritesProject)
		}
	}

	// Add project details (favorite state, among other things)
	err = models.AddProjectDetails(s, projects, u)
	if err != nil {
		return
	}

	err = models.AddMaxPermissionToProjects(s, projects, u)
	if err != nil {
		return
	}

	return projects, resultCount, totalItems, err
}

// Create creates a new project.
func (p *Project) Create(s *xorm.Session, project *models.Project, u *user.User) (*models.Project, error) {
	if project.ParentProjectID != 0 {
		// parent := &models.Project{ID: project.ParentProjectID}
		// TODO: Move this to the service
		//can, err := parent.CanWrite(s, u)
		//if err != nil {
		//	return nil, err
		//}
		//if !can {
		//	return nil, errors.New("cannot write to parent project")
		//}
	}

	project.ID = 0
	project.OwnerID = u.ID
	project.Owner = u

	err := p.validate(s, project)
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
		if err := models.AddToFavorites(s, project.ID, u, models.FavoriteKindProject); err != nil {
			return nil, err
		}
	}

	err = CreateDefaultViewsForProject(s, project, u, true, true)
	if err != nil {
		return nil, err
	}

	err = events.Dispatch(&models.ProjectCreatedEvent{
		Project: project,
		Doer:    u,
	})
	if err != nil {
		return nil, err
	}

	fullProject, err := models.GetProjectSimpleByID(s, project.ID)
	if err != nil {
		return nil, err
	}

	return fullProject, err
}

func (p *Project) validate(s *xorm.Session, project *models.Project) (err error) {
	if project.Title == "" {
		return &models.ErrProjectTitleCannotBeEmpty{}
	}

	if project.ParentProjectID < 0 {
		return &ErrProjectCannotBelongToAPseudoParentProject{ProjectID: project.ID, ParentProjectID: project.ParentProjectID}
	}

	// Check if the parent project exists
	if project.ParentProjectID > 0 {
		if project.ParentProjectID == project.ID {
			return &ErrProjectCannotBeChildOfItself{
				ProjectID: project.ID,
			}
		}

		allProjects, err := models.GetAllParentProjects(s, project.ParentProjectID)
		if err != nil {
			return err
		}

		var parent *models.Project
		parent = allProjects[project.ParentProjectID]

		// Check if there's a cycle in the parent relation
		parentsVisited := make(map[int64]bool)
		parentsVisited[project.ID] = true
		for parent.ParentProjectID != 0 {

			parent = allProjects[parent.ParentProjectID]

			if parentsVisited[parent.ID] {
				return &ErrProjectCannotHaveACyclicRelationship{
					ProjectID: project.ID,
				}
			}

			parentsVisited[parent.ID] = true
		}
	}

	// Check if the identifier is unique and not empty
	if project.Identifier != "" {
		exists, err := s.
			Where("identifier = ?", project.Identifier).
			And("id != ?", project.ID).
			Exist(&models.Project{})
		if err != nil {
			return err
		}
		if exists {
			return &models.ErrProjectIdentifierIsNotUnique{Identifier: project.Identifier}
		}
	}

	return nil
}

func calculateDefaultPosition(id int64, position float64) float64 {
	if position < 0.1 {
		return float64(id) * math.Pow(2, 32)
	}

	return position
}

func addToFavorites(s *xorm.Session, entityID int64, u *user.User, kind models.FavoriteKind) (err error) {
	fav := &models.Favorite{
		UserID:   u.ID,
		EntityID: entityID,
		Kind:     kind,
	}
	_, err = s.Insert(fav)
	return
}

func CreateDefaultViewsForProject(s *xorm.Session, project *models.Project, u *user.User, createBacklogBucket bool, createDefaultBuckets bool) (err error) {
	_, err = s.Insert([]*models.ProjectView{
		{
			ProjectID: project.ID,
			Title:     "List",
			ViewKind:  models.ProjectViewKindList,
			Position:  100,
		},
		{
			ProjectID: project.ID,
			Title:     "Gantt",
			ViewKind:  models.ProjectViewKindGantt,
			Position:  200,
		},
		{
			ProjectID: project.ID,
			Title:     "Table",
			ViewKind:  models.ProjectViewKindTable,
			Position:  300,
		},
	})
	if err != nil {
		return
	}

	kanbanView := &models.ProjectView{
		ProjectID:             project.ID,
		Title:                 "Kanban",
		ViewKind:              models.ProjectViewKindKanban,
		Position:              400,
		BucketConfigurationMode: models.BucketConfigurationModeManual,
	}
	_, err = s.Insert(kanbanView)
	if err != nil {
		return
	}

	if !createDefaultBuckets {
		return
	}

	buckets := []*models.Bucket{}
	if createBacklogBucket {
		buckets = append(buckets, &models.Bucket{
			Title:         "Backlog",
			Position:      100,
			ProjectViewID: kanbanView.ID,
		})
	}
	buckets = append(buckets, []*models.Bucket{
		{
			Title:         "To-Do",
			Position:      200,
			ProjectViewID: kanbanView.ID,
		},
		{
			Title:         "Doing",
			Position:      300,
			ProjectViewID: kanbanView.ID,
		},
		{
			Title:         "Done",
			Position:      400,
			ProjectViewID: kanbanView.ID,
		},
	}...)

	_, err = s.Insert(buckets)
	if err != nil {
		return
	}

	kanbanView.DefaultBucketID = buckets[0].ID
	kanbanView.DoneBucketID = buckets[len(buckets)-1].ID
	_, err = s.ID(kanbanView.ID).Cols("default_bucket_id", "done_bucket_id").Update(kanbanView)
	return
}
