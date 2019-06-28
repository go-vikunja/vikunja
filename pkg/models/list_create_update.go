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
	l.Owner = *doer
	l.ID = 0 // Otherwise only the first time a new list would be created

	return CreateOrUpdateList(l)
}
