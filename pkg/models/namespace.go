// Vikunja is a todo-list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/metrics"
	"code.vikunja.io/web"
	"github.com/imdario/mergo"
	"time"
)

// Namespace holds informations about a namespace
type Namespace struct {
	// The unique, numeric id of this namespace.
	ID int64 `xorm:"int(11) autoincr not null unique pk" json:"id" param:"namespace"`
	// The name of this namespace.
	Name string `xorm:"varchar(250) not null" json:"name" valid:"required,runelength(5|250)" minLength:"5" maxLength:"250"`
	// The description of the namespace
	Description string `xorm:"longtext null" json:"description"`
	OwnerID     int64  `xorm:"int(11) not null INDEX" json:"-"`

	// The user who owns this namespace
	Owner *User `xorm:"-" json:"owner" valid:"-"`

	// A unix timestamp when this namespace was created. You cannot change this value.
	Created int64 `xorm:"created not null" json:"created"`
	// A unix timestamp when this namespace was last updated. You cannot change this value.
	Updated int64 `xorm:"updated not null" json:"updated"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// PseudoNamespace is a pseudo namespace used to hold shared lists
var PseudoNamespace = Namespace{
	ID:          -1,
	Name:        "Shared Lists",
	Description: "Lists of other users shared with you via teams or directly.",
	Created:     time.Now().Unix(),
	Updated:     time.Now().Unix(),
}

// TableName makes beautiful table names
func (Namespace) TableName() string {
	return "namespaces"
}

// GetSimpleByID gets a namespace without things like the owner, it more or less only checks if it exists.
func (n *Namespace) GetSimpleByID() (err error) {
	if n.ID == 0 {
		return ErrNamespaceDoesNotExist{ID: n.ID}
	}

	// Get the namesapce with shared lists
	if n.ID == -1 {
		*n = PseudoNamespace
		return
	}

	namespaceFromDB := &Namespace{}
	exists, err := x.Where("id = ?", n.ID).Get(namespaceFromDB)
	if err != nil {
		return
	}
	if !exists {
		return ErrNamespaceDoesNotExist{ID: n.ID}
	}
	// We don't want to override the provided user struct because this would break updating, so we have to merge it
	if err := mergo.Merge(namespaceFromDB, n, mergo.WithOverride); err != nil {
		return err
	}
	*n = *namespaceFromDB

	return
}

// GetNamespaceByID returns a namespace object by its ID
func GetNamespaceByID(id int64) (namespace Namespace, err error) {
	namespace = Namespace{ID: id}
	err = namespace.GetSimpleByID()
	if err != nil {
		return
	}

	// Get the namespace Owner
	namespace.Owner, err = GetUserByID(namespace.OwnerID)
	return
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
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to that namespace."
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{id} [get]
func (n *Namespace) ReadOne() (err error) {
	// Get the namespace Owner
	n.Owner, err = GetUserByID(n.OwnerID)
	return
}

// NamespaceWithLists represents a namespace with list meta informations
type NamespaceWithLists struct {
	Namespace `xorm:"extends"`
	Lists     []*List `xorm:"-" json:"lists"`
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
// @Security JWTKeyAuth
// @Success 200 {array} models.NamespaceWithLists "The Namespaces."
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces [get]
func (n *Namespace) ReadAll(a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	if _, is := a.(*LinkSharing); is {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	doer, err := getUserWithError(a)
	if err != nil {
		return nil, 0, 0, err
	}

	all := []*NamespaceWithLists{}

	// Create our pseudo-namespace to hold the shared lists
	// We want this one at the beginning, which is why we create it here
	pseudonamespace := PseudoNamespace
	pseudonamespace.Owner = doer
	all = append(all, &NamespaceWithLists{
		pseudonamespace,
		[]*List{},
	})

	err = x.Select("namespaces.*").
		Table("namespaces").
		Join("LEFT", "team_namespaces", "namespaces.id = team_namespaces.namespace_id").
		Join("LEFT", "team_members", "team_members.team_id = team_namespaces.team_id").
		Join("LEFT", "users_namespace", "users_namespace.namespace_id = namespaces.id").
		Where("team_members.user_id = ?", doer.ID).
		Or("namespaces.owner_id = ?", doer.ID).
		Or("users_namespace.user_id = ?", doer.ID).
		GroupBy("namespaces.id").
		Limit(getLimitFromPageIndex(page, perPage)).
		Where("namespaces.name LIKE ?", "%"+search+"%").
		Find(&all)
	if err != nil {
		return all, 0, 0, err
	}

	// Get all users
	users := []*User{}
	err = x.Select("users.*").
		Table("namespaces").
		Join("LEFT", "team_namespaces", "namespaces.id = team_namespaces.namespace_id").
		Join("LEFT", "team_members", "team_members.team_id = team_namespaces.team_id").
		Join("INNER", "users", "users.id = namespaces.owner_id").
		Where("team_members.user_id = ?", doer.ID).
		Or("namespaces.owner_id = ?", doer.ID).
		GroupBy("users.id").
		Find(&users)

	if err != nil {
		return all, 0, 0, err
	}

	// Make a list of namespace ids
	var namespaceids []int64
	for _, nsp := range all {
		namespaceids = append(namespaceids, nsp.ID)
	}

	// Get all lists
	lists := []*List{}
	err = x.
		In("namespace_id", namespaceids).
		Find(&lists)
	if err != nil {
		return all, 0, 0, err
	}

	// Get all lists individually shared with our user (not via a namespace)
	individualLists := []*List{}
	err = x.Select("l.*").
		Table("list").
		Alias("l").
		Join("LEFT", []string{"team_list", "tl"}, "l.id = tl.list_id").
		Join("LEFT", []string{"team_members", "tm"}, "tm.team_id = tl.team_id").
		Join("LEFT", []string{"users_list", "ul"}, "ul.list_id = l.id").
		Where("tm.user_id = ?", doer.ID).
		Or("ul.user_id = ?", doer.ID).
		GroupBy("l.id").
		Find(&individualLists)
	if err != nil {
		return nil, 0, 0, err
	}

	// Make the namespace -1 so we now later which one it was
	// + Append it to all lists we already have
	for _, l := range individualLists {
		l.NamespaceID = -1
		lists = append(lists, l)
	}

	// Remove the pseudonamespace if we don't have any shared lists
	if len(individualLists) == 0 {
		all = append(all[:0], all[1:]...)
	}

	// More details for the lists
	err = AddListDetails(lists)
	if err != nil {
		return nil, 0, 0, err
	}

	// Put objects in our namespace list
	// TODO: Refactor this to use maps for better efficiency
	for i, n := range all {

		// Users
		for _, u := range users {
			if n.OwnerID == u.ID {
				all[i].Owner = u
				break
			}
		}

		// List infos
		for _, l := range lists {
			if n.ID == l.NamespaceID {
				all[i].Lists = append(all[i].Lists, l)
			}
		}
	}

	numberOfTotalItems, err = x.
		Table("namespaces").
		Join("LEFT", "team_namespaces", "namespaces.id = team_namespaces.namespace_id").
		Join("LEFT", "team_members", "team_members.team_id = team_namespaces.team_id").
		Join("LEFT", "users_namespace", "users_namespace.namespace_id = namespaces.id").
		Where("team_members.user_id = ?", doer.ID).
		Or("namespaces.owner_id = ?", doer.ID).
		Or("users_namespace.user_id = ?", doer.ID).
		GroupBy("namespaces.id").
		Where("namespaces.name LIKE ?", "%"+search+"%").
		Count(&NamespaceWithLists{})
	if err != nil {
		return all, 0, 0, err
	}

	return all, len(all), numberOfTotalItems, nil
}

// Create implements the creation method via the interface
// @Summary Creates a new namespace
// @Description Creates a new namespace.
// @tags namespace
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param namespace body models.Namespace true "The namespace you want to create."
// @Success 200 {object} models.Namespace "The created namespace."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid namespace object provided."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the namespace"
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces [put]
func (n *Namespace) Create(a web.Auth) (err error) {
	// Check if we have at least a name
	if n.Name == "" {
		return ErrNamespaceNameCannotBeEmpty{NamespaceID: 0, UserID: a.GetID()}
	}
	n.ID = 0 // This would otherwise prevent the creation of new lists after one was created

	// Check if the User exists
	n.Owner, err = GetUserByID(a.GetID())
	if err != nil {
		return
	}
	n.OwnerID = n.Owner.ID

	// Insert
	if _, err = x.Insert(n); err != nil {
		return err
	}

	metrics.UpdateCount(1, metrics.NamespaceCountKey)
	return
}

// Delete deletes a namespace
// @Summary Deletes a namespace
// @Description Delets a namespace
// @tags namespace
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Namespace ID"
// @Success 200 {object} models.Message "The namespace was successfully deleted."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid namespace object provided."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the namespace"
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{id} [delete]
func (n *Namespace) Delete() (err error) {

	// Check if the namespace exists
	_, err = GetNamespaceByID(n.ID)
	if err != nil {
		return
	}

	// Delete the namespace
	_, err = x.ID(n.ID).Delete(&Namespace{})
	if err != nil {
		return
	}

	// Delete all lists with their tasks
	lists, err := GetListsByNamespaceID(n.ID, &User{})
	if err != nil {
		return
	}
	var listIDs []int64
	// We need to do that for here because we need the list ids to delete two times:
	// 1) to delete the lists itself
	// 2) to delete the list tasks
	for _, l := range lists {
		listIDs = append(listIDs, l.ID)
	}

	// Delete tasks
	_, err = x.In("list_id", listIDs).Delete(&Task{})
	if err != nil {
		return
	}

	// Delete the lists
	_, err = x.In("id", listIDs).Delete(&List{})
	if err != nil {
		return
	}

	metrics.UpdateCount(-1, metrics.NamespaceCountKey)

	return
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
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid namespace object provided."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the namespace"
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespace/{id} [post]
func (n *Namespace) Update() (err error) {
	// Check if we have at least a name
	if n.Name == "" {
		return ErrNamespaceNameCannotBeEmpty{NamespaceID: n.ID}
	}

	// Check if the namespace exists
	currentNamespace, err := GetNamespaceByID(n.ID)
	if err != nil {
		return
	}

	// Check if the (new) owner exists
	n.OwnerID = n.Owner.ID
	if currentNamespace.OwnerID != n.OwnerID {
		n.Owner, err = GetUserByID(n.OwnerID)
		if err != nil {
			return
		}
	}

	// Do the actual update
	_, err = x.ID(currentNamespace.ID).Update(n)
	return
}
