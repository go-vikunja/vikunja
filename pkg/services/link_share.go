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

package services

import (
	"errors"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"xorm.io/xorm"
)

func init() {
	// Wire dependency inversion for backward compatibility
	models.LinkShareCreateFunc = func(s *xorm.Session, share *models.LinkSharing, u *user.User) error {
		service := NewLinkShareService(s.Engine())
		return service.Create(s, share, u)
	}
	models.LinkShareGetByIDFunc = func(s *xorm.Session, id int64) (*models.LinkSharing, error) {
		service := NewLinkShareService(s.Engine())
		return service.GetByID(s, id)
	}
	models.LinkShareGetByHashFunc = func(s *xorm.Session, hash string) (*models.LinkSharing, error) {
		service := NewLinkShareService(s.Engine())
		return service.GetByHash(s, hash)
	}
	models.LinkShareUpdateFunc = func(s *xorm.Session, share *models.LinkSharing, u *user.User) error {
		service := NewLinkShareService(s.Engine())
		return service.Update(s, share, u)
	}
	models.LinkShareDeleteFunc = func(s *xorm.Session, shareID int64, u *user.User) error {
		service := NewLinkShareService(s.Engine())
		return service.Delete(s, shareID, u)
	}
}

// LinkShareService handles all business logic for link sharing functionality
type LinkShareService struct {
	DB *xorm.Engine
}

// NewLinkShareService creates a new instance of LinkShareService
func NewLinkShareService(engine *xorm.Engine) *LinkShareService {
	return &LinkShareService{
		DB: engine,
	}
}

// Create creates a new link share
func (lss *LinkShareService) Create(s *xorm.Session, share *models.LinkSharing, u *user.User) (err error) {
	// Validate permission level
	err = lss.validatePermission(share.Permission)
	if err != nil {
		return
	}

	// Check if user has permission to create link shares for this project
	canCreate, err := lss.canDoLinkShare(s, share, u)
	if err != nil {
		return err
	}
	if !canCreate {
		return &models.ErrGenericForbidden{}
	}

	// Set the user who shared this
	share.SharedByID = u.GetID()

	// Generate random hash
	hash, err := utils.CryptoRandomString(40)
	if err != nil {
		return err
	}
	share.Hash = hash

	// Handle password if provided
	if share.Password != "" {
		share.SharingType = models.SharingTypeWithPassword
		share.Password, err = user.HashPassword(share.Password)
		if err != nil {
			return
		}
	} else {
		share.SharingType = models.SharingTypeWithoutPassword
	}

	// Reset ID to ensure new record
	share.ID = 0

	// Insert into database
	_, err = s.Insert(share)
	if err != nil {
		return err
	}

	// Clear password from response and set shared by user
	share.Password = ""
	share.SharedBy, _ = user.GetFromAuth(u)

	return nil
}

// GetByID retrieves a link share by its ID
func (lss *LinkShareService) GetByID(s *xorm.Session, id int64) (*models.LinkSharing, error) {
	share := &models.LinkSharing{}
	exists, err := s.Where("id = ?", id).Get(share)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &models.ErrProjectShareDoesNotExist{ID: id}
	}

	// Always clear password from responses
	share.Password = ""
	return share, nil
}

// GetByHash retrieves a link share by its hash
func (lss *LinkShareService) GetByHash(s *xorm.Session, hash string) (*models.LinkSharing, error) {
	share := &models.LinkSharing{}
	exists, err := s.Where("hash = ?", hash).Get(share)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &models.ErrProjectShareDoesNotExist{Hash: hash}
	}

	return share, nil
}

// GetByProjectID retrieves all link shares for a project that the user can read
func (lss *LinkShareService) GetByProjectID(s *xorm.Session, projectID int64, u *user.User) ([]*models.LinkSharing, error) {
	// Check if user can read link shares for this project
	project, err := models.GetProjectSimpleByID(s, projectID)
	if err != nil {
		return nil, err
	}

	canRead, _, err := project.CanRead(s, u)
	if err != nil {
		return nil, err
	}
	if !canRead {
		return nil, &models.ErrGenericForbidden{}
	}

	var shares []*models.LinkSharing
	err = s.Where("project_id = ?", projectID).Find(&shares)
	if err != nil {
		return nil, err
	}

	// Clear passwords from all responses
	for _, share := range shares {
		share.Password = ""
	}

	return shares, nil
}

// Update updates an existing link share
func (lss *LinkShareService) Update(s *xorm.Session, share *models.LinkSharing, u *user.User) error {
	// Validate permission level
	err := lss.validatePermission(share.Permission)
	if err != nil {
		return err
	}

	// Check if user has permission to update this link share
	canUpdate, err := lss.canDoLinkShare(s, share, u)
	if err != nil {
		return err
	}
	if !canUpdate {
		return &models.ErrGenericForbidden{}
	}

	// Handle password update
	if share.Password != "" {
		share.SharingType = models.SharingTypeWithPassword
		share.Password, err = user.HashPassword(share.Password)
		if err != nil {
			return err
		}
	} else {
		share.SharingType = models.SharingTypeWithoutPassword
		share.Password = ""
	}

	// Update in database
	_, err = s.Where("id = ?", share.ID).Update(share)
	if err != nil {
		return err
	}

	// Clear password from response
	share.Password = ""
	return nil
}

// Delete removes a link share
func (lss *LinkShareService) Delete(s *xorm.Session, shareID int64, u *user.User) error {
	// Get the link share first to check permissions
	share, err := lss.GetByID(s, shareID)
	if err != nil {
		return err
	}

	// Check if user has permission to delete this link share
	canDelete, err := lss.canDoLinkShare(s, share, u)
	if err != nil {
		return err
	}
	if !canDelete {
		return &models.ErrGenericForbidden{}
	}

	// Delete from database
	_, err = s.Where("id = ?", shareID).Delete(&models.LinkSharing{})
	return err
}

// VerifyPassword checks if a password matches a link share's password
func (lss *LinkShareService) VerifyPassword(share *models.LinkSharing, password string) error {
	if share.SharingType == models.SharingTypeWithPassword {
		if password == "" {
			return &models.ErrLinkSharePasswordRequired{ShareID: share.ID}
		}

		err := bcrypt.CompareHashAndPassword([]byte(share.Password), []byte(password))
		if err != nil {
			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				return &models.ErrLinkSharePasswordInvalid{ShareID: share.ID}
			}
			return err
		}
	}

	return nil
}

// ToUser converts a LinkSharing to a pseudo-User for authentication purposes
func (lss *LinkShareService) ToUser(share *models.LinkSharing) *user.User {
	suffix := "Link Share"
	if share.Name != "" {
		suffix = " (" + suffix + ")"
	}

	username := "link-share-" + strconv.FormatInt(share.ID, 10)

	return &user.User{
		ID:       lss.getUserID(share),
		Name:     share.Name + suffix,
		Username: username,
		Created:  share.Created,
		Updated:  share.Updated,
	}
}

// getUserID returns the user ID for a link share (negative of share ID)
func (lss *LinkShareService) getUserID(share *models.LinkSharing) int64 {
	return share.ID * -1
}

// canDoLinkShare checks if a user can perform link share operations on a project
func (lss *LinkShareService) canDoLinkShare(s *xorm.Session, share *models.LinkSharing, a web.Auth) (bool, error) {
	// Don't allow creating link shares if the user itself authenticated with a link share
	if _, is := a.(*models.LinkSharing); is {
		return false, nil
	}

	project, err := models.GetProjectSimpleByID(s, share.ProjectID)
	if err != nil {
		return false, err
	}

	// Check if the user is admin when the link permission is admin
	if share.Permission == models.PermissionAdmin {
		return project.IsAdmin(s, a)
	}

	return project.CanWrite(s, a)
}

// CanRead checks if a user can read link shares for a project
func (lss *LinkShareService) CanRead(s *xorm.Session, share *models.LinkSharing, a web.Auth) (bool, int, error) {
	// Don't allow reading link shares if the user itself authenticated with a link share
	if _, is := a.(*models.LinkSharing); is {
		return false, 0, nil
	}

	project, err := models.GetProjectByShareHash(s, share.Hash)
	if err != nil {
		return false, 0, err
	}
	return project.CanRead(s, a)
}

// validatePermission validates if a permission value is valid
func (lss *LinkShareService) validatePermission(p models.Permission) error {
	if p != models.PermissionAdmin && p != models.PermissionRead && p != models.PermissionWrite {
		return &models.ErrInvalidPermission{Permission: p}
	}
	return nil
}

// CreateJWTToken creates a new JWT token from a link share
func (lss *LinkShareService) CreateJWTToken(share *models.LinkSharing) (token string, err error) {
	t := jwt.New(jwt.SigningMethodHS256)

	var ttl = time.Duration(config.ServiceJWTTTL.GetInt64())
	var exp = time.Now().Add(time.Second * ttl).Unix()

	// Set claims
	claims := t.Claims.(jwt.MapClaims)
	claims["type"] = auth.AuthTypeLinkShare
	claims["id"] = share.ID
	claims["hash"] = share.Hash
	claims["project_id"] = share.ProjectID
	claims["permission"] = share.Permission
	claims["sharedByID"] = share.SharedByID
	claims["exp"] = exp
	claims["isLocalUser"] = true // Link shares are always local

	// Generate encoded token and send it as response
	return t.SignedString([]byte(config.ServiceJWTSecret.GetString()))
}

// Authenticate validates a link share hash and optional password, returning the share if valid
func (lss *LinkShareService) Authenticate(s *xorm.Session, hash, password string) (*models.LinkSharing, error) {
	// Get the link share by hash
	share, err := lss.GetByHash(s, hash)
	if err != nil {
		return nil, err
	}

	// Verify password if required
	err = lss.VerifyPassword(share, password)
	if err != nil {
		return nil, err
	}

	return share, nil
}

// GetByIDs retrieves multiple link shares by their IDs
func (lss *LinkShareService) GetByIDs(s *xorm.Session, ids []int64) (map[int64]*models.LinkSharing, error) {
	if len(ids) == 0 {
		return make(map[int64]*models.LinkSharing), nil
	}

	var shares []*models.LinkSharing
	err := s.In("id", ids).Find(&shares)
	if err != nil {
		return nil, err
	}

	shareMap := make(map[int64]*models.LinkSharing)
	for _, share := range shares {
		// Clear password from response
		share.Password = ""
		shareMap[share.ID] = share
	}

	return shareMap, nil
}

// GetUsersOrLinkSharesFromIDs converts IDs to User objects, handling both regular users and link shares
// Negative IDs are treated as link share IDs (converted to positive)
func (lss *LinkShareService) GetUsersOrLinkSharesFromIDs(s *xorm.Session, ids []int64) (map[int64]*user.User, error) {
	users := make(map[int64]*user.User)

	var userIDs []int64
	var linkShareIDs []int64

	// Separate user IDs from link share IDs
	for _, id := range ids {
		if id < 0 {
			// Negative ID indicates a link share
			linkShareIDs = append(linkShareIDs, -id)
		} else if id > 0 {
			// Positive ID indicates a regular user
			userIDs = append(userIDs, id)
		}
	}

	// Get regular users
	if len(userIDs) > 0 {
		var userList []*user.User
		err := s.In("id", userIDs).Find(&userList)
		if err != nil {
			return nil, err
		}
		for _, u := range userList {
			users[u.ID] = u
		}
	}

	// Get link shares and convert to users
	if len(linkShareIDs) > 0 {
		linkShares, err := lss.GetByIDs(s, linkShareIDs)
		if err != nil {
			return nil, err
		}
		for _, share := range linkShares {
			pseudoUser := lss.ToUser(share)
			users[pseudoUser.ID] = pseudoUser
		}
	}

	return users, nil
}
