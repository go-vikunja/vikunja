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
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"
)

// SavedFilter represents a saved bunch of filters
type SavedFilter struct {
	// The unique numeric id of this saved filter
	ID int64 `xorm:"autoincr not null unique pk" json:"id" param:"filter"`
	// The actual filters this filter contains
	Filters *TaskCollection `xorm:"JSON not null" json:"filters"`
	// The title of the filter.
	Title string `xorm:"varchar(250) not null" json:"title" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250"`
	// The description of the filter
	Description string `xorm:"longtext null" json:"description"`
	OwnerID     int64  `xorm:"bigint not null INDEX" json:"-"`

	// The user who owns this filter
	Owner *user.User `xorm:"-" json:"owner" valid:"-"`

	// A timestamp when this filter was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this filter was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// TableName returns a better table name for saved filters
func (s *SavedFilter) TableName() string {
	return "saved_filters"
}

func (s *SavedFilter) getTaskCollection() *TaskCollection {
	// We're resetting the listID to return tasks from all lists
	s.Filters.ListID = 0
	return s.Filters
}

// Returns the saved filter ID from a list ID. Will not check if the filter actually exists.
// If the returned ID is zero, means that it is probably invalid.
func getSavedFilterIDFromListID(listID int64) (filterID int64) {
	// We get the id of the saved filter by multiplying the ListID with -1 and subtracting one
	filterID = listID*-1 - 1
	// FilterIDs from listIDs are always positive
	if filterID < 0 {
		filterID = 0
	}
	return
}

func getListIDFromSavedFilterID(filterID int64) (listID int64) {
	listID = filterID*-1 - 1
	// ListIDs from saved filters are always negative
	if listID > 0 {
		listID = 0
	}
	return
}

func getSavedFiltersForUser(auth web.Auth) (filters []*SavedFilter, err error) {
	// Link shares can't view or modify saved filters, therefore we can error out right away
	if _, is := auth.(*LinkSharing); is {
		return nil, ErrSavedFilterNotAvailableForLinkShare{LinkShareID: auth.GetID()}
	}

	err = x.Where("owner_id = ?", auth.GetID()).Find(&filters)
	return
}

// Create creates a new saved filter
// @Summary Creates a new saved filter
// @Description Creates a new saved filter
// @tags filter
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} models.SavedFilter "The Saved Filter"
// @Failure 403 {object} web.HTTPError "The user does not have access to that saved filter."
// @Failure 500 {object} models.Message "Internal error"
// @Router /filters [put]
func (s *SavedFilter) Create(auth web.Auth) error {
	s.OwnerID = auth.GetID()
	_, err := x.Insert(s)
	return err
}

func getSavedFilterSimpleByID(id int64) (s *SavedFilter, err error) {
	s = &SavedFilter{}
	exists, err := x.
		Where("id = ?", id).
		Get(s)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrSavedFilterDoesNotExist{SavedFilterID: id}
	}
	return
}

// ReadOne returns one saved filter
// @Summary Gets one saved filter
// @Description Returns a saved filter by its ID.
// @tags filter
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Filter ID"
// @Success 200 {object} models.SavedFilter "The Saved Filter"
// @Failure 403 {object} web.HTTPError "The user does not have access to that saved filter."
// @Failure 500 {object} models.Message "Internal error"
// @Router /filters/{id} [get]
func (s *SavedFilter) ReadOne() error {
	// s already contains almost the full saved filter from the rights check, we only need to add the user
	u, err := user.GetUserByID(s.OwnerID)
	s.Owner = u
	return err
}

// Update updates an existing filter
// @Summary Updates a saved filter
// @Description Updates a saved filter by its ID.
// @tags filter
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Filter ID"
// @Success 200 {object} models.SavedFilter "The Saved Filter"
// @Failure 403 {object} web.HTTPError "The user does not have access to that saved filter."
// @Failure 404 {object} web.HTTPError "The saved filter does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /filters/{id} [post]
func (s *SavedFilter) Update() error {
	_, err := x.
		Where("id = ?", s.ID).
		Cols(
			"title",
			"description",
			"filters",
		).
		Update(s)
	return err
}

// Delete removes a saved filter
// @Summary Removes a saved filter
// @Description Removes a saved filter by its ID.
// @tags filter
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Filter ID"
// @Success 200 {object} models.SavedFilter "The Saved Filter"
// @Failure 403 {object} web.HTTPError "The user does not have access to that saved filter."
// @Failure 404 {object} web.HTTPError "The saved filter does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /filters/{id} [delete]
func (s *SavedFilter) Delete() error {
	_, err := x.Where("id = ?", s.ID).Delete(s)
	return err
}
