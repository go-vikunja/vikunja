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
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"slices"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"

	"golang.org/x/crypto/pbkdf2"
	"xorm.io/builder"
	"xorm.io/xorm"
)

type APIPermissions map[string][]string

type APIToken struct {
	// The unique, numeric id of this api key.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"token" readOnly:"true" doc:"The unique, numeric id of this api key."`

	// A human-readable name for this token
	Title string `xorm:"not null" json:"title" valid:"required" minLength:"1" doc:"A human-readable name for this token."`
	// The actual api key. Only visible after creation.
	Token string `xorm:"-" json:"token,omitempty" readOnly:"true" doc:"The cleartext api key. Returned only once, in the response to creating the token; never readable again."`
	// Legacy PBKDF2 columns, only populated on tokens created before TokenSha256 existed.
	TokenSalt      string `xorm:"null" json:"-"`
	TokenHash      string `xorm:"null unique" json:"-"`
	TokenLastEight string `xorm:"null index varchar(8)" json:"-"`
	TokenSha256    string `xorm:"varchar(64) null unique" json:"-"`
	// The permissions this token has. Possible values are available via the /routes endpoint and consist of the keys of the list from that endpoint. For example, if the token should be able to read all tasks as well as update existing tasks, you should add `{"tasks":["read_all","update"]}`.
	APIPermissions APIPermissions `xorm:"json not null permissions" json:"permissions" valid:"required" doc:"The permissions this token has. Possible values are available via the /routes endpoint and consist of the keys of the list from that endpoint. For example, if the token should be able to read all tasks as well as update existing tasks, you should add {\"tasks\":[\"read_all\",\"update\"]}."`
	// The date when this key expires.
	ExpiresAt time.Time `xorm:"not null" json:"expires_at" valid:"required" doc:"The date when this key expires."`

	// A timestamp when this api key was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created" readOnly:"true" doc:"A timestamp when this api key was created. You cannot change this value."`

	// The user ID of the token owner. When creating a token for a bot user, set this
	// to the bot's ID. If omitted, defaults to the authenticated user.
	OwnerID int64 `xorm:"bigint not null" json:"owner_id,omitempty" query:"owner_id" doc:"The user ID of the token owner. When creating a token for a bot user, set this to the bot's ID; the bot must be owned by the authenticated user. If omitted, defaults to the authenticated user."`

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
// @Success 201 {object} models.APIToken "The created token."
// @Failure 400 {object} web.HTTPError "Invalid token object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tokens [put]
func (t *APIToken) Create(s *xorm.Session, a web.Auth) (err error) {
	caller, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}

	owner := caller
	if t.OwnerID != 0 && t.OwnerID != caller.ID {
		owner, err = user.GetUserByID(s, t.OwnerID)
		if err != nil {
			return err
		}
		if !owner.IsBotOwnedBy(caller) {
			return &user.ErrBotNotOwned{UserID: t.OwnerID}
		}
	}

	return t.issue(s, owner, caller.ID)
}

// CreateInstanceBotToken mints a token for an instance bot. There is no API
// caller, so the doer is 0 and the audit log attributes it to the CLI.
func (t *APIToken) CreateInstanceBotToken(s *xorm.Session, bot *user.User) error {
	if !bot.IsInstanceBot {
		return &user.ErrBotNotOwned{UserID: bot.ID}
	}
	return t.issue(s, bot, 0)
}

func (t *APIToken) issue(s *xorm.Session, owner *user.User, doerID int64) error {
	if owner.IsInstanceBot {
		if err := validateInstanceBotPermissions(t.APIPermissions); err != nil {
			return err
		}
	}
	if err := PermissionsAreValid(t.APIPermissions); err != nil {
		return err
	}

	t.ID = 0
	t.OwnerID = owner.ID

	token, err := utils.CryptoRandomBytes(20)
	if err != nil {
		return err
	}
	t.Token = APITokenPrefix + hex.EncodeToString(token)
	t.TokenSha256 = HashAPIToken(t.Token)

	// Legacy columns stay NULL; without Nullable xorm would insert "" and trip the unique index on token_hash.
	_, err = s.Nullable("token_salt", "token_hash", "token_last_eight").Insert(t)
	if err != nil {
		return err
	}

	events.DispatchOnCommit(s, &APITokenIssuedEvent{
		TokenID: t.ID,
		DoerID:  doerID,
		OwnerID: t.OwnerID,
	})

	return nil
}

// HashToken is the legacy PBKDF2 hash, only kept to verify tokens created before token_sha256 existed.
func HashToken(token, salt string) string {
	tempHash := pbkdf2.Key([]byte(token), []byte(salt), 10000, 50, sha256.New)
	return hex.EncodeToString(tempHash)
}

// HashAPIToken uses plain SHA-256, same rationale as HashSessionToken: 160-bit random tokens gain nothing from a slow KDF.
func HashAPIToken(token string) string {
	return utils.Sha256Hex(token)
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

	caller, err := user.GetFromAuth(a)
	if err != nil {
		return nil, 0, 0, err
	}

	tokens := []*APIToken{}

	ownerID := caller.ID
	if t.OwnerID != 0 && t.OwnerID != caller.ID {
		botUser, lookupErr := user.GetUserByID(s, t.OwnerID)
		if lookupErr != nil {
			return nil, 0, 0, lookupErr
		}
		if !botUser.IsBotOwnedBy(caller) {
			return nil, 0, 0, &user.ErrBotNotOwned{UserID: t.OwnerID}
		}
		ownerID = t.OwnerID
	}

	var where builder.Cond = builder.Eq{"owner_id": ownerID}

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
	caller, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}

	// Ownership is verified in CanDelete; delete by ID only.
	return t.revoke(s, caller.ID)
}

// RevokeInstanceBotToken deletes a token of an instance bot on behalf of the CLI (doer 0).
func (t *APIToken) RevokeInstanceBotToken(s *xorm.Session, bot *user.User) error {
	if !bot.IsInstanceBot || t.OwnerID != bot.ID {
		return &user.ErrBotNotOwned{UserID: bot.ID}
	}
	return t.revoke(s, 0)
}

func (t *APIToken) revoke(s *xorm.Session, doerID int64) error {
	_, err := s.Where("id = ?", t.ID).Delete(&APIToken{})
	if err != nil {
		return err
	}

	events.DispatchOnCommit(s, &APITokenRevokedEvent{
		TokenID: t.ID,
		DoerID:  doerID,
	})

	return nil
}

// HasCaldavAccess checks whether the token has the caldav access permission.
func (t *APIToken) HasCaldavAccess() bool {
	perms, has := t.APIPermissions["caldav"]
	if !has {
		return false
	}
	return slices.Contains(perms, "access")
}

// HasFeedsAccess checks whether the token has the feeds access permission.
func (t *APIToken) HasFeedsAccess() bool {
	perms, has := t.APIPermissions["feeds"]
	if !has {
		return false
	}
	return slices.Contains(perms, "access")
}

// GetTokenFromTokenString returns the full token object from the original token string,
// backfilling token_sha256 when the token was only found via the legacy pbkdf2 path.
func GetTokenFromTokenString(s *xorm.Session, token string) (apiToken *APIToken, err error) {
	// The slice below would panic on a short string. Real tokens are prefix + 40
	// hex chars, so anything shorter is invalid by construction.
	if len(token) < len(APITokenPrefix)+8 {
		return nil, &ErrAPITokenInvalid{}
	}

	hash := HashAPIToken(token)

	apiToken = &APIToken{}
	found, err := s.Where(builder.Eq{"token_sha256": hash}).Get(apiToken)
	if err != nil {
		return nil, err
	}
	if found {
		return apiToken, nil
	}

	apiToken, err = getLegacyTokenFromTokenString(s, token)
	if err != nil {
		return nil, err
	}

	apiToken.TokenSha256 = hash
	backfillTokenSha256(apiToken.ID, hash)

	return apiToken, nil
}

func getLegacyTokenFromTokenString(s *xorm.Session, token string) (*APIToken, error) {
	lastEight := token[len(token)-8:]

	tokens := []*APIToken{}
	err := s.Where(builder.And(
		builder.Eq{"token_last_eight": lastEight},
		builder.IsNull{"token_sha256"},
	)).Find(&tokens)
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

// Own autocommit session because callers roll theirs back; the timeout keeps a starved pool from blocking auth.
func backfillTokenSha256(id int64, hash string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	s := db.NewAutocommitSession()
	defer s.Close()
	db.SetSessionContext(ctx, s)

	if _, err := s.ID(id).Cols("token_sha256").Update(&APIToken{TokenSha256: hash}); err != nil {
		log.Warningf("Could not backfill token_sha256 for api token %d: %s", id, err)
	}
}

// ValidateTokenAndGetOwner looks up a raw token string, checks it is not expired,
// and returns both the APIToken and its owner. Callers are responsible for checking
// permissions on the returned token (e.g. CanDoAPIRoute or HasCaldavAccess).
// Returns (nil, nil, nil) if the token is invalid or expired, or if the owner
// account is disabled/locked.
func ValidateTokenAndGetOwner(s *xorm.Session, rawToken string) (*APIToken, *user.User, error) {
	apiToken, err := GetTokenFromTokenString(s, rawToken)
	if err != nil {
		if IsErrAPITokenInvalid(err) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	if time.Now().After(apiToken.ExpiresAt) {
		return nil, nil, nil
	}

	u, err := user.GetUserByID(s, apiToken.OwnerID)
	if err != nil {
		if user.IsErrUserStatusError(err) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	return apiToken, u, nil
}
