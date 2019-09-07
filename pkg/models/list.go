//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/metrics"
	"code.vikunja.io/web"
)

// List represents a list of tasks
type List struct {
	// The unique, numeric id of this list.
	ID int64 `xorm:"int(11) autoincr not null unique pk" json:"id" param:"list"`
	// The title of the list. You'll see this in the namespace overview.
	Title string `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(3|250)" minLength:"3" maxLength:"250"`
	// The description of the list.
	Description string `xorm:"longtext null" json:"description"`
	OwnerID     int64  `xorm:"int(11) INDEX not null" json:"-"`
	NamespaceID int64  `xorm:"int(11) INDEX not null" json:"-" param:"namespace"`

	// The user who created this list.
	Owner *User `xorm:"-" json:"owner" valid:"-"`
	// An array of tasks which belong to the list.
	Tasks []*Task `xorm:"-" json:"tasks"`

	// A unix timestamp when this list was created. You cannot change this value.
	Created int64 `xorm:"created not null" json:"created"`
	// A unix timestamp when this list was last updated. You cannot change this value.
	Updated int64 `xorm:"updated not null" json:"updated"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// GetListsByNamespaceID gets all lists in a namespace
func GetListsByNamespaceID(nID int64, doer *User) (lists []*List, err error) {
	if nID == -1 {
		err = x.Select("l.*").
			Table("list").
			Alias("l").
			Join("LEFT", []string{"team_list", "tl"}, "l.id = tl.list_id").
			Join("LEFT", []string{"team_members", "tm"}, "tm.team_id = tl.team_id").
			Join("LEFT", []string{"users_list", "ul"}, "ul.list_id = l.id").
			Where("tm.user_id = ?", doer.ID).
			Or("ul.user_id = ?", doer.ID).
			GroupBy("l.id").
			Find(&lists)
	} else {
		err = x.Where("namespace_id = ?", nID).Find(&lists)
	}
	if err != nil {
		return nil, err
	}

	// get more list details
	err = AddListDetails(lists)
	return lists, err
}

// ReadAll gets all lists a user has access to
// @Summary Get all lists a user has access to
// @Description Returns all lists a user has access to.
// @tags list
// @Accept json
// @Produce json
// @Param p query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param s query string false "Search lists by title."
// @Security JWTKeyAuth
// @Success 200 {array} models.List "The lists"
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists [get]
func (l *List) ReadAll(search string, a web.Auth, page int) (interface{}, error) {
	// Check if we're dealing with a share auth
	shareAuth, ok := a.(*LinkSharing)
	if ok {
		list := &List{ID: shareAuth.ListID}
		err := list.GetSimpleByID()
		if err != nil {
			return nil, err
		}
		lists := []*List{list}
		err = AddListDetails(lists)
		return lists, err
	}

	lists, err := getRawListsForUser(search, &User{ID: a.GetID()}, page)
	if err != nil {
		return nil, err
	}

	// Add more list details
	err = AddListDetails(lists)
	return lists, err
}

// ReadOne gets one list by its ID
// @Summary Gets one list
// @Description Returns a list by its ID.
// @tags list
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "List ID"
// @Success 200 {object} models.List "The list"
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id} [get]
func (l *List) ReadOne() (err error) {
	// Get list tasks
	l.Tasks, err = GetTasksByListID(l.ID)
	if err != nil {
		return err
	}

	// Get list owner
	l.Owner, err = GetUserByID(l.OwnerID)
	return
}

// GetSimpleByID gets a list with only the basic items, aka no tasks or user objects. Returns an error if the list does not exist.
func (l *List) GetSimpleByID() (err error) {
	if l.ID < 1 {
		return ErrListDoesNotExist{ID: l.ID}
	}

	// We need to re-init our list object, because otherwise xorm creates a "where for every item in that list object,
	// leading to not finding anything if the id is good, but for example the title is different.
	id := l.ID
	*l = List{}
	exists, err := x.Where("id = ?", id).Get(l)
	if err != nil {
		return
	}

	if !exists {
		return ErrListDoesNotExist{ID: l.ID}
	}

	return
}

// GetListSimplByTaskID gets a list by a task id
func GetListSimplByTaskID(taskID int64) (l *List, err error) {
	// We need to re-init our list object, because otherwise xorm creates a "where for every item in that list object,
	// leading to not finding anything if the id is good, but for example the title is different.
	var list List
	exists, err := x.
		Select("list.*").
		Table(List{}).
		Join("INNER", "tasks", "list.id = tasks.list_id").
		Where("tasks.id = ?", taskID).
		Get(&list)
	if err != nil {
		return
	}

	if !exists {
		return &List{}, ErrListDoesNotExist{ID: l.ID}
	}

	return &list, nil
}

// Gets the lists only, without any tasks or so
func getRawListsForUser(search string, u *User, page int) (lists []*List, err error) {
	fullUser, err := GetUserByID(u.ID)
	if err != nil {
		return lists, err
	}

	// Gets all Lists where the user is either owner or in a team which has access to the list
	// Or in a team which has namespace read access
	err = x.Select("l.*").
		Table("list").
		Alias("l").
		Join("INNER", []string{"namespaces", "n"}, "l.namespace_id = n.id").
		Join("LEFT", []string{"team_namespaces", "tn"}, "tn.namespace_id = n.id").
		Join("LEFT", []string{"team_members", "tm"}, "tm.team_id = tn.team_id").
		Join("LEFT", []string{"team_list", "tl"}, "l.id = tl.list_id").
		Join("LEFT", []string{"team_members", "tm2"}, "tm2.team_id = tl.team_id").
		Join("LEFT", []string{"users_list", "ul"}, "ul.list_id = l.id").
		Join("LEFT", []string{"users_namespace", "un"}, "un.namespace_id = l.namespace_id").
		Where("tm.user_id = ?", fullUser.ID).
		Or("tm2.user_id = ?", fullUser.ID).
		Or("l.owner_id = ?", fullUser.ID).
		Or("ul.user_id = ?", fullUser.ID).
		Or("un.user_id = ?", fullUser.ID).
		GroupBy("l.id").
		Limit(getLimitFromPageIndex(page)).
		Where("l.title LIKE ?", "%"+search+"%").
		Find(&lists)

	return lists, err
}

// AddListDetails adds owner user objects and list tasks to all lists in the slice
func AddListDetails(lists []*List) (err error) {
	var listIDs []int64
	var ownerIDs []int64
	for _, l := range lists {
		listIDs = append(listIDs, l.ID)
		ownerIDs = append(ownerIDs, l.OwnerID)
	}

	// Get all tasks
	ts := []*Task{}
	err = x.In("list_id", listIDs).Find(&ts)
	if err != nil {
		return
	}

	// Get all list owners
	owners := []*User{}
	err = x.In("id", ownerIDs).Find(&owners)
	if err != nil {
		return
	}

	// Build it all into the lists slice
	for in, list := range lists {
		// Owner
		for _, owner := range owners {
			if list.OwnerID == owner.ID {
				lists[in].Owner = owner
				break
			}
		}

		// Tasks
		for _, task := range ts {
			if task.ListID == list.ID {
				lists[in].Tasks = append(lists[in].Tasks, task)
			}
		}
	}

	return
}

// CreateOrUpdateList updates a list or creates it if it doesn't exist
func CreateOrUpdateList(list *List) (err error) {

	// Check if the namespace exists
	if list.NamespaceID != 0 {
		_, err = GetNamespaceByID(list.NamespaceID)
		if err != nil {
			return err
		}
	}

	if list.ID == 0 {
		_, err = x.Insert(list)
		metrics.UpdateCount(1, metrics.ListCountKey)
	} else {
		_, err = x.ID(list.ID).Update(list)
	}

	if err != nil {
		return
	}

	err = list.GetSimpleByID()
	if err != nil {
		return
	}

	err = list.ReadOne()
	return

}

// Update implements the update method of CRUDable
// @Summary Updates a list
// @Description Updates a list. This does not include adding a task (see below).
// @tags list
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "List ID"
// @Param list body models.List true "The list with updated values you want to update."
// @Success 200 {object} models.List "The updated list."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid list object provided."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id} [post]
func (l *List) Update() (err error) {
	return CreateOrUpdateList(l)
}

func updateListLastUpdated(list *List) error {
	_, err := x.ID(list.ID).Cols("updated").Update(list)
	return err
}

func updateListByTaskID(taskID int64) (err error) {
	// need to get the task to update the list last updated timestamp
	task, err := GetTaskByIDSimple(taskID)
	if err != nil {
		return err
	}

	return updateListLastUpdated(&List{ID: task.ListID})
}

// Create implements the create method of CRUDable
// @Summary Creates a new list
// @Description Creates a new list in a given namespace. The user needs write-access to the namespace.
// @tags list
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param namespaceID path int true "Namespace ID"
// @Param list body models.List true "The list you want to create."
// @Success 200 {object} models.List "The created list."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid list object provided."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{namespaceID}/lists [put]
func (l *List) Create(a web.Auth) (err error) {
	doer, err := getUserWithError(a)
	if err != nil {
		return err
	}

	l.OwnerID = doer.ID
	l.Owner = doer
	l.ID = 0 // Otherwise only the first time a new list would be created

	return CreateOrUpdateList(l)
}

// Delete implements the delete method of CRUDable
// @Summary Deletes a list
// @Description Delets a list
// @tags list
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "List ID"
// @Success 200 {object} models.Message "The list was successfully deleted."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid list object provided."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id} [delete]
func (l *List) Delete() (err error) {

	// Delete the list
	_, err = x.ID(l.ID).Delete(&List{})
	if err != nil {
		return
	}
	metrics.UpdateCount(-1, metrics.ListCountKey)

	// Delete all todotasks on that list
	_, err = x.Where("list_id = ?", l.ID).Delete(&Task{})
	return
}
