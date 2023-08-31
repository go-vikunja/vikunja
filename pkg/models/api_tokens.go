// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2023 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/utils"

	"code.vikunja.io/web"
	"xorm.io/xorm"
)

type APIPermissions map[string][]string

type APIToken struct {
	// The unique, numeric id of this api key.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"token"`

	// A human-readable name for this token
	Title string `xorm:"not null" json:"title" valid:"required"`
	// The actual api key. Only visible after creation.
	Key string `xorm:"not null varchar(50)" json:"key,omitempty"`
	// The permissions this token has. Possible values are available via the /routes endpoint and consist of the keys of the list from that endpoint. For example, if the token should be able to read all tasks as well as update existing tasks, you should add `{"tasks":["read_all","update"]}`.
	Permissions APIPermissions `xorm:"json not null" json:"permissions" valid:"required"`
	// The date when this key expires.
	ExpiresAt time.Time `xorm:"not null" json:"expires_at" valid:"required"`

	// A timestamp when this api key was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this api key was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	OwnerID int64 `xorm:"bigint not null" json:"-"`

	web.Rights   `xorm:"-" json:"-"`
	web.CRUDable `xorm:"-" json:"-"`
}

func (*APIToken) TableName() string {
	return "api_tokens"
}

func GetAPITokenByID(s *xorm.Session, id int64) (token *APIToken, err error) {
	token = &APIToken{}
	_, err = s.Where("id = ?", id).
		Get(token)
	return
}

// Create creates a new token
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
	t.ID = 0
	t.Key = "tk_" + utils.MakeRandomString(32)
	t.OwnerID = a.GetID()

	// TODO: validate permissions

	_, err = s.Insert(t)
	return err
}

// ReadAll returns all api tokens the current user has created
// @Summary Get all api tokens of the current user
// @Description Returns all api tokens the current user has created.
// @tags api
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param page query int false "The page number for tasks. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of tasks per bucket per page. This parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search tasks by task text."
// @Success 200 {array} models.APIToken "The list of all tokens"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /tokens [get]
func (t *APIToken) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {

	tokens := []*APIToken{}

	query := s.Where("owner_id = ?", a.GetID()).
		Limit(getLimitFromPageIndex(page, perPage))

	if search != "" {
		query = query.Where(db.ILIKE("title", search))
	}

	err = query.Find(&tokens)
	if err != nil {
		return nil, 0, 0, err
	}

	for _, token := range tokens {
		token.Key = ""
	}

	totalCount, err := query.Count(&APIToken{})
	return tokens, len(tokens), totalCount, err
}

// Delete deletes a token
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
	_, err = s.Where("id = ? AND owner_id = ?", t.ID, a.GetID()).Delete(&APIToken{})
	return err
}
