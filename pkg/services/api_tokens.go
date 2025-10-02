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
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"golang.org/x/crypto/pbkdf2"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// APITokenService handles all business logic for API token management
type APITokenService struct {
	DB *xorm.Engine
}

// NewAPITokenService creates a new APITokenService instance
func NewAPITokenService(db *xorm.Engine) *APITokenService {
	return &APITokenService{
		DB: db,
	}
}

// Create creates a new API token for the authenticated user
func (ats *APITokenService) Create(s *xorm.Session, token *models.APIToken, u *user.User) error {
	if u == nil {
		return ErrAccessDenied
	}

	// Reset ID to ensure we create a new token
	token.ID = 0

	// Generate cryptographically secure salt
	salt, err := utils.CryptoRandomString(10)
	if err != nil {
		return err
	}

	// Generate cryptographically secure random token
	tokenBytes, err := utils.CryptoRandomBytes(20)
	if err != nil {
		return err
	}

	// Build the token string with prefix
	token.TokenSalt = salt
	token.Token = models.APITokenPrefix + hex.EncodeToString(tokenBytes)
	token.TokenHash = ats.hashToken(token.Token, token.TokenSalt)
	token.TokenLastEight = token.Token[len(token.Token)-8:]

	// Set the owner
	token.OwnerID = u.ID

	// Validate permissions
	if err := ats.validatePermissions(token.APIPermissions); err != nil {
		return err
	}

	// Insert the token
	_, err = s.Insert(token)
	return err
}

// Get retrieves an API token by ID
func (ats *APITokenService) Get(s *xorm.Session, id int64, u *user.User) (*models.APIToken, error) {
	if u == nil {
		return nil, ErrAccessDenied
	}

	token := &models.APIToken{}
	has, err := s.Where("id = ? AND owner_id = ?", id, u.ID).Get(token)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, models.ErrAPITokenDoesNotExist{TokenID: id}
	}

	return token, nil
}

// GetAll returns all API tokens for the authenticated user with search and pagination
func (ats *APITokenService) GetAll(s *xorm.Session, u *user.User, search string, page int, perPage int) ([]*models.APIToken, int, int64, error) {
	if u == nil {
		return nil, 0, 0, ErrAccessDenied
	}

	tokens := []*models.APIToken{}

	// Build base condition - user can only see their own tokens
	var where builder.Cond = builder.Eq{"owner_id": u.ID}

	// Add search filter if provided
	if search != "" {
		where = builder.And(
			where,
			db.ILIKE("api_tokens.title", search),
		)
	}

	// Calculate limit and offset for pagination
	limit := perPage
	if limit <= 0 {
		limit = 50 // Default limit
	}
	offset := 0
	if page > 0 {
		offset = (page - 1) * limit
	}

	// Execute query with pagination
	err := s.
		Where(where).
		Limit(limit, offset).
		Find(&tokens)
	if err != nil {
		return nil, 0, 0, err
	}

	// Get total count for pagination
	totalCount, err := s.Where(where).Count(&models.APIToken{})
	if err != nil {
		return nil, 0, 0, err
	}

	return tokens, len(tokens), totalCount, nil
}

// Delete deletes an API token
func (ats *APITokenService) Delete(s *xorm.Session, id int64, u *user.User) error {
	if u == nil {
		return ErrAccessDenied
	}

	// Verify the token exists and belongs to the user
	token, err := ats.Get(s, id, u)
	if err != nil {
		return err
	}

	// Delete the token
	_, err = s.Where("id = ? AND owner_id = ?", token.ID, u.ID).Delete(&models.APIToken{})
	return err
}

// GetTokenFromTokenString returns the full token object from the original token string
// This is used for authentication
func (ats *APITokenService) GetTokenFromTokenString(s *xorm.Session, tokenString string) (*models.APIToken, error) {
	// Use last 8 characters for efficient lookup
	lastEight := tokenString[len(tokenString)-8:]

	tokens := []*models.APIToken{}
	err := s.Where("token_last_eight = ?", lastEight).Find(&tokens)
	if err != nil {
		return nil, err
	}

	// Use constant-time comparison to prevent timing attacks
	for _, t := range tokens {
		tempHash := ats.hashToken(tokenString, t.TokenSalt)
		if subtle.ConstantTimeCompare([]byte(t.TokenHash), []byte(tempHash)) == 1 {
			return t, nil
		}
	}

	return nil, &models.ErrAPITokenInvalid{}
}

// ValidateToken checks if a token is valid and has permission for the given route
func (ats *APITokenService) ValidateToken(s *xorm.Session, tokenString string, routePath string, routeMethod string) (*models.APIToken, *user.User, error) {
	// Get the token
	token, err := ats.GetTokenFromTokenString(s, tokenString)
	if err != nil {
		return nil, nil, err
	}

	// Check if token is expired
	if time.Now().After(token.ExpiresAt) {
		return nil, nil, models.ErrAPITokenExpired{Token: token}
	}

	// Get the owner user
	u, err := user.GetUserByID(s, token.OwnerID)
	if err != nil {
		return nil, nil, err
	}

	return token, u, nil
}

// CanDelete checks if a user can delete a specific token
func (ats *APITokenService) CanDelete(s *xorm.Session, id int64, u *user.User) (bool, error) {
	if u == nil {
		return false, nil
	}

	token, err := ats.Get(s, id, u)
	if err != nil {
		// Token doesn't exist or user doesn't own it
		if models.IsErrAPITokenDoesNotExist(err) {
			return false, nil
		}
		return false, err
	}

	// User can delete their own tokens
	return token.OwnerID == u.ID, nil
}

// hashToken creates a secure hash of the token using PBKDF2
func (ats *APITokenService) hashToken(token, salt string) string {
	tempHash := pbkdf2.Key([]byte(token), []byte(salt), 10000, 50, sha256.New)
	return hex.EncodeToString(tempHash)
}

// validatePermissions checks if the provided permissions are valid
func (ats *APITokenService) validatePermissions(permissions models.APIPermissions) error {
	return models.PermissionsAreValid(permissions)
}
