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
	"errors"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// SharingType holds the sharing type
type SharingType int

// These consts represent all valid link sharing types
const (
	SharingTypeUnknown SharingType = iota
	SharingTypeWithoutPassword
	SharingTypeWithPassword
)

// LinkSharing represents a shared project
type LinkSharing struct {
	// The ID of the shared thing
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"share"`
	// The public id to get this shared project
	Hash string `xorm:"varchar(40) not null unique" json:"hash" param:"hash"`
	// The name of this link share. All actions someone takes while being authenticated with that link will appear with that name.
	Name string `xorm:"text null" json:"name"`
	// The ID of the shared project
	ProjectID int64 `xorm:"bigint not null" json:"-" param:"project"`
	// The permission this project is shared with. 0 = Read only, 1 = Read & Write, 2 = Admin. See the docs for more details.
	Permission Permission `xorm:"bigint INDEX not null default 0" json:"permission" valid:"length(0|2)" maximum:"2" default:"0"`

	// The kind of this link. 0 = undefined, 1 = without password, 2 = with password.
	SharingType SharingType `xorm:"bigint INDEX not null default 0" json:"sharing_type" valid:"length(0|2)" maximum:"2" default:"0"`

	// The password of this link share. You can only set it, not retrieve it after the link share has been created.
	Password string `xorm:"text null" json:"password"`

	// The user who shared this project
	SharedBy   *user.User `xorm:"-" json:"shared_by"`
	SharedByID int64      `xorm:"bigint INDEX not null" json:"-"`

	// A timestamp when this project was shared. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this share was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// TableName holds the table name
func (*LinkSharing) TableName() string {
	return "link_shares"
}

// GetID returns the ID of the links sharing object
func (share *LinkSharing) GetID() int64 {
	return share.ID
}

// GetLinkShareFromClaims builds a link sharing object from jwt claims
func GetLinkShareFromClaims(claims jwt.MapClaims) (share *LinkSharing, err error) {
	projectID, is := claims["project_id"].(float64)
	if !is {
		return nil, &ErrLinkShareTokenInvalid{}
	}

	permissionFloat, is := claims["permission"].(float64)
	if !is {
		return nil, &ErrLinkShareTokenInvalid{}
	}

	share = &LinkSharing{}
	share.ID = int64(claims["id"].(float64))
	share.Hash = claims["hash"].(string)
	share.ProjectID = int64(projectID)
	share.Permission = Permission(permissionFloat)
	share.SharedByID = int64(claims["sharedByID"].(float64))
	return
}

func (share *LinkSharing) getUserID() int64 {
	return share.ID * -1
}

func (share *LinkSharing) toUser() *user.User {
	suffix := "Link Share"
	if share.Name != "" {
		suffix = " (" + suffix + ")"
	}

	username := "link-share-" + strconv.FormatInt(share.ID, 10)

	return &user.User{
		ID:       share.getUserID(),
		Name:     share.Name + suffix,
		Username: username,
		Created:  share.Created,
		Updated:  share.Updated,
	}
}

// Create creates a new link share for a given project
// @Summary Share a project via link
// @Description Share a project via link. The user needs to have write-access to the project to be able do this.
// @tags sharing
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param label body models.LinkSharing true "The new link share object"
// @Success 201 {object} models.LinkSharing "The created link share object."
// @Failure 400 {object} web.HTTPError "Invalid link share object provided."
// @Failure 403 {object} web.HTTPError "Not allowed to add the project share."
// @Failure 404 {object} web.HTTPError "The project does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/shares [put]
func (share *LinkSharing) Create(s *xorm.Session, a web.Auth) (err error) {

	err = share.Permission.isValid()
	if err != nil {
		return
	}

	share.SharedByID = a.GetID()
	hash, err := utils.CryptoRandomString(40)
	if err != nil {
		return err
	}
	share.Hash = hash

	if share.Password != "" {
		share.SharingType = SharingTypeWithPassword
		share.Password, err = user.HashPassword(share.Password)
		if err != nil {
			return
		}
	} else {
		share.SharingType = SharingTypeWithoutPassword
	}

	share.ID = 0
	_, err = s.Insert(share)
	share.Password = ""
	share.SharedBy, _ = user.GetFromAuth(a)
	return
}

// ReadOne returns one share
// @Summary Get one link shares for a project
// @Description Returns one link share by its ID.
// @tags sharing
// @Accept json
// @Produce json
// @Param project path int true "Project ID"
// @Param share path int true "Share ID"
// @Security JWTKeyAuth
// @Success 200 {object} models.LinkSharing "The share links"
// @Failure 403 {object} web.HTTPError "No access to the project"
// @Failure 404 {object} web.HTTPError "Share Link not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/shares/{share} [get]
func (share *LinkSharing) ReadOne(s *xorm.Session, _ web.Auth) (err error) {
	exists, err := s.Where("id = ?", share.ID).Get(share)
	if err != nil {
		return err
	}
	if !exists {
		return ErrProjectShareDoesNotExist{ID: share.ID, Hash: share.Hash}
	}
	share.Password = ""
	return
}

// ReadAll returns all shares for a given project
// @Summary Get all link shares for a project
// @Description Returns all link shares which exist for a given project
// @tags sharing
// @Accept json
// @Produce json
// @Param project path int true "Project ID"
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search shares by hash."
// @Security JWTKeyAuth
// @Success 200 {array} models.LinkSharing "The share links"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/shares [get]
func (share *LinkSharing) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	project := &Project{ID: share.ProjectID}
	can, _, err := project.CanRead(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !can {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	limit, start := getLimitFromPageIndex(page, perPage)

	var shares []*LinkSharing
	var where []builder.Cond
	where = append(where, builder.Eq{"project_id": share.ProjectID})

	if search != "" {
		where = append(where, db.ILIKE("name", search))
	}

	query := s.Where(builder.And(where...))

	if limit > 0 {
		query = query.Limit(limit, start)
	}
	err = query.Find(&shares)
	if err != nil {
		return nil, 0, 0, err
	}

	// Find all users and add them
	var userIDs []int64
	for _, s := range shares {
		userIDs = append(userIDs, s.SharedByID)
	}

	users := make(map[int64]*user.User)
	if len(userIDs) > 0 {
		err = s.In("id", userIDs).Find(&users)
		if err != nil {
			return nil, 0, 0, err
		}
	}

	for _, s := range shares {
		if sharedBy, has := users[s.SharedByID]; has {
			s.SharedBy = sharedBy
		}
		s.Password = ""
	}

	// Total count
	totalItems, err = s.
		Where("project_id = ? AND hash LIKE ?", share.ProjectID, "%"+search+"%").
		Count(&LinkSharing{})
	if err != nil {
		return nil, 0, 0, err
	}

	return shares, len(shares), totalItems, err
}

// Delete removes a link share
// @Summary Remove a link share
// @Description Remove a link share. The user needs to have write-access to the project to be able do this.
// @tags sharing
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param project path int true "Project ID"
// @Param share path int true "Share Link ID"
// @Success 200 {object} models.Message "The link was successfully removed."
// @Failure 403 {object} web.HTTPError "Not allowed to remove the link."
// @Failure 404 {object} web.HTTPError "Share Link not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{project}/shares/{share} [delete]
func (share *LinkSharing) Delete(s *xorm.Session, _ web.Auth) (err error) {
	_, err = s.Where("id = ?", share.ID).Delete(share)
	return
}

// GetLinkShareByHash returns a link share by hash
func GetLinkShareByHash(s *xorm.Session, hash string) (share *LinkSharing, err error) {
	share = &LinkSharing{}
	has, err := s.Where("hash = ?", hash).Get(share)
	if err != nil {
		return
	}
	if !has {
		return share, ErrProjectShareDoesNotExist{Hash: hash}
	}
	return
}

// GetProjectByShareHash returns a link share by its hash
func GetProjectByShareHash(s *xorm.Session, hash string) (project *Project, err error) {
	share, err := GetLinkShareByHash(s, hash)
	if err != nil {
		return
	}

	project, err = GetProjectSimpleByID(s, share.ProjectID)
	return
}

// GetLinkShareByID returns a link share by its id.
func GetLinkShareByID(s *xorm.Session, id int64) (share *LinkSharing, err error) {
	share = &LinkSharing{}
	has, err := s.Where("id = ?", id).Get(share)
	if err != nil {
		return
	}
	if !has {
		return share, ErrProjectShareDoesNotExist{ID: id}
	}
	return
}

// GetLinkSharesByIDs returns all link shares from a slice of ids
func GetLinkSharesByIDs(s *xorm.Session, ids []int64) (shares map[int64]*LinkSharing, err error) {
	shares = make(map[int64]*LinkSharing)
	err = s.In("id", ids).Find(&shares)
	return
}

// VerifyLinkSharePassword checks if a password of a link share matches a provided one.
func VerifyLinkSharePassword(share *LinkSharing, password string) (err error) {
	if password == "" {
		return &ErrLinkSharePasswordRequired{ShareID: share.ID}
	}

	err = bcrypt.CompareHashAndPassword([]byte(share.Password), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return &ErrLinkSharePasswordInvalid{ShareID: share.ID}
		}
		return err
	}

	return nil
}
