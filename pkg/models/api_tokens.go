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
	"crypto/sha256"
	"encoding/hex"
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"golang.org/x/crypto/pbkdf2"
	"xorm.io/xorm"
)

// APITokenServiceProvider is a function type that returns an API token service instance
// This is used to avoid import cycles between models and services packages
type APITokenServiceProvider func() interface {
	Create(s *xorm.Session, token *APIToken, u *user.User) error
	GetAll(s *xorm.Session, u *user.User, search string, page int, perPage int) ([]*APIToken, int, int64, error)
	GetByID(s *xorm.Session, id int64) (*APIToken, error)
	Delete(s *xorm.Session, id int64, u *user.User) error
}

// apiTokenServiceProvider is the registered service provider function
var apiTokenServiceProvider APITokenServiceProvider

// RegisterAPITokenService registers a service provider for API token operations
// This should be called during application initialization by the services package
func RegisterAPITokenService(provider APITokenServiceProvider) {
	apiTokenServiceProvider = provider
}

// getAPITokenService returns the registered API token service instance
func getAPITokenService() interface {
	Create(s *xorm.Session, token *APIToken, u *user.User) error
	GetAll(s *xorm.Session, u *user.User, search string, page int, perPage int) ([]*APIToken, int, int64, error)
	GetByID(s *xorm.Session, id int64) (*APIToken, error)
	Delete(s *xorm.Session, id int64, u *user.User) error
} {
	if apiTokenServiceProvider == nil {
		panic("APITokenService not registered - did you forget to call services.InitializeDependencies()?")
	}
	return apiTokenServiceProvider()
}

type APIPermissions map[string][]string

type APIToken struct {
	// The unique, numeric id of this api key.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"token"`

	// A human-readable name for this token
	Title string `xorm:"not null" json:"title" valid:"required"`
	// The actual api key. Only visible after creation.
	Token          string `xorm:"-" json:"token,omitempty"`
	TokenSalt      string `xorm:"not null" json:"-"`
	TokenHash      string `xorm:"not null unique" json:"-"`
	TokenLastEight string `xorm:"not null index varchar(8)" json:"-"`
	// The permissions this token has. Possible values are available via the /routes endpoint and consist of the keys of the list from that endpoint. For example, if the token should be able to read all tasks as well as update existing tasks, you should add `{"tasks":["read_all","update"]}`.
	APIPermissions APIPermissions `xorm:"json not null permissions" json:"permissions" valid:"required"`
	// The date when this key expires.
	ExpiresAt time.Time `xorm:"not null" json:"expires_at" valid:"required"`

	// A timestamp when this api key was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`

	OwnerID int64 `xorm:"bigint not null" json:"-"`

	web.Permissions `xorm:"-" json:"-"`
	web.CRUDable    `xorm:"-" json:"-"`
}

const APITokenPrefix = `tk_`

func (*APIToken) TableName() string {
	return "api_tokens"
}

// Create creates a new token
// @Deprecated: This method is deprecated and will be removed in a future release. Use APITokenService.Create instead.
// @Summary Create a new api token
// @Description Create a new api token to use on behalf of the user creating it.
// @tags api
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param token body models.APIToken true "The token object with required fields"
// @Success 200 {object} models.APIToken "The created token."
// @Failure 400 {object} web.HTTPError "Invalid token object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tokens [put]
func (t *APIToken) Create(s *xorm.Session, a web.Auth) (err error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return
	}
	return getAPITokenService().Create(s, t, u)
}

func HashToken(token, salt string) string {
	tempHash := pbkdf2.Key([]byte(token), []byte(salt), 10000, 50, sha256.New)
	return hex.EncodeToString(tempHash)
}

// ReadAll returns all api tokens the current user has created
// @Deprecated: This method is deprecated and will be removed in a future release. Use APITokenService.GetAll instead.
// @Summary Get all api tokens of the current user
// @Description Returns all api tokens the current user has created.
// @tags api
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param page query int false "The page number, used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of tokens per page. This parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search tokens by their title."
// @Success 200 {array} models.APIToken "The list of all tokens"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /tokens [get]
func (t *APIToken) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return nil, 0, 0, err
	}
	tokens, rc, total, err := getAPITokenService().GetAll(s, u, search, page, perPage)
	return tokens, rc, total, err
}

// Delete deletes a token
// @Deprecated: This method is deprecated and will be removed in a future release. Use APITokenService.Delete instead.
// @Summary Deletes an existing api token
// @Description Delete any of the user's api tokens.
// @tags api
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param tokenID path int true "Token ID"
// @Success 200 {object} models.Message "Successfully deleted."
// @Failure 404 {object} web.HTTPError "The token does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tokens/{tokenID} [delete]
func (t *APIToken) Delete(s *xorm.Session, a web.Auth) (err error) {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return
	}
	return getAPITokenService().Delete(s, t.ID, u)
}
