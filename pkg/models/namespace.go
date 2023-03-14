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
	"sort"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/db"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"

	"code.vikunja.io/web"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// Namespace holds informations about a namespace
type Namespace struct {
	// The unique, numeric id of this namespace.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"namespace"`
	// The name of this namespace.
	Title string `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250"`
	// The description of the namespace
	Description string `xorm:"longtext null" json:"description"`
	OwnerID     int64  `xorm:"bigint not null INDEX" json:"-"`

	// The hex color of this namespace
	HexColor string `xorm:"varchar(6) null" json:"hex_color" valid:"runelength(0|6)" maxLength:"6"`

	// Whether or not a namespace is archived.
	IsArchived bool `xorm:"not null default false" json:"is_archived" query:"is_archived"`

	// The user who owns this namespace
	Owner *user.User `xorm:"-" json:"owner" valid:"-"`

	// The subscription status for the user reading this namespace. You can only read this property, use the subscription endpoints to modify it.
	// Will only returned when retreiving one namespace.
	Subscription *Subscription `xorm:"-" json:"subscription,omitempty"`

	// A timestamp when this namespace was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this namespace was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	// If set to true, will only return the namespaces, not their projects.
	NamespacesOnly bool `xorm:"-" json:"-" query:"namespaces_only"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// SharedProjectsPseudoNamespace is a pseudo namespace used to hold shared projects
var SharedProjectsPseudoNamespace = Namespace{
	ID:          -1,
	Title:       "Shared Projects",
	Description: "Projects of other users shared with you via teams or directly.",
	Created:     time.Now(),
	Updated:     time.Now(),
}

// FavoritesPseudoNamespace is a pseudo namespace used to hold favorited projects and tasks
var FavoritesPseudoNamespace = Namespace{
	ID:          -2,
	Title:       "Favorites",
	Description: "Favorite projects and tasks.",
	Created:     time.Now(),
	Updated:     time.Now(),
}

// SavedFiltersPseudoNamespace is a pseudo namespace used to hold saved filters
var SavedFiltersPseudoNamespace = Namespace{
	ID:          -3,
	Title:       "Filters",
	Description: "Saved filters.",
	Created:     time.Now(),
	Updated:     time.Now(),
}

// TableName makes beautiful table names
func (Namespace) TableName() string {
	return "namespaces"
}

// GetSimpleByID gets a namespace without things like the owner, it more or less only checks if it exists.
func getNamespaceSimpleByID(s *xorm.Session, id int64) (namespace *Namespace, err error) {
	if id == 0 {
		return nil, ErrNamespaceDoesNotExist{ID: id}
	}

	// Get the namesapce with shared projects
	if id == -1 {
		return &SharedProjectsPseudoNamespace, nil
	}

	if id == FavoritesPseudoNamespace.ID {
		return &FavoritesPseudoNamespace, nil
	}

	if id == SavedFiltersPseudoNamespace.ID {
		return &SavedFiltersPseudoNamespace, nil
	}

	namespace = &Namespace{}

	exists, err := s.Where("id = ?", id).Get(namespace)
	if err != nil {
		return
	}
	if !exists {
		return nil, ErrNamespaceDoesNotExist{ID: id}
	}

	return
}

// GetNamespaceByID returns a namespace object by its ID
func GetNamespaceByID(s *xorm.Session, id int64) (namespace *Namespace, err error) {
	namespace, err = getNamespaceSimpleByID(s, id)
	if err != nil {
		return
	}

	// Get the namespace Owner
	namespace.Owner, err = user.GetUserByID(s, namespace.OwnerID)
	return
}

// CheckIsArchived returns an ErrNamespaceIsArchived if the namepace is archived.
func (n *Namespace) CheckIsArchived(s *xorm.Session) error {
	exists, err := s.
		Where("id = ? AND is_archived = true", n.ID).
		Exist(&Namespace{})
	if err != nil {
		return err
	}
	if exists {
		return ErrNamespaceIsArchived{NamespaceID: n.ID}
	}
	return nil
}

// ReadOne gets one namespace
// @Summary Gets one namespace
// @Description Returns a namespace by its ID.
// @tags namespace
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Namespace ID"
// @Success 200 {object} models.Namespace "The Namespace"
// @Failure 403 {object} web.HTTPError "The user does not have access to that namespace."
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{id} [get]
func (n *Namespace) ReadOne(s *xorm.Session, a web.Auth) (err error) {
	nn, err := GetNamespaceByID(s, n.ID)
	if err != nil {
		return err
	}
	*n = *nn

	n.Subscription, err = GetSubscription(s, SubscriptionEntityNamespace, n.ID, a)
	return
}

// NamespaceWithProjects represents a namespace with project meta informations
type NamespaceWithProjects struct {
	Namespace `xorm:"extends"`
	Projects  []*Project `xorm:"-" json:"projects"`
}

type NamespaceWithProjectsAndTasks struct {
	Namespace
	Projects []*ProjectWithTasksAndBuckets `xorm:"-" json:"projects"`
}

func makeNamespaceSlice(namespaces map[int64]*NamespaceWithProjects, userMap map[int64]*user.User, subscriptions map[int64]*Subscription) []*NamespaceWithProjects {
	all := make([]*NamespaceWithProjects, 0, len(namespaces))
	for _, n := range namespaces {
		n.Owner = userMap[n.OwnerID]
		n.Subscription = subscriptions[n.ID]
		all = append(all, n)
		for _, l := range n.Projects {
			if n.Subscription != nil && l.Subscription == nil {
				l.Subscription = n.Subscription
			}
		}
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].ID < all[j].ID
	})

	return all
}

func getNamespaceFilterCond(search string) (filterCond builder.Cond) {
	filterCond = db.ILIKE("namespaces.title", search)

	if search == "" {
		return
	}

	vals := strings.Split(search, ",")

	if len(vals) == 0 {
		return
	}

	ids := []int64{}
	for _, val := range vals {
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			log.Debugf("Namespace search string part '%s' is not a number: %s", val, err)
			continue
		}
		ids = append(ids, v)
	}

	if len(ids) > 0 {
		filterCond = builder.In("namespaces.id", ids)
	}

	return
}

func getNamespaceArchivedCond(archived bool) builder.Cond {
	// Adding a 1=1 condition by default here because xorm always needs a condition and cannot handle nil conditions
	var isArchivedCond builder.Cond = builder.Eq{"1": 1}
	if !archived {
		isArchivedCond = builder.And(
			builder.Eq{"namespaces.is_archived": false},
		)
	}

	return isArchivedCond
}

func getNamespacesWithProjects(s *xorm.Session, namespaces *map[int64]*NamespaceWithProjects, search string, isArchived bool, page, perPage int, userID int64) (numberOfTotalItems int64, err error) {
	isArchivedCond := getNamespaceArchivedCond(isArchived)
	filterCond := getNamespaceFilterCond(search)

	limit, start := getLimitFromPageIndex(page, perPage)
	query := s.Select("namespaces.*").
		Table("namespaces").
		Join("LEFT", "team_namespaces", "namespaces.id = team_namespaces.namespace_id").
		Join("LEFT", "team_members", "team_members.team_id = team_namespaces.team_id").
		Join("LEFT", "users_namespaces", "users_namespaces.namespace_id = namespaces.id").
		Where("team_members.user_id = ?", userID).
		Or("namespaces.owner_id = ?", userID).
		Or("users_namespaces.user_id = ?", userID).
		GroupBy("namespaces.id").
		Where(filterCond).
		Where(isArchivedCond)
	if limit > 0 {
		query = query.Limit(limit, start)
	}
	err = query.Find(namespaces)
	if err != nil {
		return 0, err
	}

	numberOfTotalItems, err = s.
		Table("namespaces").
		Join("LEFT", "team_namespaces", "namespaces.id = team_namespaces.namespace_id").
		Join("LEFT", "team_members", "team_members.team_id = team_namespaces.team_id").
		Join("LEFT", "users_namespaces", "users_namespaces.namespace_id = namespaces.id").
		Where("team_members.user_id = ?", userID).
		Or("namespaces.owner_id = ?", userID).
		Or("users_namespaces.user_id = ?", userID).
		And("namespaces.is_archived = false").
		GroupBy("namespaces.id").
		Where(filterCond).
		Where(isArchivedCond).
		Count(&NamespaceWithProjects{})
	return numberOfTotalItems, err
}

func getNamespaceOwnerIDs(namespaces map[int64]*NamespaceWithProjects) (namespaceIDs, ownerIDs []int64) {
	for _, nsp := range namespaces {
		namespaceIDs = append(namespaceIDs, nsp.ID)
		ownerIDs = append(ownerIDs, nsp.OwnerID)
	}

	return
}

func getNamespaceSubscriptions(s *xorm.Session, namespaceIDs []int64, userID int64) (map[int64]*Subscription, error) {
	subscriptionsMap := make(map[int64]*Subscription)
	if len(namespaceIDs) == 0 {
		return subscriptionsMap, nil
	}

	subscriptions := []*Subscription{}
	err := s.
		Where("entity_type = ? AND user_id = ?", SubscriptionEntityNamespace, userID).
		In("entity_id", namespaceIDs).
		Find(&subscriptions)
	if err != nil {
		return nil, err
	}
	for _, sub := range subscriptions {
		sub.Entity = sub.EntityType.String()
		subscriptionsMap[sub.EntityID] = sub
	}

	return subscriptionsMap, err
}

func getProjectsForNamespaces(s *xorm.Session, namespaceIDs []int64, archived bool) ([]*Project, error) {
	projects := []*Project{}
	projectQuery := s.
		OrderBy("position").
		In("namespace_id", namespaceIDs)

	if !archived {
		projectQuery.And("is_archived = false")
	}
	err := projectQuery.Find(&projects)
	return projects, err
}

func getSharedProjectsInNamespace(s *xorm.Session, archived bool, doer *user.User) (sharedProjectsNamespace *NamespaceWithProjects, err error) {
	// Create our pseudo namespace to hold the shared projects
	sharedProjectsPseudonamespace := SharedProjectsPseudoNamespace
	sharedProjectsPseudonamespace.OwnerID = doer.ID
	sharedProjectsNamespace = &NamespaceWithProjects{
		sharedProjectsPseudonamespace,
		[]*Project{},
	}

	// Get all projects individually shared with our user (not via a namespace)
	individualProjects := []*Project{}
	iProjectQuery := s.Select("l.*").
		Table("projects").
		Alias("l").
		Join("LEFT", []string{"team_projects", "tl"}, "l.id = tl.project_id").
		Join("LEFT", []string{"team_members", "tm"}, "tm.team_id = tl.team_id").
		Join("LEFT", []string{"users_projects", "ul"}, "ul.project_id = l.id").
		Where(builder.And(
			builder.Eq{"tm.user_id": doer.ID},
			builder.Neq{"l.owner_id": doer.ID},
		)).
		Or(builder.And(
			builder.Eq{"ul.user_id": doer.ID},
			builder.Neq{"l.owner_id": doer.ID},
		)).
		GroupBy("l.id")
	if !archived {
		iProjectQuery.And("l.is_archived = false")
	}
	err = iProjectQuery.Find(&individualProjects)
	if err != nil {
		return
	}

	// Make the namespace -1 so we now later which one it was
	// + Append it to all projects we already have
	for _, l := range individualProjects {
		l.NamespaceID = sharedProjectsNamespace.ID
	}

	sharedProjectsNamespace.Projects = individualProjects

	// Remove the sharedProjectsPseudonamespace if we don't have any shared projects
	if len(individualProjects) == 0 {
		sharedProjectsNamespace = nil
	}

	return
}

func getFavoriteProjects(s *xorm.Session, projects []*Project, namespaceIDs []int64, doer *user.User) (favoriteNamespace *NamespaceWithProjects, err error) {
	// Create our pseudo namespace with favorite projects
	pseudoFavoriteNamespace := FavoritesPseudoNamespace
	pseudoFavoriteNamespace.OwnerID = doer.ID
	favoriteNamespace = &NamespaceWithProjects{
		Namespace: pseudoFavoriteNamespace,
		Projects:  []*Project{{}},
	}
	*favoriteNamespace.Projects[0] = FavoritesPseudoProject // Copying the project to be able to modify it later
	favoriteNamespace.Projects[0].Owner = doer

	for _, project := range projects {
		if !project.IsFavorite {
			continue
		}
		favoriteNamespace.Projects = append(favoriteNamespace.Projects, project)
	}

	// Check if we have any favorites or favorited projects and remove the favorites namespace from the project if not
	cond := builder.
		Select("tasks.id").
		From("tasks").
		Join("INNER", "projects", "tasks.project_id = projects.id").
		Join("INNER", "namespaces", "projects.namespace_id = namespaces.id").
		Where(builder.In("namespaces.id", namespaceIDs))

	var favoriteCount int64
	favoriteCount, err = s.
		Where(builder.And(
			builder.Eq{"user_id": doer.ID},
			builder.Eq{"kind": FavoriteKindTask},
			builder.In("entity_id", cond),
		)).
		Count(&Favorite{})
	if err != nil {
		return
	}

	// If we don't have any favorites in the favorites pseudo project, remove that pseudo project from the namespace
	if favoriteCount == 0 {
		for in, l := range favoriteNamespace.Projects {
			if l.ID == FavoritesPseudoProject.ID {
				favoriteNamespace.Projects = append(favoriteNamespace.Projects[:in], favoriteNamespace.Projects[in+1:]...)
				break
			}
		}
	}

	// If we don't have any favorites in the namespace, remove it
	if len(favoriteNamespace.Projects) == 0 {
		return nil, nil
	}

	return
}

func getSavedFilters(s *xorm.Session, doer *user.User) (savedFiltersNamespace *NamespaceWithProjects, err error) {
	savedFilters, err := getSavedFiltersForUser(s, doer)
	if err != nil {
		return
	}

	if len(savedFilters) == 0 {
		return nil, nil
	}

	savedFiltersPseudoNamespace := SavedFiltersPseudoNamespace
	savedFiltersPseudoNamespace.OwnerID = doer.ID
	savedFiltersNamespace = &NamespaceWithProjects{
		Namespace: savedFiltersPseudoNamespace,
		Projects:  make([]*Project, 0, len(savedFilters)),
	}

	for _, filter := range savedFilters {
		filterProject := filter.toProject()
		filterProject.NamespaceID = savedFiltersNamespace.ID
		filterProject.Owner = doer
		savedFiltersNamespace.Projects = append(savedFiltersNamespace.Projects, filterProject)
	}

	return
}

// ReadAll gets all namespaces a user has access to
// @Summary Get all namespaces a user has access to
// @Description Returns all namespaces a user has access to.
// @tags namespace
// @Accept json
// @Produce json
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search namespaces by name."
// @Param is_archived query bool false "If true, also returns all archived namespaces."
// @Param namespaces_only query bool false "If true, also returns only namespaces without their projects."
// @Security JWTKeyAuth
// @Success 200 {array} models.NamespaceWithProjects "The Namespaces."
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces [get]
//
//nolint:gocyclo
func (n *Namespace) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	if _, is := a.(*LinkSharing); is {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	// This map will hold all namespaces and their projects. The key is usually the id of the namespace.
	// We're using a map here because it makes a few things like adding projects or removing pseudo namespaces easier.
	namespaces := make(map[int64]*NamespaceWithProjects)

	//////////////////////////////
	// Projects with their namespaces

	doer, err := user.GetFromAuth(a)
	if err != nil {
		return nil, 0, 0, err
	}

	numberOfTotalItems, err = getNamespacesWithProjects(s, &namespaces, search, n.IsArchived, page, perPage, doer.ID)
	if err != nil {
		return nil, 0, 0, err
	}

	namespaceIDs, ownerIDs := getNamespaceOwnerIDs(namespaces)

	if len(namespaceIDs) == 0 {
		return nil, 0, 0, nil
	}

	subscriptionsMap, err := getNamespaceSubscriptions(s, namespaceIDs, doer.ID)
	if err != nil {
		return nil, 0, 0, err
	}

	ownerMap, err := user.GetUsersByIDs(s, ownerIDs)
	if err != nil {
		return nil, 0, 0, err
	}
	ownerMap[doer.ID] = doer

	if n.NamespacesOnly {
		all := makeNamespaceSlice(namespaces, ownerMap, subscriptionsMap)
		return all, len(all), numberOfTotalItems, nil
	}

	// Get all projects
	projects, err := getProjectsForNamespaces(s, namespaceIDs, n.IsArchived)
	if err != nil {
		return nil, 0, 0, err
	}

	///////////////
	// Shared Projects

	sharedProjectsNamespace, err := getSharedProjectsInNamespace(s, n.IsArchived, doer)
	if err != nil {
		return nil, 0, 0, err
	}

	if sharedProjectsNamespace != nil {
		namespaces[sharedProjectsNamespace.ID] = sharedProjectsNamespace
		projects = append(projects, sharedProjectsNamespace.Projects...)
	}

	/////////////////
	// Saved Filters

	savedFiltersNamespace, err := getSavedFilters(s, doer)
	if err != nil {
		return nil, 0, 0, err
	}

	if savedFiltersNamespace != nil {
		namespaces[savedFiltersNamespace.ID] = savedFiltersNamespace
		projects = append(projects, savedFiltersNamespace.Projects...)
	}

	/////////////////
	// Add project details (favorite state, among other things)
	err = addProjectDetails(s, projects, a)
	if err != nil {
		return
	}

	/////////////////
	// Favorite projects

	favoritesNamespace, err := getFavoriteProjects(s, projects, namespaceIDs, doer)
	if err != nil {
		return nil, 0, 0, err
	}

	if favoritesNamespace != nil {
		namespaces[favoritesNamespace.ID] = favoritesNamespace
	}

	//////////////////////
	// Put it all together

	for _, project := range projects {
		if project.NamespaceID == SharedProjectsPseudoNamespace.ID || project.NamespaceID == SavedFiltersPseudoNamespace.ID {
			// Shared projects and filtered projects are already in the namespace
			continue
		}
		namespaces[project.NamespaceID].Projects = append(namespaces[project.NamespaceID].Projects, project)
	}

	all := makeNamespaceSlice(namespaces, ownerMap, subscriptionsMap)
	return all, len(all), numberOfTotalItems, err
}

// Create implements the creation method via the interface
// @Summary Creates a new namespace
// @Description Creates a new namespace.
// @tags namespace
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param namespace body models.Namespace true "The namespace you want to create."
// @Success 201 {object} models.Namespace "The created namespace."
// @Failure 400 {object} web.HTTPError "Invalid namespace object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the namespace"
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces [put]
func (n *Namespace) Create(s *xorm.Session, a web.Auth) (err error) {
	// Check if we have at least a title
	if n.Title == "" {
		return ErrNamespaceNameCannotBeEmpty{NamespaceID: 0, UserID: a.GetID()}
	}

	n.Owner, err = user.GetUserByID(s, a.GetID())
	if err != nil {
		return
	}
	n.OwnerID = n.Owner.ID

	if _, err = s.Insert(n); err != nil {
		return err
	}

	err = events.Dispatch(&NamespaceCreatedEvent{
		Namespace: n,
		Doer:      a,
	})
	if err != nil {
		return err
	}

	return
}

// CreateNewNamespaceForUser creates a new namespace for a user. To prevent import cycles, we can't do that
// directly in the user.Create function.
func CreateNewNamespaceForUser(s *xorm.Session, user *user.User) (err error) {
	newN := &Namespace{
		Title:       user.Username,
		Description: user.Username + "'s namespace.",
	}
	return newN.Create(s, user)
}

// Delete deletes a namespace
// @Summary Deletes a namespace
// @Description Delets a namespace
// @tags namespace
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Namespace ID"
// @Success 200 {object} models.Message "The namespace was successfully deleted."
// @Failure 400 {object} web.HTTPError "Invalid namespace object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the namespace"
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{id} [delete]
func (n *Namespace) Delete(s *xorm.Session, a web.Auth) (err error) {
	return deleteNamespace(s, n, a, true)
}

func deleteNamespace(s *xorm.Session, n *Namespace, a web.Auth, withProjects bool) (err error) {
	// Check if the namespace exists
	_, err = GetNamespaceByID(s, n.ID)
	if err != nil {
		return
	}

	// Delete the namespace
	_, err = s.ID(n.ID).Delete(&Namespace{})
	if err != nil {
		return
	}

	namespaceDeleted := &NamespaceDeletedEvent{
		Namespace: n,
		Doer:      a,
	}

	if !withProjects {
		return events.Dispatch(namespaceDeleted)
	}

	// Delete all projects with their tasks
	projects, err := GetProjectsByNamespaceID(s, n.ID, &user.User{})
	if err != nil {
		return
	}

	if len(projects) == 0 {
		return events.Dispatch(namespaceDeleted)
	}

	// Looping over all projects to let the project handle properly cleaning up the tasks and everything else associated with it.
	for _, project := range projects {
		err = project.Delete(s, a)
		if err != nil {
			return err
		}
	}

	return events.Dispatch(namespaceDeleted)
}

// Update implements the update method via the interface
// @Summary Updates a namespace
// @Description Updates a namespace.
// @tags namespace
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Namespace ID"
// @Param namespace body models.Namespace true "The namespace with updated values you want to update."
// @Success 200 {object} models.Namespace "The updated namespace."
// @Failure 400 {object} web.HTTPError "Invalid namespace object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the namespace"
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespace/{id} [post]
func (n *Namespace) Update(s *xorm.Session, a web.Auth) (err error) {
	// Check if we have at least a name
	if n.Title == "" {
		return ErrNamespaceNameCannotBeEmpty{NamespaceID: n.ID}
	}

	// Check if the namespace exists
	currentNamespace, err := GetNamespaceByID(s, n.ID)
	if err != nil {
		return
	}

	// Check if the namespace is archived and the update is not un-archiving it
	if currentNamespace.IsArchived && n.IsArchived {
		return ErrNamespaceIsArchived{NamespaceID: n.ID}
	}

	// Check if the (new) owner exists
	if n.Owner != nil {
		n.OwnerID = n.Owner.ID
		if currentNamespace.OwnerID != n.OwnerID {
			n.Owner, err = user.GetUserByID(s, n.OwnerID)
			if err != nil {
				return
			}
		}
	}

	// We need to specify the cols we want to update here to be able to un-archive projects
	colsToUpdate := []string{
		"title",
		"is_archived",
		"hex_color",
	}
	if n.Description != "" {
		colsToUpdate = append(colsToUpdate, "description")
	}

	// Do the actual update
	_, err = s.
		ID(currentNamespace.ID).
		Cols(colsToUpdate...).
		Update(n)
	if err != nil {
		return err
	}

	return events.Dispatch(&NamespaceUpdatedEvent{
		Namespace: n,
		Doer:      a,
	})
}
