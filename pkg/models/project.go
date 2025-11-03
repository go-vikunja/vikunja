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
	"math"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// Project represents a project of tasks
type Project struct {
	// The unique, numeric id of this project.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"project"`
	// The title of the project. You'll see this in the overview.
	Title string `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250"`
	// The description of the project.
	Description string `xorm:"longtext null" json:"description"`
	// The unique project short identifier. Used to build task identifiers.
	Identifier string `xorm:"varchar(10) null" json:"identifier" valid:"runelength(0|10)" minLength:"0" maxLength:"10"`
	// The hex color of this project
	HexColor string `xorm:"varchar(6) null" json:"hex_color" valid:"runelength(0|7)" maxLength:"7"`

	OwnerID         int64    `xorm:"bigint INDEX not null" json:"-"`
	ParentProjectID int64    `xorm:"bigint INDEX null" json:"parent_project_id"`
	ParentProject   *Project `xorm:"-" json:"-"`

	// The user who created this project.
	Owner *user.User `xorm:"-" json:"owner" valid:"-"`

	// Whether a project is archived.
	IsArchived bool `xorm:"not null default false" json:"is_archived" query:"is_archived"`

	// The id of the file this project has set as background
	BackgroundFileID int64 `xorm:"null" json:"-"`
	// Holds extra information about the background set since some background providers require attribution or similar. If not null, the background can be accessed at /projects/{projectID}/background
	BackgroundInformation interface{} `xorm:"-" json:"background_information"`
	// Contains a very small version of the project background to use as a blurry preview until the actual background is loaded. Check out https://blurha.sh/ to learn how it works.
	BackgroundBlurHash string `xorm:"varchar(50) null" json:"background_blur_hash"`

	// True if a project is a favorite. Favorite projects show up in a separate parent project. This value depends on the user making the call to the api.
	IsFavorite bool `xorm:"-" json:"is_favorite"`

	// The subscription status for the user reading this project. You can only read this property, use the subscription endpoints to modify it.
	// Will only returned when retreiving one project.
	Subscription *Subscription `xorm:"-" json:"subscription,omitempty"`

	// The position this project has when querying all projects. See the tasks.position property on how to use this.
	Position float64 `xorm:"double null" json:"position"`

	Views []*ProjectView `xorm:"-" json:"views"`

	Expand        ProjectExpandable `xorm:"-" json:"-" query:"expand"`
	MaxPermission Permission        `xorm:"-" json:"max_permission"`

	// A timestamp when this project was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this project was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

type ProjectExpandable string

const ProjectExpandableRights = `permissions`

type ProjectWithTasksAndBuckets struct {
	Project
	ChildProjects []*ProjectWithTasksAndBuckets `xorm:"-" json:"child_projects"`

	// An array of tasks which belong to the project.
	Tasks []*TaskWithComments `xorm:"-" json:"tasks"`
	// Only used for migration.
	Buckets          []*Bucket       `xorm:"-" json:"buckets"`
	TaskBuckets      []*TaskBucket   `xorm:"-" json:"task_buckets"`
	Positions        []*TaskPosition `xorm:"-" json:"positions"`
	BackgroundFileID int64           `xorm:"null" json:"background_file_id"`
}

// TableName returns a better name for the projects table
func (p *Project) TableName() string {
	return "projects"
}

// ProjectBackgroundType holds a project background type
type ProjectBackgroundType struct {
	Type string
}

// ProjectBackgroundUpload represents the project upload background type
const ProjectBackgroundUpload string = "upload"

const FavoritesPseudoProjectID = -1

// FavoritesPseudoProject holds all tasks marked as favorites
var FavoritesPseudoProject = Project{
	ID:          FavoritesPseudoProjectID,
	Title:       "Favorites",
	Description: "This project has all tasks marked as favorites.",
	IsFavorite:  true,
	Position:    -1,

	Views: []*ProjectView{
		{
			ID:        -1,
			ProjectID: FavoritesPseudoProjectID,
			Title:     "List",
			ViewKind:  ProjectViewKindList,
			Position:  100,
			Filter:    &TaskCollection{Filter: "done = false"},
		},
		{
			ID:        -2,
			ProjectID: FavoritesPseudoProjectID,
			Title:     "Gantt",
			ViewKind:  ProjectViewKindGantt,
			Position:  200,
		},
		{
			ID:        -3,
			ProjectID: FavoritesPseudoProjectID,
			Title:     "Table",
			ViewKind:  ProjectViewKindTable,
			Position:  300,
		},
	},

	Created: time.Now(),
	Updated: time.Now(),
}

// ReadAll gets all projects a user has access to
// @Summary Get all projects a user has access to
// @Description Returns all projects a user has access to.
// @tags project
// @Accept json
// @Produce json
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search projects by title."
// @Param is_archived query bool false "If true, also returns all archived projects."
// @Param expand query string false "If set to `permissions`, Vikunja will return the max permission the current user has on this project. You can currently only set this to `permissions`."
// @Security JWTKeyAuth
// @Success 200 {array} models.Project "The projects"
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects [get]
func (p *Project) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	prs, resultCount, totalItems, err := getAllRawProjects(s, a, search, page, perPage, p.IsArchived)
	if err != nil {
		return nil, 0, 0, err
	}

	_, is := a.(*LinkSharing)
	if is {
		// If we're dealing with a link share, we should just return the list of projects
		// (which will contain only one) here because it would not do anything meaningful afterward.
		return prs, resultCount, totalItems, nil
	}

	/////////////////
	// Add project details (favorite state, among other things)
	err = addProjectDetails(s, prs, a)
	if err != nil {
		return
	}

	if p.Expand == ProjectExpandableRights {
		var doer *user.User
		doer, err = user.GetFromAuth(a)
		if err != nil {
			return
		}
		err = addMaxPermissionToProjects(s, prs, doer)
		if err != nil {
			return
		}
	} else {
		for _, pr := range prs {
			pr.MaxPermission = PermissionUnknown
		}
	}

	//////////////////////////
	// Putting it all together

	return prs, resultCount, totalItems, err
}

func getAllRawProjects(s *xorm.Session, a web.Auth, search string, page int, perPage int, isArchived bool) (projects []*Project, resultCount int, totalItems int64, err error) {
	// Check if we're dealing with a share auth
	shareAuth, is := a.(*LinkSharing)
	if is {
		project, err := GetProjectSimpleByID(s, shareAuth.ProjectID)
		if err != nil {
			return nil, 0, 0, err
		}
		projects := []*Project{project}
		err = addProjectDetails(s, projects, a)
		if err == nil && len(projects) > 0 {
			projects[0].ParentProjectID = 0
		}
		return projects, 0, 0, err
	}

	doer, err := user.GetFromAuth(a)
	if err != nil {
		return nil, 0, 0, err
	}

	prs, resultCount, totalItems, err := getRawProjectsForUser(
		s,
		&projectOptions{
			search:      search,
			user:        doer,
			page:        page,
			perPage:     perPage,
			getArchived: isArchived,
		})
	if err != nil {
		return nil, 0, 0, err
	}

	/////////////////
	// Saved Filters

	savedFiltersProject, err := getSavedFilterProjects(s, doer, search)
	if err != nil {
		return nil, 0, 0, err
	}

	if len(savedFiltersProject) > 0 {
		prs = append(prs, savedFiltersProject...)
	}

	return prs, resultCount, totalItems, err
}

// ReadOne gets one project by its ID
// @Summary Gets one project
// @Description Returns a project by its ID.
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Project ID"
// @Success 200 {object} models.Project "The project"
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{id} [get]
func (p *Project) ReadOne(s *xorm.Session, a web.Auth) (err error) {

	if p.ID == FavoritesPseudoProject.ID {
		p.Views = FavoritesPseudoProject.Views
		// Already "built" the project in CanRead
		return nil
	}

	// Check for saved filters
	filterID := GetSavedFilterIDFromProjectID(p.ID)
	isFilter := filterID > 0
	if isFilter {
		sf, err := GetSavedFilterSimpleByID(s, filterID)
		if err != nil {
			return err
		}
		p.Title = sf.Title
		p.Description = sf.Description
		p.Created = sf.Created
		p.Updated = sf.Updated
		p.OwnerID = sf.OwnerID
	}

	_, isShareAuth := a.(*LinkSharing)
	if isShareAuth {
		p.ParentProjectID = 0
	}

	// Get project owner
	p.Owner, err = user.GetUserByID(s, p.OwnerID)
	if user.IsErrUserDoesNotExist(err) {
		p.Owner = nil
	} else if err != nil {
		return err
	}

	// Check if the project is archived and set it to archived if it is not already archived individually.
	if !p.IsArchived && !isFilter {
		err = p.CheckIsArchived(s)
		if err != nil {
			p.IsArchived = true
		}
	}

	// Get any background information if there is one set
	if p.BackgroundFileID != 0 {
		// Unsplash image
		p.BackgroundInformation, err = GetUnsplashPhotoByFileID(s, p.BackgroundFileID)
		if err != nil && !files.IsErrFileIsNotUnsplashFile(err) {
			return
		}

		if err != nil && files.IsErrFileIsNotUnsplashFile(err) {
			p.BackgroundInformation = &ProjectBackgroundType{Type: ProjectBackgroundUpload}
		}
	}

	p.IsFavorite, err = isFavorite(s, p.ID, a, FavoriteKindProject)
	if err != nil {
		return
	}

	subs, err := GetSubscriptionForUser(s, SubscriptionEntityProject, p.ID, a)
	if err != nil && IsErrProjectDoesNotExist(err) && isFilter {
		return nil
	}
	if subs != nil {
		p.Subscription = &subs.Subscription
	}

	p.Views, err = getViewsForProject(s, p.ID)
	return
}

// GetProjectSimpleByID gets a project with only the basic items, aka no tasks or user objects. Returns an error if the project does not exist.
func GetProjectSimpleByID(s *xorm.Session, projectID int64) (project *Project, err error) {
	if projectID < 1 {
		return nil, ErrProjectDoesNotExist{ID: projectID}
	}

	project, exists, err := getProjectSimple(s, builder.Eq{"id": projectID})
	if !exists {
		return nil, ErrProjectDoesNotExist{ID: projectID}
	}

	return
}

func getProjectSimple(s *xorm.Session, cond builder.Cond) (project *Project, exists bool, err error) {
	project = &Project{}
	exists, err = s.
		Where(cond).
		OrderBy("position").
		Get(project)
	return
}

// GetProjectSimpleByTaskID gets a project by a task id
func GetProjectSimpleByTaskID(s *xorm.Session, taskID int64) (l *Project, err error) {
	// We need to re-init our project object, because otherwise xorm creates a "where for every item in that project object,
	// leading to not finding anything if the id is good, but for example the title is different.
	var project Project
	exists, err := s.
		Select("projects.*").
		Table(Project{}).
		Join("INNER", "tasks", "projects.id = tasks.project_id").
		Where("tasks.id = ?", taskID).
		Get(&project)
	if err != nil {
		return
	}

	if !exists {
		return nil, ErrProjectDoesNotExist{}
	}

	return &project, nil
}

// GetProjectsMapSimpleByTaskIDs gets a list of projects by a task ids
func GetProjectsMapSimpleByTaskIDs(s *xorm.Session, taskIDs []int64) (ps map[int64]*Project, err error) {
	ps = make(map[int64]*Project)
	err = s.
		Select("projects.*").
		Table(Project{}).
		Join("INNER", "tasks", "projects.id = tasks.project_id").
		In("tasks.id", taskIDs).
		Find(&ps)
	return
}

func GetProjectsSimpleByTaskIDs(s *xorm.Session, taskIDs []int64) (ps []*Project, err error) {
	err = s.
		Select("projects.*").
		Table(Project{}).
		Join("INNER", "tasks", "projects.id = tasks.project_id").
		In("tasks.id", taskIDs).
		Find(&ps)
	return
}

// GetProjectsMapByIDs returns a map of projects from a slice with project ids
func GetProjectsMapByIDs(s *xorm.Session, projectIDs []int64) (projects map[int64]*Project, err error) {
	projects = make(map[int64]*Project, len(projectIDs))

	if len(projectIDs) == 0 {
		return
	}

	err = s.In("id", projectIDs).Find(&projects)
	return
}

func GetProjectsByIDs(s *xorm.Session, projectIDs []int64) (projects []*Project, err error) {
	projects = make([]*Project, 0, len(projectIDs))

	if len(projectIDs) == 0 {
		return
	}

	err = s.In("id", projectIDs).Find(&projects)
	return
}

type projectOptions struct {
	search      string
	user        *user.User
	page        int
	perPage     int
	getArchived bool
}

func getUserProjectsStatement(userID int64, search string, getArchived bool) *builder.Builder {
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

func getAllProjectsForUser(s *xorm.Session, userID int64, opts *projectOptions) (projects []*Project, totalCount int64, err error) {

	limit, start := getLimitFromPageIndex(opts.page, opts.perPage)
	query := getUserProjectsStatement(userID, opts.search, opts.getArchived)

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
	currentProjects := []*Project{}
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
		Count(&Project{})
	if err != nil {
		return nil, 0, err
	}

	return currentProjects, totalCount, err
}

// Gets the projects with their children without any tasks
func getRawProjectsForUser(s *xorm.Session, opts *projectOptions) (projects []*Project, resultCount int, totalItems int64, err error) {
	fullUser, err := user.GetUserByID(s, opts.user.ID)
	if err != nil {
		return nil, 0, 0, err
	}

	allProjects, totalItems, err := getAllProjectsForUser(s, fullUser.ID, opts)
	if err != nil {
		return
	}

	favoriteCount, err := s.
		Where(builder.And(
			builder.Eq{"user_id": opts.user.ID},
			builder.Eq{"kind": FavoriteKindTask},
		)).
		Count(&Favorite{})
	if err != nil {
		return
	}

	if favoriteCount > 0 {
		favoritesProject := &Project{}
		*favoritesProject = FavoritesPseudoProject
		allProjects = append(allProjects, favoritesProject)
	}

	if len(allProjects) == 0 {
		return nil, 0, totalItems, nil
	}

	return allProjects, len(allProjects), totalItems, err
}

func getSavedFilterProjects(s *xorm.Session, doer *user.User, search string) (savedFiltersProjects []*Project, err error) {
	savedFilters, err := getSavedFiltersForUser(s, doer, search)
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

// GetAllParentProjects returns all parents of a given project
func GetAllParentProjects(s *xorm.Session, projectID int64) (allProjects map[int64]*Project, err error) {
	allProjects = make(map[int64]*Project)
	err = s.SQL(`WITH RECURSIVE all_projects AS (
		    SELECT
		        p.*
		    FROM
		        projects p
		    WHERE
		        p.id = ?
		    UNION ALL
		    SELECT
		        p.*
		    FROM
		        projects p
		            INNER JOIN all_projects pc ON p.ID = pc.parent_project_id
		)
		SELECT DISTINCT * FROM all_projects`, projectID).Find(&allProjects)
	return
}

// addProjectDetails adds owner user objects and project tasks to all projects in the slice
func addProjectDetails(s *xorm.Session, projects []*Project, a web.Auth) (err error) {
	if len(projects) == 0 {
		return
	}

	var ownerIDs []int64
	var projectIDs []int64
	var fileIDs []int64
	for _, p := range projects {
		ownerIDs = append(ownerIDs, p.OwnerID)
		projectIDs = append(projectIDs, p.ID)
		fileIDs = append(fileIDs, p.BackgroundFileID)
	}

	owners, err := user.GetUsersByIDs(s, ownerIDs)
	if err != nil {
		return err
	}

	favs, err := getFavorites(s, projectIDs, a, FavoriteKindProject)
	if err != nil {
		return err
	}

	var subscriptions = make(map[int64][]*Subscription)
	u, is := a.(*user.User)
	if is {
		subscriptionsWithUser, err := GetSubscriptionsForEntitiesAndUser(s, SubscriptionEntityProject, projectIDs, u)
		if err != nil {
			log.Errorf("An error occurred while getting project subscriptions for a project: %s", err.Error())
		}
		if err == nil {
			for pID, subs := range subscriptionsWithUser {
				for _, sub := range subs {
					if _, has := subscriptions[pID]; !has {
						subscriptions[pID] = []*Subscription{}
					}
					subscriptions[pID] = append(subscriptions[pID], &sub.Subscription)
				}
			}
		}
	}

	views := []*ProjectView{}
	err = s.
		In("project_id", projectIDs).
		OrderBy("position asc").
		Find(&views)
	if err != nil {
		return
	}

	viewMap := make(map[int64][]*ProjectView)
	for _, v := range views {
		if _, has := viewMap[v.ProjectID]; !has {
			viewMap[v.ProjectID] = []*ProjectView{}
		}

		viewMap[v.ProjectID] = append(viewMap[v.ProjectID], v)
	}

	for _, p := range projects {
		if o, exists := owners[p.OwnerID]; exists {
			p.Owner = o
		}
		if p.BackgroundFileID != 0 {
			p.BackgroundInformation = &ProjectBackgroundType{Type: ProjectBackgroundUpload}
		}

		// Don't override the favorite state if it was already set from before (favorite saved filters do this)
		if p.IsFavorite {
			continue
		}
		p.IsFavorite = favs[p.ID]

		if subscription, exists := subscriptions[p.ID]; exists && len(subscription) > 0 {
			p.Subscription = subscription[0]
		}

		vs, has := viewMap[p.ID]
		if has {
			p.Views = vs
		}
	}

	if len(fileIDs) == 0 {
		return
	}

	// Unsplash background file info
	us := []*UnsplashPhoto{}
	err = s.In("file_id", fileIDs).Find(&us)
	if err != nil {
		return
	}
	unsplashPhotos := make(map[int64]*UnsplashPhoto, len(us))
	for _, u := range us {
		unsplashPhotos[u.FileID] = u
	}

	// Build it all into the projects slice
	for _, l := range projects {
		// Only override the file info if we have info for unsplash backgrounds
		if _, exists := unsplashPhotos[l.BackgroundFileID]; exists {
			l.BackgroundInformation = unsplashPhotos[l.BackgroundFileID]
		}
	}

	return
}

func addMaxPermissionToProjects(s *xorm.Session, projects []*Project, u *user.User) (err error) {
	projectIDs := make([]int64, 0, len(projects))
	for _, project := range projects {
		if GetSavedFilterIDFromProjectID(project.ID) > 0 {
			project.MaxPermission = PermissionAdmin
			continue
		}
		projectIDs = append(projectIDs, project.ID)
	}

	permissions, err := checkPermissionsForProjects(s, u, projectIDs)
	if err != nil {
		return err
	}

	for _, project := range projects {
		permission, has := permissions[project.ID]
		if has {
			project.MaxPermission = permission.MaxPermission
		}
	}

	return
}

// CheckIsArchived returns an ErrProjectIsArchived if the project or any of its parent projects is archived.
func (p *Project) CheckIsArchived(s *xorm.Session) (err error) {
	if p.ParentProjectID > 0 {
		p := &Project{ID: p.ParentProjectID}
		return p.CheckIsArchived(s)
	}

	if p.ID == 0 { // don't check new projects
		return nil
	}

	project, err := GetProjectSimpleByID(s, p.ID)
	if err != nil {
		return err
	}

	if project.IsArchived {
		return ErrProjectIsArchived{ProjectID: p.ID}
	}

	return nil
}

func checkProjectBeforeUpdateOrDelete(s *xorm.Session, project *Project) (err error) {
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

		allProjects, err := GetAllParentProjects(s, project.ParentProjectID)
		if err != nil {
			return err
		}

		var parent *Project
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
			Exist(&Project{})
		if err != nil {
			return err
		}
		if exists {
			return ErrProjectIdentifierIsNotUnique{Identifier: project.Identifier}
		}
	}

	return nil
}

func CreateProject(s *xorm.Session, project *Project, auth web.Auth, createBacklogBucket bool, createDefaultViews bool) (err error) {
	err = project.CheckIsArchived(s)
	if err != nil {
		return err
	}

	doer, err := user.GetFromAuth(auth)
	if err != nil {
		return err
	}

	project.ID = 0
	project.OwnerID = doer.ID
	project.Owner = doer

	err = checkProjectBeforeUpdateOrDelete(s, project)
	if err != nil {
		return
	}

	project.HexColor = utils.NormalizeHex(project.HexColor)

	_, err = s.Insert(project)
	if err != nil {
		return
	}

	project.Position = calculateDefaultPosition(project.ID, project.Position)
	_, err = s.Where("id = ?", project.ID).Update(project)
	if err != nil {
		return
	}
	if project.IsFavorite {
		if err := addToFavorites(s, project.ID, auth, FavoriteKindProject); err != nil {
			return err
		}
	}

	if createDefaultViews {
		err = CreateDefaultViewsForProject(s, project, auth, createBacklogBucket, true)
		if err != nil {
			return
		}
	}

	return events.Dispatch(&ProjectCreatedEvent{
		Project: project,
		Doer:    doer,
	})
}

// CreateNewProjectForUser creates a new inbox project for a user. To prevent import cycles, we can't do that
// directly in the user.Create function.
func CreateNewProjectForUser(s *xorm.Session, u *user.User) (err error) {
	p := &Project{
		Title: "Inbox",
	}
	err = p.Create(s, u)
	if err != nil {
		return err
	}

	if u.DefaultProjectID != 0 {
		return err
	}

	u.DefaultProjectID = p.ID
	_, err = user.UpdateUser(s, u, false)
	return err
}

func UpdateProject(s *xorm.Session, project *Project, auth web.Auth, updateProjectBackground bool) (err error) {
	err = checkProjectBeforeUpdateOrDelete(s, project)
	if err != nil {
		return
	}

	if project.IsArchived {
		isDefaultProject, err := project.isDefaultProject(s)
		if err != nil {
			return err
		}

		if isDefaultProject {
			return &ErrCannotArchiveDefaultProject{ProjectID: project.ID}
		}
	}

	err = setArchiveStateForProjectDescendants(s, project.ID, project.IsArchived)
	if err != nil {
		return err
	}

	// We need to specify the cols we want to update here to be able to un-archive projects
	colsToUpdate := []string{
		"title",
		"is_archived",
		"identifier",
		"hex_color",
		"parent_project_id",
		"position",
		"done_bucket_id",
		"default_bucket_id",
	}
	if project.Description != "" {
		colsToUpdate = append(colsToUpdate, "description")
	}

	if updateProjectBackground {
		colsToUpdate = append(colsToUpdate, "background_file_id", "background_blur_hash")
	}

	if project.Position < 0.1 {
		err = recalculateProjectPositions(s, project.ParentProjectID)
		if err != nil {
			return err
		}
	}

	wasFavorite, err := isFavorite(s, project.ID, auth, FavoriteKindProject)
	if err != nil {
		return err
	}
	if project.IsFavorite && !wasFavorite {
		if err := addToFavorites(s, project.ID, auth, FavoriteKindProject); err != nil {
			return err
		}
	}

	if !project.IsFavorite && wasFavorite {
		if err := removeFromFavorite(s, project.ID, auth, FavoriteKindProject); err != nil {
			return err
		}
	}

	project.HexColor = utils.NormalizeHex(project.HexColor)

	_, err = s.
		ID(project.ID).
		Cols(colsToUpdate...).
		Update(project)
	if err != nil {
		return err
	}

	err = events.Dispatch(&ProjectUpdatedEvent{
		Project: project,
		Doer:    auth,
	})
	if err != nil {
		return err
	}

	l, err := GetProjectSimpleByID(s, project.ID)
	if err != nil {
		return err
	}

	*project = *l
	err = project.ReadOne(s, auth)
	return
}

func recalculateProjectPositions(s *xorm.Session, parentProjectID int64) (err error) {

	allProjects := []*Project{}
	err = s.
		Where("parent_project_id = ?", parentProjectID).
		OrderBy("position asc").
		Find(&allProjects)
	if err != nil {
		return
	}

	maxPosition := math.Pow(2, 32)

	for i, project := range allProjects {

		currentPosition := maxPosition / float64(len(allProjects)) * (float64(i + 1))

		_, err = s.Cols("position").
			Where("id = ?", project.ID).
			Update(&Project{Position: currentPosition})
		if err != nil {
			return
		}
	}

	return
}

// Update implements the update method of CRUDable
// @Summary Updates a project
// @Description Updates a project. This does not include adding a task (see below).
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Project ID"
// @Param project body models.Project true "The project with updated values you want to update."
// @Success 200 {object} models.Project "The updated project."
// @Failure 400 {object} web.HTTPError "Invalid project object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{id} [post]
func (p *Project) Update(s *xorm.Session, a web.Auth) (err error) {
	fid := GetSavedFilterIDFromProjectID(p.ID)
	if fid > 0 {
		f, err := GetSavedFilterSimpleByID(s, fid)
		if err != nil {
			return err
		}

		f.Title = p.Title
		f.Description = p.Description
		f.IsFavorite = p.IsFavorite
		err = f.Update(s, a)
		if err != nil {
			return err
		}

		*p = *f.ToProject()
		return nil
	}

	return UpdateProject(s, p, a, false)
}

func updateProjectLastUpdated(s *xorm.Session, project *Project) error {
	_, err := s.ID(project.ID).Cols("updated").Update(project)
	return err
}

func updateProjectByTaskID(s *xorm.Session, taskID int64) (err error) {
	// need to get the task to update the project last updated timestamp
	task, err := GetTaskByIDSimple(s, taskID)
	if err != nil {
		return err
	}

	return updateProjectLastUpdated(s, &Project{ID: task.ProjectID})
}

// Create implements the create method of CRUDable
// @Summary Creates a new project
// @Description Creates a new project. If a parent project is provided the user needs to have write access to that project.
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project body models.Project true "The project you want to create."
// @Success 201 {object} models.Project "The created project."
// @Failure 400 {object} web.HTTPError "Invalid project object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects [put]
func (p *Project) Create(s *xorm.Session, a web.Auth) (err error) {
	err = CreateProject(s, p, a, true, true)
	if err != nil {
		return
	}

	fullProject, err := GetProjectSimpleByID(s, p.ID)
	if err != nil {
		return
	}

	return fullProject.ReadOne(s, a)
}

func (p *Project) isDefaultProject(s *xorm.Session) (is bool, err error) {
	return s.
		Where("default_project_id = ?", p.ID).
		Exist(&user.User{})
}

// Delete implements the delete method of CRUDable
// @Summary Deletes a project
// @Description Delets a project
// @tags project
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Project ID"
// @Success 200 {object} models.Message "The project was successfully deleted."
// @Failure 400 {object} web.HTTPError "Invalid project object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{id} [delete]
func (p *Project) Delete(s *xorm.Session, a web.Auth) (err error) {

	isDefaultProject, err := p.isDefaultProject(s)
	if err != nil {
		return err
	}
	// Owners should be allowed to delete the default project
	if isDefaultProject && p.OwnerID != a.GetID() {
		return &ErrCannotDeleteDefaultProject{ProjectID: p.ID}
	}

	// Delete all tasks on that project
	// Using the loop to make sure all related entities to all tasks are properly deleted as well.
	tasks, _, _, err := getRawTasksForProjects(s, []*Project{p}, a, &taskSearchOptions{})
	if err != nil {
		return
	}

	for _, task := range tasks {
		err = task.Delete(s, a)
		if err != nil {
			return err
		}
	}

	fullProject, err := GetProjectSimpleByID(s, p.ID)
	if err != nil {
		return
	}

	err = fullProject.DeleteBackgroundFileIfExists()
	if err != nil {
		return
	}

	// If we're deleting a default project, remove it as default
	if isDefaultProject {
		_, err = s.Where("default_project_id = ?", p.ID).
			Cols("default_project_id").
			Update(&user.User{DefaultProjectID: 0})
		if err != nil {
			return
		}
	}

	// Delete related project entities
	views, err := getViewsForProject(s, p.ID)
	if err != nil {
		return
	}
	viewIDs := []int64{}
	for _, v := range views {
		viewIDs = append(viewIDs, v.ID)
	}

	_, err = s.In("project_view_id", viewIDs).Delete(&Bucket{})
	if err != nil {
		return
	}

	_, err = s.In("id", viewIDs).Delete(&ProjectView{})
	if err != nil {
		return
	}

	err = removeFromFavorite(s, p.ID, a, FavoriteKindProject)
	if err != nil {
		return
	}

	_, err = s.Where("project_id = ?", p.ID).Delete(&LinkSharing{})
	if err != nil {
		return
	}

	_, err = s.Where("project_id = ?", p.ID).Delete(&ProjectUser{})
	if err != nil {
		return
	}

	_, err = s.Where("project_id = ?", p.ID).Delete(&TeamProject{})
	if err != nil {
		return
	}

	// Delete the project
	_, err = s.ID(p.ID).Delete(&Project{})
	if err != nil {
		return
	}

	err = events.Dispatch(&ProjectDeletedEvent{
		Project: fullProject,
		Doer:    a,
	})
	if err != nil {
		return
	}

	childProjects := []*Project{}
	err = s.Where("parent_project_id = ?", fullProject.ID).Find(&childProjects)
	if err != nil {
		return
	}

	for _, child := range childProjects {
		err = child.Delete(s, a)
		if err != nil {
			return
		}
	}

	return
}

// DeleteBackgroundFileIfExists deletes the list's background file from the db and the filesystem,
// if one exists
func (p *Project) DeleteBackgroundFileIfExists() (err error) {
	if p.BackgroundFileID == 0 {
		return
	}

	s := db.NewSession()
	defer s.Close()

	file := files.File{ID: p.BackgroundFileID}
	err = file.Delete(s)
	if err != nil && files.IsErrFileDoesNotExist(err) {
		return nil
	}

	return err
}

// SetProjectBackground sets a background file as project background in the db
func SetProjectBackground(s *xorm.Session, projectID int64, background *files.File, blurHash string) (err error) {
	l := &Project{
		ID:                 projectID,
		BackgroundFileID:   background.ID,
		BackgroundBlurHash: blurHash,
	}
	_, err = s.
		Where("id = ?", l.ID).
		Cols("background_file_id", "background_blur_hash").
		Update(l)
	return
}

// setArchiveStateForProjectDescendants uses a recursive CTE to find and set the archived status of all descendant projects.
func setArchiveStateForProjectDescendants(s *xorm.Session, parentProjectID int64, shouldBeArchived bool) error {
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
		Update(&Project{IsArchived: shouldBeArchived})
	if err != nil {
		log.Errorf("Error updating is_archived for descendant projects for parent ID %d to %t: %v", parentProjectID, shouldBeArchived, err)
		return fmt.Errorf("failed to update is_archived for descendant projects for parent ID %d to %t: %w", parentProjectID, shouldBeArchived, err)
	}
	return nil
}
