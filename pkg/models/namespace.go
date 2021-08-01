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

	// If set to true, will only return the namespaces, not their lists.
	NamespacesOnly bool `xorm:"-" json:"-" query:"namespaces_only"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// SharedListsPseudoNamespace is a pseudo namespace used to hold shared lists
var SharedListsPseudoNamespace = Namespace{
	ID:          -1,
	Title:       "Shared Lists",
	Description: "Lists of other users shared with you via teams or directly.",
	Created:     time.Now(),
	Updated:     time.Now(),
}

// FavoritesPseudoNamespace is a pseudo namespace used to hold favorited lists and tasks
var FavoritesPseudoNamespace = Namespace{
	ID:          -2,
	Title:       "Favorites",
	Description: "Favorite lists and tasks.",
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

	// Get the namesapce with shared lists
	if id == -1 {
		return &SharedListsPseudoNamespace, nil
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

// NamespaceWithLists represents a namespace with list meta informations
type NamespaceWithLists struct {
	Namespace `xorm:"extends"`
	Lists     []*List `xorm:"-" json:"lists"`
}

func makeNamespaceSlice(namespaces map[int64]*NamespaceWithLists, userMap map[int64]*user.User, subscriptions map[int64]*Subscription) []*NamespaceWithLists {
	all := make([]*NamespaceWithLists, 0, len(namespaces))
	for _, n := range namespaces {
		n.Owner = userMap[n.OwnerID]
		n.Subscription = subscriptions[n.ID]
		all = append(all, n)
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

func getNamespacesWithLists(s *xorm.Session, namespaces *map[int64]*NamespaceWithLists, search string, isArchived bool, page, perPage int, userID int64) (numberOfTotalItems int64, err error) {
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
		Count(&NamespaceWithLists{})
	return numberOfTotalItems, err
}

func getNamespaceOwnerIDs(namespaces map[int64]*NamespaceWithLists) (namespaceIDs, ownerIDs []int64) {
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

func getListsForNamespaces(s *xorm.Session, namespaceIDs []int64, archived bool) ([]*List, error) {
	lists := []*List{}
	listQuery := s.
		OrderBy("position").
		In("namespace_id", namespaceIDs)

	if !archived {
		listQuery.And("is_archived = false")
	}
	err := listQuery.Find(&lists)
	return lists, err
}

func getSharedListsInNamespace(s *xorm.Session, archived bool, doer *user.User) (sharedListsNamespace *NamespaceWithLists, err error) {
	// Create our pseudo namespace to hold the shared lists
	sharedListsPseudonamespace := SharedListsPseudoNamespace
	sharedListsPseudonamespace.Owner = doer
	sharedListsNamespace = &NamespaceWithLists{
		sharedListsPseudonamespace,
		[]*List{},
	}

	// Get all lists individually shared with our user (not via a namespace)
	individualLists := []*List{}
	iListQuery := s.Select("l.*").
		Table("lists").
		Alias("l").
		Join("LEFT", []string{"team_lists", "tl"}, "l.id = tl.list_id").
		Join("LEFT", []string{"team_members", "tm"}, "tm.team_id = tl.team_id").
		Join("LEFT", []string{"users_lists", "ul"}, "ul.list_id = l.id").
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
		iListQuery.And("l.is_archived = false")
	}
	err = iListQuery.Find(&individualLists)
	if err != nil {
		return
	}

	// Make the namespace -1 so we now later which one it was
	// + Append it to all lists we already have
	for _, l := range individualLists {
		l.NamespaceID = sharedListsNamespace.ID
	}

	sharedListsNamespace.Lists = individualLists

	// Remove the sharedListsPseudonamespace if we don't have any shared lists
	if len(individualLists) == 0 {
		sharedListsNamespace = nil
	}

	return
}

func getFavoriteLists(s *xorm.Session, lists []*List, namespaceIDs []int64, doer *user.User) (favoriteNamespace *NamespaceWithLists, err error) {
	// Create our pseudo namespace with favorite lists
	pseudoFavoriteNamespace := FavoritesPseudoNamespace
	pseudoFavoriteNamespace.Owner = doer
	favoriteNamespace = &NamespaceWithLists{
		Namespace: pseudoFavoriteNamespace,
		Lists:     []*List{{}},
	}
	*favoriteNamespace.Lists[0] = FavoritesPseudoList // Copying the list to be able to modify it later

	for _, list := range lists {
		if !list.IsFavorite {
			continue
		}
		favoriteNamespace.Lists = append(favoriteNamespace.Lists, list)
	}

	// Check if we have any favorites or favorited lists and remove the favorites namespace from the list if not
	cond := builder.
		Select("tasks.id").
		From("tasks").
		Join("INNER", "lists", "tasks.list_id = lists.id").
		Join("INNER", "namespaces", "lists.namespace_id = namespaces.id").
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

	// If we don't have any favorites in the favorites pseudo list, remove that pseudo list from the namespace
	if favoriteCount == 0 {
		for in, l := range favoriteNamespace.Lists {
			if l.ID == FavoritesPseudoList.ID {
				favoriteNamespace.Lists = append(favoriteNamespace.Lists[:in], favoriteNamespace.Lists[in+1:]...)
				break
			}
		}
	}

	// If we don't have any favorites in the namespace, remove it
	if len(favoriteNamespace.Lists) == 0 {
		return nil, nil
	}

	return
}

func getSavedFilters(s *xorm.Session, doer *user.User) (savedFiltersNamespace *NamespaceWithLists, err error) {
	savedFilters, err := getSavedFiltersForUser(s, doer)
	if err != nil {
		return
	}

	if len(savedFilters) == 0 {
		return nil, nil
	}

	savedFiltersPseudoNamespace := SavedFiltersPseudoNamespace
	savedFiltersPseudoNamespace.Owner = doer
	savedFiltersNamespace = &NamespaceWithLists{
		Namespace: savedFiltersPseudoNamespace,
		Lists:     make([]*List, 0, len(savedFilters)),
	}

	for _, filter := range savedFilters {
		filterList := filter.toList()
		filterList.NamespaceID = savedFiltersNamespace.ID
		filterList.Owner = doer
		savedFiltersNamespace.Lists = append(savedFiltersNamespace.Lists, filterList)
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
// @Param namespaces_only query bool false "If true, also returns only namespaces without their lists."
// @Security JWTKeyAuth
// @Success 200 {array} models.NamespaceWithLists "The Namespaces."
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces [get]
//nolint:gocyclo
func (n *Namespace) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	if _, is := a.(*LinkSharing); is {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	// This map will hold all namespaces and their lists. The key is usually the id of the namespace.
	// We're using a map here because it makes a few things like adding lists or removing pseudo namespaces easier.
	namespaces := make(map[int64]*NamespaceWithLists)

	//////////////////////////////
	// Lists with their namespaces

	doer, err := user.GetFromAuth(a)
	if err != nil {
		return nil, 0, 0, err
	}

	numberOfTotalItems, err = getNamespacesWithLists(s, &namespaces, search, n.IsArchived, page, perPage, doer.ID)
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

	if n.NamespacesOnly {
		all := makeNamespaceSlice(namespaces, ownerMap, subscriptionsMap)
		return all, len(all), numberOfTotalItems, nil
	}

	// Get all lists
	lists, err := getListsForNamespaces(s, namespaceIDs, n.IsArchived)
	if err != nil {
		return nil, 0, 0, err
	}

	///////////////
	// Shared Lists

	sharedListsNamespace, err := getSharedListsInNamespace(s, n.IsArchived, doer)
	if err != nil {
		return nil, 0, 0, err
	}

	if sharedListsNamespace != nil {
		namespaces[sharedListsNamespace.ID] = sharedListsNamespace
		lists = append(lists, sharedListsNamespace.Lists...)
	}

	/////////////////
	// Saved Filters

	savedFiltersNamespace, err := getSavedFilters(s, doer)
	if err != nil {
		return nil, 0, 0, err
	}

	if savedFiltersNamespace != nil {
		namespaces[savedFiltersNamespace.ID] = savedFiltersNamespace
		lists = append(lists, savedFiltersNamespace.Lists...)
	}

	/////////////////
	// Add list details (favorite state, among other things)
	err = addListDetails(s, lists, a)
	if err != nil {
		return
	}

	/////////////////
	// Favorite lists

	favoritesNamespace, err := getFavoriteLists(s, lists, namespaceIDs, doer)
	if err != nil {
		return nil, 0, 0, err
	}

	if favoritesNamespace != nil {
		namespaces[favoritesNamespace.ID] = favoritesNamespace
	}

	//////////////////////
	// Put it all together

	for _, list := range lists {
		if list.NamespaceID == SharedListsPseudoNamespace.ID || list.NamespaceID == SavedFiltersPseudoNamespace.ID {
			// Shared lists and filtered lists are already in the namespace
			continue
		}
		namespaces[list.NamespaceID].Lists = append(namespaces[list.NamespaceID].Lists, list)
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
	// Check if we have at least a name
	if n.Title == "" {
		return ErrNamespaceNameCannotBeEmpty{NamespaceID: 0, UserID: a.GetID()}
	}
	n.ID = 0 // This would otherwise prevent the creation of new lists after one was created

	// Check if the User exists
	n.Owner, err = user.GetUserByID(s, a.GetID())
	if err != nil {
		return
	}
	n.OwnerID = n.Owner.ID

	// Insert
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

	// Delete all lists with their tasks
	lists, err := GetListsByNamespaceID(s, n.ID, &user.User{})
	if err != nil {
		return
	}

	if len(lists) == 0 {
		return events.Dispatch(&NamespaceDeletedEvent{
			Namespace: n,
			Doer:      a,
		})
	}

	var listIDs []int64
	// We need to do that for here because we need the list ids to delete two times:
	// 1) to delete the lists itself
	// 2) to delete the list tasks
	for _, l := range lists {
		listIDs = append(listIDs, l.ID)
	}

	if len(listIDs) == 0 {
		return events.Dispatch(&NamespaceDeletedEvent{
			Namespace: n,
			Doer:      a,
		})
	}

	// Delete tasks
	_, err = s.In("list_id", listIDs).Delete(&Task{})
	if err != nil {
		return
	}

	// Delete the lists
	_, err = s.In("id", listIDs).Delete(&List{})
	if err != nil {
		return
	}

	return events.Dispatch(&NamespaceDeletedEvent{
		Namespace: n,
		Doer:      a,
	})
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

	// We need to specify the cols we want to update here to be able to un-archive lists
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
