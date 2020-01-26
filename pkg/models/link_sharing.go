// Copyright 2018-2020 Vikunja and contriubtors. All rights reserved.
//
// This file is part of Vikunja.
//
// Vikunja is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Vikunja is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Vikunja.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/web"
	"github.com/dgrijalva/jwt-go"
)

// SharingType holds the sharing type
type SharingType int

// These consts represent all valid link sharing types
const (
	SharingTypeUnknown SharingType = iota
	SharingTypeWithoutPassword
	SharingTypeWithPassword
)

// LinkSharing represents a shared list
type LinkSharing struct {
	// The ID of the shared thing
	ID int64 `xorm:"int(11) autoincr not null unique pk" json:"id" param:"share"`
	// The public id to get this shared list
	Hash string `xorm:"varchar(40) not null unique" json:"hash" param:"hash"`
	// The ID of the shared list
	ListID int64 `xorm:"int(11) not null" json:"-" param:"list"`
	// The right this list is shared with. 0 = Read only, 1 = Read & Write, 2 = Admin. See the docs for more details.
	Right Right `xorm:"int(11) INDEX not null default 0" json:"right" valid:"length(0|2)" maximum:"2" default:"0"`

	// The kind of this link. 0 = undefined, 1 = without password, 2 = with password (currently not implemented).
	SharingType SharingType `xorm:"int(11) INDEX not null default 0" json:"sharing_type" valid:"length(0|2)" maximum:"2" default:"0"`

	// The user who shared this list
	SharedBy   *user.User `xorm:"-" json:"shared_by"`
	SharedByID int64      `xorm:"int(11) INDEX not null" json:"-"`

	// A unix timestamp when this list was shared. You cannot change this value.
	Created int64 `xorm:"created not null" json:"created"`
	// A unix timestamp when this share was last updated. You cannot change this value.
	Updated int64 `xorm:"updated not null" json:"updated"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// TableName holds the table name
func (LinkSharing) TableName() string {
	return "link_sharing"
}

// GetID returns the ID of the links sharing object
func (share *LinkSharing) GetID() int64 {
	return share.ID
}

// GetLinkShareFromClaims builds a link sharing object from jwt claims
func GetLinkShareFromClaims(claims jwt.MapClaims) (share *LinkSharing, err error) {
	share = &LinkSharing{}
	share.ID = int64(claims["id"].(float64))
	share.Hash = claims["hash"].(string)
	share.ListID = int64(claims["listID"].(float64))
	share.Right = Right(claims["right"].(float64))
	share.SharedByID = int64(claims["sharedByID"].(float64))
	return
}

// Create creates a new link share for a given list
// @Summary Share a list via link
// @Description Share a list via link. The user needs to have write-access to the list to be able do this.
// @tags sharing
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param list path int true "List ID"
// @Param label body models.LinkSharing true "The new link share object"
// @Success 200 {object} models.LinkSharing "The created link share object."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid link share object provided."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "Not allowed to add the list share."
// @Failure 404 {object} code.vikunja.io/web.HTTPError "The list does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{list}/shares [put]
func (share *LinkSharing) Create(a web.Auth) (err error) {
	share.SharedByID = a.GetID()
	share.Hash = utils.MakeRandomString(40)
	_, err = x.Insert(share)
	share.SharedBy, _ = a.(*user.User)
	return
}

// ReadOne returns one share
// @Summary Get one link shares for a list
// @Description Returns one link share by its ID.
// @tags sharing
// @Accept json
// @Produce json
// @Param list path int true "List ID"
// @Param share path int true "Share ID"
// @Security JWTKeyAuth
// @Success 200 {object} models.LinkSharing "The share links"
// @Failure 403 {object} code.vikunja.io/web.HTTPError "No access to the list"
// @Failure 404 {object} code.vikunja.io/web.HTTPError "Share Link not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{list}/shares/{share} [get]
func (share *LinkSharing) ReadOne() (err error) {
	exists, err := x.Where("id = ?", share.ID).Get(share)
	if err != nil {
		return err
	}
	if !exists {
		return ErrListShareDoesNotExist{ID: share.ID, Hash: share.Hash}
	}
	return
}

// ReadAll returns all shares for a given list
// @Summary Get all link shares for a list
// @Description Returns all link shares which exist for a given list
// @tags sharing
// @Accept json
// @Produce json
// @Param list path int true "List ID"
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search shares by hash."
// @Security JWTKeyAuth
// @Success 200 {array} models.LinkSharing "The share links"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{list}/shares [get]
func (share *LinkSharing) ReadAll(a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	list := &List{ID: share.ListID}
	can, err := list.CanRead(a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !can {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	var shares []*LinkSharing
	err = x.
		Where("list_id = ? AND hash LIKE ?", share.ListID, "%"+search+"%").
		Limit(getLimitFromPageIndex(page, perPage)).
		Find(&shares)
	if err != nil {
		return nil, 0, 0, err
	}

	// Find all users and add them
	var userIDs []int64
	for _, s := range shares {
		userIDs = append(userIDs, s.SharedByID)
	}

	users := make(map[int64]*user.User)
	err = x.In("id", userIDs).Find(&users)
	if err != nil {
		return nil, 0, 0, err
	}

	for _, s := range shares {
		s.SharedBy = users[s.SharedByID]
	}

	// Total count
	totalItems, err = x.
		Where("list_id = ? AND hash LIKE ?", share.ListID, "%"+search+"%").
		Count(&LinkSharing{})
	if err != nil {
		return nil, 0, 0, err
	}

	return shares, len(shares), totalItems, err
}

// Delete removes a link share
// @Summary Remove a link share
// @Description Remove a link share. The user needs to have write-access to the list to be able do this.
// @tags sharing
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param list path int true "List ID"
// @Param share path int true "Share Link ID"
// @Success 200 {object} models.Message "The link was successfully removed."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "Not allowed to remove the link."
// @Failure 404 {object} code.vikunja.io/web.HTTPError "Share Link not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{list}/shares/{share} [delete]
func (share *LinkSharing) Delete() (err error) {
	_, err = x.Where("id = ?", share.ID).Delete(share)
	return
}

// GetLinkShareByHash returns a link share by hash
func GetLinkShareByHash(hash string) (share *LinkSharing, err error) {
	share = &LinkSharing{}
	has, err := x.Where("hash = ?", hash).Get(share)
	if err != nil {
		return
	}
	if !has {
		return share, ErrListShareDoesNotExist{Hash: hash}
	}
	return
}

// GetListByShareHash returns a link share by its hash
func GetListByShareHash(hash string) (list *List, err error) {
	share, err := GetLinkShareByHash(hash)
	if err != nil {
		return
	}

	list = &List{ID: share.ListID}
	err = list.GetSimpleByID()
	return
}
