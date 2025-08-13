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
	"crypto/subtle"
	"encoding/hex"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"

	"golang.org/x/crypto/pbkdf2"
	"xorm.io/builder"
	"xorm.io/xorm"
)

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

	salt, err := utils.CryptoRandomString(10)
	if err != nil {
		return err
	}
	token, err := utils.CryptoRandomBytes(20)
	if err != nil {
		return err
	}
	t.TokenSalt = salt
	t.Token = APITokenPrefix + hex.EncodeToString(token)
	t.TokenHash = HashToken(t.Token, t.TokenSalt)
	t.TokenLastEight = t.Token[len(t.Token)-8:]

	t.OwnerID = a.GetID()

	if err := PermissionsAreValid(t.APIPermissions); err != nil {
		return err
	}

	_, err = s.Insert(t)
	return err
}

func HashToken(token, salt string) string {
	tempHash := pbkdf2.Key([]byte(token), []byte(salt), 10000, 50, sha256.New)
	return hex.EncodeToString(tempHash)
}

// ReadAll returns all api tokens the current user has created
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

	tokens := []*APIToken{}

	var where builder.Cond = builder.Eq{"owner_id": a.GetID()}

	if search != "" {
		where = builder.And(
			where,
			db.ILIKE("api_tokens.title", search),
		)
	}

	err = s.
		Where(where).
		Limit(getLimitFromPageIndex(page, perPage)).
		Find(&tokens)
	if err != nil {
		return nil, 0, 0, err
	}

	totalCount, err := s.Where(where).Count(&APIToken{})
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

// GetTokenFromTokenString returns the full token object from the original token string.
func GetTokenFromTokenString(s *xorm.Session, token string) (apiToken *APIToken, err error) {
	lastEight := token[len(token)-8:]

	tokens := []*APIToken{}
	err = s.Where("token_last_eight = ?", lastEight).Find(&tokens)
	if err != nil {
		return nil, err
	}

	for _, t := range tokens {
		tempHash := HashToken(token, t.TokenSalt)
		if subtle.ConstantTimeCompare([]byte(t.TokenHash), []byte(tempHash)) == 1 {
			return t, nil
		}
	}

	return nil, &ErrAPITokenInvalid{}
}
