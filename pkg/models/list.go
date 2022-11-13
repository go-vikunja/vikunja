// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/db"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// Project represents a project of tasks
type Project struct {
	// The unique, numeric id of this project.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"project"`
	// The title of the project. You'll see this in the namespace overview.
	Title string `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250"`
	// The description of the project.
	Description string `xorm:"longtext null" json:"description"`
	// The unique project short identifier. Used to build task identifiers.
	Identifier string `xorm:"varchar(10) null" json:"identifier" valid:"runelength(0|10)" minLength:"0" maxLength:"10"`
	// The hex color of this project
	HexColor string `xorm:"varchar(6) null" json:"hex_color" valid:"runelength(0|6)" maxLength:"6"`

	OwnerID     int64 `xorm:"bigint INDEX not null" json:"-"`
	NamespaceID int64 `xorm:"bigint INDEX not null" json:"namespace_id" param:"namespace"`

	// The user who created this project.
	Owner *user.User `xorm:"-" json:"owner" valid:"-"`

	// Whether or not a project is archived.
	IsArchived bool `xorm:"not null default false" json:"is_archived" query:"is_archived"`

	// The id of the file this project has set as background
	BackgroundFileID int64 `xorm:"null" json:"-"`
	// Holds extra information about the background set since some background providers require attribution or similar. If not null, the background can be accessed at /projects/{projectID}/background
	BackgroundInformation interface{} `xorm:"-" json:"background_information"`
	// Contains a very small version of the project background to use as a blurry preview until the actual background is loaded. Check out https://blurha.sh/ to learn how it works.
	BackgroundBlurHash string `xorm:"varchar(50) null" json:"background_blur_hash"`

	// True if a project is a favorite. Favorite projects show up in a separate namespace. This value depends on the user making the call to the api.
	IsFavorite bool `xorm:"-" json:"is_favorite"`

	// The subscription status for the user reading this project. You can only read this property, use the subscription endpoints to modify it.
	// Will only returned when retreiving one project.
	Subscription *Subscription `xorm:"-" json:"subscription,omitempty"`

	// The position this project has when querying all projects. See the tasks.position property on how to use this.
	Position float64 `xorm:"double null" json:"position"`

	// A timestamp when this project was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this project was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

type ProjectWithTasksAndBuckets struct {
	Project
	// An array of tasks which belong to the project.
	Tasks []*TaskWithComments `xorm:"-" json:"tasks"`
	// Only used for migration.
	Buckets          []*Bucket `xorm:"-" json:"buckets"`
	BackgroundFileID int64     `xorm:"null" json:"background_file_id"`
}

// TableName returns a better name for the projects table
func (l *Project) TableName() string {
	return "projects"
}

// ProjectBackgroundType holds a project background type
type ProjectBackgroundType struct {
	Type string
}

// ProjectBackgroundUpload represents the project upload background type
const ProjectBackgroundUpload string = "upload"

// FavoritesPseudoProject holds all tasks marked as favorites
var FavoritesPseudoProject = Project{
	ID:          -1,
	Title:       "Favorites",
	Description: "This project has all tasks marked as favorites.",
	NamespaceID: FavoritesPseudoNamespace.ID,
	IsFavorite:  true,
	Created:     time.Now(),
	Updated:     time.Now(),
}

// GetProjectsByNamespaceID gets all projects in a namespace
func GetProjectsByNamespaceID(s *xorm.Session, nID int64, doer *user.User) (projects []*Project, err error) {
	switch nID {
	case SharedProjectsPseudoNamespace.ID:
		nnn, err := getSharedProjectsInNamespace(s, false, doer)
		if err != nil {
			return nil, err
		}
		if nnn != nil && nnn.Projects != nil {
			projects = nnn.Projects
		}
	case FavoritesPseudoNamespace.ID:
		namespaces := make(map[int64]*NamespaceWithProjects)
		_, err := getNamespacesWithProjects(s, &namespaces, "", false, 0, -1, doer.ID)
		if err != nil {
			return nil, err
		}
		namespaceIDs, _ := getNamespaceOwnerIDs(namespaces)
		ls, err := getProjectsForNamespaces(s, namespaceIDs, false)
		if err != nil {
			return nil, err
		}
		nnn, err := getFavoriteProjects(s, ls, namespaceIDs, doer)
		if err != nil {
			return nil, err
		}
		if nnn != nil && nnn.Projects != nil {
			projects = nnn.Projects
		}
	case SavedFiltersPseudoNamespace.ID:
		nnn, err := getSavedFilters(s, doer)
		if err != nil {
			return nil, err
		}
		if nnn != nil && nnn.Projects != nil {
			projects = nnn.Projects
		}
	default:
		err = s.Select("l.*").
			Alias("l").
			Join("LEFT", []string{"namespaces", "n"}, "l.namespace_id = n.id").
			Where("l.is_archived = false").
			Where("n.is_archived = false OR n.is_archived IS NULL").
			Where("namespace_id = ?", nID).
			Find(&projects)
	}
	if err != nil {
		return nil, err
	}

	// get more project details
	err = addProjectDetails(s, projects, doer)
	return projects, err
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
// @Security JWTKeyAuth
// @Success 200 {array} models.Project "The projects"
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects [get]
func (l *Project) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	// Check if we're dealing with a share auth
	shareAuth, ok := a.(*LinkSharing)
	if ok {
		project, err := GetProjectSimpleByID(s, shareAuth.ProjectID)
		if err != nil {
			return nil, 0, 0, err
		}
		projects := []*Project{project}
		err = addProjectDetails(s, projects, a)
		return projects, 0, 0, err
	}

	projects, resultCount, totalItems, err := getRawProjectsForUser(
		s,
		&projectOptions{
			search:     search,
			user:       &user.User{ID: a.GetID()},
			page:       page,
			perPage:    perPage,
			isArchived: l.IsArchived,
		})
	if err != nil {
		return nil, 0, 0, err
	}

	// Add more project details
	err = addProjectDetails(s, projects, a)
	return projects, resultCount, totalItems, err
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
func (l *Project) ReadOne(s *xorm.Session, a web.Auth) (err error) {

	if l.ID == FavoritesPseudoProject.ID {
		// Already "built" the project in CanRead
		return nil
	}

	// Check for saved filters
	if getSavedFilterIDFromProjectID(l.ID) > 0 {
		sf, err := getSavedFilterSimpleByID(s, getSavedFilterIDFromProjectID(l.ID))
		if err != nil {
			return err
		}
		l.Title = sf.Title
		l.Description = sf.Description
		l.Created = sf.Created
		l.Updated = sf.Updated
		l.OwnerID = sf.OwnerID
	}

	// Get project owner
	l.Owner, err = user.GetUserByID(s, l.OwnerID)
	if err != nil {
		return err
	}
	// Check if the namespace is archived and set the namespace to archived if it is not already archived individually.
	if !l.IsArchived {
		err = l.CheckIsArchived(s)
		if err != nil {
			if !IsErrNamespaceIsArchived(err) && !IsErrProjectIsArchived(err) {
				return
			}
			l.IsArchived = true
		}
	}

	// Get any background information if there is one set
	if l.BackgroundFileID != 0 {
		// Unsplash image
		l.BackgroundInformation, err = GetUnsplashPhotoByFileID(s, l.BackgroundFileID)
		if err != nil && !files.IsErrFileIsNotUnsplashFile(err) {
			return
		}

		if err != nil && files.IsErrFileIsNotUnsplashFile(err) {
			l.BackgroundInformation = &ProjectBackgroundType{Type: ProjectBackgroundUpload}
		}
	}

	l.IsFavorite, err = isFavorite(s, l.ID, a, FavoriteKindProject)
	if err != nil {
		return
	}

	l.Subscription, err = GetSubscription(s, SubscriptionEntityProject, l.ID, a)
	return
}

// GetProjectSimpleByID gets a project with only the basic items, aka no tasks or user objects. Returns an error if the project does not exist.
func GetProjectSimpleByID(s *xorm.Session, projectID int64) (project *Project, err error) {

	project = &Project{}

	if projectID < 1 {
		return nil, ErrProjectDoesNotExist{ID: projectID}
	}

	exists, err := s.
		Where("id = ?", projectID).
		OrderBy("position").
		Get(project)
	if err != nil {
		return
	}

	if !exists {
		return nil, ErrProjectDoesNotExist{ID: projectID}
	}

	return
}

// GetProjectSimplByTaskID gets a project by a task id
func GetProjectSimplByTaskID(s *xorm.Session, taskID int64) (l *Project, err error) {
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
		return &Project{}, ErrProjectDoesNotExist{ID: l.ID}
	}

	return &project, nil
}

// GetProjectsByIDs returns a map of projects from a slice with project ids
func GetProjectsByIDs(s *xorm.Session, projectIDs []int64) (projects map[int64]*Project, err error) {
	projects = make(map[int64]*Project, len(projectIDs))

	if len(projectIDs) == 0 {
		return
	}

	err = s.In("id", projectIDs).Find(&projects)
	return
}

type projectOptions struct {
	search     string
	user       *user.User
	page       int
	perPage    int
	isArchived bool
}

func getUserProjectsStatement(userID int64) *builder.Builder {
	dialect := config.DatabaseType.GetString()
	if dialect == "sqlite" {
		dialect = builder.SQLITE
	}

	return builder.Dialect(dialect).
		Select("l.*").
		From("projects", "l").
		Join("INNER", "namespaces n", "l.namespace_id = n.id").
		Join("LEFT", "team_namespaces tn", "tn.namespace_id = n.id").
		Join("LEFT", "team_members tm", "tm.team_id = tn.team_id").
		Join("LEFT", "team_projects tl", "l.id = tl.project_id").
		Join("LEFT", "team_members tm2", "tm2.team_id = tl.team_id").
		Join("LEFT", "users_projects ul", "ul.project_id = l.id").
		Join("LEFT", "users_namespaces un", "un.namespace_id = l.namespace_id").
		Where(builder.Or(
			builder.Eq{"tm.user_id": userID},
			builder.Eq{"tm2.user_id": userID},
			builder.Eq{"ul.user_id": userID},
			builder.Eq{"un.user_id": userID},
			builder.Eq{"l.owner_id": userID},
		)).
		OrderBy("position").
		GroupBy("l.id")
}

// Gets the projects only, without any tasks or so
func getRawProjectsForUser(s *xorm.Session, opts *projectOptions) (projects []*Project, resultCount int, totalItems int64, err error) {
	fullUser, err := user.GetUserByID(s, opts.user.ID)
	if err != nil {
		return nil, 0, 0, err
	}

	// Adding a 1=1 condition by default here because xorm always needs a condition and cannot handle nil conditions
	var isArchivedCond builder.Cond = builder.Eq{"1": 1}
	if !opts.isArchived {
		isArchivedCond = builder.And(
			builder.Eq{"l.is_archived": false},
			builder.Eq{"n.is_archived": false},
		)
	}

	limit, start := getLimitFromPageIndex(opts.page, opts.perPage)

	var filterCond builder.Cond
	ids := []int64{}
	if opts.search != "" {
		vals := strings.Split(opts.search, ",")
		for _, val := range vals {
			v, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				log.Debugf("Project search string part '%s' is not a number: %s", val, err)
				continue
			}
			ids = append(ids, v)
		}
	}

	filterCond = db.ILIKE("l.title", opts.search)
	if len(ids) > 0 {
		filterCond = builder.In("l.id", ids)
	}

	// Gets all Projects where the user is either owner or in a team which has access to the project
	// Or in a team which has namespace read access

	query := getUserProjectsStatement(fullUser.ID).
		Where(filterCond).
		Where(isArchivedCond)
	if limit > 0 {
		query = query.Limit(limit, start)
	}
	err = s.SQL(query).Find(&projects)
	if err != nil {
		return nil, 0, 0, err
	}

	query = getUserProjectsStatement(fullUser.ID).
		Where(filterCond).
		Where(isArchivedCond)
	totalItems, err = s.
		SQL(query.Select("count(*)")).
		Count(&Project{})
	return projects, len(projects), totalItems, err
}

// addProjectDetails adds owner user objects and project tasks to all projects in the slice
func addProjectDetails(s *xorm.Session, projects []*Project, a web.Auth) (err error) {
	if len(projects) == 0 {
		return
	}

	var ownerIDs []int64
	for _, l := range projects {
		ownerIDs = append(ownerIDs, l.OwnerID)
	}

	// Get all project owners
	owners := map[int64]*user.User{}
	if len(ownerIDs) > 0 {
		err = s.In("id", ownerIDs).Find(&owners)
		if err != nil {
			return
		}
	}

	var fileIDs []int64
	var projectIDs []int64
	for _, l := range projects {
		projectIDs = append(projectIDs, l.ID)
		if o, exists := owners[l.OwnerID]; exists {
			l.Owner = o
		}
		if l.BackgroundFileID != 0 {
			l.BackgroundInformation = &ProjectBackgroundType{Type: ProjectBackgroundUpload}
		}
		fileIDs = append(fileIDs, l.BackgroundFileID)
	}

	favs, err := getFavorites(s, projectIDs, a, FavoriteKindProject)
	if err != nil {
		return err
	}

	subscriptions, err := GetSubscriptions(s, SubscriptionEntityProject, projectIDs, a)
	if err != nil {
		log.Errorf("An error occurred while getting project subscriptions for a namespace item: %s", err.Error())
		subscriptions = make(map[int64]*Subscription)
	}

	for _, project := range projects {
		// Don't override the favorite state if it was already set from before (favorite saved filters do this)
		if project.IsFavorite {
			continue
		}
		project.IsFavorite = favs[project.ID]

		if subscription, exists := subscriptions[project.ID]; exists {
			project.Subscription = subscription
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

// NamespaceProject is a meta type to be able  to join a project with its namespace
type NamespaceProject struct {
	Project   Project   `xorm:"extends"`
	Namespace Namespace `xorm:"extends"`
}

// CheckIsArchived returns an ErrProjectIsArchived or ErrNamespaceIsArchived if the project or its namespace is archived.
func (l *Project) CheckIsArchived(s *xorm.Session) (err error) {
	// When creating a new project, we check if the namespace is archived
	if l.ID == 0 {
		n := &Namespace{ID: l.NamespaceID}
		return n.CheckIsArchived(s)
	}

	nl := &NamespaceProject{}
	exists, err := s.
		Table("projects").
		Join("LEFT", "namespaces", "projects.namespace_id = namespaces.id").
		Where("projects.id = ? AND (projects.is_archived = true OR namespaces.is_archived = true)", l.ID).
		Get(nl)
	if err != nil {
		return
	}
	if exists && nl.Project.ID != 0 && nl.Project.IsArchived {
		return ErrProjectIsArchived{ProjectID: l.ID}
	}
	if exists && nl.Namespace.ID != 0 && nl.Namespace.IsArchived {
		return ErrNamespaceIsArchived{NamespaceID: nl.Namespace.ID}
	}
	return nil
}

func checkProjectBeforeUpdateOrDelete(s *xorm.Session, project *Project) error {
	if project.NamespaceID < 0 {
		return &ErrProjectCannotBelongToAPseudoNamespace{ProjectID: project.ID, NamespaceID: project.NamespaceID}
	}

	// Check if the namespace exists
	if project.NamespaceID > 0 {
		_, err := GetNamespaceByID(s, project.NamespaceID)
		if err != nil {
			return err
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

func CreateProject(s *xorm.Session, project *Project, auth web.Auth) (err error) {
	err = project.CheckIsArchived(s)
	if err != nil {
		return err
	}

	doer, err := user.GetFromAuth(auth)
	if err != nil {
		return err
	}

	project.OwnerID = doer.ID
	project.Owner = doer
	project.ID = 0 // Otherwise only the first time a new project would be created

	err = checkProjectBeforeUpdateOrDelete(s, project)
	if err != nil {
		return
	}

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

	// Create a new first bucket for this project
	b := &Bucket{
		ProjectID: project.ID,
		Title:     "Backlog",
	}
	err = b.Create(s, auth)
	if err != nil {
		return
	}

	return events.Dispatch(&ProjectCreatedEvent{
		Project: project,
		Doer:    doer,
	})
}

func UpdateProject(s *xorm.Session, project *Project, auth web.Auth, updateProjectBackground bool) (err error) {
	err = checkProjectBeforeUpdateOrDelete(s, project)
	if err != nil {
		return
	}

	if project.NamespaceID == 0 {
		return &ErrProjectMustBelongToANamespace{
			ProjectID:   project.ID,
			NamespaceID: project.NamespaceID,
		}
	}

	// We need to specify the cols we want to update here to be able to un-archive projects
	colsToUpdate := []string{
		"title",
		"is_archived",
		"identifier",
		"hex_color",
		"namespace_id",
		"position",
	}
	if project.Description != "" {
		colsToUpdate = append(colsToUpdate, "description")
	}

	if updateProjectBackground {
		colsToUpdate = append(colsToUpdate, "background_file_id", "background_blur_hash")
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
func (l *Project) Update(s *xorm.Session, a web.Auth) (err error) {
	fid := getSavedFilterIDFromProjectID(l.ID)
	if fid > 0 {
		f, err := getSavedFilterSimpleByID(s, fid)
		if err != nil {
			return err
		}

		f.Title = l.Title
		f.Description = l.Description
		f.IsFavorite = l.IsFavorite
		err = f.Update(s, a)
		if err != nil {
			return err
		}

		*l = *f.toProject()
		return nil
	}

	return UpdateProject(s, l, a, false)
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
// @Description Creates a new project in a given namespace. The user needs write-access to the namespace.
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param namespaceID path int true "Namespace ID"
// @Param project body models.Project true "The project you want to create."
// @Success 201 {object} models.Project "The created project."
// @Failure 400 {object} web.HTTPError "Invalid project object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{namespaceID}/projects [put]
func (l *Project) Create(s *xorm.Session, a web.Auth) (err error) {
	err = CreateProject(s, l, a)
	if err != nil {
		return
	}

	return l.ReadOne(s, a)
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
func (l *Project) Delete(s *xorm.Session, a web.Auth) (err error) {

	fullList, err := GetProjectSimpleByID(s, l.ID)
	if err != nil {
		return
	}

	// Delete the project
	_, err = s.ID(l.ID).Delete(&Project{})
	if err != nil {
		return
	}

	// Delete all tasks on that project
	// Using the loop to make sure all related entities to all tasks are properly deleted as well.
	tasks, _, _, err := getRawTasksForProjects(s, []*Project{l}, a, &taskOptions{})
	if err != nil {
		return
	}

	for _, task := range tasks {
		err = task.Delete(s, a)
		if err != nil {
			return err
		}
	}

	err = fullList.DeleteBackgroundFileIfExists()
	if err != nil {
		return
	}

	return events.Dispatch(&ProjectDeletedEvent{
		Project: l,
		Doer:    a,
	})
}

// DeleteBackgroundFileIfExists deletes the list's background file from the db and the filesystem,
// if one exists
func (l *Project) DeleteBackgroundFileIfExists() (err error) {
	if l.BackgroundFileID == 0 {
		return
	}

	file := files.File{ID: l.BackgroundFileID}
	return file.Delete()
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
