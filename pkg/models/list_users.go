// Vikunja is a to-do list application to facilitate your life.
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
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"
)

// ListUser represents a list <-> user relation
type ListUser struct {
	// The unique, numeric id of this list <-> user relation.
	ID int64 `xorm:"int(11) autoincr not null unique pk" json:"id" param:"namespace"`
	// The username.
	Username string `xorm:"-" json:"userID" param:"user"`
	// Used internally to reference the user
	UserID int64 `xorm:"int(11) not null INDEX" json:"-"`
	// The list id.
	ListID int64 `xorm:"int(11) not null INDEX" json:"-" param:"list"`
	// The right this user has. 0 = Read only, 1 = Read & Write, 2 = Admin. See the docs for more details.
	Right Right `xorm:"int(11) INDEX not null default 0" json:"right" valid:"length(0|2)" maximum:"2" default:"0"`

	// A unix timestamp when this relation was created. You cannot change this value.
	Created int64 `xorm:"created not null" json:"created"`
	// A unix timestamp when this relation was last updated. You cannot change this value.
	Updated int64 `xorm:"updated not null" json:"updated"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// TableName is the table name for ListUser
func (ListUser) TableName() string {
	return "users_list"
}

// UserWithRight represents a user in combination with the right it can have on a list/namespace
type UserWithRight struct {
	user.User `xorm:"extends"`
	Right     Right `json:"right"`
}

// Create creates a new list <-> user relation
// @Summary Add a user to a list
// @Description Gives a user access to a list.
// @tags sharing
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "List ID"
// @Param list body models.ListUser true "The user you want to add to the list."
// @Success 200 {object} models.ListUser "The created user<->list relation."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid user list object provided."
// @Failure 404 {object} code.vikunja.io/web.HTTPError "The user does not exist."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id}/users [put]
func (lu *ListUser) Create(a web.Auth) (err error) {

	// Check if the right is valid
	if err := lu.Right.isValid(); err != nil {
		return err
	}

	// Check if the list exists
	l := &List{ID: lu.ListID}
	if err = l.GetSimpleByID(); err != nil {
		return
	}

	// Check if the user exists
	user, err := user.GetUserByUsername(lu.Username)
	if err != nil {
		return err
	}
	lu.UserID = user.ID

	// Check if the user already has access or is owner of that list
	// We explicitly DONT check for teams here
	if l.OwnerID == lu.UserID {
		return ErrUserAlreadyHasAccess{UserID: lu.UserID, ListID: lu.ListID}
	}

	exist, err := x.Where("list_id = ? AND user_id = ?", lu.ListID, lu.UserID).Get(&ListUser{})
	if err != nil {
		return
	}
	if exist {
		return ErrUserAlreadyHasAccess{UserID: lu.UserID, ListID: lu.ListID}
	}

	// Insert user <-> list relation
	_, err = x.Insert(lu)
	if err != nil {
		return err
	}

	err = updateListLastUpdated(l)
	return
}

// Delete deletes a list <-> user relation
// @Summary Delete a user from a list
// @Description Delets a user from a list. The user won't have access to the list anymore.
// @tags sharing
// @Produce json
// @Security JWTKeyAuth
// @Param listID path int true "List ID"
// @Param userID path int true "User ID"
// @Success 200 {object} models.Message "The user was successfully removed from the list."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the list"
// @Failure 404 {object} code.vikunja.io/web.HTTPError "user or list does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{listID}/users/{userID} [delete]
func (lu *ListUser) Delete() (err error) {

	// Check if the user exists
	user, err := user.GetUserByUsername(lu.Username)
	if err != nil {
		return
	}
	lu.UserID = user.ID

	// Check if the user has access to the list
	has, err := x.Where("user_id = ? AND list_id = ?", lu.UserID, lu.ListID).
		Get(&ListUser{})
	if err != nil {
		return
	}
	if !has {
		return ErrUserDoesNotHaveAccessToList{ListID: lu.ListID, UserID: lu.UserID}
	}

	_, err = x.Where("user_id = ? AND list_id = ?", lu.UserID, lu.ListID).
		Delete(&ListUser{})
	if err != nil {
		return err
	}

	err = updateListLastUpdated(&List{ID: lu.ListID})
	return
}

// ReadAll gets all users who have access to a list
// @Summary Get users on a list
// @Description Returns a list with all users which have access on a given list.
// @tags sharing
// @Accept json
// @Produce json
// @Param id path int true "List ID"
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search users by its name."
// @Security JWTKeyAuth
// @Success 200 {array} models.UserWithRight "The users with the right they have."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "No right to see the list."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id}/users [get]
func (lu *ListUser) ReadAll(a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	// Check if the user has access to the list
	l := &List{ID: lu.ListID}
	canRead, err := l.CanRead(a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !canRead {
		return nil, 0, 0, ErrNeedToHaveListReadAccess{UserID: a.GetID(), ListID: lu.ListID}
	}

	// Get all users
	all := []*UserWithRight{}
	err = x.
		Join("INNER", "users_list", "user_id = users.id").
		Where("users_list.list_id = ?", lu.ListID).
		Limit(getLimitFromPageIndex(page, perPage)).
		Where("users.username LIKE ?", "%"+search+"%").
		Find(&all)
	if err != nil {
		return nil, 0, 0, err
	}

	// Obfuscate all user emails
	for _, u := range all {
		u.Email = ""
	}

	numberOfTotalItems, err = x.
		Join("INNER", "users_list", "user_id = users.id").
		Where("users_list.list_id = ?", lu.ListID).
		Where("users.username LIKE ?", "%"+search+"%").
		Count(&UserWithRight{})

	return all, len(all), numberOfTotalItems, err
}

// Update updates a user <-> list relation
// @Summary Update a user <-> list relation
// @Description Update a user <-> list relation. Mostly used to update the right that user has.
// @tags sharing
// @Accept json
// @Produce json
// @Param listID path int true "List ID"
// @Param userID path int true "User ID"
// @Param list body models.ListUser true "The user you want to update."
// @Security JWTKeyAuth
// @Success 200 {object} models.ListUser "The updated user <-> list relation."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have admin-access to the list"
// @Failure 404 {object} code.vikunja.io/web.HTTPError "User or list does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{listID}/users/{userID} [post]
func (lu *ListUser) Update() (err error) {

	// Check if the right is valid
	if err := lu.Right.isValid(); err != nil {
		return err
	}

	// Check if the user exists
	user, err := user.GetUserByUsername(lu.Username)
	if err != nil {
		return err
	}
	lu.UserID = user.ID

	_, err = x.
		Where("list_id = ? AND user_id = ?", lu.ListID, lu.UserID).
		Cols("right").
		Update(lu)
	if err != nil {
		return err
	}

	err = updateListLastUpdated(&List{ID: lu.ListID})
	return
}
